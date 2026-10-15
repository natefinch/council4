package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Persist is the card definition for Persist.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Return target nonlegendary creature card from your graveyard to the battlefield with a -1/-1 counter on it.
var Persist = newPersist

func newPersist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Persist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonlegendary creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSupertype: types.Legendary, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							EntryCounters: []game.CounterPlacement{
								game.CounterPlacement{
									Kind:   counter.MinusOneMinusOne,
									Amount: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target nonlegendary creature card from your graveyard to the battlefield with a -1/-1 counter on it.
		`,
		},
	}
}
