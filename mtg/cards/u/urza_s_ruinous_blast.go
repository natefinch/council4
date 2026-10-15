package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrzaSRuinousBlast is the card definition for Urza's Ruinous Blast.
//
// Type: Legendary Sorcery
// Cost: {4}{W}
//
// Oracle text:
//
//	(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
//	Exile all nonland permanents that aren't legendary.
var UrzaSRuinousBlast = newUrzaSRuinousBlast

func newUrzaSRuinousBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Urza's Ruinous Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, ExcludedSupertype: types.Legendary}),
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
			Exile all nonland permanents that aren't legendary.
		`,
		},
	}
}
