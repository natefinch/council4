package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestSequenceCounterTaxCapabilities(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, text string
		outerGate  bool
	}{
		{"fixed life rider", "Counter target spell unless its controller pays {3}. You gain 2 life.", false},
		{"X life rider", "Counter target spell unless its controller pays {X}. You gain 2 life.", false},
		{"dynamic multiplier", "Counter target spell unless its controller pays {1} for each card in your graveyard. You gain 2 life.", false},
		{"outer condition", "If you control a creature, counter target spell unless its controller pays {3}. You gain 2 life.", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Tax Probe", Layout: "normal", TypeLine: "Instant", OracleText: tt.text}
			assertCardPaths(t, card,
				"SpellAbility.Val.Modes[0].Sequence[0].PublishResult = \"unless-paid\"",
				"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Key = \"unless-paid\"",
				"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Succeeded = game.TriFalse",
				"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.GainLife).Amount.fixed = 2",
			)
			face := lowerSingleFace(t, card)
			sequence := face.SpellAbility.Val.Modes[0].Sequence
			if len(sequence) != 3 {
				t.Fatalf("instructions = %d, want Pay, CounterObject, GainLife", len(sequence))
			}

			pay, ok := sequence[0].Primitive.(game.Pay)
			if !ok || pay.Payment.Payer.Val != game.ObjectControllerReference(game.TargetStackObjectReference(0)) {
				t.Fatalf("payment = %#v, want target spell controller", sequence[0])
			}
			if _, ok := sequence[1].Primitive.(game.CounterObject); !ok {
				t.Fatalf("counter = %T", sequence[1].Primitive)
			}
			if _, ok := sequence[2].Primitive.(game.GainLife); !ok {
				t.Fatalf("rider = %T", sequence[2].Primitive)
			}
			if sequence[0].Condition.Exists != tt.outerGate || sequence[1].Condition.Exists != tt.outerGate ||
				sequence[2].Condition.Exists || sequence[2].ResultGate.Exists {
				t.Fatalf("condition routing = %#v", sequence)
			}
		})
	}
}

func TestSequenceCounterTaxRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Counter target spell unless you pay {3}. You gain 2 life.",
		"Counter target spell unless its controller pays {X}, where X is this creature's power. You gain 2 life.",
		"Counter target spell unless its controller pays {3} and 2 life. You gain 2 life.",
		"Counter target spell unless its controller pays {3}. If they do, draw a card.",
		"Counter target spell unless its controller pays {3}. Counter another target spell unless its controller pays {2}.",
		"Counter target spell unless its controller pays {3}. If you win, draw a card.",
		"Counter target spell unless its controller pays {3}. You gain 2 life and jump over the moon.",
		"Counter target spell unless its controller pays {3}. If you control a creature, counter another target spell.",
		"Counter target spell unless its controller pays {Y}. You gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Tax Near Miss", Layout: "normal", TypeLine: "Instant", OracleText: text})
		})
	}
}

func TestSequenceCounterTaxWithSeparateOptionalFlow(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Tax Probe", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell unless its controller pays {3}. You may draw a card. If you do, you gain 2 life.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[0].PublishResult = \"unless-paid\"",
		"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Key = \"unless-paid\"",
		"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.Draw).Amount.fixed = 1",
		"SpellAbility.Val.Modes[0].Sequence[3].Primitive.(game.GainLife).Amount.fixed = 2",
	)
	face := lowerSingleFace(t, card)
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 4 || sequence[2].PublishResult == "" ||
		sequence[2].PublishResult == sequence[0].PublishResult ||
		sequence[3].ResultGate.Val.Key != sequence[2].PublishResult {
		t.Fatalf("independent optional result routing = %#v", sequence)
	}
}

func TestSequenceCounterTaxOuterConditionWithoutRider(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Tax Probe", Layout: "normal", TypeLine: "Instant",
		OracleText: "If you control a creature, counter target spell unless its controller pays {3}.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[0].Condition.Exists = true",
		"SpellAbility.Val.Modes[0].Sequence[1].Condition.Exists = true",
		"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Key = \"unless-paid\"",
	)
}

func TestSequenceCounterTaxTypedOwnership(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*compiler.AbilityContent)
	}{
		{"wrong node", func(c *compiler.AbilityContent) { c.Effects[0].Payment.FailureConditionNodeID++ }},
		{"missing node", func(c *compiler.AbilityContent) { c.Effects[0].Payment.FailureConditionNodeID = -1 }},
		{"wrong owner", func(c *compiler.AbilityContent) { c.Effects[0].Kind = compiler.EffectDraw }},
		{"unsupported predicate", func(c *compiler.AbilityContent) { c.Conditions[0].Predicate = compiler.ConditionPredicateUnsupported }},
		{"intervening", func(c *compiler.AbilityContent) { c.Conditions[0].Intervening = true }},
		{"wrong kind", func(c *compiler.AbilityContent) { c.Conditions[0].Kind = compiler.ConditionIf }},
		{"wrong polarity", func(c *compiler.AbilityContent) { c.Conditions[0].Negated = false }},
		{"duplicate condition", func(c *compiler.AbilityContent) { c.Conditions = append(c.Conditions, c.Conditions[0]) }},
		{"wrong form", func(c *compiler.AbilityContent) { c.Effects[0].Payment.Form = parser.EffectPaymentFormMayPayThenIfDo }},
		{"extra cost", func(c *compiler.AbilityContent) { c.Effects[0].Payment.AdditionalCost = &compiler.CompiledCost{} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(
				"Counter target spell unless its controller pays {3}. You gain 2 life.",
				parser.Context{CardName: "Tax Probe", InstantOrSorcery: true}, compiler.Context{},
			)
			if len(diagnostics) != 0 {
				t.Fatalf("compile = %#v", diagnostics)
			}
			content := compilation.Abilities[0].Content
			tt.mutate(&content)
			flow, ok := planOptionalFlow(content)
			if !ok {
				t.Fatal("unexpected optional-flow refusal")
			}
			if _, _, ok := planSequenceConditions(content, flow); ok {
				t.Fatal("malformed payment ownership was accepted")
			}
		})
	}
}

func TestCounterTaxMalformedManaFailsClosed(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileTestOracle(
		"Counter target spell unless its controller pays {3}.",
		parser.Context{CardName: "Tax Probe", InstantOrSorcery: true}, compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("compile = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	content.Effects[0].Payment.ManaCost = cost.Mana{cost.X, cost.U}
	if _, ok := lowerCounterUnlessPaysSpell(contentCtx{content: content}); ok {
		t.Fatal("mixed variable mana payment was accepted")
	}
}
