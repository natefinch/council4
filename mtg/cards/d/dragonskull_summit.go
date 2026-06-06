package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DragonskullSummit is the card definition for Dragonskull Summit.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Swamp or a Mountain.
//	{T}: Add {B} or {R}.
var DragonskullSummit = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Dragonskull Summit",
			Types: []types.Card{types.Land},
			OracleText: `
				This land enters tapped unless you control a Swamp or a Mountain.
				{T}: Add {B} or {R}.
			`,
		},
	}
	card.ReplacementAbilities = append(card.ReplacementAbilities,
		game.EntersTappedIfReplacement("This land enters tapped unless you control a Swamp or a Mountain.", &game.Condition{
			Negate: true,
			ControllerControls: game.PermanentFilter{
				SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
			},
		}),
	)
	card.ManaAbilities = append(card.ManaAbilities, game.ManaAbilityBody{
		Text: `
			{T}: Add {B} or {R}.
		`,
		AdditionalCosts: cost.Tap,
		Content: game.PlainAbilityContent{
			Sequence: []game.Instruction{
				{
					Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{
								mana.B,
								mana.R,
							},
						},
						PublishChoice: game.ChoiceKey("dragonskull-summit-color"),
					},
				},
				{
					Primitive: game.AddMana{
						Amount:     game.Fixed(1),
						ChoiceFrom: game.ChoiceKey("dragonskull-summit-color"),
					},
				},
			},
		},
	})
	return card
}()
