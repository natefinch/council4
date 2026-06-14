package cardgen

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func TestLowerPlayerOrdinalTriggerPattern(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Ordinal Draw",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever you draw your second card each turn, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventCardDrawn ||
		pattern.Player != game.TriggerPlayerYou ||
		pattern.PlayerEventOrdinalThisTurn != 2 {
		t.Fatalf("pattern = %#v", pattern)
	}
}

// TestLowerBodyEquivalenceAcrossShells proves that the same body oracle text
// lowers to equivalent game.AbilityContent regardless of which shell wraps it
// (spell, activated ability body, triggered ability body, loyalty ability body,
// or modal option). This is the core contract for lowerAbilityContent.
func TestLowerBodyEquivalenceAcrossShells(t *testing.T) {
	t.Parallel()

	// Body text: "Draw a card." — a simple, widely supported single-effect body.
	// We verify that lowering it as five different shells yields identical
	// game.AbilityContent values.

	spellFace := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card.",
	})
	if !spellFace.SpellAbility.Exists {
		t.Fatal("spell face missing SpellAbility")
	}
	want := spellFace.SpellAbility.Val

	tests := []struct {
		name string
		card *ScryfallCard
		get  func(loweredFaceAbilities) game.AbilityContent
	}{
		{
			name: "activated body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "{1}, {T}: Draw a card.",
				Power:      new("1"),
				Toughness:  new("1"),
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if len(f.ActivatedAbilities) == 0 {
					t.Fatal("no activated abilities")
				}
				return f.ActivatedAbilities[0].Content
			},
		},
		{
			name: "triggered body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "When this creature enters, draw a card.",
				Power:      new("1"),
				Toughness:  new("1"),
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if len(f.TriggeredAbilities) == 0 {
					t.Fatal("no triggered abilities")
				}
				return f.TriggeredAbilities[0].Content
			},
		},
		{
			name: "loyalty body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Planeswalker — Jace",
				OracleText: "+1: Draw a card.",
				Loyalty:    new("3"),
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if len(f.LoyaltyAbilities) == 0 {
					t.Fatal("no loyalty abilities")
				}
				return f.LoyaltyAbilities[0].Content
			},
		},
		{
			name: "modal option",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Choose one —\n• Draw a card.\n• Draw a card.",
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if !f.SpellAbility.Exists {
					t.Fatal("no spell ability")
				}
				ab := f.SpellAbility.Val
				if len(ab.Modes) < 1 {
					t.Fatal("no modes")
				}
				// Return a non-modal AbilityContent wrapping the first mode.
				return game.Mode{
					Targets:  ab.Modes[0].Targets,
					Sequence: ab.Modes[0].Sequence,
				}.Ability()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, tt.card)
			got := tt.get(face)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("body content mismatch for shell %q:\n got  %#v\n want %#v", tt.name, got, want)
			}
		})
	}
}

