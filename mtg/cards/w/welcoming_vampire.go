package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WelcomingVampire is the card definition for Welcoming Vampire.
//
// Type: Creature — Vampire
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	Whenever one or more other creatures you control with power 2 or less enter, draw a card. This ability triggers only once each turn.
var WelcomingVampire = newWelcomingVampire

func newWelcomingVampire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Welcoming Vampire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more other creatures you control with power 2 or less enter, draw a card. This ability triggers only once each turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:       game.EventPermanentEnteredBattlefield,
							Controller:  game.TriggerControllerYou,
							ExcludeSelf: true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 2,
							})},
							StackObjectKind: game.StackSpell,
							OneOrMore:       true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Whenever one or more other creatures you control with power 2 or less enter, draw a card. This ability triggers only once each turn.
		`,
		},
	}
}
