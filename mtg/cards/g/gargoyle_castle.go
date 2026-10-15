package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GargoyleCastle is the card definition for Gargoyle Castle.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{5}, {T}, Sacrifice this land: Create a 3/4 colorless Gargoyle artifact creature token with flying.
var GargoyleCastle = newGargoyleCastle

func newGargoyleCastle() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Gargoyle Castle",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, {T}, Sacrifice this land: Create a 3/4 colorless Gargoyle artifact creature token with flying.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this land",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(gargoyleCastleToken),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{5}, {T}, Sacrifice this land: Create a 3/4 colorless Gargoyle artifact creature token with flying.
		`,
		},
	}
}

var gargoyleCastleToken = newGargoyleCastleToken()

func newGargoyleCastleToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Gargoyle",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
