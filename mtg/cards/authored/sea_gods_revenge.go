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

const seaGodsRevengeOracleID = "8ab71a49-ce81-44f7-b940-50e7aba64df9"

var seaGodsRevengeSpell = game.ProgramReference{ID: seaGodsRevengeOracleID + "/spell/0", Version: "1"}

var SeaGodsRevenge = cardset.Entry{
	OracleID: seaGodsRevengeOracleID,
	Name:     "Sea God's Revenge",
	New:      newSeaGodsRevenge,
	Programs: []cardprogram.Program{{Reference: seaGodsRevengeSpell, Resolve: resolveSeaGodsRevenge}},
	Verification: "mtg/rules/scry_interaction_test.go: " +
		"TestSeaGodsRevengeDefaultRegistryPaidCastTimingAndTargetCounts; " +
		"TestSeaGodsRevengeSimultaneousOwnerRelativeReturns; " +
		"TestSeaGodsRevengeRevalidatesEachTargetAndSkipsScryWhenAllIllegal; " +
		"TestSeaGodsRevengeZeroTargetsScry. " +
		"mtg/rules/commander_sea_gods_revenge_test.go: " +
		"TestSeaGodsRevengeCommanderChoicesBeforeSimultaneousReturnAndScry; " +
		"mtg/rules/commander_simultaneous_choice_test.go: " +
		"TestSimultaneousCommanderReplacementChoicesAPNAP; " +
		"TestSimultaneousMixedZoneReplacementChoicesAPNAP. " +
		"Sources: pinned oracle-cards.json SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; " +
		"THS 61 printing 5acdf4e4-5933-43dc-bf8b-25f89415db5b; complete six-ruling response at " +
		"https://api.scryfall.com/cards/5acdf4e4-5933-43dc-bf8b-25f89415db5b/rulings retrieved 2026-10-22 " +
		"and retained in the rules test source record. Official CR " +
		"https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt, effective 2026-09-25, " +
		"SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca: " +
		"101.4 requires APNAP choices before simultaneous actions; 108.3, 400.7, and 903.9b make each " +
		"creature return to its owner's hand unless that owner chooses the commander-to-command-zone " +
		"replacement (distinct from the post-exile state-based action in 903.9a). 115.1a/601.2c enforce " +
		"distinct chosen targets and their fixed count; 608.2b/c/f require current target revalidation, " +
		"ordered instructions, and simultaneous movement; 701.22a-d governs the following scry.",
}

func newSeaGodsRevenge() *game.CardDef {
	const text = "Return up to three target creatures your opponents control to their owners' hands. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)"
	targets := []game.TargetSpec{{
		MinTargets: 0,
		MaxTargets: 3,
		Constraint: "up to three target creatures your opponents control",
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerOpponent,
		}),
	}}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Sea God's Revenge",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.U}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Targets: targets,
				Program: seaGodsRevengeSpell,
			}.Ability()),
		},
	}
}

func resolveSeaGodsRevenge(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.MovePermanent{
		Object:      game.AllTargetPermanentsReference(0),
		Destination: zone.Hand,
	}); err != nil {
		return err
	}
	_, err := ctx.Execute(game.Scry{Player: game.ControllerReference(), Amount: game.Fixed(1)})
	return err
}
