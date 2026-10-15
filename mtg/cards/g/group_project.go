package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GroupProject is the card definition for Group Project.
//
// Type: Sorcery
// Cost: {1}{W}
//
// Oracle text:
//
//	Create a 2/2 red and white Spirit creature token.
//	Flashback—Tap three untapped creatures you control. (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var GroupProject = newGroupProject

func newGroupProject() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Group Project",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Flashback,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Flashback",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap three untapped creatures you control",
							Amount:             3,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
						},
					},
					Mechanic: cost.AlternativeMechanicFlashback,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(groupProjectToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 2/2 red and white Spirit creature token.
			Flashback—Tap three untapped creatures you control. (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var groupProjectToken = newGroupProjectToken()

func newGroupProjectToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
