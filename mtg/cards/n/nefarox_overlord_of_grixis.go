package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NefaroxOverlordOfGrixis is the card definition for Nefarox, Overlord of Grixis.
//
// Type: Legendary Creature — Demon
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Flying
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	Whenever Nefarox attacks alone, defending player sacrifices a creature of their choice.
var NefaroxOverlordOfGrixis = newNefaroxOverlordOfGrixis

func newNefaroxOverlordOfGrixis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nefarox, Overlord of Grixis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ExaltedStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Nefarox attacks alone, defending player sacrifices a creature of their choice.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							AttackAlone:     true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:    game.DefendingPlayerReference(),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			Flying
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
			Whenever Nefarox attacks alone, defending player sacrifices a creature of their choice.
		`,
		},
	}
}
