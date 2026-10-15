package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarbaraWright is the card definition for Barbara Wright.
//
// Type: Legendary Creature — Human Advisor
// Cost: {1}{W}
//
// Oracle text:
//
//	History Teacher — Sagas you control have read ahead. (As a Saga enters, choose a chapter and start with that many lore counters. Skipped chapters don't trigger.)
//	Doctor's companion (You can have two commanders if the other is the Doctor.)
var BarbaraWright = newBarbaraWright

func newBarbaraWright() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Barbara Wright",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Advisor},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "History Teacher — Sagas you control have read ahead. (As a Saga enters, choose a chapter and start with that many lore counters. Skipped chapters don't trigger.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Saga")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.ReadAhead},
						},
					},
				},
				game.CompanionStaticBody,
			},
			OracleText: `
			History Teacher — Sagas you control have read ahead. (As a Saga enters, choose a chapter and start with that many lore counters. Skipped chapters don't trigger.)
			Doctor's companion (You can have two commanders if the other is the Doctor.)
		`,
		},
	}
}
