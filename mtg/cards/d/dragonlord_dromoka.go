package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonlordDromoka is the card definition for Dragonlord Dromoka.
//
// Type: Legendary Creature — Elder Dragon
// Cost: {4}{G}{W}
//
// Oracle text:
//
//	This spell can't be countered.
//	Flying, lifelink
//	Your opponents can't cast spells during your turn.
var DragonlordDromoka = newDragonlordDromoka

func newDragonlordDromoka() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Dragonlord Dromoka",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elder, types.Dragon},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Your opponents can't cast spells during your turn.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectCantCastSpells,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerOpponent,
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
			OracleText: `
			This spell can't be countered.
			Flying, lifelink
			Your opponents can't cast spells during your turn.
		`,
		},
	}
}
