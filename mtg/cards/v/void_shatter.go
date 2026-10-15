package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VoidShatter is the card definition for Void Shatter.
//
// Type: Instant
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
var VoidShatter = newVoidShatter

func newVoidShatter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Void Shatter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Types: []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object:       game.TargetStackObjectReference(0),
							ExileInstead: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Devoid (This card has no color.)
			Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
		`,
		},
	}
}
