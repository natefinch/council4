package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const ghostVacuumOracleID = "4ef583ab-e0a0-4e4b-bd65-af766038b1fd"

const ghostVacuumProducer game.LinkedProducerKey = ghostVacuumOracleID + "/activated/exile"

var (
	ghostVacuumExile  = game.ProgramReference{ID: ghostVacuumOracleID + "/activated/exile/0", Version: "1"}
	ghostVacuumReturn = game.ProgramReference{ID: ghostVacuumOracleID + "/activated/return/1", Version: "1"}
)

var GhostVacuum = cardset.Entry{
	OracleID: ghostVacuumOracleID,
	Name:     "Ghost Vacuum",
	New:      newGhostVacuum,
	Verification: "mtg/rules/ghost_vacuum_test.go: TestGhostVacuumProductionFilteredReturnAndEntryReplacements; " +
		"TestGhostVacuumProductionRejectsStaleGraveyardTarget; TestGhostVacuumProductionUsesSorceryTimingAndActivationCosts. " +
		"Sources: Scryfall DSK 248 printing 8ac39c01-127f-4471-bc74-11a90c48e306, pinned oracle-cards.json SHA256 " +
		"6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; rulings " +
		"https://api.scryfall.com/cards/8ac39c01-127f-4471-bc74-11a90c48e306/rulings retrieved 2026-10-17 (empty data). " +
		"CR effective 2026-09-25, verified local copy " +
		"/Users/natefinch/.copilot/session-state/2d21805f-c9bb-4910-a872-eb7b8d5d9459/files/high-review-rules.txt " +
		"SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca: §§115.1c-2, 602.2-2b/5d, " +
		"307.5 require an explicit graveyard target, paid activation costs, and sorcery timing; 400.7 and 608.2b " +
		"require a returned-to-graveyard target to be a new, illegal target; 607.1-2a link only cards successfully " +
		"exiled by this object's first ability; 122.1b and 613.1d/f, 613.4b require a flying counter, an added Spirit " +
		"subtype without losing existing types, and 1/1 setting.",
	Programs: []cardprogram.Program{
		{Reference: ghostVacuumExile, Resolve: resolveGhostVacuumExile},
		{Reference: ghostVacuumReturn, Resolve: resolveGhostVacuumReturn},
	},
}

func newGhostVacuum() *game.CardDef {
	exile := game.Mode{
		Text: "Exile target card from a graveyard.",
		Targets: []game.TargetSpec{{
			Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, MinTargets: 1, MaxTargets: 1,
		}},
		Program: ghostVacuumExile,
	}.Ability()
	exile.LinkedProducers = []game.LinkedProducerKey{ghostVacuumProducer}

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(),
		CardFace: game.CardFace{
			Name:       "Ghost Vacuum",
			ManaCost:   opt.Val(cost.Mana{cost.O(1)}),
			Types:      []types.Card{types.Artifact},
			OracleText: "{T}: Exile target card from a graveyard.\n{6}, {T}, Sacrifice this artifact: Put each creature card exiled with this artifact onto the battlefield under your control with a flying counter on it. Each of them is a 1/1 Spirit in addition to its other types. Activate only as a sorcery.",
			ActivatedAbilities: []game.ActivatedAbility{
				{
					Text:            "{T}: Exile target card from a graveyard.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content:         exile,
				},
				{
					Text:     "{6}, {T}, Sacrifice this artifact: Put each creature card exiled with this artifact onto the battlefield under your control with a flying counter on it. Each of them is a 1/1 Spirit in addition to its other types. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(6)}),
					AdditionalCosts: []cost.Additional{
						{Kind: cost.AdditionalTap},
						{Kind: cost.AdditionalSacrificeSource},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Text:    "Put each creature card exiled with this artifact onto the battlefield under your control with a flying counter on it. Each of them is a 1/1 Spirit in addition to its other types.",
						Program: ghostVacuumReturn,
					}.Ability(),
				},
			},
		},
	}
}

func resolveGhostVacuumExile(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MoveCard{
		Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard,
		Destination: zone.Exile, PublishLinkedProducer: ghostVacuumProducer,
	})
	return err
}

func resolveGhostVacuumReturn(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.PutOnBattlefield{
		Source:    game.OriginalSourceLinkedBattlefieldSource(ghostVacuumProducer),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Recipient: opt.Val(game.ControllerReference()),
		EntryCounters: []game.CounterPlacement{{
			Kind: counter.Flying, Amount: 1,
		}},
		ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerType, AddSubtypes: []types.Sub{types.Spirit}, Duration: game.DurationPermanent},
			{Layer: game.LayerPowerToughnessSet, SetPower: opt.Val(game.PT{Value: 1}),
				SetToughness: opt.Val(game.PT{Value: 1}), Duration: game.DurationPermanent},
		},
	})
	return err
}
