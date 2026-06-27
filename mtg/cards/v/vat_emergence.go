package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VatEmergence is the card definition for Vat Emergence.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	Put target creature card from a graveyard onto the battlefield under your control. Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var VatEmergence = newVatEmergence()

func newVatEmergence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vat Emergence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient: opt.Val(game.ControllerReference()),
						},
					},
					{
						Primitive: game.Proliferate{
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put target creature card from a graveyard onto the battlefield under your control. Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
