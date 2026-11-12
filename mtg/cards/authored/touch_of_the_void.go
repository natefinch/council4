package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const touchOfTheVoidOracleID = "6b530534-5c02-4874-9256-501102ef8a5f"
const touchOfTheVoidOracleText = "Devoid (This card has no color.)\nTouch of the Void deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead."
const touchOfTheVoidDamageResult = game.ResultKey("touch-of-the-void-damage")

var TouchOfTheVoid = cardset.Entry{
	OracleID: touchOfTheVoidOracleID,
	Name:     "Touch of the Void",
	New:      newTouchOfTheVoid,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestTouchOfTheVoidScenario",
}

func newTouchOfTheVoid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Touch of the Void",
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
			Types:    []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle},
					}),
				}},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						PublishResult: touchOfTheVoidDamageResult,
					},
					{
						Primitive: game.CreateReplacement{
							Replacement: &game.ReplacementEffect{
								Duration:                     game.DurationPermanent,
								MatchEvent:                   game.EventZoneChanged,
								MatchFromZone:                true,
								FromZone:                     zone.Battlefield,
								MatchToZone:                  true,
								ToZone:                       zone.Graveyard,
								ReplaceToZone:                zone.Exile,
								AffectedObjectMustBeCreature: true,
							},
							Object:   game.TargetPermanentReference(0),
							Duration: game.DurationThisTurn,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       touchOfTheVoidDamageResult,
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: touchOfTheVoidOracleText,
		},
	}
}
