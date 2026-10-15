package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SafewrightQuest is the card definition for Safewright Quest.
//
// Type: Sorcery
// Cost: {G/W}
//
// Oracle text:
//
//	Search your library for a Forest or Plains card, reveal it, put it into your hand, then shuffle.
var SafewrightQuest = newSafewrightQuest

func newSafewrightQuest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Safewright Quest",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{SubtypesAny: []types.Sub{types.Forest, types.Plains}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a Forest or Plains card, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}
