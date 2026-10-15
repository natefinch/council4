package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheTigerGod is the card definition for The Tiger God.
//
// Type: Token Legendary Creature — Cat God
//
// Oracle text:
//   The Tiger God can't be blocked by more than one creature.

// TheTigerGodToken26ada256e5fc4c1d966feae96c4673b8 is the card definition for The Tiger God.
var TheTigerGodToken26ada256e5fc4c1d966feae96c4673b8 = newTheTigerGodToken26ada256e5fc4c1d966feae96c4673b8()

func newTheTigerGodToken26ada256e5fc4c1d966feae96c4673b8() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "The Tiger God",
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cat, types.God},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "The Tiger God can't be blocked by more than one creature.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByMoreThanOne,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			The Tiger God can't be blocked by more than one creature.
		`,
		},
	}
}
