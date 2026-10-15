package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IzzetGuildmage is the card definition for Izzet Guildmage.
//
// Type: Creature — Human Wizard
// Cost: {U/R}{U/R}
//
// Oracle text:
//
//	({U/R} can be paid with either {U} or {R}.)
//	{2}{U}: Copy target instant spell you control with mana value 2 or less. You may choose new targets for the copy.
//	{2}{R}: Copy target sorcery spell you control with mana value 2 or less. You may choose new targets for the copy.
var IzzetGuildmage = newIzzetGuildmage

func newIzzetGuildmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Izzet Guildmage",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.R),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{U}: Copy target instant spell you control with mana value 2 or less. You may choose new targets for the copy.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target instant spell you control with mana value 2 or less",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											SpellCardTypes:   []types.Card{types.Instant},
											StackObjectKinds: []game.StackObjectKind{game.StackSpell},
											Controller:       game.ControllerYou,
											ManaValue: opt.Val(compare.Int{
												Op:    compare.LessOrEqual,
												Value: 2,
											}),
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
				game.ActivatedAbility{
					Text: "{2}{R}: Copy target sorcery spell you control with mana value 2 or less. You may choose new targets for the copy.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target sorcery spell you control with mana value 2 or less",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											SpellCardTypes:   []types.Card{types.Sorcery},
											StackObjectKinds: []game.StackObjectKind{game.StackSpell},
											Controller:       game.ControllerYou,
											ManaValue: opt.Val(compare.Int{
												Op:    compare.LessOrEqual,
												Value: 2,
											}),
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
			OracleText: `
			({U/R} can be paid with either {U} or {R}.)
			{2}{U}: Copy target instant spell you control with mana value 2 or less. You may choose new targets for the copy.
			{2}{R}: Copy target sorcery spell you control with mana value 2 or less. You may choose new targets for the copy.
		`,
		},
	}
}
