package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCombatStunPossessiveRequiresOwnedRelatedSubject(t *testing.T) {
	t.Parallel()
	compile := func() AbilityContent {
		document, diagnostics := parser.Parse("Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step.", parser.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := Compile(document, Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return compilation.Abilities[0].Content
	}
	content := compile()
	var subject, possessive CompiledReference
	for _, reference := range content.References {
		if reference.Kind == ReferenceThatObject {
			subject = reference
		}
		if reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounIts {
			possessive = reference
		}
	}
	if !combatStunPossessiveBindsRelated(possessive, []CompiledReference{subject}, content.Effects) {
		t.Fatal("qualifying clause-owned subject refused")
	}
	for _, invalid := range []string{"missing", "ambiguous", "different subject", "foreign"} {
		t.Run(invalid, func(t *testing.T) {
			t.Parallel()
			prior := []CompiledReference{subject}
			candidate := possessive
			switch invalid {
			case "missing":
				prior = nil
			case "ambiguous":
				prior[0].Binding = ReferenceBindingAmbiguous
			case "different subject":
				prior[0].NodeID++
			case "foreign":
				for _, reference := range compile().References {
					if reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounIts {
						candidate = reference
					}
				}
			default:
				t.Fatal("unknown invalid subject")
			}
			if combatStunPossessiveBindsRelated(candidate, prior, content.Effects) {
				t.Fatal("unowned related possessive admitted")
			}
		})
	}
}
