package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KataraTheFearless is the card definition for Katara, the Fearless.
//
// Type: Legendary Creature — Human Warrior Ally
// Cost: {G}{W}{U}
//
// Oracle text:
//
//	If a triggered ability of an Ally you control triggers, that ability triggers an additional time.
var KataraTheFearless = newKataraTheFearless

func newKataraTheFearless() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Katara, the Fearless",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior, types.Ally},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "If a triggered ability of an Ally you control triggers, that ability triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
							Duration:          game.DurationPermanent,
							AffectedSelection: game.Selection{SubtypesAny: []types.Sub{types.Ally}},
						},
					},
				},
			},
			OracleText: `
			If a triggered ability of an Ally you control triggers, that ability triggers an additional time.
		`,
		},
	}
}
