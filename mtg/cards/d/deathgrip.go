package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Deathgrip is the card definition for Deathgrip.
//
// Type: Enchantment
// Cost: {B}{B}
//
// Oracle text:
//
//	{B}{B}: Counter target green spell.
var Deathgrip = newDeathgrip

func newDeathgrip() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Deathgrip",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}{B}: Counter target green spell.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target green spell",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											StackObjectKinds: []game.StackObjectKind{game.StackSpell},
											SpellColors:      []color.Color{color.Green},
										},
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CounterObject{
											Object: game.TargetStackObjectReference(0),
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
			{B}{B}: Counter target green spell.
		`,
		},
	}
}
