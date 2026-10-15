package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Hackrobat is the card definition for Hackrobat.
//
// Type: Creature — Human Rogue
// Cost: {1}{B}{R}
//
// Oracle text:
//
//	Spectacle {B}{R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
//	{B}: This creature gains deathtouch until end of turn.
//	{R}: This creature gets +2/-2 until end of turn.
var Hackrobat = newHackrobat

func newHackrobat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Hackrobat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}: This creature gains deathtouch until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Deathtouch},
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
				game.ActivatedAbility{
					Text: "{R}: This creature gets +2/-2 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(-2),
											Duration:       game.DurationUntilEndOfTurn,
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
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Spectacle",
					ManaCost:  opt.Val(cost.Mana{cost.B, cost.R}),
					Condition: cost.AlternativeConditionOpponentLostLifeThisTurn,
				},
			},
			OracleText: `
			Spectacle {B}{R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
			{B}: This creature gains deathtouch until end of turn.
			{R}: This creature gets +2/-2 until end of turn.
		`,
		},
	}
}
