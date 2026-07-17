package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StolenByTheFae is the card definition for Stolen by the Fae.
//
// Type: Sorcery
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Return target creature with mana value X to its owner's hand. You create X 1/1 blue Faerie creature tokens with flying.
var StolenByTheFae = newStolenByTheFae

func newStolenByTheFae() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Stolen by the Fae",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:       1,
						MaxTargets:       1,
						Constraint:       "target creature with mana value X",
						Allow:            game.TargetAllowPermanent,
						Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						ManaValueEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(stolenByTheFaeToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature with mana value X to its owner's hand. You create X 1/1 blue Faerie creature tokens with flying.
		`,
		},
	}
}

var stolenByTheFaeToken = newStolenByTheFaeToken()

func newStolenByTheFaeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Faerie",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
