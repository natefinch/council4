package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// The exact printing revisions and full Scryfall rulings endpoints are:
// Igneous Inspiration, 5781ad7b-dc1b-4cc1-9e72-6e714b9ba1de,
// https://api.scryfall.com/cards/5781ad7b-dc1b-4cc1-9e72-6e714b9ba1de/rulings;
// Enthusiastic Study, 543c64ff-2c51-4a63-a940-dc8645717c85,
// https://api.scryfall.com/cards/543c64ff-2c51-4a63-a940-dc8645717c85/rulings;
// Arcane Subtraction, 68cdb00c-ea86-4b72-8f62-570820edfa1b,
// https://api.scryfall.com/cards/68cdb00c-ea86-4b72-8f62-570820edfa1b/rulings;
// Guiding Voice, d6fb3163-12ca-4a7f-a0c7-b8ddfc9408a0,
// https://api.scryfall.com/cards/d6fb3163-12ca-4a7f-a0c7-b8ddfc9408a0/rulings;
// Cram Session, c59a249f-35ed-447a-845b-32ba5a53124e,
// https://api.scryfall.com/cards/c59a249f-35ed-447a-845b-32ba5a53124e/rulings;
// Study Break, ef7cd813-0781-4fd7-8748-2716e1eeb4b9,
// https://api.scryfall.com/cards/ef7cd813-0781-4fd7-8748-2716e1eeb4b9/rulings;
// Pop Quiz, d16892d8-9d10-45de-ab79-0e645c4b5588,
// https://api.scryfall.com/cards/d16892d8-9d10-45de-ab79-0e645c4b5588/rulings.
// Each complete response was retrieved 2026-11-12 and contained the same four
// WotC rulings dated 2021-04-16: outside-game cards remain in the game until
// its end or their owner's departure; Constructed tournament cards come from
// the sideboard, casual Constructed cards from the collection, and Limited
// cards from the sealed/draft pool; when instructed to Learn, a player may do
// nothing, and both discarding and taking a Lesson are optional. The raw
// responses are retained under .cardwork/source-review/learn-20261112-a/.
//
// Official Wizards Comprehensive Rules effective 2026-09-25, matching the
// corpus rules pin: https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (verified local copy .cardwork/source-review/learn-20261112-a/MagicCompRules-2026-09-25.docx;
// SHA-256 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960).
// Read 115.1-6, 117.1a, 118.1, 119.1-3, 120.1-6, 121.1-2, 122.1a, 601.2,
// 608.2b-d, 613.4c, 701.26a, 701.48a, 514.2, and 702.19a. These require
// correct target kinds and target revalidation, printed spell timing/costs and
// ordered effects, damage/life/draw and +1/+1-counter results, P/T and ability
// duration through cleanup, tap behavior, Learn's optional discard-then-draw
// or outside-game Lesson choice, and trample's combat-only effect. The
// Commander model has no sideboard, so outside-game Lesson retrieval is not
// represented; Learn's discard/draw branch uses the existing DiscardThenDraw
// operation.

const (
	igneousInspirationOracleID = "2e31af8f-5e94-4888-857d-911253eac2cd"
	enthusiasticStudyOracleID  = "4233334d-bb19-41e4-a203-2b7bb0c48f52"
	arcaneSubtractionOracleID  = "4b5806fa-430f-4b78-b874-63d01a8a9de5"
	guidingVoiceOracleID       = "5c7e7c9e-00d1-497e-bee7-278bbd82e234"
	cramSessionOracleID        = "6e5b8bb2-9dd8-4986-9d0a-659114be5a2a"
	studyBreakOracleID         = "7aae7c7b-0b31-4b48-a8d3-81c913f604cf"
	popQuizOracleID            = "dda9ac6c-1ab2-4f77-b499-f03d587163d8"
)

type learnScenarioFixtures struct {
	discard id.ID
	draws   []id.ID
	chooser *orderedHandChoiceAgent
}

func learnCardDefinition(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s registry lookup returned %q", name, def.Name)
	}
	return def
}

func seedLearnScenario(g *game.Game, drawNames ...string) learnScenarioFixtures {
	fixtures := learnScenarioFixtures{
		discard: addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Learn discard",
		}}),
		chooser: &orderedHandChoiceAgent{order: []string{"Learn discard"}},
	}
	for _, name := range drawNames {
		fixtures.draws = append(fixtures.draws, addCardToLibrary(g, game.Player1, &game.CardDef{
			CardFace: game.CardFace{Name: name},
		}))
	}
	return fixtures
}

