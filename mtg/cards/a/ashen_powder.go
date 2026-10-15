package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AshenPowder is the card definition for Ashen Powder.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Put target creature card from an opponent's graveyard onto the battlefield under your control.
var AshenPowder = newAshenPowder

func newAshenPowder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ashen Powder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from an opponent's graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient: opt.Val(game.ControllerReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put target creature card from an opponent's graveyard onto the battlefield under your control.
		`,
		},
	}
}
