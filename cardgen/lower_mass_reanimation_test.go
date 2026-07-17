package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerBreachTheMultiverse proves the mass reanimation base lowers Breach the
// Multiverse end to end: a group mill fills every graveyard, one chooser picks a
// creature or planeswalker card in each player's graveyard, the chosen cards
// enter the battlefield at once under the controller, and the controlled-creature
// type grant permanently adds Phyrexian.
func TestLowerBreachTheMultiverse(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Breach the Multiverse",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Each player mills ten cards. " +
			"For each player, choose a creature or planeswalker card in that player's graveyard. " +
			"Put those cards onto the battlefield under your control. " +
			"Then each creature you control becomes a Phyrexian in addition to its other types.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 4 {
		t.Fatalf("mode = %+v, want no targets and four instructions", mode)
	}
	mill, ok := mode.Sequence[0].Primitive.(game.Mill)
	if !ok || mill.PlayerGroup != game.AllPlayersReference() || mill.Amount.Value() != 10 {
		t.Fatalf("first primitive = %+v, want each player mills ten", mode.Sequence[0].Primitive)
	}
	choose, ok := mode.Sequence[1].Primitive.(game.ChooseCardFromEachGraveyard)
	if !ok {
		t.Fatalf("second primitive = %+v, want ChooseCardFromEachGraveyard", mode.Sequence[1].Primitive)
	}
	if choose.Chooser != game.ControllerReference() ||
		choose.Players != game.AllPlayersReference() ||
		choose.Optional ||
		choose.LinkedKey != massReanimationChosenKey {
		t.Fatalf("choose = %+v, want mandatory controller choice over all players", choose)
	}
	if got := choose.Selection.RequiredTypesAny; len(got) != 2 ||
		!slices.Contains(got, types.Creature) || !slices.Contains(got, types.Planeswalker) {
		t.Fatalf("choose selection types = %+v, want creature or planeswalker", got)
	}
	reanimate, ok := mode.Sequence[2].Primitive.(game.ReanimateLinkedCards)
	if !ok || reanimate.Controller != game.ControllerReference() ||
		reanimate.LinkedKey != massReanimationChosenKey {
		t.Fatalf("third primitive = %+v, want reanimate under controller", mode.Sequence[2].Primitive)
	}
	apply, ok := mode.Sequence[3].Primitive.(game.ApplyContinuous)
	if !ok || apply.Duration != game.DurationPermanent {
		t.Fatalf("fourth primitive = %+v, want permanent ApplyContinuous", mode.Sequence[3].Primitive)
	}
	if len(apply.ContinuousEffects) == 0 || apply.ContinuousEffects[0].Layer != game.LayerType {
		t.Fatalf("apply continuous effects = %+v, want a type layer", apply.ContinuousEffects)
	}
	if len(apply.ContinuousEffects[0].AddSubtypes) != 1 || apply.ContinuousEffects[0].AddSubtypes[0] != types.Phyrexian {
		t.Fatalf("apply add subtypes = %+v, want Phyrexian", apply.ContinuousEffects[0].AddSubtypes)
	}
}

// TestLowerMassReanimationUpToOneNoMillNoRider proves the optional "up to one"
// variant without a leading mill or trailing rider lowers to just the per-player
// choice and the simultaneous reanimation, with the choice marked optional.
func TestLowerMassReanimationUpToOneNoMillNoRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Optional Reanimation",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "For each player, choose up to one creature card in that player's graveyard. " +
			"Put those cards onto the battlefield under your control.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.ChooseCardFromEachGraveyard)
	if !ok || !choose.Optional || choose.LinkedKey != massReanimationChosenKey {
		t.Fatalf("first primitive = %+v, want optional per-player choice", mode.Sequence[0].Primitive)
	}
	if got := choose.Selection.RequiredTypes; len(got) != 1 || !slices.Contains(got, types.Creature) {
		t.Fatalf("choose selection required types = %+v, want creature", got)
	}
	reanimate, ok := mode.Sequence[1].Primitive.(game.ReanimateLinkedCards)
	if !ok || reanimate.LinkedKey != massReanimationChosenKey {
		t.Fatalf("second primitive = %+v, want reanimate", mode.Sequence[1].Primitive)
	}
}

// TestLowerMassReanimationFailsClosedOnTarget proves a per-player graveyard
// choice that also carries a target does not lower to a partial reanimation.
func TestLowerMassReanimationFailsClosedOnTarget(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Test Bad Reanimation",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Destroy target creature. " +
			"For each player, choose a creature card in that player's graveyard. " +
			"Put those cards onto the battlefield under your control.",
	})
}

// TestLowerMassReanimationFailsClosedOnTargetedChoice proves the targeted
// per-player graveyard choice "choose up to one target creature card" (Afterlife
// from the Loam, The Moonbase) fails closed rather than routing through the
// non-targeted primitive and silently dropping the targeting.
func TestLowerMassReanimationFailsClosedOnTargetedChoice(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Test Targeted Reanimation",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "For each player, choose up to one target creature card in that player's graveyard. " +
			"Put those cards onto the battlefield under your control.",
	})
}

// TestLowerMassReanimationFailsClosedWithoutPut proves a per-player graveyard
// choice with no following "Put those cards..." clause fails closed instead of
// indexing past the effects. A leading each-player mill puts the choose last in
// the sequence, which must not panic.
func TestLowerMassReanimationFailsClosedWithoutPut(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Test Choose Without Put",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Each player mills ten cards. " +
			"For each player, choose a creature card in that player's graveyard.",
	})
}
