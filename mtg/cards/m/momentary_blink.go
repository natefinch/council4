package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MomentaryBlink is the card definition for Momentary Blink.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Exile target creature you control, then return it to the battlefield under its owner's control.
//	Flashback {3}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var MomentaryBlink = newMomentaryBlink

func newMomentaryBlink() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Momentary Blink",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.U,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(0),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("blink-1"),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target creature you control, then return it to the battlefield under its owner's control.
			Flashback {3}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
