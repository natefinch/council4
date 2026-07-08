package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IntoTheFaeCourt is the card definition for Into the Fae Court.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Draw three cards. Create a 1/1 blue Faerie creature token with flying and "This token can block only creatures with flying."
var IntoTheFaeCourt = newIntoTheFaeCourt

func newIntoTheFaeCourt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Into the Fae Court",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(intoTheFaeCourtToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards. Create a 1/1 blue Faerie creature token with flying and "This token can block only creatures with flying."
		`,
		},
	}
}

var intoTheFaeCourtToken = newIntoTheFaeCourtToken()

func newIntoTheFaeCourtToken() *game.CardDef {
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
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCanBlockOnlyCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
		},
	}
}