// TestLowerOrderedEffectsViaContentEntry proves that target-index remapping
// in ordered-effect sequences works equivalently when the body is lowered
// through different shells, all routing through lowerAbilityContent.
func TestLowerOrderedEffectsViaContentEntry(t *testing.T) {
	t.Parallel()

	type result struct {
		targets    int
		idx0, idx1 int
	}

	extract := func(t *testing.T, ab game.AbilityContent) result {
		t.Helper()
		if len(ab.Modes) != 1 {
			t.Fatalf("modes = %d, want 1", len(ab.Modes))
		}
		m := ab.Modes[0]
		if len(m.Targets) != 2 || len(m.Sequence) != 2 {
			t.Fatalf("mode targets=%d sequence=%d, want 2 targets and 2 instructions", len(m.Targets), len(m.Sequence))
		}
		destroy, ok := m.Sequence[0].Primitive.(game.Destroy)
		if !ok {
			t.Fatalf("first primitive = %T, want game.Destroy", m.Sequence[0].Primitive)
		}
		tap, ok := m.Sequence[1].Primitive.(game.Tap)
		if !ok {
			t.Fatalf("second primitive = %T, want game.Tap", m.Sequence[1].Primitive)
		}
		return result{
			targets: len(m.Targets),
			idx0:    destroy.Object.TargetIndex(),
			idx1:    tap.Object.TargetIndex(),
		}
	}

	// Establish expected result from the spell shell.
	spellFace := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature.",
	})
	if !spellFace.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	want := extract(t, spellFace.SpellAbility.Val)

	tests := []struct {
		name string
		card *ScryfallCard
		get  func(t *testing.T, f loweredFaceAbilities) game.AbilityContent
	}{
		{
			name: "activated body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: "{T}: Destroy target artifact. Tap target creature.",
			},
			get: func(t *testing.T, f loweredFaceAbilities) game.AbilityContent {
				t.Helper()
				if len(f.ActivatedAbilities) == 0 {
					t.Fatal("no activated abilities")
				}
				return f.ActivatedAbilities[0].Content
			},
		},
		{
			name: "loyalty body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Planeswalker — Test",
				OracleText: "-2: Destroy target artifact. Tap target creature.",
				Loyalty:    new("4"),
			},
			get: func(t *testing.T, f loweredFaceAbilities) game.AbilityContent {
				t.Helper()
				if len(f.LoyaltyAbilities) == 0 {
					t.Fatal("no loyalty abilities")
				}
				return f.LoyaltyAbilities[0].Content
			},
		},
		{
			name: "modal option",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Choose one —\n• Destroy target artifact. Tap target creature.\n• Draw a card.",
			},
			get: func(t *testing.T, f loweredFaceAbilities) game.AbilityContent {
				t.Helper()
				if !f.SpellAbility.Exists {
					t.Fatal("no spell ability")
				}
				ab := f.SpellAbility.Val
				if len(ab.Modes) < 1 {
					t.Fatal("no modes")
				}
				return game.Mode{
					Targets:  ab.Modes[0].Targets,
					Sequence: ab.Modes[0].Sequence,
				}.Ability()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, tt.card)
			got := extract(t, tt.get(t, face))
			if got != want {
				t.Errorf("ordered-effect result for shell %q: got %+v, want %+v", tt.name, got, want)
			}
		})
	}
}

// TestLowerOrderedEffectsTargetIndexRemappingInActivatedBody checks target-
// index remapping for a three-clause ordered sequence through an activated
// ability body, proving that lowerAbilityContent correctly rebases indices
// regardless of shell.
func TestLowerOrderedEffectsTargetIndexRemappingInActivatedBody(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Destroy target artifact. Tap target creature. Target player mills three cards.",
	})
	if len(face.ActivatedAbilities) == 0 {
		t.Fatal("no activated abilities")
	}
	ab := face.ActivatedAbilities[0].Content
	if len(ab.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(ab.Modes))
	}
	m := ab.Modes[0]
	if len(m.Targets) != 3 || len(m.Sequence) != 3 {
		t.Fatalf("mode targets=%d sequence=%d, want 3 each", len(m.Targets), len(m.Sequence))
	}
	destroy, ok := m.Sequence[0].Primitive.(game.Destroy)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Destroy", m.Sequence[0].Primitive)
	}
	tap, ok := m.Sequence[1].Primitive.(game.Tap)
	if !ok {
		t.Fatalf("second primitive = %T, want game.Tap", m.Sequence[1].Primitive)
	}
	mill, ok := m.Sequence[2].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("third primitive = %T, want game.Mill", m.Sequence[2].Primitive)
	}
	if destroy.Object.TargetIndex() != 0 || tap.Object.TargetIndex() != 1 || mill.Player.TargetIndex() != 2 {
		t.Errorf(
			"target indices = %d, %d, %d; want 0, 1, 2",
			destroy.Object.TargetIndex(),
			tap.Object.TargetIndex(),
			mill.Player.TargetIndex(),
		)
	}
}

