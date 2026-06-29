package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SalvagingStation is the card definition for Salvaging Station.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	{T}: Return target noncreature artifact card with mana value 1 or less from your graveyard to the battlefield.
//	Whenever a creature dies, you may untap this artifact.
var SalvagingStation = newSalvagingStation()

func newSalvagingStation() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Salvaging Station",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Return target noncreature artifact card with mana value 1 or less from your graveyard to the battlefield.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target noncreature artifact card with mana value 1 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
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
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			{T}: Return target noncreature artifact card with mana value 1 or less from your graveyard to the battlefield.
			Whenever a creature dies, you may untap this artifact.
		`,
		},
	}
}
