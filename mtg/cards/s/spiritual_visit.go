package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpiritualVisit is the card definition for Spiritual Visit.
//
// Type: Instant — Arcane
// Cost: {W}
//
// Oracle text:
//
//	Create a 1/1 colorless Spirit creature token.
//	Splice onto Arcane {W} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
var SpiritualVisit = newSpiritualVisit

func newSpiritualVisit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Spiritual Visit",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SpliceKeyword{Cost: cost.Mana{cost.W}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(spiritualVisitToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 colorless Spirit creature token.
			Splice onto Arcane {W} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
		`,
		},
	}
}

var spiritualVisitToken = newSpiritualVisitToken()

func newSpiritualVisitToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
