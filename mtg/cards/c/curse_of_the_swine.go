package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfTheSwine is the card definition for Curse of the Swine.
//
// Type: Sorcery
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Exile X target creatures. For each creature exiled this way, its controller creates a 2/2 green Boar creature token.
var CurseOfTheSwine = newCurseOfTheSwine

func newCurseOfTheSwine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Curse of the Swine",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "target creatures",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.RemoveTargetsForToken{
							Exile:     true,
							LinkedKey: game.LinkedKey("removed-targets-for-token"),
						},
					},
					{
						Primitive: game.CreateTokenForEachDestroyed{
							Source:    game.TokenDef(curseOfTheSwineToken),
							LinkedKey: game.LinkedKey("removed-targets-for-token"),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile X target creatures. For each creature exiled this way, its controller creates a 2/2 green Boar creature token.
		`,
		},
	}
}

var curseOfTheSwineToken = newCurseOfTheSwineToken()

func newCurseOfTheSwineToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Boar",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Boar},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
