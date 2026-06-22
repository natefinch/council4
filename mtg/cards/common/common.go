// Package common provides commonly used card components for Magic: The Gathering cards.
package common

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// RampLand configures an ability that searches the controller's library for a land and puts it into play.
type RampLand struct {
	Basic, Tapped bool
	SubTypes      []types.Sub
}

// Ability returns non-modal content that searches the controller's library for a land and puts it into play.
func (r RampLand) Ability() game.AbilityContent {
	return game.Mode{
		Sequence: []game.Instruction{
			r.Instruction(),
		},
	}.Ability()
}

// Instruction returns a single instruction for the ramp land ability.
func (r RampLand) Instruction() game.Instruction {
	filter := game.Selection{
		RequiredTypes: []types.Card{types.Land},
		SubtypesAny:   r.SubTypes,
	}
	if r.Basic {
		filter.Supertypes = []types.Super{types.Basic}
	}
	return game.Instruction{
		Primitive: game.Search{
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				Filter:       filter,
				EntersTapped: r.Tapped,
			},
		},
	}
}

// ETB is a trigger condition for when a permanent enters the battlefield.
var ETB = game.TriggerCondition{
	Type: game.TriggerWhen,
	Pattern: game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	},
}
