package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// tibaltsTrickeryOracleText is the authoritative current Oracle text of Tibalt's
// Trickery, copied from the Scryfall snapshot the corpus compiler reads.
const tibaltsTrickeryOracleText = "Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, " +
	"then exiles cards from the top of their library until they exile a nonland card with a different " +
	"name than that spell. They may cast that card without paying its mana cost. Then they put the " +
	"exiled cards on the bottom of their library in a random order."

func tibaltsTrickeryCard(oracleText string) *ScryfallCard {
	return &ScryfallCard{
		Name:       "Tibalt's Trickery",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{R}",
		OracleText: oracleText,
	}
}

// TestLowerTibaltsTrickerySequence proves the whole six-effect Oracle body folds
// into the exact six-instruction resolution sequence: counter the single spell
// target, choose 1..3 at random into the shared mill-count key, mill that many
// from the countered spell's controller, publish the iterative process outputs,
// optionally cast the found card, and random-bottom the linked remainder.
func TestLowerTibaltsTrickerySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, tibaltsTrickeryCard(tibaltsTrickeryOracleText))

	if !face.SpellAbility.Exists {
		t.Fatal("Tibalt's Trickery produced no spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]

	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want a single spell target", mode.Targets)
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowStackObject ||
		len(target.Predicate.StackObjectKinds) != 1 ||
		target.Predicate.StackObjectKinds[0] != game.StackSpell {
		t.Fatalf("target spec = %#v, want a stack-spell target", target)
	}

	if len(mode.Sequence) != 6 {
		t.Fatalf("sequence length = %d, want 6:\n%#v", len(mode.Sequence), mode.Sequence)
	}

	targetRef := game.TargetStackObjectReference(0)
	controller := game.ObjectControllerReference(targetRef)

	counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
	if !ok || counter.Object != targetRef || counter.ExileInstead {
		t.Fatalf("sequence[0] = %#v, want CounterObject on target 0", mode.Sequence[0].Primitive)
	}

	choose, ok := mode.Sequence[1].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want Choose", mode.Sequence[1].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceNumber ||
		choose.Choice.MinNumber != 1 || choose.Choice.MaxNumber != 3 || !choose.Choice.AtRandom {
		t.Fatalf("choose = %#v, want a random 1..3 number choice", choose.Choice)
	}
	if string(choose.PublishChoice) != "tibalts-trickery-mill-count" {
		t.Fatalf("PublishChoice = %q, want the shared mill-count key", choose.PublishChoice)
	}

	mill, ok := mode.Sequence[2].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("sequence[2] = %#v, want Mill", mode.Sequence[2].Primitive)
	}
	if mill.Player != controller {
		t.Fatalf("mill player = %#v, want the countered spell's controller", mill.Player)
	}
	amount := mill.Amount.DynamicAmount()
	if !amount.Exists || amount.Val.Kind != game.DynamicAmountChosenNumber ||
		string(amount.Val.ResultKey) != "tibalts-trickery-mill-count" {
		t.Fatalf("mill amount = %#v, want the chosen mill-count number", mill.Amount)
	}

	process, ok := mode.Sequence[3].Primitive.(game.IterativeLibraryProcess)
	if !ok {
		t.Fatalf("sequence[3] = %#v, want IterativeLibraryProcess", mode.Sequence[3].Primitive)
	}
	if process.Player != controller {
		t.Fatalf("process player = %#v, want the countered spell's controller", process.Player)
	}
	if process.Stop != game.IterativeLibraryStopDifferentNameNonland {
		t.Fatalf("process stop = %v, want DifferentNameNonland", process.Stop)
	}
	if process.DifferentNameFrom != targetRef {
		t.Fatalf("process DifferentNameFrom = %#v, want target 0", process.DifferentNameFrom)
	}
	if process.PublishLinked != tibaltsTrickeryExiledKey {
		t.Fatalf("process PublishLinked = %q, want %q", process.PublishLinked, tibaltsTrickeryExiledKey)
	}
	if process.ChooseName || process.Reveal || process.OptionalTake || process.AllowAbsentName {
		t.Fatalf("process carries unexpected iterative knobs: %#v", process)
	}
	if mode.Sequence[3].PublishResult != tibaltsTrickeryFoundKey {
		t.Fatalf("process PublishResult = %q, want %q", mode.Sequence[3].PublishResult, tibaltsTrickeryFoundKey)
	}
	cast, ok := mode.Sequence[4].Primitive.(game.CastForFree)
	if !ok || cast.Player != controller || cast.Zone != zone.Exile ||
		cast.Card.Kind != game.CardReferenceLinked ||
		cast.Card.LinkID != string(tibaltsTrickeryExiledKey) {
		t.Fatalf("sequence[4] = %#v, want linked free cast", mode.Sequence[4])
	}
	if !mode.Sequence[4].Optional || !mode.Sequence[4].OptionalActor.Exists ||
		mode.Sequence[4].OptionalActor.Val != controller || !mode.Sequence[4].ResultGate.Exists ||
		mode.Sequence[4].ResultGate.Val.Key != tibaltsTrickeryFoundKey ||
		mode.Sequence[4].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("sequence[4] gates = %#v, want optional found-result gate", mode.Sequence[4])
	}
	bottom, ok := mode.Sequence[5].Primitive.(game.PutLinkedExiledCardsInLibrary)
	if !ok || bottom.LinkedKey != tibaltsTrickeryExiledKey || !bottom.Bottom || !bottom.RandomOrder {
		t.Fatalf("sequence[5] = %#v, want linked random-bottom disposal", mode.Sequence[5])
	}
}

