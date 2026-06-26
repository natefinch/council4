package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TapestryOfTheAges is the card definition for Tapestry of the Ages.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{2}, {T}: Draw a card. Activate only if you've cast a noncreature spell this turn.
var TapestryOfTheAges = newTapestryOfTheAges()

func newTapestryOfTheAges() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tapestry of the Ages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: Draw a card. Activate only if you've cast a noncreature spell this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						}, Window: game.EventHistoryCurrentTurn}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}, {T}: Draw a card. Activate only if you've cast a noncreature spell this turn.
		`,
		},
	}
}
