package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func actualCardQuantityKey(clauseID int) game.ResultKey {
	return game.ResultKey(fmt.Sprintf("actual-card-quantity-%d", clauseID))
}

func actualCardQuantitySelection(amount compiler.CompiledAmount) (game.Selection, bool) {
	if !amount.ActualCardQuantitySupported() {
		return game.Selection{}, false
	}
	selector := amount.Selector()
	permanentCard := selector.Kind == compiler.SelectorPermanent
	if selector.Kind == compiler.SelectorCard {
		selector.Kind = compiler.SelectorPermanent
	}
	selection, ok := SelectionForSelector(selector)
	if permanentCard {
		cardSelection, cardOK := cardSelectionForSelector(selector)
		if !cardOK {
			return game.Selection{}, false
		}
		selection.RequiredTypesAny = cardSelection.RequiredTypesAny
	}
	selection.Colored = selector.Colored
	return selection, ok && len(game.ValidateResultObjectSelection(selection)) == 0
}

func actualCardQuantityPublisher(primitive game.Primitive, outcome compiler.EffectKind) bool {
	if primitive == nil {
		return false
	}
	switch outcome {
	case compiler.EffectExile:
		switch primitive.Kind() {
		case game.PrimitiveMoveCard:
			move, ok := primitive.(game.MoveCard)
			return ok && move.Destination == zone.Exile && !move.IncludeEventPermanentComponents
		case game.PrimitiveMoveTopOfLibrary:
			move, ok := primitive.(game.MoveTopOfLibrary)
			return ok && move.Destination == zone.Exile && !move.FaceDown
		case game.PrimitiveChooseFromZone:
			choose, ok := primitive.(game.ChooseFromZone)
			return ok && choose.Destination.Zone == zone.Exile && !choose.SplitSecondary.Exists && !choose.Riders.FaceDown
		default:
			return false
		}
	case compiler.EffectMill:
		move, ok := primitive.(game.MoveTopOfLibrary)
		return ok && move.Destination == zone.Graveyard && !move.FaceDown
	case compiler.EffectDiscard:
		_, ok := primitive.(game.Discard)
		return ok
	default:
		return false
	}
}

func linkActualCardQuantities(effects []compiler.CompiledEffect, ranges [][2]int, sequence []game.Instruction) string {
	for ci := range effects {
		amount := effects[ci].Amount
		if amount.DynamicKind != compiler.DynamicAmountActualCardResultCount {
			continue
		}
		pi, ok := amount.ActualCardQuantityProducer(effects, effects[ci].ClauseID)
		if !ok || ranges[pi][1]-ranges[pi][0] != 1 {
			return "structural — actual-card quantity requires one owned card-result publisher"
		}
		publisher := &sequence[ranges[pi][0]]
		if publisher.TemptingOffer {
			return "structural — actual-card quantity producer has offer-acceptance semantics"
		}
		if !actualCardQuantityPublisher(publisher.Primitive, effects[pi].Kind) {
			return "structural — actual-card quantity producer lacks a card-only result protocol"
		}
		key := actualCardQuantityKey(amount.ProducerClauseID)
		if publisher.PublishResult != "" && publisher.PublishResult != key {
			return "structural — actual-card quantity publication conflicts with another result"
		}
		publisher.PublishResult = key
		if !publisher.LocalProducts.HasResult(key) {
			publisher.LocalProducts.Results = append(publisher.LocalProducts.Results, key)
		}
		for i := ranges[ci][0]; i < ranges[ci][1]; i++ {
			instruction := &sequence[i]
			if !slices.Contains(game.ConsumedResultKeys(instruction.Primitive), key) {
				return "structural — actual-card quantity was not consumed by the primitive"
			}
			if instruction.ResultGate.Exists {
				return "structural — actual-card quantity consumer carries another result gate"
			}
			instruction.ResultGate = opt.Val(game.InstructionResultGate{Key: key, Accepted: game.TriTrue})
		}
	}
	return ""
}
