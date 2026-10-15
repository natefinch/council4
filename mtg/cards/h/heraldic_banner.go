package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeraldicBanner is the card definition for Heraldic Banner.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	As this artifact enters, choose a color.
//	Creatures you control of the chosen color get +1/+0.
//	{T}: Add one mana of the chosen color.
var HeraldicBanner = newHeraldicBanner

func newHeraldicBanner() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Heraldic Banner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control of the chosen color get +1/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorChoice: game.ColorChoiceSourceEntry}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this artifact enters, choose a color."),
			},
			OracleText: `
			As this artifact enters, choose a color.
			Creatures you control of the chosen color get +1/+0.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
