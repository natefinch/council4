package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IllusionistSStratagem is the card definition for Illusionist's Stratagem.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Exile up to two target creatures you control, then return those cards to the battlefield under their owner's control.
//	Draw a card.
var IllusionistSStratagem = newIllusionistSStratagem

func newIllusionistSStratagem() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Illusionist's Stratagem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to two target creatures you control, then return those cards to the battlefield under their owner's control.
			Draw a card.
		`,
		},
	}
}
