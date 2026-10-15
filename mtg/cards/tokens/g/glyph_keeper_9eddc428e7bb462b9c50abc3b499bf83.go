package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlyphKeeper is the card definition for Glyph Keeper.
//
// Type: Token Creature — Zombie Sphinx
//
// Oracle text:
//   Flying
//   Whenever Glyph Keeper becomes the target of a spell or ability for the first time each turn, counter that spell or ability.

// GlyphKeeperToken9eddc428e7bb462b9c50abc3b499bf83 is the card definition for Glyph Keeper.
var GlyphKeeperToken9eddc428e7bb462b9c50abc3b499bf83 = newGlyphKeeperToken9eddc428e7bb462b9c50abc3b499bf83()

func newGlyphKeeperToken9eddc428e7bb462b9c50abc3b499bf83() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Glyph Keeper",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Sphinx},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Glyph Keeper becomes the target of a spell or ability for the first time each turn, counter that spell or ability.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventObjectBecameTarget,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CounterObject{
											Object: game.EventStackObjectReference(),
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
			Whenever Glyph Keeper becomes the target of a spell or ability for the first time each turn, counter that spell or ability.
		`,
		},
	}
}
