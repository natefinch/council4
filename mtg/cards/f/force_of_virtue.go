package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ForceOfVirtue is the card definition for Force of Virtue.
//
// Type: Enchantment
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	If it's not your turn, you may exile a white card from your hand rather than pay this spell's mana cost.
//	Flash
//	Creatures you control get +1/+1.
var ForceOfVirtue = newForceOfVirtue

func newForceOfVirtue() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Force of Virtue",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "Creatures you control get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a white card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.White,
						},
					},
					Condition: cost.AlternativeConditionNotYourTurn,
				},
			},
			OracleText: `
			If it's not your turn, you may exile a white card from your hand rather than pay this spell's mana cost.
			Flash
			Creatures you control get +1/+1.
		`,
		},
	}
}
