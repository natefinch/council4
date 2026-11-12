package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Saiba Cryptomancer 15f89b06-5025-4c9f-bf51-e25bc5c2dd6b, printing
// 3934d535-740d-471a-bbf1-c3b26b1cd596: {1}{U}, 0/1 Moonfolk Ninja,
// "Flash / Backup 1 (When this creature enters, put a +1/+1 counter on target
// creature. If that's another creature, it gains the following ability until
// end of turn.) / Hexproof." Its Scryfall rulings URI
// https://api.scryfall.com/cards/3934d535-740d-471a-bbf1-c3b26b1cd596/rulings
// returned five WotC rulings dated 2023-04-14: self-targets get counters but
// no abilities, only printed abilities below Backup are granted, and the
// granted set is fixed when the trigger is put on the stack. Verified official
// CR effective 2026-09-25, .cardwork/sources/backup-n/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// read 113.7a, 603.1-3, 611.2a, 613.1f, 702.8a, 702.11, and 702.165a-d.
const saibaCryptomancerScenarioOracleID = "15f89b06-5025-4c9f-bf51-e25bc5c2dd6b"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Chomping Kavu 20a362c3-d278-437f-881c-e7d9fe78fc2d, printing
// 31c249d9-37c0-451d-866b-e834c4c57214: {3}{G}, 3/3 Kavu, "Backup 1 (When
// this creature enters, put a +1/+1 counter on target creature. If that's
// another creature, it gains the following ability until end of turn.) / This
// creature can't be blocked by creatures with power 2 or less." Its
// Scryfall rulings URI
// https://api.scryfall.com/cards/31c249d9-37c0-451d-866b-e834c4c57214/rulings
// returned six WotC rulings dated 2023-04-14; one says that changing a
// creature's power after it legally blocks Chomping Kavu does not undo that
// block. Verified official CR effective 2026-09-25, local copy above; read
// 113.7a, 603.1-3, 611.2a, 613.1f, 702.165a-d, and 509.1.
const chompingKavuScenarioOracleID = "20a362c3-d278-437f-881c-e7d9fe78fc2d"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Consuming Aetherborn 4d472308-5bc7-4a3c-81b5-0a757f00d9e2, printing
// 7311ade8-eb75-40f8-b018-668762aa3b77: {3}{B}, 2/2 Aetherborn Vampire,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target
// creature. If that's another creature, it gains the following ability until
// end of turn.) / Lifelink." Its Scryfall rulings URI
// https://api.scryfall.com/cards/7311ade8-eb75-40f8-b018-668762aa3b77/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.15, and 702.165a-d.
const consumingAetherbornScenarioOracleID = "4d472308-5bc7-4a3c-81b5-0a757f00d9e2"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Gloomfang Mauler 5437f560-7f61-4326-be8e-15d74bf4e6ec, printing
// 025a5338-133f-486d-9f73-0896226685c0: {5}{B}{B}, 5/5 Nightmare,
// "Swampcycling {2} ({2}, Discard this card: Search your library for a Swamp
// card, reveal it, put it into your hand, then shuffle.) / Backup 2 (When this
// creature enters, put two +1/+1 counters on target creature. If that's
// another creature, it gains the following ability until end of turn.) /
// Menace." Its Scryfall rulings URI
// https://api.scryfall.com/cards/025a5338-133f-486d-9f73-0896226685c0/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.29e-f, 702.111, 701.23-24, and 702.165a-d.
const gloomfangMaulerScenarioOracleID = "5437f560-7f61-4326-be8e-15d74bf4e6ec"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Cragsmasher Yeti 57120486-0316-4591-abca-438ab8283d3f, printing
// ee7a4ccb-c0c9-491f-9475-e78895fa9d03: {4}{R}{R}, 4/2 Yeti,
// "Mountaincycling {2} ({2}, Discard this card: Search your library for a
// Mountain card, reveal it, put it into your hand, then shuffle.) / Backup 2
// (When this creature enters, put two +1/+1 counters on target creature. If
// that's another creature, it gains the following ability until end of turn.) /
// Trample." Its Scryfall rulings URI
// https://api.scryfall.com/cards/ee7a4ccb-c0c9-491f-9475-e78895fa9d03/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.19, 702.29e-f, 701.23-24, and 702.165a-d.
const cragsmasherYetiScenarioOracleID = "57120486-0316-4591-abca-438ab8283d3f"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Serpent-Blade Assailant 66a131ae-dbb2-4e7b-b6ce-7695f47c8f87, printing
// 6fabc49f-b665-4c56-b06e-03220f7918bb: {2}{G}, 2/1 Elf Warrior,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target
// creature. If that's another creature, it gains the following ability until
// end of turn.) / Deathtouch." Its Scryfall rulings URI
// https://api.scryfall.com/cards/6fabc49f-b665-4c56-b06e-03220f7918bb/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.2b, and 702.165a-d.
const serpentBladeAssailantScenarioOracleID = "66a131ae-dbb2-4e7b-b6ce-7695f47c8f87"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Sigiled Sentinel 79c90a68-1c31-4e16-be53-1304132da80f, printing
// 7f65b178-51d6-4dcb-a1cc-4b4dcb2237cd: {2}{W}, 2/2 Human Knight,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target
// creature. If that's another creature, it gains the following ability until
// end of turn.) / Vigilance." Its Scryfall rulings URI
// https://api.scryfall.com/cards/7f65b178-51d6-4dcb-a1cc-4b4dcb2237cd/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.20, and 702.165a-d.
const sigiledSentinelScenarioOracleID = "79c90a68-1c31-4e16-be53-1304132da80f"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Death-Greeter's Champion 81abb451-5ee0-4f90-bacc-a8ff69726a27, printing
// 7cb2b582-1c45-4bb2-8aef-59a71a5a9e94: {2}{R}, 2/1 Human Warrior,
// "Dash {3}{R} (You may cast this spell for its dash cost. If you do, it gains
// haste, and it's returned from the battlefield to its owner's hand at the
// beginning of the next end step.) / Backup 1 (When this creature enters, put
// a +1/+1 counter on target creature. If that's another creature, it gains the
// following ability until end of turn.) / Double strike." Its Scryfall rulings
// URI https://api.scryfall.com/cards/7cb2b582-1c45-4bb2-8aef-59a71a5a9e94/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.4, 702.109, and 702.165a-d.
const deathGreetersChampionScenarioOracleID = "81abb451-5ee0-4f90-bacc-a8ff69726a27"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Scorn-Blade Berserker 93107139-b743-4d62-b0ac-9997202e0562, printing
// 792e7386-4c2d-4fa9-b499-fa3681f2a50e: {B}, 0/1 Human Berserker,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
// If that's another creature, it gains the following ability until end of
// turn.) / {1}, Sacrifice this creature: Draw a card." Its Scryfall rulings URI
// https://api.scryfall.com/cards/792e7386-4c2d-4fa9-b499-fa3681f2a50e/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 602.2, 603.1-3, 611.2a,
// 701.21, 702.29, 702.165a-d, and 121.
const scornBladeBerserkerScenarioOracleID = "93107139-b743-4d62-b0ac-9997202e0562"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Voldaren Thrillseeker b3cc2ed9-1da3-4ad2-a9a7-de7665d800ff, printing
// 1dbcd466-fb99-4388-a118-4534b85544f0: {2}{R}, 1/1 Vampire Warrior,
// "Backup 2 (When this creature enters, put two +1/+1 counters on target
// creature. If that's another creature, it gains the following ability until
// end of turn.) / {1}, Sacrifice this creature: It deals damage equal to its
// power to any target." Its Scryfall rulings URI
// https://api.scryfall.com/cards/1dbcd466-fb99-4388-a118-4534b85544f0/rulings
// returned six WotC rulings dated 2023-04-14; one requires its power as last
// seen on the battlefield. Verified official CR effective 2026-09-25, local
// copy above; read 113.7a, 120.3, 603.1-3, 611.2a, 613.1f, 701.21,
// 702.165a-d.
const voldarenThrillseekerScenarioOracleID = "b3cc2ed9-1da3-4ad2-a9a7-de7665d800ff"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Enduring Bondwarden c17d4818-5825-4bff-8fc5-5fd740cc0c70, printing
// eb0680eb-ceca-44d9-9654-78ea2ccfce17: {W}, 0/1 Human Scout,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
// If that's another creature, it gains the following ability until end of
// turn.) / When this creature dies, put its counters on target creature you
// control." Its Scryfall rulings URI
// https://api.scryfall.com/cards/eb0680eb-ceca-44d9-9654-78ea2ccfce17/rulings
// returned nine WotC rulings dated 2023-04-14; these clarify that the same
// number of each kind is put on the target, rather than moving counters, and
// that -1/-1 counters are included. Verified official CR effective 2026-09-25,
// local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f, 700.4, 702.165a-d.
const enduringBondwardenScenarioOracleID = "c17d4818-5825-4bff-8fc5-5fd740cc0c70"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Hangar Scrounger d1b74fb9-8ef3-40be-bc79-45ffb4c597eb, printing
// c2551b01-4d88-47f3-b54d-96ec03e093f8: {2}{R}, 2/1 Dwarf Pilot,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
// If that's another creature, it gains the following ability until end of
// turn.) / Whenever this creature becomes tapped, you may discard a card. If
// you do, draw a card." Its Scryfall rulings URI
// https://api.scryfall.com/cards/c2551b01-4d88-47f3-b54d-96ec03e093f8/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 701.9, 701.26a, 702.165a-d, and 121.
const hangarScroungerScenarioOracleID = "d1b74fb9-8ef3-40be-bc79-45ffb4c597eb"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Redcap Heelslasher d2134b68-d73f-4090-aaec-b2ee5c9f9f04, printing
// 3c7ca435-aa84-4d35-80db-d1c74b878b12: {3}{R}, 2/3 Goblin Rogue,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
// If that's another creature, it gains the following ability until end of
// turn.) / First strike." Its Scryfall rulings URI
// https://api.scryfall.com/cards/3c7ca435-aa84-4d35-80db-d1c74b878b12/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.7, and 702.165a-d.
const redcapHeelslasherScenarioOracleID = "d2134b68-d73f-4090-aaec-b2ee5c9f9f04"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Golden-Scale Aeronaut d388080d-5ad4-4e27-b361-340f410423a7, printing
// 8b2f3c52-8d6a-411a-aa62-cbb08a144351: {4}{W}, 2/3 Dwarf Pilot,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
// If that's another creature, it gains the following ability until end of
// turn.) / Flying." Its Scryfall rulings URI
// https://api.scryfall.com/cards/8b2f3c52-8d6a-411a-aa62-cbb08a144351/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.9a-b, and 702.165a-d.
const goldenScaleAeronautScenarioOracleID = "d388080d-5ad4-4e27-b361-340f410423a7"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Fearless Skald f055a941-8b83-464e-a57f-8f188dc69aaa, printing
// 4e317d1d-3f20-4f0d-8b1e-31df351e8f83: {4}{R}, 3/2 Dwarf Berserker,
// "Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
// If that's another creature, it gains the following ability until end of
// turn.) / Double strike." Its Scryfall rulings URI
// https://api.scryfall.com/cards/4e317d1d-3f20-4f0d-8b1e-31df351e8f83/rulings
// returned five WotC rulings dated 2023-04-14, including the Backup
// self-target and fixed-printed-ability rulings. Verified official CR effective
// 2026-09-25, local copy above; read 113.7a, 603.1-3, 611.2a, 613.1f,
// 702.4, and 702.165a-d.
const fearlessSkaldScenarioOracleID = "f055a941-8b83-464e-a57f-8f188dc69aaa"

