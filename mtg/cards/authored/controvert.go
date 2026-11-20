package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	controvertOracleID    = "f7ad79e3-d128-4aee-8542-99fd3527f9d3"
	controvertSpellText   = "Counter target spell."
	controvertRecoverText = "Recover {2}{U}{U} (When a creature is put into your graveyard from the battlefield, you may pay {2}{U}{U}. If you do, return this card from your graveyard to your hand. Otherwise, exile this card.)"
)

var controvertProgram = game.ProgramReference{ID: controvertOracleID + "/spell/0", Version: "1"}

var Controvert = cardset.Entry{
	OracleID:     controvertOracleID,
	Name:         "Controvert",
	New:          newControvert,
	Programs:     []cardprogram.Program{{Reference: controvertProgram, Resolve: resolveCounterTarget}},
	Verification: "mtg/rules/controvert_test.go: TestControvertCountersSpellAndRecovers (counter spell, cast cost, Recover payment and exile choices)",
}

func newControvert() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Controvert",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Instant},
			OracleText: controvertSpellText + "\n" + controvertRecoverText,
			SpellAbility: opt.Val(game.Mode{
				Text:    controvertSpellText,
				Program: controvertProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowStackObject,
					Predicate:  game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
				}},
			}.Ability()),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text:           controvertRecoverText,
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
							Prompt:   "Pay {2}{U}{U}?",
							ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
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
