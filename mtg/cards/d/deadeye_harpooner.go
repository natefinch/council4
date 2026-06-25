package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DeadeyeHarpooner is the card definition for Deadeye Harpooner.
//
// Type: Creature — Dwarf Warrior
// Cost: {2}{W}
//
// Oracle text:
//
//	Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, destroy target tapped creature an opponent controls.
var DeadeyeHarpooner = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name: "Deadeye Harpooner",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.W,
		}),
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dwarf, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{
			game.TriggeredAbility{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
					InterveningIf: "if a permanent left the battlefield under your control this turn",
					InterveningCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Controller:    game.TriggerControllerYou,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						}, Window: game.EventHistoryCurrentTurn}),
					}),
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{
						game.TargetSpec{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target tapped creature an opponent controls",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent, Tapped: game.TriTrue}),
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.Destroy{
								Object: game.TargetPermanentReference(0),
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, destroy target tapped creature an opponent controls.
		`,
	},
}
