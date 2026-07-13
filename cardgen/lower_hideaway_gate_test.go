package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
)

// hideawayPlayInstruction returns the sole instruction of a Hideaway land's
// {cost}, {T} play ability, asserting the lowered shape is exactly one activated
// ability whose single instruction is an optional PlayHideawayCard.
func hideawayPlayInstruction(t *testing.T, face loweredFaceAbilities) game.Instruction {
	t.Helper()
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	modes := face.ActivatedAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("ability content = %#v, want one mode with one instruction", face.ActivatedAbilities[0].Content)
	}
	instruction := modes[0].Sequence[0]
	if _, ok := instruction.Primitive.(game.PlayHideawayCard); !ok {
		t.Fatalf("primitive = %#v, want PlayHideawayCard", instruction.Primitive)
	}
	if !instruction.Optional {
		t.Fatal("Hideaway play instruction must be optional (the controller may decline)")
	}
	return instruction
}

// hideawayGateCondition returns the shared condition gating a Hideaway play
// instruction, asserting an effect condition carries one.
func hideawayGateCondition(t *testing.T, instruction game.Instruction) game.Condition {
	t.Helper()
	if !instruction.Condition.Exists || !instruction.Condition.Val.Condition.Exists {
		t.Fatalf("instruction condition = %#v, want a gating condition", instruction.Condition)
	}
	return instruction.Condition.Val.Condition.Val
}

// TestLowerSpinerockKnollHideawayGate lowers Spinerock Knoll's exact printed
// Oracle text and proves its {R}, {T} Hideaway play ability gates on an opponent
// having been dealt 7 or more damage this turn via the reusable aggregate.
func TestLowerSpinerockKnollHideawayGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Spinerock Knoll",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)\n" +
			"This land enters tapped.\n" +
			"{T}: Add {R}.\n" +
			"{R}, {T}: You may play the exiled card without paying its mana cost if an opponent was dealt 7 or more damage this turn.",
	})
	condition := hideawayGateCondition(t, hideawayPlayInstruction(t, face))
	if got := condition.Aggregates; len(got) != 1 ||
		got[0].Aggregate != game.AggregateAnyOpponentDamageTakenThisTurn ||
		got[0].Op != compare.GreaterOrEqual ||
		got[0].Value != 7 {
		t.Fatalf("condition = %#v, want opponent-damage-this-turn >= 7", condition)
	}
}

// TestLowerShelldockIsleHideawayGate lowers Shelldock Isle's exact printed
// Oracle text and proves its {U}, {T} Hideaway play ability gates on a library
// holding twenty or fewer cards via the reusable minimum-library aggregate.
func TestLowerShelldockIsleHideawayGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Shelldock Isle",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)\n" +
			"This land enters tapped.\n" +
			"{T}: Add {U}.\n" +
			"{U}, {T}: You may play the exiled card without paying its mana cost if a library has twenty or fewer cards in it.",
	})
	condition := hideawayGateCondition(t, hideawayPlayInstruction(t, face))
	if got := condition.Aggregates; len(got) != 1 ||
		got[0].Aggregate != game.AggregateMinPlayerLibrarySize ||
		got[0].Op != compare.LessOrEqual ||
		got[0].Value != 20 {
		t.Fatalf("condition = %#v, want min-library <= 20", condition)
	}
}

// TestLowerHowltoothHollowHideawayGate lowers Howltooth Hollow's exact printed
// Oracle text and proves its {B}, {T} Hideaway play ability gates on every
// player having an empty hand via the reusable universal predicate.
func TestLowerHowltoothHollowHideawayGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Howltooth Hollow",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)\n" +
			"This land enters tapped.\n" +
			"{T}: Add {B}.\n" +
			"{B}, {T}: You may play the exiled card without paying its mana cost if each player has no cards in hand.",
	})
	condition := hideawayGateCondition(t, hideawayPlayInstruction(t, face))
	if !condition.AllPlayersHandEmpty {
		t.Fatalf("condition = %#v, want AllPlayersHandEmpty", condition)
	}
	if len(condition.Aggregates) != 0 {
		t.Fatalf("condition = %#v, want no aggregates for a universal hand gate", condition)
	}
}

// TestHideawayGatePredicatesFailClosedInReplacementContext confirms the three
// reusable gate predicates are rejected in the replacement context, where no
// resolving object exists to evaluate turn-scoped or all-player board state.
func TestHideawayGatePredicatesFailClosedInReplacementContext(t *testing.T) {
	t.Parallel()
	predicates := []compiler.ConditionPredicate{
		compiler.ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
		compiler.ConditionPredicateAnyLibrarySizeAtMost,
		compiler.ConditionPredicateAllPlayersHandEmpty,
	}
	for _, predicate := range predicates {
		if conditionPredicateAllowedInContext(predicate, conditionContextReplacement) {
			t.Fatalf("predicate %v must fail closed in the replacement context", predicate)
		}
		if !conditionPredicateAllowedInContext(predicate, conditionContextEffectGate) {
			t.Fatalf("predicate %v must be allowed in the Hideaway effect-gate context", predicate)
		}
	}
}
