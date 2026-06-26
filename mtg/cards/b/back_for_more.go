package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BackForMore is the card definition for Back for More.
//
// Type: Instant
// Cost: {4}{B}{G}
//
// Oracle text:
//
//	Return target creature card from your graveyard to the battlefield. When you do, it fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
var BackForMore = newBackForMore()

func newBackForMore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Back for More",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Instant},
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
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target creature you don't control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						},
					},
					{
						Primitive: game.Fight{
							Object:        game.TargetPermanentReference(0),
							RelatedObject: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card from your graveyard to the battlefield. When you do, it fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
		`,
		},
	}
}
