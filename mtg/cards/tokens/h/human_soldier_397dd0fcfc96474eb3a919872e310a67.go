package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HumanSoldier is the card definition for Human Soldier.
//
// Type: Token Creature — Human Soldier
//
// Oracle text:
//   Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)

// HumanSoldierToken397dd0fcfc96474eb3a919872e310a67 is the card definition for Human Soldier.
var HumanSoldierToken397dd0fcfc96474eb3a919872e310a67 = newHumanSoldierToken397dd0fcfc96474eb3a919872e310a67()

func newHumanSoldierToken397dd0fcfc96474eb3a919872e310a67() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TrainingTriggeredBody,
			},
			OracleText: `
			Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
		`,
		},
	}
}
