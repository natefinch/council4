package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrimsonCowlMasterOfEvil is the card definition for Crimson Cowl, Master of Evil.
//
// Type: Legendary Creature — Human Villain
// Cost: {2}{B}
//
// Oracle text:
//
//	Whenever one or more nontoken Villains you control attack a player, you create a 2/1 black Villain creature token with menace. (It can't be blocked except by two or more creatures.)
var CrimsonCowlMasterOfEvil = newCrimsonCowlMasterOfEvil()

func newCrimsonCowlMasterOfEvil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crimson Cowl, Master of Evil",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Villain},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Controller:       game.TriggerControllerYou,
							OneOrMore:        true,
							AttackRecipient:  game.AttackRecipientPlayer,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Villain")}, NonToken: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(crimsonCowlMasterOfEvilToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever one or more nontoken Villains you control attack a player, you create a 2/1 black Villain creature token with menace. (It can't be blocked except by two or more creatures.)
		`,
		},
	}
}

var crimsonCowlMasterOfEvilToken = newCrimsonCowlMasterOfEvilToken()

func newCrimsonCowlMasterOfEvilToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Villain",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}
