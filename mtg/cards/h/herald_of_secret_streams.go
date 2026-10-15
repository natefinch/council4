package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeraldOfSecretStreams is the card definition for Herald of Secret Streams.
//
// Type: Creature — Merfolk Warrior
// Cost: {3}{U}
//
// Oracle text:
//
//	Creatures you control with +1/+1 counters on them can't be blocked.
var HeraldOfSecretStreams = newHeraldOfSecretStreams

func newHeraldOfSecretStreams() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Herald of Secret Streams",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control with +1/+1 counters on them can't be blocked.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeBlocked,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
							AffectedSelection:  game.Selection{MatchCounter: true},
						},
					},
				},
			},
			OracleText: `
			Creatures you control with +1/+1 counters on them can't be blocked.
		`,
		},
	}
}
