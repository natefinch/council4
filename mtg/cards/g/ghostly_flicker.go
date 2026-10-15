package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GhostlyFlicker is the card definition for Ghostly Flicker.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Exile two target artifacts, creatures, and/or lands you control, then return those cards to the battlefield under your control.
var GhostlyFlicker = newGhostlyFlicker

func newGhostlyFlicker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ghostly Flicker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 2,
						MaxTargets: 2,
						Constraint: "two target artifacts, creatures, and/or lands you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}, Controller: game.ControllerYou}),
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
							Source:    game.LinkedBattlefieldSource(game.LinkedKey("group-blink-0")),
							Recipient: opt.Val(game.ControllerReference()),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source:    game.LinkedBattlefieldSource(game.LinkedKey("group-blink-1")),
							Recipient: opt.Val(game.ControllerReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile two target artifacts, creatures, and/or lands you control, then return those cards to the battlefield under your control.
		`,
		},
	}
}
