package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

const counterExileReplacement = "If that spell is countered this way, exile it instead of putting it into its owner's graveyard."

func TestCounterDestinationComposition(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text string
		paths      []string
	}{
		{"tax", "Counter target spell unless its controller pays {3}. " + counterExileReplacement,
			[]string{
				"SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Pay).Payment.ManaCost.Val[0] = {3}",
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).ExileInstead = true",
				"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Key = \"unless-paid\"",
			}},
		{"x tax with riders", "You gain 2 life. Counter target spell unless its controller pays {X}. " + counterExileReplacement + " Scry 1.",
			[]string{
				"SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.GainLife).Amount.fixed = 2",
				"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.CounterObject).ExileInstead = true",
				"SpellAbility.Val.Modes[0].Sequence[3].Primitive.(game.Scry).Amount.fixed = 1",
			}},
		{"nonzero counter slot", "Tap target creature. Counter target spell. " + counterExileReplacement + " You gain 2 life.",
			[]string{
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).Object.targetIndex = 1",
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).ExileInstead = true",
				"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.GainLife).Amount.fixed = 2",
			}},
		{"preceding independent exile", "Exile target creature. Counter target spell. " + counterExileReplacement + " You gain 2 life.",
			[]string{
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).Object.targetIndex = 1",
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).ExileInstead = true",
			}},
		{"independent counter destinations", "Counter target spell. " + counterExileReplacement +
			" Counter target spell. If that spell is countered this way, put it into its owner's hand instead of into that player's graveyard. You gain 2 life.",
			[]string{
				"SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CounterObject).ExileInstead = true",
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).Destination = game.CounteredSpellHand",
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).Object.targetIndex = 1",
			}},
		{"conditioned tax", "If you control a creature, counter target spell unless its controller pays {3}. " + counterExileReplacement + " You gain 2 life.",
			[]string{
				"SpellAbility.Val.Modes[0].Sequence[0].PublishCondition = \"condition-0\"",
				"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).ExileInstead = true",
			}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertCardPaths(t, &ScryfallCard{
				Name: "Counter Probe", Layout: "normal", TypeLine: "Instant", OracleText: tt.text,
			}, tt.paths...)
		})
	}
}

func TestCounterActualSuccessComposition(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text      string
		publisher int
		consumer  int
	}{
		{"Counter target spell. You gain 2 life. If that spell is countered this way, draw a card. Scry 1.", 0, 2},
		{"Exile target creature. Counter target spell. If that spell is countered this way, draw a card.", 1, 2},
		{"Counter target spell unless its controller pays {3}. If that spell is countered this way, you gain 2 life. Scry 1.", 1, 2},
		{"Counter target spell unless its controller pays {X}. " + counterExileReplacement + " If that spell is countered this way, you gain 2 life.", 1, 2},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			defs, diagnostics, err := CompileCardDefs(&ScryfallCard{
				Name: "Counter Probe", Layout: "normal", TypeLine: "Instant", OracleText: tt.text,
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: %v %#v", err, diagnostics)
			}
			sequence := defs[0].SpellAbility.Val.Modes[0].Sequence
			if sequence[tt.publisher].PublishResult == "" ||
				!sequence[tt.consumer].ResultGate.Exists ||
				sequence[tt.consumer].ResultGate.Val.Key != sequence[tt.publisher].PublishResult {
				t.Fatalf("actual counter result wiring = %#v", sequence)
			}
		})
	}
}

func TestCounterCompositionRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Counter target spell. Exile target creature. " + counterExileReplacement,
		"If that spell is countered this way, draw a card. Counter target spell.",
		"Counter up to two target spells. " + counterExileReplacement,
		"Counter target spell. If that spell is countered this way, put it on the bottom of its owner's library instead of into that player's graveyard.",
		"Counter target spell. If that spell is countered this way, put that card on your choice of the top or bottom of its owner's library instead of into that player's graveyard.",
		"Counter target non-Faerie spell. " + counterExileReplacement,
		"Counter target spell. If you control a creature, exile it instead of putting it into its owner's graveyard.",
		"Counter target spell. " + counterExileReplacement + " " + counterExileReplacement,
		"Counter target spell unless its controller pays {3}. If you do, draw a card.",
		"Counter target spell unless its controller pays {3}. Counter target spell unless its controller pays {4}. " + counterExileReplacement,
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Counter Probe", Layout: "normal", TypeLine: "Instant", OracleText: text,
			})
		})
	}

}

