package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RainOfBlades is the card definition for Rain of Blades.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Rain of Blades deals 1 damage to each attacking creature.
var RainOfBlades = newRainOfBlades

func newRainOfBlades() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rain of Blades",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Rain of Blades deals 1 damage to each attacking creature.
		`,
		},
	}
}
