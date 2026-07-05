package parser

import "testing"

// TestRecognizeEachPlayerChooseDestroySequence verifies the "Starting with you,
// each player may choose <permanent>. Destroy each permanent chosen this way."
// construct (Druid of Purification) folds onto ability.EachPlayerChooseDestroy
// with the shared candidate pool typed from the choose sentence and both
// sentences' effects shed.
func TestRecognizeEachPlayerChooseDestroySequence(t *testing.T) {
	src := "When this creature enters, starting with you, each player may choose an artifact or enchantment you don't control. Destroy each permanent chosen this way."
	document, diagnostics := Parse(src, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	clause := document.Abilities[0].EachPlayerChooseDestroy
	if clause == nil {
		t.Fatal("EachPlayerChooseDestroy clause = nil, want recognized construct")
	}
	if !clause.Optional {
		t.Fatal("clause.Optional = false, want true for the \"may\" wording")
	}
	if clause.Pool.Controller != SelectionControllerNotYou {
		t.Fatalf("pool controller = %v, want SelectionControllerNotYou", clause.Pool.Controller)
	}
	wantTypes := []string{"CardTypeArtifact", "CardTypeEnchantment"}
	if len(clause.Pool.RequiredTypesAny) != len(wantTypes) {
		t.Fatalf("pool types = %v, want artifact and enchantment", clause.Pool.RequiredTypesAny)
	}
	for i, want := range wantTypes {
		if string(clause.Pool.RequiredTypesAny[i]) != want {
			t.Fatalf("pool types = %v, want %v", clause.Pool.RequiredTypesAny, wantTypes)
		}
	}
	for si := range document.Abilities[0].Sentences {
		if len(document.Abilities[0].Sentences[si].Effects) != 0 {
			t.Fatalf("sentence %d retained effects after recognition", si)
		}
	}
}

// TestRecognizeEachPlayerChooseDestroyFailsClosed verifies the recognizer leaves
// abilities untouched when the two-sentence shape is not an exact match, so
// neighbouring constructs keep their ordinary lowering.
func TestRecognizeEachPlayerChooseDestroyFailsClosed(t *testing.T) {
	cases := map[string]string{
		"missing destroy half":   "When this creature enters, starting with you, each player may choose an artifact or enchantment you don't control.",
		"non-may choose":         "Starting with you, each player chooses an artifact or enchantment you don't control. Destroy each permanent chosen this way.",
		"different destroy verb": "Starting with you, each player may choose an artifact or enchantment you don't control. Exile each permanent chosen this way.",
		"not starting with you":  "Each player may choose an artifact or enchantment you don't control. Destroy each permanent chosen this way.",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			document, _ := Parse(src, Context{})
			for ai := range document.Abilities {
				if document.Abilities[ai].EachPlayerChooseDestroy != nil {
					t.Fatalf("ability %d recognized an each-player-choose-destroy construct, want none", ai)
				}
			}
		})
	}
}
