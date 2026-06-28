package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DissensionInTheRanks is the card definition for Dissension in the Ranks.
//
// Type: Instant
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Target blocking creature fights another target blocking creature.
var DissensionInTheRanks = newDissensionInTheRanks()

func newDissensionInTheRanks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dissension in the Ranks",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target blocking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateBlocking}),
					},
					game.TargetSpec{
						MinTargets:               1,
						MaxTargets:               1,
						Constraint:               "another target blocking creature",
						Allow:                    game.TargetAllowPermanent,
						Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateBlocking}),
						DistinctFromPriorTargets: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Fight{
							Object:        game.TargetPermanentReference(0),
							RelatedObject: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target blocking creature fights another target blocking creature.
		`,
		},
	}
}
