package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StormSWrath is the card definition for Storm's Wrath.
//
// Type: Sorcery
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Storm's Wrath deals 4 damage to each creature and each planeswalker.
var StormSWrath = newStormSWrath

func newStormSWrath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Storm's Wrath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Storm's Wrath deals 4 damage to each creature and each planeswalker.
		`,
		},
	}
}