func backupFamilyCard(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("Oracle ID %s resolved to %q, want %q", oracleID, def.Name, name)
	}
	return def
}

func backupFamilyCreature(name string, power, toughness int, static ...game.StaticAbility) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: power}), Toughness: opt.Val(game.PT{Value: toughness}),
		StaticAbilities: static,
	}}
}

func addBackupFamilyMana(g *game.Game, player game.PlayerID) int {
	for _, c := range []mana.Color{mana.C, mana.W, mana.U, mana.B, mana.R, mana.G} {
		g.Players[player].ManaPool.Add(c, 10)
	}
	return g.Players[player].ManaPool.Total()
}

type backupFamilyChoiceAgent struct {
	target    game.Target
	costLabel string
	acceptMay bool
}

func (backupFamilyChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a backupFamilyChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoicePayment:
		for _, option := range request.Options {
			if option.Label == a.costLabel {
				return []int{option.Index}
			}
		}
	case game.ChoiceTarget:
		for _, option := range request.Options {
			for _, target := range option.Targets {
				if target == a.target {
					return []int{option.Index}
				}
			}
		}
	case game.ChoiceMay:
		if a.acceptMay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func castAndResolveBackupFamilyCard(t *testing.T, g *game.Game, engine *Engine, oracleID, name string, manaValue, backupCounters int, target *game.Permanent, flash bool) (*game.CardDef, *game.Permanent, [game.NumPlayers]PlayerAgent) {
	t.Helper()
	def := backupFamilyCard(t, oracleID, name)
	cardID := addCardToHand(g, game.Player1, def)
	manaBefore := addBackupFamilyMana(g, game.Player1)
	if flash {
		g.Turn.ActivePlayer = game.Player2
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player1
	} else {
		setMainPhasePriority(g, game.Player1)
	}
	agents := simPassPolicies()
	if target != nil {
		agents[game.Player1] = backupFamilyChoiceAgent{target: game.PermanentTarget(target.ObjectID)}
	}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatalf("%s could not be cast", name)
	}
	if paid := manaBefore - g.Players[game.Player1].ManaPool.Total(); paid != manaValue {
		t.Fatalf("%s paid %d mana, want printed cost value %d", name, paid, manaValue)
	}
	resolveStackWithTriggers(engine, g, agents)
	permanent := permanentForCard(g, cardID)
	if permanent == nil {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
	if target != nil {
		if got := target.Counters.Get(counter.PlusOnePlusOne); got != backupCounters {
			t.Fatalf("%s Backup put %d +1/+1 counters on its target, want %d", name, got, backupCounters)
		}
	}
	return def, permanent, agents
}

func TestSaibaCryptomancerFlashesInAndGrantsHexproof(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Unprotected target", 2, 2))
	_, cryptomancer, _ := castAndResolveBackupFamilyCard(t, g, engine,
		saibaCryptomancerScenarioOracleID, "Saiba Cryptomancer", 2, 1, target, true)
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Backup put %d +1/+1 counters on target, want 1", got)
	}
	if !hasKeyword(g, target, game.Hexproof) {
		t.Fatal("Saiba Cryptomancer's Backup did not grant hexproof until end of turn")
	}
	if !hasKeyword(g, cryptomancer, game.Flash) {
		t.Fatal("Saiba Cryptomancer lacks its printed flash ability")
	}

	bolt := cards.NewDefaultRegistry().Lookup("Lightning Bolt")
	if bolt == nil {
		t.Fatal("default registry has no Lightning Bolt for the hexproof interaction")
	}
	boltID := addCardToHand(g, game.Player2, bolt)
	g.Players[game.Player2].ManaPool.Add(mana.R, 1)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player2
	if engine.applyAction(g, game.Player2, action.CastSpell(
		boltID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("opponent cast Lightning Bolt targeting the creature with Backup-granted hexproof")
	}

	selfGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	selfEngine := NewEngine(nil)
	selfCardID := addCardToHand(selfGame, game.Player1, backupFamilyCard(t,
		saibaCryptomancerScenarioOracleID, "Saiba Cryptomancer"))
	addBackupFamilyMana(selfGame, game.Player1)
	setMainPhasePriority(selfGame, game.Player1)
	if !selfEngine.applyAction(selfGame, game.Player1, action.CastSpell(selfCardID, nil, 0, nil)) {
		t.Fatal("could not cast Saiba Cryptomancer for its self-target Backup ruling")
	}
	selfEngine.resolveTopOfStack(selfGame, &TurnLog{})
	self := permanentForCard(selfGame, selfCardID)
	if self == nil {
		t.Fatal("self-target Saiba Cryptomancer did not enter")
	}
	selfAgents := simPassPolicies()
	selfAgents[game.Player1] = backupFamilyChoiceAgent{target: game.PermanentTarget(self.ObjectID)}
	if !selfEngine.putTriggeredAbilitiesOnStackWithChoices(selfGame, selfAgents, &TurnLog{}) {
		t.Fatal("self-target Backup trigger was not put on the stack")
	}
	selfEngine.resolveTopOfStackWithChoices(selfGame, selfAgents, &TurnLog{})
	if got := self.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("self-target Backup put %d counters on Saiba, want 1", got)
	}
	if len(selfGame.ContinuousEffects) != 0 {
		t.Fatalf("self-target Backup created %d temporary effects, want none", len(selfGame.ContinuousEffects))
	}
}

