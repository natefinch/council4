package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DesperateRitual is the card definition for Desperate Ritual.
//
// Type: Instant — Arcane
// Cost: {1}{R}
//
// Oracle text:
//
//	Add {R}{R}{R}.
//	Splice onto Arcane {1}{R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
var DesperateRitual = newDesperateRitual

func newDesperateRitual() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Desperate Ritual",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SpliceKeyword{Cost: cost.Mana{cost.O(1), cost.R}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.R,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.R,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.R,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Add {R}{R}{R}.
			Splice onto Arcane {1}{R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
		`,
		},
	}
}
