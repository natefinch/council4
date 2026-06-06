package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// HauntedRidge is the card definition for Haunted Ridge.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {B} or {R}.
var HauntedRidge = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Haunted Ridge",
			Types: []types.Card{types.Land},
			OracleText: `
				This land enters tapped unless you control two or more other lands.
				{T}: Add {B} or {R}.
			`,
			ReplacementAbilities: []game.ReplacementAbilityDef{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or more other lands.", &game.Condition{
					Negate: true,
					ControllerControls: game.PermanentFilter{
						Types:    []types.Card{types.Land},
						MinCount: 2,
					},
				}),
			},
		},
	}

	card.ManaAbilities = append(card.ManaAbilities,
		game.ManaAbilityBody{
			Text: `
				{T}: Add {B} or {R}.
			`,
			AdditionalCosts: []cost.Additional{
				{
					Kind: cost.AdditionalTap,
				},
			},
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
							PublishChoice: game.ChoiceKey("haunted-ridge-color"),
						},
					},
					{
						Primitive: game.AddMana{
							Amount:     game.Fixed(1),
							ChoiceFrom: game.ChoiceKey("haunted-ridge-color"),
						},
					},
				},
			},
		},
	)
	return card
}()
