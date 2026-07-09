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

// AlchemistSRefuge is the card definition for Alchemist's Refuge.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{G}{U}, {T}: You may cast spells this turn as though they had flash.
var AlchemistSRefuge = newAlchemistSRefuge

func newAlchemistSRefuge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:  "Alchemist's Refuge",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{G}{U}, {T}: You may cast spells this turn as though they had flash.",
					ManaCost:        opt.Val(cost.Mana{cost.G, cost.U}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:           game.RuleEffectCastSpellsAsThoughFlash,
											AffectedPlayer: game.PlayerYou,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{G}{U}, {T}: You may cast spells this turn as though they had flash.
		`,
		},
	}
}
