package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GarrukSHorde is the card definition for Garruk's Horde.
//
// Type: Creature — Beast
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Trample
//	Play with the top card of your library revealed.
//	You may cast creature spells from the top of your library. (Do this only any time you could cast that creature spell. You still pay the spell's costs.)
var GarrukSHorde = newGarrukSHorde

func newGarrukSHorde() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Garruk's Horde",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.PlayWithTopCardRevealedStaticBody,
				game.StaticAbility{
					Text: "You may cast creature spells from the top of your library. (Do this only any time you could cast that creature spell. You still pay the spell's costs.)",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Creature},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
					},
				},
			},
			OracleText: `
			Trample
			Play with the top card of your library revealed.
			You may cast creature spells from the top of your library. (Do this only any time you could cast that creature spell. You still pay the spell's costs.)
		`,
		},
	}
}
