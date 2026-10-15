package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CollectiveRestraint is the card definition for Collective Restraint.
//
// Type: Enchantment
// Cost: {3}{U}
//
// Oracle text:
//
//	Domain — Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control.
var CollectiveRestraint = newCollectiveRestraint

func newCollectiveRestraint() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Collective Restraint",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Domain — Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                  game.RuleEffectAttackTaxPerCreature,
							Duration:              game.DurationPermanent,
							AffectedPlayer:        game.PlayerYou,
							AttackTaxScaledAmount: game.AggregateControllerBasicLandTypeCount,
						},
					},
				},
			},
			OracleText: `
			Domain — Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control.
		`,
		},
	}
}
