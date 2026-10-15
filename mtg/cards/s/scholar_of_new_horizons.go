package s

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

// ScholarOfNewHorizons is the card definition for Scholar of New Horizons.
//
// Type: Creature — Human Scout
// Cost: {1}{W}
//
// Oracle text:
//
//	This creature enters with a +1/+1 counter on it.
//	{T}, Remove a counter from a permanent you control: Search your library for a Plains card and reveal it. If an opponent controls more lands than you, you may put that card onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand. Then shuffle.
var ScholarOfNewHorizons = newScholarOfNewHorizons

func newScholarOfNewHorizons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Scholar of New Horizons",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Scout},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove a counter from a permanent you control: Search your library for a Plains card and reveal it. If an opponent controls more lands than you, you may put that card onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand. Then shuffle.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:           cost.AdditionalRemoveCounterAmong,
							Text:           "Remove a counter from a permanent you control",
							Amount:         1,
							CounterKind:    counter.PlusOnePlusOne,
							AnyCounterKind: true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Search your library for a Plains card and reveal it. If an opponent controls more lands than you, you may put that card onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand. Then shuffle.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone: zone.Library,
												Filter:     game.Selection{SubtypesAny: []types.Sub{types.Plains}},
												Reveal:     true,
												RevealOnly: true,
											},
											Amount:        game.Fixed(1),
											PublishLinked: game.LinkedKey("conditional-destination-card"),
										},
									},
									game.Instruction{
										Primitive: game.ConditionalDestinationPlace{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "conditional-destination-card",
											},
											FromZone: zone.Library,
											Condition: opt.Val(game.EffectCondition{
												Condition: opt.Val(game.Condition{
													Text: "If an opponent controls more lands than you",
													ControlComparison: opt.Val(game.ControlCountComparison{
														Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
														Left:      game.ControlPlayerAnyOpponent,
														Op:        compare.GreaterThan,
													}),
													SourceCounterKind: counter.PlusOnePlusOne,
												}),
											}),
											EntryTapped: true,
											Else:        zone.Hand,
										},
									},
									game.Instruction{
										Primitive: game.ShuffleLibrary{
											Player: game.ControllerReference(),
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
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			This creature enters with a +1/+1 counter on it.
			{T}, Remove a counter from a permanent you control: Search your library for a Plains card and reveal it. If an opponent controls more lands than you, you may put that card onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand. Then shuffle.
		`,
		},
	}
}
