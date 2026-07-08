package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Shriekmaw is the card definition for Shriekmaw.
//
// Type: Creature — Elemental
// Cost: {4}{B}
//
// Oracle text:
//
//	Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	When this creature enters, destroy target nonartifact, nonblack creature.
//	Evoke {1}{B} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
var Shriekmaw = newShriekmaw

func newShriekmaw() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shriekmaw",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FearStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonartifact, nonblack creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Artifact}, ExcludedColors: []color.Color{color.Black}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.EvokeSacrificeTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Evoke",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					Mechanic: cost.AlternativeMechanicEvoke,
				},
			},
			OracleText: `
			Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
			When this creature enters, destroy target nonartifact, nonblack creature.
			Evoke {1}{B} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
		`,
		},
	}
}
