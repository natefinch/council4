package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerTauntFromTheRampart(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Taunt from the Rampart",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{3}{R}{W}",
		OracleText: "Goad all creatures your opponents control. Until your next turn, those creatures can't block. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want goad then can't block", mode.Sequence)
	}

	goad, ok := mode.Sequence[0].Primitive.(game.Goad)
	if !ok || !goad.Group.Valid() {
		t.Fatalf("goad = %#v", mode.Sequence[0].Primitive)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyRule)
	if !ok ||
		len(apply.RuleEffects) != 1 ||
		!apply.RuleEffects[0].AffectedSelection.MatchGoaded ||
		apply.Duration != game.DurationUntilYourNextTurn {
		t.Fatalf("can't-block rule = %#v", mode.Sequence[1].Primitive)
	}
}

func TestLowerGoadAllOpponentCreatures(t *testing.T) {
	t.Parallel()
	lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Goad",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Goad all creatures your opponents control.",
	})
}
