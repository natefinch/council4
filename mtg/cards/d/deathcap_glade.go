package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeathcapGlade is the card definition for Deathcap Glade.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {B} or {G}.
var DeathcapGlade = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:  "Deathcap Glade",
			Types: []types.Card{types.Land},
			OracleText: `
				This land enters tapped unless you control two or more other lands.
				{T}: Add {B} or {G}.
			`,
		},
	}
	card.ReplacementAbilities = append(card.ReplacementAbilities,
		game.EntersTappedIfReplacement("This land enters tapped unless you control two or more other lands.", &game.Condition{
			Negate: true,
			ControllerControls: game.PermanentFilter{
				Types:    []types.Card{types.Land},
				MinCount: 2,
			},
		}),
	)
	card.ManaAbilities = append(card.ManaAbilities, game.ManaAbilityBody{
		Text: `
			{T}: Add {B} or {G}.
		`,
		AdditionalCosts: []game.AdditionalCost{
			{Kind: game.AdditionalCostTap},
		},
		Content: game.PlainAbilityContent{
			Sequence: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.B, mana.G},
					}),
					LinkID: "deathcap-glade-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "deathcap-glade-color",
				},
			},
		},
	})
	return card
}()
