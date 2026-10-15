package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnclaimedTerritory is the card definition for Unclaimed Territory.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, choose a creature type.
//	{T}: Add {C}.
//	{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type.
var UnclaimedTerritory = newUnclaimedTerritory

func newUnclaimedTerritory() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Unclaimed Territory",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					Text: "{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type.",
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
												Condition:         game.ManaSpendCastChosenCreatureType,
												Restriction:       game.ManaSpendRestrictedToCondition,
												ChosenSubtypeFrom: game.EntryTypeChoiceKey,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this land enters, choose a creature type."),
			},
			OracleText: `
			As this land enters, choose a creature type.
			{T}: Add {C}.
			{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type.
		`,
		},
	}
}
