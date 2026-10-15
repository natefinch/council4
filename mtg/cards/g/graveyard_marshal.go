package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GraveyardMarshal is the card definition for Graveyard Marshal.
//
// Type: Creature — Zombie Soldier
// Cost: {B}{B}
//
// Oracle text:
//
//	{2}{B}, Exile a creature card from your graveyard: Create a tapped 2/2 black Zombie creature token.
var GraveyardMarshal = newGraveyardMarshal

func newGraveyardMarshal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Graveyard Marshal",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{B}, Exile a creature card from your graveyard: Create a tapped 2/2 black Zombie creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:          cost.AdditionalExile,
							Text:          "Exile a creature card from your graveyard",
							Amount:        1,
							MatchCardType: true,
							CardType:      types.Creature,
							Source:        zone.Graveyard,
							CounterKind:   counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:      game.Fixed(1),
											Source:      game.TokenDef(graveyardMarshalToken),
											EntryTapped: true,
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
			{2}{B}, Exile a creature card from your graveyard: Create a tapped 2/2 black Zombie creature token.
		`,
		},
	}
}

var graveyardMarshalToken = newGraveyardMarshalToken()

func newGraveyardMarshalToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
