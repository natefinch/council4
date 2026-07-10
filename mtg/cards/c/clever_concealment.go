package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CleverConcealment is the card definition for Clever Concealment.
//
// Type: Instant
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	Any number of target nonland permanents you control phase out. (Treat them and anything attached to them as though they don't exist until your next turn.)
var CleverConcealment = newCleverConcealment

func newCleverConcealment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Clever Concealment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target nonland permanents you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PhaseOut{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			Any number of target nonland permanents you control phase out. (Treat them and anything attached to them as though they don't exist until your next turn.)
		`,
		},
	}
}
