package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var SuperCombo = cardset.Entry{
	OracleID: "68555673-2833-459a-b415-e2805f837778",
	Name:     "Super Combo",
	New:      newSuperCombo,
	Verification: "mtg/rules/super_combo_test.go: TestSuperComboReplicateScenario; " +
		"printing 52bed0d4-3bcf-4770-b9c9-2a1d0666d220; Scryfall WotC rulings 2026-01-27; " +
		"CR 2026-11-06 601.2b/c/f-i, 603.2-4, 608.2b/h, 702.56a-b, 707.10/.10c, 113.7a, 120.1",
}

func newSuperCombo() *game.CardDef {
	const replicateText = "Replicate {2} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)"
	const damageText = "Target creature you control deals damage equal to its power to target creature an opponent controls."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Super Combo",
			OracleText: replicateText + "\n" + damageText,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Green},
			StaticAbilities: []game.StaticAbility{{
				Text: replicateText,
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.O(2)}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: damageText,
				Targets: []game.TargetSpec{
					{
						MinTargets: 1, MaxTargets: 1, Constraint: "target creature you control",
						Allow: game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou,
						}),
					},
					{
						MinTargets: 1, MaxTargets: 1, Constraint: "target creature an opponent controls",
						Allow: game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent,
						}),
					},
				},
				Sequence: []game.Instruction{{
					Primitive: game.Damage{
						Amount: game.Dynamic(game.DynamicAmount{
							Kind: game.DynamicAmountObjectPower, Object: game.TargetPermanentReference(0),
						}),
						Recipient:    game.ObjectDamageRecipient(game.TargetPermanentReference(1)),
						DamageSource: opt.Val(game.TargetPermanentReference(0)),
					},
				}},
			}.Ability()),
		},
	}
}
