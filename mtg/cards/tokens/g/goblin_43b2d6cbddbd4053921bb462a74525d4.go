package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin is the card definition for Goblin.
//
// Type: Token Creature — Goblin
//
// Oracle text:
//   Creatures you control attack each combat if able.

// GoblinToken43b2d6cbddbd4053921bb462a74525d4 is the card definition for Goblin.
var GoblinToken43b2d6cbddbd4053921bb462a74525d4 = newGoblinToken43b2d6cbddbd4053921bb462a74525d4()

func newGoblinToken43b2d6cbddbd4053921bb462a74525d4() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control attack each combat if able.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectMustAttack,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			Creatures you control attack each combat if able.
		`,
		},
	}
}
