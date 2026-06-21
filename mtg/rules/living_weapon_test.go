package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func addLivingWeaponEquipment(g *game.Game, controller game.PlayerID) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               "Living Weapon Equipment",
		Types:              []types.Card{types.Artifact},
		Subtypes:           []types.Sub{types.Equipment},
		TriggeredAbilities: []game.TriggeredAbility{game.LivingWeaponTriggeredAbility()},
	}}
	return addCombatPermanent(g, controller, def)
}

func TestLivingWeaponCreatesGermAndAttaches(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addLivingWeaponEquipment(g, game.Player1)

	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: equipment.ObjectID})
	agents := [game.NumPlayers]PlayerAgent{}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("living weapon entry trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	var germ *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Germ" {
			germ = permanent
			break
		}
	}
	if germ == nil {
		t.Fatal("living weapon did not create a Germ token")
	}
	if !permanentHasType(g, germ, types.Creature) {
		t.Fatal("Germ token is not a creature")
	}

	if !equipment.AttachedTo.Exists || equipment.AttachedTo.Val != germ.ObjectID {
		t.Fatalf("equipment is not attached to the Germ token: AttachedTo=%v germ=%v", equipment.AttachedTo, germ.ObjectID)
	}
	if !permanentIDsContain(germ.Attachments, equipment.ObjectID) {
		t.Fatal("Germ token does not list the equipment as an attachment")
	}
}
