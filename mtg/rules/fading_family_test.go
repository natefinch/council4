package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Skyshroud Ridgeback Oracle ID 34a160bc-5d7e-483c-b24c-ab30b7d313b0,
// printing revision 410896ab-d3dc-478c-bfd1-c0cad5b1180a:
// "Fading 2 (This creature enters with two fade counters on it. At the beginning
// of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)"
// Its Scryfall rulings URI (https://api.scryfall.com/cards/410896ab-d3dc-478c-bfd1-c0cad5b1180a/rulings)
// returned an empty list (0 rulings) on 2026-11-20. Read the official WotC
// Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
// 109.5, 122.6, 503.1/1a, 603.2b/3a, 608.2c, 609.3, 614.1c, 701.21a, and
// 702.32a. They require two fade counters on entry, one removed at the
// controller's upkeep, and sacrifice on the next upkeep when none can be removed.
func TestSkyshroudRidgebackUpkeepRemovesFadeThenSacrifices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ridgeback := castFadingCard(t, g, engine, "34a160bc-5d7e-483c-b24c-ab30b7d313b0", 0, mana.G)
	assertFadeCounters(t, ridgeback, 2)

	if resolveFadingUpkeep(t, g, engine, game.Player2) {
		t.Fatal("Skyshroud Ridgeback triggered during an opponent's upkeep")
	}
	assertFadeCounters(t, ridgeback, 2)

	resolveFadingLifecycle(t, g, engine, ridgeback, 2)
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Cloudskate Oracle ID 648e4072-b991-46fd-b057-1e807091dcdf, printing
// revision e7f97e50-3aeb-4c79-81b1-505a2f32d8ac. Oracle text:
// Flying
// Fading 3 (This creature enters with three fade counters on it. At the
// beginning of your upkeep, remove a fade counter from it. If you can't,
// sacrifice it.)
// Its Scryfall rulings URI
// (https://api.scryfall.com/cards/e7f97e50-3aeb-4c79-81b1-505a2f32d8ac/rulings)
// returned an empty list (0 rulings) on 2026-11-20. Read the official WotC
// Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
// 109.5, 122.6, 503.1/1a, 603.2b/3a, 608.2c, 609.3, 614.1c, 701.21a,
// 702.9b, and 702.32a. These require three fade counters and their upkeep
// removal/sacrifice sequence, with flying preventing it from being blocked
// except by creatures with flying or reach.
func TestCloudskateFadingAndFlying(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cloudskate := castFadingCard(t, g, engine, "648e4072-b991-46fd-b057-1e807091dcdf", 1, mana.U)
	if !hasKeyword(g, cloudskate, game.Flying) {
		t.Fatal("Cloudskate does not have flying")
	}
	assertFadeCounters(t, cloudskate, 3)
	resolveFadingLifecycle(t, g, engine, cloudskate, 3)
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Skyshroud Behemoth Oracle ID 77bee041-6f96-44e3-a18c-b52e508b9bdf,
// printing revision 1c01d17e-45a2-4b6f-aaa5-2af9c8f26181. Oracle text:
// Fading 2 (This creature enters with two fade counters on it. At the beginning
// of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)
// This creature enters tapped.
// Its Scryfall rulings URI
// (https://api.scryfall.com/cards/1c01d17e-45a2-4b6f-aaa5-2af9c8f26181/rulings)
// returned an empty list (0 rulings) on 2026-11-20. Read the official WotC
// Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
// 109.5, 122.6, 503.1/1a, 603.2b/3a, 608.2c, 609.3, 614.1c, 701.21a, and
// 702.32a. These require two fade counters on entry, entry tapped, then one
// fade counter removed per controller's upkeep until it is sacrificed when
// removal is impossible.
func TestSkyshroudBehemothEntersTappedAndFades(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	behemoth := castFadingCard(t, g, engine, "77bee041-6f96-44e3-a18c-b52e508b9bdf", 5, mana.G, mana.G)
	if !behemoth.Tapped {
		t.Fatal("Skyshroud Behemoth did not enter tapped")
	}
	assertFadeCounters(t, behemoth, 2)
	resolveFadingLifecycle(t, g, engine, behemoth, 2)
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Rejuvenation Chamber Oracle ID c1ebe189-db71-47cd-a763-8cb8c305383a,
// printing revision b97f86ce-4758-4ae3-af29-b08a4d771652. Oracle text:
// Fading 2 (This artifact enters with two fade counters on it. At the beginning
// of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)
// {T}: You gain 2 life.
// Its Scryfall rulings URI
// (https://api.scryfall.com/cards/b97f86ce-4758-4ae3-af29-b08a4d771652/rulings)
// returned an empty list (0 rulings) on 2026-11-20. Read the official WotC
// Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
// 109.5, 119.3, 122.6, 503.1/1a, 602.2a-b, 603.2b/3a, 608.2c, 609.3,
// 614.1c, 701.21a, and 702.32a. They require the tap cost to be paid before
// gaining two life, and two fade counters to be removed over the first two
// upkeeps before sacrificing the artifact at the next upkeep.
func TestRejuvenationChamberActivatesForLifeAndFades(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chamber := castFadingCard(t, g, engine, "c1ebe189-db71-47cd-a763-8cb8c305383a", 3)
	assertFadeCounters(t, chamber, 2)

	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(chamber.ObjectID, 0, nil, 0)) {
		t.Fatal("Rejuvenation Chamber's tap ability did not activate")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if g.Players[game.Player1].Life != 42 || !chamber.Tapped {
		t.Fatalf("life = %d, tapped = %t; want 42 life and a tapped Chamber", g.Players[game.Player1].Life, chamber.Tapped)
	}
	resolveFadingLifecycle(t, g, engine, chamber, 2)
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Blastoderm Oracle ID d6ee8bd6-0207-46ee-9892-c1d1b23b934e, printing
// revision 1aeb6b3a-1e84-4181-9e7f-d5a88ebf08b1. Oracle text:
// Shroud (This creature can't be the target of spells or abilities.)
// Fading 3 (This creature enters with three fade counters on it. At the
// beginning of your upkeep, remove a fade counter from it. If you can't,
// sacrifice it.)
// Its Scryfall rulings URI
// (https://api.scryfall.com/cards/1aeb6b3a-1e84-4181-9e7f-d5a88ebf08b1/rulings)
// returned an empty list (0 rulings) on 2026-11-20. Read the official WotC
// Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
// 109.5, 122.6, 503.1/1a, 603.2b/3a, 608.2c, 609.3, 614.1c, 701.21a,
// 702.18a, and 702.32a. These require shroud to prevent spell and ability
// targeting, alongside the three-counter fading sequence ending in sacrifice.
func TestBlastodermShroudAndFading(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	blastoderm := castFadingCard(t, g, engine, "d6ee8bd6-0207-46ee-9892-c1d1b23b934e", 2, mana.G, mana.G)
	if !hasKeyword(g, blastoderm, game.Shroud) {
		t.Fatal("Blastoderm does not have shroud")
	}
	assertFadeCounters(t, blastoderm, 3)

	registry := cards.NewDefaultRegistry()
	bolt := registry.Lookup("Lightning Bolt")
	if bolt == nil {
		t.Fatal("production Lightning Bolt is missing")
	}
	ordinaryCreature := addCombatPermanent(g, game.Player2, creatureDef("Targetable creature"))
	firstBolt := addCardToHand(g, game.Player1, bolt)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		firstBolt, []game.Target{game.PermanentTarget(ordinaryCreature.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Lightning Bolt could not target an ordinary creature")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	secondBolt := addCardToHand(g, game.Player1, bolt)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpell(
		secondBolt, []game.Target{game.PermanentTarget(blastoderm.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Lightning Bolt illegally targeted Blastoderm through shroud")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("failed shroud-restricted cast spent mana; pool = %d, want 1", got)
	}
	resolveFadingLifecycle(t, g, engine, blastoderm, 3)
}

func castFadingCard(t *testing.T, g *game.Game, engine *Engine, oracleID string, generic int, colors ...mana.Color) *game.Permanent {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default-registry lookup for %s failed: %v", oracleID, err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	if generic > 0 {
		g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	}
	for _, c := range colors {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", def.Name)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("casting %s left %d mana, want exact payment", def.Name, got)
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	t.Fatalf("%s did not enter the battlefield", def.Name)
	return nil
}

func resolveFadingLifecycle(t *testing.T, g *game.Game, engine *Engine, permanent *game.Permanent, initial int) {
	t.Helper()
	for remaining := initial - 1; remaining >= 0; remaining-- {
		if !resolveFadingUpkeep(t, g, engine, game.Player1) {
			t.Fatal("fading permanent did not trigger during its controller's upkeep")
		}
		assertFadeCounters(t, permanent, remaining)
		if !fadingPermanentOnBattlefield(g, permanent) {
			t.Fatal("fading permanent was sacrificed before all fade counters were removed")
		}
	}
	if !resolveFadingUpkeep(t, g, engine, game.Player1) {
		t.Fatal("fading permanent did not trigger when it had no fade counters")
	}
	if fadingPermanentOnBattlefield(g, permanent) {
		t.Fatal("fading permanent remained on the battlefield after its last fade counter was gone")
	}
	if !g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("fading permanent was not sacrificed to its owner's graveyard")
	}
}

func resolveFadingUpkeep(t *testing.T, g *game.Game, engine *Engine, activePlayer game.PlayerID) bool {
	t.Helper()
	g.Turn.ActivePlayer = activePlayer
	g.Turn.Step = game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	triggered := engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if triggered {
		engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	}
	return triggered
}

func assertFadeCounters(t *testing.T, permanent *game.Permanent, want int) {
	t.Helper()
	if got := permanent.Counters.Get(counter.Fade); got != want {
		t.Fatalf("fade counters = %d, want %d", got, want)
	}
}

func fadingPermanentOnBattlefield(g *game.Game, want *game.Permanent) bool {
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == want.ObjectID {
			return true
		}
	}
	return false
}
