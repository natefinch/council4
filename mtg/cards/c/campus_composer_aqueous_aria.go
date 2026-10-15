package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CampusComposer is the card definition for Campus Composer // Aqueous Aria.
//
// Type: Creature — Merfolk Bard // Sorcery
// Cost: {3}{U} // {4}{U}
// Face: Aqueous Aria — Sorcery ({4}{U})
//
// Oracle text:
//
//	Ward {2}
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var CampusComposer = newCampusComposer

func newCampusComposer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Campus Composer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:         []color.Color{color.Blue},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Merfolk, types.Bard},
			Power:          opt.Val(game.PT{Value: 3}),
			Toughness:      opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Ward {2}
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Aqueous Aria",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(campusComposerToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 3/3 blue and red Elemental creature token with flying.
		`,
		}),
	}
}

var campusComposerToken = newCampusComposerToken()

func newCampusComposerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Blue, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
