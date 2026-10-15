package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PyreRhymer is the card definition for Pyre Rhymer // Molten Tide.
//
// Type: Creature — Elemental Sorcerer // Instant
// Cost: {1}{R}{R} // {R}
// Face: Molten Tide — Instant ({R})
//
// Oracle text:
//
//	Prowess
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var PyreRhymer = newPyreRhymer

func newPyreRhymer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pyre Rhymer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:         []color.Color{color.Red},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Elemental, types.Sorcerer},
			Power:          opt.Val(game.PT{Value: 3}),
			Toughness:      opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
			},
			OracleText: `
			Prowess
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Molten Tide",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
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
														ManaColor: mana.R,
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								EventPattern: opt.Val(game.TriggerPattern{
									Event:                game.EventPermanentTapped,
									Controller:           game.TriggerControllerYou,
									SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{types.Mountain}},
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
			Until end of turn, whenever you tap a Mountain for mana, add an additional {R}.
		`,
		}),
	}
}
