package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BairdStewardOfArgive is the card definition for Baird, Steward of Argive.
//
// Type: Legendary Creature — Human Soldier
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Vigilance
//	Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.
var BairdStewardOfArgive = newBairdStewardOfArgive

func newBairdStewardOfArgive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Baird, Steward of Argive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectAttackTaxPerCreature,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerYou,
							AttackTaxGeneric:               1,
							AttackTaxIncludesPlaneswalkers: true,
						},
					},
				},
			},
			OracleText: `
			Vigilance
			Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.
		`,
		},
	}
}
