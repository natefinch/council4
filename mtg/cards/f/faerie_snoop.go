package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FaerieSnoop is the card definition for Faerie Snoop.
//
// Type: Creature — Faerie Detective
// Cost: {1}{U}{B}
//
// Oracle text:
//
//	Flying
//	Disguise {1}{U/B}{U/B} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
//	When this creature is turned face up, look at the top two cards of your library. Put one into your hand and the other into your graveyard.
var FaerieSnoop = newFaerieSnoop

func newFaerieSnoop() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Faerie Snoop",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Detective},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Disguise {1}{U/B}{U/B}",
					KeywordAbilities: []game.KeywordAbility{
						game.DisguiseKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.HybridMana(mana.U, mana.B),
								cost.HybridMana(mana.U, mana.B),
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature is turned face up, look at the top two cards of your library. Put one into your hand and the other into your graveyard.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentTurnedFaceUp,
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
										Primitive: game.Dig{
											Player: game.ControllerReference(),
											Look:   game.Fixed(2),
											Take:   game.Fixed(1),
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
			Disguise {1}{U/B}{U/B} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
			When this creature is turned face up, look at the top two cards of your library. Put one into your hand and the other into your graveyard.
		`,
		},
	}
}
