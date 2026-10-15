package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HorriblyAwry is the card definition for Horribly Awry.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Counter target creature spell with mana value 4 or less. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
var HorriblyAwry = newHorriblyAwry

func newHorriblyAwry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Horribly Awry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
						Constraint: "target creature spell with mana value 4 or less",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypes:   []types.Card{types.Creature},
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
							ManaValue:        opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4}),
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
			Counter target creature spell with mana value 4 or less. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
		`,
		},
	}
}
