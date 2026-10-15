package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SynthInfiltrator is the card definition for Synth Infiltrator.
//
// Type: Artifact Creature — Synth
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
//	You may have this creature enter as a copy of any creature on the battlefield, except it's a Synth artifact creature in addition to its other types.
var SynthInfiltrator = newSynthInfiltrator

func newSynthInfiltrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Synth Infiltrator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Synth},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.ImproviseStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any creature on the battlefield, except it's a Synth artifact creature in addition to its other types.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, false, nil, []types.Sub{types.Synth}, types.Artifact, types.Creature),
			},
			OracleText: `
			Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
			You may have this creature enter as a copy of any creature on the battlefield, except it's a Synth artifact creature in addition to its other types.
		`,
		},
	}
}
