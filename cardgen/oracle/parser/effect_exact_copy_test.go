package parser

import (
	"slices"
	"testing"
)

// copyStackObjectAbility parses a single-ability copy source and returns its
// folded sentences so tests can assert the copy effect's exactness and the
// "choose new targets for the copy" rider folding.
func copyStackObjectAbility(t *testing.T, source string) []Sentence {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	return document.Abilities[0].Sentences
}

func TestParseCopyStackObjectExact(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Copy target triggered ability you control.",
		"Copy target activated ability you control.",
	}
	for _, source := range accepted {
		sentences := copyStackObjectAbility(t, source)
		if len(sentences) != 1 || len(sentences[0].Effects) != 1 {
			t.Fatalf("Parse(%q) shape = %#v", source, sentences)
		}
		effect := sentences[0].Effects[0]
		if effect.Kind != EffectCopyStackObject {
			t.Fatalf("Parse(%q) kind = %v, want EffectCopyStackObject", source, effect.Kind)
		}
		if !effect.Exact {
			t.Errorf("Parse(%q) exact = false, want true", source)
		}
		if effect.CopyMayChooseNewTargets {
			t.Errorf("Parse(%q) CopyMayChooseNewTargets = true, want false", source)
		}
	}
}

func TestParseCopyChooseNewTargetsRiderFolds(t *testing.T) {
	t.Parallel()
	source := "Copy target triggered ability you control. You may choose new targets for the copy."
	sentences := copyStackObjectAbility(t, source)
	if len(sentences) != 2 {
		t.Fatalf("Parse(%q) sentences = %d, want 2", source, len(sentences))
	}

	if len(sentences[0].Effects) != 1 {
		t.Fatalf("Parse(%q) copy sentence effects = %#v", source, sentences[0].Effects)
	}
	effect := sentences[0].Effects[0]
	if effect.Kind != EffectCopyStackObject || !effect.Exact {
		t.Fatalf("Parse(%q) copy effect = %#v", source, effect)
	}
	if !effect.CopyMayChooseNewTargets {
		t.Errorf("Parse(%q) CopyMayChooseNewTargets = false, want true", source)
	}
	if len(sentences[1].Effects) != 0 {
		t.Errorf("Parse(%q) rider sentence effects = %#v, want folded", source, sentences[1].Effects)
	}
	if !sentences[1].CopyChooseNewTargetsRider {
		t.Errorf("Parse(%q) rider sentence not credited", source)
	}
}

func TestParseDynamicCopyBatchChooseNewTargetsRiderFolds(t *testing.T) {
	t.Parallel()
	source := "Copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies."
	sentences := copyStackObjectAbility(t, source)
	if len(sentences) != 2 || len(sentences[0].Effects) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, sentences)
	}
	effect := sentences[0].Effects[0]
	if effect.Kind != EffectCopyStackObject ||
		!effect.Exact ||
		effect.Amount.DynamicKind != EffectDynamicAmountCommanderCastCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		!effect.CopyMayChooseNewTargets {
		t.Fatalf("Parse(%q) copy effect = %#v", source, effect)
	}
	if effect.RequiresOrderedLowering {
		t.Fatal("dynamic copy batch still requires ordered lowering after rider fold")
	}
	if effect.Selection.Kind != SelectionUnknown || effect.FromZone != 0 {
		t.Fatalf("dynamic count leaked into copy object selection: selection=%#v from=%v", effect.Selection, effect.FromZone)
	}
	if len(sentences[1].Effects) != 0 || !sentences[1].CopyChooseNewTargetsRider {
		t.Fatalf("retarget rider not folded: %#v", sentences[1])
	}
}

func TestParseCopyColorExceptionAndRetargetRider(t *testing.T) {
	t.Parallel()
	source := "Copy target instant or sorcery spell, except that the copy is red. You may choose new targets for the copy."
	sentences := copyStackObjectAbility(t, source)
	if len(sentences) != 2 || len(sentences[0].Effects) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, sentences)
	}
	effect := sentences[0].Effects[0]
	if effect.Kind != EffectCopyStackObject || !effect.Exact {
		t.Fatalf("Parse(%q) copy effect = %#v", source, effect)
	}
	if !slices.Equal(effect.CopySetColors, []Color{ColorRed}) {
		t.Fatalf("CopySetColors = %v, want [ColorRed]", effect.CopySetColors)
	}
	if effect.CopySetColorsRiderSpan == effect.Span {
		t.Fatal("color exception was not recorded as a distinct rider span")
	}
	if !effect.CopyMayChooseNewTargets || !sentences[1].CopyChooseNewTargetsRider {
		t.Fatalf("retarget rider not folded: effect=%#v sentence=%#v", effect, sentences[1])
	}
}

// TestParseCopyNounNotEffect guards that the "copy" noun in token-copy wording
// ("a copy of ...") is never misread as a copy effect verb.
func TestParseCopyNounNotEffect(t *testing.T) {
	t.Parallel()
	source := "Create a token that's a copy of target creature you control."
	sentences := copyStackObjectAbility(t, source)
	if len(sentences) != 1 || len(sentences[0].Effects) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, sentences)
	}
	if sentences[0].Effects[0].Kind == EffectCopyStackObject {
		t.Errorf("Parse(%q) misclassified copy noun as EffectCopyStackObject", source)
	}
}

