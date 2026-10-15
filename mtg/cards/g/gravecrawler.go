package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Gravecrawler is the card definition for Gravecrawler.
//
// Type: Creature — Zombie
// Cost: {B}
//
// Oracle text:
//
//	This creature can't block.
//	You may cast this card from your graveyard as long as you control a Zombie.
var Gravecrawler = newGravecrawler

func newGravecrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gravecrawler",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
				game.StaticAbility{
					Text: "You may cast this card from your graveyard as long as you control a Zombie.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a Zombie",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Zombie}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ZoneOfFunction: zone.Graveyard,
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastFromZone,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Graveyard,
						},
					},
				},
			},
			OracleText: `
			This creature can't block.
			You may cast this card from your graveyard as long as you control a Zombie.
		`,
		},
	}
}
