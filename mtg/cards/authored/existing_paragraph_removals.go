package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	forcedMarchOracleID       = "457dc141-bc72-42ad-b482-c78183462e34"
	incineratingBlastOracleID = "49f0da4a-e046-4220-b90e-e105a5f2ef00"
)

var (
	forcedMarchProgram       = game.ProgramReference{ID: forcedMarchOracleID + "/spell/0", Version: "1"}
	incineratingBlastProgram = game.ProgramReference{ID: incineratingBlastOracleID + "/spell/0", Version: "1"}

	ForcedMarch = cardset.Entry{
		OracleID: forcedMarchOracleID, Name: "Forced March", New: newForcedMarch,
		Programs: []cardprogram.Program{{Reference: forcedMarchProgram, Resolve: resolveForcedMarch}},
		Verification: "mtg/rules/existing_paragraph_removals_test.go: TestForcedMarchScenario; " +
			"pinned Oracle corpus SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; " +
			"no published Scryfall rulings; official CR 2026-09-25 §§ 107.3a, 202.3, 608.2c, 701.8",
	}

	IncineratingBlast = cardset.Entry{
		OracleID: incineratingBlastOracleID, Name: "Incinerating Blast", New: newIncineratingBlast,
		Programs: []cardprogram.Program{{Reference: incineratingBlastProgram, Resolve: resolveIncineratingBlast}},
		Verification: "mtg/rules/existing_paragraph_removals_test.go: TestIncineratingBlastScenario; " +
			"pinned Oracle corpus SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; " +
			"Scryfall ruling 2024-11-08; official CR 2026-09-25 §§ 601.2c, 608.2b-c, 120.4b-d, 701.9",
	}
)

func newForcedMarch() *game.CardDef {
	const text = "Destroy all creatures with mana value X or less."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Forced March",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.X, cost.B, cost.B, cost.B}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Black},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Program: forcedMarchProgram,
			}.Ability()),
		},
	}
}

func newIncineratingBlast() *game.CardDef {
	const text = "Incinerating Blast deals 6 damage to target creature.\nYou may discard a card. If you do, draw a card."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Incinerating Blast",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Red},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Program: incineratingBlastProgram,
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

func resolveForcedMarch(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		ManaValue: opt.Val(compare.Int{
			Op:    compare.LessOrEqual,
			Value: ctx.View().Source().X,
		}),
	})})
	return err
}

func resolveIncineratingBlast(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.Damage{
		Amount:    game.Fixed(6),
		Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
	}); err != nil {
		return err
	}

	selected, err := ctx.Ask(game.ChoiceRequest{
		Kind:             game.ChoiceMay,
		Player:           ctx.View().Source().Controller,
		Prompt:           "Discard a card to draw a card?",
		Options:          []game.ChoiceOption{{Index: 0, Label: "No"}, {Index: 1, Label: "Yes"}},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	})
	if err != nil || len(selected) == 0 || selected[0] != 1 {
		return err
	}

	discard, err := ctx.Execute(game.Discard{
		Player: game.ControllerReference(),
		Amount: game.Fixed(1),
	})
	if err != nil || !discard.Succeeded {
		return err
	}
	_, err = ctx.Execute(game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)})
	return err
}
