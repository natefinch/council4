package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DanithaCapashenParagon is the card definition for Danitha Capashen, Paragon.
//
// Type: Legendary Creature — Human Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	First strike, vigilance, lifelink
//	Aura and Equipment spells you cast cost {1} less to cast.
var DanithaCapashenParagon = newDanithaCapashenParagon

func newDanithaCapashenParagon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Danitha Capashen, Paragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Knight},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Aura and Equipment spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Aura, types.Equipment}},
							},
						},
					},
				},
			},
			OracleText: `
			First strike, vigilance, lifelink
			Aura and Equipment spells you cast cost {1} less to cast.
		`,
		},
	}
}
