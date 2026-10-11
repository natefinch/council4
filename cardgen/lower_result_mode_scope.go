package cardgen

import "github.com/natefinch/council4/mtg/game"

func modeResultScopesCompatible(modes []game.Mode) bool {
	previous := make(map[game.ResultKey]bool)
	previousProducts := make(map[game.LinkedKey]bool)
	for _, mode := range modes {
		publications := make(map[game.ResultKey]bool)
		collectModeResultPublications(mode.Sequence, false, publications)
		for key, conditional := range publications {
			if earlier, exists := previous[key]; exists && (earlier || conditional) {
				return false
			}
			previous[key] = conditional
		}
		products := make(map[game.LinkedKey]bool)
		collectModeLocalProductPublications(mode.Sequence, false, products)
		for key, conditional := range products {
			if earlier, exists := previousProducts[key]; exists && (earlier || conditional) {
				return false
			}
			previousProducts[key] = conditional
		}
	}
	return true
}

func collectModeResultPublications(sequence []game.Instruction, inheritedConditional bool, publications map[game.ResultKey]bool) {
	for _, instruction := range sequence {
		conditional := inheritedConditional || instruction.Condition.Exists ||
			instruction.ConditionGate != "" || instruction.ResultGate.Exists ||
			instruction.OptionalDecisionGate != ""
		if instruction.PublishResult != "" && !instruction.LocalProducts.HasResult(instruction.PublishResult) {
			publications[instruction.PublishResult] = publications[instruction.PublishResult] || conditional
		}
		for _, key := range game.PublishedScalarKeys(instruction.Primitive) {
			if !instruction.LocalProducts.HasResult(key) {
				publications[key] = publications[key] || conditional
			}
		}
		// Repeat bodies share the stack object's results and may execute zero times.
		if repeat, ok := instruction.Primitive.(game.RepeatProcess); ok {
			for _, mode := range repeat.Body.Modes {
				collectModeResultPublications(mode.Sequence, true, publications)
			}
		}
	}
}
