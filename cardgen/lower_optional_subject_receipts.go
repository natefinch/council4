package cardgen

import "github.com/natefinch/council4/cardgen/oracle/compiler"

// Binding may demand the existing action owner's receipt. It does not elect a
// second owner or turn acceptance into an actual entered-object publication.
func (groups optionalActionGroups) requireEnteredSubjectReceipts(content compiler.AbilityContent) string {
	require := func(reference compiler.CompiledReference) string {
		if !reference.EnteredSubjectSupported() {
			return ""
		}
		index := reference.PriorInstruction
		if !content.OwnsSubject(reference) || index < 0 || index >= len(content.Effects) ||
			!reference.SubjectProducerMatches(content.Effects[index]) {
			return "structural — optional entered subject has incompatible producer proof"
		}
		owner, optional := groups.owners[index]
		if !optional {
			return ""
		}
		if groups.keys[owner] == "" {
			return "structural — optional entered subject has no existing action receipt"
		}
		groups.requiredReceipts[owner] = true
		return ""
	}
	for _, reference := range content.References {
		if reason := require(reference); reason != "" {
			return reason
		}
		if reference.ReachedCardProducerClauseID != 0 {
			reached, ok := reference.CardLineageSubject(content.Effects)
			if !ok {
				return "structural — optional card lineage has no compatible entered proof"
			}
			if reason := require(reached); reason != "" {
				return reason
			}
		}
	}
	return ""
}
