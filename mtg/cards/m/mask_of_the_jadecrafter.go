package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MaskOfTheJadecrafter is the card definition for Mask of the Jadecrafter.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{X}, {T}, Sacrifice this artifact: Create an X/X colorless Golem artifact creature token. Activate only as a sorcery.
//	Unearth {2}{G} ({2}{G}: Return this card from your graveyard to the battlefield. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
var MaskOfTheJadecrafter = newMaskOfTheJadecrafter

func newMaskOfTheJadecrafter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mask of the Jadecrafter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{X}, {T}, Sacrifice this artifact: Create an X/X colorless Golem artifact creature token. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.X}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(maskOfTheJadecrafterToken),
									Power: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									})),
									Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									})),
								},
							},
						},
					}.Ability(),
				},
				game.UnearthActivatedAbility(cost.Mana{cost.O(2), cost.G}),
			},
			OracleText: `
			{X}, {T}, Sacrifice this artifact: Create an X/X colorless Golem artifact creature token. Activate only as a sorcery.
			Unearth {2}{G} ({2}{G}: Return this card from your graveyard to the battlefield. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
		`,
		},
	}
}

var maskOfTheJadecrafterToken = newMaskOfTheJadecrafterToken()

func newMaskOfTheJadecrafterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Golem",
			Types:    []types.Card{types.Artifact, types.Creature},
			Subtypes: []types.Sub{types.Golem},
		},
	}
}
