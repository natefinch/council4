package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CryptcallerChariot is the card definition for Cryptcaller Chariot.
//
// Type: Artifact — Vehicle
//
// Oracle text:
//
//	Menace
//	Whenever you discard one or more cards, create that many tapped 2/2 black Zombie creature tokens.
//	Crew 2
var CryptcallerChariot = newCryptcallerChariot

func newCryptcallerChariot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cryptcaller Chariot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(2),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:     game.EventCardDiscarded,
							Player:    game.TriggerPlayerYou,
							OneOrMore: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventCardCount,
										Multiplier: 1,
									}),
									Source:      game.TokenDef(cryptcallerChariotToken),
									EntryTapped: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace
			Whenever you discard one or more cards, create that many tapped 2/2 black Zombie creature tokens.
			Crew 2
		`,
		},
	}
}

var cryptcallerChariotToken = newCryptcallerChariotToken()

func newCryptcallerChariotToken() *game.CardDef {
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
