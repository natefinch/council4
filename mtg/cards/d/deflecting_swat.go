package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeflectingSwat is the card definition for Deflecting Swat.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	If you control a commander, you may cast this spell without paying its mana cost.
//	You may choose new targets for target spell or ability.
var DeflectingSwat = newDeflectingSwat

func newDeflectingSwat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Deflecting Swat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Cast without paying mana cost",
					Condition: cost.AlternativeConditionControlsCommander,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell or ability",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseNewTargets{
							Object: game.TargetStackObjectReference(0),
							Rule:   game.RetargetChooseNewTargets,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			If you control a commander, you may cast this spell without paying its mana cost.
			You may choose new targets for target spell or ability.
		`,
		},
	}
}