// TestParseChangeTargetRetarget verifies the redirect wording "Change the
// target of target spell with a single target." parses to a single exact
// EffectChooseNewTargets with one clean spell target, and that the spurious
// "target" nouns in the sentence are reconciled away from the ability's target
// list so the redirect lowering sees exactly one target.
func TestParseChangeTargetRetarget(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Change the target of target spell with a single target.",
		"Change the targets of target spell with a single target.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		sentences := document.Abilities[0].Sentences
		if len(sentences) != 1 || len(sentences[0].Effects) != 1 {
			t.Fatalf("Parse(%q) shape = %#v", source, sentences)
		}
		effect := sentences[0].Effects[0]
		if effect.Kind != EffectChooseNewTargets || !effect.Exact {
			t.Fatalf("Parse(%q) effect = %#v", source, effect)
		}
		if len(effect.Targets) != 1 || effect.Targets[0].Selection.Kind != SelectionSpell {
			t.Fatalf("Parse(%q) effect targets = %#v", source, effect.Targets)
		}
		if len(sentences[0].Targets) != 1 || sentences[0].Targets[0].Selection.Kind != SelectionSpell {
			t.Fatalf("Parse(%q) sentence targets = %#v", source, sentences[0].Targets)
		}
	}
}

// TestParseChangeTargetRetargetGeneralizedForms verifies the redirect recognizer
// covers its full grammatical domain: an activated-ability selection (Reroute),
// an optional leading "You may" that rides the effect's Optional flag and a
// no-qualifier form without the trailing "with a single target" (Goblin
// Flectomancer). It also confirms the redirect-to-a-named-object form
// ("... to this creature", Muck Drubb) fails closed instead of producing a
// spurious EffectChooseNewTargets.
func TestParseChangeTargetRetargetGeneralizedForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		source       string
		wantKind     SelectionKind
		wantOptional bool
	}{
		{
			name:     "activated ability with qualifier",
			source:   "Change the target of target activated ability with a single target.",
			wantKind: SelectionActivatedAbility,
		},
		{
			name:         "optional no qualifier spell",
			source:       "You may change the targets of target instant or sorcery spell.",
			wantKind:     SelectionSpell,
			wantOptional: true,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("Parse(%q) diagnostics = %#v", test.source, diagnostics)
			}
			sentences := document.Abilities[0].Sentences
			if len(sentences) != 1 || len(sentences[0].Effects) != 1 {
				t.Fatalf("Parse(%q) shape = %#v", test.source, sentences)
			}
			effect := sentences[0].Effects[0]
			if effect.Kind != EffectChooseNewTargets || !effect.Exact {
				t.Fatalf("Parse(%q) effect = %#v", test.source, effect)
			}
			if effect.Optional != test.wantOptional {
				t.Fatalf("Parse(%q) optional = %v, want %v", test.source, effect.Optional, test.wantOptional)
			}
			if len(effect.Targets) != 1 || effect.Targets[0].Selection.Kind != test.wantKind {
				t.Fatalf("Parse(%q) effect targets = %#v, want kind %v", test.source, effect.Targets, test.wantKind)
			}
		})
	}
}

// TestParseChangeTargetRedirectToObjectRejected verifies the redirect-to-a-named
// object wording ("... to this creature", Muck Drubb) is not recognized as a
// free retarget, so it fails closed rather than lowering as EffectChooseNewTargets.
func TestParseChangeTargetRedirectToObjectRejected(t *testing.T) {
	t.Parallel()
	source := "Change the target of target spell that targets only a single creature to this creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if effect.Kind == EffectChooseNewTargets {
				t.Fatalf("Parse(%q) wrongly recognized redirect-to-object as EffectChooseNewTargets", source)
			}
		}
	}
}

// TestParseCopyTokenOneOfThem verifies the "create a token that's a copy of one
// of them." copy source (Twilight Diviner) is recognized as an exact
// copy-of-triggering-set create whose "them" pronoun names the triggering set.
func TestParseCopyTokenOneOfThem(t *testing.T) {
	t.Parallel()
	source := "Whenever one or more other creatures you control enter, create a token that's a copy of one of them."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	var effect *EffectSyntax
	for i := range document.Abilities[0].Sentences {
		for j := range document.Abilities[0].Sentences[i].Effects {
			if document.Abilities[0].Sentences[i].Effects[j].Kind == EffectCreate {
				effect = &document.Abilities[0].Sentences[i].Effects[j]
			}
		}
	}
	if effect == nil {
		t.Fatalf("Parse(%q) found no create effect", source)
	}
	if !effect.TokenCopyOfTriggeringSet {
		t.Errorf("Parse(%q) TokenCopyOfTriggeringSet = false, want true", source)
	}
	if !effect.Exact {
		t.Errorf("Parse(%q) exact = false, want true", source)
	}
	if effect.TokenCopyOfTarget || effect.TokenCopyOfReference || effect.TokenCopyOfAttached {
		t.Errorf("Parse(%q) misclassified copy source: %#v", source, effect)
	}
}
