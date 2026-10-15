package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LlanowarEmpath is the card definition for Llanowar Empath.
//
// Type: Creature — Elf Shaman
// Cost: {3}{G}
//
// Oracle text:
//
//	When this creature enters, scry 2, then reveal the top card of your library. If it's a creature card, put it into your hand. (To scry 2, look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)
var LlanowarEmpath = newLlanowarEmpath

func newLlanowarEmpath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Llanowar Empath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, scry 2, then reveal the top card of your library. If it's a creature card, put it into your hand. (To scry 2, look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Reveal{
											Amount:        game.Fixed(1),
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("sequence-effect-1-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-1-product")},
										},
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-1-product",
											},
											FromZone:    zone.Library,
											Destination: zone.Hand,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a creature card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-1-product")),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
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
			When this creature enters, scry 2, then reveal the top card of your library. If it's a creature card, put it into your hand. (To scry 2, look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)
		`,
		},
	}
}
