package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// addendumAbilityWord is the flavor ability-word label that prefixes the
// Addendum paragraph. It is a rules-free label (see rulesFreeAbilityWordLabel),
// so matching it does not interpret resolving Oracle wording.
const addendumAbilityWord = "Addendum"

// lowerControlledGroupGrantThenAddendumGroupBonus lowers a two-paragraph spell
// whose first paragraph grants a keyword to a battlefield group ("Creatures you
// control gain <keyword> until end of turn.") and whose Addendum paragraph,
// gated on casting during the main phase, places counters on and grants a second
// keyword to "those creatures"/"they" — a plural group back-reference to the
// first paragraph's affected group (Unbreakable Formation).
//
// The two paragraphs compile to independent abilities, so the back-reference has
// no antecedent within the Addendum ability. This combiner resolves it by
// reusing the first paragraph's lowered group for the gated counter placement
// and keyword grant, then fuses both paragraphs into one spell ability. It fails
// closed for any other shape so unrelated two-ability spells stay unaffected.
func lowerControlledGroupGrantThenAddendumGroupBonus(
	cardName string,
	compilation compiler.Compilation,
) (game.AbilityContent, bool) {
	if len(compilation.Abilities) != 2 || len(compilation.Syntax.Abilities) != 2 {
		return game.AbilityContent{}, false
	}
	grantInstructions, group, ok := lowerGroupKeywordGrantSpell(
		cardName, compilation.Abilities[0], &compilation.Syntax.Abilities[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	bonusInstructions, ok := lowerAddendumGroupBonus(compilation.Abilities[1], group)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, len(grantInstructions)+len(bonusInstructions))
	sequence = append(sequence, grantInstructions...)
	sequence = append(sequence, bonusInstructions...)
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerGroupKeywordGrantSpell lowers an unconditional spell paragraph that grants
// a keyword to a battlefield group and returns its instructions together with the
// affected group. It fails closed unless the paragraph lowers cleanly to a
// single non-modal, untargeted spell whose lone instruction applies to one valid
// group.
func lowerGroupKeywordGrantSpell(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) ([]game.Instruction, game.GroupReference, bool) {
	if ability.Kind != compiler.AbilitySpell ||
		ability.AbilityWord != "" ||
		ability.Trigger != nil ||
		ability.Cost != nil ||
		ability.Static != nil ||
		ability.Optional {
		return nil, game.GroupReference{}, false
	}
	lowered, diagnostic := lowerExecutableAbility(cardName, false, nil, -1, ability, syntax)
	if diagnostic != nil ||
		!lowered.complete(ability, syntax) ||
		!lowered.spellAbility.Exists ||
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
		return nil, game.GroupReference{}, false
	}
	spell := lowered.spellAbility.Val
	if spell.IsModal() ||
		len(spell.SharedTargets) != 0 ||
		len(spell.Modes) != 1 ||
		len(spell.Modes[0].Targets) != 0 {
		return nil, game.GroupReference{}, false
	}
	group, ok := spellGroupKeywordTarget(spell.Modes[0].Sequence)
	if !ok {
		return nil, game.GroupReference{}, false
	}
	return spell.Modes[0].Sequence, group, true
}

// spellGroupKeywordTarget extracts the battlefield group a group keyword-grant
// spell applies to. It fails closed unless the sequence is exactly one ungated
// ApplyContinuous on a valid group with no per-object affected target.
func spellGroupKeywordTarget(sequence []game.Instruction) (game.GroupReference, bool) {
	if len(sequence) != 1 ||
		sequence[0].Condition.Exists ||
		sequence[0].ResultGate.Exists {
		return game.GroupReference{}, false
	}
	apply, ok := sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || apply.Object.Exists || len(apply.ContinuousEffects) == 0 {
		return game.GroupReference{}, false
	}
	group := apply.ContinuousEffects[0].Group
	if !group.Valid() {
		return game.GroupReference{}, false
	}
	return group, true
}

// lowerAddendumGroupBonus lowers the Addendum paragraph's gated counter placement
// and keyword grant onto the supplied group (the first paragraph's affected
// group, denoted "those creatures"/"they"). Both instructions are gated on the
// Addendum's cast-during-main-phase condition. It fails closed for any other
// shape.
func lowerAddendumGroupBonus(
	ability compiler.CompiledAbility,
	group game.GroupReference,
) ([]game.Instruction, bool) {
	if ability.Kind != compiler.AbilitySpell ||
		ability.AbilityWord != addendumAbilityWord ||
		ability.Trigger != nil ||
		ability.Cost != nil ||
		ability.Static != nil ||
		ability.Optional {
		return nil, false
	}
	content := ability.Content
	if len(content.Conditions) != 1 ||
		len(content.Effects) != 2 ||
		len(content.Modes) != 0 ||
		len(content.Targets) != 0 {
		return nil, false
	}
	gate, ok := lowerCondition(content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return nil, false
	}
	counterEffect := content.Effects[0]
	keywordEffect := content.Effects[1]
	if !addendumGroupCounterEffect(&counterEffect) ||
		!addendumGroupKeywordEffect(&keywordEffect) {
		return nil, false
	}
	keywords, abilities, ok := partitionTemporaryKeywords(
		keywordsWithinSpan(content.Keywords, keywordEffect.ClauseSpan))
	if !ok || (len(keywords) == 0 && len(abilities) == 0) {
		return nil, false
	}
	sequence := []game.Instruction{
		{Primitive: game.AddCounter{
			Amount:      game.Fixed(counterEffect.Amount.Value),
			Group:       group,
			CounterKind: counterEffect.CounterKind,
		}},
		{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:        game.LayerAbility,
				Group:        group,
				AddKeywords:  keywords,
				AddAbilities: abilities,
			}},
			Duration: game.DurationUntilEndOfTurn,
		}},
	}
	gateCondition := game.EffectCondition{Condition: opt.Val(gate)}
	if !applyEffectConditionGate(sequence, &gateCondition) {
		return nil, false
	}
	return sequence, true
}

// addendumGroupCounterEffect reports whether an effect is "put a +N/+N counter on
// each of those creatures" — a fixed, supported counter placement on a plural
// group back-reference.
func addendumGroupCounterEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectPut &&
		effect.Context == parser.EffectContextController &&
		!effect.Negated &&
		!effect.Optional &&
		effect.Duration == compiler.DurationNone &&
		effect.CounterKindKnown &&
		compiler.CounterKindPlacementSupported(effect.CounterKind) &&
		!effect.CounterKind.PlayerOnly() &&
		effect.Amount.Known &&
		effect.Amount.Value >= 1 &&
		groupBackReferencePronoun(effect.References)
}

// addendumGroupKeywordEffect reports whether an effect is "they gain <keyword>
// until end of turn" — a temporary keyword grant to a plural group
// back-reference.
func addendumGroupKeywordEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectGain &&
		!effect.Negated &&
		!effect.Optional &&
		!effect.KeywordGrantChoice &&
		effect.Duration == compiler.DurationUntilEndOfTurn &&
		effect.StaticSubject == compiler.StaticSubjectNone &&
		groupBackReferencePronoun(effect.SubjectReferences)
}

// groupBackReferencePronoun reports whether the references carry a plural group
// demonstrative ("those creatures") or pronoun ("they") that denotes a prior
// paragraph's affected group rather than a self-contained selection.
func groupBackReferencePronoun(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Kind == compiler.ReferencePronoun &&
			(references[i].Pronoun == compiler.ReferencePronounThose ||
				references[i].Pronoun == compiler.ReferencePronounThey) {
			return true
		}
	}
	return false
}
