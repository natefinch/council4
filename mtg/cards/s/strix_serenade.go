package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StrixSerenade is the card definition for Strix Serenade.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Counter target artifact, creature, or planeswalker spell. Its controller creates a 2/2 blue Bird creature token with flying.
var StrixSerenade = newStrixSerenade

func newStrixSerenade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Strix Serenade",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact, creature, or planeswalker spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Artifact, types.Creature, types.Planeswalker},
							StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
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
							Amount:    game.Fixed(1),
							Source:    game.TokenDef(strixSerenadeToken),
							Recipient: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target artifact, creature, or planeswalker spell. Its controller creates a 2/2 blue Bird creature token with flying.
		`,
		},
	}
}

var strixSerenadeToken = newStrixSerenadeToken()

func newStrixSerenadeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
