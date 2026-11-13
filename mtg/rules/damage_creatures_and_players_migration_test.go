package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: existing cards migrated to one creature-and-player damage
// batch (#3815). Comprehensive Rules effective 2026-11-06 (verified copy
// retained with #3762): 120.4 (one damage event, prevention/replacement
// applied to the whole event), 608.2f (an action on several objects and players
// is processed simultaneously), 113.7a and 603.10a (a dies trigger's source
// deals damage using last known information), 702.15e (one lifelink life-gain
// event per source per simultaneous damage), 702.33d (kicked spells).
// Scryfall Oracle (pinned corpus 6da9ced0...):
//   - Breath of Darigaaz 1954994f {1}{R} "Kicker {2} ... Breath of Darigaaz
//     deals 1 damage to each creature without flying and each player. If this
//     spell was kicked, it deals 4 damage to each creature without flying and
//     each player instead." Rulings 2024-11-08 are generic kicker rulings.
//   - Pestilence Demon 2572add6 {5}{B}{B}{B} 7/6 "Flying / {B}: This creature
//     deals 1 damage to each creature and each player." Ruling 2010-06-15: it
//     deals 1 damage to itself as well as each other creature and each player.
//   - Fire Tempest ede6a352 {5}{R}{R} "Fire Tempest deals 6 damage to each
//     creature and each player." (no rulings)
//   - Magma Phoenix 9c6edb99 {3}{R}{R} 3/3 "Flying / When this creature dies,
//     it deals 3 damage to each creature and each player. / {3}{R}{R}: Return
//     this card from your graveyard to your hand." (no rulings)
//   - Tectonic Hazard 9097243a {R} "Tectonic Hazard deals 1 damage to each
//     opponent and each creature they control." (no rulings)
//   - Chandra's Ignition f61680da {3}{R}{R} "Target creature you control deals
//     damage equal to its power to each other creature and each opponent."
//     Rulings 2015-06-22: the creature is the source; its power is read on
//     resolution; an illegal target means no damage is dealt.

// requireOneDamageBatch reports the damage events after skip and requires
// them to be want events sharing one nonzero simultaneous ID.
func requireOneDamageBatch(t *testing.T, g *game.Game, skip, want int) []game.Event {
	t.Helper()
	var events []game.Event
	for _, event := range g.Events[skip:] {
		if event.Kind == game.EventDamageDealt {
			events = append(events, event)
		}
	}
	if len(events) != want {
		t.Fatalf("damage events = %d, want %d", len(events), want)
	}
	for _, event := range events {
		if event.SimultaneousID == 0 || event.SimultaneousID != events[0].SimultaneousID {
			t.Fatalf("damage event batch IDs differ or are zero: %+v", events)
		}
	}
	return events
}

func requireLives(t *testing.T, g *game.Game, want [game.NumPlayers]int) {
	t.Helper()
	for player := range g.Players {
		if g.Players[player].Life != want[player] {
			t.Fatalf("player %d life = %d, want %d", player, g.Players[player].Life, want[player])
		}
	}
}

func allLives(life int) [game.NumPlayers]int {
	var lives [game.NumPlayers]int
	for player := range lives {
		lives[player] = life
	}
	return lives
}

func castRegistrySpell(t *testing.T, engine *Engine, g *game.Game, name string, cast func(id.ID) action.Action, pool map[mana.Color]int) {
	t.Helper()
	spell := addCardToHand(g, game.Player1, registryCard(t, name))
	for color, amount := range pool {
		g.Players[game.Player1].ManaPool.Add(color, amount)
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, cast(spell)) {
		t.Fatalf("%s did not cast", name)
	}
	if left := g.Players[game.Player1].ManaPool.Total(); left != 0 {
		t.Fatalf("%s left %d mana unpaid", name, left)
	}
}

// TestBreathOfDarigaazScenario casts Breath of Darigaaz unkicked and kicked:
// each version deals its amount to each creature without flying and each
// player as one damage event, and the flyer is dealt none.
func TestBreathOfDarigaazScenario(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kicked bool
		damage int
	}{
		{name: "unkicked", damage: 1},
		{name: "kicked", kicked: true, damage: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
			flyer := addCombatCreaturePermanentWithPower(g, game.Player3, 5, game.Flying)
			cast := func(spell id.ID) action.Action { return action.CastSpell(spell, nil, 0, nil) }
			pool := map[mana.Color]int{mana.R: 2}
			if tc.kicked {
				cast = func(spell id.ID) action.Action { return action.CastKickedSpell(spell, nil, 0, nil) }
				pool[mana.R] = 4
			}
			castRegistrySpell(t, engine, g, "Breath of Darigaaz", cast, pool)
			skip := len(g.Events)
			engine.resolveTopOfStack(g, &TurnLog{})

			if grounded.MarkedDamage != tc.damage || flyer.MarkedDamage != 0 {
				t.Fatalf("grounded damage %d, flyer %d; want %d, 0", grounded.MarkedDamage, flyer.MarkedDamage, tc.damage)
			}
			requireLives(t, g, allLives(40-tc.damage))
			requireOneDamageBatch(t, g, skip, 1+game.NumPlayers)
		})
	}
}

