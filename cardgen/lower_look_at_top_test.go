package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// sarinthSteelseekerOracleText is the complete printed Sarinth Steelseeker rules
// text: an artifact-enters trigger whose body is the conditional look-at-top
// reveal sequence guarded on a land card.
const sarinthSteelseekerOracleText = "Whenever an artifact you control enters, look at the top card of your library. " +
	"If it's a land card, you may reveal it and put it into your hand. " +
	"If you don't put the card into your hand, you may put it into your graveyard."

// travelingBotanistOracleText is the complete printed Traveling Botanist rules
// text: the same conditional look-at-top reveal sequence on a becomes-tapped
// trigger, proving the building block is trigger-agnostic.
const travelingBotanistOracleText = "Whenever this creature becomes tapped, look at the top card of your library. " +
	"If it's a land card, you may reveal it and put it into your hand. " +
	"If you don't put the card into your hand, you may put it into your graveyard."

func TestLowerConditionalLookAtTopReveal(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sarinth Steelseeker",
		Layout:     "normal",
		TypeLine:   "Creature — Human Artificer Scout",
		ManaCost:   "{2}{G}",
		Power:      new("2"),
		Toughness:  new("3"),
		OracleText: sarinthSteelseekerOracleText,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one look-at-top sequence", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhenever {
		t.Fatalf("trigger type = %v, want whenever", trigger.Trigger.Type)
	}
	if trigger.Optional {
		t.Fatal("trigger should not be optional; the optional choices are on the reveal and graveyard move")
	}
	if len(trigger.Content.Modes) != 1 {
		t.Fatalf("trigger modes = %d, want one", len(trigger.Content.Modes))
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want look/conditional-destination", len(sequence))
	}

	look, ok := sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.PublishLinked == "" {
		t.Fatalf("sequence[0] = %#v, want LookAtLibraryTop with a published link", sequence[0].Primitive)
	}

	place, ok := sequence[1].Primitive.(game.ConditionalDestinationPlace)
	if !ok ||
		place.Card.Kind != game.CardReferenceLinked || place.Card.LinkID != string(look.PublishLinked) ||
		place.FromZone != zone.Library {
		t.Fatalf("sequence[1] = %#v, want ConditionalDestinationPlace of the looked-at library card", sequence[1].Primitive)
	}
	if place.Then != zone.Hand || !place.ThenReveal {
		t.Fatalf("conditional destination then = %#v, want a revealing put into hand", place)
	}
	if place.Else != zone.Graveyard || !place.ElseOptional {
		t.Fatalf("conditional destination else = %#v, want an optional graveyard fallback", place)
	}
	condition := place.CardCondition
	if !condition.Exists ||
		len(condition.Val.Selection.RequiredTypesAny) != 1 ||
		condition.Val.Selection.RequiredTypesAny[0] != types.Land {
		t.Fatalf("conditional destination card condition = %#v, want a land card", place.CardCondition)
	}
}

// TestLowerConditionalLookAtTopRevealTriggerAgnostic proves the building block
// lowers identically under a different trigger, here a becomes-tapped trigger,
// confirming the recognizer and lowering own the body text-blind regardless of
// the surrounding trigger.
func TestLowerConditionalLookAtTopRevealTriggerAgnostic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Traveling Botanist",
		Layout:     "normal",
		TypeLine:   "Creature — Dog Scout",
		ManaCost:   "{2}{G}",
		Power:      new("1"),
		Toughness:  new("4"),
		OracleText: travelingBotanistOracleText,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one look-at-top sequence", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want look/conditional-destination", len(sequence))
	}
}
