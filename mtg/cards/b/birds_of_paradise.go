package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BirdsOfParadise is the card definition for Birds of Paradise.
//
// Type: Creature — Bird
// Cost: {G}
//
// Oracle text:
//
//	Flying
//	{T}: Add one mana of any color.
var BirdsOfParadise = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Birds of Paradise",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			OracleText: `
				Flying
				{T}: Add one mana of any color.
			`,
		},
	}
	card.StaticAbilities = append(card.StaticAbilities,
		game.FlyingStaticBody,
	)

	card.ManaAbilities = append(card.ManaAbilities,
		game.ManaAbilityBody{
			Text: `
				{T}: Add one mana of any color.
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
									mana.W,
									mana.U,
									mana.B,
									mana.R,
									mana.G,
								},
							},
							PublishChoice: game.ChoiceKey("birds-color"),
						},
					},
					{
						Primitive: game.AddMana{
							Amount:     game.Fixed(1),
							ChoiceFrom: game.ChoiceKey("birds-color"),
						},
					},
				},
			},
		},
	)
	return card
}()
