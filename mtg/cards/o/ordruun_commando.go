package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OrdruunCommando is the card definition for Ordruun Commando.
//
// Type: Creature — Minotaur Soldier
// Cost: {3}{R}
//
// Oracle text:
//
//	{W}: Prevent the next 1 damage that would be dealt to this creature this turn.
var OrdruunCommando = newOrdruunCommando

func newOrdruunCommando() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Ordruun Commando",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Minotaur, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{W}: Prevent the next 1 damage that would be dealt to this creature this turn.",
					ManaCost:       opt.Val(cost.Mana{cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}: Prevent the next 1 damage that would be dealt to this creature this turn.
		`,
		},
	}
}
