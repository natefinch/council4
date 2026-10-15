package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UnbreakableBond is the card definition for Unbreakable Bond.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	Return target creature card from your graveyard to the battlefield with a lifelink counter on it.
var UnbreakableBond = newUnbreakableBond

func newUnbreakableBond() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Unbreakable Bond",
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
							EntryCounters: []game.CounterPlacement{
								game.CounterPlacement{
									Kind:   counter.Lifelink,
									Amount: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card from your graveyard to the battlefield with a lifelink counter on it.
		`,
		},
	}
}
