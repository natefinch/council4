package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned corpus printing revision beaefa77-6e4a-4724-a443-fa6b45803db5,
// Oracle ID 138439ad-76c6-4d75-8025-e68b94f68fa3: Breezekeeper costs {3}{U},
// is a 4/4 blue Djinn, and has flying and phasing. Its Scryfall rulings URI,
// https://api.scryfall.com/cards/beaefa77-6e4a-4724-a443-fa6b45803db5/rulings,
// returned no rulings when retrieved 2026-11-19. Official Comprehensive Rules,
// effective 2026-09-25 as pinned by testdata/card-preservation/corpus.json,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx:
// 302.1-2 and 601.2 govern casting and resolving the creature; 702.9a-b define
// flying; 502.1 and 502.3 require phasing before untapping; 702.26a-d define
// simultaneous phasing, its status changes, and that it doesn't change zones or
// control. This scenario casts and resolves the registered card for {3}{U},
// then verifies it phases out before its controller untaps and phases back in
// as the same permanent at the next untap.
func TestBreezekeeperPhasesOutBeforeItsControllerUntaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, authored.Breezekeeper.New())
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("could not cast Breezekeeper for {3}{U}")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatal("casting Breezekeeper did not spend its full {3}{U} cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	breezekeeper := permanentByCardID(g, cardID)
	if breezekeeper == nil {
		t.Fatal("Breezekeeper did not resolve onto the battlefield")
	}
	if !hasKeyword(g, breezekeeper, game.Flying) {
		t.Fatal("Breezekeeper did not have flying after resolving")
	}
	objectID := breezekeeper.ObjectID
	setPermanentTapped(g, breezekeeper, true)

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player 1 draw"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player 1 next draw"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player 2 draw"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player 2 next draw"}})
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

	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !breezekeeper.PhasedOut || !breezekeeper.Tapped {
		t.Fatal("Breezekeeper changed status or untapped during the opponent's next untap")
	}

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	returned := permanentByCardID(g, cardID)
	if returned == nil || returned.ObjectID != objectID || returned.PhasedOut || returned.Tapped {
		t.Fatalf("Breezekeeper did not phase in and untap as the same permanent at its next untap (object %v)", objectID)
	}
}
