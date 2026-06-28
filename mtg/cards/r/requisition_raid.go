package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RequisitionRaid is the card definition for Requisition Raid.
//
// Type: Sorcery
// Cost: {W}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {1} — Destroy target artifact.
//	+ {1} — Destroy target enchantment.
//	+ {1} — Put a +1/+1 counter on each creature target player controls.
var RequisitionRaid = newRequisitionRaid()

func newRequisitionRaid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Requisition Raid",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{1} — Destroy target artifact.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "{1} — Destroy target enchantment.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target enchantment",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "{1} — Put a +1/+1 counter on each creature target player controls.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 3,
			}),
			OracleText: `
			Spree (Choose one or more additional costs.)
			+ {1} — Destroy target artifact.
			+ {1} — Destroy target enchantment.
			+ {1} — Put a +1/+1 counter on each creature target player controls.
		`,
		},
	}
}
