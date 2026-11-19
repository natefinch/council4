package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle revision beaefa77-6e4a-4724-a443-fa6b45803db5,
// Oracle ID 138439ad-76c6-4d75-8025-e68b94f68fa3: Breezekeeper ({3}{U},
// Creature — Djinn, 4/4) has flying and phasing. Its Scryfall rulings URL,
// https://api.scryfall.com/cards/beaefa77-6e4a-4724-a443-fa6b45803db5/rulings,
// returned an empty list when retrieved 2026-11-19. Official Comprehensive
// Rules effective November 6, 2026,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 302.1 and 601.2 govern casting/resolving the creature; 702.9a-b define
// flying; 502.1, 502.3 and 702.26a place simultaneous phasing before untapping;
// 702.26b-d define the status changes without changing zones or identity.
// Issue #3999: this test-only fixture casts and resolves Breezekeeper, then
// checks that it phases out and back in on consecutive controller untaps.
func TestBreezekeeperPhasesOutBeforeItsControllerUntaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	breezekeeperDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Breezekeeper",
		ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.U}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Djinn},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		OracleText: "Flying\nPhasing (This phases in or out before you untap during each of your " +
			"untap steps. While it's phased out, it's treated as though it doesn't exist.)",
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.PhasingStaticBody,
		},
	}}
	cardID := addCardToHand(g, game.Player1, breezekeeperDef)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("could not cast Breezekeeper for {3}{U}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	breezekeeper := permanentByCardID(g, cardID)
	if breezekeeper == nil {
		t.Fatal("Breezekeeper did not resolve onto the battlefield")
	}
	objectID := breezekeeper.ObjectID
	setPermanentTapped(g, breezekeeper, true)

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player 1 draw"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player 2 draw"}})
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if breezekeeper.PhasedOut || !breezekeeper.Tapped {
		t.Fatal("Breezekeeper phased out or untapped during its opponent's untap")
	}

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !breezekeeper.PhasedOut || !breezekeeper.Tapped {
		t.Fatalf("Breezekeeper after its controller's first untap: phased out=%v tapped=%v; want both true",
			breezekeeper.PhasedOut, breezekeeper.Tapped)
	}

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player 1 next draw"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	returned := permanentByCardID(g, cardID)
	if returned == nil || returned.ObjectID != objectID || returned.PhasedOut || returned.Tapped {
		t.Fatalf("Breezekeeper did not phase in and untap as the same permanent (object %v)", objectID)
	}
}
