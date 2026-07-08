package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BoundlessRealms is the card definition for Boundless Realms.
//
// Type: Sorcery
// Cost: {6}{G}
//
// Oracle text:
//
//	Search your library for up to X basic land cards, where X is the number of lands you control, put them onto the battlefield tapped, then shuffle.
var BoundlessRealms = newBoundlessRealms

func newBoundlessRealms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Boundless Realms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
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
								SourceZone:   zone.Library,
								Destination:  zone.Battlefield,
								Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
								EntersTapped: true,
							},
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to X basic land cards, where X is the number of lands you control, put them onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
