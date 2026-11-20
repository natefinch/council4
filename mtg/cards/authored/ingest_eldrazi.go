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
	ruinationGuideOracleID     = "3394ce74-c352-4ae1-a7bb-f6ef8ac1c810"
	benthicInfiltratorOracleID = "b2977807-5f46-424a-a6e9-14262f8a723a"
	cullingDroneOracleID       = "bda0ca68-78a6-4b32-842f-0bee9a9c84ef"
	sludgeCrawlerOracleID      = "dd05294b-2c5b-48c2-a9f6-8a528baa3dff"
	mistIntruderOracleID       = "e4e9b48b-92f9-46ac-b55a-573bc328b11a"

	devoidOracleText = "Devoid (This card has no color.)"
	ingestOracleText = "Ingest (Whenever this creature deals combat damage to a player, that player exiles the top card of their library.)"
	ingestEffectText = "That player exiles the top card of their library."
)

var (
	ruinationGuideIngest     = game.ProgramReference{ID: ruinationGuideOracleID + "/trigger/0", Version: "1"}
	benthicInfiltratorIngest = game.ProgramReference{ID: benthicInfiltratorOracleID + "/trigger/0", Version: "1"}
	cullingDroneIngest       = game.ProgramReference{ID: cullingDroneOracleID + "/trigger/0", Version: "1"}
	sludgeCrawlerIngest      = game.ProgramReference{ID: sludgeCrawlerOracleID + "/trigger/0", Version: "1"}
	mistIntruderIngest       = game.ProgramReference{ID: mistIntruderOracleID + "/trigger/0", Version: "1"}
)

var (
	RuinationGuide = cardset.Entry{
		OracleID: ruinationGuideOracleID,
		Name:     "Ruination Guide",
		New:      newRuinationGuide,
		Programs: []cardprogram.Program{{Reference: ruinationGuideIngest, Resolve: resolveEldraziIngest}},
		Verification: "mtg/rules/ingest_eldrazi_test.go: TestRuinationGuideIngestAndColorlessPump; " +
			"printing f0bdbe64-d985-4bc5-bb1d-4cbf12ef60b2; CR 2026-09-25 " +
			"401.2, 406.1, 406.3, 510.2, 603.1, 603.2c, 603.3a, 604.3, 611.3a, 613.1e, 613.4c, 702.114a",
	}
	BenthicInfiltrator = cardset.Entry{
		OracleID: benthicInfiltratorOracleID,
		Name:     "Benthic Infiltrator",
		New:      newBenthicInfiltrator,
		Programs: []cardprogram.Program{{Reference: benthicInfiltratorIngest, Resolve: resolveEldraziIngest}},
		Verification: "mtg/rules/ingest_eldrazi_test.go: TestBenthicInfiltratorCannotBeBlockedAndIngests; " +
			"printing f380ae58-08cc-42df-ac74-3c6890a62463; CR 2026-09-25 " +
			"113.12, 401.2, 406.1, 406.3, 509.1, 510.2, 603.1, 603.2c, 603.3a, 702.114a",
	}
	CullingDrone = cardset.Entry{
		OracleID: cullingDroneOracleID,
		Name:     "Culling Drone",
		New:      newCullingDrone,
		Programs: []cardprogram.Program{{Reference: cullingDroneIngest, Resolve: resolveEldraziIngest}},
		Verification: "mtg/rules/ingest_eldrazi_test.go: TestCullingDroneIngestsOnCombatDamage; " +
			"printing 1a623415-46c0-45aa-ab75-1e2d9708013a; CR 2026-09-25 " +
			"401.2, 406.1, 406.3, 510.2, 603.1, 603.2c, 603.3a, 604.3, 702.114a",
	}
	SludgeCrawler = cardset.Entry{
		OracleID: sludgeCrawlerOracleID,
		Name:     "Sludge Crawler",
		New:      newSludgeCrawler,
		Programs: []cardprogram.Program{{Reference: sludgeCrawlerIngest, Resolve: resolveEldraziIngest}},
		Verification: "mtg/rules/ingest_eldrazi_test.go: TestSludgeCrawlerPaysAndResolvesPumpBeforeIngest; " +
			"printing 9782b463-0677-40f4-ab2c-4c3ff19f1168; CR 2026-09-25 " +
			"401.2, 406.1, 406.3, 510.2, 601.2f, 602.2a-b, 603.1, 603.2c, 603.3a, 604.3, 611.2a, 613.4c, 702.114a",
	}
	MistIntruder = cardset.Entry{
		OracleID: mistIntruderOracleID,
		Name:     "Mist Intruder",
		New:      newMistIntruder,
		Programs: []cardprogram.Program{{Reference: mistIntruderIngest, Resolve: resolveEldraziIngest}},
		Verification: "mtg/rules/ingest_eldrazi_test.go: TestMistIntruderFlyingAndIngest; " +
			"printing f8c2e662-d8c3-44d4-b666-fe0d4a046271; CR 2026-09-25 " +
			"401.2, 406.1, 406.3, 509.1, 510.2, 603.1, 603.2c, 603.3a, 604.3, 702.9a-b, 702.114a",
	}
)

