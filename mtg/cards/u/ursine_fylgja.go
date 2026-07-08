package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrsineFylgja is the card definition for Ursine Fylgja.
//
// Type: Creature — Spirit Bear
// Cost: {4}{W}
//
// Oracle text:
//
//	This creature enters with four healing counters on it.
//	Remove a healing counter from this creature: Prevent the next 1 damage that would be dealt to this creature this turn.
//	{2}{W}: Put a healing counter on this creature.
var UrsineFylgja = newUrsineFylgja

func newUrsineFylgja() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ursine Fylgja",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Bear},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove a healing counter from this creature: Prevent the next 1 damage that would be dealt to this creature this turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a healing counter from this creature",
							Amount:      1,
							CounterKind: counter.Healing,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{2}{W}: Put a healing counter on this creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Healing,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with four healing counters on it.", game.CounterPlacement{Kind: counter.Healing, Amount: 4}),
			},
			OracleText: `
			This creature enters with four healing counters on it.
			Remove a healing counter from this creature: Prevent the next 1 damage that would be dealt to this creature this turn.
			{2}{W}: Put a healing counter on this creature.
		`,
		},
	}
}
