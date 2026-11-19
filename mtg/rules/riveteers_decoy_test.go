package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Oracle ID 1c4918e6-3810-458e-a11b-949bbc43e7cc, pinned
// printing 2d21701b-6a99-4e1c-82b2-0219f170b09f; exact Oracle fields retained
// in .cardwork/sources/riveteers-decoy-oracle.json. Scryfall rulings URI
// https://api.scryfall.com/cards/2d21701b-6a99-4e1c-82b2-0219f170b09f/rulings,
// retrieved 2026-11-19: five WotC rulings dated 2022-04-29 confirm ordinary
// cast timing/countering, optional attack, draw on any death, a sacrifice that
// affects only the original permanent if still present, and no Blitz effects
// on a permanent copy. Official Comprehensive Rules effective 2026-09-25,
// retained at .cardwork/sources/MagicCompRules-20260925.txt (SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), read
// 118.9, 601.2b/f-h, 302.6, 509.1c, 513.1-2, 603.2b, 603.7a-c, 603.10a,
// 611.2a/c, 700.4, 702.10b, and 702.152a: Blitz is an announced alternative
// cost whose payment grants haste and a death draw and schedules one next-end-
// step sacrifice; haste bypasses summoning sickness, the Decoy's own static
// requirement compels a legal block when possible, and a death means a
// battlefield-to-graveyard move.
func TestAuthoredRiveteersDecoyScenario(t *testing.T) {
	engine := NewEngine(nil)
	g := castCopyProvenanceGame()
	card := addCardToHand(g, game.Player1, authored.RiveteersDecoy.New())
	draw := addCardToLibraryNamed(g, game.Player1, "Riveteers Decoy draw")
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)

	if !engine.applyActionWithChoices(
		g,
		game.Player1,
		action.CastSpell(card, nil, 0, nil),
		blitzAgents("Blitz"),
		&TurnLog{},
	) {
		t.Fatal("could not cast Riveteers Decoy for its Blitz cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !spell.Blitzed || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Blitz cast = %+v, mana remaining = %d", spell, g.Players[game.Player1].ManaPool.Total())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	decoy := permanentForCard(g, card)
	requireBlitzEntry(t, g, decoy, true, true)

	blocker := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: decoy.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}))
	ce := combatEngine{engine}
	if !ce.applyAttackers(g, game.Player1, attack) {
		t.Fatal("Blitz's haste did not let the newly cast Decoy attack")
	}

	g.Turn.Step = game.StepDeclareBlockers
	legal := legalDeclareBlockersActions(g, game.Player2)
	wantBlock := action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: decoy.ObjectID,
	}})
	if len(legal) != 1 || !actionsEqual(legal[0], wantBlock) {
		t.Fatalf("legal blocks = %+v, want the Decoy's only able blocker", legal)
	}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))) {
		t.Fatal("the Decoy's must-be-blocked requirement allowed no blockers")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, wantBlock)) {
		t.Fatal("the Decoy's required blocker was rejected")
	}

	blitzEndStep(engine, g)
	if g.Stack.Size() != 1 || permanentForCard(g, card) == nil {
		t.Fatal("the Blitz sacrifice did not trigger at the next end step")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Graveyard.Contains(card) ||
		!g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("the Blitz sacrifice did not put the Decoy in the graveyard and draw a card")
	}
}
