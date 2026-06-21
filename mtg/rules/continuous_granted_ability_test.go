package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// grantedTriggeredAbilityEquipment is an Equipment that grants its attached
// creature a quoted triggered ability, exercising the
// StaticContinuousGrantAbility lowering: a continuous LayerAbility effect whose
// AddAbilities contains a full *game.TriggeredAbility.
func grantedTriggeredAbilityEquipment(g *game.Game, controller game.PlayerID) *game.Permanent {
	granted := &game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				RequireCombatDamage: true,
				DamageRecipient:     game.DamageRecipientPlayer,
			},
		},
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			}},
		}.Ability(),
	}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Granting Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:        game.LayerAbility,
				Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
				AddAbilities: []game.Ability{granted},
			}},
		}},
	}})
}

func countGrantedTriggeredAbilities(g *game.Game, permanent *game.Permanent) int {
	count := 0
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		if _, ok := ability.(*game.TriggeredAbility); ok {
			count++
		}
	}
	return count
}

// TestEquippedCreatureGainsGrantedTriggeredAbility confirms an Equipment's
// continuous AddAbilities grant adds a quoted triggered ability to the creature
// it is attached to, and only while attached.
func TestEquippedCreatureGainsGrantedTriggeredAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	equipment := grantedTriggeredAbilityEquipment(g, game.Player1)

	if got := countGrantedTriggeredAbilities(g, creature); got != 0 {
		t.Fatalf("granted triggered abilities before attaching = %d, want 0", got)
	}

	creature.Attachments = append(creature.Attachments, equipment.ObjectID)
	equipment.AttachedTo = opt.Val(creature.ObjectID)

	if got := countGrantedTriggeredAbilities(g, creature); got != 1 {
		t.Fatalf("granted triggered abilities while attached = %d, want 1", got)
	}

	// The grant stops applying once the Equipment leaves the battlefield.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if got := countGrantedTriggeredAbilities(g, creature); got != 0 {
		t.Fatalf("granted triggered abilities after equipment leaves = %d, want 0", got)
	}
}
