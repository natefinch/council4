package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GarlandRoyalKidnapper is the card definition for Garland, Royal Kidnapper.
//
// Type: Legendary Creature — Human Knight
// Cost: {2}{U}{B}
//
// Oracle text:
//
//	When Garland enters, target opponent becomes the monarch.
//	Whenever an opponent becomes the monarch, gain control of target creature that player controls for as long as they're the monarch.
//	Creatures you control but don't own get +2/+2 and can't be sacrificed.
var GarlandRoyalKidnapper = newGarlandRoyalKidnapper()

func newGarlandRoyalKidnapper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Garland, Royal Kidnapper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Knight},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, OwnerNotController: true}),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeSacrificed,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
							AffectedSelection:  game.Selection{OwnerNotController: true},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventBecameMonarch,
							Player: game.TriggerPlayerOpponent,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature that player controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ControlledByEventPlayer: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:         game.LayerControl,
											NewController: opt.Val(game.Player1),
											ExpiresForRef: opt.Val(game.EventPlayerReference()),
										},
									},
									Duration: game.DurationForAsLongAsPlayerIsMonarch,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Garland enters, target opponent becomes the monarch.
			Whenever an opponent becomes the monarch, gain control of target creature that player controls for as long as they're the monarch.
			Creatures you control but don't own get +2/+2 and can't be sacrificed.
		`,
		},
	}
}
