package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HexingSquelcher is the card definition for Hexing Squelcher.
//
// Type: Creature — Goblin Sorcerer
// Cost: {1}{R}
//
// Oracle text:
//
//	This spell can't be countered.
//	Ward—Pay 2 life.
//	Spells you control can't be countered.
//	Other creatures you control have "Ward—Pay 2 life."
var HexingSquelcher = newHexingSquelcher

func newHexingSquelcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hexing Squelcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Sorcerer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalPayLife,
						Text:   "Pay 2 life",
						Amount: 2,
					},
				}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeCountered,
							AffectedController: game.ControllerYou,
						},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
									{
										Kind:   cost.AdditionalPayLife,
										Text:   "Pay 2 life",
										Amount: 2,
									},
								})),
							},
						},
					},
				},
			},
			OracleText: `
			This spell can't be countered.
			Ward—Pay 2 life.
			Spells you control can't be countered.
			Other creatures you control have "Ward—Pay 2 life."
		`,
		},
	}
}
