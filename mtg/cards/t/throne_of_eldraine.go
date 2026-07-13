package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThroneOfEldraine is the card definition for Throne of Eldraine.
//
// Type: Legendary Artifact
// Cost: {5}
//
// Oracle text:
//
//	As Throne of Eldraine enters, choose a color.
//	{T}: Add four mana of the chosen color. Spend this mana only to cast monocolored spells of that color.
//	{3}, {T}: Draw two cards. Spend only mana of the chosen color to activate this ability.
var ThroneOfEldraine = newThroneOfEldraine

func newThroneOfEldraine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Throne of Eldraine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:                                 "{3}, {T}: Draw two cards. Spend only mana of the chosen color to activate this ability.",
					ManaCost:                             opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts:                      cost.Tap,
					ZoneOfFunction:                       zone.Battlefield,
					ManaCostRestrictedToEntryChosenColor: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:          game.Fixed(4),
									EntryChoiceFrom: game.ChoiceKey("oracle-entry-color"),
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastMonocoloredSpellOfChosenColor,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As Throne of Eldraine enters, choose a color."),
			},
			OracleText: `
			As Throne of Eldraine enters, choose a color.
			{T}: Add four mana of the chosen color. Spend this mana only to cast monocolored spells of that color.
			{3}, {T}: Draw two cards. Spend only mana of the chosen color to activate this ability.
		`,
		},
	}
}
