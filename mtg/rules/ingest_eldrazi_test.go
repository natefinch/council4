package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	ruinationGuideTestOracleID     = "3394ce74-c352-4ae1-a7bb-f6ef8ac1c810"
	benthicInfiltratorTestOracleID = "b2977807-5f46-424a-a6e9-14262f8a723a"
	cullingDroneTestOracleID       = "bda0ca68-78a6-4b32-842f-0bee9a9c84ef"
	sludgeCrawlerTestOracleID      = "dd05294b-2c5b-48c2-a9f6-8a528baa3dff"
	mistIntruderTestOracleID       = "e4e9b48b-92f9-46ac-b55a-573bc328b11a"
)

// Source review: BFZ #64, Oracle ID 3394ce74-c352-4ae1-a7bb-f6ef8ac1c810,
// printing revision f0bdbe64-d985-4bc5-bb1d-4cbf12ef60b2; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Oracle:
// “Devoid (This card has no color.); Ingest (Whenever this creature deals
// combat damage to a player, that player exiles the top card of their library.);
// Other colorless creatures you control get +1/+0.” Scryfall rulings URI
// https://api.scryfall.com/cards/f0bdbe64-d985-4bc5-bb1d-4cbf12ef60b2/rulings
// retrieved 2026-11-20 returned seven WotC rulings dated 2015-08-25, including
// that Ingest exiles face up and does nothing to an empty library. Read the
// official CR effective 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (verified local copy .cardwork/MagicCompRules_20260925.docx), §§ 401.2,
// 406.1, 406.3, 510.2, 603.1, 603.2c, 603.3a, 604.3, 611.3a, 613.1a, 613.1e, 613.4c,
// 702.114a. The event is combat damage, Ingest resolves for the damaged player,
// Devoid makes Guide colorless, and the static +1/+0 applies to other matching
// creatures while excluding Guide itself.
func TestRuinationGuideIngestAndColorlessPump(t *testing.T) {
	g, engine, guide, topID := ingestEldraziScenario(t, ruinationGuideTestOracleID, "Ruination Guide")
	colorless := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	green := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Green Combat Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	setIngestCombatAttack(g, guide, colorless, green)

	log := resolveIngestCombat(t, g, engine, topID)
	assertIngestCombatDamage(t, log, guide, 3)
	assertIngestCombatDamage(t, log, colorless, 3)
	assertIngestCombatDamage(t, log, green, 2)
	if g.Players[game.Player2].Life != 32 {
		t.Fatalf("defending player life = %d, want 32", g.Players[game.Player2].Life)
	}
}

// Source review: BFZ #55, Oracle ID b2977807-5f46-424a-a6e9-14262f8a723a,
// printing revision f380ae58-08cc-42df-ac74-3c6890a62463; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Oracle:
// “Devoid (This card has no color.); Ingest (Whenever this creature deals
// combat damage to a player, that player exiles the top card of their library.);
// This creature can't be blocked.” Scryfall rulings URI
// https://api.scryfall.com/cards/f380ae58-08cc-42df-ac74-3c6890a62463/rulings
// retrieved 2026-11-20 returned seven WotC rulings dated 2015-08-25, including
// that Ingest exiles face up and does nothing to an empty library. Read the
// official CR effective 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (verified local copy .cardwork/MagicCompRules_20260925.docx), §§ 113.12,
// 401.2, 406.1, 406.3, 509.1, 510.2, 603.1, 603.2c, 603.3a, 604.3, 702.114a.
// A vanilla 1/4 control demonstrates both blockers can otherwise block; Benthic's
// own rule effect then disallows all blocks before it deals combat damage.
func TestBenthicInfiltratorCannotBeBlockedAndIngests(t *testing.T) {
	g, engine, infiltrator, topID := ingestEldraziScenario(t, benthicInfiltratorTestOracleID, "Benthic Infiltrator")
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flying := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Flying Blocker", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}), StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	vanilla := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Vanilla 1/4", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 4}),
	}})
	baseline := ingestLegalBlockActions(g, vanilla)
	if !ingestHasLegalBlock(baseline, vanilla, ground) || !ingestHasLegalBlock(baseline, vanilla, flying) {
		t.Fatal("control 1/4 did not offer both otherwise-legal blockers")
	}

	legal := ingestLegalBlockActions(g, infiltrator)
	if ingestHasLegalBlock(legal, infiltrator, ground) || ingestHasLegalBlock(legal, infiltrator, flying) {
		t.Fatal("Benthic Infiltrator was offered a blocker despite its can't-be-blocked text")
	}
	noBlocks := ingestNoBlockAction(t, legal)
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, noBlocks)) {
		t.Fatal("Engine rejected the legal no-block declaration")
	}

	log := resolveIngestCombat(t, g, engine, topID)
	assertIngestCombatDamage(t, log, infiltrator, 1)
}

