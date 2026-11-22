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

// Source review: Greenbelt Guardian, Oracle 2d8aa053-289d-40d9-baa7-9bd1c5b8e957,
// printing revision 2f6b0000-5fd3-483b-bac5-f9b888d8755b (dft #164): {1}{G},
// 2/2 Creature — Elf Ranger. Oracle: "{G}: Target creature gains trample until
// end of turn. / Exhaust — {3}{G}: Put three +1/+1 counters on this creature.
// (Activate each exhaust ability only once.)" The pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and the
// live Scryfall card endpoint agree. Its rulings URL
// https://api.scryfall.com/cards/2f6b0000-5fd3-483b-bac5-f9b888d8755b/rulings
// returned three 2025-02-07 rulings: exhaust can be activated at normal ability
// timing; leaving and returning makes a new object whose exhaust ability is
// usable again; and triggers on activating exhaust resolve before it. Rules
// source read before implementation: Wizards Comprehensive Rules effective
// 2026-11-06, official
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx
// (verified local copy .cardwork/source-review/MagicCompRules-20261106.docx).
// Read 118.2, 122.1/122.6, 400.7, 602.2/602.5b, 702.19a, and 702.177a:
// pay {G} and target a creature for the temporary trample ability; pay {3}{G}
// for the once-per-object ability and put three +1/+1 counters on this source.

// Source review: Skystreak Engineer, Oracle 2f920bf2-6a49-47c2-be3c-9787c377673b,
// printing revision 5bc9c501-098d-4560-9826-329b05689e0f (dft #61): {1}{U},
// 1/3 Creature — Human Pilot. Oracle: "Flying / Exhaust — {4}{U}: Put two
// +1/+1 counters on this creature. (Activate each exhaust ability only once.)"
// The pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4
// and live Scryfall card endpoint agree. Its rulings URL
// https://api.scryfall.com/cards/5bc9c501-098d-4560-9826-329b05689e0f/rulings
// returned three 2025-02-07 rulings: exhaust can be activated at normal ability
// timing; leaving and returning makes a new object whose exhaust ability is
// usable again; and triggers on activating exhaust resolve before it. Rules
// source read before implementation: Wizards Comprehensive Rules effective
// 2026-11-06,
// official https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx
// (verified local copy .cardwork/source-review/MagicCompRules-20261106.docx).
// Read 118.2, 122.1/122.6, 400.7, 602.2/602.5b, 702.9a, and 702.177a:
// pay {4}{U}, retain flying, and resolve two +1/+1 counters once on this object.

// Source review: Prowcatcher Specialist, Oracle
// 594132ac-32a7-41d5-b7f0-3692bbed7f6f, printing revision
// affbc8db-4ff7-4a86-954a-910493e5a2ef (dft #142): {1}{R}, 2/1 Creature —
// Goblin Warrior. Oracle: "Haste / Exhaust — {3}{R}: Put two +1/+1 counters on
// this creature. (Activate each exhaust ability only once.)" The pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and live
// Scryfall card endpoint agree. Its rulings URL
// https://api.scryfall.com/cards/affbc8db-4ff7-4a86-954a-910493e5a2ef/rulings
// returned three 2025-02-07 rulings: exhaust can be activated at normal ability
// timing; leaving and returning makes a new object whose exhaust ability is
// usable again; and triggers on activating exhaust resolve before it. Rules
// source read before implementation: Wizards Comprehensive Rules effective
// 2026-11-06,
// official https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx
// (verified local copy .cardwork/source-review/MagicCompRules-20261106.docx).
// Read 118.2, 122.1/122.6, 400.7, 602.2/602.5b, 702.10a, and 702.177a:
// pay {3}{R}, retain haste, and resolve two +1/+1 counters once on this object.

// Source review: Hazard of the Dunes, Oracle 799664fc-279c-40c1-8e40-44a68c38801b,
// printing revision c3fbee7a-7e73-43cf-bc39-ccb1c63bf837 (dft #165): {3}{G},
// 4/4 Creature — Wurm. Oracle: "Reach, trample / Exhaust — {6}{G}: Put three
// +1/+1 counters on this creature. (Activate each exhaust ability only once.)"
// The pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4
// and live Scryfall card endpoint agree. Its rulings URL
// https://api.scryfall.com/cards/c3fbee7a-7e73-43cf-bc39-ccb1c63bf837/rulings
// returned three 2025-02-07 rulings: exhaust can be activated at normal ability
// timing; leaving and returning makes a new object whose exhaust ability is
// usable again; and triggers on activating exhaust resolve before it. Rules
// source read before implementation: Wizards Comprehensive Rules effective
// 2026-11-06,
// official https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx
// (verified local copy .cardwork/source-review/MagicCompRules-20261106.docx).
// Read 118.2, 122.1/122.6, 400.7, 602.2/602.5b, 702.17a, 702.19a, and
// 702.177a: pay {6}{G}, retain reach and trample, and resolve three +1/+1
// counters once on this object.

