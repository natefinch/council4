package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID cf3b1070-2c08-49b9-aa78-fe676fab38aa, printing revision
// c011b432-ec39-44c8-bf2a-1c3fec32a576 (Guildpact #121), {U}{R}, Instant.
// Exact text: “Replicate {U}{R} (When you cast this spell, copy it for each
// time you paid its replicate cost. You may choose new targets for the copies.)
// Target creature gets +1/+0 and gains flying and first strike until end of
// turn.” Scryfall's card endpoint https://api.scryfall.com/cards/c011b432-ec39-44c8-bf2a-1c3fec32a576
// confirmed the pinned text on 2026-11-19; its rulings endpoint
// https://api.scryfall.com/cards/c011b432-ec39-44c8-bf2a-1c3fec32a576/rulings
// returned an empty list. Read the official Wizards Comprehensive Rules
// effective 2026-11-06 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/source/MagicCompRules-20261106.txt): 115.1a
// and 601.2a-i govern the target, announcement, additional costs and casting;
// 603.2-4 and 702.56a govern the paid Replicate trigger and copies per payment;
// 608.2b rechecks target legality; 611.2a and 613.1f-g govern the end-of-turn
// ability and power effects; 707.10/.10c govern copied spell choices and new
// target choices.
func TestLeapOfFlameReplicateBuffsThreeCreatures(t *testing.T) {
	const oracleID = "cf3b1070-2c08-49b9-aa78-fe676fab38aa"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Leap of Flame by Oracle ID: %v", err)
	}
	if def.Name != "Leap of Flame" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, def)
	creatures := []*game.Permanent{
		addCreaturePermanent(g, game.Player2),
		addCreaturePermanent(g, game.Player2),
		addCreaturePermanent(g, game.Player2),
	}
	targets := []game.Target{
		game.PermanentTarget(creatures[0].ObjectID),
		game.PermanentTarget(creatures[1].ObjectID),
		game.PermanentTarget(creatures[2].ObjectID),
	}
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	setMainPhasePriority(g, game.Player1)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID:          spellID,
		SourceZone:      zone.Hand,
		Targets:         targets[:1],
		ReplicateCounts: []int{2},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("could not cast Leap of Flame paying its Replicate cost twice")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 6 || len(original.ReplicateCounts) != 1 || original.ReplicateCounts[0] != 2 ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("expected {U}{R} plus two {U}{R} Replicate payments: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("the two Replicate payments did not trigger above Leap of Flame")
	}

	nextTarget := 1
	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			if nextTarget >= len(targets) {
				t.Fatal("Replicate asked for more than two new targets")
			}
			for _, option := range req.Options {
				if slices.Equal(option.Targets, targets[nextTarget:nextTarget+1]) {
					nextTarget++
					return []int{option.Index}
				}
			}
			t.Fatalf("Replicate did not offer creature %d as a legal target", nextTarget+1)
		}
		return req.DefaultSelection
	})
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if nextTarget != len(targets) || g.Stack.Size() != 3 {
		t.Fatalf("two Replicate payments produced %d stack objects with %d new targets; want two copies and the original", g.Stack.Size(), nextTarget-1)
	}
	for range 3 {
		engine.resolveTopOfStackWithChoices(g, agents, log)
	}

	for _, target := range creatures {
		permanent, ok := permanentByObjectID(g, target.ObjectID)
		if !ok || effectivePower(g, permanent) != 2 || effectiveToughnessMust(t, g, permanent) != 1 ||
			!hasKeyword(g, permanent, game.Flying) || !hasKeyword(g, permanent, game.FirstStrike) {
			t.Fatalf("Leap of Flame did not give creature %d +1/+0, flying, and first strike until end of turn", target.ObjectID)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Leap of Flame did not finish in its owner's graveyard")
	}
}
