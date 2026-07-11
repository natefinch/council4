package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GaiusVanBaelsar is the card definition for Gaius van Baelsar.
//
// Type: Legendary Creature — Human Soldier
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	When Gaius van Baelsar enters, choose one —
//	• Each player sacrifices a creature token of their choice.
//	• Each player sacrifices a nontoken creature of their choice.
//	• Each player sacrifices an enchantment of their choice.
var GaiusVanBaelsar = newGaiusVanBaelsar

func newGaiusVanBaelsar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gaius van Baelsar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Each player sacrifices a creature token of their choice.",
								Sequence: []game.Instruction{
									{
										Primitive: game.SacrificePermanents{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
										},
									},
								},
							},
							game.Mode{
								Text: "Each player sacrifices a nontoken creature of their choice.",
								Sequence: []game.Instruction{
									{
										Primitive: game.SacrificePermanents{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
										},
									},
								},
							},
							game.Mode{
								Text: "Each player sacrifices an enchantment of their choice.",
								Sequence: []game.Instruction{
									{
										Primitive: game.SacrificePermanents{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
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
			OracleText: `
			When Gaius van Baelsar enters, choose one —
			• Each player sacrifices a creature token of their choice.
			• Each player sacrifices a nontoken creature of their choice.
			• Each player sacrifices an enchantment of their choice.
		`,
		},
	}
}
