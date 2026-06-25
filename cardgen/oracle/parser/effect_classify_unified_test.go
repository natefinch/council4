package parser

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// nthWordIndex returns the token index of the occurrence-th (1-based) token whose
// text equals word (case-insensitive), so classifier tests can anchor on a
// specific verb without hard-coding token positions.
func nthWordIndex(t *testing.T, tokens []shared.Token, word string, occurrence int) int {
	t.Helper()
	seen := 0
	for i := range tokens {
		if equalWord(tokens[i], word) {
			seen++
			if seen == occurrence {
				return i
			}
		}
	}
	t.Fatalf("word %q occurrence %d not found in %q", word, occurrence, sourceText(tokens))
	return -1
}

func sourceText(tokens []shared.Token) string {
	texts := make([]string, len(tokens))
	for i, token := range tokens {
		texts[i] = token.Text
	}
	return strings.Join(texts, " ")
}

// TestEffectKindAtVerbOverrides pins the authoritative effectKindAt classifier on
// every verb case that previously diverged between effectKindAt and the deleted
// legacyEffectKindAt. The single classifier now owns all of these overrides, so
// both the real effect segmentation and the ordered-lowering count agree.
func TestEffectKindAtVerbOverrides(t *testing.T) {
	cases := []struct {
		name       string
		source     string
		word       string
		occurrence int
		want       EffectKind
	}{
		// Cases effectKindAt classified that legacyEffectKindAt did not.
		{"win game", "You win the game.", "win", 1, EffectWinGame},
		{"win without game", "You win.", "win", 1, EffectUnknown},
		{"pay life", "Pay 3 life.", "pay", 1, EffectLose},
		{"remove counter", "Remove a counter from target permanent.", "remove", 1, EffectRemoveCounter},
		{"remove from combat", "Remove target attacking creature from combat.", "remove", 1, EffectRemoveFromCombat},
		{"become monarch", "You become the monarch.", "become", 1, EffectBecomeMonarch},
		{"manifest", "Manifest the top card of your library.", "manifest", 1, EffectManifest},
		{"manifest dread", "Manifest dread.", "manifest", 1, EffectManifestDread},
		{"choose new targets", "Choose new targets for target spell.", "choose", 1, EffectChooseNewTargets},
		{"tap or untap", "Tap or untap target permanent.", "tap", 1, EffectTapOrUntap},
		// Cast static permission suppressed only in effectKindAt before unification.
		{"cast this from graveyard", "You may cast this card from your graveyard.", "cast", 1, EffectUnknown},
		// GrantKeyword player-possession suppression unique to effectKindAt.
		{"player possession grant", "You have no maximum hand size.", "have", 1, EffectUnknown},
		// Cases legacyEffectKindAt suppressed, now folded into effectKindAt.
		{"spell cost modifier first cast", "Artifact spells you cast this turn cost {1} less to cast.", "cast", 1, EffectUnknown},
		{"spell cost modifier second cast", "Artifact spells you cast this turn cost {1} less to cast.", "cast", 2, EffectUnknown},
		{"every creature type rider", "Creatures you control have base power and toughness 4/4 and gain all creature types.", "gain", 1, EffectUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := lexedWords(t, tc.source)
			index := nthWordIndex(t, tokens, tc.word, tc.occurrence)
			if got := effectKindAt(tokens, index); got != tc.want {
				t.Fatalf("effectKindAt(%q @ %d) = %v; want %v", tc.source, index, got, tc.want)
			}
		})
	}
}

// TestOrderedEffectCountConditionExclusion verifies that a verb inside a leading
// if/unless condition clause is excluded from the ordered-lowering count, while
// the segmentation index list still includes it. This preserves the distinction
// that legacyEffectCount drew via effectWithinCondition.
func TestOrderedEffectCountConditionExclusion(t *testing.T) {
	tokens := lexedWords(t, "If you control a creature, draw a card.")
	atoms := Atoms{}

	if got := orderedEffectCount(tokens, atoms); got != 1 {
		t.Fatalf("orderedEffectCount = %d; want 1 (condition verb excluded)", got)
	}
	// The "draw" effect outside the condition is the only counted verb.
	if got := len(effectIndices(tokens, atoms)); got != 1 {
		t.Fatalf("len(effectIndices) = %d; want 1", got)
	}
}

// TestOrderedEffectCountMultiEffect confirms a genuine two-effect sentence counts
// both verbs and so drives the ordered-lowering path.
func TestOrderedEffectCountMultiEffect(t *testing.T) {
	tokens := lexedWords(t, "Draw a card and gain 2 life.")
	atoms := Atoms{}
	if got := orderedEffectCount(tokens, atoms); got != 2 {
		t.Fatalf("orderedEffectCount = %d; want 2", got)
	}
	if got := len(effectIndices(tokens, atoms)); got != 2 {
		t.Fatalf("len(effectIndices) = %d; want 2", got)
	}
}

// TestEffectIndicesVsOrderedCountDivergence pins the deliberate per-consumer
// filtering: the noun-form "next untap step" verb is excluded from the effect
// segmentation (effectNounAt) but still contributes to the ordered count, exactly
// as the two classifiers diverged before unification.
func TestEffectIndicesVsOrderedCountDivergence(t *testing.T) {
	tokens := lexedWords(t, "Untap it during the next untap step.")
	atoms := Atoms{}

	indices := effectIndices(tokens, atoms)
	if len(indices) != 1 {
		t.Fatalf("len(effectIndices) = %d; want 1 (noun-form untap excluded)", len(indices))
	}
	if got := orderedEffectCount(tokens, atoms); got != 2 {
		t.Fatalf("orderedEffectCount = %d; want 2 (noun-form untap counted)", got)
	}
}

// TestSuppressedShapesYieldNoEffects verifies the folded suppressions remove both
// the segmentation entry and the ordered-lowering contribution so neither path
// double-counts the spell-cost-modifier or every-creature-type rider shapes.
func TestSuppressedShapesYieldNoEffects(t *testing.T) {
	atoms := Atoms{}

	spellCost := lexedWords(t, "Artifact spells you cast this turn cost {1} less to cast.")
	if got := len(effectIndices(spellCost, atoms)); got != 0 {
		t.Fatalf("spell-cost effectIndices = %d; want 0", got)
	}
	if got := orderedEffectCount(spellCost, atoms); got != 0 {
		t.Fatalf("spell-cost orderedEffectCount = %d; want 0", got)
	}
}

// TestClassifyEffectVerbsSharedSource confirms both derivations read the same
// classification pass: every effectIndices entry has a matching classifiedVerb
// with the same kind effectKindAt reports for that token.
func TestClassifyEffectVerbsSharedSource(t *testing.T) {
	tokens := lexedWords(t, "Draw a card and gain 2 life.")
	atoms := Atoms{}

	verbs := classifyEffectVerbs(tokens, atoms)
	for _, verb := range verbs {
		if verb.Kind != effectKindAt(tokens, verb.Index) {
			t.Fatalf("classifiedVerb kind %v != effectKindAt %v at %d", verb.Kind, effectKindAt(tokens, verb.Index), verb.Index)
		}
	}
	for _, index := range effectIndices(tokens, atoms) {
		found := false
		for _, verb := range verbs {
			if verb.Index == index {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("effectIndices entry %d absent from classifyEffectVerbs", index)
		}
	}
}
