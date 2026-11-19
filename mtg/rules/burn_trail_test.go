package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle revision 7f01f9a0-f1d0-4241-a270-df4ed673d1fd, Oracle ID
// 566b4dcc-b7d7-4443-b2d0-954f1fc33a11, Sorcery {3}{R}. Oracle text:
// "Burn Trail deals 3 damage to any target. Conspire (As you cast this spell,
// you may tap two untapped creatures you control that share a color with it.
// When you do, copy it and you may choose a new target for the copy.)" The
// printing's Scryfall rulings
// URL, https://api.scryfall.com/cards/7f01f9a0-f1d0-4241-a270-df4ed673d1fd/rulings,
// returned an empty list on 2026-11-19. Read the official Wizards Comprehensive
// Rules effective November 6, 2026 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/MagicCompRules_20261106.txt): 105.2 identifies
// the spell's color from its red mana symbol; 115.4 defines "any target" as a
// creature, player, planeswalker, or battle; 120.1, 120.2b, and 120.3a govern
// spell damage and life loss; 601.2a-i governs targeting and paying the
// optional Conspire cost;
// 603.2-3 and 702.78a-b make the paid cast trigger a copy; 608.2b checks targets
// on resolution; 707.10/.10c govern copied decisions and optional retargeting.
func TestBurnTrailDealsDamageWithConspire(t *testing.T) {
	const oracleID = "566b4dcc-b7d7-4443-b2d0-954f1fc33a11"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Burn Trail by Oracle ID: %v", err)
	}
	if def.Name != "Burn Trail" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, def)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	initialLife := g.Players[game.Player2].Life
	target := game.PlayerTarget(game.Player2)
	setMainPhasePriority(g, game.Player1)

	agents := conspireTestAgent(chooseConspireCost)
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1,
		action.CastSpell(spellID, []game.Target{target}, 0, nil), agents, log) {
		t.Fatal("could not cast Burn Trail targeting the opponent")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 4 {
		t.Fatalf("expected Burn Trail's mana and Conspire costs to be paid: %+v", original)
	}
	for _, creature := range []*game.Permanent{first, second} {
		if permanent, _ := permanentByObjectID(g, creature.ObjectID); !permanent.Tapped {
			t.Fatalf("Conspire did not tap creature %d", creature.ObjectID)
		}
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("paid Conspire did not put its cast trigger on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.resolveTopOfStack(g, log)
	if got := g.Players[game.Player2].Life; got != initialLife-3 {
		t.Fatalf("Burn Trail copy left opponent at %d life, want %d", got, initialLife-3)
	}
	engine.resolveTopOfStack(g, log)
	if got := g.Players[game.Player2].Life; got != initialLife-6 {
		t.Fatalf("Burn Trail left opponent at %d life, want %d", got, initialLife-6)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Burn Trail did not finish in its owner's graveyard")
	}
}
