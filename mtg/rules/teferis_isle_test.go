package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: exact pinned Oracle ID ce55657d-d82f-4528-a83e-5cad7de111fd,
// printing b6ed7ca8-fd91-46e3-9149-a3de23c7078e (Teferi's Isle, legendary land):
// "Phasing (This phases in or out before you untap during each of your untap
// steps. While it's phased out, it's treated as though it doesn't exist.)
// Teferi's Isle enters tapped. {T}: Add {U}{U}."
// Its printing-specific Scryfall rulings URI,
// https://api.scryfall.com/cards/b6ed7ca8-fd91-46e3-9149-a3de23c7078e/rulings,
// returned an empty list on 2026-11-20.
// Official Comprehensive Rules effective 2026-09-25 (the corpus pin), downloaded
// from https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// and verified against SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca:
// 305.1 permits the land play as a special action; 614.1c and 614.12 make its
// enters-tapped text a replacement applied using its would-be battlefield
// characteristics; 502.1, 502.3, and 702.26a-d require phasing before untap
// without changing zones or object identity; 106.12 and 602.2b define the
// tap-for-mana activation cost; 605.1a and 605.3b classify the untargeted mana
// ability and require it to resolve immediately.
func TestTeferisIsleEntersTappedPhasesAndAddsTwoBlue(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID("ce55657d-d82f-4528-a83e-5cad7de111fd")
	if err != nil {
		t.Fatalf("exact Teferi's Isle registry lookup failed: %v", err)
	}
	if def.Name != "Teferi's Isle" {
		t.Fatalf("exact Teferi's Isle lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.PlayLand(cardID)) {
		t.Fatal("playing Teferi's Isle from hand failed")
	}
	isle := permanentByCardID(g, cardID)
	if isle == nil || !isle.Tapped {
		t.Fatal("Teferi's Isle did not enter the battlefield tapped")
	}
	if !hasKeyword(g, isle, game.Phasing) {
		t.Fatal("Teferi's Isle did not have its printed phasing ability")
	}

	for player := game.Player1; player < game.NumPlayers; player++ {
		for i := range 2 {
			addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{
				Name: fmt.Sprintf("Teferi's Isle library card %d for player %d", i, player),
			}})
		}
	}

	untaps := []struct {
		player     game.PlayerID
		phasedOut  bool
		isleTapped bool
	}{
		{player: game.Player2, phasedOut: false, isleTapped: true},
		{player: game.Player1, phasedOut: true, isleTapped: true},
		{player: game.Player2, phasedOut: true, isleTapped: true},
		{player: game.Player1, phasedOut: false, isleTapped: false},
	}
	for _, untap := range untaps {
		g.Turn.ActivePlayer = untap.player
		g.Turn.PriorityPlayer = untap.player
		engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if isle.PhasedOut != untap.phasedOut || isle.Tapped != untap.isleTapped {
			t.Fatalf("after player %d's untap, Teferi's Isle state = phased out %t, tapped %t; want %t, %t",
				untap.player, isle.PhasedOut, isle.Tapped, untap.phasedOut, untap.isleTapped)
		}
	}

	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(isle.ObjectID, 0, nil, 0)) {
		t.Fatal("Teferi's Isle mana ability failed to activate")
	}
	if !isle.Tapped {
		t.Fatal("activating Teferi's Isle's mana ability did not pay its tap cost")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 2 {
		t.Fatalf("Teferi's Isle added %d blue mana, want 2", got)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 2 {
		t.Fatalf("Teferi's Isle added %d total mana, want only two blue", got)
	}
}
