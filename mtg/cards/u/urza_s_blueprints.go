package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrzaSBlueprints is the card definition for Urza's Blueprints.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	Echo {6} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	{T}: Draw a card.
var UrzaSBlueprints = newUrzaSBlueprints

func newUrzaSBlueprints() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Urza's Blueprints",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Draw a card.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(6)}),
			},
			OracleText: `
			Echo {6} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			{T}: Draw a card.
		`,
		},
	}
}
