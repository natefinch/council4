package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Order is the card definition for Order // Chaos.
//
// Type: Instant // Instant
// Cost: {3}{W} // {2}{R}
// Face: Chaos — Instant ({2}{R})
//
// Oracle text:
//
//	Exile target attacking creature.
var Order = newOrder()

func newOrder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Order",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target attacking creature.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Chaos",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectCantBlock,
									PermanentTypes: []types.Card{types.Creature},
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures can't block this turn.
		`,
		}),
	}
}
