package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Expansion is the card definition for Expansion // Explosion.
//
// Type: Instant // Instant
// Cost: {U/R}{U/R} // {X}{U}{U}{R}{R}
// Face: Explosion — Instant ({X}{U}{U}{R}{R})
//
// Oracle text:
//
//	Copy target instant or sorcery spell with mana value 4 or less. You may choose new targets for the copy.
var Expansion = newExpansion

func newExpansion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Expansion",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.R),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors: []color.Color{color.Blue, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target instant or sorcery spell with mana value 4 or less",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
							StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
							ManaValue:         opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4}),
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CopyStackObject{
							Object:              game.TargetStackObjectReference(0),
							Count:               1,
							MayChooseNewTargets: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Copy target instant or sorcery spell with mana value 4 or less. You may choose new targets for the copy.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Explosion",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Blue, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.TargetPlayerReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Explosion deals X damage to any target. Target player draws X cards.
		`,
		}),
	}
}
