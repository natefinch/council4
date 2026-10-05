package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestFixedPhaseCapturedSubjectComposes(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, text string
		group      bool
		target     int
	}{
		{"product after draw", "Create a 1/1 green Insect creature token. Draw a card. Sacrifice it at the beginning of the next end step.", true, -1},
		{"product after life", "Create two 1/1 green Insect creature tokens. You gain 2 life. Exile them at end of combat.", true, -1},
		{"optional product", "You may create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step.", true, -1},
		{"optional product result", "You may create a 1/1 green Insect creature token. If you do, sacrifice it at the beginning of the next end step.", true, -1},
		{"future group state", "Create a 1/1 green Insect creature token. At the beginning of the next end step, sacrifice it if you have 20 or more life.", true, -1},
		{"target after draw", "Target creature gets +2/+2 until end of turn. Draw a card. Return it to its owner's hand at the beginning of the next end step.", false, 0},
		{"nonzero target", "Tap target creature. Untap target creature. Draw a card. Exile it at the beginning of the next end step.", false, 1},
		{"typed decoy", "Tap target creature. Untap target land. Sacrifice that creature at the beginning of the next end step.", false, 0},
		{"future subject predicate", "Tap target creature. At the beginning of the next end step, if it's an Elf, destroy it.", false, 0},
		{"optional target", "Up to one target creature gets +2/+2 until end of turn. Return that creature to its owner's hand at end of combat.", false, 0},
		{"returned product", "Return target creature card from your graveyard to the battlefield. You gain 1 life. Sacrifice that creature at the beginning of the next end step.", false, -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{Name: "Captured Subject", Layout: "normal", TypeLine: "Sorcery", OracleText: test.text})
			sequence := face.SpellAbility.Val.Modes[0].Sequence
			delayed, ok := sequence[len(sequence)-1].Primitive.(game.CreateDelayedTrigger)
			if !ok {
				t.Fatalf("last primitive = %T", sequence[len(sequence)-1].Primitive)
			}
			if test.group {
				if !delayed.Trigger.CapturedObjectGroup.Exists {
					t.Fatal("actual produced group was not captured")
				}
			} else if test.target < 0 {
				if !delayed.Trigger.CapturedObject.Exists || delayed.Trigger.CapturedObject.Val.Kind() != game.ObjectReferenceLinkedObject {
					t.Fatal("returned actual permanent was not captured")
				}
			} else if !delayed.Trigger.CapturedObject.Exists ||
				delayed.Trigger.CapturedObject.Val != game.TargetPermanentReference(test.target) {
				t.Fatalf("capture = %#v, want target %d", delayed.Trigger.CapturedObject, test.target)
			}
		})
	}
}

func TestFixedPhaseCapturedCard(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text     string
		timing   game.DelayedTriggerTiming
		faceDown bool
	}{
		{"Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step.", game.DelayedAtBeginningOfYourNextEndStep, true},
		{"Pay 1 life: Exile the top card of your library. Put that card into your hand at the beginning of your next end step.", game.DelayedAtBeginningOfYourNextEndStep, false},
		{"Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of the next end step.", game.DelayedAtBeginningOfNextEndStep, true},
		{"Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of the next turn's upkeep.", game.DelayedAtBeginningOfNextUpkeep, true},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{Name: "Captured Card", Layout: "normal", TypeLine: "Enchantment", OracleText: test.text})
			sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
			move := sequence[0].Primitive.(game.MoveTopOfLibrary)
			trigger := sequence[1].Primitive.(game.CreateDelayedTrigger).Trigger
			if move.FaceDown != test.faceDown || trigger.Timing != test.timing ||
				!trigger.CapturedCard.Exists || !sequence[0].ClearLinkedBeforeGate {
				t.Fatal("captured card lost its publication, visibility, or phase timing")
			}

		})
	}
}

func TestFixedPhaseSubjectRefusesUnavailableDomains(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Tap target creature. Sacrifice that land at the beginning of the next end step.",
		"Create a 1/1 green Insect creature token. Exile that spell at end of combat.",
		"Whenever you cast a creature spell, exile it at the beginning of the next end step.",
		"Exile this spell at the beginning of the next end step.",
		"Whenever this creature attacks, exile that land at end of combat.",
		"Sacrifice Unavailable Capture at the beginning of the next end step.",
		"Create a 1/1 green Insect creature token. Sacrifice it at the beginning of the next end step if it's an Elf.",
		"Tap target creature. Remove X +1/+1 counters from it at end of combat.",
		"Choose two —\n• Create a 1/1 green Insect creature token. Sacrifice it at end of combat.\n• Create a 1/1 green Insect creature token. Exile it at end of combat.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Unavailable Capture", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
		})
	}
}

func TestFixedPhasePriorTargetCardProduct(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Captured Card", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Exile target creature card from your graveyard. Draw a card. Return that card to the battlefield at the beginning of the next end step.",
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	move := sequence[0].Primitive.(game.MoveCard)
	trigger := sequence[2].Primitive.(game.CreateDelayedTrigger).Trigger
	if move.PublishLinked == "" || !move.ReplacePublishedLinked || !sequence[0].ClearLinkedBeforeGate ||
		!trigger.CapturedCard.Exists || trigger.CapturedCard.Val != game.LinkedObjectReference(string(move.PublishLinked)) {
		t.Fatal("delayed card body did not consume the exact actual card publisher")
	}
}

func TestFixedPhaseTriggeredOptionalTiming(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text   string
		future bool
	}{
		{"Whenever this creature attacks, you may exile it at the beginning of the next end step.", false},
		{"Whenever this creature attacks, at the beginning of the next end step, you may exile it.", true},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{Name: "Captured Source", Layout: "normal", TypeLine: "Creature", OracleText: test.text})
			ability := face.TriggeredAbilities[0]
			sequence := ability.Content.Modes[0].Sequence
			delayed := sequence[0].Primitive.(game.CreateDelayedTrigger).Trigger
			if ability.Optional == test.future || delayed.Optional != test.future ||
				!delayed.CapturedObject.Exists || delayed.CapturedObject.Val != game.SourcePermanentReference() {
				t.Fatal("triggered source choice/capture evaluated at the wrong time")
			}
		})
	}
}

func TestFixedPhaseDepartedEventCardExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Captured Card", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever a creature you control dies, exile that card at the beginning of the next end step.",
	})
	delayed := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CreateDelayedTrigger).Trigger
	body := delayed.Content.Modes[0].Sequence[0].Primitive.(game.MoveCard)
	if !delayed.CapturedCard.Exists || body.Card != game.CapturedCardReference() {
		t.Fatal("departed event card was replaced by a live source/event reference")
	}
}
