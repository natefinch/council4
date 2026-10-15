package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MysticSanctuary is the card definition for Mystic Sanctuary.
//
// Type: Land — Island
//
// Oracle text:
//
//	({T}: Add {U}.)
//	This land enters tapped unless you control three or more other Islands.
//	When this land enters untapped, you may put target instant or sorcery card from your graveyard on top of your library.
var MysticSanctuary = newMysticSanctuary

func newMysticSanctuary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Mystic Sanctuary",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters untapped, you may put target instant or sorcery card from your graveyard on top of your library.",
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
										Constraint: "target instant or sorcery card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou}),
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
				game.EntersTappedIfReplacement("This land enters tapped unless you control three or more other Islands.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}, ExcludeSource: true},
						MinCount:  3,
					}),
				}),
			},
			OracleText: `
			({T}: Add {U}.)
			This land enters tapped unless you control three or more other Islands.
			When this land enters untapped, you may put target instant or sorcery card from your graveyard on top of your library.
		`,
		},
	}
}
