package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeismicShudder is the card definition for Seismic Shudder.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Seismic Shudder deals 1 damage to each creature without flying.
var SeismicShudder = newSeismicShudder

func newSeismicShudder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Seismic Shudder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Seismic Shudder deals 1 damage to each creature without flying.
		`,
		},
	}
}
