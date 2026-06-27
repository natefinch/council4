package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CargoShip is the card definition for Cargo Ship.
//
// Type: Artifact — Vehicle
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying, vigilance
//	{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.
//	Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)
var CargoShip = newCargoShip()

func newCargoShip() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cargo Ship",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(1),
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateArtifact,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance
			{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.
			Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}
