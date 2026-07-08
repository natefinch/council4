package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GracefulRestoration is the card definition for Graceful Restoration.
//
// Type: Sorcery
// Cost: {3}{W}{B}
//
// Oracle text:
//
//	Choose one —
//	• Return target creature card from your graveyard to the battlefield with an additional +1/+1 counter on it.
//	• Return up to two target creature cards with power 2 or less from your graveyard to the battlefield.
var GracefulRestoration = newGracefulRestoration

func newGracefulRestoration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Graceful Restoration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Return target creature card from your graveyard to the battlefield with an additional +1/+1 counter on it.",
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
									Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									EntryCounters: []game.CounterPlacement{game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}},
								},
							},
						},
					},
					game.Mode{
						Text: "Return up to two target creature cards with power 2 or less from your graveyard to the battlefield.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target creature cards with power 2 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Return target creature card from your graveyard to the battlefield with an additional +1/+1 counter on it.
			• Return up to two target creature cards with power 2 or less from your graveyard to the battlefield.
		`,
		},
	}
}
