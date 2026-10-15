package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShootDown is the card definition for Shoot Down.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Exile target artifact, enchantment, or creature with flying.
var ShootDown = newShootDown

func newShootDown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Shoot Down",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}, game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}, game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Keyword: game.Flying}}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target artifact, enchantment, or creature with flying.
		`,
		},
	}
}
