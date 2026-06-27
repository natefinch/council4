package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheHowlingCommandos is the card definition for The Howling Commandos.
//
// Type: Legendary Creature — Human Soldier Hero
// Cost: {1}{W}
//
// Oracle text:
//
//	{5}: Creatures you control get +1/+1 until end of turn. Soldiers you control gain vigilance until end of turn.
var TheHowlingCommandos = newTheHowlingCommandos()

func newTheHowlingCommandos() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "The Howling Commandos",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{5}: Creatures you control get +1/+1 until end of turn. Soldiers you control gain vigilance until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(5)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Soldier")}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Vigilance,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{5}: Creatures you control get +1/+1 until end of turn. Soldiers you control gain vigilance until end of turn.
		`,
		},
	}
}
