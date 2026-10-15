package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RuralRecruit is the card definition for Rural Recruit.
//
// Type: Creature — Human Peasant
// Cost: {3}{G}
//
// Oracle text:
//
//	Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
//	When this creature enters, create a 3/1 green Boar creature token.
var RuralRecruit = newRuralRecruit

func newRuralRecruit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rural Recruit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Peasant},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TrainingTriggeredBody,
				game.TriggeredAbility{
					Text: "When this creature enters, create a 3/1 green Boar creature token.",
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
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(ruralRecruitToken),
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
			Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
			When this creature enters, create a 3/1 green Boar creature token.
		`,
		},
	}
}

var ruralRecruitToken = newRuralRecruitToken()

func newRuralRecruitToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Boar",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Boar},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
