package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkromaSMemorial is the card definition for Akroma's Memorial.
//
// Type: Legendary Artifact
// Cost: {7}
//
// Oracle text:
//
//	Creatures you control have flying, first strike, vigilance, trample, haste, and protection from black and from red.
var AkromaSMemorial = newAkromaSMemorial

func newAkromaSMemorial() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Akroma's Memorial",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have flying, first strike, vigilance, trample, haste, and protection from black and from red.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying, game.FirstStrike, game.Vigilance, game.Trample, game.Haste},
							AddAbilities: []game.Ability{
								new(game.ProtectionFromColorsStaticAbility(color.Black, color.Red)),
							},
						},
					},
				},
			},
			OracleText: `
			Creatures you control have flying, first strike, vigilance, trample, haste, and protection from black and from red.
		`,
		},
	}
}
