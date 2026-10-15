package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AshiokSSkulker is the card definition for Ashiok's Skulker.
//
// Type: Creature — Nightmare
// Cost: {4}{U}
//
// Oracle text:
//
//	{3}{U}: This creature can't be blocked this turn.
var AshiokSSkulker = newAshiokSSkulker

func newAshiokSSkulker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ashiok's Skulker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nightmare},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{U}: This creature can't be blocked this turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
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
			{3}{U}: This creature can't be blocked this turn.
		`,
		},
	}
}
