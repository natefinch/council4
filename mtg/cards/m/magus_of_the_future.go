package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MagusOfTheFuture is the card definition for Magus of the Future.
//
// Type: Creature — Human Wizard
// Cost: {2}{U}{U}{U}
//
// Oracle text:
//
//	Play with the top card of your library revealed.
//	You may play lands and cast spells from the top of your library.
var MagusOfTheFuture = newMagusOfTheFuture

func newMagusOfTheFuture() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Magus of the Future",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.PlayWithTopCardRevealedStaticBody,
				game.StaticAbility{
					Text: "You may play lands and cast spells from the top of your library.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectPlayLandsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							PermanentTypes: []types.Card{types.Land},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
					},
				},
			},
			OracleText: `
			Play with the top card of your library revealed.
			You may play lands and cast spells from the top of your library.
		`,
		},
	}
}
