package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BilboBirthdayCelebrant is the card definition for Bilbo, Birthday Celebrant.
//
// Type: Legendary Creature — Halfling Rogue
// Cost: {W}{B}{G}
//
// Oracle text:
//
//	If you would gain life, you gain that much life plus 1 instead.
//	{2}{W}{B}{G}, {T}, Exile Bilbo: Search your library for any number of creature cards, put them onto the battlefield, then shuffle. Activate only if you have 111 or more life.
var BilboBirthdayCelebrant = newBilboBirthdayCelebrant

func newBilboBirthdayCelebrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Bilbo, Birthday Celebrant",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Halfling, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{W}{B}{G}, {T}, Exile Bilbo: Search your library for any number of creature cards, put them onto the battlefield, then shuffle. Activate only if you have 111 or more life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W, cost.B, cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalExileSource,
							Text:   "Exile Bilbo",
							Amount: 1,
							Source: zone.Battlefield,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 111}},
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain that much life plus 1 instead.", 1, 1),
			},
			OracleText: `
			If you would gain life, you gain that much life plus 1 instead.
			{2}{W}{B}{G}, {T}, Exile Bilbo: Search your library for any number of creature cards, put them onto the battlefield, then shuffle. Activate only if you have 111 or more life.
		`,
		},
	}
}
