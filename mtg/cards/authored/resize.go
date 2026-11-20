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

const (
	resizeOracleID    = "2c092ed4-4dbb-41ec-9e07-e59b9676802e"
	resizeSpellText   = "Target creature gets +3/+3 until end of turn."
	resizeRecoverText = "Recover {1}{G} (When a creature is put into your graveyard from the battlefield, " +
		"you may pay {1}{G}. If you do, return this card from your graveyard to your hand. " +
		"Otherwise, exile this card.)"
)

var Resize = cardset.Entry{
	OracleID:     resizeOracleID,
	Name:         "Resize",
	New:          newResize,
	Verification: "mtg/rules/resize_test.go: TestResizeSpellAndRecover",
}

func newResize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Resize",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Instant},
			OracleText: resizeSpellText + "\n" + resizeRecoverText,
			SpellAbility: opt.Val(game.Mode{
				Text: resizeSpellText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ModifyPT{
						Object:         game.TargetPermanentReference(0),
						PowerDelta:     game.Fixed(3),
						ToughnessDelta: game.Fixed(3),
						Duration:       game.DurationUntilEndOfTurn,
					},
				}},
			}.Ability()),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text:           resizeRecoverText,
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
							Prompt:   "Pay {1}{G}?",
							ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
						}},
						PublishResult: "paid",
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceSource},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriTrue}),
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceSource},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriFalse}),
					},
				}}.Ability(),
			}},
		},
	}
}
