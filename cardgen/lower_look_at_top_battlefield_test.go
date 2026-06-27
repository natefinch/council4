package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// intoTheWildsOracleText is the complete printed Into the Wilds rules text: an
// upkeep trigger whose body is the conditional look-at-top battlefield sequence
// with no trailing else clause, so a declined or non-matching card stays atop
// the library.
const intoTheWildsOracleText = "At the beginning of your upkeep, look at the top card of your library. " +
	"If it's a land card, you may put it onto the battlefield."

// risenReefStyleOracleText is the conditional look-at-top battlefield sequence in
// its tapped-with-hand-fallback form, proving the building block carries both the
// EntryTapped flag and the hand else disposition independently of the trigger.
const risenReefStyleOracleText = "Whenever this creature attacks, look at the top card of your library. " +
	"If it's a land card, you may put it onto the battlefield tapped. " +
	"If you don't put the card onto the battlefield, put it into your hand."

// TestLowerConditionalLookAtTopBattlefieldNoElse proves the no-else form lowers to
// a look followed by a conditional destination that puts the matching card onto
// the battlefield (Then = zone.None) with no fallback (Else = zone.None).
func TestLowerConditionalLookAtTopBattlefieldNoElse(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Into the Wilds",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{3}{G}",
		OracleText: intoTheWildsOracleText,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one look-at-top sequence", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want at-the-beginning", trigger.Trigger.Type)
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
		t.Fatalf("sequence[1] = %#v, want ConditionalDestinationPlace of the looked-at card", sequence[1].Primitive)
	}
	if place.Then != zone.None {
		t.Fatalf("conditional destination then = %v, want the implicit battlefield (zone.None)", place.Then)
	}
	if place.EntryTapped {
		t.Fatal("conditional destination should not enter tapped without the word \"tapped\"")
	}
	if place.Else != zone.None {
		t.Fatalf("conditional destination else = %v, want no fallback (zone.None)", place.Else)
	}
	condition := place.CardCondition
	if !condition.Exists ||
		len(condition.Val.Selection.RequiredTypesAny) != 1 ||
		condition.Val.Selection.RequiredTypesAny[0] != types.Land {
		t.Fatalf("conditional destination card condition = %#v, want a land card", place.CardCondition)
	}
}

// TestLowerConditionalLookAtTopBattlefieldTappedHandFallback proves the tapped
// hand-fallback form carries EntryTapped and an Else of zone.Hand, and that the
// recognizer is trigger-agnostic (an attacks trigger here).
func TestLowerConditionalLookAtTopBattlefieldTappedHandFallback(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reef Mimic",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		ManaCost:   "{1}{G}{U}",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: risenReefStyleOracleText,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one look-at-top sequence", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want look/conditional-destination", len(sequence))
	}
	place, ok := sequence[1].Primitive.(game.ConditionalDestinationPlace)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want ConditionalDestinationPlace", sequence[1].Primitive)
	}
	if place.Then != zone.None || !place.EntryTapped {
		t.Fatalf("conditional destination = %#v, want a tapped battlefield put", place)
	}
	if place.Else != zone.Hand {
		t.Fatalf("conditional destination else = %v, want a hand fallback", place.Else)
	}
}
