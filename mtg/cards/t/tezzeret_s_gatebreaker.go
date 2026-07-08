package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TezzeretSGatebreaker is the card definition for Tezzeret's Gatebreaker.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	When this artifact enters, look at the top five cards of your library. You may reveal a blue or artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
//	{5}{U}, {T}, Sacrifice this artifact: Creatures you control can't be blocked this turn.
var TezzeretSGatebreaker = newTezzeretSGatebreaker

func newTezzeretSGatebreaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tezzeret's Gatebreaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{5}{U}, {T}, Sacrifice this artifact: Creatures you control can't be blocked this turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(5), cost.U}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:               game.RuleEffectCantBeBlocked,
											AffectedController: game.ControllerYou,
											PermanentTypes:     []types.Card{types.Creature},
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
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
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(5),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, ColorsAny: []color.Color{color.Blue}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters, look at the top five cards of your library. You may reveal a blue or artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
			{5}{U}, {T}, Sacrifice this artifact: Creatures you control can't be blocked this turn.
		`,
		},
	}
}
