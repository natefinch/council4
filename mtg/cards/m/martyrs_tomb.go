package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MartyrsTomb is the card definition for Martyrs' Tomb.
//
// Type: Enchantment
// Cost: {2}{W}{B}
//
// Oracle text:
//
//	Pay 2 life: Prevent the next 1 damage that would be dealt to target creature this turn.
var MartyrsTomb = newMartyrsTomb

func newMartyrsTomb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Martyrs' Tomb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Pay 2 life: Prevent the next 1 damage that would be dealt to target creature this turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "Pay 2 life",
							Amount: 2,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Pay 2 life: Prevent the next 1 damage that would be dealt to target creature this turn.
		`,
		},
	}
}