const (
	greenbeltGuardianOracleID     = "2d8aa053-289d-40d9-baa7-9bd1c5b8e957"
	skystreakEngineerOracleID     = "2f920bf2-6a49-47c2-be3c-9787c377673b"
	prowcatcherSpecialistOracleID = "594132ac-32a7-41d5-b7f0-3692bbed7f6f"
	hazardOfTheDunesOracleID      = "799664fc-279c-40c1-8e40-44a68c38801b"
)

type exhaustCounterFixture struct {
	engine *Engine
	game   *game.Game
	def    *game.CardDef
	source *game.Permanent
}

func newExhaustCounterFixture(t *testing.T, oracleID, name string) exhaustCounterFixture {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("default registry Oracle ID %s returned %q, want %q", oracleID, def.Name, name)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	return exhaustCounterFixture{
		engine: NewEngineWithPrograms(nil, registry.Programs()),
		game:   g,
		def:    def,
		source: source,
	}
}

func (f exhaustCounterFixture) addMana(generic int, color mana.Color) {
	for range generic {
		f.game.Players[game.Player1].ManaPool.Add(mana.C, 1)
	}
	f.game.Players[game.Player1].ManaPool.Add(color, 1)
}

func (f exhaustCounterFixture) activate(t *testing.T, index, generic int, color mana.Color, counters int) {
	t.Helper()
	f.addMana(generic, color)
	setMainPhasePriority(f.game, game.Player1)
	activation := action.ActivateAbility(f.source.ObjectID, f.def.ActivatedAbilityIndex(index), nil, 0)
	if !containsAction(f.engine.legalActions(f.game, game.Player1), activation) {
		t.Fatalf("%s exhaust ability was not offered with %d generic and one %s mana", f.def.Name, generic, color)
	}
	if !f.engine.applyAction(f.game, game.Player1, activation) {
		t.Fatalf("%s exhaust activation failed", f.def.Name)
	}
	if got := f.game.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("%s left %d mana after paying its exhaust cost", f.def.Name, got)
	}
	if got := f.source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("%s received counters before its ability resolved: %d", f.def.Name, got)
	}
	if top, ok := f.game.Stack.Peek(); !ok || top.Kind != game.StackActivatedAbility {
		t.Fatalf("%s exhaust activation did not put an activated ability on the stack: %+v", f.def.Name, top)
	}
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
	if got := f.source.Counters.Get(counter.PlusOnePlusOne); got != counters {
		t.Fatalf("%s received %d +1/+1 counters, want %d", f.def.Name, got, counters)
	}

	f.addMana(generic, color)
	setMainPhasePriority(f.game, game.Player1)
	if containsAction(f.engine.legalActions(f.game, game.Player1), activation) {
		t.Fatalf("%s exhaust ability was offered again after its activation", f.def.Name)
	}
}

func exhaustScenarioCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Exhaust scenario creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func TestGreenbeltGuardianScenario(t *testing.T) {
	f := newExhaustCounterFixture(t, greenbeltGuardianOracleID, "Greenbelt Guardian")
	target := addCombatPermanent(f.game, game.Player1, exhaustScenarioCreature())

	f.game.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(f.game, game.Player1)
	pump := action.ActivateAbility(f.source.ObjectID, f.def.ActivatedAbilityIndex(0),
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0)
	if !containsAction(f.engine.legalActions(f.game, game.Player1), pump) {
		t.Fatal("Greenbelt Guardian's {G} trample ability was not offered with a creature target")
	}
	if !f.engine.applyAction(f.game, game.Player1, pump) {
		t.Fatal("Greenbelt Guardian's {G} trample ability failed")
	}
	if got := f.game.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("Greenbelt Guardian left %d mana after paying {G}", got)
	}
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
	if !hasKeyword(f.game, target, game.Trample) {
		t.Fatal("Greenbelt Guardian did not give its target trample")
	}

	f.activate(t, 1, 3, mana.G, 3)
	f.engine.runEndingPhase(f.game, simPassPolicies())
	if hasKeyword(f.game, target, game.Trample) {
		t.Fatal("Greenbelt Guardian's granted trample lasted beyond end of turn")
	}
}

func TestSkystreakEngineerScenario(t *testing.T) {
	f := newExhaustCounterFixture(t, skystreakEngineerOracleID, "Skystreak Engineer")
	if !hasKeyword(f.game, f.source, game.Flying) {
		t.Fatal("Skystreak Engineer does not have flying")
	}
	f.activate(t, 0, 4, mana.U, 2)
}

func TestProwcatcherSpecialistScenario(t *testing.T) {
	f := newExhaustCounterFixture(t, prowcatcherSpecialistOracleID, "Prowcatcher Specialist")
	if !hasKeyword(f.game, f.source, game.Haste) {
		t.Fatal("Prowcatcher Specialist does not have haste")
	}
	f.activate(t, 0, 3, mana.R, 2)
}

func TestHazardOfTheDunesScenario(t *testing.T) {
	f := newExhaustCounterFixture(t, hazardOfTheDunesOracleID, "Hazard of the Dunes")
	if !hasKeyword(f.game, f.source, game.Reach) || !hasKeyword(f.game, f.source, game.Trample) {
		t.Fatal("Hazard of the Dunes does not have reach and trample")
	}
	f.activate(t, 0, 6, mana.G, 3)
}
