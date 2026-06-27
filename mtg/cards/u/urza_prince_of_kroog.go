package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrzaPrinceOfKroog is the card definition for Urza, Prince of Kroog.
//
// Type: Legendary Creature — Human Artificer
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Artifact creatures you control get +2/+2.
//	{6}: Create a token that's a copy of target artifact you control, except it's a 1/1 Soldier creature in addition to its other types.
var UrzaPrinceOfKroog = newUrzaPrinceOfKroog()

func newUrzaPrinceOfKroog() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Urza, Prince of Kroog",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Artificer},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{6}: Create a token that's a copy of target artifact you control, except it's a 1/1 Soldier creature in addition to its other types.",
					ManaCost:       opt.Val(cost.Mana{cost.O(6)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:       game.TokenCopySourceObject,
										Object:       game.TargetPermanentReference(0),
										SetPower:     opt.Val(game.PT{Value: 1}),
										SetToughness: opt.Val(game.PT{Value: 1}),
										AddTypes:     []types.Card{types.Creature},
										AddSubtypes:  []types.Sub{types.Soldier},
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Artifact creatures you control get +2/+2.
			{6}: Create a token that's a copy of target artifact you control, except it's a 1/1 Soldier creature in addition to its other types.
		`,
		},
	}
}
