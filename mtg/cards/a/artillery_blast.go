package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArtilleryBlast is the card definition for Artillery Blast.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Domain — Artillery Blast deals X damage to target tapped creature, where X is 1 plus the number of basic land types among lands you control.
var ArtilleryBlast = newArtilleryBlast

func newArtilleryBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Artillery Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target tapped creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Tapped: game.TriTrue}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier: 1,
								Addend:     1,
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Domain — Artillery Blast deals X damage to target tapped creature, where X is 1 plus the number of basic land types among lands you control.
		`,
		},
	}
}
