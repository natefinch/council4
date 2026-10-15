package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UndeadGladiator is the card definition for Undead Gladiator.
//
// Type: Creature — Zombie Barbarian
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	{1}{B}, Discard a card: Return this card from your graveyard to your hand. Activate only during your upkeep.
//	Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)
var UndeadGladiator = newUndeadGladiator

func newUndeadGladiator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Undead Gladiator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Barbarian},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B}, Discard a card: Return this card from your graveyard to your hand. Activate only during your upkeep.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard a card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.DuringUpkeep,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.B}),
			},
			OracleText: `
			{1}{B}, Discard a card: Return this card from your graveyard to your hand. Activate only during your upkeep.
			Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)
		`,
		},
	}
}
