package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const icefallDestroyText = "Destroy target artifact or land."
const icefallRecoverText = "Recover {R}{R} (When a creature is put into your graveyard from the battlefield, you may pay {R}{R}. If you do, return this card from your graveyard to your hand. Otherwise, exile this card.)"

var Icefall = cardset.Entry{
	OracleID:     "30baca42-50e2-45f8-8f99-8e2f9795b4da",
	Name:         "Icefall",
	New:          newIcefall,
	Verification: "mtg/rules/icefall_test.go: TestIcefallDestroyTargetAndRecoverChoices",
}

func newIcefall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Icefall",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R, cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Sorcery},
			OracleText: icefallDestroyText + "\n" + icefallRecoverText,
			SpellAbility: opt.Val(game.Mode{
				Text: icefallDestroyText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Land}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability()),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text:           icefallRecoverText,
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
							Prompt:   "Pay {R}{R}?",
							ManaCost: opt.Val(cost.Mana{cost.R, cost.R}),
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
