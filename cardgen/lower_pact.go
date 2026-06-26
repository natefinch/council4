package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerSpellFaceCombiner attempts each whole-face spell combiner that fuses
// multiple compiled abilities into a single spell ability. It returns the first
// combiner that recognizes the face.
func lowerSpellFaceCombiner(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	if spell, ok := lowerCounterThenNextTurnUpkeepDrawAbilities(cardName, compilation); ok {
		return spell, true
	}
	if spell, ok := lowerPactSpellAbilities(cardName, compilation); ok {
		return spell, true
	}
	if spell, ok := lowerThresholdInsteadManaSpellAbilities(cardName, compilation); ok {
		return spell, true
	}
	if spell, ok := lowerInsteadModifyPTSpellAbilities(cardName, compilation); ok {
		return spell, true
	}
	if spell, ok := lowerInsteadDamageSpellAbilities(cardName, compilation); ok {
		return spell, true
	}
	if spell, ok := lowerDamageDieToExileSpellAbilities(cardName, compilation); ok {
		return spell, true
	}
	if spell, ok := lowerControlledGroupGrantThenAddendumGroupBonus(cardName, compilation); ok {
		return spell, true
	}
	return game.AbilityContent{}, false
}

// lowerPactSpellAbilities lowers the Pact cycle: a spell whose main effect is
// followed by a one-shot "At the beginning of your next upkeep, pay {cost}. If
// you don't, you lose the game." delayed payment (Pact of Negation, Slaughter
// Pact, Summoner's Pact, ...). The main effect resolves immediately and sets up
// the pay-or-lose obligation as a delayed triggered ability (CR 603.7, 104.3a).
func lowerPactSpellAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	if len(compilation.Abilities) != 2 ||
		len(compilation.Syntax.Abilities) != 2 {
		return game.AbilityContent{}, false
	}
	mainAbility := compilation.Abilities[0]
	tailAbility := compilation.Abilities[1]
	if mainAbility.Kind != compiler.AbilitySpell ||
		mainAbility.Trigger != nil ||
		mainAbility.Cost != nil ||
		mainAbility.Static != nil ||
		mainAbility.Optional {
		return game.AbilityContent{}, false
	}
	tail, ok := lowerNextUpkeepPayOrLoseTail(tailAbility, &compilation.Syntax.Abilities[1])
	if !ok {
		return game.AbilityContent{}, false
	}
	lowered, diagnostic := lowerExecutableAbility(cardName, false, nil, mainAbility, &compilation.Syntax.Abilities[0])
	if diagnostic != nil || !lowered.complete(mainAbility, &compilation.Syntax.Abilities[0]) {
		return game.AbilityContent{}, false
	}
	if !lowered.spellAbility.Exists ||
		lowered.activatedAbility.Exists ||
		lowered.triggeredAbility.Exists ||
		lowered.manaAbility.Exists ||
		lowered.loyaltyAbility.Exists ||
		lowered.chapterAbility.Exists ||
		lowered.replacementAbility.Exists ||
		len(lowered.staticAbilities) != 0 ||
		lowered.overloadCost.Exists ||
		len(lowered.additionalCosts) != 0 ||
		len(lowered.alternativeCosts) != 0 {
		return game.AbilityContent{}, false
	}
	spell := lowered.spellAbility.Val
	if spell.IsModal() ||
		len(spell.SharedTargets) != 0 ||
		len(spell.Modes) != 1 {
		return game.AbilityContent{}, false
	}
	spell.Modes[0].Sequence = append(spell.Modes[0].Sequence, game.Instruction{
		Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing:  game.DelayedAtBeginningOfNextUpkeep,
			Content: tail,
		}},
	})
	return spell, true
}

// lowerNextUpkeepPayOrLoseTail validates that a triggered ability is the
// one-shot "At the beginning of your next upkeep, pay {cost}. If you don't, you
// lose the game." Pact obligation and lowers its body to the delayed trigger's
// content. The body pays a fixed mana cost during resolution; on failure the
// controller loses the game.
func lowerNextUpkeepPayOrLoseTail(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ability.Kind != compiler.AbilityTriggered ||
		ability.Trigger == nil ||
		ability.Optional ||
		ability.Cost != nil ||
		ability.Static != nil ||
		ability.Trigger.Kind != compiler.TriggerAt ||
		ability.Trigger.Condition != nil ||
		ability.Trigger.MaxTriggersPerTurn != 0 {
		return game.AbilityContent{}, false
	}
	pattern := &ability.Trigger.Pattern
	if pattern.Event != compiler.TriggerEventBeginningOfStep ||
		pattern.Step != compiler.TriggerStepUpkeep ||
		pattern.Controller != compiler.ControllerYou ||
		!pattern.NextOccurrence ||
		pattern.InterveningCondition != nil {
		return game.AbilityContent{}, false
	}
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Conditions) != 1 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ability.Content.Effects[0]
	payment := effect.Payment
	condition := ability.Content.Conditions[0]
	if effect.Kind != compiler.EffectLoseGame ||
		!effect.Exact ||
		effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Context != parser.EffectContextController ||
		effect.Duration != compiler.DurationNone ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 ||
		payment.Form != parser.EffectPaymentFormMayPayThenIfDoesNot ||
		payment.Payer != parser.EffectPaymentPayerController ||
		len(payment.ManaCost) == 0 ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted ||
		condition.NodeID != payment.FailureConditionNodeID ||
		payment.Span.End.Offset >= condition.Span.Start.Offset {
		return game.AbilityContent{}, false
	}
	_ = syntax
	return game.Mode{Sequence: []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay " + payment.ManaCost.String() + "?",
				ManaCost: opt.Val(slices.Clone(payment.ManaCost)),
			}},
			PublishResult: pactPaidResultKey,
		},
		{
			Primitive: game.PlayerLosesGame{Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       pactPaidResultKey,
				Succeeded: game.TriFalse,
			}),
		},
	}}.Ability(), true
}

const pactPaidResultKey = game.ResultKey("pact-paid")
