package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FungalBloom is the card definition for Fungal Bloom.
//
// Type: Enchantment
// Cost: {G}{G}
//
// Oracle text:
//
//	{G}{G}: Put a spore counter on target Fungus.
var FungalBloom = newFungalBloom

func newFungalBloom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fungal Bloom",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{G}{G}: Put a spore counter on target Fungus.",
					ManaCost:       opt.Val(cost.Mana{cost.G, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Fungus",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Fungus")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.Spore,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{G}{G}: Put a spore counter on target Fungus.
		`,
		},
	}
}
