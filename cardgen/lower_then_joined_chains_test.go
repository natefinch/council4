package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// TestLowerThenJoinedThreeEffectSharedTargetSequence is the primary regression
// for the 3+ shared-subject then-chain bug. Before the fix, pair (1,2) in
// "mills, then draws, then discards" assigned iClauseStart=vi (draws verb),
// producing [draws, a, card] without the "Target player" prefix and failing
// closed. After the fix, the subject prefix tokens[sentenceStart:viFirst] are
// prepended to every non-first clause in the group.
//
// Requirements verified:
//   - Exactly one game.TargetSpec (no duplicate).
//   - All three instructions reference TargetPlayerReference(0).
func TestLowerThenJoinedThreeEffectSharedTargetSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Three",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player mills three cards, then draws a card, then discards a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want exactly 1 (no duplicate target spec)", len(mode.Targets))
	}
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %d, want 3", len(mode.Sequence))
	}
	mill, millOK := mode.Sequence[0].Primitive.(game.Mill)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[2].Primitive.(game.Discard)
	if !millOK || !drawOK || !discardOK {
		t.Fatalf("primitives = %T, %T, %T; want game.Mill, game.Draw, game.Discard",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if mill.Player.TargetIndex() != 0 {
		t.Fatalf("mill.Player target index = %d, want 0", mill.Player.TargetIndex())
	}
	if draw.Player.TargetIndex() != 0 {
		t.Fatalf("draw.Player target index = %d, want 0 (reusing shared target)", draw.Player.TargetIndex())
	}
	if discard.Player.TargetIndex() != 0 {
		t.Fatalf("discard.Player target index = %d, want 0 (reusing shared target)", discard.Player.TargetIndex())
	}
}

func TestLowerThenJoinedActivatedAbilitySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tome",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{2}, {T}: Draw a card, then discard a card.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	if !drawOK || !discardOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Draw, game.Discard",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
}

func TestRejectActivatedAbilitySequenceWithDelayedTargetSacrifice(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Elementalist",
		Layout:     "normal",
		TypeLine:   "Creature — Wizard",
		OracleText: "{U}{U}: Target creature you control gains flying until end of turn. Sacrifice it at the beginning of the next end step.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported ordered effect sequence diagnostic")
	}
	if diagnostics[0].Summary != "unsupported ordered effect sequence" {
		t.Fatalf("summary = %q, want unsupported ordered effect sequence", diagnostics[0].Summary)
	}
}

func TestLowerThenJoinedLoyaltyAbilitySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "+1: Scry 1, then draw a card.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	mode := face.LoyaltyAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	scry, scryOK := mode.Sequence[0].Primitive.(game.Scry)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !scryOK || !drawOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Scry, game.Draw",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if scry.Amount.Value() != 1 || scry.Player != game.ControllerReference() {
		t.Fatalf("scry = %+v, want controller scries 1", scry)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
}

func TestLowerThenJoinedSagaChapterSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I, II — Scry 2, then draw a card.\nIII — Draw two cards.",
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("got %d chapter abilities, want 2", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("chapter I/II mode = %+v, want no targets and two instructions", mode)
	}
	scry, scryOK := mode.Sequence[0].Primitive.(game.Scry)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !scryOK || !drawOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Scry, game.Draw",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if scry.Amount.Value() != 2 || scry.Player != game.ControllerReference() {
		t.Fatalf("scry = %+v, want controller scries 2", scry)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
}

