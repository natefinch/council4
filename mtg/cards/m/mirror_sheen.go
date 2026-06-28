package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MirrorSheen is the card definition for Mirror Sheen.
//
// Type: Enchantment
// Cost: {1}{U/R}{U/R}
//
// Oracle text:
//
//	{1}{U/R}{U/R}: Copy target instant or sorcery spell that targets you. You may choose new targets for the copy.
var MirrorSheen = newMirrorSheen()

func newMirrorSheen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Mirror Sheen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.U, mana.R),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{U/R}{U/R}: Copy target instant or sorcery spell that targets you. You may choose new targets for the copy.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.HybridMana(mana.U, mana.R), cost.HybridMana(mana.U, mana.R)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant or sorcery spell that targets you",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
									SpellTargets: []game.SpellTargetRequirement{game.SpellTargetRequirement{
										Kind:   game.SpellTargetRequirementPlayer,
										Player: game.PlayerYou,
									}},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyStackObject{
									Object:              game.TargetStackObjectReference(0),
									MayChooseNewTargets: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{1}{U/R}{U/R}: Copy target instant or sorcery spell that targets you. You may choose new targets for the copy.
		`,
		},
	}
}
