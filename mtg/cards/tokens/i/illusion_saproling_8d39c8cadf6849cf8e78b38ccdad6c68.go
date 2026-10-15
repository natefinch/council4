package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Illusion is the card definition for Illusion // Saproling.
//
// Type: Token Creature — Illusion // Token Creature — Saproling
// Face: Saproling — Token Creature — Saproling
//
// Oracle text:
//   When this creature becomes the target of a spell, sacrifice it.

// IllusionToken8d39c8cadf6849cf8e78b38ccdad6c68 is the card definition for Illusion.
var IllusionToken8d39c8cadf6849cf8e78b38ccdad6c68 = newIllusionToken8d39c8cadf6849cf8e78b38ccdad6c68()

func newIllusionToken8d39c8cadf6849cf8e78b38ccdad6c68() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:      "Illusion",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature becomes the target of a spell, sacrifice it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:                game.EventObjectBecameTarget,
							Source:               game.TriggerSourceSelf,
							MatchStackObjectKind: true,
							StackObjectKind:      game.StackSpell,
							CounterKind:          counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Sacrifice{
											Object: game.EventPermanentReference(),
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
			When this creature becomes the target of a spell, sacrifice it.
		`,
		},
		Layout: game.LayoutDoubleFacedToken,
		Back: opt.Val(game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		}),
	}
}
