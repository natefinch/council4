package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BubblingMuck is the card definition for Bubbling Muck.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Until end of turn, whenever a player taps a Swamp for mana, that player adds an additional {B}.
var BubblingMuck = newBubblingMuck

func newBubblingMuck() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bubbling Muck",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.AddMana{
														Amount:    game.Fixed(1),
														ManaColor: mana.B,
														Player:    opt.Val(game.EventPlayerReference()),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								EventPattern: opt.Val(game.TriggerPattern{
									Event:                game.EventManaProduced,
									SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{types.Swamp}},
									StackObjectKind:      game.StackSpell,
									CounterKind:          counter.PlusOnePlusOne,
									RequireTappedForMana: true,
								}),
								Window: game.DelayedWindowThisTurn,
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, whenever a player taps a Swamp for mana, that player adds an additional {B}.
		`,
		},
	}
}
