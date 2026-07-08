package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TwoHeadedDragon is the card definition for Two-Headed Dragon.
//
// Type: Creature — Dragon
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Flying
//	Menace (This creature can't be blocked except by two or more creatures.)
//	This creature can block an additional creature each combat.
//	{1}{R}: This creature gets +2/+0 until end of turn.
var TwoHeadedDragon = newTwoHeadedDragon

func newTwoHeadedDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Two-Headed Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.MenaceStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{R}: This creature gets +2/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Menace (This creature can't be blocked except by two or more creatures.)
			This creature can block an additional creature each combat.
			{1}{R}: This creature gets +2/+0 until end of turn.
		`,
		},
	}
}
