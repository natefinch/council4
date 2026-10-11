package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestResolvingAlternativeParagraphsRefuseUnownedReplacements(t *testing.T) {
	for _, text := range []string{
		"Create a 4/4 white Angel creature token with flying instead.",
		"Put a +1/+1 counter on target creature.\nUnknown label — Put three +1/+1 counters on that creature instead if a creature died this turn.",
		"The next time damage would be dealt to target creature this turn, destroy that creature instead.",
		"Choose one —\n• Draw a card.\n• Gain 2 life.\nMorbid — Draw three cards instead if a creature died this turn.",
		"Brimstone Volley deals 3 damage to any target.\nMorbid — Brimstone Volley deals 5 damage and draws a card instead if a creature died this turn.",
		"Brimstone Volley deals 5 damage instead if a creature died this turn.",
		"Put a +1/+1 counter on target creature.\nMorbid — Put three +1/+1 counters on that creature instead if a creature died this turn.\nFateful hour — If you have 5 or less life, put four +1/+1 counters on that creature instead.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unowned Alternative", Layout: "normal", TypeLine: "Instant", OracleText: text,
			})
		})
	}
}

func TestResolvingAlternativeParagraphCounterSharesTypedTarget(t *testing.T) {
	card := &ScryfallCard{
		Name: "Resolving Counter Alternative", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target noncreature spell unless its controller pays {1}.\nFerocious — If you control a creature with power 4 or greater, counter that spell instead.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Pay).Payment.Payer.Val.object.Val.kind = game.ObjectReferenceTargetStackObject",
		"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CounterObject).Object.kind = game.ObjectReferenceTargetStackObject",
		"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.CounterObject).Object.kind = game.ObjectReferenceTargetStackObject",
	)
	for _, text := range []string{
		"Counter that spell.",
		"Counter target spell. Counter that creature.",
		"Counter target spell. Counter another spell.",
	} {
		assertCardUnsupported(t, &ScryfallCard{
			Name: "Unproved Counter Recipient", Layout: "normal", TypeLine: "Instant", OracleText: text,
		})
	}
}

func TestResolvingAlternativeParagraphsUseSharedAssembly(t *testing.T) {
	for _, test := range []struct {
		name, text string
	}{
		{"Hunger of the Howlpack", "Put a +1/+1 counter on target creature.\nMorbid — Put three +1/+1 counters on that creature instead if a creature died this turn."},
		{"Might Beyond Reason", "Put two +1/+1 counters on target creature.\nDelirium — Put three +1/+1 counters on that creature instead if there are four or more card types among cards in your graveyard."},
	} {
		t.Run(test.name, func(t *testing.T) {
			face := lowerSingleFace(t, &ScryfallCard{Name: test.name, Layout: "normal", TypeLine: "Instant", OracleText: test.text})
			content := face.SpellAbility.Val
			if len(content.Modes) != 1 || len(content.Modes[0].Targets) != 1 ||
				len(content.Modes[0].Sequence) != 2 {
				t.Fatalf("shared resolving branch shape=%+v", content)
			}

			sequence := content.Modes[0].Sequence
			for _, instruction := range sequence {
				if instruction.Primitive.Kind() != game.PrimitiveAddCounter {
					t.Fatalf("counter branch=%+v", instruction)
				}
			}

			assertConditionConsumer(t, sequence[0], sequence[1], true)
		})
	}
}
