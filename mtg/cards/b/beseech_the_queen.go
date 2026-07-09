package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BeseechTheQueen is the card definition for Beseech the Queen.
//
// Type: Sorcery
// Cost: {2/B}{2/B}{2/B}
//
// Oracle text:
//
//	({2/B} can be paid with any two mana or with {B}. This card's mana value is 6.)
//	Search your library for a card with mana value less than or equal to the number of lands you control, reveal it, put it into your hand, then shuffle.
var BeseechTheQueen = newBeseechTheQueen

func newBeseechTheQueen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Beseech the Queen",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.B),
				cost.Twobrid(mana.B),
				cost.Twobrid(mana.B),
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{Kind: game.DynamicAmountCountSelector, Multiplier: 1, Group: game.GroupRef(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}))})},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			({2/B} can be paid with any two mana or with {B}. This card's mana value is 6.)
			Search your library for a card with mana value less than or equal to the number of lands you control, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}
