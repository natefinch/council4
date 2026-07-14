package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestTriggeringEventTotalPowerSizesTokenFromMatchingDeathBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Fungus Dinosaur",
		Colors:   []color.Color{color.Green},
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Fungus, types.Dinosaur},
	}}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "The Skullspore Nexus",
		Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event:      game.EventPermanentDied,
				Controller: game.TriggerControllerYou,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					NonToken:      true,
				},
				OneOrMore: true,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
				Amount:    game.Fixed(1),
				Source:    game.TokenDef(tokenDef),
				Power:     opt.Val(game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountTriggeringEventTotalPower})),
				Toughness: opt.Val(game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountTriggeringEventTotalPower})),
			}}}}.Ability(),
		}},
	}})
	addCombatPermanent(g, game.Player1, poweredCreatureDef("Three", 3))
	addCombatPermanent(g, game.Player1, poweredCreatureDef("Four", 4))
	addCombatTokenCreaturePermanent(g, game.Player1, 10)
	addCombatPermanent(g, game.Player2, poweredCreatureDef("Twenty", 20))

	resolveInstruction(engine, g, &game.StackObject{
		ID:         g.IDGen.Next(),
		SourceID:   g.IDGen.Next(),
		Controller: game.Player1,
	}, game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
	})}, &TurnLog{})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("death batch did not put the Skullspore trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	var created *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == source.ObjectID {
			continue
		}
		if def, ok := permanentCardDef(g, permanent); ok && def.Name == "Fungus Dinosaur" {
			created = permanent
			break
		}
	}
	if created == nil {
		t.Fatal("Fungus Dinosaur token was not created")
	}
	if power := effectivePower(g, created); power != 7 {
		t.Fatalf("token power = %d, want 7", power)
	}
	if toughness, ok := effectiveToughness(g, created); !ok || toughness != 7 {
		t.Fatalf("token toughness = %d, %v, want 7, true", toughness, ok)
	}
}

func poweredCreatureDef(name string, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}}
}
