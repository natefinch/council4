package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SelvalaHeartOfTheWilds is the card definition for Selvala, Heart of the Wilds.
//
// Type: Legendary Creature — Elf Scout
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Whenever another creature enters, its controller may draw a card if its power is greater than each other creature's power.
//	{G}, {T}: Add X mana in any combination of colors, where X is the greatest power among creatures you control.
var SelvalaHeartOfTheWilds = newSelvalaHeartOfTheWilds

func newSelvalaHeartOfTheWilds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Selvala, Heart of the Wilds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Scout},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.G}),
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountGreatestPowerInGroup,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										EventPermanentPowerGreaterThanEachOtherCreature: true,
									}),
								}),
								Optional:      true,
								OptionalActor: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever another creature enters, its controller may draw a card if its power is greater than each other creature's power.
			{G}, {T}: Add X mana in any combination of colors, where X is the greatest power among creatures you control.
		`,
		},
	}
}