// TestPestilenceDemonScenario activates Pestilence Demon's {B} ability: it
// deals 1 damage to itself, each other creature and each player in one event.
func TestPestilenceDemonScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	demon := addCombatPermanent(g, game.Player1, registryCard(t, "Pestilence Demon"))
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(demon.ObjectID, 0, nil, 0)) {
		t.Fatal("Pestilence Demon's ability did not activate")
	}
	skip := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if demon.MarkedDamage != 1 || other.MarkedDamage != 1 {
		t.Fatalf("creature damage = %d, %d; want 1, 1", demon.MarkedDamage, other.MarkedDamage)
	}
	requireLives(t, g, allLives(39))
	for _, event := range requireOneDamageBatch(t, g, skip, 2+game.NumPlayers) {
		if event.SourceObjectID != demon.ObjectID {
			t.Fatalf("damage source = %v, want Pestilence Demon %v", event.SourceObjectID, demon.ObjectID)
		}
	}
}

// TestFireTempestKillsMagmaPhoenixScenario casts Fire Tempest, which deals 6
// damage to each creature and each player in one event; Magma Phoenix dies and
// its trigger deals 3 damage to each creature and each player as a second,
// separate event from the Phoenix's last known information.
func TestFireTempestKillsMagmaPhoenixScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	phoenix := addCombatPermanent(g, game.Player1, registryCard(t, "Magma Phoenix"))
	big := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	castRegistrySpell(t, engine, g, "Fire Tempest", func(spell id.ID) action.Action {
		return action.CastSpell(spell, nil, 0, nil)
	}, map[mana.Color]int{mana.R: 7})
	skip := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if phoenix.MarkedDamage != 6 || big.MarkedDamage != 6 {
		t.Fatalf("creature damage = %d, %d; want 6, 6", phoenix.MarkedDamage, big.MarkedDamage)
	}
	requireLives(t, g, allLives(34))
	tempest := requireOneDamageBatch(t, g, skip, 2+game.NumPlayers)

	engine.applyStateBasedActions(g)
	if _, ok := g.PermanentByID(phoenix.ObjectID); ok {
		t.Fatal("Magma Phoenix survived 6 damage")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("Magma Phoenix's dies trigger not on the stack; stack size %d", g.Stack.Size())
	}
	skip = len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if big.MarkedDamage != 9 {
		t.Fatalf("surviving creature damage = %d, want 9", big.MarkedDamage)
	}
	requireLives(t, g, allLives(31))
	dies := requireOneDamageBatch(t, g, skip, 1+game.NumPlayers)
	if dies[0].SimultaneousID == tempest[0].SimultaneousID {
		t.Fatal("Magma Phoenix's damage shares Fire Tempest's damage event")
	}
	for _, event := range dies {
		if event.SourceObjectID != phoenix.ObjectID {
			t.Fatalf("dies-trigger damage source = %v, want Magma Phoenix %v", event.SourceObjectID, phoenix.ObjectID)
		}
	}
}

// TestTectonicHazardScenario casts Tectonic Hazard: each opponent and each
// creature they control is dealt 1 damage in one event, while the caster and
// the caster's creature are not.
func TestTectonicHazardScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	own := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	another := addCombatCreaturePermanentWithPower(g, game.Player3, 2, game.Flying)
	castRegistrySpell(t, engine, g, "Tectonic Hazard", func(spell id.ID) action.Action {
		return action.CastSpell(spell, nil, 0, nil)
	}, map[mana.Color]int{mana.R: 1})
	skip := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if own.MarkedDamage != 0 || theirs.MarkedDamage != 1 || another.MarkedDamage != 1 {
		t.Fatalf("creature damage = own %d, opponents %d, %d; want 0, 1, 1", own.MarkedDamage, theirs.MarkedDamage, another.MarkedDamage)
	}
	lives := allLives(39)
	lives[game.Player1] = 40
	requireLives(t, g, lives)
	requireOneDamageBatch(t, g, skip, 2+game.NumPlayers-1)
}

// TestChandrasIgnitionScenario casts Chandra's Ignition targeting a 4-power
// lifelink creature: it deals 4 damage to each other creature and each
// opponent in one event, so its controller gains life once for the total.
func TestChandrasIgnitionScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 4, game.Lifelink)
	ownOther := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	castRegistrySpell(t, engine, g, "Chandra's Ignition", func(spell id.ID) action.Action {
		return action.CastSpell(spell, []game.Target{game.PermanentTarget(source.ObjectID)}, 0, nil)
	}, map[mana.Color]int{mana.R: 5})
	skip := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if source.MarkedDamage != 0 || ownOther.MarkedDamage != 4 || theirs.MarkedDamage != 4 {
		t.Fatalf("creature damage = source %d, others %d, %d; want 0, 4, 4", source.MarkedDamage, ownOther.MarkedDamage, theirs.MarkedDamage)
	}
	opponents := game.NumPlayers - 1
	total := 4 * (2 + opponents)
	lives := allLives(36)
	lives[game.Player1] = 40 + total
	requireLives(t, g, lives)
	for _, event := range requireOneDamageBatch(t, g, skip, 2+opponents) {
		if event.SourceObjectID != source.ObjectID {
			t.Fatalf("damage source = %v, want the targeted creature %v", event.SourceObjectID, source.ObjectID)
		}
	}
	gains := 0
	for _, event := range g.Events[skip:] {
		if event.Kind == game.EventLifeGained {
			gains++
		}
	}
	if gains != 1 {
		t.Fatalf("life-gain events = %d, want one for the whole event", gains)
	}
}
