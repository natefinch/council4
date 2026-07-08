package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QueenMarchesa is the card definition for Queen Marchesa.
//
// Type: Legendary Creature — Human Assassin
// Cost: {1}{R}{W}{B}
//
// Oracle text:
//
//	Deathtouch, haste
//	When Queen Marchesa enters, you become the monarch.
//	At the beginning of your upkeep, if an opponent is the monarch, create a 1/1 black Assassin creature token with deathtouch and haste.
var QueenMarchesa = newQueenMarchesa

func newQueenMarchesa() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Queen Marchesa",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Assassin},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if an opponent is the monarch",
						InterveningCondition: opt.Val(game.Condition{
							AnOpponentIsMonarch: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(queenMarchesaToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Deathtouch, haste
			When Queen Marchesa enters, you become the monarch.
			At the beginning of your upkeep, if an opponent is the monarch, create a 1/1 black Assassin creature token with deathtouch and haste.
		`,
		},
	}
}

var queenMarchesaToken = newQueenMarchesaToken()

func newQueenMarchesaToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Assassin",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Assassin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