func TestChompingKavuBackupRestrictsBlockersByPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Kavu's target", 3, 3))
	castAndResolveBackupFamilyCard(t, g, engine, chompingKavuScenarioOracleID, "Chomping Kavu", 4, 1, target, false)
	weak := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	strong := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: target.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: weak.ObjectID, Blocking: target.ObjectID,
	}}))) {
		t.Fatal("power-2 creature legally blocked Chomping Kavu's Backup-granted restriction")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: strong.ObjectID, Blocking: target.ObjectID,
	}}))) {
		t.Fatal("power-3 creature could not block the creature with the granted restriction")
	}
}

func TestConsumingAetherbornBackupLifelinkGainsLifeFromCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Lifelink target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		consumingAetherbornScenarioOracleID, "Consuming Aetherborn", 4, 1, target, false)
	if !hasKeyword(g, target, game.Lifelink) {
		t.Fatal("Backup did not grant lifelink")
	}
	before := g.Players[game.Player1].Life
	power := effectivePower(g, target)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: target.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != before+power {
		t.Fatalf("controller life = %d, want %d after the granted lifelink damage", got, before+power)
	}
}

func activateTypecycling(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, landType types.Sub) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	land := &game.CardDef{CardFace: game.CardFace{
		Name: "typecycling target", Types: []types.Card{types.Land}, Subtypes: []types.Sub{landType},
	}}
	landID := addCardToLibrary(g, game.Player1, land)
	addCardToLibrary(g, game.Player1, backupFamilyCreature("Nonmatching library card", 1, 1))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	manaBefore := g.Players[game.Player1].ManaPool.Total()
	activation := action.ActivateAbility(cardID, def.ActivatedAbilityIndex(0), nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), activation) || !engine.applyAction(g, game.Player1, activation) {
		t.Fatalf("%s typecycling activation from hand failed", def.Name)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) ||
		manaBefore-g.Players[game.Player1].ManaPool.Total() != 2 {
		t.Fatalf("%s did not pay {2} and discard itself for typecycling", def.Name)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(landID) || g.Players[game.Player1].Library.Contains(landID) {
		t.Fatalf("%s did not find and reveal the matching land subtype", def.Name)
	}
}

