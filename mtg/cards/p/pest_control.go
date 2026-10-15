package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PestControl is the card definition for Pest Control.
//
// Type: Sorcery
// Cost: {W}{B}
//
// Oracle text:
//
//	Destroy all nonland permanents with mana value 1 or less.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var PestControl = newPestControl

func newPestControl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Pest Control",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1})}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonland permanents with mana value 1 or less.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
