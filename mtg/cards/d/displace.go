package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Displace is the card definition for Displace.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Exile up to two target creatures you control, then return those cards to the battlefield under their owner's control.
var Displace = newDisplace

func newDisplace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Displace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target creatures you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(0),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("group-blink-0"),
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(1),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("group-blink-1"),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.LinkedBattlefieldSource(game.LinkedKey("group-blink-0")),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.LinkedBattlefieldSource(game.LinkedKey("group-blink-1")),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to two target creatures you control, then return those cards to the battlefield under their owner's control.
		`,
		},
	}
}
