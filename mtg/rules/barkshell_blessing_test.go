package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned-corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Barkshell Blessing Oracle ID baf3d4d9-0590-4d94-8f4e-1e726194b920,
// printing revision cd273ef2-4aed-4c7e-8c97-fe8b1af9ce69, {G/W} Instant.
// Oracle text: "Target creature gets +2/+2 until end of turn. Conspire (As you
// cast this spell, you may tap two untapped creatures you control that share a
// color with it. When you do, copy it and you may choose a new target for the
// copy.)" Scryfall rulings at
// https://api.scryfall.com/cards/cd273ef2-4aed-4c7e-8c97-fe8b1af9ce69/rulings
// returned an empty list on 2026-11-19. Read the official Wizards Comprehensive
// Rules effective November 6, 2026 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/card-sources/barkshell-blessing/MagicCompRules-20261106.txt):
// 202.2d makes this hybrid-cost spell both green and white; 115.1a and
// 601.2a-i govern its required creature target and spell-casting payment;
// 603.2-3 and 702.78a-b govern the optional tap cost and paid-cast copy;
// 608.2b checks the target on resolution; 611.2a and 613.4c place the +2/+2
// effect through end of turn in layer 7c; 707.10/10c govern copied decisions
// and the optional new target.
func TestBarkshellBlessingPumpsTargetWithConspire(t *testing.T) {
	const oracleID = "baf3d4d9-0590-4d94-8f4e-1e726194b920"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Barkshell Blessing by Oracle ID: %v", err)
	}
	if def.Name != "Barkshell Blessing" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, def)
	target := addCreatureWithPowerToughness(g, game.Player2, 2, 2)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(g, game.Player1)

	agents := conspireTestAgent(chooseConspireCost)
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1,
		action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil), agents, log) {
		t.Fatal("could not cast Barkshell Blessing targeting a creature")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 1 {
		t.Fatalf("expected the hybrid mana and Conspire costs to be paid: %+v", original)
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
	if g.Stack.Size() != 2 {
		t.Fatalf("Conspire left %d stack objects, want the copy and original spell", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, log)
	engine.resolveTopOfStack(g, log)
	if got := effectivePower(g, target); got != 6 {
		t.Fatalf("target power after the original and copied pumps = %d, want 6", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 6 {
		t.Fatalf("target toughness after the original and copied pumps = %d (available %t), want 6", got, ok)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Barkshell Blessing did not finish in its owner's graveyard")
	}
}