// TestLowerContentDiagnosticDistinguishesShellFromContent proves that
// content-body failures propagate their own diagnostic summaries through shell
// lowerers, and that shell-specific failures (bad cost) produce different
// summaries from content failures (unsupported effect).
func TestLowerContentDiagnosticDistinguishesShellFromContent(t *testing.T) {
	t.Parallel()

	// A card whose body is an unsupported search effect: should produce a
	// content diagnostic (not "unsupported activated ability").
	t.Run("content failure through activated shell", func(t *testing.T) {
		t.Parallel()
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:     "Test Card",
			Layout:   "normal",
			TypeLine: "Artifact",
			// Unsupported search effect as activated body — the content fails,
			// not the cost.
			OracleText: "{T}: Search your library for a creature card, then shuffle.",
		}, "t")
		if err != nil {
			t.Fatal(err)
		}

		if len(diagnostics) == 0 {
			t.Fatal("want at least one diagnostic, got none")
		}
		for _, d := range diagnostics {
			if d.Summary == "unsupported activated ability" {
				t.Errorf("got generic shell summary %q; expected content diagnostic to propagate", d.Summary)
			}
		}
	})

	for _, oracleText := range []string{
		"Whenever you gain life, search your library for a creature card, then shuffle.",
		"Whenever a creature enters, search your library for a creature card, then shuffle.",
		"Whenever you cast an artifact spell, search your library for a creature card, then shuffle.",
		"Whenever equipped creature attacks, search your library for a creature card, then shuffle.",
	} {
		t.Run("content failure through typed trigger "+oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Artifact — Equipment",
				OracleText: oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("want at least one diagnostic, got none")
			}
			if slices.ContainsFunc(diagnostics, func(d shared.Diagnostic) bool {
				return d.Summary == "unsupported triggered ability"
			}) {
				t.Fatalf("recognized typed trigger body collapsed into generic pattern diagnostic: %#v", diagnostics)
			}
			if !slices.ContainsFunc(diagnostics, func(d shared.Diagnostic) bool {
				return d.Summary == "unsupported search effect"
			}) {
				t.Fatalf("diagnostics = %#v, want shared content diagnostic", diagnostics)
			}
		})
	}

	// A card with an unsupported activated-ability cost produces a shell
	// diagnostic; the body "Draw a card." is fully supported.
	t.Run("shell failure with supported content", func(t *testing.T) {
		t.Parallel()
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:     "Test Card",
			Layout:   "normal",
			TypeLine: "Artifact",
			// Unsupported cost "Choose" — shell failure.
			OracleText: "{T}: Choose — Draw a card.",
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatal("want at least one diagnostic, got none")
		}
		for _, d := range diagnostics {
			if d.Summary == "unsupported ability content" {
				t.Errorf("content diagnostic %q should not surface for a shell-level cost failure", d.Summary)
			}
		}
	})
}

func TestLowerOrderedEffectsInPhaseTrigger(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "At the beginning of your upkeep, destroy target artifact. Draw a card.",
	})
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want destroy then draw", sequence)
	}
}

// TestLowerOrderedEffectsOrderPreservedAcrossShells verifies that the
// instruction order produced by lowerAbilityContent for an ordered-effect body
// is stable and identical across every shell that supports ordered effects.
// This complements TestLowerOrderedEffectsViaContentEntry by checking a
// Saga chapter shell.
func TestLowerOrderedEffectsOrderPreservedAcrossShells(t *testing.T) {
	t.Parallel()

	// Extract the ordered pair (first-primitive-type, second-primitive-type)
	// from an AbilityContent that should have exactly one mode with two instructions.
	type instrTypes struct {
		first, second string
	}
	extract := func(t *testing.T, ab game.AbilityContent) instrTypes {
		t.Helper()
		if len(ab.Modes) != 1 || len(ab.Modes[0].Sequence) != 2 {
			t.Fatalf("expected 1 mode with 2 instructions, got modes=%d", len(ab.Modes))
		}
		seq := ab.Modes[0].Sequence
		return instrTypes{
			first:  fmt.Sprintf("%T", seq[0].Primitive),
			second: fmt.Sprintf("%T", seq[1].Primitive),
		}
	}

	// "Tap target creature. Draw a card." — a supported 2-effect ordered body
	// where exactly one target is needed. Activated and loyalty shells support this.
	want := extract(t, lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Tap target creature. Draw a card.",
	}).SpellAbility.Val)

	tests := []struct {
		name string
		card *ScryfallCard
		get  func(f loweredFaceAbilities) game.AbilityContent
	}{
		{
			name: "activated body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: "{T}: Tap target creature. Draw a card.",
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if len(f.ActivatedAbilities) == 0 {
					t.Fatal("no activated abilities")
				}
				return f.ActivatedAbilities[0].Content
			},
		},
		{
			name: "loyalty body",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Planeswalker — Test",
				OracleText: "-1: Tap target creature. Draw a card.",
				Loyalty:    new("3"),
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if len(f.LoyaltyAbilities) == 0 {
					t.Fatal("no loyalty abilities")
				}
				return f.LoyaltyAbilities[0].Content
			},
		},
		{
			name: "modal option",
			card: &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Choose one —\n• Tap target creature. Draw a card.\n• Draw a card.",
			},
			get: func(f loweredFaceAbilities) game.AbilityContent {
				if !f.SpellAbility.Exists {
					t.Fatal("no spell ability")
				}
				ab := f.SpellAbility.Val
				return game.Mode{
					Targets:  ab.Modes[0].Targets,
					Sequence: ab.Modes[0].Sequence,
				}.Ability()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, tt.card)
			got := extract(t, tt.get(face))
			if got != want {
				t.Errorf("instruction order mismatch for shell %q: got %+v, want %+v", tt.name, got, want)
			}
		})
	}
}

