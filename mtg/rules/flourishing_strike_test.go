package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Flourishing Strike printing revision 9c4b23a5-d997-4f02-b136-44ee7be51c7b,
// Oracle ID 98c44298-7b6a-4f98-bcb7-1a63a4c29f87, exact text:
// "Choose one —
// • Flourishing Strike deals 5 damage to target creature with flying.
// • Target creature gets +3/+3 until end of turn.
// Entwine {2}{G} (Choose both if you pay the entwine cost.)"
// The current Scryfall record at
// https://api.scryfall.com/cards/9c4b23a5-d997-4f02-b136-44ee7be51c7b,
// retrieved 2026-11-21, matches the pinned identity and text. Its rulings URI
// https://api.scryfall.com/cards/9c4b23a5-d997-4f02-b136-44ee7be51c7b/rulings
// returned an empty list.
// Official Wizards Comprehensive Rules effective 2026-09-25 (the corpus rules
// pin), https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// retained at .cardwork/flourishing-strike/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 700.2a/c and 601.2b/c require choosing modes before choosing only their
// targets; 601.2f-h and 702.42a require paying the additional entwine cost to
// choose both modes, and 702.42b/608.2c resolve them in printed order. Rule
// 115.1a identifies creature targets, 608.2b rechecks their legality, 120.3e
// marks creature damage, 613.4c applies +3/+3 in layer 7c, and 514.2 ends the
// pump during cleanup. Runtime trace: payment/options.go charges EscalateCost
// once per extra selected mode; effects.go binds each selected mode's target
// segment before resolving its Sequence; primitive_handlers_damage.go handles
// Damage; and primitive_handlers.go applies the continuous effect evaluated by
// continuous.go in layer 7c.
func TestFlourishingStrikeEntwineDealsDamageAndPumps(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("98c44298-7b6a-4f98-bcb7-1a63a4c29f87")
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Flourishing Strike" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flyingCreature := addCombatPermanent(
		g, game.Player2, vanillaCreature("Flourishing Strike flying target", 6, 6, game.Flying),
	)
	pumpedCreature := addCombatPermanent(
		g, game.Player2, vanillaCreature("Flourishing Strike pump target", 2, 2),
	)
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	setMainPhasePriority(g, game.Player1)

	engine := NewEngineWithPrograms(nil, registry.Programs())
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(flyingCreature.ObjectID),
		game.PermanentTarget(pumpedCreature.ObjectID),
	}, 0, []int{0, 1})) {
		t.Fatal("Flourishing Strike did not cast both modes with the entwine cost")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining after entwined cast = %d, want 0", got)
	}
	stackObject, ok := g.Stack.Peek()
	if !ok || len(stackObject.ChosenModes) != 2 ||
		stackObject.ChosenModes[0] != 0 || stackObject.ChosenModes[1] != 1 {
		t.Fatalf("entwined stack modes = %v, want [0 1]", stackObject.ChosenModes)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if flyingCreature.MarkedDamage != 5 {
		t.Fatalf("damage marked on flying creature = %d, want 5", flyingCreature.MarkedDamage)
	}
	if got := effectivePower(g, pumpedCreature); got != 5 {
		t.Fatalf("pumped creature power = %d, want 5", got)
	}
	if got, ok := effectiveToughness(g, pumpedCreature); !ok || got != 5 {
		t.Fatalf("pumped creature toughness = %d (defined %v), want 5", got, ok)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Flourishing Strike did not go to its owner's graveyard")
	}

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := effectivePower(g, pumpedCreature); got != 2 {
		t.Fatalf("pumped creature power after cleanup = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, pumpedCreature); !ok || got != 2 {
		t.Fatalf("pumped creature toughness after cleanup = %d (defined %v), want 2", got, ok)
	}
}