func TestGloomfangMaulerMenaceAndSwampcycling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Menace target", 2, 2))
	def, _, _ := castAndResolveBackupFamilyCard(t, g, engine, gloomfangMaulerScenarioOracleID, "Gloomfang Mauler", 7, 2, target, false)
	if !hasKeyword(g, target, game.Menace) {
		t.Fatal("Backup did not grant menace")
	}
	oneBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	secondBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	single := []game.BlockDeclaration{{Blocker: oneBlocker.ObjectID, Blocking: target.ObjectID}}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(single))) {
		t.Fatal("one creature legally blocked the creature with granted menace")
	}
	pair := append(single, game.BlockDeclaration{Blocker: secondBlocker.ObjectID, Blocking: target.ObjectID})
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(pair))) {
		t.Fatal("two creatures could not block the creature with granted menace")
	}
	activateTypecycling(t, g, engine, def, types.Swamp)
}

func TestCragsmasherYetiTramplesAndMountaincycles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Trample target", 2, 2))
	def, _, _ := castAndResolveBackupFamilyCard(t, g, engine, cragsmasherYetiScenarioOracleID, "Cragsmasher Yeti", 6, 2, target, false)
	if !hasKeyword(g, target, game.Trample) {
		t.Fatal("Backup did not grant trample")
	}
	blocker := addCombatPermanent(g, game.Player2, backupFamilyCreature("Small blocker", 1, 1))
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: target.ObjectID}},
	}
	lifeBefore := g.Players[game.Player2].Life
	trampleDamage := effectivePower(g, target) - 1
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != lifeBefore-trampleDamage {
		t.Fatalf("trample damage to defending player = %d, want %d", lifeBefore-got, trampleDamage)
	}
	activateTypecycling(t, g, engine, def, types.Mountain)
}

