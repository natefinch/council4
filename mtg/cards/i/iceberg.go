package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Iceberg is the card definition for Iceberg.
//
// Type: Enchantment
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	This enchantment enters with X ice counters on it.
//	{3}: Put an ice counter on this enchantment.
//	Remove an ice counter from this enchantment: Add {C}.
var Iceberg = newIceberg()

func newIceberg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Iceberg",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{3}: Put an ice counter on this enchantment.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Ice,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove an ice counter from this enchantment",
							Amount:      1,
							CounterKind: counter.Ice,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This enchantment enters with X ice counters on it.", game.CounterPlacement{Kind: counter.Ice, AmountFromX: true}),
			},
			OracleText: `
			This enchantment enters with X ice counters on it.
			{3}: Put an ice counter on this enchantment.
			Remove an ice counter from this enchantment: Add {C}.
		`,
		},
	}
}
