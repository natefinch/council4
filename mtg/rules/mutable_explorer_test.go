package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: ECL printing 8f35d95a-caea-4d5e-b98e-55da1ba7c92d;
// https://api.scryfall.com/cards/8f35d95a-caea-4d5e-b98e-55da1ba7c92d/rulings
// returned five WotC rulings dated 2025-11-17. The token is a copy of the card
// Mutavault ("{T}: Add {C}." and "{1}: becomes a 2/2 creature with all creature
// types until end of turn. It's still a land."), and changeling makes the card
// every creature type in all zones. CR 2026-09-25 from the verified official
// copy governs tokens (111.1-111.4), enters-the-battlefield triggers (603.6a),
// mana abilities (605.1a) and type-changing effects (205.1b, 613.1d). The card
// is an installed-compiled definition (oracle 34b93a66-de1e-4133-a479-e0d90ba94a3a)
// registered by name only.

func TestMutableExplorerCreatesTappedMutavaultTokenWithItsAbilities(t *testing.T) {
	def := cards.NewDefaultRegistry().Lookup("Mutable Explorer")
	if def == nil {
		t.Fatal("production Mutable Explorer is missing")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	explorerID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	if !e.applyAction(g, game.Player1, action.CastSpell(explorerID, nil, 0, nil)) {
		t.Fatal("Mutable Explorer cast failed")
	}
	e.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	explorer := g.Battlefield[len(g.Battlefield)-1]
	if explorer.CardInstanceID != explorerID || !permanentHasSubtype(g, explorer, types.Elf) {
		t.Fatal("Mutable Explorer did not enter as a changeling creature")
	}
	e.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %d, want Mutable Explorer's enters trigger", g.Stack.Size())
	}
	e.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	token := g.Battlefield[len(g.Battlefield)-1]
	if !token.Token || permanentEffectiveName(g, token) != "Mutavault" || !token.Tapped {
		t.Fatalf("created %q token=%v tapped=%v, want a tapped Mutavault token",
			permanentEffectiveName(g, token), token.Token, token.Tapped)
	}
	if !permanentHasType(g, token, types.Land) || permanentHasType(g, token, types.Creature) {
		t.Fatal("Mutavault token is not just a land before its ability resolves")
	}

	token.Tapped = false
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 0 {
		t.Fatalf("pool {C} = %d before tapping the token, want Mutable Explorer's cost paid", got)
	}
	setMainPhasePriority(g, game.Player1)
	manaIdx, animateIdx := -1, -1
	for idx, ability := range permanentEffectiveAbilitiesView(g, token) {
		switch ability.(type) {
		case *game.ManaAbility:
			manaIdx = idx
		case *game.ActivatedAbility:
			animateIdx = idx
		}
	}
	if manaIdx < 0 || animateIdx < 0 {
		t.Fatalf("token abilities: mana %d, animate %d", manaIdx, animateIdx)
	}
	if !e.applyAction(g, game.Player1, action.ActivateAbility(token.ObjectID, manaIdx, nil, 0)) {
		t.Fatal("token's {T}: Add {C} could not be activated")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 1 || !token.Tapped {
		t.Fatalf("pool {C} = %d, tapped = %v after the token's mana ability", got, token.Tapped)
	}
	if !e.applyAction(g, game.Player1, action.ActivateAbility(token.ObjectID, animateIdx, nil, 0)) {
		t.Fatal("token's {1} animate ability could not be activated with its own {C}")
	}
	e.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if !permanentHasType(g, token, types.Creature) || !permanentHasType(g, token, types.Land) ||
		!permanentHasSubtype(g, token, types.Goblin) {
		t.Fatal("animated Mutavault token is not a land creature with all creature types")
	}
	if toughness, _ := effectiveToughness(g, token); effectivePower(g, token) != 2 || toughness != 2 {
		t.Fatalf("animated Mutavault token = %d/%d, want 2/2", effectivePower(g, token), toughness)
	}
}
