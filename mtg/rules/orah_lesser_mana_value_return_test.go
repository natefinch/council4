package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lesserManaValueClericDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Cleric},
	}}
}

// TestReturnTargetRequiresLesserManaValueThanEventPermanent covers the
// event-relative "with lesser mana value" graveyard-return target filter that
// Orah, Skyclave Hierophant's dies trigger uses: Selection.
// ManaValueLessThanEventPermanent requires the candidate graveyard card's mana
// value to be strictly less than the mana value of the permanent named by the
// triggering event (the Cleric that died). A died Cleric of mana value 3 lets the
// return target a lesser-mana-value Cleric card (2) but not one of equal (3) or
// greater (4) mana value, the subtype filter still excludes a lesser non-Cleric,
// and without a triggering event permanent the bound fails closed.
func TestReturnTargetRequiresLesserManaValueThanEventPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	diedCleric := addCardInstance(g, game.Player1, lesserManaValueClericDef("Died Cleric", 3))
	lesser := addCardInstance(g, game.Player1, lesserManaValueClericDef("Lesser Cleric", 2))
	equal := addCardInstance(g, game.Player1, lesserManaValueClericDef("Equal Cleric", 3))
	greater := addCardInstance(g, game.Player1, lesserManaValueClericDef("Greater Cleric", 4))
	lesserNonCleric := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Lesser Soldier",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Soldier},
	}})
	for _, cardID := range []game.ObjectID{diedCleric, lesser, equal, greater, lesserNonCleric} {
		g.Players[game.Player1].Graveyard.Add(cardID)
	}

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			SubtypesAny:                     []types.Sub{types.Cleric},
			Controller:                      game.ControllerYou,
			ManaValueLessThanEventPermanent: true,
		}),
	}

	// The died Cleric (mana value 3) is the permanent named by the trigger event.
	event := game.Event{Kind: game.EventPermanentDied, CardID: diedCleric, PermanentID: g.IDGen.Next()}

	if !targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(lesser)) {
		t.Fatal("a Cleric card with lesser mana value than the died creature should be a legal target")
	}
	if targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(equal)) {
		t.Fatal("a Cleric card with equal mana value must not match a lesser-mana-value filter")
	}
	if targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(greater)) {
		t.Fatal("a Cleric card with greater mana value must not match a lesser-mana-value filter")
	}
	if targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(lesserNonCleric)) {
		t.Fatal("a lesser-mana-value non-Cleric card must not match the Cleric subtype filter")
	}
	if targetMatchesSpec(g, game.Player1, 0, game.Event{}, &spec, game.CardTarget(lesser)) {
		t.Fatal("without a triggering event permanent the mana-value bound must fail closed")
	}
}
