package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AmuletOfKroog is the card definition for Amulet of Kroog.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{2}, {T}: Prevent the next 1 damage that would be dealt to any target this turn.
var AmuletOfKroog = newAmuletOfKroog

func newAmuletOfKroog() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Amulet of Kroog",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: Prevent the next 1 damage that would be dealt to any target this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
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
			{2}, {T}: Prevent the next 1 damage that would be dealt to any target this turn.
		`,
		},
	}
}
