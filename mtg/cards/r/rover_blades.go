package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoverBlades is the card definition for Rover Blades.
//
// Type: Artifact — Equipment Vehicle
// Cost: {3}
//
// Oracle text:
//
//	Double strike
//	Equipped creature has double strike.
//	Equip {4}
//	Crew 2 (Tap any number of creatures you control with total power 2 or more: This Vehicle becomes an artifact creature until end of turn. Creatures can't be attached to other permanents.)
var RoverBlades = newRoverBlades

func newRoverBlades() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Rover Blades",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Equipment, types.Vehicle},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.StaticAbility{
					Text: "Equipped creature has double strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.DoubleStrike},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
				game.CrewActivatedAbility(2),
			},
			OracleText: `
			Double strike
			Equipped creature has double strike.
			Equip {4}
			Crew 2 (Tap any number of creatures you control with total power 2 or more: This Vehicle becomes an artifact creature until end of turn. Creatures can't be attached to other permanents.)
		`,
		},
	}
}
