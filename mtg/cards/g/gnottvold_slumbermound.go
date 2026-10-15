package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GnottvoldSlumbermound is the card definition for Gnottvold Slumbermound.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {R}.
//	{3}{R}{G}{G}, {T}, Sacrifice this land: Destroy target land. Create a 4/4 green Troll Warrior creature token with trample.
var GnottvoldSlumbermound = newGnottvoldSlumbermound

func newGnottvoldSlumbermound() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:  "Gnottvold Slumbermound",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{R}{G}{G}, {T}, Sacrifice this land: Destroy target land. Create a 4/4 green Troll Warrior creature token with trample.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.R,
						cost.G,
						cost.G,
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target land",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(gnottvoldSlumbermoundToken),
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
				game.TapManaAbility(mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {R}.
			{3}{R}{G}{G}, {T}, Sacrifice this land: Destroy target land. Create a 4/4 green Troll Warrior creature token with trample.
		`,
		},
	}
}

var gnottvoldSlumbermoundToken = newGnottvoldSlumbermoundToken()

func newGnottvoldSlumbermoundToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Troll Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
