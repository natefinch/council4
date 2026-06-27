package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FreyaliseSRadiance is the card definition for Freyalise's Radiance.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Cumulative upkeep {2} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
//	Snow permanents don't untap during their controllers' untap steps.
var FreyaliseSRadiance = newFreyaliseSRadiance()

func newFreyaliseSRadiance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Freyalise's Radiance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							AffectedSelection: game.Selection{Supertypes: []types.Super{types.Snow}},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.CumulativeUpkeepTriggeredAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Cumulative upkeep {2} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
			Snow permanents don't untap during their controllers' untap steps.
		`,
		},
	}
}
