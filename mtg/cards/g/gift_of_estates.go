package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GiftOfEstates is the card definition for Gift of Estates.
//
// Type: Sorcery
// Cost: {1}{W}
//
// Oracle text:
//
//	If an opponent controls more lands than you, search your library for up to three Plains cards, reveal them, put them into your hand, then shuffle.
var GiftOfEstates = newGiftOfEstates

func newGiftOfEstates() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gift of Estates",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
								Filter:      game.Selection{SubtypesAny: []types.Sub{types.Plains}},
								Reveal:      true,
							},
							Amount: game.Fixed(3),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlComparison: opt.Val(game.ControlCountComparison{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
									Left:      game.ControlPlayerAnyOpponent,
									Right:     game.ControlPlayerController,
									Op:        compare.GreaterThan,
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			If an opponent controls more lands than you, search your library for up to three Plains cards, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
