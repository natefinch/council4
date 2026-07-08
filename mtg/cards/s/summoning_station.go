package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SummoningStation is the card definition for Summoning Station.
//
// Type: Artifact
// Cost: {7}
//
// Oracle text:
//
//	{T}: Create a 2/2 colorless Pincher creature token.
//	Whenever an artifact is put into a graveyard from the battlefield, you may untap this artifact.
var SummoningStation = newSummoningStation

func newSummoningStation() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Summoning Station",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Create a 2/2 colorless Pincher creature token.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(summoningStationToken),
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
							Event:            game.EventZoneChanged,
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Create a 2/2 colorless Pincher creature token.
			Whenever an artifact is put into a graveyard from the battlefield, you may untap this artifact.
		`,
		},
	}
}

var summoningStationToken = newSummoningStationToken()

func newSummoningStationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Pincher",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pincher},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
