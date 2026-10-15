package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MarrowShards is the card definition for Marrow Shards.
//
// Type: Instant
// Cost: {W/P}
//
// Oracle text:
//
//	({W/P} can be paid with either {W} or 2 life.)
//	Marrow Shards deals 1 damage to each attacking creature.
var MarrowShards = newMarrowShards

func newMarrowShards() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Marrow Shards",
			ManaCost: opt.Val(cost.Mana{
				cost.PhyrexianMana(mana.W),
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
			({W/P} can be paid with either {W} or 2 life.)
			Marrow Shards deals 1 damage to each attacking creature.
		`,
		},
	}
}
