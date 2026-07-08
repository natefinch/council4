package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Conservator is the card definition for Conservator.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{3}, {T}: Prevent the next 2 damage that would be dealt to you this turn.
var Conservator = newConservator

func newConservator() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Conservator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{3}, {T}: Prevent the next 2 damage that would be dealt to you this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player: game.ControllerReference(),
									Amount: game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{3}, {T}: Prevent the next 2 damage that would be dealt to you this turn.
		`,
		},
	}
}
