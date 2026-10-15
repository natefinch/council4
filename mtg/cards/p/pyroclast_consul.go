package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PyroclastConsul is the card definition for Pyroclast Consul.
//
// Type: Creature — Elemental Shaman
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, this creature deals 2 damage to each creature.
var PyroclastConsul = newPyroclastConsul

func newPyroclastConsul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pyroclast Consul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Shaman},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, this creature deals 2 damage to each creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "You may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, this creature deals 2 damage to each creature.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("kinship-top"),
										},
										Optional: true,
									},
									game.Instruction{
										Primitive: game.Reveal{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "kinship-top",
											},
										},
										CardCondition: opt.Val(game.CardSelection{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "kinship-top",
											},
											Selection: game.Selection{SharesCreatureTypeWithSource: true},
										}),
										Optional:      true,
										PublishResult: game.ResultKey("kinship-revealed"),
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
											Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("kinship-revealed"),
											Accepted:  game.TriTrue,
											Succeeded: game.TriTrue,
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
			Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, this creature deals 2 damage to each creature.
		`,
		},
	}
}
