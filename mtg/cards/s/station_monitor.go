package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StationMonitor is the card definition for Station Monitor.
//
// Type: Creature — Lizard Artificer
// Cost: {W}{U}
//
// Oracle text:
//
//	Whenever you cast your second spell each turn, create a 1/1 colorless Drone artifact creature token with flying and "This token can block only creatures with flying."
var StationMonitor = newStationMonitor

func newStationMonitor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Station Monitor",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Artificer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(stationMonitorToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast your second spell each turn, create a 1/1 colorless Drone artifact creature token with flying and "This token can block only creatures with flying."
		`,
		},
	}
}

var stationMonitorToken = newStationMonitorToken()

func newStationMonitorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Drone",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Drone},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCanBlockOnlyCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
		},
	}
}
