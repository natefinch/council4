package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LifeSLegacy is the card definition for Life's Legacy.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature.
//	Draw cards equal to the sacrificed creature's power.
var LifeSLegacy = newLifeSLegacy

func newLifeSLegacy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Life's Legacy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
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
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectPower,
								Multiplier: 1,
								Object:     game.SacrificedCostReference(),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature.
			Draw cards equal to the sacrificed creature's power.
		`,
		},
	}
}
