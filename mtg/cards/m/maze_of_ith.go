package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MazeOfIth is the card definition for Maze of Ith.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.
var MazeOfIth = newMazeOfIth

func newMazeOfIth() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Maze of Ith",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target attacking creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.PreventDamage{
											Object:     game.TargetPermanentReference(0),
											All:        true,
											CombatOnly: true,
										},
									},
									game.Instruction{
										Primitive: game.PreventDamage{
											Object:     game.TargetPermanentReference(0),
											All:        true,
											CombatOnly: true,
											BySource:   true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.
		`,
		},
	}
}