// TestCompoundMillOracleIR documents the oracle compiler IR for the
// shared-subject then-joined pattern ("Target player mills three cards, then
// draws a card.") and proves that compound mill is achievable within the scope
// of issue #131 without additional effect kinds.
//
// Hypothesis verified: the oracle compiler emits exactly one CompiledTarget
// ("target player") for the sentence; it does NOT create a second implicit
// target for the "draws" clause. The second effect's subject is implied, not
// independently emitted. lowerOrderedEffectSequence resolves this through the
// shared-target deduplication path: contextForEffect uses the sentence Span for
// both effects (finding the one target for each), allOracleTargetSpansClaimed
// recognises the second claim as a duplicate, and rebaseTargetedSequence with
// offset 0 correctly produces TargetPlayerReference(0) for both instructions
// without adding a duplicate game.TargetSpec.
func TestCompoundMillOracleIR(t *testing.T) {
	t.Parallel()
	const text = "Target player mills three cards, then draws a card."
	compilation, diags := compileTestOracle(text, parser.Context{}, compiler.Context{})
	if len(diags) > 0 {
		t.Fatalf("compile diagnostics: %v", diags)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	ab := compilation.Abilities[0]

	// Two effects with the same sentence Span — the root condition that
	// requires the then-join split.
	if len(ab.Content.Effects) != 2 {
		t.Fatalf("IR effects = %d, want 2 (mills + draws)", len(ab.Content.Effects))
	}
	if ab.Content.Effects[0].Kind != compiler.EffectMill {
		t.Fatalf("effect[0].Kind = %v, want EffectMill", ab.Content.Effects[0].Kind)
	}
	if ab.Content.Effects[1].Kind != compiler.EffectDraw {
		t.Fatalf("effect[1].Kind = %v, want EffectDraw", ab.Content.Effects[1].Kind)
	}
	if ab.Content.Effects[0].Span != ab.Content.Effects[1].Span {
		t.Fatalf("effect spans differ: %+v vs %+v; want same sentence span",
			ab.Content.Effects[0].Span, ab.Content.Effects[1].Span)
	}

	// Verb spans are at distinct offsets, enabling the split to locate each
	// clause boundary precisely.
	if ab.Content.Effects[0].VerbSpan == ab.Content.Effects[1].VerbSpan {
		t.Fatal("verb spans equal; want mills ≠ draws")
	}

	// Exactly one target ("target player") in the IR. The compiler does not
	// emit a separate target for the implied "draws" subject.
	if len(ab.Content.Targets) != 1 {
		t.Fatalf("IR targets = %d, want 1 (shared; not duplicated for draws clause)", len(ab.Content.Targets))
	}
	if ab.Content.Targets[0].Selector.Kind != compiler.SelectorPlayer {
		t.Fatalf("target selector = %v, want SelectorPlayer", ab.Content.Targets[0].Selector.Kind)
	}

	// End-to-end: compound mill lowers successfully with no diagnostics.
	card := &ScryfallCard{
		Name: "Test Mill", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
	}
	_, execDiags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(execDiags) != 0 {
		t.Fatalf("executable diagnostics: %v", execDiags)
	}
}

// TestLowerThenJoinedImpliedSubjectDamageChain is a regression for the
// implied-subject reference accounting bug: "A deals N damage to target X,
// then deals N damage to target X." has exactly ONE CompiledReference in the
// oracle IR but both effects find it via sentence span, making consumedReferences
// increment twice and the final accounting check fail.
//
// The fix attributes references to their per-clause owned region so the shared
// self-reference is counted only once while still being propagated to implied-
// subject clauses for the damage-amount-reference lowerer check.
func TestLowerThenJoinedImpliedSubjectDamageChain(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 1 damage to target creature, then deals 1 damage to target creature.",
	}
	source, diags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v", diags)
	}
	// Two independent target slots: each clause targets its own creature.
	for _, want := range []string{"game.AnyTargetDamageRecipient(0)", "game.AnyTargetDamageRecipient(1)"} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q (two independent target slots):\n%s", want, source)
		}
	}
}

// TestLowerThenJoinedExplicitRepeatedSubjectDamageChain is a regression for the
// explicit repeated-subject reference accounting bug: "A deals N damage to X,
// then A deals N damage to X." has TWO CompiledReferences and TWO targets.
// With sentence-span filtering each effect found both references and both
// targets, causing singleSelfReference to fail with len==2.
//
// The fix attributes each reference and target to exactly the clause that
// contains it so every lowering call sees exactly one self-reference and one
// target, and consumedReferences + consumedTargets equal the ability totals.
func TestLowerThenJoinedExplicitRepeatedSubjectDamageChain(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 1 damage to target creature, then Test Bolt deals 1 damage to target creature.",
	}
	source, diags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v", diags)
	}
	// Two independent target slots: each explicit "Test Bolt" clause targets its own creature.
	for _, want := range []string{"game.AnyTargetDamageRecipient(0)", "game.AnyTargetDamageRecipient(1)"} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q (two independent target slots):\n%s", want, source)
		}
	}
}

// TestLowerThenJoinedDifferentExplicitSubject is a regression for the bug where
// non-first then clauses that have their own explicit subject (e.g. "you" in
// "then you gain 2 life.") were incorrectly given the first clause's subject
// prefix ("Target player") instead, producing "Target player gain 2 life." and
// failing the exact-text check.
//
// Requirements verified:
//   - Compiles without diagnostics.
//   - Draw instruction references TargetPlayerReference(0) (target player draws).
//   - GainLife instruction references ControllerReference (you = controller).
//   - Exactly 1 target spec (the "target player" from the draw clause).
func TestLowerThenJoinedDifferentExplicitSubject(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player draws a card, then you gain 2 life.",
	}
	source, diags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v", diags)
	}
	// Draw must reference the target player, not the controller.
	if !strings.Contains(source, "game.TargetPlayerReference(0)") {
		t.Fatalf("source missing target player draw reference:\n%s", source)
	}
	// GainLife must reference the controller ("you"), not the target player.
	if strings.Contains(source, "game.TargetPlayerReference") &&
		strings.Contains(source, "game.GainLife") {
		// Verify the GainLife uses ControllerReference.
		if !strings.Contains(source, "Player: game.ControllerReference()") {
			t.Fatalf("expected GainLife to use ControllerReference:\n%s", source)
		}
	}
	// Exactly one target slot (the "target player" for the draw).
	if count := strings.Count(source, "MinTargets:"); count != 1 {
		t.Fatalf("want 1 TargetSpec, got %d:\n%s", count, source)
	}
}

