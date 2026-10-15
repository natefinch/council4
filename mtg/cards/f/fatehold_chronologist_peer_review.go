package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FateholdChronologist is the card definition for Fatehold Chronologist // Peer Review.
//
// Type: Creature — Bird Wizard // Sorcery
// Cost: {1}{W/U} // {2}{W/U}
// Face: Peer Review — Sorcery ({2}{W/U})
//
// Oracle text:
//
//	Flying
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var FateholdChronologist = newFateholdChronologist

func newFateholdChronologist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Fatehold Chronologist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.W, mana.U),
			}),
			Colors:         []color.Color{color.White, color.Blue},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Bird, types.Wizard},
			Power:          opt.Val(game.PT{Value: 1}),
			Toughness:      opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Peer Review",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.W, mana.U),
			}),
			Colors: []color.Color{color.White, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(fateholdChronologistToken),
						},
					},
					{
						Primitive: game.Surveil{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 2/2 colorless Wizard Soldier creature token named Cadet. Surveil 1.
		`,
		}),
	}
}

var fateholdChronologistToken = newFateholdChronologistToken()

func newFateholdChronologistToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cadet",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
