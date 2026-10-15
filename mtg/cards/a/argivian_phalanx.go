package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArgivianPhalanx is the card definition for Argivian Phalanx.
//
// Type: Creature — Human Kor Soldier
// Cost: {5}{W}
//
// Oracle text:
//
//	Affinity for creatures (This spell costs {1} less to cast for each creature you control.)
//	Vigilance (Attacking doesn't cause this creature to tap.)
var ArgivianPhalanx = newArgivianPhalanx

func newArgivianPhalanx() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Argivian Phalanx",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Kor, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each creature you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.VigilanceStaticBody,
			},
			OracleText: `
			Affinity for creatures (This spell costs {1} less to cast for each creature you control.)
			Vigilance (Attacking doesn't cause this creature to tap.)
		`,
		},
	}
}
