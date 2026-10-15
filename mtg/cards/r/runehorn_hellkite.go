package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RunehornHellkite is the card definition for Runehorn Hellkite.
//
// Type: Creature — Dragon
// Cost: {5}{R}
//
// Oracle text:
//
//	Flying
//	{5}{R}, Exile this card from your graveyard: Each player discards their hand, then draws seven cards.
var RunehornHellkite = newRunehornHellkite

func newRunehornHellkite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Runehorn Hellkite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{R}, Exile this card from your graveyard: Each player discards their hand, then draws seven cards.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.R,
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
										Primitive: game.Discard{
											PlayerGroup: game.AllPlayersReference(),
											EntireHand:  true,
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount:      game.Fixed(7),
											PlayerGroup: game.AllPlayersReference(),
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
			OracleText: `
			Flying
			{5}{R}, Exile this card from your graveyard: Each player discards their hand, then draws seven cards.
		`,
		},
	}
}
