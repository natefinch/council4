package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestReferencedPlayerActionOwnsActualResult(t *testing.T) {
	for _, text := range []string{
		"Target player gains 1 life. That player discards a card. If that player does, you gain 2 life.",
		"Target player gains 1 life. That player mills a card. You gain 1 life. If that player does, you gain 2 life.",
		"Target player gains 1 life. That player may discard a card. You gain 1 life. If that player does, you gain 2 life.",
		"Target player gains 1 life. That player may mill a card. You gain 1 life. If that player does, you gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			compiled, diagnostics := compileTestOracle(text, parser.Context{InstantOrSorcery: true}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}

			body := compiled.Abilities[0].Content
			found := false
			for _, c := range body.Conditions {
				if c.Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
					continue
				}
				found = true
				if c.Ownership.ResultProducerClauseID != 2 {
					t.Fatalf("producer=%d want discard/mill clause 2", c.Ownership.ResultProducerClauseID)
				}
			}
			if !found {
				t.Fatal("missing actual-result condition")
			}
			defs, diagnostics, err := CompileCardDefs(&ScryfallCard{Name: "Actor Result", Layout: "normal", TypeLine: "Instant", OracleText: text})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: %v %v", err, diagnostics)
			}
			sequence := defs[0].SpellAbility.Val.Modes[0].Sequence
			if len(sequence) != len(body.Effects) || sequence[0].PublishResult != "" ||
				sequence[1].PublishResult != "if-you-do" ||
				!slices.Contains(sequence[1].LocalProducts.Results, game.ResultKey("if-you-do")) {
				t.Fatalf("actual action does not exclusively own its local publication: instructions=%d want=%d publishers=%q/%q locals=%v firstLocals=%v", len(sequence), len(body.Effects), sequence[0].PublishResult, sequence[1].PublishResult, sequence[1].LocalProducts.Results, sequence[0].LocalProducts.Results)
			}
			consumer := sequence[len(sequence)-1]
			if !consumer.ResultGate.Exists || consumer.ResultGate.Val.Key != sequence[1].PublishResult ||
				consumer.ResultGate.Val.Succeeded != game.TriTrue {
				t.Fatal("consequence does not consume the action's actual success")
			}
			if len(sequence) == 4 && (sequence[2].ResultGate.Exists || sequence[2].PublishResult != "") {
				t.Fatal("independent controller rider acquired the actor's receipt")
			}
		})
	}
}

func TestControllerActorOwnsActualResult(t *testing.T) {
	for _, text := range []string{
		"You gain 1 life. If you do, you gain 2 life.",
		"You may exile target creature. If you do, you gain 2 life.",
		"You may destroy target creature. If you do, you gain 2 life.",
	} {
		assertCardPaths(t, &ScryfallCard{Name: "Controller Actor", Layout: "normal", TypeLine: "Instant", OracleText: text},
			"PublishResult", "ResultGate.Val.Succeeded = game.TriTrue")
	}
}

func TestControllerCausativeActorOwnsActualResult(t *testing.T) {
	const text = "Whenever a card is put into an opponent's graveyard from anywhere, you may have that player lose 2 life. If you do, you gain 2 life."
	document, diagnostics := parser.Parse(text, parser.Context{CardName: "Causative Actor"})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	ctx := contentCtx{
		content: compilation.Abilities[0].Content, text: text, span: document.Abilities[0].Span,
		enclosingKind: compiler.AbilityTriggered, triggerEvent: game.EventZoneChanged,
	}
	effects, controllerActions := collapseCausativeHavePairs(ctx.content.Effects, ctx.content.Keywords)
	if len(effects) != 2 || !controllerActions[effects[0].ClauseID] ||
		effects[0].Context != parser.EffectContextReferencedPlayer ||
		len(effects[0].SubjectReferences) != 1 ||
		effects[0].SubjectReferences[0].NodeID != ctx.content.Effects[1].SubjectReferences[0].NodeID {
		t.Fatal("causative chooser proof overwrote the affected player's canonical identity")
	}
	for _, ids := range [][]int{nil, {1}, {2, 1}, {1, 99}, {1, 1}} {
		malformed := slices.Clone(ctx.content.Effects)
		malformed[0].OptionalActionClauseIDs = ids
		_, actors := collapseCausativeHavePairs(malformed, ctx.content.Keywords)
		if len(actors) != 0 {
			t.Fatalf("invalid original optional group %v issued controller ownership", ids)
		}
	}
	content, diagnostic := lowerOrderedEffectSequence("Causative Actor", ctx, &document.Abilities[0])
	if diagnostic != nil {
		t.Fatalf("controller-caused action lost its owned chooser: %#v", diagnostic)
	}
	sequence := content.Modes[0].Sequence
	if len(sequence) != 2 || !sequence[0].Optional || sequence[0].OptionalActor.Exists ||
		sequence[0].PublishResult == "" || !sequence[1].ResultGate.Exists ||
		sequence[1].ResultGate.Val.Key != sequence[0].PublishResult {
		t.Fatal("controller chooser and actual action receipt not preserved")
	}
}

