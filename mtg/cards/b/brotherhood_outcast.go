package b

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

// BrotherhoodOutcast is the card definition for Brotherhood Outcast.
//
// Type: Creature — Human Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Return target Aura or Equipment card with mana value 3 or less from your graveyard to the battlefield.
//	• Put a shield counter on target creature. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
var BrotherhoodOutcast = newBrotherhoodOutcast()

func newBrotherhoodOutcast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Brotherhood Outcast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Return target Aura or Equipment card with mana value 3 or less from your graveyard to the battlefield.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target Aura or Equipment card with mana value 3 or less from your graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Aura"), types.Sub("Equipment")}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
										},
									},
								},
							},
							game.Mode{
								Text: "Put a shield counter on target creature. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)",
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.Shield,
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
			When this creature enters, choose one —
			• Return target Aura or Equipment card with mana value 3 or less from your graveyard to the battlefield.
			• Put a shield counter on target creature. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
		`,
		},
	}
}
