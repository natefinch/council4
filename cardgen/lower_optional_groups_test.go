package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func TestOptionalActionGroupCardPaths(t *testing.T) {
	for _, tt := range []struct {
		name, text string
		paths      []string
		absent     []string
	}{
		{
			name: "distinct primitives",
			text: "You may draw a card and gain 2 life. Scry 1.",
			paths: []string{
				"Sequence[0].Optional = true",
				"Sequence[0].PublishOptionalDecision = \"optional-clause-1\"",
				"Sequence[1].OptionalDecisionGate = \"optional-clause-1\"",
			},
			absent: []string{"Sequence[1].Optional = true", "Sequence[2].OptionalDecisionGate", "ResultGate.Exists = true"},
		},
		{
			name: "conditioned expanded action",
			text: "Tap target artifact. If you have no cards in hand, you may put a +1/+1 counter on each of up to two target creatures.",
			paths: []string{
				"Sequence[1].Optional = true",
				"Sequence[1].PublishOptionalDecision = \"optional-clause-2\"",
				"Sequence[2].OptionalDecisionGate = \"optional-clause-2\"",
				"Sequence[1].PublishCondition",
				"Sequence[2].ConditionGate",
			},
			absent: []string{"Sequence[0].Optional = true", "Sequence[2].Optional = true"},
		},
		{
			name: "single expanded action",
			text: "You may put a +1/+1 counter on each of up to two target creatures.",
			paths: []string{
				"Sequence[0].Optional = true",
				"Sequence[0].PublishOptionalDecision",
				"Sequence[1].OptionalDecisionGate",
			},
			absent: []string{"Sequence[1].Optional = true"},
		},
		{
			name: "independent decisions and riders",
			text: "You may draw a card and gain 2 life. Scry 1. You may draw a card and gain 3 life. You gain 1 life.",
			paths: []string{
				"Sequence[0].PublishOptionalDecision = \"optional-clause-1\"",
				"Sequence[1].OptionalDecisionGate = \"optional-clause-1\"",
				"Sequence[3].PublishOptionalDecision = \"optional-clause-4\"",
				"Sequence[4].OptionalDecisionGate = \"optional-clause-4\"",
			},
			absent: []string{"Sequence[2].OptionalDecisionGate", "Sequence[5].OptionalDecisionGate"},
		},
		{
			name: "repeat real body",
			text: "Repeat the following process two times. You may draw a card and gain 2 life.",
			paths: []string{
				"Primitive.(game.RepeatProcess).Times.fixed = 2",
				"Primitive.(game.RepeatProcess).Body.Modes[0].Sequence[0].PublishOptionalDecision",
				"Primitive.(game.RepeatProcess).Body.Modes[0].Sequence[1].OptionalDecisionGate",
			},
		},
		{
			name:   "mixed color expanded action",
			text:   "You may add {R}{G}.",
			paths:  []string{"Sequence[0].PublishOptionalDecision", "Sequence[1].OptionalDecisionGate"},
			absent: []string{"Sequence[1].Optional = true"},
		},
		{
			name:   "independent same subject rider",
			text:   "You may untap target creature and gain control of it until end of turn. It gains haste until end of turn.",
			paths:  []string{"Sequence[0].PublishOptionalDecision", "Sequence[1].OptionalDecisionGate"},
			absent: []string{"Sequence[2].Optional = true", "Sequence[2].OptionalDecisionGate", "gain-keyword-2"},
		},
		{
			name: "actual entered object rider",
			text: "You may gain 1 life and return target creature card from your graveyard to the battlefield. It gains haste until end of turn.",
			paths: []string{
				"Sequence[1].Primitive.(game.PutOnBattlefield).PublishLinked = \"sequence-effect-1-product\"",
				"Sequence[2].Primitive.(game.ApplyContinuous).Object.Val.linkID = \"sequence-effect-1-product\"",
			},
			absent: []string{"Sequence[2].Optional = true", "Sequence[2].OptionalDecisionGate"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Optional Group Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: tt.text}
			assertCardPaths(t, card, tt.paths...)
			assertCardPathsAbsent(t, card, tt.absent...)
		})
	}
}

func TestOptionalActionGroupCompatibleShells(t *testing.T) {
	for _, tt := range []struct{ name, typ, text, path string }{
		{"spell", "Sorcery", "You may draw a card and gain 2 life.", "SpellAbility.Val.Modes[0].Sequence"},
		{"activated", "Artifact", "{1}: You may draw a card and gain 2 life.", "ActivatedAbilities[0].Content.Modes[0].Sequence"},
		{"triggered", "Enchantment", "At the beginning of your upkeep, you may draw a card and gain 2 life.", "TriggeredAbilities[0].Content.Modes[0].Sequence"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Optional Shell", Layout: "normal", TypeLine: tt.typ, OracleText: tt.text}
			assertCardPaths(t, card, tt.path+"[0].Optional = true",
				tt.path+"[0].PublishOptionalDecision", tt.path+"[1].OptionalDecisionGate")
			assertCardPathsAbsent(t, card, tt.path+"[1].Optional = true")
		})
	}
}

