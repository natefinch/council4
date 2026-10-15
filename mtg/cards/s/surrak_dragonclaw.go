package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SurrakDragonclaw is the card definition for Surrak Dragonclaw.
//
// Type: Legendary Creature — Human Warrior
// Cost: {2}{G}{U}{R}
//
// Oracle text:
//
//	Flash
//	This spell can't be countered.
//	Creature spells you control can't be countered.
//	Other creatures you control have trample.
var SurrakDragonclaw = newSurrakDragonclaw

func newSurrakDragonclaw() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Surrak Dragonclaw",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
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
				game.StaticAbility{
					Text: "Other creatures you control have trample.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
					},
				},
			},
			OracleText: `
			Flash
			This spell can't be countered.
			Creature spells you control can't be countered.
			Other creatures you control have trample.
		`,
		},
	}
}