// Source review: BFZ #91, Oracle ID bda0ca68-78a6-4b32-842f-0bee9a9c84ef,
// printing revision 1a623415-46c0-45aa-ab75-1e2d9708013a; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Oracle:
// “Devoid (This card has no color.); Ingest (Whenever this creature deals
// combat damage to a player, that player exiles the top card of their library.)”
// Scryfall rulings URI
// https://api.scryfall.com/cards/1a623415-46c0-45aa-ab75-1e2d9708013a/rulings
// retrieved 2026-11-20 returned seven WotC rulings dated 2015-08-25, including
// that Ingest exiles face up and does nothing to an empty library. Read the
// official CR effective 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (verified local copy .cardwork/MagicCompRules_20260925.docx), §§ 401.2,
// 406.1, 406.3, 510.2, 603.1, 603.2c, 603.3a, 604.3, 702.114a. The source's
// combat-damage trigger resolves by exiling the damaged player's library top.
func TestCullingDroneIngestsOnCombatDamage(t *testing.T) {
	g, engine, drone, topID := ingestEldraziScenario(t, cullingDroneTestOracleID, "Culling Drone")
	setIngestCombatAttack(g, drone)

	log := resolveIngestCombat(t, g, engine, topID)
	assertIngestCombatDamage(t, log, drone, 2)
}

// Source review: BFZ #98, Oracle ID dd05294b-2c5b-48c2-a9f6-8a528baa3dff,
// printing revision 9782b463-0677-40f4-ab2c-4c3ff19f1168; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Oracle:
// “Devoid (This card has no color.); Ingest (Whenever this creature deals
// combat damage to a player, that player exiles the top card of their library.);
// {2}: This creature gets +1/+1 until end of turn.” Scryfall rulings URI
// https://api.scryfall.com/cards/9782b463-0677-40f4-ab2c-4c3ff19f1168/rulings
// retrieved 2026-11-20 returned seven WotC rulings dated 2015-08-25, including
// that Ingest exiles face up and does nothing to an empty library. Read the
// official CR effective 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (verified local copy .cardwork/MagicCompRules_20260925.docx), §§ 401.2,
// 406.1, 406.3, 510.2, 601.2f, 602.2a-b, 603.1, 603.2c, 603.3a, 604.3, 611.2a,
// 613.4c, 702.114a. The Engine charges {2}, resolves the self-pump through end
// of turn, then combat damage triggers Ingest for the defending player.
func TestSludgeCrawlerPaysAndResolvesPumpBeforeIngest(t *testing.T) {
	g, engine, crawler, topID := ingestEldraziScenario(t, sludgeCrawlerTestOracleID, "Sludge Crawler")
	setMainPhasePriority(g, game.Player1)
	activate := action.ActivateAbility(crawler.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), activate) {
		t.Fatal("Sludge Crawler's {2} pump was legal without mana")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if !containsAction(engine.legalActions(g, game.Player1), activate) {
		t.Fatal("Sludge Crawler's {2} pump was unavailable with two mana")
	}
	if !engine.applyAction(g, game.Player1, activate) {
		t.Fatal("Sludge Crawler's {2} pump could not be activated")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining after paying {2} = %d, want 0", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, crawler); got != 2 {
		t.Fatalf("Sludge Crawler power after pump = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, crawler); !ok || got != 2 {
		t.Fatalf("Sludge Crawler toughness after pump = %d (%t), want 2", got, ok)
	}

	setIngestCombatAttack(g, crawler)
	log := resolveIngestCombat(t, g, engine, topID)
	assertIngestCombatDamage(t, log, crawler, 2)
}

// Source review: BFZ #61, Oracle ID e4e9b48b-92f9-46ac-b55a-573bc328b11a,
// printing revision f8c2e662-d8c3-44d4-b666-fe0d4a046271; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Oracle:
// “Devoid (This card has no color.); Flying; Ingest (Whenever this creature
// deals combat damage to a player, that player exiles the top card of their
// library.)” Scryfall rulings URI
// https://api.scryfall.com/cards/f8c2e662-d8c3-44d4-b666-fe0d4a046271/rulings
// retrieved 2026-11-20 returned seven WotC rulings dated 2015-08-25, including
// that Ingest exiles face up and does nothing to an empty library. Read the
// official CR effective 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (verified local copy .cardwork/MagicCompRules_20260925.docx), §§ 401.2,
// 406.1, 406.3, 509.1, 510.2, 603.1, 603.2c, 603.3a, 604.3, 702.9a-b,
// 702.114a. A ground blocker is illegal, while a flying blocker could block;
// when unblocked, Mist Intruder deals combat damage and triggers Ingest.
func TestMistIntruderFlyingAndIngest(t *testing.T) {
	g, engine, intruder, topID := ingestEldraziScenario(t, mistIntruderTestOracleID, "Mist Intruder")
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flying := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Flying Blocker", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}), StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	vanilla := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Vanilla 1/2", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	baseline := ingestLegalBlockActions(g, vanilla)
	if !ingestHasLegalBlock(baseline, vanilla, ground) || !ingestHasLegalBlock(baseline, vanilla, flying) {
		t.Fatal("control 1/2 did not offer both otherwise-legal blockers")
	}

	legal := ingestLegalBlockActions(g, intruder)
	if ingestHasLegalBlock(legal, intruder, ground) {
		t.Fatal("Mist Intruder was offered a ground blocker despite flying")
	}
	if !ingestHasLegalBlock(legal, intruder, flying) {
		t.Fatal("Mist Intruder could not be blocked by a flying creature")
	}
	noBlocks := ingestNoBlockAction(t, legal)
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, noBlocks)) {
		t.Fatal("Engine rejected the legal no-block declaration")
	}

	log := resolveIngestCombat(t, g, engine, topID)
	assertIngestCombatDamage(t, log, intruder, 1)
}