func TestSerpentBladeAssailantBackupDeathtouchMarksLethalDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Deathtouch target", 3, 3))
	castAndResolveBackupFamilyCard(t, g, engine,
		serpentBladeAssailantScenarioOracleID, "Serpent-Blade Assailant", 3, 1, target, false)
	blocker := addCombatPermanent(g, game.Player2, backupFamilyCreature("Large blocker", 5, 5))
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: target.ObjectID}},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	if !blocker.MarkedDeathtouchDamage || blocker.MarkedDamage == 0 {
		t.Fatal("the blocker did not receive damage marked as lethal by granted deathtouch")
	}
}

func TestSigiledSentinelBackupVigilanceLeavesAttackerUntapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Vigilance target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		sigiledSentinelScenarioOracleID, "Sigiled Sentinel", 3, 1, target, false)
	if !hasKeyword(g, target, game.Vigilance) {
		t.Fatal("Backup did not grant vigilance")
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Combat = &game.CombatState{}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}))) {
		t.Fatal("the creature with granted vigilance could not attack")
	}
	if target.Tapped {
		t.Fatal("attacking with granted vigilance tapped the creature")
	}
}

func TestDeathGreetersChampionDashesAndGrantsDoubleStrike(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Double-strike target", 2, 2))
	def := backupFamilyCard(t, deathGreetersChampionScenarioOracleID, "Death-Greeter's Champion")
	cardID := addCardToHand(g, game.Player1, def)
	manaBefore := addBackupFamilyMana(g, game.Player1)
	setMainPhasePriority(g, game.Player1)
	agents := simPassPolicies()
	agents[game.Player1] = backupFamilyChoiceAgent{
		target: game.PermanentTarget(target.ObjectID), costLabel: "Dash",
	}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("Death-Greeter's Champion could not be cast for its Dash cost")
	}
	if paid := manaBefore - g.Players[game.Player1].ManaPool.Total(); paid != 4 {
		t.Fatalf("Dash paid %d mana, want {3}{R} (4)", paid)
	}
	resolveStackWithTriggers(engine, g, agents)
	dasher := permanentForCard(g, cardID)
	if dasher == nil || !hasKeyword(g, dasher, game.Haste) {
		t.Fatal("Dash did not give Death-Greeter's Champion haste")
	}
	if !hasKeyword(g, target, game.DoubleStrike) || target.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatal("Backup did not add its counter and grant double strike to the other creature")
	}

	before := g.Players[game.Player2].Life
	power := effectivePower(g, target)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: target.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != before-2*power {
		t.Fatalf("double-strike combat damage reduced life to %d, want %d", got, before-2*power)
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the dashed creature did not trigger its next-end-step return")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if permanentForCard(g, cardID) != nil || !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("the dashed creature was not returned to its owner's hand at the end step")
	}
}

