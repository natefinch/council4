package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerClassLevelGain lowers a Class enchantment's level-up activated ability
// ("{cost}: Level N") into a sorcery-speed ActivatedAbility that raises the
// source's class level to N (CR 716). Its activation is gated to the level band
// directly below N: the source must currently be below level N, and at the
// previous band's level when that band is 2 or higher. The cost must be pure
// mana; any additional cost component leaves the ability unsupported.
func lowerClassLevelGain(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
	previousLevel int,
) (abilityLowering, *shared.Diagnostic) {
	const unsupportedDetail = "the executable source backend supports only a pure mana cost on a Class level-up ability"
	if ability.Cost == nil {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported Class level ability", unsupportedDetail)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok || manaCost == nil || len(additionalCosts) != 0 {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported Class level ability", unsupportedDetail)
	}
	level := ability.ClassLevelGain
	condition := game.Condition{SourceClassLevelLessThan: level}
	if previousLevel >= 2 {
		condition.SourceClassLevelAtLeast = previousLevel
	}
	activated := game.ActivatedAbility{
		Text:                ability.Text,
		ManaCost:            opt.Val(manaCost),
		Timing:              game.SorceryOnly,
		ActivationCondition: opt.Val(condition),
		Content: game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.SetClassLevel{
						Object: game.SourcePermanentReference(),
						Amount: game.Fixed(level),
					},
				},
			},
		}.Ability(),
	}
	return abilityLowering{
		activatedAbility: opt.Val(activated),
		consumed:         semanticConsumption{cost: true},
		sourceSpans:      []shared.Span{ability.Span},
	}, nil
}

// gateLoweredAbilityByClassLevel restricts a lowered Class-band ability to the
// levels at or above its band by merging SourceClassLevelAtLeast into the
// ability's own condition (CR 716.2c). It covers the ability shapes Class bands
// use — activated, triggered, and static abilities — and fails closed for any
// other lowered output so an unexpected shape stays unsupported.
func gateLoweredAbilityByClassLevel(
	lowered *abilityLowering,
	ability compiler.CompiledAbility,
	level int,
) *shared.Diagnostic {
	const unsupportedDetail = "the executable source backend cannot gate this Class-band ability shape by class level"
	gated := false
	if lowered.activatedAbility.Exists {
		activated := lowered.activatedAbility.Val
		activated.ActivationCondition = mergeClassLevelGate(activated.ActivationCondition, level)
		lowered.activatedAbility = opt.Val(activated)
		gated = true
	}
	if lowered.triggeredAbility.Exists {
		triggered := lowered.triggeredAbility.Val
		triggered.Trigger.InterveningCondition = mergeClassLevelGate(triggered.Trigger.InterveningCondition, level)
		lowered.triggeredAbility = opt.Val(triggered)
		gated = true
	}
	for i := range lowered.staticAbilities {
		if lowered.staticAbilities[i].VarName != "" {
			return executableDiagnostic(ability, "unsupported Class level ability", unsupportedDetail)
		}
		lowered.staticAbilities[i].Body.Condition = mergeClassLevelGate(lowered.staticAbilities[i].Body.Condition, level)
		gated = true
	}
	if lowered.manaAbility.Exists ||
		lowered.loyaltyAbility.Exists ||
		lowered.chapterAbility.Exists ||
		lowered.replacementAbility.Exists ||
		lowered.spellAbility.Exists {
		return executableDiagnostic(ability, "unsupported Class level ability", unsupportedDetail)
	}
	if !gated {
		return executableDiagnostic(ability, "unsupported Class level ability", unsupportedDetail)
	}
	return nil
}

// mergeClassLevelGate sets SourceClassLevelAtLeast on a condition, preserving any
// existing condition fields so a banded ability that carries its own condition is
// gated by both its condition and the class level.
func mergeClassLevelGate(condition opt.V[game.Condition], level int) opt.V[game.Condition] {
	merged := condition.Val
	merged.SourceClassLevelAtLeast = level
	return opt.Val(merged)
}
