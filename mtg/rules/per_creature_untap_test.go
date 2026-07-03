package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// perCreatureUntapContent mirrors the lowering of the "each player's upkeep,
// that player may choose any number of tapped <filter> creatures they control
// and pay {N} for each creature chosen this way. If the player does, untap those
// creatures." resolution (Dream Tides, Magnetic Mountain, Thelon's Curse). The
// upkeep player pays {2} per creature and chooses that many of their own tapped
// nongreen creatures to untap. It exercises the event-player PayRepeatedly payer
// together with the Untap event-player chooser.
func perCreatureUntapContent() game.AbilityContent {
	const countKey = game.ResultKey("per-creature-untap-count")
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.PayRepeatedly{
					Payment: game.ResolutionPayment{
						ManaCost: opt.Val(cost.Mana{cost.O(2)}),
						Payer:    opt.Val(game.EventPlayerReference()),
					},
					PublishCount: countKey,
					Prompt:       "Pay {2}?",
				},
				PublishResult: countKey,
			},
			{
				Primitive: game.Untap{
					Group: game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{
						RequiredTypes:  []types.Card{types.Creature},
						Tapped:         game.TriTrue,
						ExcludedColors: []color.Color{color.Green},
					}),
					ChooseUpTo: true,
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:      game.DynamicAmountChosenNumber,
						ResultKey: countKey,
					}),
					Chooser: game.EventPlayerReference(),
				},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       countKey,
					Succeeded: game.TriTrue,
				}),
			},
		},
	}.Ability()
}

func perCreatureUntapCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Dream Tides",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event: game.EventBeginningOfStep,
					Step:  game.StepUpkeep,
				},
			},
			Content: perCreatureUntapContent(),
		}},
	}}
}

// TestPerCreatureUntapUntapsUpkeepPlayerCreatures proves the upkeep player pays
// {2} for each of their tapped nongreen creatures and untaps them, even though
// the enchantment is controlled by a different player. Player1 controls the
// enchantment; on Player2's upkeep Player2 funds two payments with four Islands
// and untaps both blue creatures they control.
func TestPerCreatureUntapUntapsUpkeepPlayerCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	if issues := game.ValidateCardDef(perCreatureUntapCard()); len(issues) != 0 {
		t.Fatalf("carddef invalid: %v", issues)
	}
	addCombatPermanent(g, game.Player1, perCreatureUntapCard())
	blue1 := addColoredCreaturePermanent(g, game.Player2, 2, color.Blue)
	blue2 := addColoredCreaturePermanent(g, game.Player2, 2, color.Blue)
	setPermanentTapped(g, blue1, true)
	setPermanentTapped(g, blue2, true)
	for range 4 {
		addBasicLandPermanent(g, game.Player2, types.Island)
	}

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("per-creature untap trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if blue1.Tapped {
		t.Error("blue1 was not untapped")
	}
	if blue2.Tapped {
		t.Error("blue2 was not untapped")
	}
}

// TestPerCreatureUntapLeavesCreaturesTappedWhenUpkeepPlayerCannotPay proves the
// gate is fail-closed: with no mana the upkeep player pays nothing and their
// tapped creature stays tapped.
func TestPerCreatureUntapLeavesCreaturesTappedWhenUpkeepPlayerCannotPay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, perCreatureUntapCard())
	blue := addColoredCreaturePermanent(g, game.Player2, 2, color.Blue)
	setPermanentTapped(g, blue, true)

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("per-creature untap trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !blue.Tapped {
		t.Error("blue creature was untapped without payment")
	}
}

// TestPerCreatureUntapExcludesFilteredCreatures proves the folded selection is
// honored: a tapped green creature is outside the "nongreen" filter, so it
// remains tapped even when the upkeep player has ample mana.
func TestPerCreatureUntapExcludesFilteredCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, perCreatureUntapCard())
	green := addColoredCreaturePermanent(g, game.Player2, 2, color.Green)
	setPermanentTapped(g, green, true)
	for range 4 {
		addBasicLandPermanent(g, game.Player2, types.Island)
	}

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("per-creature untap trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !green.Tapped {
		t.Error("green creature was untapped despite the nongreen filter")
	}
}
