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

const grimHarvestOracleID = "231ca538-b277-4c12-979f-57640a0e87f0"

var GrimHarvest = cardset.Entry{
	OracleID:     grimHarvestOracleID,
	Name:         "Grim Harvest",
	New:          newGrimHarvest,
	Verification: "mtg/rules/grim_harvest_test.go: TestGrimHarvestSpellAndRecoverChoices",
}

func newGrimHarvest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Grim Harvest",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Instant},
			OracleText: "Return target creature card from your graveyard to your hand.\nRecover {2}{B} (When a creature is put into your graveyard from the battlefield, you may pay {2}{B}. If you do, return this card from your graveyard to your hand. Otherwise, exile this card.)",
			SpellAbility: opt.Val(game.Mode{
				Text: "Return target creature card from your graveyard to your hand.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.MoveCard{
						Card:        game.CardReference{Kind: game.CardReferenceTarget},
						FromZone:    zone.Graveyard,
						Destination: zone.Hand,
					},
				}},
			}.Ability()),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text:           "Recover {2}{B} (When a creature is put into your graveyard from the battlefield, you may pay {2}{B}. If you do, return this card from your graveyard to your hand. Otherwise, exile this card.)",
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
							Prompt:   "Pay {2}{B}?",
							ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
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
