package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WitchSCottage is the card definition for Witch's Cottage.
//
// Type: Land — Swamp
//
// Oracle text:
//
//	({T}: Add {B}.)
//	This land enters tapped unless you control three or more other Swamps.
//	When this land enters untapped, you may put target creature card from your graveyard on top of your library.
var WitchSCottage = newWitchSCottage

func newWitchSCottage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:     "Witch's Cottage",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters untapped, you may put target creature card from your graveyard on top of your library.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{Tapped: game.TriFalse},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Library,
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
				game.EntersTappedIfReplacement("This land enters tapped unless you control three or more other Swamps.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Swamp")}, ExcludeSource: true},
						MinCount:  3,
					}),
				}),
			},
			OracleText: `
			({T}: Add {B}.)
			This land enters tapped unless you control three or more other Swamps.
			When this land enters untapped, you may put target creature card from your graveyard on top of your library.
		`,
		},
	}
}
