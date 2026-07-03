package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TectonicRift is the card definition for Tectonic Rift.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Destroy target land. Creatures without flying can't block this turn.
var TectonicRift = newTectonicRift()

func newTectonicRift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tectonic Rift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:              game.RuleEffectCantBlock,
									PermanentTypes:    []types.Card{types.Creature},
									AffectedSelection: game.Selection{ExcludedKeyword: game.Flying},
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target land. Creatures without flying can't block this turn.
		`,
		},
	}
}
