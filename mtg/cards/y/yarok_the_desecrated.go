package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YarokTheDesecrated is the card definition for Yarok, the Desecrated.
//
// Type: Legendary Creature — Elemental Horror
// Cost: {2}{B}{G}{U}
//
// Oracle text:
//
//	Deathtouch, lifelink
//	If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
var YarokTheDesecrated = newYarokTheDesecrated

func newYarokTheDesecrated() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Yarok, the Desecrated",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elemental, types.Horror},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectAdditionalTriggerForEnteringPermanent,
							Duration: game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			Deathtouch, lifelink
			If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
		`,
		},
	}
}
