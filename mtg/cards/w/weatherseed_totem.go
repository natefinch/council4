package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WeatherseedTotem is the card definition for Weatherseed Totem.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {G}.
//	{2}{G}{G}{G}: This artifact becomes a 5/3 green Treefolk artifact creature with trample until end of turn.
//	When this artifact is put into a graveyard from the battlefield, if it was a creature, return this card to its owner's hand.
var WeatherseedTotem = newWeatherseedTotem()

func newWeatherseedTotem() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Weatherseed Totem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{G}{G}{G}: This artifact becomes a 5/3 green Treefolk artifact creature with trample until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.G, cost.G, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:     game.LayerColor,
											SetColors: []color.Color{color.Green},
										},
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature, types.Artifact},
											AddSubtypes: []types.Sub{types.Treefolk},
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 5}),
											SetToughness: opt.Val(game.PT{Value: 3}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
							MatchToZone:   true,
							ToZone:        zone.Graveyard,
						},
						InterveningIf: "if it was a creature",
						InterveningCondition: opt.Val(game.Condition{
							Object:        opt.Val(game.EventPermanentReference()),
							ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {G}.
			{2}{G}{G}{G}: This artifact becomes a 5/3 green Treefolk artifact creature with trample until end of turn.
			When this artifact is put into a graveyard from the battlefield, if it was a creature, return this card to its owner's hand.
		`,
		},
	}
}
