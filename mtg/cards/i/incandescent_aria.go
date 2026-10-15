package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IncandescentAria is the card definition for Incandescent Aria.
//
// Type: Sorcery
// Cost: {R}{G}{W}
//
// Oracle text:
//
//	Incandescent Aria deals 3 damage to each nontoken creature.
var IncandescentAria = newIncandescentAria

func newIncandescentAria() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Incandescent Aria",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Incandescent Aria deals 3 damage to each nontoken creature.
		`,
		},
	}
}
