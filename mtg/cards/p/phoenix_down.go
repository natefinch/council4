package p

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

// PhoenixDown is the card definition for Phoenix Down.
//
// Type: Artifact
// Cost: {W}
//
// Oracle text:
//
//	{1}{W}, {T}, Exile this artifact: Choose one —
//	• Return target creature card with mana value 4 or less from your graveyard to the battlefield tapped.
//	• Exile target Skeleton, Spirit, or Zombie.
var PhoenixDown = newPhoenixDown

func newPhoenixDown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Phoenix Down",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{W}, {T}, Exile this artifact: Choose one —\n• Return target creature card with mana value 4 or less from your graveyard to the battlefield tapped.\n• Exile target Skeleton, Spirit, or Zombie.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.W,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this artifact",
							Amount:      1,
							Source:      zone.Battlefield,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Return target creature card with mana value 4 or less from your graveyard to the battlefield tapped.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card with mana value 4 or less from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{
											Op:    compare.LessOrEqual,
											Value: 4,
										})}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
											EntryTapped: true,
										},
									},
								},
							},
							game.Mode{
								Text: "Exile target Skeleton, Spirit, or Zombie.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target Skeleton, Spirit, or Zombie",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Skeleton, types.Spirit, types.Zombie}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Exile,
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
			{1}{W}, {T}, Exile this artifact: Choose one —
			• Return target creature card with mana value 4 or less from your graveyard to the battlefield tapped.
			• Exile target Skeleton, Spirit, or Zombie.
		`,
		},
	}
}
