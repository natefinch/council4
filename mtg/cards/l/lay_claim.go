package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LayClaim is the card definition for Lay Claim.
//
// Type: Enchantment — Aura
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Enchant permanent
//	You control enchanted permanent.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var LayClaim = newLayClaim

func newLayClaim() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Lay Claim",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				}),
				game.StaticAbility{
					Text: "You control enchanted permanent.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:      game.DurationPermanent,
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Enchant permanent
			You control enchanted permanent.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
