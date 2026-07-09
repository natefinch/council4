package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WindingCanyons is the card definition for Winding Canyons.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{2}, {T}: You may cast creature spells this turn as though they had flash.
var WindingCanyons = newWindingCanyons

func newWindingCanyons() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Winding Canyons",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: You may cast creature spells this turn as though they had flash.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
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
											SpellTypes:     []types.Card{types.Creature},
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
			{2}, {T}: You may cast creature spells this turn as though they had flash.
		`,
		},
	}
}
