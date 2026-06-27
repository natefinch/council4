package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoulCollector is the card definition for Soul Collector.
//
// Type: Creature — Vampire
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Flying
//	Whenever a creature dealt damage by this creature this turn dies, return that card to the battlefield under your control.
//	Morph {B}{B}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var SoulCollector = newSoulCollector()

func newSoulCollector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Soul Collector",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{Cost: cost.Mana{cost.B, cost.B, cost.B}},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentDied,
							DyingDamagedBySource: true,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever a creature dealt damage by this creature this turn dies, return that card to the battlefield under your control.
			Morph {B}{B}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
