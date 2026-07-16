package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MorkrutBehemoth is the card definition for Morkrut Behemoth.
//
// Type: Creature — Zombie Giant
// Cost: {4}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature or pay {1}{B}.
//	Menace (This creature can't be blocked except by two or more creatures.)
var MorkrutBehemoth = newMorkrutBehemoth

func newMorkrutBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Morkrut Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Giant},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Sacrifice a creature",
							Costs: []cost.Additional{
								{
									Kind:               cost.AdditionalSacrifice,
									Text:               "sacrifice a creature",
									Amount:             1,
									MatchPermanentType: true,
									PermanentType:      types.Creature,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {1}{B}",
							Mana:  cost.Mana{cost.O(1), cost.B},
						},
					},
				},
			},
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature or pay {1}{B}.
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
