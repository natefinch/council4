package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NemataGroveGuardian is the card definition for Nemata, Grove Guardian.
//
// Type: Legendary Creature — Treefolk
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	{2}{G}: Create a 1/1 green Saproling creature token.
//	Sacrifice a Saproling: Saproling creatures get +1/+1 until end of turn.
var NemataGroveGuardian = newNemataGroveGuardian()

func newNemataGroveGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nemata, Grove Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Treefolk},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{G}: Create a 1/1 green Saproling creature token.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(nemataGroveGuardianToken),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text: "Sacrifice a Saproling: Saproling creatures get +1/+1 until end of turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Saproling",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Saproling},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Saproling")}}),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}{G}: Create a 1/1 green Saproling creature token.
			Sacrifice a Saproling: Saproling creatures get +1/+1 until end of turn.
		`,
		},
	}
}

var nemataGroveGuardianToken = newNemataGroveGuardianToken()

func newNemataGroveGuardianToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
