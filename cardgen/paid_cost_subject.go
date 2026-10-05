package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func paidCostSubjectKey(producer *parser.PaidCostProducer) (string, bool) {
	if producer == nil || producer.ClauseID <= 0 || producer.ComponentNodeID <= 0 {
		return "", false
	}
	return fmt.Sprintf("paid-cost-%d-%d", producer.ClauseID, producer.ComponentNodeID), true
}

func lowerPaidCostReference(reference compiler.CompiledReference, subjectTypes []types.Card) (game.ObjectReference, bool) {
	binding := reference.PaidCost
	if reference.Binding != compiler.ReferenceBindingPaidCost || binding == nil || !binding.Known ||
		binding.ConsumerNodeID != reference.NodeID {
		return game.ObjectReference{}, false
	}
	key, ok := paidCostSubjectKey(&binding.Producer)
	if !ok {
		return game.ObjectReference{}, false
	}
	var kind game.PaidCostKind
	switch binding.Domain {
	case parser.PaidCostDomainSacrificedPermanent:
		kind = game.PaidCostSacrifice
	case parser.PaidCostDomainDiscardedCard:
		kind = game.PaidCostDiscard
	default:
		return game.ObjectReference{}, false
	}
	if len(subjectTypes) > 1 || len(subjectTypes) == 1 && (kind != game.PaidCostSacrifice || !subjectTypes[0].IsPermanent()) {
		return game.ObjectReference{}, false
	}
	var noun types.Card
	if len(subjectTypes) == 1 {
		noun = subjectTypes[0]
	}
	return game.PaidCostReference(key, kind, noun), true
}
