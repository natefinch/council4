package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// Every proof issued for the complete retained originals keeps its domain's
// permitted operations: abstract policy, player-group, and card domains never
// acquire live permanent mutation authority.
func TestReferenceRoleCohortProofDomainsRestrictUse(t *testing.T) {
	forbidden := map[compiler.ReferenceSubjectDomain][]compiler.ReferenceUse{
		compiler.ReferenceSubjectPolicy: {compiler.ReferenceUseMutation, compiler.ReferenceUseCharacteristic,
			compiler.ReferenceUseAttribution, compiler.ReferenceUseCardAction},
		compiler.ReferenceSubjectPlayerGroup: {compiler.ReferenceUseMutation, compiler.ReferenceUseAttribution,
			compiler.ReferenceUseCardAction},
		compiler.ReferenceSubjectCard: {compiler.ReferenceUseMutation},
	}
	checked := 0
	var walk func(*testing.T, compiler.AbilityContent)
	walk = func(t *testing.T, content compiler.AbilityContent) {
		for _, reference := range content.References {
			if !reference.SubjectSupported() {
				continue
			}
			checked++
			if !content.OwnsSubject(reference) {
				t.Errorf("proof is not owned by its body: %+v", reference)
			}
			for _, use := range forbidden[reference.SubjectDomain()] {
				if reference.SupportsUse(use) {
					t.Errorf("%s: domain %v gained use %v", reference.Text, reference.SubjectDomain(), use)
				}
			}
		}
		for _, mode := range content.Modes {
			walk(t, mode.Content)
		}
	}
	for name, card := range referenceRoleOriginals(t) {
		t.Run(name, func(t *testing.T) {
			faces := []struct{ name, text, typeLine string }{{card.Name, card.OracleText, card.TypeLine}}
			if len(card.CardFaces) > 0 {
				faces = faces[:0]
				for _, face := range card.CardFaces {
					faces = append(faces, struct{ name, text, typeLine string }{face.Name, face.OracleText, face.TypeLine})
				}
			}
			for _, face := range faces {
				document, _ := parser.Parse(face.text, parser.Context{CardName: face.name,
					InstantOrSorcery: strings.Contains(face.typeLine, "Instant") || strings.Contains(face.typeLine, "Sorcery")})
				compilation, _ := compiler.Compile(document, compiler.Context{})
				for _, ability := range compilation.Abilities {
					walk(t, ability.Content)
				}
			}
		})
	}
	if checked == 0 {
		t.Fatal("cohort issued no supported proofs")
	}
}
