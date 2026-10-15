package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JaheiraFriendOfTheForest is the card definition for Jaheira, Friend of the Forest.
//
// Type: Legendary Creature — Human Elf Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	Tokens you control have "{T}: Add {G}."
//	Choose a Background (You can have a Background as a second commander.)
var JaheiraFriendOfTheForest = newJaheiraFriendOfTheForest

func newJaheiraFriendOfTheForest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Jaheira, Friend of the Forest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Elf, types.Druid},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Tokens you control have \"{T}: Add {G}.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{TokenOnly: true}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaAbility(mana.G)),
							},
						},
					},
				},
				game.ChooseABackgroundStaticBody,
			},
			OracleText: `
			Tokens you control have "{T}: Add {G}."
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}
