package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonloftIdol is the card definition for Dragonloft Idol.
//
// Type: Artifact Creature — Gargoyle
// Cost: {4}
//
// Oracle text:
//
//	As long as you control a Dragon, this creature gets +1/+1 and has flying and trample.
var DragonloftIdol = newDragonloftIdol

func newDragonloftIdol() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dragonloft Idol",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Dragon, this creature gets +1/+1 and has flying and trample.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Dragon",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Dragon}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Flying, game.Trample},
						},
					},
				},
			},
			OracleText: `
			As long as you control a Dragon, this creature gets +1/+1 and has flying and trample.
		`,
		},
	}
}
