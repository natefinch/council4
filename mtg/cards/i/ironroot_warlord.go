package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IronrootWarlord is the card definition for Ironroot Warlord.
//
// Type: Creature — Treefolk Soldier
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Ironroot Warlord's power is equal to the number of creatures you control.
//	{3}{G}{W}: Create a 1/1 white Soldier creature token.
var IronrootWarlord = newIronrootWarlord

func newIronrootWarlord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Ironroot Warlord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:       []color.Color{color.Green, color.White},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Treefolk, types.Soldier},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 5}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{G}{W}: Create a 1/1 white Soldier creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.G,
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(ironrootWarlordToken),
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
			Ironroot Warlord's power is equal to the number of creatures you control.
			{3}{G}{W}: Create a 1/1 white Soldier creature token.
		`,
		},
	}
}

var ironrootWarlordToken = newIronrootWarlordToken()

func newIronrootWarlordToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
