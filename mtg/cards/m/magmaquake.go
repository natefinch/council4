package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Magmaquake is the card definition for Magmaquake.
//
// Type: Instant
// Cost: {X}{R}{R}
//
// Oracle text:
//
//	Magmaquake deals X damage to each creature without flying and each planeswalker.
var Magmaquake = newMagmaquake

func newMagmaquake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Magmaquake",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
						},
					},
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Magmaquake deals X damage to each creature without flying and each planeswalker.
		`,
		},
	}
}
