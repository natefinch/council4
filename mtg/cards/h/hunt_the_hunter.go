package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HuntTheHunter is the card definition for Hunt the Hunter.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Target green creature you control gets +2/+2 until end of turn. It fights target green creature an opponent controls.
var HuntTheHunter = newHuntTheHunter()

func newHuntTheHunter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hunt the Hunter",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target green creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}, Controller: game.ControllerYou}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target green creature an opponent controls",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(2),
							Duration:       game.DurationUntilEndOfTurn,
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
			Target green creature you control gets +2/+2 until end of turn. It fights target green creature an opponent controls.
		`,
		},
	}
}