func TestEventPlayerActionOwnsActualResult(t *testing.T) {
	for _, text := range []string{
		"At the beginning of each player's upkeep, that player discards a card. You gain 1 life. If that player does, you gain 2 life.",
		"At the beginning of each player's upkeep, that player may mill a card. You gain 1 life. If they do, you gain 2 life.",
		"Whenever a creature deals combat damage to one of your opponents, its controller may pay 1 life. If they do, they draw a card.",
	} {
		assertCardPaths(t, &ScryfallCard{Name: "Event Actor", Layout: "normal", TypeLine: "Enchantment", OracleText: text},
			"PublishResult", "ResultGate.Val.Succeeded = game.TriTrue")
	}
}

func TestReferencedPlayerActionRejectsInvalidOwnership(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
	}{
		{"missing producer", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultProducerClauseID = 0 }},
		{"forward producer", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultProducerClauseID = 4 }},
		{"declaration instead of action", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultProducerClauseID = 1 }},
		{"foreign actor", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultSubjectReferenceNodeID = 99 }},
		{"missing actor membership", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ReferenceNodeIDs = nil }},
		{"action reference substituted for condition actor", func(c *compiler.AbilityContent) {
			c.Conditions[0].Ownership.ReferenceNodeIDs = []int{c.Effects[1].SubjectReferences[0].NodeID}
		}},
		{"duplicate actor", func(c *compiler.AbilityContent) { c.References = append(c.References, c.References[1]) }},
		{"missing actor proof", func(c *compiler.AbilityContent) { c.References[1].Subject = compiler.ReferenceSubjectProof{} }},
		{"missing action proof", func(c *compiler.AbilityContent) {
			c.Effects[1].SubjectReferences[0].Subject = compiler.ReferenceSubjectProof{}
		}},
		{"foreign target", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultSubjectTargetOccurrence = 99 }},
		{"plural target", func(c *compiler.AbilityContent) { c.Targets[0].Cardinality.Max = 2 }},
		{"foreign action node", func(c *compiler.AbilityContent) { c.Effects[1].SubjectReferences[0].NodeID = 99 }},
		{"forward action reference", func(c *compiler.AbilityContent) { c.References[0].Order = c.References[1].Order }},
		{"controller action", func(c *compiler.AbilityContent) { c.Effects[1].Context = parser.EffectContextController }},
		{"duplicate producer", func(c *compiler.AbilityContent) { c.Effects[2].ClauseID = 2 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compiled, diagnostics := compileTestOracle(
				"Target player gains 1 life. That player discards a card. You gain 1 life. If that player does, you gain 2 life.",
				parser.Context{InstantOrSorcery: true}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compiled.Abilities[0].Content
			tc.mutate(&content)
			if _, ok := planOptionalFlow(content); ok {
				t.Fatal("invalid actor/action proof acquired a result flow")
			}
		})
	}
	for _, text := range []string{
		"Target player gains 1 life. Each player discards a card. If that player does, you gain 2 life.",
		"That player discards a card. If that player does, you gain 2 life.",
		"That player discards a card. If that player does, you gain 2 life. Target player gains 1 life.",
		"Up to two target players each gain 1 life. They discard a card. If they do, you gain 2 life.",
		"Whenever a creature deals combat damage to one of your opponents, its controller may pay 1 life. If you do, draw a card.",
		"Whenever a creature deals combat damage to one of your opponents, its owner may pay 1 life. If they do, they draw a card.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Actor Near Miss", Layout: "normal", TypeLine: "Instant", OracleText: text})
	}
}
