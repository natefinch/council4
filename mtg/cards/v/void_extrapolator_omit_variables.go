package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VoidExtrapolator is the card definition for Void Extrapolator // Omit Variables.
//
// Type: Creature — Aetherborn Warlock // Sorcery
// Cost: {1}{B} // {U/B}
// Face: Omit Variables — Sorcery ({U/B})
//
// Oracle text:
//
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
//	Threshold — This creature gets +1/+1 as long as there are seven or more cards in your graveyard.
var VoidExtrapolator = newVoidExtrapolator

func newVoidExtrapolator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Void Extrapolator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:         []color.Color{color.Black},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Aetherborn, types.Warlock},
			Power:          opt.Val(game.PT{Value: 2}),
			Toughness:      opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Threshold — This creature gets +1/+1 as long as there are seven or more cards in your graveyard.",
					Condition: opt.Val(game.Condition{
						Text: "as long as there are seven or more cards in your graveyard",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerGraveyardCardCount,
								Op:        compare.GreaterOrEqual,
								Value:     7,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
			Threshold — This creature gets +1/+1 as long as there are seven or more cards in your graveyard.
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Omit Variables",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors: []color.Color{color.Blue, color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(3),
							Player:      game.ControllerReference(),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mill three cards. (Put the top three cards of your library into your graveyard.)
		`,
		}),
	}
}
