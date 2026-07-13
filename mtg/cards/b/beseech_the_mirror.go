package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BeseechTheMirror is the card definition for Beseech the Mirror.
//
// Type: Sorcery
// Cost: {1}{B}{B}{B}
//
// Oracle text:
//
//	Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
//	Search your library for a card, exile it face down, then shuffle. If this spell was bargained, you may cast the exiled card without paying its mana cost if that spell's mana value is 4 or less. Put the exiled card into your hand if it wasn't cast this way.
var BeseechTheMirror = newBeseechTheMirror

func newBeseechTheMirror() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Beseech the Mirror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.BargainStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: "Search your library for a card, exile it face down, then shuffle. If this spell was bargained, you may cast the exiled card without paying its mana cost if that spell's mana value is 4 or less. Put the exiled card into your hand if it wasn't cast this way.",
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:    zone.Library,
								Destination:   zone.Exile,
								ExileFaceDown: true,
							},
							Amount:        game.Fixed(1),
							PublishLinked: game.LinkedKey("beseech-exiled"),
						},
					},
					{
						Primitive: game.CastForFree{
							Player: game.ControllerReference(),
							Zone:   zone.Exile,
							Card:   game.CardReference{Kind: game.CardReferenceLinked, LinkID: "beseech-exiled"},
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasBargained: true,
							}),
						}),
						CardCondition: opt.Val(game.CardSelection{
							Card:      game.CardReference{Kind: game.CardReferenceLinked, LinkID: "beseech-exiled"},
							Selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4})},
						}),
						Optional: true,
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceLinked, LinkID: "beseech-exiled"},
							FromZone:    zone.Exile,
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
			Search your library for a card, exile it face down, then shuffle. If this spell was bargained, you may cast the exiled card without paying its mana cost if that spell's mana value is 4 or less. Put the exiled card into your hand if it wasn't cast this way.
		`,
		},
	}
}
