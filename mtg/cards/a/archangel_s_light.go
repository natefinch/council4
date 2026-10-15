package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArchangelSLight is the card definition for Archangel's Light.
//
// Type: Sorcery
// Cost: {7}{W}
//
// Oracle text:
//
//	You gain 2 life for each card in your graveyard, then shuffle your graveyard into your library.
var ArchangelSLight = newArchangelSLight

func newArchangelSLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Archangel's Light",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 2,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{},
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ShuffleGraveyardIntoLibrary{
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You gain 2 life for each card in your graveyard, then shuffle your graveyard into your library.
		`,
		},
	}
}
