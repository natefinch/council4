package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type typeMatchState struct {
	fakePaymentState

	landTypes map[id.ID]bool
}

func (s typeMatchState) PermanentHasType(permanent *game.Permanent, cardType types.Card) bool {
	return cardType == types.Land && s.landTypes[permanent.ObjectID]
}

func (s typeMatchState) PermanentMatchesSelection(permanent *game.Permanent, sel game.Selection) bool {
	if len(sel.RequiredTypesAny) > 0 && !slices.ContainsFunc(sel.RequiredTypesAny, func(t types.Card) bool {
		return s.PermanentHasType(permanent, t)
	}) {
		return false
	}
	return !slices.ContainsFunc(sel.ExcludedTypes, func(t types.Card) bool {
		return s.PermanentHasType(permanent, t)
	})
}

// TestAdditionalCostExcludesPermanentType verifies that the
// ExcludePermanentType filter (Bolas's Citadel's "Sacrifice ten nonland
// permanents") bars permanents of the excluded type from satisfying the cost
// while still admitting every other permanent.
func TestAdditionalCostExcludesPermanentType(t *testing.T) {
	land := &game.Permanent{ObjectID: 1}
	creature := &game.Permanent{ObjectID: 2}
	state := typeMatchState{landTypes: map[id.ID]bool{1: true}}
	additional := cost.Additional{
		Kind:                 cost.AdditionalSacrifice,
		ExcludePermanentType: types.Land,
	}
	if additionalCostMatchesPermanent(state, land, additional) {
		t.Fatal("a land matched a nonland-permanent sacrifice cost")
	}
	if !additionalCostMatchesPermanent(state, creature, additional) {
		t.Fatal("a creature did not match a nonland-permanent sacrifice cost")
	}
}

// TestChooseSacrificePermanentsSkipsExcludedType proves the sacrifice planner
// never selects a permanent of the excluded type.
func TestChooseSacrificePermanentsSkipsExcludedType(t *testing.T) {
	land := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	creature := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := typeMatchState{
		fakePaymentState: fakePaymentState{battlefield: []*game.Permanent{land, creature}},
		landTypes:        map[id.ID]bool{1: true},
	}
	additional := cost.Additional{
		Kind:                 cost.AdditionalSacrifice,
		Amount:               1,
		ExcludePermanentType: types.Land,
	}
	chosen := chooseSacrificePermanents(state, game.Player1, additional, 1, nil, nil)
	if len(chosen) != 1 || chosen[0].ObjectID != creature.ObjectID {
		t.Fatalf("chosen = %#v, want only the nonland permanent", chosen)
	}

	landOnly := typeMatchState{
		fakePaymentState: fakePaymentState{battlefield: []*game.Permanent{land}},
		landTypes:        map[id.ID]bool{1: true},
	}
	if chosen := chooseSacrificePermanents(landOnly, game.Player1, additional, 1, nil, nil); len(chosen) != 0 {
		t.Fatalf("chosen = %#v, want no permanent when only a land is present", chosen)
	}
}
