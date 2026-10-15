package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Howlgeist is the card definition for Howlgeist.
//
// Type: Creature — Spirit Wolf
// Cost: {5}{G}
//
// Oracle text:
//
//	Creatures with power less than this creature's power can't block it.
//	Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)
var Howlgeist = newHowlgeist

func newHowlgeist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Howlgeist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Wolf},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures with power less than this creature's power can't block it.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCantBlock,
							Duration:          game.DurationPermanent,
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{PowerLessThanSource: true},
							BlockedSource:     true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.UndyingTriggeredBody,
			},
			OracleText: `
			Creatures with power less than this creature's power can't block it.
			Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)
		`,
		},
	}
}
