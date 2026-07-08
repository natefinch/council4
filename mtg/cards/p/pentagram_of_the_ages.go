package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PentagramOfTheAges is the card definition for Pentagram of the Ages.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{4}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage.
var PentagramOfTheAges = newPentagramOfTheAges

func newPentagramOfTheAges() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Pentagram of the Ages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{4}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
					ManaCost:        opt.Val(cost.Mana{cost.O(4)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player:  game.ControllerReference(),
									All:     true,
									OneShot: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{4}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage.
		`,
		},
	}
}
