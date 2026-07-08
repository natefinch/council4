package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wight is the card definition for Wight.
//
// Type: Creature — Zombie Soldier
// Cost: {1}{B}
//
// Oracle text:
//
//	This creature enters tapped.
//	Life Drain — Whenever a creature dealt damage by this creature this turn dies, create a tapped 2/2 black Zombie creature token and exile that card.
var Wight = newWight

func newWight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Wight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentDied,
							DyingDamagedBySource: true,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:      game.Fixed(1),
									Source:      game.TokenDef(wightToken),
									EntryTapped: true,
								},
							},
							{
								Primitive: game.Exile{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			OracleText: `
			This creature enters tapped.
			Life Drain — Whenever a creature dealt damage by this creature this turn dies, create a tapped 2/2 black Zombie creature token and exile that card.
		`,
		},
	}
}

var wightToken = newWightToken()

func newWightToken() *game.CardDef {
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