func TestCounterCompositionShells(t *testing.T) {
	t.Parallel()
	body := "Counter target spell unless its controller pays {3}. " + counterExileReplacement +
		" If that spell is countered this way, you gain 2 life. You gain 1 life."
	for _, tt := range []struct {
		name, typ, text, path string
	}{
		{"spell", "Instant", body, "SpellAbility.Val.Modes[0].Sequence"},
		{"activated", "Artifact", "{1}: " + body, "ActivatedAbilities[0].Content.Modes[0].Sequence"},
		{"triggered", "Artifact", "When this artifact enters, " + body, "TriggeredAbilities[0].Content.Modes[0].Sequence"},
		{"modal", "Instant", "Choose one \u2014\n\u2022 " + body + "\n\u2022 Draw a card.", "SpellAbility.Val.Modes[0].Sequence"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Shell Counter", Layout: "normal", TypeLine: tt.typ, OracleText: tt.text}
			assertCardPaths(t, card,
				tt.path+"[1].Primitive.(game.CounterObject).ExileInstead = true",
				tt.path+"[1].PublishResult = \"if-you-do\"",
				tt.path+"[2].ResultGate.Val.Key = \"if-you-do\"",
			)
			assertCardPathsAbsent(t, card, tt.path+"[3].ResultGate.Exists = true",
				"Primitive.(game.MoveCard)", "Primitive.(game.ExileTargetSpells)")
		})
	}
	assertCardPaths(t, &ScryfallCard{
		Name: "Counter Token Result", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell. If that spell is countered this way, create a Treasure token.",
	}, "Sequence[1].Primitive.(game.CreateToken)", "Sequence[1].ResultGate.Val.Succeeded = game.TriTrue")
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Counter Modes Refusal", Layout: "normal", TypeLine: "Instant",
		OracleText: "Choose two \u2014\n\u2022 " + body + "\n\u2022 " + body,
	}, "cross-mode conditional result publication requires scoped namespaces")
}

func TestCounterDestinationPlannerTypedOwnersAndNearMisses(t *testing.T) {
	t.Parallel()
	text := "Tap target creature. Counter target spell unless its controller pays {3}. " + counterExileReplacement
	document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compiled, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compiled.Abilities[0].Content
	for ei := range content.Effects {
		content.Effects[ei].Text = "opaque"
		content.Effects[ei].Span = shared.Span{}
		content.Effects[ei].ClauseSpan = shared.Span{}
	}
	for ci := range content.Conditions {
		content.Conditions[ci].Text = "opaque"
		content.Conditions[ci].Span = shared.Span{}
	}
	plan, _, reason := planCounterDestinations(content)
	if reason != "" || !plan.modifiers[1].ExileInstead || !plan.absorbed[2] {
		t.Fatalf("typed destination plan=%#v, reason=%q", plan, reason)
	}
	for _, tt := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
	}{
		{"missing producer", func(c *compiler.AbilityContent) {
			c.Conditions[1].Ownership.ResultProducerClauseID = 999
		}},
		{"forward producer", func(c *compiler.AbilityContent) {
			c.Conditions[1].Ownership.ResultProducerClauseID = c.Effects[2].ClauseID
		}},
		{"duplicate producer", func(c *compiler.AbilityContent) {
			c.Effects[0].ClauseID = c.Effects[1].ClauseID
		}},
		{"missing subject", func(c *compiler.AbilityContent) {
			c.Conditions[1].Ownership.ResultSubjectReferenceNodeID = 999
		}},
		{"wrong target occurrence", func(c *compiler.AbilityContent) {
			for i := range c.References {
				if c.References[i].NodeID == c.Conditions[1].Ownership.ResultSubjectReferenceNodeID {
					c.References[i].Occurrence = 0
				}
			}
		}},
		{"different predicate", func(c *compiler.AbilityContent) {
			c.Conditions[1].Predicate = compiler.ConditionPredicatePriorInstructionAccepted
		}},
		{"duplicate target identity", func(c *compiler.AbilityContent) {
			c.Targets[0].Order = c.Targets[1].Order
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Compile afresh so each mutation owns its nested slices.
			fresh, diags := compiler.Compile(document, compiler.Context{})
			if len(diags) != 0 {
				t.Fatal(diags)
			}
			candidate := fresh.Abilities[0].Content
			tt.mutate(&candidate)
			if _, _, reason := planCounterDestinations(candidate); reason == "" {
				t.Fatal("unproven ownership was admitted")
			}
		})
	}
}
