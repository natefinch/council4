package parser

import "strings"

func libraryCardAntecedent(reference Reference, latest int, observations map[EffectKind]int) int {
	if kind := explicitLibraryCardObservation(reference); kind != EffectUnknown {
		return observations[kind]
	}
	return latest
}

func explicitLibraryCardObservation(reference Reference) EffectKind {
	if reference.Kind == ReferenceThatObject {
		if strings.EqualFold(reference.Text, "the looked-at card") {
			return EffectLookAtLibraryTop
		}
		if strings.EqualFold(reference.Text, "the revealed card") {
			return EffectReveal
		}
	}
	return EffectUnknown
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
