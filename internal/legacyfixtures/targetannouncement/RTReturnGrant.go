// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTReturnGrant is the card definition for RT Return Grant.
//
// Type: Instant
// Cost: {0}
//
// Oracle text:
//
//	Return target creature card from your graveyard to the battlefield.
var RTReturnGrant = newRTReturnGrant

func newRTReturnGrant() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "RT Return Grant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card from your graveyard to the battlefield.
		`,
		},
	}
}
