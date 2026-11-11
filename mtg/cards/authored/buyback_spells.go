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
	recruitTheWorthyOracleID = "e25e2a3f-4add-4744-b1f8-2e880998a51f"
	mindGamesOracleID        = "ea43372f-b0a7-4d1b-be3e-a4a738b06d59"
	hazeOfRageOracleID       = "f17d0fb8-c157-43b8-be26-f5ba4c6aed14"

	recruitTheWorthyOracleText = "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Create a 1/1 white Soldier creature token."
	mindGamesOracleText = "Buyback {2}{U} (You may pay an additional {2}{U} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Tap target artifact, creature, or land."
	hazeOfRageOracleText = "Buyback {2} (You may pay an additional {2} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Creatures you control get +1/+0 until end of turn.\n" +
		"Storm (When you cast this spell, copy it for each spell cast before it this turn.)"
)

var (
	recruitTheWorthyProgram = game.ProgramReference{ID: recruitTheWorthyOracleID + "/spell/0", Version: "1"}
	mindGamesProgram        = game.ProgramReference{ID: mindGamesOracleID + "/spell/0", Version: "1"}
	hazeOfRageProgram       = game.ProgramReference{ID: hazeOfRageOracleID + "/spell/0", Version: "1"}

	RecruitTheWorthy = cardset.Entry{
		OracleID: recruitTheWorthyOracleID,
		Name:     "Recruit the Worthy",
		New:      newRecruitTheWorthy,
		Programs: []cardprogram.Program{{Reference: recruitTheWorthyProgram, Resolve: resolveRecruitTheWorthy}},
		Verification: "mtg/rules/buyback_spells_test.go: " +
			"TestRecruitTheWorthyScenario",
	}
	MindGames = cardset.Entry{
		OracleID: mindGamesOracleID,
		Name:     "Mind Games",
		New:      newMindGames,
		Programs: []cardprogram.Program{{Reference: mindGamesProgram, Resolve: resolveMindGames}},
		Verification: "mtg/rules/buyback_spells_test.go: " +
			"TestMindGamesScenario",
	}
	HazeOfRage = cardset.Entry{
		OracleID: hazeOfRageOracleID,
		Name:     "Haze of Rage",
		New:      newHazeOfRage,
		Programs: []cardprogram.Program{{Reference: hazeOfRageProgram, Resolve: resolveHazeOfRage}},
		Verification: "mtg/rules/buyback_spells_test.go: " +
			"TestHazeOfRageScenario",
	}
)

func newRecruitTheWorthy() *game.CardDef {
	face := instantBuybackFace(
		"Recruit the Worthy",
		recruitTheWorthyOracleText,
		cost.Mana{cost.W},
		color.White,
		3,
	)
	face.SpellAbility = opt.Val(game.Mode{
		Text:    "Create a 1/1 white Soldier creature token.",
		Program: recruitTheWorthyProgram,
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace:      face,
	}
}

func newMindGames() *game.CardDef {
	face := game.CardFace{
		Name:       "Mind Games",
		OracleText: mindGamesOracleText,
		ManaCost:   opt.Val(cost.Mana{cost.U}),
		Types:      []types.Card{types.Instant},
		Colors:     []color.Color{color.Blue},
		StaticAbilities: []game.StaticAbility{{
			Text:           "Buyback {2}{U}",
			ZoneOfFunction: zone.Stack,
			KeywordAbilities: []game.KeywordAbility{
				game.BuybackKeyword{Cost: cost.Mana{cost.O(2), cost.U}},
			},
		}},
	}
	face.SpellAbility = opt.Val(game.Mode{
		Text:    "Tap target artifact, creature, or land.",
		Program: mindGamesProgram,
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "target artifact, creature, or land",
			Allow:      game.TargetAllowPermanent,
			Selection: opt.Val(game.Selection{
				RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land},
			}),
		}},
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace:      face,
	}
}

func newHazeOfRage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:            "Haze of Rage",
			OracleText:      hazeOfRageOracleText,
			ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.R}),
			Types:           []types.Card{types.Sorcery},
			Colors:          []color.Color{color.Red},
			StaticAbilities: []game.StaticAbility{buybackStaticWithCost(2), game.StormStaticBody},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Creatures you control get +1/+0 until end of turn.",
				Program: hazeOfRageProgram,
			}.Ability()),
		},
	}
}

func resolveRecruitTheWorthy(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenDef(removalCreatureToken(
			"Soldier Token", 1, []color.Color{color.White}, types.Soldier,
		)),
	})
	return err
}

func resolveMindGames(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Tap{Object: game.TargetPermanentReference(0)})
	return err
}

func resolveHazeOfRage(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}),
			PowerDelta: 1,
		}},
		Duration: game.DurationUntilEndOfTurn,
	})
	return err
}
