package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OakenSiren is the card definition for Oaken Siren.
//
// Type: Artifact Creature — Siren Pirate
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying, vigilance
//	{T}: Add {U}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.
var OakenSiren = newOakenSiren

func newOakenSiren() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Oaken Siren",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Siren, types.Pirate},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
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
			{T}: Add {U}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.
		`,
		},
	}
}
