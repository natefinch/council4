package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TajicBladeOfTheLegion is the card definition for Tajic, Blade of the Legion.
//
// Type: Legendary Creature — Human Soldier
// Cost: {2}{R}{W}
//
// Oracle text:
//
//	Indestructible
//	Battalion — Whenever Tajic and at least two other creatures attack, Tajic gets +5/+5 until end of turn.
var TajicBladeOfTheLegion = newTajicBladeOfTheLegion

func newTajicBladeOfTheLegion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Tajic, Blade of the Legion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Battalion — Whenever Tajic and at least two other creatures attack, Tajic gets +5/+5 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventAttackerDeclared,
							Source:               game.TriggerSourceSelf,
							StackObjectKind:      game.StackSpell,
							AttackerCountAtLeast: 3,
							CounterKind:          counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(5),
											ToughnessDelta: game.Fixed(5),
											Duration:       game.DurationUntilEndOfTurn,
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
			Indestructible
			Battalion — Whenever Tajic and at least two other creatures attack, Tajic gets +5/+5 until end of turn.
		`,
		},
	}
}
