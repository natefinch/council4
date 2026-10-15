package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WillOfTheAllHunter is the card definition for Will of the All-Hunter.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Target creature gets +2/+2 until end of turn. If it's blocking, instead put two +1/+1 counters on it.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var WillOfTheAllHunter = newWillOfTheAllHunter

func newWillOfTheAllHunter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Will of the All-Hunter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			SpellAbility: opt.Val(game.Mode{
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
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(2),
							Duration:       game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:        true,
								Object:        opt.Val(game.TargetPermanentReference(0)),
								ObjectMatches: opt.Val(game.Selection{CombatState: game.CombatStateBlocking}),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +2/+2 until end of turn. If it's blocking, instead put two +1/+1 counters on it.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
