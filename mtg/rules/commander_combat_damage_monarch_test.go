package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// commanderCombatDamageBecomeMonarch builds Archivist of Gondor's first ability:
// "When your commander deals combat damage to a player, if there is no monarch,
// you become the monarch." The trigger subject is the damage source restricted
// to a commander you control (DamageSourceSelection.MatchCommander), gated by the
// NoMonarch intervening condition.
func commanderCombatDamageBecomeMonarch(g *game.Game, controller game.PlayerID) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Archivist",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:                 game.EventDamageDealt,
					Controller:            game.TriggerControllerYou,
					Subject:               game.TriggerSubjectDamageSource,
					RequireCombatDamage:   true,
					DamageRecipient:       game.DamageRecipientPlayer,
					DamageSourceSelection: game.Selection{MatchCommander: true},
				},
				InterveningIf:        "if there is no monarch",
				InterveningCondition: opt.Val(game.Condition{NoMonarch: true}),
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.BecomeMonarch{Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
	return addCombatPermanent(g, controller, def)
}

// TestCommanderCombatDamageBecomeMonarch proves Archivist's first ability: when
// a commander you control deals combat damage to a player and there is no
// monarch, you become the monarch.
func TestCommanderCombatDamageBecomeMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	commanderCombatDamageBecomeMonarch(g, game.Player1)
	commander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Commander",
		Types: []types.Card{types.Creature},
	}})
	g.CommanderIDs[commander.CardInstanceID] = true

	dealPlayerDamage(g, commander.ObjectID, commander.ObjectID, game.Player1, game.Player2, 3, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("commander combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("controller did not become the monarch after their commander dealt combat damage")
	}
}

// TestCommanderCombatDamageMonarchGate proves the NoMonarch intervening
// condition: while a monarch already exists, the commander's combat damage does
// not make you the monarch.
func TestCommanderCombatDamageMonarchGate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player3].IsMonarch = true
	commanderCombatDamageBecomeMonarch(g, game.Player1)
	commander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Commander",
		Types: []types.Card{types.Creature},
	}})
	g.CommanderIDs[commander.CardInstanceID] = true

	dealPlayerDamage(g, commander.ObjectID, commander.ObjectID, game.Player1, game.Player2, 3, true)
	// The intervening NoMonarch gate fails, so no trigger goes on the stack.
	engine.putTriggeredAbilitiesOnStack(g)
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("controller became monarch despite an existing monarch")
	}
	if !g.Players[game.Player3].IsMonarch {
		t.Fatal("existing monarch lost the crown")
	}
}

// TestCommanderCombatDamageRequiresCommander proves the commander restriction:
// combat damage from a non-commander creature you control does not trigger the
// ability.
func TestCommanderCombatDamageRequiresCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	commanderCombatDamageBecomeMonarch(g, game.Player1)
	nonCommander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
	}})

	dealPlayerDamage(g, nonCommander.ObjectID, nonCommander.ObjectID, game.Player1, game.Player2, 3, true)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired for combat damage from a non-commander creature")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("controller became monarch from non-commander combat damage")
	}
}
