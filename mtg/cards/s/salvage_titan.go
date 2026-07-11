package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SalvageTitan is the card definition for Salvage Titan.
//
// Type: Artifact Creature — Golem
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	You may sacrifice three artifacts rather than pay this spell's mana cost.
//	Exile three artifact cards from your graveyard: Return this card from your graveyard to your hand.
var SalvageTitan = newSalvageTitan

func newSalvageTitan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Salvage Titan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Exile three artifact cards from your graveyard: Return this card from your graveyard to your hand.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile three artifact cards from your graveyard",
							Amount:        3,
							Source:        zone.Graveyard,
							MatchCardType: true,
							CardType:      types.Artifact,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceSource},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice three artifacts",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "sacrifice three artifacts",
							Amount:             3,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
				},
			},
			OracleText: `
			You may sacrifice three artifacts rather than pay this spell's mana cost.
			Exile three artifact cards from your graveyard: Return this card from your graveyard to your hand.
		`,
		},
	}
}
