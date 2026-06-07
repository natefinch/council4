// Package common provides commonly used card components for Magic: The Gathering cards.
package common

import (
	"fmt"
	"strconv"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TapForOneOfAny creates a mana ability body for a tap mana ability with a choice from any of the five basic mana colors.
func TapForOneOfAny(key game.ChoiceKey) game.ManaAbilityBody {
	return TapForOneOf(key, mana.W, mana.U, mana.B, mana.R, mana.G)
}

// TapForOneOf creates a mana ability body for a tap mana ability with a choice from two mana colors.
func TapForOneOf(key game.ChoiceKey, colors ...mana.Color) game.ManaAbilityBody {
	var text string
	switch len(colors) {
	case 2:
		text = fmt.Sprintf(`
			{T}: Add {%s} or {%s}.
		`, colors[0], colors[1])
	case 3:
		text = fmt.Sprintf(`
			{T}: Add {%s}, {%s}, or {%s}.
		`, colors[0], colors[1], colors[2])
	case 4:
		text = fmt.Sprintf(`
			{T}: Add {%s}, {%s}, {%s}, or {%s}.
		`, colors[0], colors[1], colors[2], colors[3])
	case 5:
		text = `
			{T}: Add one mana of any color.
		`
	default:
		panic("invalid number of colors: " + strconv.Itoa(len(colors)))
	}
	return game.ManaAbilityBody{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: game.PlainAbilityContent{
			Sequence: []game.Instruction{
				{
					Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: colors,
						},
						PublishChoice: key,
					},
				},
				{
					Primitive: game.AddMana{
						Amount:     game.Fixed(1),
						ChoiceFrom: key,
					},
				},
			},
		},
	}
}

// TapForOne creates a mana ability body for a tap mana ability.
func TapForOne(clr mana.Color) game.ManaAbilityBody {
	return game.ManaAbilityBody{
		Text: fmt.Sprintf(`
			{T}: Add {%s}.
		`, clr),
		AdditionalCosts: cost.Tap,
		Content: game.PlainAbilityContent{
			Sequence: []game.Instruction{
				{
					Primitive: game.AddMana{
						Amount:    game.Fixed(1),
						ManaColor: clr,
					},
				},
			},
		},
	}
}

// RampLand creates a plain ability content to searches the controller's library for a land and puts it into play.
type RampLand struct {
	Basic, Tapped bool
	SubTypes      []types.Sub
}

// Ability returns a plain ability content that searches the controller's library for a land and puts it into play.
func (r RampLand) Ability() game.PlainAbilityContent {
	return game.PlainAbilityContent{
		Sequence: []game.Instruction{
			r.Instruction(),
		},
	}
}

// Instruction returns a single instruction for the ramp land ability.
func (r RampLand) Instruction() game.Instruction {
	var basics opt.V[types.Super]
	if r.Basic {
		basics = opt.Val(types.Basic)
	}
	return game.Instruction{
		Primitive: game.Search{
			TargetIndex: game.TargetIndexController,
			Spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				SubtypesAny:  r.SubTypes,
				CardType:     opt.Val(types.Land),
				Supertype:    basics,
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
