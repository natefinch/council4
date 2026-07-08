package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChitteringSkitterling is the card definition for Chittering Skitterling.
//
// Type: Creature — Phyrexian Rat
// Cost: {2}{B}
//
// Oracle text:
//
//	Corrupted — Sacrifice an artifact or creature: Draw a card. Activate only if an opponent has three or more poison counters and only once each turn.
var ChitteringSkitterling = newChitteringSkitterling

func newChitteringSkitterling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Chittering Skitterling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Rat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Corrupted — Sacrifice an artifact or creature: Draw a card. Activate only if an opponent has three or more poison counters and only once each turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an artifact or creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							PermanentTypeAlt:   types.Creature,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					ActivationCondition: opt.Val(game.Condition{
						AnyOpponentPoisonAtLeast: 3,
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Corrupted — Sacrifice an artifact or creature: Draw a card. Activate only if an opponent has three or more poison counters and only once each turn.
		`,
		},
	}
}
