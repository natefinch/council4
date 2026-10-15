package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StormshriekFeral is the card definition for Stormshriek Feral // Flush Out.
//
// Type: Creature — Dragon // Sorcery — Omen
// Cost: {4}{R} // {1}{R}
// Face: Flush Out — Sorcery — Omen ({1}{R})
//
// Oracle text:
//
//	Flying, haste
//	{1}{R}: This creature gets +1/+0 until end of turn.
var StormshriekFeral = newStormshriekFeral

func newStormshriekFeral() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Stormshriek Feral",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{R}: This creature gets +1/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
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
			Flying, haste
			{1}{R}: This creature gets +1/+0 until end of turn.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Flush Out",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Omen},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						PublishResult: game.ResultKey("if-you-do"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("if-you-do")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "if-you-do",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Discard a card. If you do, draw two cards. (Then shuffle this card into its owner's library.)
		`,
		}),
	}
}
