package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ViciousShadows is the card definition for Vicious Shadows.
//
// Type: Enchantment
// Cost: {6}{R}
//
// Oracle text:
//
//	Whenever a creature dies, you may have this enchantment deal damage to target player equal to the number of cards in that player's hand.
var ViciousShadows = newViciousShadows()

func newViciousShadows() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Vicious Shadows",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.TargetPlayerReference(0); return &ref }(),
										CardZone:   zone.Hand,
										Selection:  &game.Selection{},
									}),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature dies, you may have this enchantment deal damage to target player equal to the number of cards in that player's hand.
		`,
		},
	}
}
