package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SageOfTheBeyond is the card definition for Sage of the Beyond.
//
// Type: Creature — Spirit Giant
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Flying
//	Spells you cast from anywhere other than your hand cost {2} less to cast.
//	Foretell {4}{U} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
var SageOfTheBeyond = newSageOfTheBeyond

func newSageOfTheBeyond() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sage of the Beyond",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Giant},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								SourceZones:      []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command},
								GenericReduction: 2,
							},
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.ForetellKeyword{Cost: cost.Mana{cost.O(4), cost.U}},
					},
				},
			},
			OracleText: `
			Flying
			Spells you cast from anywhere other than your hand cost {2} less to cast.
			Foretell {4}{U} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
		`,
		},
	}
}
