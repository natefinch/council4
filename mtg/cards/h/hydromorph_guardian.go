package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HydromorphGuardian is the card definition for Hydromorph Guardian.
//
// Type: Creature — Elemental
// Cost: {2}{U}
//
// Oracle text:
//
//	{U}, Sacrifice this creature: Counter target spell that targets a creature you control.
var HydromorphGuardian = newHydromorphGuardian()

func newHydromorphGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hydromorph Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{U}, Sacrifice this creature: Counter target spell that targets a creature you control.",
					ManaCost: opt.Val(cost.Mana{cost.U}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target spell that targets a creature you control",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackSpell},
									SpellTargets: []game.SpellTargetRequirement{game.SpellTargetRequirement{
										Kind:          game.SpellTargetRequirementPermanent,
										RequiredTypes: []types.Card{types.Creature},
										Controller:    game.ControllerYou,
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
					}.Ability(),
				},
			},
			OracleText: `
			{U}, Sacrifice this creature: Counter target spell that targets a creature you control.
		`,
		},
	}
}
