package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func groupEntersBecomesEffect(t *testing.T, name, text string) EffectSyntax {
	t.Helper()
	doc, _ := Parse(text, Context{CardName: name})
	for a := range doc.Abilities {
		ability := &doc.Abilities[a]
		if ability.Kind != AbilityReplacement {
			continue
		}
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				effect := ability.Sentences[s].Effects[e]
				if effect.EntersBecomesGroup() {
					return effect
				}
			}
		}
	}
	t.Fatalf("no group enters-becomes effect parsed for %q", text)
	return EffectSyntax{}
}

// TestParseGroupEntersBecomesDisplacedDinosaurs proves the reusable group ETB
// characteristic replacement parses Displaced Dinosaurs into a
// GroupEntryModificationBecomes carrying the historic filter, the you-control
// scope, the added Creature type and Dinosaur subtype, and the 7/7 base P/T.
func TestParseGroupEntersBecomesDisplacedDinosaurs(t *testing.T) {
	effect := groupEntersBecomesEffect(t, "Displaced Dinosaurs",
		"As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types.")
	mod := effect.GroupEntryModification
	if mod.Kind != GroupEntryModificationBecomes {
		t.Fatalf("modification kind = %v, want GroupEntryModificationBecomes", mod.Kind)
	}
	if !mod.Historic {
		t.Error("modification is not historic, want historic")
	}
	if mod.ControllerScope != EntersTappedGroupControllerYou {
		t.Errorf("controller scope = %v, want you", mod.ControllerScope)
	}
	if len(mod.AddTypes) != 1 || mod.AddTypes[0] != types.Creature {
		t.Errorf("add types = %v, want [Creature]", mod.AddTypes)
	}
	if len(mod.AddSubtypes) != 1 || mod.AddSubtypes[0] != types.Dinosaur {
		t.Errorf("add subtypes = %v, want [Dinosaur]", mod.AddSubtypes)
	}
	if !mod.BasePower.Exists || mod.BasePower.Val != 7 {
		t.Errorf("base power = %v, want 7", mod.BasePower)
	}
	if !mod.BaseToughness.Exists || mod.BaseToughness.Val != 7 {
		t.Errorf("base toughness = %v, want 7", mod.BaseToughness)
	}
	if len(mod.Colors) != 0 {
		t.Errorf("colors = %v, want none", mod.Colors)
	}
}

// TestParseGroupEntersBecomesRejectsNonReplacementShapes proves the recognizer
// does not fire for sentences that are not "As <subject> enters, it becomes ..."
// characteristic replacements.
func TestParseGroupEntersBecomesRejectsNonReplacementShapes(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{
			name: "targeted becomes is not a group ETB replacement",
			text: "Target permanent becomes a 7/7 Dinosaur creature in addition to its other types.",
		},
		{
			name: "enters-tapped group is not a becomes",
			text: "Artifacts you control enter tapped.",
		},
		{
			name: "triggered pump is not a becomes",
			text: "Whenever a historic permanent you control enters, put a +1/+1 counter on it.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			doc, _ := Parse(test.text, Context{CardName: "Test"})
			for a := range doc.Abilities {
				ability := &doc.Abilities[a]
				for s := range ability.Sentences {
					for e := range ability.Sentences[s].Effects {
						if ability.Sentences[s].Effects[e].EntersBecomesGroup() {
							t.Fatalf("unexpected group enters-becomes effect for %q", test.text)
						}
					}
				}
			}
		})
	}
}
