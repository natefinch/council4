package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AttuneWithAether is the card definition for Attune with Aether.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Search your library for a basic land card, reveal it, put it into your hand, then shuffle. You get {E}{E} (two energy counters).
var AttuneWithAether = newAttuneWithAether

func newAttuneWithAether() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Attune with Aether",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
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
			Search your library for a basic land card, reveal it, put it into your hand, then shuffle. You get {E}{E} (two energy counters).
		`,
		},
	}
}
