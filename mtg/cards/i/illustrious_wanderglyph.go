package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IllustriousWanderglyph is the card definition for Illustrious Wanderglyph.
//
// Type: Artifact Creature — Golem
// Cost: {4}{W}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	Other artifact creatures you control get +2/+2 as long as you have the city's blessing.
//	At the beginning of each upkeep, create a 1/1 colorless Gnome artifact creature token.
var IllustriousWanderglyph = newIllustriousWanderglyph

func newIllustriousWanderglyph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Illustrious Wanderglyph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.AscendStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerHasCityBlessing: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(illustriousWanderglyphToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			Other artifact creatures you control get +2/+2 as long as you have the city's blessing.
			At the beginning of each upkeep, create a 1/1 colorless Gnome artifact creature token.
		`,
		},
	}
}

var illustriousWanderglyphToken = newIllustriousWanderglyphToken()

func newIllustriousWanderglyphToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Gnome",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gnome},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
