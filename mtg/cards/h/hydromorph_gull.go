package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HydromorphGull is the card definition for Hydromorph Gull.
//
// Type: Creature — Elemental Bird
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Flying
//	{U}, Sacrifice this creature: Counter target spell that targets a creature you control.
var HydromorphGull = newHydromorphGull()

func newHydromorphGull() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hydromorph Gull",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Bird},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
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
			Flying
			{U}, Sacrifice this creature: Counter target spell that targets a creature you control.
		`,
		},
	}
}
