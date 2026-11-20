package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	krovikanRotDestroyText = "Destroy target creature with power 2 or less."
	krovikanRotRecoverText = "Recover {1}{B}{B} (When a creature is put into your graveyard from the battlefield, you may pay {1}{B}{B}. If you do, return this card from your graveyard to your hand. Otherwise, exile this card.)"
)

var KrovikanRot = cardset.Entry{
	OracleID:     "f9a272c5-6ebc-4d29-a47c-a2d570184196",
	Name:         "Krovikan Rot",
	New:          newKrovikanRot,
	Verification: "mtg/rules/krovikan_rot_test.go: TestKrovikanRotDestroysSmallCreatureAndRecovers (cast cost, target condition, spell result, optional Recover payment choices)",
}

func newKrovikanRot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Krovikan Rot",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Instant},
			OracleText: krovikanRotDestroyText + "\n" + krovikanRotRecoverText,
			SpellAbility: opt.Val(game.Mode{
				Text: krovikanRotDestroyText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature with power 2 or less",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature},
						Power:            opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability()),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text:           krovikanRotRecoverText,
				ZoneOfFunction: zone.Graveyard,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:            game.EventPermanentDied,
						Player:           game.TriggerPlayerYou,
						SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive: game.Pay{Payment: game.ResolutionPayment{
							Prompt:   "Pay {1}{B}{B}?",
							ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B, cost.B}),
						}},
						PublishResult: "paid",
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceSource},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "paid",
							Succeeded: game.TriTrue,
						}),
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceSource},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "paid",
							Succeeded: game.TriFalse,
						}),
					},
				}}.Ability(),
			}},
		},
	}
}
