package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AugurOfAutumn is the card definition for Augur of Autumn.
//
// Type: Creature — Human Druid
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	You may look at the top card of your library any time.
//	You may play lands from the top of your library.
//	Coven — As long as you control three or more creatures with different powers, you may cast creature spells from the top of your library.
var AugurOfAutumn = newAugurOfAutumn

func newAugurOfAutumn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Augur of Autumn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LookAtTopCardAnyTimeStaticBody,
				game.PlayLandsFromLibraryTopStaticBody,
				game.StaticAbility{
					Text: "Coven — As long as you control three or more creatures with different powers, you may cast creature spells from the top of your library.",
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerCreaturePowerDiversity,
								Op:        compare.GreaterOrEqual,
								Value:     3,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Creature},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
					},
				},
			},
			OracleText: `
			You may look at the top card of your library any time.
			You may play lands from the top of your library.
			Coven — As long as you control three or more creatures with different powers, you may cast creature spells from the top of your library.
		`,
		},
	}
}
