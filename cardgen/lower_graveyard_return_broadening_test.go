package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerGraveyardReturnSerialTypeUnion covers three-or-more card-type unions,
// which render with the serial-comma "X, Y, or Z" form. This is the wording on
// Treasury Thrull and Possessed Skaab, which previously failed closed because the
// parser only reconstructed two-member " or " joins.
func TestLowerGraveyardReturnSerialTypeUnion(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Serial Union",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target artifact, creature, or enchantment card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!slices.Equal(target.Selection.Val.RequiredTypesAny, []types.Card{types.Artifact, types.Creature, types.Enchantment}) {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("move = %#v", move)
	}
}

// TestLowerGraveyardReturnPluralAndOrUnion covers plural multi-target type unions,
// which join with "and/or" instead of "or". This is the wording on Scholar of
// the Ages.
func TestLowerGraveyardReturnPluralAndOrUnion(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test And Or Union",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to two target instant and/or sorcery cards from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.MaxTargets != 2 || target.TargetZone != zone.Graveyard ||
		!slices.Equal(target.Selection.Val.RequiredTypesAny, []types.Card{types.Instant, types.Sorcery}) {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Destination != zone.Hand {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerGraveyardReturnOneOrTwoCardinality covers the "one or two target"
// count, the wording on Infernal Rebirth and Archenemy's Charm.
func TestLowerGraveyardReturnOneOrTwoCardinality(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test One Or Two",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return one or two target creature cards from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.MinTargets != 1 || target.MaxTargets != 2 || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
}

// TestLowerGraveyardReturnOwnerRelativeHand covers owner-relative hand
// destinations. A returned card always moves to its owner's hand, so "to its
// owner's hand" and "to their hand" lower identically to "to your hand". This is
// the wording on Pulse of Murasa and Forcemage Advocate.
func TestLowerGraveyardReturnOwnerRelativeHand(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Return target creature card from a graveyard to its owner's hand.",
		"Return target card from an opponent's graveyard to their hand.",
	} {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Owner Hand",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracle,
		})
		mode := face.SpellAbility.Val.Modes[0]
		move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
		if !ok {
			t.Fatalf("oracle %q primitive = %T, want game.MoveCard", oracle, mode.Sequence[0].Primitive)
		}
		if move.Card.Kind != game.CardReferenceTarget || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
			t.Fatalf("oracle %q move = %#v", oracle, move)
		}
	}
}
