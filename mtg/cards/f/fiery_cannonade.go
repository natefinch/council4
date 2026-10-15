package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FieryCannonade is the card definition for Fiery Cannonade.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Fiery Cannonade deals 2 damage to each non-Pirate creature.
var FieryCannonade = newFieryCannonade

func newFieryCannonade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fiery Cannonade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Pirate")})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Fiery Cannonade deals 2 damage to each non-Pirate creature.
		`,
		},
	}
}
