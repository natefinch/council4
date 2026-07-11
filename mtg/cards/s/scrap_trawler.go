package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScrapTrawler is the card definition for Scrap Trawler.
//
// Type: Artifact Creature — Construct
// Cost: {3}
//
// Oracle text:
//
//	Whenever this creature dies or another artifact you control is put into a graveyard from the battlefield, return to your hand target artifact card in your graveyard with lesser mana value.
var ScrapTrawler = newScrapTrawler

func newScrapTrawler() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scrap Trawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventZoneChanged,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							MatchFromZone:          true,
							FromZone:               zone.Battlefield,
							MatchToZone:            true,
							ToZone:                 zone.Graveyard,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact card in your graveyard with lesser mana value",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou, ManaValueLessThanEventPermanent: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature dies or another artifact you control is put into a graveyard from the battlefield, return to your hand target artifact card in your graveyard with lesser mana value.
		`,
		},
	}
}
