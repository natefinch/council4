package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KunorosHoundOfAthreos is the card definition for Kunoros, Hound of Athreos.
//
// Type: Legendary Creature — Dog
// Cost: {1}{W}{B}
//
// Oracle text:
//
//	Vigilance, menace, lifelink
//	Creature cards in graveyards can't enter the battlefield.
//	Players can't cast spells from graveyards.
var KunorosHoundOfAthreos = newKunorosHoundOfAthreos

func newKunorosHoundOfAthreos() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Kunoros, Hound of Athreos",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dog},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.MenaceStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Creature cards in graveyards can't enter the battlefield.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantEnterFromZones,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
							EnterFromZones: []zone.Type{zone.Graveyard},
						},
					},
				},
				game.StaticAbility{
					Text: "Players can't cast spells from graveyards.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCantCastFromZones,
							Duration:          game.DurationPermanent,
							CantCastFromZones: []zone.Type{zone.Graveyard},
						},
					},
				},
			},
			OracleText: `
			Vigilance, menace, lifelink
			Creature cards in graveyards can't enter the battlefield.
			Players can't cast spells from graveyards.
		`,
		},
	}
}
