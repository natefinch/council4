package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummonerSBane is the card definition for Summoner's Bane.
//
// Type: Instant
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Counter target creature spell. Create a 2/2 blue Illusion creature token.
var SummonerSBane = newSummonerSBane()

func newSummonerSBane() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Summoner's Bane",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypes:   []types.Card{types.Creature},
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(summonerSBaneToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target creature spell. Create a 2/2 blue Illusion creature token.
		`,
		},
	}
}

var summonerSBaneToken = newSummonerSBaneToken()

func newSummonerSBaneToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Illusion",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