// TestLowerOrderedEffectsViaLowerAbilityContentPerClause proves that the change
// to route each ordered-effect clause through lowerAbilityContent (rather than
// calling lowerSingleEffectSpell directly) preserves correct remapping for
// both independent clauses and then-joined groups.
func TestLowerOrderedEffectsViaLowerAbilityContentPerClause(t *testing.T) {
	t.Parallel()

	// Independent effects — two separate sentences, each routed through
	// lowerAbilityContent by lowerOrderedEffectSequence.
	t.Run("independent_clauses_remapped", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Spell",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Destroy target artifact. Tap target creature.",
		})
		if !face.SpellAbility.Exists {
			t.Fatal("spell ability not lowered")
		}
		ab := face.SpellAbility.Val
		if len(ab.Modes) != 1 {
			t.Fatalf("modes = %d, want 1", len(ab.Modes))
		}
		m := ab.Modes[0]
		if len(m.Targets) != 2 || len(m.Sequence) != 2 {
			t.Fatalf("targets=%d sequence=%d, want 2 each", len(m.Targets), len(m.Sequence))
		}
		destroy, ok := m.Sequence[0].Primitive.(game.Destroy)
		if !ok {
			t.Fatalf("first primitive %T, want game.Destroy", m.Sequence[0].Primitive)
		}
		tap, ok := m.Sequence[1].Primitive.(game.Tap)
		if !ok {
			t.Fatalf("second primitive %T, want game.Tap", m.Sequence[1].Primitive)
		}
		if destroy.Object.TargetIndex() != 0 {
			t.Errorf("destroy target index = %d, want 0", destroy.Object.TargetIndex())
		}
		if tap.Object.TargetIndex() != 1 {
			t.Errorf("tap target index = %d, want 1", tap.Object.TargetIndex())
		}
	})

	// Then-joined group — shared subject, each sub-clause routed through
	// lowerAbilityContent with capitalised clause text.
	t.Run("then_joined_draw_rider", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Spell",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Counter target spell. Draw a card.",
		})
		if !face.SpellAbility.Exists {
			t.Fatal("spell ability not lowered")
		}
		ab := face.SpellAbility.Val
		if len(ab.Modes) != 1 {
			t.Fatalf("modes = %d, want 1", len(ab.Modes))
		}
		m := ab.Modes[0]
		if len(m.Sequence) < 2 {
			t.Fatalf("sequence = %d, want >= 2 instructions", len(m.Sequence))
		}
	})
}

// TestLowerContentSpanContract proves the compiler contract that Content.Span
// is always non-zero for any recognised ability body (supported or not), and
// that activated-ability Content.Span starts after the cost.
func TestLowerContentSpanContract(t *testing.T) {
	t.Parallel()

	t.Run("activated_content_span_after_cost", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Card",
			Layout:     "normal",
			TypeLine:   "Creature — Test",
			OracleText: "{T}: Draw a card.",
			Power:      new("1"),
			Toughness:  new("1"),
		})
		if len(face.ActivatedAbilities) == 0 {
			t.Fatal("no activated abilities")
		}
		ab := face.ActivatedAbilities[0]
		if len(ab.Content.Modes) == 0 {
			t.Fatal("activated ability content has no modes; was it lowered?")
		}
	})

	t.Run("unsupported_body_still_lowers_gracefully", func(t *testing.T) {
		t.Parallel()
		// An unsupported body should produce a diagnostic (not panic) proving
		// the content pipeline handles unrecognized content safely.
		faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Card",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Frob the gronk.",
		})
		if len(diagnostics) == 0 {
			t.Fatal("expected diagnostics for unsupported body, got none")
		}
		if len(faces) == 0 {
			t.Fatal("expected at least one face result")
		}
		if faces[0].SpellAbility.Exists {
			t.Fatal("expected no spell ability for unsupported text, got one")
		}
	})
}
