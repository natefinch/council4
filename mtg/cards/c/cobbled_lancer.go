package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CobbledLancer is the card definition for Cobbled Lancer.
//
// Type: Creature — Zombie Horse
// Cost: {U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile a creature card from your graveyard.
//	{3}{U}, Exile this card from your graveyard: Draw a card.
var CobbledLancer = newCobbledLancer

func newCobbledLancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cobbled Lancer",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Horse},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{U}, Exile this card from your graveyard: Draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this card from your graveyard",
							Amount:      1,
							Source:      zone.Graveyard,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalExile,
					Text:          "exile a creature card from your graveyard",
					Amount:        1,
					Source:        zone.Graveyard,
					MatchCardType: true,
					CardType:      types.Creature,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, exile a creature card from your graveyard.
			{3}{U}, Exile this card from your graveyard: Draw a card.
		`,
		},
	}
}
