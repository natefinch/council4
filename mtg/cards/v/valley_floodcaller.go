package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ValleyFloodcaller is the card definition for Valley Floodcaller.
//
// Type: Creature — Otter Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	Flash
//	You may cast noncreature spells as though they had flash.
//	Whenever you cast a noncreature spell, Birds, Frogs, Otters, and Rats you control get +1/+1 until end of turn. Untap them.
var ValleyFloodcaller = newValleyFloodcaller

func newValleyFloodcaller() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Valley Floodcaller",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Otter, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCastSpellsAsThoughFlash,
							AffectedPlayer:     game.PlayerYou,
							ExcludedSpellTypes: []types.Card{types.Creature},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Bird"), types.Sub("Frog"), types.Sub("Otter"), types.Sub("Rat")}, Controller: game.ControllerYou}),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.Untap{
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Bird"), types.Sub("Frog"), types.Sub("Otter"), types.Sub("Rat")}, Controller: game.ControllerYou}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			You may cast noncreature spells as though they had flash.
			Whenever you cast a noncreature spell, Birds, Frogs, Otters, and Rats you control get +1/+1 until end of turn. Untap them.
		`,
		},
	}
}
