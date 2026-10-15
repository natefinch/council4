package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PleaForGuidance is the card definition for Plea for Guidance.
//
// Type: Sorcery
// Cost: {5}{W}
//
// Oracle text:
//
//	Search your library for up to two enchantment cards, reveal them, put them into your hand, then shuffle.
var PleaForGuidance = newPleaForGuidance

func newPleaForGuidance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Plea for Guidance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
								Reveal:      true,
							},
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to two enchantment cards, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
