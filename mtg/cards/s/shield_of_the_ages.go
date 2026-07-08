package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShieldOfTheAges is the card definition for Shield of the Ages.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{2}: Prevent the next 1 damage that would be dealt to you this turn.
var ShieldOfTheAges = newShieldOfTheAges

func newShieldOfTheAges() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Shield of the Ages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}: Prevent the next 1 damage that would be dealt to you this turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player: game.ControllerReference(),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}: Prevent the next 1 damage that would be dealt to you this turn.
		`,
		},
	}
}
