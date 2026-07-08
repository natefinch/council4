package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Incubation is the card definition for Incubation // Incongruity.
//
// Type: Sorcery // Instant
// Cost: {G/U} // {1}{G}{U}
// Face: Incongruity — Instant ({1}{G}{U})
//
// Oracle text:
//
//	Look at the top five cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var Incubation = newIncubation

func newIncubation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Incubation",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors: []color.Color{color.Blue, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(5),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top five cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Incongruity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.Green},
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
						Primitive: game.Exile{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:    game.Fixed(1),
							Source:    game.TokenDef(incubationToken),
							Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target creature. That creature's controller creates a 3/3 green Frog Lizard creature token.
		`,
		}),
	}
}

var incubationToken = newIncubationToken()

func newIncubationToken() *game.CardDef {
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
