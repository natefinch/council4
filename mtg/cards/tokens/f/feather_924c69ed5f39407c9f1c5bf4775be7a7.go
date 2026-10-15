package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Feather is the card definition for Feather.
//
// Type: Token Artifact
//
// Oracle text:
//   {1}, Sacrifice Feather: Return target Phoenix card from your graveyard to the battlefield tapped.

// FeatherToken924c69ed5f39407c9f1c5bf4775be7a7 is the card definition for Feather.
var FeatherToken924c69ed5f39407c9f1c5bf4775be7a7 = newFeatherToken924c69ed5f39407c9f1c5bf4775be7a7()

func newFeatherToken924c69ed5f39407c9f1c5bf4775be7a7() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:   "Feather",
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, Sacrifice Feather: Return target Phoenix card from your graveyard to the battlefield tapped.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice Feather",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target Phoenix card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Phoenix}, Controller: game.ControllerYou}),
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
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{1}, Sacrifice Feather: Return target Phoenix card from your graveyard to the battlefield tapped.
		`,
		},
	}
}
