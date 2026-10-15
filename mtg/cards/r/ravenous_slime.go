package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RavenousSlime is the card definition for Ravenous Slime.
//
// Type: Creature — Ooze
// Cost: {2}{G}
//
// Oracle text:
//
//	This creature can't be blocked by creatures with power 2 or less.
//	If a creature an opponent controls would die, instead exile it and put a number of +1/+1 counters equal to that creature's power on this creature.
var RavenousSlime = newRavenousSlime

func newRavenousSlime() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ravenous Slime",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ooze},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 2 or less.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("If a creature an opponent controls would die, instead exile it and put a number of +1/+1 counters equal to that creature's power on this creature.", game.TriggerControllerAny, game.TriggerControllerOpponent, true, types.Creature),
			},
			OracleText: `
			This creature can't be blocked by creatures with power 2 or less.
			If a creature an opponent controls would die, instead exile it and put a number of +1/+1 counters equal to that creature's power on this creature.
		`,
		},
	}
}
