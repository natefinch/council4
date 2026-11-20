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
	sunsBountyOracleID    = "5c15d9b1-5ec7-4476-a3e6-539d899a9957"
	sunsBountySpellText   = "You gain 4 life."
	sunsBountyRecoverText = "Recover {1}{W} (When a creature is put into your graveyard from the battlefield, " +
		"you may pay {1}{W}. If you do, return this card from your graveyard to your hand. " +
		"Otherwise, exile this card.)"
)

var SunsBounty = cardset.Entry{
	OracleID:     sunsBountyOracleID,
	Name:         "Sun's Bounty",
	New:          newSunsBounty,
	Verification: "mtg/rules/suns_bounty_test.go: TestSunsBountySpellAndRecoverChoices",
}

func newSunsBounty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Sun's Bounty",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Instant},
			OracleText: sunsBountySpellText + "\n" + sunsBountyRecoverText,
			SpellAbility: opt.Val(game.Mode{
				Text: sunsBountySpellText,
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{
						Player: game.ControllerReference(),
						Amount: game.Fixed(4),
					},
				}},
			}.Ability()),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text:           sunsBountyRecoverText,
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
							Prompt:   "Pay {1}{W}?",
							ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
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