func TestScornBladeBerserkerBackupGrantsSacrificeDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Sacrifice-draw target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		scornBladeBerserkerScenarioOracleID, "Scorn-Blade Berserker", 1, 1, target, false)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw for the granted ability"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	activation := action.ActivateAbility(target.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), activation) || !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("Backup-granted {1}, sacrifice, draw ability could not be activated")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, stillPresent := permanentByObjectID(g, target.ObjectID); stillPresent {
		t.Fatal("the creature with the granted ability was not sacrificed")
	}
	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("the granted sacrifice ability did not draw a card")
	}
}

func TestVoldarenThrillseekerBackupAbilityUsesLastKnownPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Damage-source target", 1, 1))
	castAndResolveBackupFamilyCard(t, g, engine,
		voldarenThrillseekerScenarioOracleID, "Voldaren Thrillseeker", 3, 2, target, false)
	if got := effectivePower(g, target); got != 3 {
		t.Fatalf("Backup target power = %d, want 3 after two counters", got)
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	activation := action.ActivateAbility(target.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), activation) || !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("Backup-granted sacrifice-and-damage ability could not target an opponent")
	}
	if _, stillPresent := permanentByObjectID(g, target.ObjectID); stillPresent {
		t.Fatal("the damage source was not sacrificed as an activation cost")
	}
	before := g.Players[game.Player2].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != before-3 {
		t.Fatalf("damage from sacrificed target = %d, want 3 from its last-known power", before-got)
	}
}

