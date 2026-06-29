package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KumenaTyrantOfOrazca is the card definition for Kumena, Tyrant of Orazca.
//
// Type: Legendary Creature — Merfolk Shaman
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Tap another untapped Merfolk you control: Kumena can't be blocked this turn.
//	Tap three untapped Merfolk you control: Draw a card.
//	Tap five untapped Merfolk you control: Put a +1/+1 counter on each Merfolk you control.
var KumenaTyrantOfOrazca = newKumenaTyrantOfOrazca()

func newKumenaTyrantOfOrazca() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Kumena, Tyrant of Orazca",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Merfolk, types.Shaman},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap another untapped Merfolk you control: Kumena can't be blocked this turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalTapPermanents,
							Text:          "Tap another untapped Merfolk you control",
							Amount:        1,
							ExcludeSource: true,
							SubtypesAny:   cost.SubtypeSet{types.Merfolk},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.SourcePermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text: "Tap three untapped Merfolk you control: Draw a card.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap three untapped Merfolk you control",
							Amount:      3,
							SubtypesAny: cost.SubtypeSet{types.Merfolk},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text: "Tap five untapped Merfolk you control: Put a +1/+1 counter on each Merfolk you control.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap five untapped Merfolk you control",
							Amount:      5,
							SubtypesAny: cost.SubtypeSet{types.Merfolk},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Merfolk")}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Tap another untapped Merfolk you control: Kumena can't be blocked this turn.
			Tap three untapped Merfolk you control: Draw a card.
			Tap five untapped Merfolk you control: Put a +1/+1 counter on each Merfolk you control.
		`,
		},
	}
}
