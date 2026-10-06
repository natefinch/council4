package parser

import "strings"

func libraryCardAntecedent(reference Reference, latest int, observations map[EffectKind]int) int {
	if reference.Kind == ReferenceThatObject {
		if strings.EqualFold(reference.Text, "the looked-at card") {
			return observations[EffectLookAtLibraryTop]
		}
		if strings.EqualFold(reference.Text, "the revealed card") {
			return observations[EffectReveal]
		}
	}
	return latest
}

func linkedLibraryCardRevealProducer(effect EffectSyntax, latest int, observations map[EffectKind]int) int {
	if effect.Kind != EffectReveal || !recognizeLibraryCardAction(&effect) {
		return 0
	}
	producer := 0
	for _, list := range [][]Reference{effect.References, effect.SubjectReferences} {
		for _, reference := range list {
			if !libraryCardReference(reference) {
				continue
			}
			subject := libraryCardAntecedent(reference, latest, observations)
			if subject == 0 || producer != 0 && producer != subject {
				return 0
			}
			producer = subject
		}
	}
	return producer
}