// TestGenerateTibaltsTrickerySource proves the full parse-through-render pipeline
// emits a clean, diagnostic-free CardDef whose source carries every typed node of
// the resolution sequence.
func TestGenerateTibaltsTrickerySource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(tibaltsTrickeryCard(tibaltsTrickeryOracleText), "testcards")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	normalized := strings.Join(strings.Fields(source), " ")
	for _, want := range []string{
		"game.CounterObject{",
		"game.TargetStackObjectReference(0)",
		"game.Choose{",
		"Kind: game.ResolutionChoiceNumber,",
		"MinNumber: 1,",
		"MaxNumber: 3,",
		"AtRandom: true,",
		"PublishChoice: game.ChoiceKey(\"tibalts-trickery-mill-count\"),",
		"game.Mill{",
		"Kind: game.DynamicAmountChosenNumber,",
		"ResultKey: game.ResultKey(\"tibalts-trickery-mill-count\"),",
		"game.ObjectControllerReference(game.TargetStackObjectReference(0))",
		"game.IterativeLibraryProcess{",
		"Stop: game.IterativeLibraryStopDifferentNameNonland,",
		"PublishLinked: game.LinkedKey(\"tibalts-trickery-exiled\"),",
		"game.CastForFree{",
		"Kind: game.CardReferenceLinked,",
		"LinkID: \"tibalts-trickery-exiled\"",
		"game.PutLinkedExiledCardsInLibrary{",
		"RandomOrder: true,",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("generated source missing %q:\n%s", want, normalized)
		}
	}
}

// tibaltPrimitivePresent reports whether any spell-ability instruction in the
// lowered face is the different-name-nonland iterative process, i.e. whether the
// face was recognized as Tibalt's Trickery.
func tibaltPrimitivePresent(face loweredFaceAbilities) bool {
	if !face.SpellAbility.Exists {
		return false
	}
	for _, mode := range face.SpellAbility.Val.Modes {
		for _, instruction := range mode.Sequence {
			if process, ok := instruction.Primitive.(game.IterativeLibraryProcess); ok &&
				process.Stop == game.IterativeLibraryStopDifferentNameNonland {
				return true
			}
		}
	}
	return false
}

// TestTibaltsTrickeryNearMissesFailClosed proves the recognizer is strict: any
// wording that deviates from Tibalt's Trickery's exact sequence is never lowered
// into the different-name-nonland iterative process, so a near-miss can only be
// rejected, never silently mis-implemented.
func TestTibaltsTrickeryNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
	}{
		{
			name: "different mana value predicate",
			oracleText: "Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, " +
				"then exiles cards from the top of their library until they exile a nonland card with a different " +
				"mana value than that spell. They may cast that card without paying its mana cost. Then they put the " +
				"exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "fixed mill count without random choice",
			oracleText: "Counter target spell. Its controller mills three cards, " +
				"then exiles cards from the top of their library until they exile a nonland card with a different " +
				"name than that spell. They may cast that card without paying its mana cost. Then they put the " +
				"exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "bottom in fixed order",
			oracleText: "Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, " +
				"then exiles cards from the top of their library until they exile a nonland card with a different " +
				"name than that spell. They may cast that card without paying its mana cost. Then they put the " +
				"exiled cards on the bottom of their library.",
		},
		{
			name: "cast with payment",
			oracleText: "Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, " +
				"then exiles cards from the top of their library until they exile a nonland card with a different " +
				"name than that spell. They may cast that card. Then they put the " +
				"exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "missing counter step",
			oracleText: "Choose 1, 2, or 3 at random. Its controller mills that many cards, " +
				"then exiles cards from the top of their library until they exile a nonland card with a different " +
				"name than that spell. They may cast that card without paying its mana cost. Then they put the " +
				"exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "missing mill step",
			oracleText: "Counter target spell. Choose 1, 2, or 3 at random. Its controller " +
				"exiles cards from the top of their library until they exile a nonland card with a different " +
				"name than that spell. They may cast that card without paying its mana cost. Then they put the " +
				"exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "must cast rather than may",
			oracleText: "Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, " +
				"then exiles cards from the top of their library until they exile a nonland card with a different " +
				"name than that spell. They cast that card without paying its mana cost. Then they put the " +
				"exiled cards on the bottom of their library in a random order.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			faces, _ := lowerExecutableFaces(tibaltsTrickeryCard(test.oracleText))
			for _, face := range faces {
				if tibaltPrimitivePresent(face) {
					t.Fatalf("near-miss %q was mis-recognized as Tibalt's Trickery", test.name)
				}
			}
		})
	}
}
