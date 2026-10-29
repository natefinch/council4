package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Misdirection is the card definition for Misdirection.
//
// Type: Instant
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	You may exile a blue card from your hand rather than pay this spell's mana cost.
//	Change the target of target spell with a single target.
var Misdirection = newMisdirection

func newMisdirection() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Misdirection",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a blue card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Blue,
						},
					},
				},
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
							SingleTarget:     true,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseNewTargets{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may exile a blue card from your hand rather than pay this spell's mana cost.
			Change the target of target spell with a single target.
		`,
		},
	}
}
