package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MuddleTheMixture is the card definition for Muddle the Mixture.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	Counter target instant or sorcery spell.
//	Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var MuddleTheMixture = newMuddleTheMixture

func newMuddleTheMixture() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Muddle the Mixture",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.U, cost.U}, 2),
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target instant or sorcery spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
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
				},
			}.Ability()),
			OracleText: `
			Counter target instant or sorcery spell.
			Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
