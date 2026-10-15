package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HowlpackResurgence is the card definition for Howlpack Resurgence.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Flash
//	Each creature you control that's a Wolf or a Werewolf gets +1/+1 and has trample.
var HowlpackResurgence = newHowlpackResurgence

func newHowlpackResurgence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Howlpack Resurgence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "Each creature you control that's a Wolf or a Werewolf gets +1/+1 and has trample.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Wolf"), types.Sub("Werewolf")}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Wolf"), types.Sub("Werewolf")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
					},
				},
			},
			OracleText: `
			Flash
			Each creature you control that's a Wolf or a Werewolf gets +1/+1 and has trample.
		`,
		},
	}
}
