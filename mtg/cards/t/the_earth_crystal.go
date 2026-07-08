package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheEarthCrystal is the card definition for The Earth Crystal.
//
// Type: Legendary Artifact
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Green spells you cast cost {1} less to cast.
//	If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.
//	{4}{G}{G}, {T}: Distribute two +1/+1 counters among one or two target creatures you control.
var TheEarthCrystal = newTheEarthCrystal

func newTheEarthCrystal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "The Earth Crystal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.Green}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{4}{G}{G}, {T}: Distribute two +1/+1 counters among one or two target creatures you control.",
					ManaCost:        opt.Val(cost.Mana{cost.O(4), cost.G, cost.G}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 2,
								Constraint: "one or two target creatures you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.AllTargetPermanentsReference(0),
									CounterKind: counter.PlusOnePlusOne,
									Distribute:  true,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.CounterPlacementReplacement("If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.", 2, 0, counter.PlusOnePlusOne, game.TriggerControllerYou),
			},
			OracleText: `
			Green spells you cast cost {1} less to cast.
			If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.
			{4}{G}{G}, {T}: Distribute two +1/+1 counters among one or two target creatures you control.
		`,
		},
	}
}
