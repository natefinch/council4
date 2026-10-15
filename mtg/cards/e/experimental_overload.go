package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ExperimentalOverload is the card definition for Experimental Overload.
//
// Type: Sorcery
// Cost: {2}{U}{R}
//
// Oracle text:
//
//	Create an X/X blue and red Weird creature token, where X is the number of instant and sorcery cards in your graveyard. Then you may return an instant or sorcery card from your graveyard to your hand. Exile Experimental Overload.
var ExperimentalOverload = newExperimentalOverload

func newExperimentalOverload() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Experimental Overload",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(experimentalOverloadToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							})),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Hand,
							},
							Prompt: "Choose a card to return to your hand",
						},
						Optional: true,
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create an X/X blue and red Weird creature token, where X is the number of instant and sorcery cards in your graveyard. Then you may return an instant or sorcery card from your graveyard to your hand. Exile Experimental Overload.
		`,
		},
	}
}

var experimentalOverloadToken = newExperimentalOverloadToken()

func newExperimentalOverloadToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Weird",
			Colors:   []color.Color{color.Blue, color.Red},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Weird},
		},
	}
}
