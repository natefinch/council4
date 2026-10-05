package cardgen

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

func sequenceLocalProductKey(key game.LinkedKey) bool {
	text, prefix := strings.CutPrefix(string(key), "sequence-effect-")
	text, suffix := strings.CutSuffix(text, "-product")
	if !prefix || !suffix {
		return false
	}
	index, err := strconv.Atoi(text)
	return err == nil && index >= 0 && sequenceProductKey(index) == key
}

// Only canonical planner products are local. CR 607 links and paid-cost facts
// are persistent and must not be cleared or renamed to isolate mode frames.
func collectModeLocalProductPublications(sequence []game.Instruction, inheritedConditional bool, publications map[game.LinkedKey]bool) {
	for _, instruction := range sequence {
		conditional := inheritedConditional || instruction.Condition.Exists ||
			instruction.ConditionGate != "" || instruction.ResultGate.Exists
		if instruction.Primitive == nil {
			continue
		}
		key := game.PublishedLinkedKey(instruction.Primitive)
		if sequenceLocalProductKey(key) && !instruction.LocalProducts.HasLink(key) {
			publications[key] = publications[key] || conditional
		}
		if repeat, ok := instruction.Primitive.(game.RepeatProcess); ok {
			for _, mode := range repeat.Body.Modes {
				collectModeLocalProductPublications(mode.Sequence, true, publications)
			}
		}
	}
}
