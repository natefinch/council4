package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
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
	Name: "Chaos Warp",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Red),
	}),
	Colors:        []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Types:         []types.Card{types.Instant},
	OracleText:    "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectShufflePermanentIntoLibrary,
					TargetIndex: 0,
				},
				{
					Type:        game.EffectReveal,
					Amount:      1,
					TargetIndex: 0,
					LinkID:      "chaos-warp-revealed",
					Recipient: opt.Val(game.PlayerReference{
						Kind: game.PlayerReferenceObjectOwner,
						Object: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceTargetPermanent,
							TargetIndex: 0,
						}),
					}),
				},
				{
					Type:        game.EffectPutOnBattlefield,
					TargetIndex: 0,
					LinkID:      "chaos-warp-revealed",
					Recipient: opt.Val(game.PlayerReference{
						Kind: game.PlayerReferenceObjectOwner,
						Object: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceTargetPermanent,
							TargetIndex: 0,
						}),
					}),
					CardCondition: opt.Val(game.CardCondition{
						Card: game.CardReference{
							Kind:   game.CardReferenceLinked,
							LinkID: "chaos-warp-revealed",
						},
						RequirePermanentCard: true,
					}),
				},
			},
		},
	},
}
