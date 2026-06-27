package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DestructiveForce is the card definition for Destructive Force.
//
// Type: Sorcery
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Each player sacrifices five lands of their choice. Destructive Force deals 5 damage to each creature.
var DestructiveForce = newDestructiveForce()

func newDestructiveForce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Destructive Force",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:      game.Fixed(5),
							PlayerGroup: game.AllPlayersReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices five lands of their choice. Destructive Force deals 5 damage to each creature.
		`,
		},
	}
}
