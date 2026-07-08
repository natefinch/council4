package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnfathomableTruths is the card definition for Unfathomable Truths.
//
// Type: Instant
// Cost: {4}{U}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Draw three cards and create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
var UnfathomableTruths = newUnfathomableTruths

func newUnfathomableTruths() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Unfathomable Truths",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Types: []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
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
							Source: game.TokenDef(unfathomableTruthsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Devoid (This card has no color.)
			Draw three cards and create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
		`,
		},
	}
}

var unfathomableTruthsToken = newUnfathomableTruthsToken()

func newUnfathomableTruthsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
