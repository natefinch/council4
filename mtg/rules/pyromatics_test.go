package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle revision 795a7929-2c32-4006-bdf1-def0a4400f59, Oracle ID
// 9c01e2a6-61e7-455c-b09f-8c0de2266782, Instant {1}{R}. Oracle text:
// “Replicate {1}{R} (When you cast this spell, copy it for each time you paid
// its replicate cost. You may choose new targets for the copies.) Pyromatics
// deals 1 damage to any target.” The printing's Scryfall rulings URL,
// https://api.scryfall.com/cards/795a7929-2c32-4006-bdf1-def0a4400f59/rulings,
// returned an empty list on 2026-11-19. Read the official Wizards Comprehensive
// Rules effective November 6, 2026 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/source/MagicCompRules-20261106.txt): 115.4
// defines “any target”; 120.1, 120.2b, and 120.3a govern damage and player life
// loss; 601.2a-i covers announcing the target and paying the spell and
// additional costs; 603.2-4 and 702.56a-b govern the cast trigger, its stack
// placement, and Replicate payment count; 608.2b rechecks target legality;
// 707.10/.10c govern the copies and their target choices.
func TestPyromaticsReplicateDealsDamage(t *testing.T) {
	const oracleID = "9c01e2a6-61e7-455c-b09f-8c0de2266782"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Pyromatics by Oracle ID: %v", err)
	}
	if def.Name != "Pyromatics" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	initialLife := g.Players[game.Player2].Life
	target := game.PlayerTarget(game.Player2)
	setMainPhasePriority(g, game.Player1)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID:          spellID,
		SourceZone:      zone.Hand,
		Targets:         []game.Target{target},
		ReplicateCounts: []int{2},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("could not cast Pyromatics paying its Replicate cost twice")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 6 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("expected Pyromatics and both Replicate payments to cost {3}{R}{R}{R}: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("paid Replicate did not trigger when Pyromatics was cast")
	}

	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		return req.DefaultSelection
	})
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if g.Stack.Size() != 3 {
		t.Fatalf("two Replicate payments should make two copies plus the original, got %d stack objects", g.Stack.Size())
	}
	for range 3 {
		engine.resolveTopOfStack(g, log)
	}
	if got := g.Players[game.Player2].Life; got != initialLife-3 {
		t.Fatalf("Pyromatics and its two copies dealt %d damage, want 3", initialLife-got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Pyromatics did not finish in its owner's graveyard")
	}
}
