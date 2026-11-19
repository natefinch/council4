package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Oracle ID 469f25de-c180-47e9-9b1b-e23269139bbd, pinned
// printing revision 8ab5152d-767a-47d5-acfc-f439810aaeea. Rulings URI
// https://api.scryfall.com/cards/8ab5152d-767a-47d5-acfc-f439810aaeea/rulings,
// retrieved 2026-11-19: five 2022-04-29 rulings confirm normal casting and
// timing, sacrifice only if the original permanent remains on the battlefield,
// no required attack, copies don't inherit Blitz, and draw on any death.
// Official CR effective 2026-09-25:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (verified local SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 111.1-4/10a, 118.9, 121.1, 302.6, 513.1, 601.2b/f-h, 603.2/3/7a-d,
// 700.4, 701.21a, 702.10b and 702.152a-b: Blitz is an alternative cost;
// when paid it grants haste and a dies-draw trigger and schedules one
// next-end-step sacrifice, while this card's own death trigger creates a
// Treasure. "Dies" means battlefield-to-graveyard, and copies do not inherit
// the paid cost.
func TestRiveteersRequisitionerBlitzDiesCreatesTreasureAndDraws(t *testing.T) {
	engine := NewEngine(nil)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	draw := addCardToLibraryNamed(g, game.Player1, "Riveteers Requisitioner draw")
	card := addCardToHand(g, game.Player1, authored.RiveteersRequisitioner.New())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), blitzAgents("Blitz"), &TurnLog{}) {
		t.Fatal("could not cast Riveteers Requisitioner for its Blitz cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !spell.Blitzed || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Blitz cost was not paid and retained: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, card)
	if permanent == nil || !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("Riveteers Requisitioner did not enter with Blitz-granted haste")
	}

	blitzEndStep(engine, g)
	if g.Stack.Size() != 1 {
		t.Fatalf("next end step has %d pending abilities, want the Blitz sacrifice", g.Stack.Size())
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Graveyard.Contains(card) {
		t.Fatal("Blitz did not sacrifice Riveteers Requisitioner at the next end step")
	}
	if !g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("the Blitz death trigger did not draw a card")
	}
	if got := namedTokenCount(g, game.Player1, "Treasure"); got != 1 {
		t.Fatalf("Riveteers Requisitioner created %d Treasures, want 1", got)
	}
}
