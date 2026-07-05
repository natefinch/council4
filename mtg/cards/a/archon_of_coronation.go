package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArchonOfCoronation is the card definition for Archon of Coronation.
//
// Type: Creature — Archon
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Flying
//	When this creature enters, you become the monarch.
//	As long as you're the monarch, damage doesn't cause you to lose life. (When a creature deals combat damage to you, its controller still becomes the monarch.)
var ArchonOfCoronation = newArchonOfCoronation()

func newArchonOfCoronation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Archon of Coronation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Archon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerIsMonarch: true,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDamageDoesntCauseLifeLoss,
							AffectedPlayer: game.PlayerYou,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature enters, you become the monarch.
			As long as you're the monarch, damage doesn't cause you to lose life. (When a creature deals combat damage to you, its controller still becomes the monarch.)
		`,
		},
	}
}
