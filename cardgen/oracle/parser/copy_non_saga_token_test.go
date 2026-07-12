package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseCopyTargetNonSagaToken(t *testing.T) {
	t.Parallel()
	source := "Create a token that's a copy of target non-Saga token you control."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	effect := document.Abilities[0].Sentences[0].Effects[0]
	if len(effect.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(effect.Targets))
	}
	target := effect.Targets[0]
	if !effect.Exact || !effect.TokenCopyOfTarget {
		t.Fatalf("effect exact = %v, copy target = %v, target exact = %v, selection = %#v",
			effect.Exact, effect.TokenCopyOfTarget, target.Exact, target.Selection)
	}
	if !target.Exact ||
		!target.Selection.TokenOnly ||
		len(target.Selection.ExcludedSubtypes) != 1 ||
		target.Selection.ExcludedSubtypes[0] != types.Saga {
		t.Fatalf("target = %#v, want non-Saga token", target)
	}
}

func TestParseCopyTargetNonSagaTokenSagaChapter(t *testing.T) {
	t.Parallel()
	source := "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after VI.)\n" +
		"I — Create a Treasure token.\n" +
		"II, III, IV, V, VI — Create a token that's a copy of target non-Saga token you control."
	document, diagnostics := Parse(source, Context{Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if len(effect.Targets) == 1 && effect.Targets[0].Selection.TokenOnly {
					if !effect.Exact || !effect.TokenCopyOfTarget || !effect.Targets[0].Exact {
						t.Fatalf("effect exact = %v, copy target = %v, target = %#v",
							effect.Exact, effect.TokenCopyOfTarget, effect.Targets[0])
					}
					return
				}
			}
		}
	}
	t.Fatal("copy-target chapter effect not found")
}
