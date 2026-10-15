package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArmyOfTheDamned is the card definition for Army of the Damned.
//
// Type: Sorcery
// Cost: {5}{B}{B}{B}
//
// Oracle text:
//
//	Create thirteen tapped 2/2 black Zombie creature tokens.
//	Flashback {7}{B}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ArmyOfTheDamned = newArmyOfTheDamned

func newArmyOfTheDamned() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Army of the Damned",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {7}{B}{B}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(7),
								cost.B,
								cost.B,
								cost.B,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:      game.Fixed(13),
							Source:      game.TokenDef(armyOfTheDamnedToken),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create thirteen tapped 2/2 black Zombie creature tokens.
			Flashback {7}{B}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var armyOfTheDamnedToken = newArmyOfTheDamnedToken()

func newArmyOfTheDamnedToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
