package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BeastWithin is the card definition for Beast Within.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Destroy target permanent. Its controller creates a 3/3 green Beast creature token.
var BeastWithin = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Beast Within",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Instant},
		OracleText: `
			Destroy target permanent. Its controller creates a 3/3 green Beast creature token.
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
						Primitive: game.Destroy{
							TargetIndex: 0,
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(beastWithinToken),
							Recipient: opt.Val(game.PlayerReference{
								Kind: game.PlayerReferenceObjectController,
								Object: opt.Val(game.ObjectReference{
									Kind:        game.ObjectReferenceTargetPermanent,
									TargetIndex: 0,
								}),
							}),
						},
					},
				},
			}.Ability(),
		),
	},
}

var beastWithinToken = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Beast",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
