package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AndraditeLeech is the card definition for Andradite Leech.
//
// Type: Creature — Leech
// Cost: {2}{B}
//
// Oracle text:
//
//	Black spells you cast cost {B} more to cast.
//	{B}: This creature gets +1/+1 until end of turn.
var AndraditeLeech = newAndraditeLeech()

func newAndraditeLeech() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Andradite Leech",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leech},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ColorsAny: []color.Color{color.Black}},
								ColoredIncrease: []mana.Color{mana.B},
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{B}: This creature gets +1/+1 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Black spells you cast cost {B} more to cast.
			{B}: This creature gets +1/+1 until end of turn.
		`,
		},
	}
}
