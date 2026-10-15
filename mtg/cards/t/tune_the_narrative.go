package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TuneTheNarrative is the card definition for Tune the Narrative.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Draw a card. You get {E}{E} (two energy counters).
var TuneTheNarrative = newTuneTheNarrative

func newTuneTheNarrative() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tune the Narrative",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.AddPlayerCounter{
							Amount:      game.Fixed(2),
							Player:      game.ControllerReference(),
							CounterKind: counter.Energy,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card. You get {E}{E} (two energy counters).
		`,
		},
	}
}
