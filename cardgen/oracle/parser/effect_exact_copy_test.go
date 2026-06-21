package parser

import "testing"

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