func TestEnduringBondwardenBackupDeathTriggerCopiesEveryCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Death-trigger target", 2, 2))
	_, _, agents := castAndResolveBackupFamilyCard(t, g, engine,
		enduringBondwardenScenarioOracleID, "Enduring Bondwarden", 1, 1, target, false)
	target.Counters.Add(counter.Charge, 2)
	destination := addCombatPermanent(g, game.Player1, backupFamilyCreature("Counter recipient", 1, 1))
	addEffectSpellToStack(g, game.Player2, game.Destroy{
		Object: game.TargetPermanentReference(0),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, stillPresent := permanentByObjectID(g, target.ObjectID); stillPresent {
		t.Fatal("the setup destruction did not kill the creature with the granted death ability")
	}
	agents[game.Player1] = backupFamilyChoiceAgent{target: game.PermanentTarget(destination.ObjectID)}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the Backup-granted death ability did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := destination.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("death ability put %d +1/+1 counters on target, want 1", got)
	}
	if got := destination.Counters.Get(counter.Charge); got != 2 {
		t.Fatalf("death ability put %d charge counters on target, want 2", got)
	}
}

func TestHangarScroungerBackupTapTriggerDiscardsAndDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Tap-trigger target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		hangarScroungerScenarioOracleID, "Hangar Scrounger", 3, 1, target, false)
	discarded := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discard for the granted ability"}})
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw for the granted ability"}})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Combat = &game.CombatState{}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}))) {
		t.Fatal("the creature with the Backup-granted tap trigger could not attack")
	}
	agents := simPassPolicies()
	agents[game.Player1] = backupFamilyChoiceAgent{acceptMay: true}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the target's tap-triggered ability did not trigger on attack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(discarded) || !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("accepting the tap trigger did not discard a card and draw a card")
	}
}

func TestRedcapHeelslasherFirstStrikeKillsBlockerBeforeDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("First-strike target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		redcapHeelslasherScenarioOracleID, "Redcap Heelslasher", 4, 1, target, false)
	blocker := addCombatPermanent(g, game.Player2, backupFamilyCreature("First-strike blocker", 3, 3))
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepFirstStrikeDamage
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: target.ObjectID}},
	}
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	if !hasKeyword(g, target, game.FirstStrike) {
		t.Fatal("Backup did not grant first strike")
	}
	if changed, _ := engine.checkPermanentStateBasedActions(g, newPassBatchID(g)); !changed {
		t.Fatal("lethal first-strike damage did not cause a permanent state-based action")
	}
	if _, stillPresent := permanentByObjectID(g, blocker.ObjectID); stillPresent {
		t.Fatal("the first-strike attacker did not destroy its lethal blocker before normal damage")
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("first-strike target received %d damage before the blocker left combat, want 0", target.MarkedDamage)
	}
}

func TestGoldenScaleAeronautFlyingAllowsOnlyAerialBlocks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Flying target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		goldenScaleAeronautScenarioOracleID, "Golden-Scale Aeronaut", 5, 1, target, false)
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("Backup did not grant flying")
	}
	grounded := addCombatPermanent(g, game.Player2, backupFamilyCreature("Grounded blocker", 2, 2))
	flyer := addCombatPermanent(g, game.Player2, backupFamilyCreature("Flying blocker", 2, 2, game.FlyingStaticBody))
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: grounded.ObjectID, Blocking: target.ObjectID,
	}}))) {
		t.Fatal("a creature without flying or reach blocked the Backup-granted flyer")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: flyer.ObjectID, Blocking: target.ObjectID,
	}}))) {
		t.Fatal("a flying creature could not block the creature with granted flying")
	}
}

func TestFearlessSkaldBackupDoubleStrikeDealsDamageTwice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, backupFamilyCreature("Double-strike target", 2, 2))
	castAndResolveBackupFamilyCard(t, g, engine,
		fearlessSkaldScenarioOracleID, "Fearless Skald", 5, 1, target, false)
	if !hasKeyword(g, target, game.DoubleStrike) {
		t.Fatal("Backup did not grant double strike")
	}
	before := g.Players[game.Player2].Life
	power := effectivePower(g, target)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != before-2*power {
		t.Fatalf("granted double strike reduced life to %d, want %d", got, before-2*power)
	}
}
