package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Oasis is the card definition for Oasis.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Prevent the next 1 damage that would be dealt to target creature this turn.
var Oasis = newOasis()

func newOasis() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Oasis",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Prevent the next 1 damage that would be dealt to target creature this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Prevent the next 1 damage that would be dealt to target creature this turn.
		`,
		},
	}
}
