package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func damagedPlayerDestroySourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Heavy Power Hammer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:               game.EventDamageDealt,
					Source:              game.TriggerSourceSelf,
					RequireCombatDamage: true,
					DamageRecipient:     game.DamageRecipientPlayer,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny:        []types.Card{types.Artifact, types.Enchantment},
						ControlledByEventPlayer: true,
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability(),
		}},
	}}
}

func addDamagedPlayerTargetPermanent(g *game.Game, controller game.PlayerID, name string, cardTypes ...types.Card) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: cardTypes,
	}})
}

func emitCombatDamageToPlayer(g *game.Game, source *game.Permanent, player game.PlayerID) {
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        source.CardInstanceID,
		SourceObjectID:  source.ObjectID,
		Controller:      effectiveController(g, source),
		Player:          player,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Amount:          1,
	})
}

func TestDamagedPlayerTargetRelationAndArtifactEnchantmentUnion(t *testing.T) {
	for _, targetType := range []types.Card{types.Artifact, types.Enchantment} {
		t.Run(string(targetType), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, damagedPlayerDestroySourceDef())
			bystander := addDamagedPlayerTargetPermanent(g, game.Player3, "Bystander", targetType)
			damagedPlayersPermanent := addDamagedPlayerTargetPermanent(g, game.Player2, "Damaged Player's Permanent", targetType)

			emitCombatDamageToPlayer(g, source, game.Player2)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("combat-damage destroy trigger found no legal target")
			}
			top, _ := g.Stack.Peek()
			if len(top.Targets) != 1 || top.Targets[0].PermanentID != damagedPlayersPermanent.ObjectID {
				t.Fatalf("targets = %#v, want damaged player's permanent %v", top.Targets, damagedPlayersPermanent.ObjectID)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, ok := permanentByObjectID(g, damagedPlayersPermanent.ObjectID); ok {
				t.Fatal("damaged player's target was not destroyed")
			}
			if _, ok := permanentByObjectID(g, bystander.ObjectID); !ok {
				t.Fatal("other opponent's permanent was destroyed")
			}
		})
	}
}

func TestDamagedPlayerTargetRechecksLiveControllerAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, damagedPlayerDestroySourceDef())
	target := addDamagedPlayerTargetPermanent(g, game.Player2, "Changing Allegiance", types.Artifact)

	emitCombatDamageToPlayer(g, source, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage destroy trigger found no legal target")
	}
	target.Controller = game.Player3
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("target was destroyed after it stopped being controlled by the damaged player")
	}
}

func TestDamagedPlayerTargetRequiresLegalPermanentAndUsesStandardDestroy(t *testing.T) {
	t.Run("no legal target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, damagedPlayerDestroySourceDef())
		addDamagedPlayerTargetPermanent(g, game.Player3, "Wrong Opponent's Artifact", types.Artifact)

		emitCombatDamageToPlayer(g, source, game.Player2)
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("trigger reached stack without a legal damaged-player-controlled target")
		}
	})

	t.Run("indestructible", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, damagedPlayerDestroySourceDef())
		target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name:  "Indestructible Relic",
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
		}})

		emitCombatDamageToPlayer(g, source, game.Player2)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("trigger found no indestructible artifact target")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
			t.Fatal("standard destruction ignored indestructible")
		}
	})
}
