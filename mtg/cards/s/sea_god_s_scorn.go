package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SeaGodSScorn is the card definition for Sea God's Scorn.
//
// Type: Sorcery
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Return up to three target creatures and/or enchantments to their owners' hands.
var SeaGodSScorn = newSeaGodSScorn

func newSeaGodSScorn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sea God's Scorn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 3,
						Constraint: "up to three target creatures and/or enchantments",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(1),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(2),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return up to three target creatures and/or enchantments to their owners' hands.
		`,
		},
	}
}
