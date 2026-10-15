package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhythmOfTheWild is the card definition for Rhythm of the Wild.
//
// Type: Enchantment
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Creature spells you control can't be countered.
//	Nontoken creatures you control have riot. (They enter with your choice of a +1/+1 counter or haste.)
var RhythmOfTheWild = newRhythmOfTheWild

func newRhythmOfTheWild() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Rhythm of the Wild",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature spells you control can't be countered.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeCountered,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							SpellTypes:         []types.Card{types.Creature},
						},
					},
				},
				game.StaticAbility{
					Text: "Nontoken creatures you control have riot. (They enter with your choice of a +1/+1 counter or haste.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Riot},
						},
					},
				},
			},
			OracleText: `
			Creature spells you control can't be countered.
			Nontoken creatures you control have riot. (They enter with your choice of a +1/+1 counter or haste.)
		`,
		},
	}
}
