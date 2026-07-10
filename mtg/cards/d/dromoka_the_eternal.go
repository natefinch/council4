package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DromokaTheEternal is the card definition for Dromoka, the Eternal.
//
// Type: Legendary Creature — Dragon
// Cost: {3}{G}{W}
//
// Oracle text:
//
//	Flying
//	Whenever a Dragon you control attacks, bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
var DromokaTheEternal = newDromokaTheEternal

func newDromokaTheEternal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Dromoka, the Eternal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Dragon")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bolster{
									Amount: game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever a Dragon you control attacks, bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
		`,
		},
	}
}
