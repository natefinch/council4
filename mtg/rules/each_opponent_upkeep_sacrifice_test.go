package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// eachOpponentUpkeepSacrifice is the generic "At the beginning of each
// opponent's upkeep, that player sacrifices a creature of their choice."
// trigger (Sheoldred, Whispering One): a TriggerAt beginning-of-upkeep pattern
// scoped to opponents, whose edict targets the player named by the triggering
// event (EventPlayerReference), i.e. the player whose upkeep it is.
func eachOpponentUpkeepSacrifice() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Sheoldred",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:  game.EventBeginningOfStep,
					Step:   game.StepUpkeep,
					Player: game.TriggerPlayerOpponent,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.SacrificePermanents{
					Player:    game.EventPlayerReference(),
					Amount:    game.Fixed(1),
					Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				},
			}}}.Ability(),
		}},
	}}
}

// TestEachOpponentUpkeepSacrificeFiresOnOpponentUpkeep proves the each-opponent
// upkeep edict forces the player whose upkeep it is to sacrifice a creature.
// Player1 controls the source; on Player2's upkeep the trigger fires and
// Player2 sacrifices a creature of their choice.
func TestEachOpponentUpkeepSacrificeFiresOnOpponentUpkeep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, eachOpponentUpkeepSacrifice())
	creature1 := addCreaturePermanent(g, game.Player2)
	creature2 := addCreaturePermanent(g, game.Player2)

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("each-opponent upkeep sacrifice trigger was not put on the stack")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &sacrificeChoiceAgent{t: t, g: g, choice: []int{1}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := permanentByObjectID(g, creature1.ObjectID); !ok {
		t.Fatal("unchosen creature was removed from battlefield")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("chosen creature was not sacrificed")
	}
}

// TestEachOpponentUpkeepSacrificeSkipsControllerUpkeep proves the opponent
// scope: on the source controller's own upkeep the trigger does not fire, so
// the controller never sacrifices.
func TestEachOpponentUpkeepSacrificeSkipsControllerUpkeep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, eachOpponentUpkeepSacrifice())
	ownCreature := addCreaturePermanent(g, game.Player1)

	g.Turn.ActivePlayer = game.Player1
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired on the controller's own upkeep, want opponent-only")
	}
	if _, ok := permanentByObjectID(g, ownCreature.ObjectID); !ok {
		t.Fatal("controller's creature was sacrificed on their own upkeep")
	}
}
