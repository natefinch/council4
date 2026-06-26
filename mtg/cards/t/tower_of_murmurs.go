package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TowerOfMurmurs is the card definition for Tower of Murmurs.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{8}, {T}: Target player mills eight cards.
var TowerOfMurmurs = newTowerOfMurmurs()

func newTowerOfMurmurs() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tower of Murmurs",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{8}, {T}: Target player mills eight cards.",
					ManaCost:        opt.Val(cost.Mana{cost.O(8)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(8),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{8}, {T}: Target player mills eight cards.
		`,
		},
	}
}
