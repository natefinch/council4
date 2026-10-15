package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KrarkClanGrunt is the card definition for Krark-Clan Grunt.
//
// Type: Creature — Goblin Warrior
// Cost: {2}{R}
//
// Oracle text:
//
//	Sacrifice an artifact: This creature gets +1/+0 and gains first strike until end of turn.
var KrarkClanGrunt = newKrarkClanGrunt

func newKrarkClanGrunt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Krark-Clan Grunt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice an artifact: This creature gets +1/+0 and gains first strike until end of turn.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.FirstStrike},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Sacrifice an artifact: This creature gets +1/+0 and gains first strike until end of turn.
		`,
		},
	}
}
