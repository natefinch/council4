package parser

import "testing"

// TestParseStartEnginesKeyword verifies that "Start your engines!" parses as the
// KeywordStartEngines keyword ability, including with its reminder text.
func TestParseStartEnginesKeyword(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Start your engines!",
		"Start your engines! (If you have no speed, it starts at 1. It increases " +
			"once on each of your turns when an opponent loses life. Max speed is 4.)",
	} {
		keywords := keywordsFor(t, source)
		if len(keywords) != 1 {
			t.Fatalf("%q keywords = %+v; want exactly one", source, keywords)
		}
		if keywords[0].Kind != KeywordStartEngines {
			t.Fatalf("%q kind = %v; want %v", source, keywords[0].Kind, KeywordStartEngines)
		}
	}
}

// TestParseYourSpeedDynamicAmount verifies that "where X is your speed" parses
// the shared variable amount to the controller-speed dynamic kind.
func TestParseYourSpeedDynamicAmount(t *testing.T) {
	t.Parallel()
	source := "You draw X cards and lose X life, where X is your speed."
	document, _ := Parse(source, Context{InstantOrSorcery: true})
	var lose *EffectSyntax
	for si := range document.Abilities[0].Sentences {
		sentence := &document.Abilities[0].Sentences[si]
		for ei := range sentence.Effects {
			if sentence.Effects[ei].Kind == EffectLose {
				lose = &sentence.Effects[ei]
			}
		}
	}
	if lose == nil {
		t.Fatalf("no lose effect parsed from %q", source)
	}
	if lose.Amount.DynamicKind != EffectDynamicAmountControllerSpeed {
		t.Fatalf("lose dynamic kind = %v, want %v",
			lose.Amount.DynamicKind, EffectDynamicAmountControllerSpeed)
	}
}
