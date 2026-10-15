package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JayaVeneratedFiremage is the card definition for Jaya, Venerated Firemage.
//
// Type: Legendary Planeswalker — Jaya
// Cost: {4}{R}
//
// Oracle text:
//
//	If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.
//	−2: Jaya deals 2 damage to any target.
var JayaVeneratedFiremage = newJayaVeneratedFiremage

func newJayaVeneratedFiremage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Jaya, Venerated Firemage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Jaya},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "−2: Jaya deals 2 damage to any target.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "any target",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementExcludingSource("If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.", 0, 1, []color.Color{color.Red}, game.TriggerControllerYou),
			},
			OracleText: `
			If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.
			−2: Jaya deals 2 damage to any target.
		`,
		},
	}
}
