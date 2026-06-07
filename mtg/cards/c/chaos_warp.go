package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChaosWarp is the card definition for Chaos Warp.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.
var ChaosWarp = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name: "Chaos Warp",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.R,
		}),
		Colors: []color.Color{color.Red},
		Types:  []types.Card{types.Instant},
		OracleText: `
			The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.
		`,
		SpellAbility: opt.Val(
			game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "permanent",
						Allow:      game.TargetAllowPermanent,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ShufflePermanentIntoLibrary{
							TargetIndex: 0,
						},
					},
					{
						Primitive: game.Reveal{
							Amount:      game.Fixed(1),
							TargetIndex: 0,
							Recipient: opt.Val(game.PlayerReference{
								Kind: game.PlayerReferenceObjectOwner,
								Object: opt.Val(game.ObjectReference{
									Kind:        game.ObjectReferenceTargetPermanent,
									TargetIndex: 0,
								}),
							}),
							PublishLinked: game.LinkedKey("chaos-warp-revealed"),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							TargetIndex: 0,
							Source:      game.LinkedBattlefieldSource(game.LinkedKey("chaos-warp-revealed")),
							Recipient: opt.Val(game.PlayerReference{
								Kind: game.PlayerReferenceObjectOwner,
								Object: opt.Val(game.ObjectReference{
									Kind:        game.ObjectReferenceTargetPermanent,
									TargetIndex: 0,
								}),
							}),
						},
						CardCondition: opt.Val(game.CardCondition{
							Card: game.CardReference{
								Kind:   game.CardReferenceLinked,
								LinkID: "chaos-warp-revealed",
							},
							RequirePermanentCard: true,
						}),
					},
				},
			}.Ability(),
		),
	},
}
