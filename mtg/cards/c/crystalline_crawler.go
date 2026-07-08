package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CrystallineCrawler is the card definition for Crystalline Crawler.
//
// Type: Artifact Creature — Construct
// Cost: {4}
//
// Oracle text:
//
//	Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
//	Remove a +1/+1 counter from this creature: Add one mana of any color.
//	{T}: Put a +1/+1 counter on this creature.
var CrystallineCrawler = func() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Crystalline Crawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a +1/+1 counter from this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountColorsOfManaSpentToCast,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
			Remove a +1/+1 counter from this creature: Add one mana of any color.
			{T}: Put a +1/+1 counter on this creature.
		`,
		},
	}
}
