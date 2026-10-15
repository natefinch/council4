package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeafeningClarion is the card definition for Deafening Clarion.
//
// Type: Sorcery
// Cost: {1}{R}{W}
//
// Oracle text:
//
//	Choose one or both —
//	• Deafening Clarion deals 3 damage to each creature.
//	• Creatures you control gain lifelink until end of turn.
var DeafeningClarion = newDeafeningClarion

func newDeafeningClarion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Deafening Clarion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Deafening Clarion deals 3 damage to each creature.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control gain lifelink until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.Lifelink},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Choose one or both —
			• Deafening Clarion deals 3 damage to each creature.
			• Creatures you control gain lifelink until end of turn.
		`,
		},
	}
}
