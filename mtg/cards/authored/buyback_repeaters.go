package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	searingTouchOracleID  = "1cd9a3ba-a115-4e59-983d-605e4435f3a8"
	seethingAngerOracleID = "2effcd1c-5833-4e0e-bb67-2d2ffb22a014"
	mindPeelOracleID      = "4f424d94-1544-4086-a57d-2add17332094"
	elvishFuryOracleID    = "557d2af0-04c5-461d-826b-7208c4535dc2"

	searingTouchOracleText = "Buyback {4} (You may pay an additional {4} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Searing Touch deals 1 damage to any target."
	seethingAngerOracleText = "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Target creature gets +3/+0 until end of turn."
	mindPeelOracleText = "Buyback {2}{B}{B} (You may pay an additional {2}{B}{B} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Target player discards a card."
	elvishFuryOracleText = "Buyback {4} (You may pay an additional {4} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Target creature gets +2/+2 until end of turn."
)

var (
	searingTouchProgram  = game.ProgramReference{ID: searingTouchOracleID + "/spell/0", Version: "1"}
	seethingAngerProgram = game.ProgramReference{ID: seethingAngerOracleID + "/spell/0", Version: "1"}
	mindPeelProgram      = game.ProgramReference{ID: mindPeelOracleID + "/spell/0", Version: "1"}
	elvishFuryProgram    = game.ProgramReference{ID: elvishFuryOracleID + "/spell/0", Version: "1"}

	SearingTouch = cardset.Entry{
		OracleID: searingTouchOracleID,
		Name:     "Searing Touch",
		New:      newSearingTouch,
		Programs: []cardprogram.Program{{Reference: searingTouchProgram, Resolve: resolveSearingTouch}},
		Verification: "mtg/rules/buyback_repeaters_test.go: " +
			"TestSearingTouchScenario",
	}
	SeethingAnger = cardset.Entry{
		OracleID: seethingAngerOracleID,
		Name:     "Seething Anger",
		New:      newSeethingAnger,
		Programs: []cardprogram.Program{{Reference: seethingAngerProgram, Resolve: resolveSeethingAnger}},
		Verification: "mtg/rules/buyback_repeaters_test.go: " +
			"TestSeethingAngerScenario",
	}
	MindPeel = cardset.Entry{
		OracleID: mindPeelOracleID,
		Name:     "Mind Peel",
		New:      newMindPeel,
		Programs: []cardprogram.Program{{Reference: mindPeelProgram, Resolve: resolveMindPeel}},
		Verification: "mtg/rules/buyback_repeaters_test.go: " +
			"TestMindPeelScenario",
	}
	ElvishFury = cardset.Entry{
		OracleID: elvishFuryOracleID,
		Name:     "Elvish Fury",
		New:      newElvishFury,
		Programs: []cardprogram.Program{{Reference: elvishFuryProgram, Resolve: resolveElvishFury}},
		Verification: "mtg/rules/buyback_repeaters_test.go: " +
			"TestElvishFuryScenario",
	}
)

func newSearingTouch() *game.CardDef {
	face := instantBuybackFace(
		"Searing Touch",
		searingTouchOracleText,
		cost.Mana{cost.R},
		color.Red,
		4,
	)
	face.SpellAbility = opt.Val(game.Mode{
		Text:    "Searing Touch deals 1 damage to any target.",
		Program: searingTouchProgram,
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "any target",
			Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
			Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{
				types.Creature, types.Planeswalker, types.Battle,
			}}),
		}},
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace:      face,
	}
}

func newSeethingAnger() *game.CardDef {
	const spellText = "Target creature gets +3/+0 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:            "Seething Anger",
			OracleText:      seethingAngerOracleText,
			ManaCost:        opt.Val(cost.Mana{cost.R}),
			Types:           []types.Card{types.Sorcery},
			Colors:          []color.Color{color.Red},
			StaticAbilities: []game.StaticAbility{buybackStaticWithCost(3)},
			SpellAbility: opt.Val(game.Mode{
				Text:    spellText,
				Program: seethingAngerProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
			}.Ability()),
		},
	}
}

func newMindPeel() *game.CardDef {
	const spellText = "Target player discards a card."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Mind Peel",
			OracleText: mindPeelOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.B}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Black},
			StaticAbilities: []game.StaticAbility{{
				Text:           "Buyback {2}{B}{B}",
				ZoneOfFunction: zone.Stack,
				KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{
					Cost: cost.Mana{cost.O(2), cost.B, cost.B},
				}},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    spellText,
				Program: mindPeelProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target player",
					Allow:      game.TargetAllowPlayer,
				}},
			}.Ability()),
		},
	}
}

func newElvishFury() *game.CardDef {
	face := instantBuybackFace(
		"Elvish Fury",
		elvishFuryOracleText,
		cost.Mana{cost.G},
		color.Green,
		4,
	)
	face.SpellAbility = opt.Val(game.Mode{
		Text:    "Target creature gets +2/+2 until end of turn.",
		Program: elvishFuryProgram,
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "target creature",
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		}},
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace:      face,
	}
}

func resolveSearingTouch(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Damage{
		Amount:    game.Fixed(1),
		Recipient: game.AnyTargetDamageRecipient(0),
	})
	return err
}

func resolveSeethingAnger(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Duration:       game.DurationPermanent,
			Layer:          game.LayerPowerToughnessModify,
			PowerDelta:     3,
			ToughnessDelta: 0,
		}},
		Duration: game.DurationUntilEndOfTurn,
	})
	return err
}

func resolveMindPeel(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Discard{
		Player: game.TargetPlayerReference(0),
		Amount: game.Fixed(1),
	})
	return err
}

func resolveElvishFury(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Duration:       game.DurationPermanent,
			Layer:          game.LayerPowerToughnessModify,
			PowerDelta:     2,
			ToughnessDelta: 2,
		}},
		Duration: game.DurationUntilEndOfTurn,
	})
	return err
}
