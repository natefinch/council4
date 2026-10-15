package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RedemptionChoir is the card definition for Redemption Choir.
//
// Type: Creature — Vampire Cleric
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Lifelink
//	Coven — Whenever this creature enters or attacks, if you control three or more creatures with different powers, return target permanent card with mana value 3 or less from your graveyard to the battlefield.
var RedemptionChoir = newRedemptionChoir

func newRedemptionChoir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Redemption Choir",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Cleric},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Coven — Whenever this creature enters or attacks, if you control three or more creatures with different powers, return target permanent card with mana value 3 or less from your graveyard to the battlefield.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							UnionEvent:      game.EventAttackerDeclared,
						},
						InterveningIf: "if you control three or more creatures with different powers",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if you control three or more creatures with different powers",
							Aggregates: []game.AggregateComparison{
								game.AggregateComparison{
									Aggregate: game.AggregateControllerCreaturePowerDiversity,
									Op:        compare.GreaterOrEqual,
									Value:     3,
								},
							},
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent card with mana value 3 or less from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{
											Op:    compare.LessOrEqual,
											Value: 3,
										})}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
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
			Lifelink
			Coven — Whenever this creature enters or attacks, if you control three or more creatures with different powers, return target permanent card with mana value 3 or less from your graveyard to the battlefield.
		`,
		},
	}
}
