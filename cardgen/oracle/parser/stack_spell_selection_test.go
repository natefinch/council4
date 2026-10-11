package parser

import "testing"

func TestStackSpellSelectionGrammar(t *testing.T) {
	for _, tc := range []struct {
		text  string
		exact bool
	}{
		{"target enchantment spell", true},
		{"target battle spell", true},
		{"target planeswalker spell", true},
		{"target blue creature spell", true},
		{"target nonblue enchantment spell you don't control with mana value 3 or less", true},
		{"target colorless artifact spell an opponent controls", true},
		{"target multicolored noncreature spell with mana value 4 or greater", true},
		{"target noncreature enchantment spell", true},
		{"target instant or sorcery spell you control with mana value 2", true},
		{"target artifact, enchantment, or battle spell with mana value 3 or less", true},
		{"up to one target blue enchantment spell", true},
		{"two target blue enchantment spells you don't control with mana value 3 or less", true},
		{"any number of target artifact or enchantment spells an opponent controls", true},
		{"target blue enchantment or instant spell", false},
		{"target blue enchantment or red instant spell", false},
		{"target enchantment or blue instant spell", false},
		{"target noncreature artifact or enchantment spell", false},
		{"target artifact creature spell", false},
		{"target blue or red spell", false},
		{"target monocolored spell", false},
		{"target legendary enchantment spell", false},
		{"target Aura spell", false},
		{"target enchantment spell with flying", false},
		{"target enchantment spell without flying", false},
		{"target enchantment spell you own", false},
		{"target enchantment spell from your graveyard", false},
		{"target tapped enchantment spell", false},
		{"target enchantment spell with mana value X", false},
		{"target enchantment spell with mana value 3 or less you control", false},
		{"target enchantment spell with power 3", false},
		{"target enchantment spell named Test", false},
		{"target enchantment spell that was dealt damage this turn", false},
		{"two target blue enchantment spell", false},
		{"up to one target enchantment spells", false},
		{"target enchantment spell with an additional rider", false},
	} {
		t.Run(tc.text, func(t *testing.T) {
			doc, diagnostics := Parse("Counter "+tc.text+".", Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("parse diagnostics: %#v", diagnostics)
			}
			effects := doc.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one counter target", effects)
			}
			if got := effects[0].Targets[0].Exact; got != tc.exact {
				t.Fatalf("target exact = %v, want %v; target = %#v", got, tc.exact, effects[0].Targets[0])
			}
		})
	}
}
