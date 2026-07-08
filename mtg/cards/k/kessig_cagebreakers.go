package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KessigCagebreakers is the card definition for Kessig Cagebreakers.
//
// Type: Creature — Human Rogue
// Cost: {4}{G}
//
// Oracle text:
//
//	Whenever this creature attacks, create a 2/2 green Wolf creature token that's tapped and attacking for each creature card in your graveyard.
var KessigCagebreakers = newKessigCagebreakers

func newKessigCagebreakers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kessig Cagebreakers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
										CardZone:   zone.Graveyard,
										Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
									}),
									Source:         game.TokenDef(kessigCagebreakersToken),
									EntryTapped:    true,
									EntryAttacking: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, create a 2/2 green Wolf creature token that's tapped and attacking for each creature card in your graveyard.
		`,
		},
	}
}

var kessigCagebreakersToken = newKessigCagebreakersToken()

func newKessigCagebreakersToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wolf",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
