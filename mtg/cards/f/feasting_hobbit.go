package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FeastingHobbit is the card definition for Feasting Hobbit.
//
// Type: Creature — Halfling Citizen
// Cost: {1}{G}
//
// Oracle text:
//
//	Devour Food 3 (As this creature enters, you may sacrifice any number of Foods. It enters with three times that many +1/+1 counters on it.)
//	Creatures with power less than this creature's power can't block it.
var FeastingHobbit = newFeastingHobbit

func newFeastingHobbit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Feasting Hobbit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Halfling, types.Citizen},
			Power:     opt.Val(game.PT{Value: 2}),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.DevourSubtypeReplacement("As this creature enters, you may sacrifice any number of Foods, then it enters with 3 +1/+1 counters on it for each Food sacrificed.", 3, types.Food),
			},
			OracleText: `
			Devour Food 3 (As this creature enters, you may sacrifice any number of Foods. It enters with three times that many +1/+1 counters on it.)
			Creatures with power less than this creature's power can't block it.
		`,
		},
	}
}