func newRuinationGuide() *game.CardDef {
	const buffText = "Other colorless creatures you control get +1/+0."
	return eldraziDrone("Ruination Guide", cost.Mana{cost.O(2), cost.U}, color.Blue, 3, 2,
		devoidOracleText+"\n"+ingestOracleText+"\n"+buffText,
		[]game.StaticAbility{
			game.DevoidStaticBody,
			{
				Text: buffText,
				ContinuousEffects: []game.ContinuousEffect{{
					Duration: game.DurationPermanent,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Colorless:     true,
						Controller:    game.ControllerYou,
						ExcludeSource: true,
					}),
					Layer:      game.LayerPowerToughnessModify,
					PowerDelta: 1,
				}},
			},
		}, ruinationGuideIngest)
}

func newBenthicInfiltrator() *game.CardDef {
	const evasionText = "This creature can't be blocked."
	return eldraziDrone("Benthic Infiltrator", cost.Mana{cost.O(2), cost.U}, color.Blue, 1, 4,
		devoidOracleText+"\n"+ingestOracleText+"\n"+evasionText,
		[]game.StaticAbility{
			game.DevoidStaticBody,
			{
				Text: evasionText,
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCantBeBlocked,
					AffectedSource: true,
				}},
			},
		}, benthicInfiltratorIngest)
}

func newCullingDrone() *game.CardDef {
	return eldraziDrone("Culling Drone", cost.Mana{cost.O(1), cost.B}, color.Black, 2, 2,
		devoidOracleText+"\n"+ingestOracleText,
		[]game.StaticAbility{game.DevoidStaticBody}, cullingDroneIngest)
}

func newSludgeCrawler() *game.CardDef {
	const pumpText = "{2}: This creature gets +1/+1 until end of turn."
	def := eldraziDrone("Sludge Crawler", cost.Mana{cost.B}, color.Black, 1, 1,
		devoidOracleText+"\n"+ingestOracleText+"\n"+pumpText,
		[]game.StaticAbility{game.DevoidStaticBody}, sludgeCrawlerIngest)
	def.ActivatedAbilities = []game.ActivatedAbility{{
		Text:           pumpText,
		ZoneOfFunction: zone.Battlefield,
		ManaCost:       opt.Val(cost.Mana{cost.O(2)}),
		Content: game.Mode{
			Text: pumpText,
			Sequence: []game.Instruction{{Primitive: game.ModifyPT{
				Object:         game.SourcePermanentReference(),
				PowerDelta:     game.Fixed(1),
				ToughnessDelta: game.Fixed(1),
				Duration:       game.DurationUntilEndOfTurn,
			}}},
		}.Ability(),
	}}
	return def
}

func newMistIntruder() *game.CardDef {
	const flyingText = "Flying"
	return eldraziDrone("Mist Intruder", cost.Mana{cost.O(1), cost.U}, color.Blue, 1, 2,
		devoidOracleText+"\n"+flyingText+"\n"+ingestOracleText,
		[]game.StaticAbility{game.DevoidStaticBody, game.FlyingStaticBody}, mistIntruderIngest)
}

func eldraziDrone(name string, manaCost cost.Mana, identity color.Color, power, toughness int, oracleText string, staticAbilities []game.StaticAbility, ingest game.ProgramReference) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(identity),
		CardFace: game.CardFace{
			Name:            name,
			ManaCost:        opt.Val(manaCost),
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Eldrazi, types.Drone},
			Colors:          []color.Color{},
			Power:           opt.Val(game.PT{Value: power}),
			Toughness:       opt.Val(game.PT{Value: toughness}),
			OracleText:      oracleText,
			StaticAbilities: staticAbilities,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: ingestOracleText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:               game.EventDamageDealt,
						Source:              game.TriggerSourceSelf,
						Subject:             game.TriggerSubjectDamageSource,
						DamageRecipient:     game.DamageRecipientPlayer,
						RequireCombatDamage: true,
					},
				},
				Content: game.Mode{Text: ingestEffectText, Program: ingest}.Ability(),
			}},
		},
	}
}

func resolveEldraziIngest(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MoveTopOfLibrary{
		Player:      game.EventPlayerReference(),
		Amount:      game.Fixed(1),
		Destination: zone.Exile,
	})
	return err
}