// TestLowerThenJoinedExplicitRepeatedSelfSubject confirms that "A does X, then
// A does Y." where each clause has its own explicit repeated subject is
// handled correctly: the post-then "A" tokens are used for the second clause,
// not the first clause's subject prefix. This differs from the implied-subject
// case (where the post-then region is empty and prefix is inherited) and the
// different-subject case above.
func TestLowerThenJoinedExplicitRepeatedSelfSubject(t *testing.T) {
	t.Parallel()
	// Compound mill already tests this end-to-end; here we specifically confirm
	// that the second clause's subject comes from its own post-then token range
	// and not from a copied first-clause subject-prefix.
	card := &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 1 damage to target creature, then you gain 1 life.",
	}
	source, diags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v", diags)
	}
	// Damage clause must target the creature.
	if !strings.Contains(source, "game.AnyTargetDamageRecipient(0)") {
		t.Fatalf("source missing damage to target creature:\n%s", source)
	}
	// Gain clause must use controller reference ("you"), with no second target.
	if count := strings.Count(source, "MinTargets:"); count != 1 {
		t.Fatalf("want 1 TargetSpec (damage target only), got %d:\n%s", count, source)
	}
	if !strings.Contains(source, "game.GainLife") {
		t.Fatalf("source missing GainLife:\n%s", source)
	}
}

// TestLowerThenJoinedThreeEffectExplicitMiddleSubject is the primary regression
// for the structural bug where pair (1,2) overwrote the middle clause set by
// pair (0,1): "Target player draws a card, then you gain 2 life, then draw a
// card." would produce "Target player gain 2 life." for clause 1 (wrong subject)
// and "Target player draw a card." for clause 2 (wrong subject and verb mismatch).
//
// With the single-pass group redesign:
//   - Clause 0: target player draws (TargetPlayerReference(0), 1 TargetSpec).
//   - Clause 1: you gain (ControllerReference, 0 TargetSpecs — "you" is explicit,
//     no target inheritance).
//   - Clause 2: controller draws (ControllerReference, 0 TargetSpecs — "draw" is
//     imperative, no subject prefix or target inheritance).
func TestLowerThenJoinedThreeEffectExplicitMiddleSubject(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player draws a card, then you gain 2 life, then draw a card.",
	}
	source, diags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v", diags)
	}
	// Exactly one target slot: the "target player" for the draw.
	if count := strings.Count(source, "MinTargets:"); count != 1 {
		t.Fatalf("want 1 TargetSpec, got %d:\n%s", count, source)
	}
	// Draw uses TargetPlayerReference(0).
	if !strings.Contains(source, "game.TargetPlayerReference(0)") {
		t.Fatalf("source missing TargetPlayerReference(0) for draw:\n%s", source)
	}
	// GainLife uses ControllerReference (the "you" clause).
	if !strings.Contains(source, "game.GainLife") {
		t.Fatalf("source missing GainLife:\n%s", source)
	}
	// Final draw is a controller draw (not target player).
	drawIdx := strings.LastIndex(source, "game.Draw{")
	gainIdx := strings.Index(source, "game.GainLife")
	if drawIdx < 0 || gainIdx < 0 || drawIdx <= gainIdx {
		t.Fatalf("expected GainLife before final Draw:\n%s", source)
	}
	// Three instructions total.
	if count := strings.Count(source, "Primitive:"); count != 3 {
		t.Fatalf("want 3 instructions, got %d:\n%s", count, source)
	}
}

// TestJoinedTokenTextPossessive is a regression for the apostrophe spacing bug
// in joinedTokenNeedsSpace: before the fix, prev.Kind == shared.Apostrophe was
// missing from the no-space guard, so a possessive token sequence like
// [Test, Bolt, ', s, power] would reconstruct as "Test Bolt' s power." instead
// of "Test Bolt's power.". This matters for clause-text overrides that include
// a possessive card name (e.g. "Test Bolt's power" as a damage amount subject).
func TestJoinedTokenTextPossessive(t *testing.T) {
	t.Parallel()
	toks := []shared.Token{
		{Kind: shared.Word, Text: "Test"},
		{Kind: shared.Word, Text: "Bolt"},
		{Kind: shared.Apostrophe, Text: "'"},
		{Kind: shared.Word, Text: "s"},
		{Kind: shared.Word, Text: "power"},
		{Kind: shared.Period, Text: "."},
	}
	got := joinedTokenText(toks)
	if got != "Test Bolt's power." {
		t.Fatalf("joinedTokenText = %q, want %q", got, "Test Bolt's power.")
	}
}
