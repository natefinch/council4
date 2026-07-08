package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FormlessGenesis is the card definition for Formless Genesis.
//
// Type: Kindred Sorcery — Shapeshifter
// Cost: {2}{G}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Create an X/X colorless Shapeshifter creature token with changeling and deathtouch, where X is the number of land cards in your graveyard.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var FormlessGenesis = newFormlessGenesis

func newFormlessGenesis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Formless Genesis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Shapeshifter},
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.RetraceStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(formlessGenesisToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Land}},
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Land}},
							})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Changeling (This card is every creature type.)
			Create an X/X colorless Shapeshifter creature token with changeling and deathtouch, where X is the number of land cards in your graveyard.
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}

var formlessGenesisToken = newFormlessGenesisToken()

func newFormlessGenesisToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Shapeshifter",
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Shapeshifter},
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.DeathtouchStaticBody,
			},
		},
	}
}
