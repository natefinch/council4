package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Twincast is the card definition for Twincast.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	Copy target instant or sorcery spell. You may choose new targets for the copy.
var Twincast = newTwincast

func newTwincast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Twincast",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target instant or sorcery spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
							StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CopyStackObject{
							Object:              game.TargetStackObjectReference(0),
							Count:               1,
							MayChooseNewTargets: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Copy target instant or sorcery spell. You may choose new targets for the copy.
		`,
		},
	}
}
