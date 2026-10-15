package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiantAnkheg is the card definition for Giant Ankheg.
//
// Type: Creature — Insect
// Cost: {6}{G}{G}
//
// Oracle text:
//
//	Trample
//	Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
//	Other creatures you control have trample and ward {2}.
var GiantAnkheg = newGiantAnkheg

func newGiantAnkheg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Giant Ankheg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
				game.StaticAbility{
					Text: "Other creatures you control have trample and ward {2}.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(2)})),
							},
						},
					},
				},
			},
			OracleText: `
			Trample
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
			Other creatures you control have trample and ward {2}.
		`,
		},
	}
}
