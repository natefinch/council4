package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

func counterPumpCardDefinition(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s lookup returned %q", name, def.Name)
	}
	return def
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; Common Bond RTR 151 printing revision 59965953-1522-4103-9a6c-5534205d34d9, mana cost {1}{G}{W}.
// Oracle text: “Put a +1/+1 counter on target creature.
// Put a +1/+1 counter on target creature.”
// https://api.scryfall.com/cards/59965953-1522-4103-9a6c-5534205d34d9/rulings retrieved 2026-11-11 returned the WotC 2012-10-01 ruling allowing the same or different creatures.
// Official Wizards CR effective 2026-09-25, https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf; read 115.3, 601.2c/h, 608.2c, 122.1a, 613.4c, and 304.2. These require separate target instances (which may name the same object), complete target/cost announcement, ordered resolution, +1/+1 counters modifying P/T, and resolved instants going to their owner's graveyard.
func TestCommonBondScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Common Bond target", 1, 1))
	spell := addCardToHand(g, game.Player1, counterPumpCardDefinition(t, "d52e9255-62e7-479f-8def-e92b96efdb89", "Common Bond"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(target.ObjectID),
		game.PermanentTarget(target.ObjectID),
	}, 0, nil)) {
		t.Fatal("Common Bond did not cast with the same creature chosen for both target instances")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Common Bond did not pay {1}{G}{W}")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("target +1/+1 counters = %d, want 2", got)
	}
	if got := effectivePower(g, target); got != 3 {
		t.Fatalf("target power = %d, want 3", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 3 {
		t.Fatalf("target toughness = %d, valid %t; want 3", got, ok)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Common Bond did not go to its owner's graveyard")
	}
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; Seeds of Strength RAV 227 printing revision bfa1ac13-adbd-439a-90b2-9c506aec0836, mana cost {G}{W}.
// Oracle text:
// “Target creature gets +1/+1 until end of turn.
// Target creature gets +1/+1 until end of turn.
// Target creature gets +1/+1 until end of turn.”
// https://api.scryfall.com/cards/bfa1ac13-adbd-439a-90b2-9c506aec0836/rulings retrieved 2026-11-11 returned the WotC 2005-10-01 ruling permitting any assignment across one, two, or three creatures.
// Official Wizards CR effective 2026-09-25, https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf; read 115.3, 601.2c/h, 608.2c, 611.2a, 613.4c, and 304.2. Each clause has its own target occurrence; each resolving P/T effect lasts until end of turn in layer 7c, and the instant then goes to its owner's graveyard.
func TestSeedsOfStrengthScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targets := []*game.Permanent{
		addCombatPermanent(g, game.Player2, vanillaCreature("Seeds target one", 1, 1)),
		addCombatPermanent(g, game.Player2, vanillaCreature("Seeds target two", 1, 1)),
		addCombatPermanent(g, game.Player2, vanillaCreature("Seeds target three", 1, 1)),
	}
	spell := addCardToHand(g, game.Player1, counterPumpCardDefinition(t, "dd3d4d96-1e88-49e1-8276-7f62c22ad91a", "Seeds of Strength"))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(targets[0].ObjectID),
		game.PermanentTarget(targets[1].ObjectID),
		game.PermanentTarget(targets[2].ObjectID),
	}, 0, nil)) {
		t.Fatal("Seeds of Strength did not cast with three creatures targeted")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Seeds of Strength did not pay {G}{W}")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	for i, target := range targets {
		if got := effectivePower(g, target); got != 2 {
			t.Errorf("target %d power = %d, want 2", i, got)
		}
		if got, ok := effectiveToughness(g, target); !ok || got != 2 {
			t.Errorf("target %d toughness = %d, valid %t; want 2", i, got, ok)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Seeds of Strength did not go to its owner's graveyard")
	}
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; Martial Glory 2X2 249 printing revision 9bf662e7-e63f-40c2-b163-1ebce47d8696, mana cost {R}{W}.
// Oracle text: “Target creature gets +3/+0 until end of turn.
// Target creature gets +0/+3 until end of turn.”
// https://api.scryfall.com/cards/9bf662e7-e63f-40c2-b163-1ebce47d8696/rulings retrieved 2026-11-11 returned the WotC 2013-01-24 ruling that both targets may be the same or different creatures.
// Official Wizards CR effective 2026-09-25, https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf; read 115.3, 601.2c/h, 608.2c, 611.2a, 613.4c, and 304.2. Separate target words allow a repeated object; the two ordered P/T effects last until end of turn in layer 7c, and the instant then goes to its owner's graveyard.
func TestMartialGloryScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	powerTarget := addCombatPermanent(g, game.Player2, vanillaCreature("Martial Glory power target", 2, 2))
	toughnessTarget := addCombatPermanent(g, game.Player2, vanillaCreature("Martial Glory toughness target", 2, 2))
	spell := addCardToHand(g, game.Player1, counterPumpCardDefinition(t, "e8aa12f8-8922-41ed-8cdb-219c9e4b9e36", "Martial Glory"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(powerTarget.ObjectID),
		game.PermanentTarget(toughnessTarget.ObjectID),
	}, 0, nil)) {
		t.Fatal("Martial Glory did not cast with two creatures targeted")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Martial Glory did not pay {R}{W}")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, powerTarget); got != 5 {
		t.Fatalf("first target power = %d, want 5", got)
	}
	if got, ok := effectiveToughness(g, powerTarget); !ok || got != 2 {
		t.Fatalf("first target toughness = %d, valid %t; want 2", got, ok)
	}
	if got := effectivePower(g, toughnessTarget); got != 2 {
		t.Fatalf("second target power = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, toughnessTarget); !ok || got != 5 {
		t.Fatalf("second target toughness = %d, valid %t; want 5", got, ok)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Martial Glory did not go to its owner's graveyard")
	}
}
