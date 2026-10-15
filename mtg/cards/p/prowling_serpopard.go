package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ProwlingSerpopard is the card definition for Prowling Serpopard.
//
// Type: Creature — Cat Snake
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Creature spells you control can't be countered.
var ProwlingSerpopard = newProwlingSerpopard

func newProwlingSerpopard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Prowling Serpopard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Snake},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.StaticAbility{
					Text: "Creature spells you control can't be countered.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeCountered,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							SpellTypes:         []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			This spell can't be countered.
			Creature spells you control can't be countered.
		`,
		},
	}
}
