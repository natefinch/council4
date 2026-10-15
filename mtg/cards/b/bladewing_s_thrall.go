package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BladewingSThrall is the card definition for Bladewing's Thrall.
//
// Type: Creature — Zombie
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	This creature has flying as long as you control a Dragon.
//	When a Dragon enters, you may return this card from your graveyard to the battlefield.
var BladewingSThrall = newBladewingSThrall

func newBladewingSThrall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bladewing's Thrall",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has flying as long as you control a Dragon.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a Dragon",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Dragon}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Flying},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When a Dragon enters, you may return this card from your graveyard to the battlefield.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Dragon}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
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
			This creature has flying as long as you control a Dragon.
			When a Dragon enters, you may return this card from your graveyard to the battlefield.
		`,
		},
	}
}
