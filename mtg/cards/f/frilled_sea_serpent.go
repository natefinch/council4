package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FrilledSeaSerpent is the card definition for Frilled Sea Serpent.
//
// Type: Creature — Serpent
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	{5}{U}{U}: This creature can't be blocked this turn.
var FrilledSeaSerpent = newFrilledSeaSerpent

func newFrilledSeaSerpent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Frilled Sea Serpent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Serpent},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{U}{U}: This creature can't be blocked this turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.U,
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											Object: opt.Val(game.SourcePermanentReference()),
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:     game.RuleEffectCantBeBlocked,
													Duration: game.DurationPermanent,
												},
											},
											Duration: game.DurationThisTurn,
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
			{5}{U}{U}: This creature can't be blocked this turn.
		`,
		},
	}
}
