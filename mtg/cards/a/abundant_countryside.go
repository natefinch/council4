package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AbundantCountryside is the card definition for Abundant Countryside.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add one mana of any color. Spend this mana only to cast a creature spell.
//	{6}, {T}: Create a 1/1 colorless Shapeshifter creature token with changeling. (It's every creature type.)
var AbundantCountryside = newAbundantCountryside

func newAbundantCountryside() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Abundant Countryside",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{6}, {T}: Create a 1/1 colorless Shapeshifter creature token with changeling. (It's every creature type.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(6),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(abundantCountrysideToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					Text: "{T}: Add one mana of any color. Spend this mana only to cast a creature spell.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
											SpendRider: opt.Val(game.ManaSpendRider{
												Condition:   game.ManaSpendCastCreatureSpell,
												Restriction: game.ManaSpendRestrictedToCondition,
											}),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{T}: Add {C}.
			{T}: Add one mana of any color. Spend this mana only to cast a creature spell.
			{6}, {T}: Create a 1/1 colorless Shapeshifter creature token with changeling. (It's every creature type.)
		`,
		},
	}
}

var abundantCountrysideToken = newAbundantCountrysideToken()

func newAbundantCountrysideToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
		},
	}
}
