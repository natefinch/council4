package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MacCreadyLamplightMayor is the card definition for MacCready, Lamplight Mayor.
//
// Type: Legendary Creature — Human Advisor
// Cost: {W}{B}
//
// Oracle text:
//
//	Whenever a creature you control with power 2 or less attacks, it gains skulk until end of turn. (It can't be blocked by creatures with greater power.)
//	Whenever a creature with power 4 or greater attacks you, its controller loses 2 life and you gain 2 life.
var MacCreadyLamplightMayor = newMacCreadyLamplightMayor

func newMacCreadyLamplightMayor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "MacCready, Lamplight Mayor",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Advisor},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.EventPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Skulk,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Player:           game.TriggerPlayerYou,
							AttackRecipient:  game.AttackRecipientPlayer,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature you control with power 2 or less attacks, it gains skulk until end of turn. (It can't be blocked by creatures with greater power.)
			Whenever a creature with power 4 or greater attacks you, its controller loses 2 life and you gain 2 life.
		`,
		},
	}
}