func resolveLearnScenario(t *testing.T, g *game.Game, engine *Engine, spell id.ID, fixtures learnScenarioFixtures) {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = fixtures.chooser
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolving spell did not reach its owner's graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(fixtures.discard) {
		t.Fatal("Learn did not discard the chosen card")
	}
	for _, card := range fixtures.draws {
		if !g.Players[game.Player1].Hand.Contains(card) {
			t.Fatalf("Learn or the spell did not draw card %v", card)
		}
	}
	if len(fixtures.chooser.requests) != 1 {
		t.Fatalf("Learn choice requests = %d, want 1", len(fixtures.chooser.requests))
	}
	if got := fixtures.chooser.requests[0].MaxChoices; got != 1 {
		t.Fatalf("Learn maximum discard choices = %d, want 1", got)
	}
}

func TestIgneousInspirationScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	startingLife := g.Players[game.Player2].Life
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, igneousInspirationOracleID, "Igneous Inspiration"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell,
		[]game.Target{game.PlayerTarget(game.Player2)}, 0, nil,
	)) {
		t.Fatal("Igneous Inspiration did not cast targeting an opponent")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Igneous Inspiration did not pay {2}{R}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	if got := g.Players[game.Player2].Life; got != startingLife-3 {
		t.Fatalf("opponent life = %d, want %d after 3 damage", got, startingLife-3)
	}
}

func TestEnthusiasticStudyScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Enthusiastic Study target", 2, 2))
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, enthusiasticStudyOracleID, "Enthusiastic Study"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Enthusiastic Study did not cast targeting a creature")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Enthusiastic Study did not pay {2}{R}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	if got := effectivePower(g, target); got != 5 {
		t.Fatalf("target power = %d, want 5", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 3 {
		t.Fatalf("target toughness = %d, valid %t; want 3", got, ok)
	}
	if !hasKeyword(g, target, game.Trample) {
		t.Fatal("target did not gain trample")
	}
}

func TestArcaneSubtractionScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Arcane Subtraction target", 5, 5))
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, arcaneSubtractionOracleID, "Arcane Subtraction"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Arcane Subtraction did not cast targeting a creature")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Arcane Subtraction did not pay {1}{U}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	if got := effectivePower(g, target); got != 1 {
		t.Fatalf("target power = %d, want 1 after -4/-0", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 5 {
		t.Fatalf("target toughness = %d, valid %t; want 5", got, ok)
	}
}

func TestGuidingVoiceScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Guiding Voice target", 1, 1))
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, guidingVoiceOracleID, "Guiding Voice"))
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Guiding Voice did not cast targeting a creature")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Guiding Voice did not pay {W}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("target +1/+1 counters = %d, want 1", got)
	}
	if got := effectivePower(g, target); got != 2 {
		t.Fatalf("target power = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 2 {
		t.Fatalf("target toughness = %d, valid %t; want 2", got, ok)
	}
}

func TestCramSessionScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	startingLife := g.Players[game.Player1].Life
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, cramSessionOracleID, "Cram Session"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Cram Session did not cast for {1}{B/G}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Cram Session did not pay {1}{B/G}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	if got := g.Players[game.Player1].Life; got != startingLife+4 {
		t.Fatalf("life total = %d, want %d", got, startingLife+4)
	}
}

func TestStudyBreakScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targets := []*game.Permanent{
		addCombatPermanent(g, game.Player2, vanillaCreature("Study Break target one", 2, 2)),
		addCombatPermanent(g, game.Player2, vanillaCreature("Study Break target two", 2, 2)),
	}
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, studyBreakOracleID, "Study Break"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(targets[0].ObjectID),
		game.PermanentTarget(targets[1].ObjectID),
	}, 0, nil)) {
		t.Fatal("Study Break did not cast targeting two creatures")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Study Break did not pay {1}{W}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	for i, target := range targets {
		if !target.Tapped {
			t.Errorf("target %d was not tapped", i)
		}
	}
}

func TestStudyBreakCanBeCastWithoutTargets(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Untapped Study Break creature", 2, 2))
	fixtures := seedLearnScenario(g, "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, studyBreakOracleID, "Study Break"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Study Break was not castable with zero targets")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)
	if target.Tapped {
		t.Fatal("Study Break tapped a creature when cast with no targets")
	}
}

func TestPopQuizScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fixtures := seedLearnScenario(g, "Pop Quiz draw", "Learn draw")
	spell := addCardToHand(g, game.Player1, learnCardDefinition(t, popQuizOracleID, "Pop Quiz"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Pop Quiz did not cast for {2}{U}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Pop Quiz did not pay {2}{U}")
	}
	resolveLearnScenario(t, g, engine, spell, fixtures)

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d after Pop Quiz and Learn, want 2", got)
	}
}
