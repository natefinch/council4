package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HugsGrislyGuardian is the card definition for Hugs, Grisly Guardian.
//
// Type: Legendary Creature — Badger Warrior
// Cost: {X}{R}{R}{G}{G}
//
// Oracle text:
//
//	Trample
//	When Hugs enters, exile the top X cards of your library. Until the end of your next turn, you may play those cards.
//	You may play an additional land on each of your turns.
var HugsGrislyGuardian = newHugsGrislyGuardian

func newHugsGrislyGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Hugs, Grisly Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Badger, types.Warrior},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "You may play an additional land on each of your turns.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                game.RuleEffectAdditionalLandPlays,
							Duration:            game.DurationPermanent,
							AffectedPlayer:      game.PlayerYou,
							AdditionalLandPlays: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Hugs enters, exile the top X cards of your library. Until the end of your next turn, you may play those cards.",
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
										Primitive: game.ImpulseExile{
											Player: game.ControllerReference(),
											Amount: game.Dynamic(game.DynamicAmount{
												Kind: game.DynamicAmountX,
											}),
											Duration: game.DurationUntilEndOfYourNextTurn,
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
			Trample
			When Hugs enters, exile the top X cards of your library. Until the end of your next turn, you may play those cards.
			You may play an additional land on each of your turns.
		`,
		},
	}
}