func TestOptionalEnteredObjectUsesTypedProducer(t *testing.T) {
	for _, tt := range []struct{ name, text, producer, rider, key string }{
		{"nonadjacent", "You may return target creature card from your graveyard to the battlefield and gain 1 life. It gains haste until end of turn.", "Sequence[0]", "Sequence[2]", "sequence-effect-0-product"},
		{"expanded intervener", "You may return target creature card from your graveyard to the battlefield and add {R}{G}. It gains haste until end of turn.", "Sequence[0]", "Sequence[3]", "sequence-effect-0-product"},
		{"unconditional intervener", "You may gain 1 life and return target creature card from your graveyard to the battlefield. You gain 2 life. It gains haste until end of turn.", "Sequence[1]", "Sequence[3]", "sequence-effect-1-product"},
		{"nonzero target", "Tap target artifact. You may gain 1 life and return target creature card from your graveyard to the battlefield. It gains haste until end of turn.", "Sequence[2]", "Sequence[3]", "sequence-effect-2-product"},
		{"permanent keyword", "You may return target creature card from your graveyard to the battlefield and gain 1 life. It gains haste.", "Sequence[0]", "Sequence[2]", "sequence-effect-0-product"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Optional Entered Subject", Layout: "normal", TypeLine: "Sorcery", OracleText: tt.text}
			assertCardPaths(t, card,
				tt.producer+".Primitive.(game.PutOnBattlefield).PublishLinked = \""+tt.key+"\"",
				tt.rider+".Primitive.(game.ApplyContinuous).Object.Val.linkID = \""+tt.key+"\"")
			assertCardPathsAbsent(t, card, tt.rider+".Optional = true", tt.rider+".OptionalDecisionGate")
		})
	}
}

func TestOptionalGroupCaptureRequiresAvailablePublication(t *testing.T) {
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Optional Capture Near Miss", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may gain 1 life and return target creature card from your graveyard to the battlefield. It gains haste until end of turn. Exile it at the beginning of the next end step.",
	}, "optional linked-object publication has no skipped-availability contract")
}

func TestOptionalGroupIdentityIsTyped(t *testing.T) {
	document, diagnostics := parser.Parse("You may draw a card and gain 2 life. Scry 1.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for i := range content.Effects {
		content.Effects[i].Text = "opaque"
		content.Effects[i].Order = shared.SourceOrder{}
		content.Effects[i].Span = shared.Span{}
	}
	plan, ok := planOptionalFlow(content)
	if !ok || plan.scoped.actions.owners[1] != 0 {
		t.Fatalf("typed group was not retained: %#v", plan)
	}
	for _, ids := range [][]int{nil, {1, 999}, {2, 1}, {1, 3}, {1, 1}, {1, 2, 2}} {
		content.Effects[0].OptionalActionClauseIDs = ids
		if _, ok := planOptionalFlow(content); ok {
			t.Fatalf("invalid group admitted: %v", ids)
		}
	}
}

func TestOptionalDecisionPreservesActualResultEnvelope(t *testing.T) {
	groups := optionalActionGroups{
		owners: map[int]int{0: 0, 1: 0},
		keys:   map[int]game.OptionalDecisionKey{0: "choice"}, compound: map[int]bool{0: true},
	}
	sequence := []game.Instruction{{
		Primitive:     game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		PublishResult: "actual",
	}}
	if reason := groups.apply(0, sequence); reason != "" {
		t.Fatal(reason)
	}
	if sequence[0].PublishResult != "actual" || sequence[0].PublishOptionalDecision != "choice" {
		t.Fatal("choice overwrote primitive result publication")
	}
}

func TestOptionalGroupCannotInventWholeActionOutcome(t *testing.T) {
	document, diagnostics := parser.Parse("You may draw a card and gain 2 life. If you do, scry 1.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for _, producer := range []int{content.Effects[0].ClauseID, content.Effects[1].ClauseID} {
		content.Conditions[0].Ownership.ResultProducerClauseID = producer
		plan, ok := planOptionalFlow(content)
		if ok || plan.failureCategory != "structural — whole optional action outcome not modeled" {
			t.Fatalf("invented group outcome from primitive %d: %#v", producer, plan)
		}
	}
}

func TestOptionalGroupDoesNotClaimAnotherGroupsProduct(t *testing.T) {
	content := compiler.AbilityContent{Effects: []compiler.CompiledEffect{
		{Optional: true}, {Optional: true}, {References: []compiler.CompiledReference{{
			Binding: compiler.ReferenceBindingPriorInstructionResult, PriorInstruction: 0,
		}}},
	}}
	groups := optionalActionGroups{owners: map[int]int{0: 0, 1: 1, 2: 1}}
	if groups.ownsOptionalAntecedents(content, 2) {
		t.Fatal("continuing a group bypassed another optional producer's availability boundary")
	}
	content.Effects[2].References[0].PriorInstruction = 1
	if !groups.ownsOptionalAntecedents(content, 2) {
		t.Fatal("group did not own its accepted producer")
	}
}
