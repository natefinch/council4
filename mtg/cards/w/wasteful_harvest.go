package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WastefulHarvest is the card definition for Wasteful Harvest.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Mill five cards. You may put a permanent card from among the cards milled this way into your hand. (To mill a card, put the top card of your library into your graveyard.)
var WastefulHarvest = newWastefulHarvest

func newWastefulHarvest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wasteful Harvest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:   game.ControllerReference(),
							Look:     game.Fixed(5),
							Take:     game.Fixed(1),
							Filter:   opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}}),
							TakeUpTo: true,
							Reveal:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mill five cards. You may put a permanent card from among the cards milled this way into your hand. (To mill a card, put the top card of your library into your graveyard.)
		`,
		},
	}
}
