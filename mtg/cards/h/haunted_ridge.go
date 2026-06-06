package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
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
			AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
			Content: game.PlainAbilityContent{
				Sequence: []game.Effect{
					{
						Type:        game.EffectChoose,
						TargetIndex: game.TargetIndexController,
						Choice: opt.Val(game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{mana.B, mana.R},
						}),
						LinkID: "haunted-ridge-color",
					},
					{
						Type:         game.EffectAddMana,
						Amount:       1,
						TargetIndex:  game.TargetIndexController,
						ChoiceLinkID: "haunted-ridge-color",
					},
				},
			},
		},
	)
	return card
}()
