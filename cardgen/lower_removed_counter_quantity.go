package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func removedCounterQuantityKey(clauseID int) game.ResultKey {
	return game.ResultKey(fmt.Sprintf("removed-counter-quantity-%d", clauseID))
}

func lowerRemoveAllOfKindCounters(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	object, targets, ok := removeCounterObjectAndTargets(ctx)
	if !ok || !effect.Exact || effect.Negated || effect.Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 || len(ctx.content.Modes) != 0 ||
		!effect.CounterKindKnown || !compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{Primitive: game.RemoveCounter{
			Object: object, CounterKind: effect.CounterKind,
			Amount: game.Dynamic(game.DynamicAmount{
				Kind: game.DynamicAmountObjectCounters, CounterKind: effect.CounterKind, Object: object,
			}),
		}}},
	}.Ability(), nil
}

func linkRemovedCounterQuantities(effects []compiler.CompiledEffect, ranges [][2]int, sequence []game.Instruction) string {
	for ci := range effects {
		amount := effects[ci].Amount
		if amount.DynamicKind != compiler.DynamicAmountRemovedCounterCount {
			continue
		}
		if effects[ci].DelayedTiming != 0 {
			return "structural — delayed removed-counter quantity requires schedule-time scalar capture"
		}
		producer := -1
		for pi := range effects[:ci] {
			if amount.ProducerClauseID > 0 && effects[pi].ClauseID == amount.ProducerClauseID {
				if producer >= 0 {
					return "structural — ambiguous removed-counter quantity producer"
				}
				producer = pi
			}
		}
		if producer < 0 || effects[producer].Kind != compiler.EffectRemoveCounter ||
			ranges[producer][1]-ranges[producer][0] != 1 {
			return "structural — removed-counter quantity producer unavailable or expanded"
		}
		publisher := &sequence[ranges[producer][0]]
		if publisher.Primitive.Kind() != game.PrimitiveRemoveCounter {
			return "structural — removed-counter quantity requires an actual counter-removal primitive"
		}
		key := removedCounterQuantityKey(amount.ProducerClauseID)
		if publisher.PublishResult != "" && publisher.PublishResult != key {
			return "structural — removed-counter quantity publication conflicts with another result"
		}
		publisher.PublishResult = key
		for i := ranges[ci][0]; i < ranges[ci][1]; i++ {
			instr := &sequence[i]
			if instr.ResultGate.Exists {
				return "structural — removed-counter quantity consumer carries another result gate"
			}
			instr.ResultGate = opt.Val(game.InstructionResultGate{Key: key, AmountAvailable: true})
		}
	}
	return ""
}

func modalRemovedCounterQuantities(content compiler.AbilityContent) bool {
	for _, mode := range content.Modes {
		for _, effect := range mode.Content.Effects {
			if effect.Amount.DynamicKind == compiler.DynamicAmountRemovedCounterCount {
				return true
			}
		}
		if modalRemovedCounterQuantities(mode.Content) {
			return true
		}
	}
	return false
}
