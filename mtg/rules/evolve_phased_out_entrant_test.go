package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// Source review (#3560): Fathom Mage (Oracle 93d0e129-e3b5-4aff-9e50-f34771ed00ff,
// printing 22be6c55-f95e-498e-b6d2-65b7ed57d2d3, pinned corpus), 1/1 "Evolve
// (Whenever a creature you control enters, if that creature has greater power
// or toughness than this creature, put a +1/+1 counter on this creature.)" and
// "Whenever a +1/+1 counter is put on this creature, you may draw a card."
// Its 2013-04-15 rulings: the comparison happens again as evolve tries to
// resolve, and if the entered creature has left the battlefield its last known
// power and toughness are used. Intangible Virtue (Oracle
// d21c3c8f-d105-4ba9-bf69-e5f26f0f8ec5, printing
// 716d82a1-fb5d-4607-b1fa-e8e9edd4dd77): "Creature tokens you control get
// +1/+1 and have vigilance." Official CR effective 2026-09-25
// (media.wizards.com MagicCompRules 20260925.txt) read: 603.4 (an intervening
// "if" is rechecked on resolution), 608.2h and 113.7a (an object no longer in
// its expected zone is read from last known information), 702.26b (a
// phased-out permanent is treated as though it does not exist), 702.26c (a
// permanent that phases in is again treated as existing), 702.26d (phasing
// doesn't change zones; the phased-out permanent is treated as not on the
// battlefield) and 702.100a/c (evolve compares power to power and toughness to
// toughness against creatures). A phased-out entrant is therefore compared
// using its stats as it phased out, unaffected by static effects that change
// afterward; once phased in as the same object it is compared by its current
// stats again.

func evolvePhasedOutEntrantScenario(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mage := addCombatPermanent(g, game.Player1, entryHistoryRegistryCard(t, "Fathom Mage"))
	mage.Counters.Add(counter.PlusOnePlusOne, 1)
	virtue := addCombatPermanent(g, game.Player1, entryHistoryRegistryCard(t, "Intangible Virtue"))
	tokens, ok := createTokenPermanentsCollectingWithChoices(engine, g, game.Player1,
		vanillaCreature("Two-Two Token", 2, 2), 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !ok || len(tokens) != 1 {
		t.Fatalf("created %d tokens, want 1", len(tokens))
	}
	token := tokens[0]
	if values := effectivePermanentValues(g, token); values.power != 3 || values.toughness != 3 {
		t.Fatalf("token with Intangible Virtue is %d/%d, want 3/3", values.power, values.toughness)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("Fathom Mage's evolve did not trigger for a 3/3 token entering beside the 2/2 Mage")
	}
	phaseOutForTest(t, g, token)
	addEffectSpellToStack(g, game.Player1, game.Destroy{Object: game.TargetPermanentReference(0)},
		[]game.Target{game.PermanentTarget(virtue.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, virtue.ObjectID); ok {
		t.Fatal("Intangible Virtue is still on the battlefield")
	}
	return g, engine, mage, token, virtue
}

func TestEvolveUsesPhasedOutEntrantLastKnownStats(t *testing.T) {
	t.Parallel()
	g, engine, mage, token, _ := evolvePhasedOutEntrantScenario(t)
	if !token.PhasedOut {
		t.Fatal("token phased in before evolve resolved")
	}
	if values := effectivePermanentValues(g, token); values.power != 2 || values.toughness != 2 {
		t.Fatalf("phased-out token now computes as %d/%d, want 2/2 without Intangible Virtue", values.power, values.toughness)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := mage.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("Fathom Mage has %d +1/+1 counters, want 2: the token was last known as a 3/3 when it phased out", got)
	}
}

func TestEvolveUsesPhasedInEntrantCurrentStats(t *testing.T) {
	t.Parallel()
	g, engine, mage, token, _ := evolvePhasedOutEntrantScenario(t)
	// Phasing in happens only in the untap step, so the pending evolve ability
	// is held aside while the real untap-step phase-in runs.
	evolve, ok := g.Stack.Pop()
	if !ok || evolve.Kind != game.StackTriggeredAbility || evolve.SourceID != mage.ObjectID {
		t.Fatal("evolve trigger was not the remaining stack object")
	}
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if token.PhasedOut {
		t.Fatal("token did not phase in during its controller's untap step")
	}
	g.Stack.Push(evolve)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := mage.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Fathom Mage has %d +1/+1 counters, want 1: the phased-in token is a 2/2 again", got)
	}
}
