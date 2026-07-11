package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LegionLoyalty is the card definition for Legion Loyalty.
//
// Type: Enchantment
// Cost: {6}{W}{W}
//
// Oracle text:
//
//	Creatures you control have myriad. (Whenever a creature with myriad attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
var LegionLoyalty = newLegionLoyalty

func newLegionLoyalty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Legion Loyalty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							AddAbilities: []game.Ability{
								new(game.MyriadTriggeredBody),
							},
						},
					},
				},
			},
			OracleText: `
			Creatures you control have myriad. (Whenever a creature with myriad attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
		`,
		},
	}
}
