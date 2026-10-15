package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrainingDrone is the card definition for Training Drone.
//
// Type: Artifact Creature — Drone
// Cost: {3}
//
// Oracle text:
//
//	This creature can't attack or block unless it's equipped.
var TrainingDrone = newTrainingDrone

func newTrainingDrone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Training Drone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Drone},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't attack or block unless it's equipped.",
					Condition: opt.Val(game.Condition{
						Text:              "unless it's equipped",
						Negate:            true,
						Object:            opt.Val(game.SourcePermanentReference()),
						ObjectMatches:     opt.Val(game.Selection{MatchEquipped: true}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			This creature can't attack or block unless it's equipped.
		`,
		},
	}
}
