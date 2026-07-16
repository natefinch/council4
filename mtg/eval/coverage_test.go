package eval

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// knownPrimitiveCount is the number of game.Primitive implementations the
// translator was last reconciled against. ScorableEffect classifies the
// value-dominant primitives and treats the rest as value-neutral; when this
// guard fails because a primitive was added or removed, review the new
// primitive in appendPrimitiveAtoms (give it a value atom or confirm neutral)
// and update this constant.
//
// EachPlayerChooseDestroy is treated as value-neutral (no atom), matching the
// sibling distributive removals DestroyForEachPlayer and ExileForEachPlayer:
// its per-player player-relative choices are not modeled by the single-target
// atom heuristic.
//
// CantBecomeMonarch is value-neutral (no atom): the monarchy restriction has no
// measurable card/life/board effect, matching BecomeMonarch.
//
// PlayerMayPayGenericOrRule is value-neutral (no atom): its outcome is either an
// optional mana payment or an installed combat rule restriction, neither of
// which the single-target card/life/board atom heuristic models, matching Pay
// and ApplyRule.
//
// ExileForEachOpponent and DrawForEachExiled are value-neutral (no atom),
// matching the sibling distributive removals and per-controller payoffs
// (DestroyForEachPlayer, ExileForEachPlayer, CreateTokenForEachDestroyed): their
// per-opponent player-relative choices and linked-set payoffs are not modeled by
// the single-target atom heuristic.
// ManifestForEachLinked is likewise value-neutral: the number and controllers of
// the manifested cards depend on the linked removal's runtime choices.
//
// CreateReflexiveTrigger is value-neutral (no atom), matching CreateDelayedTrigger:
// the value lives in the reflexive ability's own content, which is scored when
// that content resolves, not in the primitive that puts it on the stack.
//
// ExilePermanentForPlay carries the EffectPermanentRemoved atom, matching Exile
// and Bounce: it is a targeted permanent removal whose owner-play grant is not
// modeled by the single-target atom heuristic.
//
// Bolster carries the EffectCounterAdded atom, matching Monstrosity: it puts
// +1/+1 counters on the controller's least-toughness creature (CR 701.37), a
// board-boosting counter placement.
//
// ChooseDrawnPayLifeOrTop is value-neutral (no atom), matching
// PlayerMayPayGenericOrRule: for each chosen card its outcome is either a life
// payment or returning a just-drawn card to the library, a wash the
// single-target atom heuristic does not model. The card advantage of the extra
// draw is already carried by the preceding Draw instruction.
//
// ExileTopEachLibraryCastFree is value-neutral (no atom), matching the sibling
// impulse free-cast primitives ImpulseExile and ExileLibraryUntilNonlandCast:
// the number and value of spells the controller ends up casting from the exiled
// pile is indeterminate and player-relative, which the single-target atom
// heuristic does not model.
//
// RecordEchoObligation is value-neutral (no atom): it only records which player
// has resolved an Echo permanent's upkeep obligation (CR 702.29) so later
// upkeeps do not re-trigger; it has no card/life/board effect of its own.
//
// GainCityBlessing is value-neutral (no atom): granting the city's blessing sets
// a persistent player designation whose payoff lives in the abilities that read
// "if you have the city's blessing", not in the primitive that installs it,
// matching BecomeMonarch and CantBecomeMonarch.
//
// CopyCard and PlayLinkedExiledCard are value-neutral (no atom): copying a
// linked exiled card and casting that copy for free (Isochron Scepter,
// Spellbinder) derive their value from the copied card's own effects, not from
// the copy/cast primitives themselves, matching CastForFree and CopyStackObject.
//
// TapChosenGroup is value-neutral (no atom): it optionally taps any number of
// the controller's own matching permanents to publish that count (Myr
// Battlesphere), so its payoff lives in the scaled ModifyPT and Damage
// instructions that read the count, not in the self-tap enabler itself.
//
// IterativeLibraryProcess carries the EffectCardsDrawn atom (amount 1) for its
// controller: the process digs through the top of the library and nets a single
// found card into hand (Tainted Pact, Demonic Consultation), matching the other
// "card into hand" library primitives.
//
// ManifestForEachLinked is value-neutral: the number and controllers of the
// manifested cards depend on the linked removal's runtime choices.
//
// Incubate is value-neutral (no atom), matching Amass: it creates a transforming
// Incubator token with +1/+1 counters for a player who is typically the exiled
// permanent's controller (an opponent), so its board payoff is player-relative
// and not modeled by the single-target atom heuristic.
//
// CorrelatedFight pairs a created-token group with a counted-permanent group and
// fights each pair (Ezuri's Predation); it is valued as dynamic targeted damage,
// mirroring the single Fight's targeted-damage atom.
//
// ExileTargetSpells is value-neutral (no atom): exiling any number of target
// spells removes them from the stack (Mindbreak Trap), so like the CounterObject
// stack-removal primitive it has no modeled value atom in this vocabulary.
//
// The dungeon/initiative primitives (VentureIntoDungeon, VentureIntoUndercity,
// TakeInitiative) are value-neutral here: they queue deferred room abilities and
// designations rather than an immediate card-advantage/life swing, so like the
// monarch-designation primitives they fall through to the default with no
// modeled value atom.
const knownPrimitiveCount = 144

// TestPrimitiveCountIsReconciled keeps a newly added resolution primitive from
// silently falling through the translator: adding one trips this guard so its
// value classification is considered.
func TestPrimitiveCountIsReconciled(t *testing.T) {
	gameDir := filepath.Join("..", "game")
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		t.Fatalf("reading %s: %v", gameDir, err)
	}
	pattern := regexp.MustCompile(`func \([A-Za-z]+\) isPrimitive\(\)`)
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		source, err := os.ReadFile(filepath.Join(gameDir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}
		for _, match := range pattern.FindAllString(string(source), -1) {
			seen[match] = true
		}
	}
	if len(seen) != knownPrimitiveCount {
		t.Fatalf("found %d game.Primitive implementations, knownPrimitiveCount = %d; "+
			"reconcile appendPrimitiveAtoms with the change and update the constant",
			len(seen), knownPrimitiveCount)
	}
}
