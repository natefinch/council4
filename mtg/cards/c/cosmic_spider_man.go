package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CosmicSpiderMan is the card definition for Cosmic Spider-Man.
//
// Type: Legendary Creature — Spider Human Hero
// Cost: {W}{U}{B}{R}{G}
//
// Oracle text:
//
//	Flying, first strike, trample, lifelink, haste
//	At the beginning of combat on your turn, other Spiders you control gain flying, first strike, trample, lifelink, and haste until end of turn.
var CosmicSpiderMan = newCosmicSpiderMan()

func newCosmicSpiderMan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Cosmic Spider-Man",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spider, types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Spider")}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											AddKeywords: []game.Keyword{
												game.Flying,
												game.FirstStrike,
												game.Trample,
												game.Lifelink,
												game.Haste,
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
			Flying, first strike, trample, lifelink, haste
			At the beginning of combat on your turn, other Spiders you control gain flying, first strike, trample, lifelink, and haste until end of turn.
		`,
		},
	}
}