func ingestEldraziScenario(t *testing.T, oracleID, name string) (*game.Game, *Engine, *game.Permanent, id.ID) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s default-registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s lookup returned %q", name, def.Name)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: name + " ingest top card"}})
	return g, NewEngineWithPrograms(nil, registry.Programs()), source, topID
}

func setIngestCombatAttack(g *game.Game, attackers ...*game.Permanent) {
	declarations := make([]game.AttackDeclaration, 0, len(attackers))
	for _, attacker := range attackers {
		declarations = append(declarations, game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		})
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{Attackers: declarations}
}

func resolveIngestCombat(t *testing.T, g *game.Game, engine *Engine, topID id.ID) TurnLog {
	t.Helper()
	var log TurnLog
	engine.resolveCombatDamage(g, &log)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat damage did not put an Ingest trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("the damaged player's top library card was not exiled")
	}
	if g.Players[game.Player2].Library.Contains(topID) {
		t.Fatal("the ingested card remained in the damaged player's library")
	}
	return log
}

func assertIngestCombatDamage(t *testing.T, log TurnLog, attacker *game.Permanent, want int) {
	t.Helper()
	for _, damage := range log.CombatDamage {
		if damage.Attacker == attacker.ObjectID {
			if damage.Damage != want {
				t.Fatalf("%v combat damage = %d, want %d", attacker.ObjectID, damage.Damage, want)
			}
			return
		}
	}
	t.Fatalf("combat damage log has no damage from %v", attacker.ObjectID)
}

func ingestLegalBlockActions(g *game.Game, attacker *game.Permanent) []action.Action {
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	return legalDeclareBlockersActions(g, game.Player2)
}

func ingestHasLegalBlock(actions []action.Action, attacker, blocker *game.Permanent) bool {
	for _, candidate := range actions {
		payload, ok := candidate.DeclareBlockersPayload()
		if !ok {
			continue
		}
		for _, declaration := range payload.Blockers {
			if declaration.Blocker == blocker.ObjectID && declaration.Blocking == attacker.ObjectID {
				return true
			}
		}
	}
	return false
}

func ingestNoBlockAction(t *testing.T, actions []action.Action) action.Action {
	t.Helper()
	for _, candidate := range actions {
		payload, ok := candidate.DeclareBlockersPayload()
		if ok && len(payload.Blockers) == 0 {
			return candidate
		}
	}
	t.Fatal("legal blocker choices did not include no blocks")
	return action.Action{}
}
