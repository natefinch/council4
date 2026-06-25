package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LordDreggInsectInvader is the card definition for Lord Dregg, Insect Invader.
//
// Type: Legendary Creature — Insect Warrior
// Cost: {3}{B}
//
// Oracle text:
//
//	Flying
//	Disappear — At the beginning of your end step, if a permanent left the battlefield under your control this turn, create a 1/1 black Insect Warrior creature token with flying.
//	{3}{G}, Sacrifice a token: Draw a card.
var LordDreggInsectInvader = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name: "Lord Dregg, Insect Invader",
		ManaCost: opt.Val(cost.Mana{
			cost.O(3),
			cost.B,
		}),
		Colors:     []color.Color{color.Black},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Insect, types.Warrior},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		ActivatedAbilities: []game.ActivatedAbility{
			game.ActivatedAbility{
				Text:     "{3}{G}, Sacrifice a token: Draw a card.",
				ManaCost: opt.Val(cost.Mana{cost.O(3), cost.G}),
				AdditionalCosts: []cost.Additional{
					{
						Kind:         cost.AdditionalSacrifice,
						Text:         "Sacrifice a token",
						Amount:       1,
						RequireToken: true,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.Draw{
								Amount: game.Fixed(1),
								Player: game.ControllerReference(),
							},
						},
					},
				}.Ability(),
			},
		},
		TriggeredAbilities: []game.TriggeredAbility{
			game.TriggeredAbility{
				Trigger: game.TriggerCondition{
					Type: game.TriggerAt,
					Pattern: game.TriggerPattern{
						Event:      game.EventBeginningOfStep,
						Controller: game.TriggerControllerYou,
						Step:       game.StepEnd,
					},
					InterveningIf: "if a permanent left the battlefield under your control this turn",
					InterveningCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Controller:    game.TriggerControllerYou,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						}, Window: game.EventHistoryCurrentTurn}),
					}),
				},
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.CreateToken{
								Amount: game.Fixed(1),
								Source: game.TokenDef(lordDreggInsectInvaderToken),
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			Flying
			Disappear — At the beginning of your end step, if a permanent left the battlefield under your control this turn, create a 1/1 black Insect Warrior creature token with flying.
			{3}{G}, Sacrifice a token: Draw a card.
		`,
	},
}

var lordDreggInsectInvaderToken = newLordDreggInsectInvaderToken()

func newLordDreggInsectInvaderToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect Warrior",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
