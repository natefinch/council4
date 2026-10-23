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
	shatteringPulseOracleID   = "77953b7a-813e-45f5-ad7f-ab72a79a0112"
	whispersOfTheMuseOracleID = "a4a8af4e-8686-45ba-94fd-ccf6cc8adaae"
	disturbedBurialOracleID   = "cdbaa21d-5896-4b8b-9a08-f0839ccfddfc"
)

var (
	shatteringPulseProgram   = game.ProgramReference{ID: shatteringPulseOracleID + "/spell/0", Version: "1"}
	whispersOfTheMuseProgram = game.ProgramReference{ID: whispersOfTheMuseOracleID + "/spell/0", Version: "1"}
	disturbedBurialProgram   = game.ProgramReference{ID: disturbedBurialOracleID + "/spell/0", Version: "1"}
)

// ShatteringPulse destroys one target artifact and supports its printed buyback.
var ShatteringPulse = cardset.Entry{
	OracleID: shatteringPulseOracleID,
	Name:     "Shattering Pulse",
	New:      newShatteringPulse,
	Programs: []cardprogram.Program{{Reference: shatteringPulseProgram, Resolve: resolveDestroyTarget}},
	Verification: "mtg/rules/buyback_interaction_test.go: TestBuybackInteractionRegistryIdentityAndCosts; " +
		"TestBuybackInteractionCastsPaymentsAndInstantTiming; " +
		"TestBuybackValueShatteringPulseResolutionAndReplacements; " +
		"TestBuybackValueShatteringPulseCounteredAndIllegalTargets; " +
		"TestBuybackValueShatteringPulseDeferredManaAndClone",
}

// WhispersOfTheMuse draws for its controller and supports its printed buyback.
var WhispersOfTheMuse = cardset.Entry{
	OracleID: whispersOfTheMuseOracleID,
	Name:     "Whispers of the Muse",
	New:      newWhispersOfTheMuse,
	Programs: []cardprogram.Program{{Reference: whispersOfTheMuseProgram, Resolve: resolveWhispersOfTheMuse}},
	Verification: "mtg/rules/buyback_interaction_test.go: TestBuybackInteractionRegistryIdentityAndCosts; " +
		"TestBuybackInteractionCastsPaymentsAndInstantTiming; " +
		"TestBuybackValueWhispersDrawControllerAndClone; " +
		"TestBuybackValueWhispersReplacementAndEmptyLibrary",
}

// DisturbedBurial returns one creature card from its controller's graveyard.
var DisturbedBurial = cardset.Entry{
	OracleID: disturbedBurialOracleID,
	Name:     "Disturbed Burial",
	New:      newDisturbedBurial,
	Programs: []cardprogram.Program{{Reference: disturbedBurialProgram, Resolve: resolveNumericHandReturn}},
	Verification: "mtg/rules/buyback_interaction_test.go: TestBuybackInteractionRegistryIdentityAndCosts; " +
		"TestBuybackInteractionCastsPaymentsAndInstantTiming; " +
		"TestBuybackValueDisturbedBurialOwnCreatureTargetsAndTiming; " +
		"TestBuybackValueDisturbedBurialIncarnationsAndReplacement; " +
		"TestBuybackValueDisturbedBurialCopyRetargetsForController",
}

func newShatteringPulse() *game.CardDef {
	const effect = "Destroy target artifact."
	face := instantBuybackFace("Shattering Pulse", buybackReminderText+"\n"+effect,
		cost.Mana{cost.O(1), cost.R}, color.Red, 3)
	face.SpellAbility = opt.Val(game.Mode{
		Text:    effect,
		Program: shatteringPulseProgram,
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
		}},
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace:      face,
	}
}

func newWhispersOfTheMuse() *game.CardDef {
	const text = "Buyback {5} (You may pay an additional {5} as you cast this spell. If you do, put this card into your hand as it resolves.)\nDraw a card."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Whispers of the Muse",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{
				buybackStaticWithCost(5),
			},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Draw a card.",
				Program: whispersOfTheMuseProgram,
			}.Ability()),
		},
	}
}

func newDisturbedBurial() *game.CardDef {
	const text = "Return target creature card from your graveyard to your hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Disturbed Burial",
			OracleText: buybackReminderText + "\n" + text,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.B}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Black},
			StaticAbilities: []game.StaticAbility{
				buybackStatic(),
			},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Program: disturbedBurialProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
					}),
				}},
			}.Ability()),
		},
	}
}

func resolveWhispersOfTheMuse(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()})
	return err
}
