package rules

import (
	"testing"

	carde "github.com/natefinch/council4/mtg/cards/e"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestEldritchEvolutionSacrificedCostBoundsSearchEndToEnd casts the real
// Eldritch Evolution card through the full cast pipeline: it sacrifices a
// mana-value-3 creature to pay the additional cost, then resolves and searches
// for "a creature card with mana value X or less, where X is 2 plus the
// sacrificed creature's mana value" (so mana value 5 or less here). A
// mana-value-5 creature is a legal find and enters the battlefield; a
// mana-value-6 creature is not a legal choice.
//
// Casting end to end is essential: the search bound reads the sacrificed
// creature's last-known mana value through obj.SacrificedAsCostIDs, which only
// the real cast path populates. A resolution-only test that hand-injects
// SacrificedAsCostIDs would pass even when the cast path forgot to record the
// sacrifice, silently degrading the bound to the bare addend (mana value 2).
func TestEldritchEvolutionSacrificedCostBoundsSearchEndToEnd(t *testing.T) {
	// castEldritchEvolution sets up a game whose only creature is a mana-value-3
	// creature, casts Eldritch Evolution (which sacrifices that creature to pay
	// its additional cost), resolves the search with an agent wanting the named
	// card, and returns the resolved game plus the in-bound (mana value 5) and
	// out-of-bound (mana value 6) library creatures.
	castEldritchEvolution := func(t *testing.T, wanted string) (*game.Game, id.ID, id.ID) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, carde.EldritchEvolution())
		// The lone creature (mana value 3) is the only legal sacrifice, so the
		// additional cost is forced onto it and the search bound is deterministic.
		sacrificed := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Sacrificed Bear",
			ManaCost: opt.Val(cost.Mana{cost.O(3)}),
			Types:    []types.Card{types.Creature},
		}})
		// {1}{G}{G} is paid from Forests; the sacrifice targets a creature, so land
		// mana and the sacrifice cost never compete for the same permanents.
		for range 3 {
			addBasicLandPermanent(g, game.Player1, types.Forest)
		}
		inBound := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Five Bear",
			ManaCost: opt.Val(cost.Mana{cost.O(5)}),
			Types:    []types.Card{types.Creature},
		}})
		outOfBound := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Six Bear",
			ManaCost: opt.Val(cost.Mana{cost.O(6)}),
			Types:    []types.Card{types.Creature},
		}})
		setSorcerySpeedTurn(g, game.Player1)

		act := action.CastSpell(spellID, nil, 0, nil)
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatal("applyAction(cast Eldritch Evolution) = false, want true")
		}
		obj, ok := g.Stack.Peek()
		if !ok {
			t.Fatal("expected Eldritch Evolution on the stack after casting")
		}
		if len(obj.SacrificedAsCostIDs) != 1 || obj.SacrificedAsCostIDs[0] != sacrificed.ObjectID {
			t.Fatalf("stack object SacrificedAsCostIDs = %v, want [%v] (the sacrificed creature)", obj.SacrificedAsCostIDs, sacrificed.ObjectID)
		}

		agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: wanted}}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		return g, inBound, outOfBound
	}

	t.Run("mana-value-5 creature is a legal find", func(t *testing.T) {
		g, inBound, outOfBound := castEldritchEvolution(t, "Five Bear")
		if permanentForCard(g, inBound) == nil {
			t.Fatal("a mana-value-5 creature within the sacrificed-cost bound did not enter the battlefield")
		}
		if !g.Players[game.Player1].Library.Contains(outOfBound) {
			t.Fatal("a mana-value-6 creature above the sacrificed-cost bound was incorrectly findable")
		}
	})

	t.Run("mana-value-6 creature is not a legal choice", func(t *testing.T) {
		g, _, outOfBound := castEldritchEvolution(t, "Six Bear")
		if permanentForCard(g, outOfBound) != nil {
			t.Fatal("a mana-value-6 creature above the sacrificed-cost bound was incorrectly put onto the battlefield")
		}
		if !g.Players[game.Player1].Library.Contains(outOfBound) {
			t.Fatal("a mana-value-6 creature should remain in the library because it exceeds the bound")
		}
	})
}
