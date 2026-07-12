package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LongRiverSPull is the card definition for Long River's Pull.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	Gift a card (You may promise an opponent a gift as you cast this spell. If you do, they draw a card before its other effects.)
//	Counter target creature spell. If the gift was promised, instead counter target spell.
var LongRiverSPull = newLongRiverSPull

func newLongRiverSPull() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Long River's Pull",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.GiftKeyword{Delivery: game.Mode{
							Sequence: []game.Instruction{
								{
									Primitive: game.Draw{
										Amount: game.Fixed(1),
										Player: game.GiftRecipientReference(),
									},
								},
							},
						}.Ability()},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypes:   []types.Card{types.Creature},
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
						Gate: game.TargetGateGiftNotPromised,
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
						Gate: game.TargetGateGiftPromised,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								GiftPromised: true,
							}),
						}),
					},
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(1),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								GiftPromised: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gift a card (You may promise an opponent a gift as you cast this spell. If you do, they draw a card before its other effects.)
			Counter target creature spell. If the gift was promised, instead counter target spell.
		`,
		},
	}
}
