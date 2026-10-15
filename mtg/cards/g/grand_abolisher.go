package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrandAbolisher is the card definition for Grand Abolisher.
//
// Type: Creature — Human Cleric
// Cost: {W}{W}
//
// Oracle text:
//
//	During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.
var GrandAbolisher = newGrandAbolisher

func newGrandAbolisher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Grand Abolisher",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectCantCastSpells,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerOpponent,
							RestrictedDuringControllerTurn: true,
						},
						game.RuleEffect{
							Kind:                           game.RuleEffectCantActivateAbilities,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerOpponent,
							PermanentTypes:                 []types.Card{types.Artifact, types.Creature, types.Enchantment},
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
			OracleText: `
			During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.
		`,
		},
	}
}
