package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThaliaTheSurvivor is the card definition for Thalia, the Survivor.
//
// Type: Legendary Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	Lifelink
//	Noncreature spells your opponents cast cost {1} more to cast.
var ThaliaTheSurvivor = newThaliaTheSurvivor

func newThaliaTheSurvivor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Thalia, the Survivor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Noncreature spells your opponents cast cost {1} more to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerOpponent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 1,
								CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							},
						},
					},
				},
			},
			OracleText: `
			Lifelink
			Noncreature spells your opponents cast cost {1} more to cast.
		`,
		},
	}
}
