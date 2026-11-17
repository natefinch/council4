package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: issue #3957. Official CR effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
// read 208.3, 208.5, 613.4b, 613.4c and 704.5f. 208.3: a noncreature permanent
// has no power or toughness. 208.5: "If a creature somehow has no value for its
// power, its power is 0. The same is true for toughness." 613.4c: effects that
// modify power and toughness (including counters) apply after 7b set effects.
// 704.5f: a creature with toughness 0 or less is put into its owner's
// graveyard. Card: Chimeric Mass (oracle 98066238-28f8-41e3-903a-d367927f23ca),
// the unregistered fixture in continuous_granted_self_static_test.go, whose
// activation grants "This creature's power and toughness are each equal to the
// number of charge counters on it." Removing that granted ability leaves an
// artifact creature with no P/T value (rulings 2011-01-01: with no charge
// counters it is 0/0 and is put into the graveyard).

// TestChimericMassWithoutPowerToughnessAbilityIsZeroZero plays an animated
// Chimeric Mass that loses its granted P/T ability. Its undefined P/T is 0
// (CR 208.5), so Giant Growth's +3/+3 keeps it alive as a 3/3. After cleanup it
// is a noncreature artifact with no P/T and stays (CR 208.3). Animated and
// drained again without a pump, it is 0/0 and dies (CR 704.5f).
func TestChimericMassWithoutPowerToughnessAbilityIsZeroZero(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	mass := castChimericMass(t, g, engine, 3)
	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	castAndResolveRegistered(t, g, engine, game.Player1, "Giant Growth", mana.G, 1, mass)
	assertPowerToughness(t, g, mass, 6, 6)

	drainAbilities(t, g, engine, mass)
	if !permanentHasType(g, mass, types.Creature) {
		t.Fatal("losing abilities ended the animation's type change")
	}
	if toughness, ok := effectiveToughness(g, mass); !ok || toughness != 3 {
		t.Fatalf("drained Chimeric Mass toughness = %d (defined %v), want 3", toughness, ok)
	}
	assertPowerToughness(t, g, mass, 3, 3)
	engine.applyStateBasedActions(g)
	if _, ok := findPermanentByCardID(g, mass.CardInstanceID); !ok {
		t.Fatal("pumped Chimeric Mass with no P/T ability died")
	}

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentHasType(g, mass, types.Creature) {
		t.Fatal("Chimeric Mass is still a creature after cleanup")
	}
	if _, ok := effectiveToughness(g, mass); ok {
		t.Fatal("noncreature Chimeric Mass has a toughness value")
	}
	engine.applyStateBasedActions(g)
	if _, ok := findPermanentByCardID(g, mass.CardInstanceID); !ok {
		t.Fatal("noncreature Chimeric Mass left the battlefield")
	}

	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	assertPowerToughness(t, g, mass, 3, 3)
	drainAbilities(t, g, engine, mass)
	assertPowerToughness(t, g, mass, 0, 0)
	engine.applyStateBasedActions(g)
	if _, ok := findPermanentByCardID(g, mass.CardInstanceID); ok ||
		!g.Players[game.Player1].Graveyard.Contains(mass.CardInstanceID) {
		t.Fatal("0/0 Chimeric Mass was not put into its owner's graveyard")
	}
}

// drainAbilities casts the "loses all abilities" fixture at target through the
// real cast path and resolves it.
func drainAbilities(t *testing.T, g *game.Game, engine *Engine, target *game.Permanent) {
	t.Helper()
	spell := addCardToHand(g, game.Player1, loseAllAbilitiesFixture())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatal("casting the ability-loss fixture failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}
