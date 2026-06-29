package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VeteranWarleader is the card definition for Veteran Warleader.
//
// Type: Creature — Human Soldier Ally
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Veteran Warleader's power and toughness are each equal to the number of creatures you control.
//	Tap another untapped Ally you control: This creature gains your choice of first strike, vigilance, or trample until end of turn.
var VeteranWarleader = newVeteranWarleader()

func newVeteranWarleader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Veteran Warleader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:           []color.Color{color.Green, color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Soldier, types.Ally},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap another untapped Ally you control: This creature gains your choice of first strike, vigilance, or trample until end of turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalTapPermanents,
							Text:          "Tap another untapped Ally you control",
							Amount:        1,
							ExcludeSource: true,
							SubtypesAny:   cost.SubtypeSet{types.Ally},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourceCardPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.FirstStrike,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourceCardPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Vigilance,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourceCardPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Trample,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			Veteran Warleader's power and toughness are each equal to the number of creatures you control.
			Tap another untapped Ally you control: This creature gains your choice of first strike, vigilance, or trample until end of turn.
		`,
		},
	}
}
