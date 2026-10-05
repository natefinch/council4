package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

type optionalActionGroups struct {
	owners          map[int]int
	keys            map[int]game.OptionalDecisionKey
	compound        map[int]bool
	controllerActor map[int]bool
}

func planOptionalActionGroups(effects []compiler.CompiledEffect, indices map[int]int) (optionalActionGroups, string) {
	groups := optionalActionGroups{
		owners: make(map[int]int), keys: make(map[int]game.OptionalDecisionKey), compound: make(map[int]bool),
		controllerActor: make(map[int]bool),
	}
	for ei, effect := range effects {
		if !effect.Optional {
			continue
		}
		if effect.ClauseID <= 0 || len(effect.OptionalActionClauseIDs) == 0 {
			return groups, "structural — optional action has no typed group identity"
		}
		if effect.DelayedSubject.OptionalAtDelayedTime && len(effect.OptionalActionClauseIDs) == 1 &&
			effect.OptionalActionClauseIDs[0] == effect.ClauseID && fixedPhaseSubjectEffectModeled(effect) {
			continue
		}
		for offset, id := range effect.OptionalActionClauseIDs {
			member, exists := indices[id]
			if !exists || member != ei+offset {
				return groups, "structural — optional action group has missing or noncontiguous owners"
			}
			if _, overlaps := groups.owners[member]; overlaps ||
				offset > 0 && effects[member].Optional {
				return groups, "structural — overlapping optional action decisions not modeled"
			}
			if effects[member].DelayedTiming != 0 &&
				(len(effect.OptionalActionClauseIDs) != 1 || !fixedPhaseSubjectEffectModeled(effects[member])) ||
				effects[member].Payment.Form != parser.EffectPaymentFormUnknown {
				return groups, "structural — optional action group contains delay or payment"
			}
			groups.owners[member] = ei
		}
		if len(effect.OptionalActionClauseIDs) > 1 {
			if effect.Context != parser.EffectContextController {
				return groups, "structural — optional action group actor not modeled"
			}
			groups.compound[ei] = true
		}
		groups.keys[ei] = game.OptionalDecisionKey(fmt.Sprintf("optional-clause-%d", effect.ClauseID))
		groups.controllerActor[ei] = effect.Context == parser.EffectContextController
	}
	return groups, ""
}

func (groups optionalActionGroups) isCompound(ei int) bool {
	owner, exists := groups.owners[ei]
	return exists && groups.compound[owner]
}

func (groups optionalActionGroups) ownsOptionalAntecedents(content compiler.AbilityContent, ei int) bool {
	owner, exists := groups.owners[ei]
	if !exists || owner == ei {
		return false
	}
	for _, reference := range content.Effects[ei].References {
		producer := reference.PriorInstruction
		if reference.Binding != compiler.ReferenceBindingPriorInstructionResult ||
			producer < 0 || producer >= ei || !content.Effects[producer].Optional {
			continue
		}
		if otherOwner, exists := groups.owners[producer]; !exists || otherOwner != owner {
			return false
		}
	}
	return true
}

func (groups optionalActionGroups) apply(ei int, sequence []game.Instruction) string {
	owner, exists := groups.owners[ei]
	if !exists {
		return ""
	}
	if owner == ei {
		if len(sequence) > 1 && !groups.controllerActor[owner] {
			return "structural — expanded optional action actor not modeled"
		}
		return markOptionalAction(sequence, groups.keys[owner], groups.compound[owner])
	}
	for i := range sequence {
		if sequence[i].Optional || sequence[i].OptionalActor.Exists || sequence[i].OptionalActorGroup.Exists ||
			sequence[i].OptionalDecisionGate != "" || sequence[i].PublishOptionalDecision != "" ||
			sequence[i].ForEachPlayerGroup.Exists || sequence[i].TemptingOffer {
			return "structural — optional action member has conflicting decision envelope"
		}
		sequence[i].OptionalDecisionGate = groups.keys[owner]
	}
	return ""
}

func markOptionalAction(sequence []game.Instruction, key game.OptionalDecisionKey, publish bool) string {
	if len(sequence) == 0 {
		return "structural — optional action produced no instructions"
	}
	for _, instruction := range sequence {
		if instruction.Optional || instruction.OptionalActor.Exists || instruction.OptionalActorGroup.Exists ||
			instruction.ForEachPlayerGroup.Exists || instruction.TemptingOffer ||
			instruction.PublishOptionalDecision != "" || instruction.OptionalDecisionGate != "" {
			return "structural — optional action has conflicting decision envelope"
		}
	}
	sequence[0].Optional = true
	if publish || len(sequence) > 1 {
		sequence[0].PublishOptionalDecision = key
		for i := 1; i < len(sequence); i++ {
			sequence[i].OptionalDecisionGate = key
		}
	}
	return ""
}
