package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const rebelSalvoOracleID = "a35b7fd3-fdef-4414-b720-978f7ce70bff"
const rebelSalvoOracleText = "Affinity for Equipment (This spell costs {1} less to cast for each Equipment you control.)\nRebel Salvo deals 5 damage to target creature or planeswalker. That permanent loses indestructible until end of turn."

var RebelSalvo = cardset.Entry{
	OracleID: rebelSalvoOracleID,
	Name:     "Rebel Salvo",
	New:      newRebelSalvo,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestRebelSalvoScenario",
}

func newRebelSalvo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Rebel Salvo",
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{{
				Text: "Affinity for Equipment (This spell costs {1} less to cast for each Equipment you control.)",
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCostModifier,
					AffectedSource: true,
					CostModifier: game.CostModifier{
						Kind:               game.CostModifierSpell,
						PerObjectReduction: 1,
						CountSelection: &game.Selection{
							RequiredTypes: []types.Card{types.Artifact},
							SubtypesAny:   []types.Sub{types.Equipment},
							Controller:    game.ControllerYou,
						},
					},
				}},
			}},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature or planeswalker",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
					}),
				}},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{{
								Layer:          game.LayerAbility,
								RemoveKeywords: []game.Keyword{game.Indestructible},
							}},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: rebelSalvoOracleText,
		},
	}
}
