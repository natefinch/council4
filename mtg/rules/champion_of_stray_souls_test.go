package rules

import (
	"testing"

	cardsc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// championGraveyardAbilityIndex is the canonical body index of Champion of Stray
// Souls' second ability ("{5}{B}{B}: Put this card from your graveyard on top of
// your library."). Champion has no spell ability, so the two activated abilities
// occupy indices 0 (battlefield reanimation) and 1 (graveyard recursion).
const championGraveyardAbilityIndex = 1

// TestChampionOfStraySoulsRecursionFunctionsFromGraveyard proves the corrected
// zone inference: Champion's "{5}{B}{B}: Put this card from your graveyard on top
// of your library." ability functions from the GRAVEYARD, moving the card from the
// graveyard onto the top of its owner's library. Before the fix it was mis-modeled
// as a battlefield self-tuck, so the intended graveyard recursion never happened.
func TestChampionOfStraySoulsRecursionFunctionsFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardID := addCardToGraveyard(g, game.Player1, cardsc.ChampionOfStraySouls())
	// {5}{B}{B} is seven mana; seven Swamps cover the generic and black pips.
	for range 7 {
		addBasicLandPermanent(g, game.Player1, types.Swamp)
	}
	setMainPhasePriority(g, game.Player1)

	act := action.ActivateAbility(cardID, championGraveyardAbilityIndex, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard-recursion ability was not offered while Champion was in the graveyard")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating the graveyard-recursion ability failed")
	}
	engine.resolveTopOfStack(g, nil)

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Champion remained in the graveyard after resolving its recursion ability")
	}
	top, ok := g.Players[game.Player1].Library.Top()
	if !ok || top != cardID {
		t.Fatalf("top of library = %v (ok=%v), want Champion %v moved onto the library", top, ok, cardID)
	}
}

// TestChampionOfStraySoulsRecursionNotOfferedOnBattlefield proves the ability is
// scoped to the graveyard: while Champion is a permanent on the battlefield, the
// "put this card from your graveyard on top of your library" ability is NOT a
// legal play. Before the fix its ZoneOfFunction was Battlefield, which illegally
// let a player tuck their own on-battlefield Champion onto their library.
func TestChampionOfStraySoulsRecursionNotOfferedOnBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	champion := addCombatPermanent(g, game.Player1, cardsc.ChampionOfStraySouls())
	for range 7 {
		addBasicLandPermanent(g, game.Player1, types.Swamp)
	}
	setMainPhasePriority(g, game.Player1)

	act := action.ActivateAbility(champion.ObjectID, championGraveyardAbilityIndex, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard-recursion ability was illegally offered while Champion was on the battlefield")
	}
}
