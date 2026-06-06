package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NatureSLore is the card definition for Nature's Lore.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Search your library for a Forest card, put that card onto the battlefield, then shuffle.
var NatureSLore = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Nature's Lore",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Search your library for a Forest card, put that card onto the battlefield, then shuffle.
		`,
		SpellAbility: opt.Val(game.SpellAbilityBody{
			Text: `
				Search your library for a Forest card, put that card onto the battlefield, then shuffle.
			`,
			Content: game.PlainAbilityContent{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							TargetIndex: game.TargetIndexController,
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								CardType:    opt.Val(types.Land),
								SubtypesAny: []types.Sub{
									types.Forest,
								},
								Shuffle: true,
							},
						},
					},
				},
			},
		}),
	},
}
