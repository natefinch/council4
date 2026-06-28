package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DawnCharm is the card definition for Dawn Charm.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Choose one —
//	• Prevent all combat damage that would be dealt this turn.
//	• Regenerate target creature.
//	• Counter target spell that targets you.
var DawnCharm = newDawnCharm()

func newDawnCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Dawn Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Prevent all combat damage that would be dealt this turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									All:        true,
									CombatOnly: true,
									Global:     true,
								},
							},
						},
					},
					game.Mode{
						Text: "Regenerate target creature.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Counter target spell that targets you.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target spell that targets you",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackSpell},
									SpellTargets: []game.SpellTargetRequirement{game.SpellTargetRequirement{
										Kind:   game.SpellTargetRequirementPlayer,
										Player: game.PlayerYou,
									}},
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
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Prevent all combat damage that would be dealt this turn.
			• Regenerate target creature.
			• Counter target spell that targets you.
		`,
		},
	}
}
