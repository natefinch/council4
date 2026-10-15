package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TimelessWitness is the card definition for Timeless Witness.
//
// Type: Token Creature — Zombie Human Shaman
//
// Oracle text:
//   When Timeless Witness enters the battlefield, return target card from your graveyard to your hand.

// TimelessWitnessTokene292b49219184c92af433bfbe39dcea4 is the card definition for Timeless Witness.
var TimelessWitnessTokene292b49219184c92af433bfbe39dcea4 = newTimelessWitnessTokene292b49219184c92af433bfbe39dcea4()

func newTimelessWitnessTokene292b49219184c92af433bfbe39dcea4() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Timeless Witness",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Timeless Witness enters the battlefield, return target card from your graveyard to your hand.",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			When Timeless Witness enters the battlefield, return target card from your graveyard to your hand.
		`,
		},
	}
}
