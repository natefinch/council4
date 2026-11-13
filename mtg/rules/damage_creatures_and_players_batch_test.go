package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: one instruction dealing damage to each creature and each
// player (#3762). Comprehensive Rules effective 2026-11-06 (local verified copy
// .cardwork/MagicCompRules-20261106.txt from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
//   - 120.3, 120.4: damage is processed as one damage event in which
//     replacement and prevention effects (120.4a) apply before all of its
//     results are dealt together and the event triggers abilities (120.4b).
//   - 608.2f: a spell's action performed on several objects or players is
//     performed on them simultaneously, so "each creature and each player" is
//     one event, not a creature event followed by a player event.
//   - 615.6, 615.7, 616.1: a "next time ... would deal damage" shield applies
//     to all of the one event it matches, with each affected player or
//     controller applying it in APNAP order.
//   - 702.15b, 702.15e: lifelink gains life equal to the damage dealt; one
//     source dealing damage simultaneously causes one life-gain event.
// Scryfall Oracle (pinned corpus 6da9ced0...): Earthquake 9a40614b {X}{R}
// "Earthquake deals X damage to each creature without flying and each
// player." with its 2004-10-04 ruling that being without flying is checked
// only on resolution; Sickening Dreams 728d2e30 {1}{B} "As an additional cost
// to cast this spell, discard X cards. Sickening Dreams deals X damage to each
// creature and each player." (no rulings; its card is authored separately).

func eachCreatureAndPlayer(amount game.Quantity) game.Damage {
	return game.Damage{Amount: amount, Recipient: game.PermanentsAndPlayersDamageRecipient(
		game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		game.AllPlayersReference(),
	)}
}

func damageEventBatchIDs(g *game.Game) []id.ID {
	var ids []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventDamageDealt {
			ids = append(ids, event.SimultaneousID)
		}
	}
	return ids
}

func castEarthquake(t *testing.T, engine *Engine, g *game.Game, x int) id.ID {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup("Earthquake")
	if def == nil {
		t.Fatal("Earthquake is not registered")
	}
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, x+1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, x, nil)) {
		t.Fatal("Earthquake did not cast")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Earthquake left %d mana unpaid for {X}{R} with X=%d", g.Players[game.Player1].ManaPool.Total(), x)
	}
	return spell
}

// TestEarthquakeScenario casts Earthquake for X=2: each creature without
// flying and each player is dealt 2 damage in one simultaneous damage event,
// and the flyer is dealt none.
func TestEarthquakeScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	ownGrounded := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	flyer := addCombatCreaturePermanentWithPower(g, game.Player3, 3, game.Flying)
	castEarthquake(t, engine, g, 2)
	engine.resolveTopOfStack(g, &TurnLog{})

	if grounded.MarkedDamage != 2 || ownGrounded.MarkedDamage != 2 || flyer.MarkedDamage != 0 {
		t.Fatalf("creature damage = %d, %d, flyer %d; want 2, 2, 0", grounded.MarkedDamage, ownGrounded.MarkedDamage, flyer.MarkedDamage)
	}
	for player := range g.Players {
		if g.Players[player].Life != 38 {
			t.Fatalf("player %d life = %d, want 38", player, g.Players[player].Life)
		}
	}
	ids := damageEventBatchIDs(g)
	if len(ids) != 2+game.NumPlayers {
		t.Fatalf("damage events = %d, want one per grounded creature and player", len(ids))
	}
	for _, batch := range ids {
		if batch == 0 || batch != ids[0] {
			t.Fatalf("damage event batch IDs = %v, want one shared nonzero ID", ids)
		}
	}
}

// TestEarthquakeChosenSourceShieldPreventsWholeEvent creates a Pilgrim of
// Justice-style "next time a red source of your choice would deal damage this
// turn" shield choosing Earthquake on the stack: its damage to every creature
// and every player is one event, so the shield prevents all of it as one use.
func TestEarthquakeChosenSourceShieldPreventsWholeEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player1, 3),
		addCombatCreaturePermanentWithPower(g, game.Player2, 3),
	}
	spell := castEarthquake(t, engine, g, 2)
	agent := sourceChoiceAgent(t, spell, nil)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, game.PreventDamage{
		Global: true, All: true, OneShot: true, SourceColors: []color.Color{color.Red},
	}, agent)
	if len(g.PreventionShields) != 1 {
		t.Fatalf("shields = %+v, want one bound to Earthquake", g.PreventionShields)
	}

	var agents [game.NumPlayers]PlayerAgent
	for player := range agents {
		agents[player] = agent
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	for _, creature := range creatures {
		if creature.MarkedDamage != 0 {
			t.Fatalf("creature of player %d marked %d, want the whole event prevented", creature.Controller, creature.MarkedDamage)
		}
	}
	for player := range g.Players {
		if g.Players[player].Life != 40 {
			t.Fatalf("player %d life = %d, want 40 with the whole event prevented", player, g.Players[player].Life)
		}
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want the shield used once", g.PreventionShields)
	}
	if ids := damageEventBatchIDs(g); len(ids) != 0 {
		t.Fatalf("damage events = %d, want none dealt", len(ids))
	}
}

// TestCreaturesAndPlayersDamageGainsLifelinkOnce proves a lifelink source's
// damage to each creature and each player is one life-gain event of the total
// dealt (CR 702.15e), and its deathtouch marks each creature it damaged.
func TestCreaturesAndPlayersDamageGainsLifelinkOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 4, game.Lifelink, game.Deathtouch)
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 4)

	resolvePermanentDamage(engine, g, source, eachCreatureAndPlayer(game.Fixed(1)))
	if !source.MarkedDeathtouchDamage || !other.MarkedDeathtouchDamage {
		t.Fatalf("deathtouch marks = %v, %v; want both creatures marked", source.MarkedDeathtouchDamage, other.MarkedDeathtouchDamage)
	}
	total := 2 + game.NumPlayers
	if got := g.Players[game.Player1].Life; got != 40-1+total {
		t.Fatalf("controller life = %d, want %d after 1 damage and one gain of %d", got, 40-1+total, total)
	}
	gains := 0
	for _, event := range g.Events {
		if event.Kind == game.EventLifeGained {
			gains++
			if event.Amount != total {
				t.Fatalf("life gained %d, want the event's total %d", event.Amount, total)
			}
		}
	}
	if gains != 1 {
		t.Fatalf("life-gain events = %d, want 1", gains)
	}
}
