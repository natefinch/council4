package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TwinningStaff is the card definition for Twinning Staff.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	If you would copy a spell one or more times, instead copy it that many times plus an additional time. You may choose new targets for the additional copy.
//	{7}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.
var TwinningStaff = newTwinningStaff

func newTwinningStaff() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Twinning Staff",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{7}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(7),
					}),
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
										Constraint: "target instant or sorcery spell you control",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
											StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
											Controller:        game.ControllerYou,
										},
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object:              game.TargetStackObjectReference(0),
											Count:               1,
											MayChooseNewTargets: true,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.AdditionalSpellCopyReplacement("If you would copy a spell one or more times, instead copy it that many times plus an additional time. You may choose new targets for the additional copy.", 1, true),
			},
			OracleText: `
			If you would copy a spell one or more times, instead copy it that many times plus an additional time. You may choose new targets for the additional copy.
			{7}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.
		`,
		},
	}
}
