package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sacrifice is the card definition for Sacrifice.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature.
//	Add an amount of {B} equal to the sacrificed creature's mana value.
var Sacrifice = newSacrifice

func newSacrifice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sacrifice",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectManaValue,
								Multiplier: 1,
								Object:     game.SacrificedCostReference(),
							}),
							ManaColor: mana.B,
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature.
			Add an amount of {B} equal to the sacrificed creature's mana value.
		`,
		},
	}
}
