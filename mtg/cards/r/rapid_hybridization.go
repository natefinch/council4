package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RapidHybridization is the card definition for Rapid Hybridization.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Destroy target creature. It can't be regenerated. That creature's controller creates a 3/3 green Frog Lizard creature token.
var RapidHybridization = newRapidHybridization

func newRapidHybridization() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Rapid Hybridization",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object:              game.TargetPermanentReference(0),
							PreventRegeneration: true,
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:    game.Fixed(1),
							Source:    game.TokenDef(rapidHybridizationToken),
							Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature. It can't be regenerated. That creature's controller creates a 3/3 green Frog Lizard creature token.
		`,
		},
	}
}

var rapidHybridizationToken = newRapidHybridizationToken()

func newRapidHybridizationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Frog Lizard",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog, types.Lizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
