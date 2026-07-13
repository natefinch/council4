package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpectralSnatcher is the card definition for Spectral Snatcher.
//
// Type: Creature — Spirit
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Ward—Discard a card. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player discards a card.)
//	Swampcycling {2} ({2}, Discard this card: Search your library for a Swamp card, reveal it, put it into your hand, then shuffle.)
var SpectralSnatcher = newSpectralSnatcher

func newSpectralSnatcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Spectral Snatcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalDiscard,
						Text:   "Discard a card",
						Amount: 1,
						Source: zone.Hand,
					},
				}),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard this card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Hand,
					KeywordAbilities: []game.KeywordAbility{
						game.CyclingKeyword{Cost: cost.Mana{cost.O(2)}},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Hand,
										Filter:      game.Selection{SubtypesAny: []types.Sub{types.Sub("Swamp")}},
										Reveal:      true,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Ward—Discard a card. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player discards a card.)
			Swampcycling {2} ({2}, Discard this card: Search your library for a Swamp card, reveal it, put it into your hand, then shuffle.)
		`,
		},
	}
}
