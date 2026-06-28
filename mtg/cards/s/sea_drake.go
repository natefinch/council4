package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeaDrake is the card definition for Sea Drake.
//
// Type: Creature — Drake
// Cost: {2}{U}
//
// Oracle text:
//
//	Flying
//	When this creature enters, return two target lands you control to their owner's hand.
var SeaDrake = newSeaDrake()

func newSeaDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sea Drake",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 2,
								MaxTargets: 2,
								Constraint: "two target lands you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature enters, return two target lands you control to their owner's hand.
		`,
		},
	}
}
