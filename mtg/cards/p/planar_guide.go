package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PlanarGuide is the card definition for Planar Guide.
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//
//	{3}{W}, Exile this creature: Exile all creatures. At the beginning of the next end step, return those cards to the battlefield under their owners' control.
var PlanarGuide = newPlanarGuide()

func newPlanarGuide() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Planar Guide",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}{W}, Exile this creature: Exile all creatures. At the beginning of the next end step, return those cards to the battlefield under their owners' control.",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalExileSource,
							Text:   "Exile this creature",
							Amount: 1,
							Source: zone.Battlefield,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									ExileLinkedKey: game.LinkedKey("group-blink"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.PutOnBattlefield{
														Source: game.LinkedBattlefieldSource(game.LinkedKey("group-blink")),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{3}{W}, Exile this creature: Exile all creatures. At the beginning of the next end step, return those cards to the battlefield under their owners' control.
		`,
		},
	}
}
