package parser

import "testing"

// coordinatedPumpEffect parses a triggered ability whose effect is body and
// returns that single resolving effect for the "Alandra, Sky Dreamer" card
// context (legendary, self-name "Alandra").
func coordinatedPumpEffect(t *testing.T, body string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(
		"Whenever you draw your fifth card each turn, "+body,
		Context{CardName: "Alandra, Sky Dreamer", Legendary: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	return effects[0]
}

// TestParseCoordinatedSelfGroupPump proves the coordinated "<self> and <group>
// each get <p>/<t>" subject records the group as its source-EXCLUDING variant and
// flags CoordinatedSourceSubject, so lowering can pump the source separately from
// the excluding group. Both the card's own name and a "this creature" self
// reference introduce the coordination.
func TestParseCoordinatedSelfGroupPump(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		body    string
		kind    EffectStaticSubjectKind
		subtype string
	}{
		{
			name:    "self name and subtype group",
			body:    "Alandra and Drakes you control each get +X/+X until end of turn, where X is the number of cards in your hand.",
			kind:    EffectStaticSubjectOtherControlledCreatureSubtype,
			subtype: "Drake",
		},
		{
			name: "this creature and controlled creatures",
			body: "This creature and creatures you control each get +1/+1 until end of turn.",
			kind: EffectStaticSubjectOtherControlledCreatures,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := coordinatedPumpEffect(t, test.body)
			if effect.Kind != EffectModifyPT {
				t.Fatalf("kind = %v, want EffectModifyPT", effect.Kind)
			}
			if !effect.CoordinatedSourceSubject {
				t.Fatal("CoordinatedSourceSubject = false, want true")
			}
			if !effect.Exact {
				t.Fatal("Exact = false, want true")
			}
			if effect.StaticSubject.Kind != test.kind {
				t.Fatalf("StaticSubject.Kind = %v, want %v", effect.StaticSubject.Kind, test.kind)
			}
			if test.subtype != "" && string(effect.StaticSubject.Subtype) != test.subtype {
				t.Fatalf("StaticSubject.Subtype = %q, want %q", effect.StaticSubject.Subtype, test.subtype)
			}
		})
	}
}

// TestParseCoordinatedSelfGroupPumpFailsClosed proves near-miss subjects do not
// engage the coordinated recognizer: a non-self first conjunct, a missing "each"
// distributive, and a non-group second conjunct all leave CoordinatedSourceSubject
// unset so lowering never emits a spurious source pump or excludes the wrong
// group.
func TestParseCoordinatedSelfGroupPumpFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body string
	}{
		{
			name: "first conjunct is not the source",
			body: "Goblins and Drakes you control each get +1/+1 until end of turn.",
		},
		{
			name: "missing each distributive",
			body: "Alandra and Drakes you control get +1/+1 until end of turn.",
		},
		{
			name: "second conjunct is a target, not a group",
			body: "Alandra and target creature each get +1/+1 until end of turn.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"Whenever you draw your fifth card each turn, "+test.body,
				Context{CardName: "Alandra, Sky Dreamer", Legendary: true},
			)
			if len(document.Abilities) == 0 {
				return
			}
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if effect.CoordinatedSourceSubject {
						t.Fatalf("CoordinatedSourceSubject = true for %q, want false", test.body)
					}
				}
			}
		})
	}
}
