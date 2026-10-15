package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MarangRiverProwler is the card definition for Marang River Prowler.
//
// Type: Creature — Human Rogue
// Cost: {2}{U}
//
// Oracle text:
//
//	This creature can't block and can't be blocked.
//	You may cast this card from your graveyard as long as you control a black or green permanent.
var MarangRiverProwler = newMarangRiverProwler

func newMarangRiverProwler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Marang River Prowler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't block and can't be blocked.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlocked,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
				game.StaticAbility{
					Text: "You may cast this card from your graveyard as long as you control a black or green permanent.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a black or green permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Black, color.Green}},
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
			This creature can't block and can't be blocked.
			You may cast this card from your graveyard as long as you control a black or green permanent.
		`,
		},
	}
}
