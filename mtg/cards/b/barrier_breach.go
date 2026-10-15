package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BarrierBreach is the card definition for Barrier Breach.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Exile up to three target enchantments.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var BarrierBreach = newBarrierBreach

func newBarrierBreach() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Barrier Breach",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 3,
						Constraint: "up to three target enchantments",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(1),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(2),
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to three target enchantments.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
