package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VigilantMartyr is the card definition for Vigilant Martyr.
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//
//	Sacrifice this creature: Regenerate target creature.
//	{W}{W}, {T}, Sacrifice this creature: Counter target spell that targets an enchantment.
var VigilantMartyr = newVigilantMartyr

func newVigilantMartyr() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Vigilant Martyr",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice this creature: Regenerate target creature.",
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
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{W}{W}, {T}, Sacrifice this creature: Counter target spell that targets an enchantment.",
					ManaCost: opt.Val(cost.Mana{cost.W, cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
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
								Constraint: "target spell that targets an enchantment",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackSpell},
									SpellTargets: []game.SpellTargetRequirement{game.SpellTargetRequirement{
										Kind:          game.SpellTargetRequirementPermanent,
										RequiredTypes: []types.Card{types.Enchantment},
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
			Sacrifice this creature: Regenerate target creature.
			{W}{W}, {T}, Sacrifice this creature: Counter target spell that targets an enchantment.
		`,
		},
	}
}
