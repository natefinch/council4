package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ResoluteRider is the card definition for Resolute Rider.
//
// Type: Creature — Human Knight
// Cost: {W/B}{W/B}{W/B}{W/B}
//
// Oracle text:
//
//	{W/B}{W/B}: This creature gains lifelink until end of turn.
//	{W/B}{W/B}{W/B}: This creature gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it.)
var ResoluteRider = newResoluteRider

func newResoluteRider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Resolute Rider",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W/B}{W/B}: This creature gains lifelink until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.W, mana.B),
						cost.HybridMana(mana.W, mana.B),
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
													AddKeywords: []game.Keyword{game.Lifelink},
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
					Text: "{W/B}{W/B}{W/B}: This creature gains indestructible until end of turn. (Damage and effects that say \"destroy\" don't destroy it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.W, mana.B),
						cost.HybridMana(mana.W, mana.B),
						cost.HybridMana(mana.W, mana.B),
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
													AddKeywords: []game.Keyword{game.Indestructible},
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
			{W/B}{W/B}: This creature gains lifelink until end of turn.
			{W/B}{W/B}{W/B}: This creature gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it.)
		`,
		},
	}
}
