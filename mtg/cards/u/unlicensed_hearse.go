package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UnlicensedHearse is the card definition for Unlicensed Hearse.
//
// Type: Artifact — Vehicle
// Cost: {2}
//
// Oracle text:
//
//	{T}: Exile up to two target cards from a single graveyard.
//	Unlicensed Hearse's power and toughness are each equal to the number of cards exiled with it.
//	Crew 2
var UnlicensedHearse = newUnlicensedHearse

func newUnlicensedHearse() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Unlicensed Hearse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:            []types.Card{types.Artifact},
			Subtypes:         []types.Sub{types.Vehicle},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueSourceLinkedExileCount, LinkedKey: game.LinkedKey("exiled-with-source"), LinkedObjectScoped: true}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueSourceLinkedExileCount, LinkedKey: game.LinkedKey("exiled-with-source"), LinkedObjectScoped: true}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Exile up to two target cards from a single graveyard.",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets:    2,
										Constraint:    "up to two target cards from a single graveyard",
										Allow:         game.TargetAllowCard,
										Selection:     opt.Val(game.Selection{}),
										TargetZone:    zone.Graveyard,
										SameGraveyard: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:                  zone.Graveyard,
											Destination:               zone.Exile,
											PublishLinked:             game.LinkedKey("exiled-with-source"),
											PublishLinkedObjectScoped: true,
										},
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:        game.CardReferenceTarget,
												TargetIndex: 1,
											},
											FromZone:                  zone.Graveyard,
											Destination:               zone.Exile,
											PublishLinked:             game.LinkedKey("exiled-with-source"),
											PublishLinkedObjectScoped: true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.CrewActivatedAbility(2),
			},
			OracleText: `
			{T}: Exile up to two target cards from a single graveyard.
			Unlicensed Hearse's power and toughness are each equal to the number of cards exiled with it.
			Crew 2
		`,
		},
	}
}
