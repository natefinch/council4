package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BreathWeapon is the card definition for Breath Weapon.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Breath Weapon deals 2 damage to each non-Dragon creature.
var BreathWeapon = newBreathWeapon

func newBreathWeapon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Breath Weapon",
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
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Dragon")})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Breath Weapon deals 2 damage to each non-Dragon creature.
		`,
		},
	}
}
