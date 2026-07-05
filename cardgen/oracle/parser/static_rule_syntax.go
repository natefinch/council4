package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseStaticRuleSyntax(tokens []shared.Token) (*StaticRuleSyntax, bool) {
	if len(tokens) < 5 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	if rule, ok := parseTrueLureStaticRule(tokens); ok {
		return rule, true
	}
	if rule, ok := parseAssignDamageAsUnblockedStaticRule(tokens); ok {
		return rule, true
	}
	subject, next, ok := parseStaticRuleSubject(tokens)
	if !ok {
		return nil, false
	}
	return parseStaticRuleOperationsForSubject(tokens, subject, next)
}

// parseTrueLureStaticRule recognizes the true-lure requirement "All creatures
// able to block <subject> do so.", where <subject> names the source creature
// ("this creature") or the creature an Aura or Equipment is attached to
// ("enchanted creature"/"equipped creature"). Every creature able to block the
// subject attacker must do so (CR 509.1c). Forms with an extra blocker filter
// ("All creatures with flying able to block ...", "All Walls able to block ...")
// or a compound subject ("... this creature or enchanted creature ...") do not
// match and fail closed.
func parseTrueLureStaticRule(tokens []shared.Token) (*StaticRuleSyntax, bool) {
	const prefix = 5 // "All creatures able to block"
	if !staticRuleWordsAt(tokens, 0, "all", "creatures", "able", "to", "block") {
		return nil, false
	}
	end := len(tokens) - 1 // index of the trailing period
	if !staticRuleWordsAt(tokens, end-2, "do", "so") {
		return nil, false
	}
	subject, ok := parseAttachableCreatureSubject(tokens[prefix : end-2])
	if !ok {
		return nil, false
	}
	rule := &StaticRuleSyntax{
		Span:    shared.SpanOf(tokens),
		Subject: subject,
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationBlockedByAll,
			Voice: StaticRuleVoicePassive,
		},
	}
	if !validStaticRuleSyntax(*rule) {
		return nil, false
	}
	return rule, true
}

// parseAssignDamageAsUnblockedStaticRule recognizes "You may have <subject> assign
// <its> combat damage as though <it> weren't blocked.", the permission for the
// subject attacker to deal its combat damage to its attack target as though it
// weren't blocked. <subject> names the source creature ("this creature"); the
// possessive and subject pronouns ("its"/"his"/"her"/"their" and "it"/"he"/
// "she"/"they") agree with the printed creature. A trailing "this turn" or other
// rider makes the form fail closed.
func parseAssignDamageAsUnblockedStaticRule(tokens []shared.Token) (*StaticRuleSyntax, bool) {
	if !staticRuleWordsAt(tokens, 0, "you", "may", "have") {
		return nil, false
	}
	subject, ok := parseAttachableCreatureSubject(tokens[3:5])
	if !ok || subject.Kind != StaticRuleSubjectSourceCreature {
		return nil, false
	}
	if !staticRuleWordsAt(tokens, 5, "assign") ||
		!staticRulePossessivePronounAt(tokens, 6) ||
		!staticRuleWordsAt(tokens, 7, "combat", "damage", "as", "though") ||
		!staticRuleSubjectPronounAt(tokens, 11) ||
		!staticRuleWordsAt(tokens, 12, "weren't", "blocked") {
		return nil, false
	}
	if len(tokens) != 15 { // 14 words plus the trailing period
		return nil, false
	}
	rule := &StaticRuleSyntax{
		Span:    shared.SpanOf(tokens),
		Subject: subject,
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationAssignDamageAsUnblocked,
			Voice: StaticRuleVoicePassive,
		},
	}
	if !validStaticRuleSyntax(*rule) {
		return nil, false
	}
	return rule, true
}

// parseAttachableCreatureSubject parses a self or attached creature subject that
// spans exactly the given tokens: "this creature" (source creature) or "enchanted
// creature"/"equipped creature" (the creature an Aura or Equipment is attached
// to). Any other or partial span fails closed.
func parseAttachableCreatureSubject(tokens []shared.Token) (StaticRuleSubject, bool) {
	subject, next, ok := parseStaticRuleSubject(tokens)
	if !ok || next != len(tokens) {
		return StaticRuleSubject{}, false
	}
	switch subject.Kind {
	case StaticRuleSubjectSourceCreature, StaticRuleSubjectAttachedObject:
		return subject, true
	default:
		return StaticRuleSubject{}, false
	}
}

// staticRulePossessivePronounAt reports whether the token at index is a
// third-person possessive pronoun ("its", "his", "her", "their").
func staticRulePossessivePronounAt(tokens []shared.Token, index int) bool {
	return staticRuleWordsAt(tokens, index, "its") ||
		staticRuleWordsAt(tokens, index, "his") ||
		staticRuleWordsAt(tokens, index, "her") ||
		staticRuleWordsAt(tokens, index, "their")
}

// staticRuleSubjectPronounAt reports whether the token at index is a third-person
// subject pronoun ("it", "he", "she", "they").
func staticRuleSubjectPronounAt(tokens []shared.Token, index int) bool {
	return staticRuleWordsAt(tokens, index, "it") ||
		staticRuleWordsAt(tokens, index, "he") ||
		staticRuleWordsAt(tokens, index, "she") ||
		staticRuleWordsAt(tokens, index, "they")
}

// parseSelfNameStaticRuleSyntax recognizes a static-rule sentence whose subject
// is the card's own name ("Toski attacks each combat if able.") rather than a
// "this creature"/"this permanent" marker. A self-name names the source object,
// so it adopts the source-creature subject; operations restricted to spells
// ("can't be countered", always printed as "This spell") fail closed.
func parseSelfNameStaticRuleSyntax(tokens []shared.Token, atoms Atoms) (*StaticRuleSyntax, bool) {
	if len(tokens) < 4 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	width, ok := selfNameStaticRuleSubjectWidth(tokens, atoms)
	if !ok {
		return nil, false
	}
	subject := StaticRuleSubject{Kind: StaticRuleSubjectSourceCreature, Span: shared.SpanOf(tokens[:width])}
	return parseStaticRuleOperationsForSubject(tokens, subject, width)
}

// selfNameStaticRuleSubjectAt reports the span and token width of a self-name
// subject beginning at tokens[0], matched against the card's source-name
// aliases. Only the printed name qualifies; "this <marker>" subjects are handled
// by parseStaticRuleSubject.
func selfNameStaticRuleSubjectWidth(tokens []shared.Token, atoms Atoms) (int, bool) {
	if len(tokens) == 0 {
		return 0, false
	}
	for _, span := range atoms.SourceNameSpans() {
		if span.Start.Offset != tokens[0].Span.Start.Offset {
			continue
		}
		if width := tokensCoveredCount(tokens, span); width > 0 {
			return width, true
		}
	}
	return 0, false
}

// parseStaticOperationQualifier parses the optional qualifier that follows a
// prohibited operation, trying each recognized qualifier form in turn and
// returning the first match with the index past it. The activation clause (the
// Arrest-family compound "..., and its activated abilities can't be activated")
// only attaches to an active "can't attack or block" prohibition.
func parseStaticOperationQualifier(tokens []shared.Token, operation *StaticRuleOperation, opNext int) (StaticRuleQualifier, int, bool) {
	limit := len(tokens) - 1
	if qualifier, qualifierNext, ok := parseStaticBlockedExceptClause(tokens, operation, opNext, limit); ok {
		return qualifier, qualifierNext, true
	}
	if qualifier, qualifierNext, ok := parseStaticBlockerRestrictionQualifier(tokens, opNext, limit); ok {
		return qualifier, qualifierNext, true
	}
	if qualifier, qualifierNext, ok := parseStaticByMoreThanOneQualifier(tokens, opNext, limit); ok {
		return qualifier, qualifierNext, true
	}
	if qualifier, qualifierNext, ok := parseStaticAloneQualifier(tokens, opNext); ok {
		return qualifier, qualifierNext, true
	}
	if operation.Kind == StaticRuleOperationAttackOrBlock && operation.Voice == StaticRuleVoiceActive {
		if qualifier, qualifierNext, ok := parseStaticActivatedAbilitiesClause(tokens, opNext); ok {
			return qualifier, qualifierNext, true
		}
	}
	return StaticRuleQualifier{}, opNext, false
}

func parseStaticRuleOperationsForSubject(tokens []shared.Token, subject StaticRuleSubject, next int) (*StaticRuleSyntax, bool) {
	rule := &StaticRuleSyntax{
		Span:    shared.SpanOf(tokens),
		Subject: subject,
	}
	if constraint, ok := parseStaticRuleProhibition(tokens, next); ok {
		rule.Constraint = constraint
		next++
		operation, opNext, ok := parseProhibitedStaticRuleOperation(tokens, next)
		if !ok {
			return nil, false
		}
		rule.Operation = operation
		if qualifier, qualifierNext, ok := parseStaticOperationQualifier(tokens, &rule.Operation, opNext); ok {
			rule.Qualifiers = append(rule.Qualifiers, qualifier)
			opNext = qualifierNext
		}
		if staticRuleHasGuardClause(tokens, opNext) {
			rule.Guarded = true
			opNext = len(tokens) - 1
		}
		if opNext != len(tokens)-1 {
			return nil, false
		}
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if attack, ok := parseRequiredAttackRule(tokens, next); ok {
		rule.Constraint = attack.Constraint
		rule.Operation = attack.Operation
		rule.Qualifiers = attack.Qualifiers
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if block, ok := parseRequiredBlockRule(tokens, next); ok {
		rule.Constraint = block.Constraint
		rule.Operation = block.Operation
		rule.Qualifiers = block.Qualifiers
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if block, ok := parseCanBlockOnlyFlyingRule(tokens, next); ok {
		rule.Constraint = block.Constraint
		rule.Operation = block.Operation
		rule.Qualifiers = block.Qualifiers
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if block, opNext, ok := parseCanBlockAdditionalRule(tokens, next); ok {
		rule.Constraint = block.Constraint
		rule.Operation = block.Operation
		rule.Qualifiers = block.Qualifiers
		if staticRuleHasGuardClause(tokens, opNext) {
			rule.Guarded = true
			opNext = len(tokens) - 1
		}
		if opNext != len(tokens)-1 {
			return nil, false
		}
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if untap, ok := parseStaticDoesntUntapRule(tokens, next); ok {
		rule.Constraint = untap.Constraint
		rule.Operation = untap.Operation
		opNext := untap.OpNext
		if staticRuleHasGuardClause(tokens, opNext) {
			rule.Guarded = true
			opNext = len(tokens) - 1
		}
		if opNext != len(tokens)-1 {
			return nil, false
		}
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	return nil, false
}

func parseStaticRuleSubject(tokens []shared.Token) (StaticRuleSubject, int, bool) {
	if !staticRuleWordsAt(tokens, 0, "this") {
		// "Enchanted permanent"/"enchanted artifact" name the object an Aura is
		// attached to, the same attached-object source "enchanted creature"
		// names. Auras that freeze a noncreature permanent ("Enchanted permanent
		// doesn't untap during its controller's untap step.", Ice Over) use these
		// wider nouns, so they thread the attached-object subject too.
		if staticRuleWordsAt(tokens, 0, "enchanted", "creature") ||
			staticRuleWordsAt(tokens, 0, "equipped", "creature") {
			return StaticRuleSubject{
				Kind: StaticRuleSubjectAttachedObject,
				Span: shared.SpanOf(tokens[:2]),
			}, 2, true
		}
		if staticRuleWordsAt(tokens, 0, "enchanted", "permanent") ||
			staticRuleWordsAt(tokens, 0, "enchanted", "artifact") {
			return StaticRuleSubject{
				Kind: StaticRuleSubjectAttachedPermanent,
				Span: shared.SpanOf(tokens[:2]),
			}, 2, true
		}
		return StaticRuleSubject{}, 0, false
	}
	if len(tokens) < 2 {
		return StaticRuleSubject{}, 0, false
	}
	subject := StaticRuleSubject{Span: shared.SpanOf(tokens[:2])}
	switch {
	case staticRuleWordsAt(tokens, 1, "creature"):
		subject.Kind = StaticRuleSubjectSourceCreature
	// A created token's own rule text reads "This token ..."; the token is the
	// self source the "this creature" form names. Only quoted token-granted
	// abilities reach this recognizer with that wording, so it threads the
	// token's self combat rule through the SourceCreature subject.
	case staticRuleWordsAt(tokens, 1, "token"):
		subject.Kind = StaticRuleSubjectSourceCreature
	case staticRuleWordsAt(tokens, 1, "artifact"),
		staticRuleWordsAt(tokens, 1, "permanent"),
		staticRuleWordsAt(tokens, 1, "land"):
		subject.Kind = StaticRuleSubjectSourcePermanent
	case staticRuleWordsAt(tokens, 1, "spell"):
		subject.Kind = StaticRuleSubjectSourceSpell
	default:
		return StaticRuleSubject{}, 0, false
	}
	return subject, 2, true
}

func parseStaticRuleProhibition(tokens []shared.Token, start int) (StaticRuleConstraint, bool) {
	if !staticRuleWordsAt(tokens, start, "can't") && !staticRuleWordsAt(tokens, start, "cannot") {
		return StaticRuleConstraint{}, false
	}
	return StaticRuleConstraint{
		Kind: StaticRuleConstraintProhibition,
		Span: tokens[start].Span,
	}, true
}

func parseProhibitedStaticRuleOperation(tokens []shared.Token, start int) (StaticRuleOperation, int, bool) {
	if staticRuleWordsAt(tokens, start, "attack", "or", "block") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationAttackOrBlock,
			Voice: StaticRuleVoiceActive,
			Span:  shared.SpanOf(tokens[start : start+3]),
		}, start + 3, true
	}
	if staticRuleWordsAt(tokens, start, "attack") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationAttack,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[start].Span,
		}, start + 1, true
	}
	if staticRuleWordsAt(tokens, start, "block", "and", "can't", "be", "blocked") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationBlockAndBeBlocked,
			Voice: StaticRuleVoiceActive,
			Span:  shared.SpanOf(tokens[start : start+5]),
		}, start + 5, true
	}
	if staticRuleWordsAt(tokens, start, "block") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[start].Span,
		}, start + 1, true
	}
	if staticRuleWordsAt(tokens, start, "be", "blocked") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[start : start+2]),
		}, start + 2, true
	}
	if staticRuleWordsAt(tokens, start, "be", "countered") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationCounter,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[start : start+2]),
		}, start + 2, true
	}
	return StaticRuleOperation{}, start, false
}

// parseStaticAloneQualifier recognizes the "alone" word that bounds an active
// "can't attack"/"can't block"/"can't attack or block" prohibition to the
// exceptional case where the subject is the only creature attacking or blocking
// ("can't attack alone"). It matches the single "alone" token immediately
// following the operation.
func parseStaticAloneQualifier(tokens []shared.Token, start int) (StaticRuleQualifier, int, bool) {
	if !staticRuleWordsAt(tokens, start, "alone") {
		return StaticRuleQualifier{}, start, false
	}
	return StaticRuleQualifier{
		Kind: StaticRuleQualifierAlone,
		Span: tokens[start].Span,
	}, start + 1, true
}

// parseStaticActivatedAbilitiesClause recognizes Arrest's trailing compound
// clause appended to an active "can't attack or block" prohibition: a comma,
// then "and its activated abilities can't be activated", optionally extended
// with "unless they're mana abilities" (Faith's Fetters). It returns the
// activation qualifier and the index past the clause so the caller can confirm
// only the closing period remains. The mana-exemption variant yields
// StaticRuleQualifierCantActivateNonManaAbilities; the plain form yields
// StaticRuleQualifierCantActivateAbilities.
func parseStaticActivatedAbilitiesClause(tokens []shared.Token, start int) (StaticRuleQualifier, int, bool) {
	if start < 0 || start >= len(tokens) || tokens[start].Kind != shared.Comma {
		return StaticRuleQualifier{}, start, false
	}
	if !staticRuleWordsAt(tokens, start+1, "and", "its", "activated", "abilities", "can't", "be", "activated") {
		return StaticRuleQualifier{}, start, false
	}
	clauseEnd := start + 8
	kind := StaticRuleQualifierCantActivateAbilities
	if staticRuleWordsAt(tokens, clauseEnd, "unless", "they're", "mana", "abilities") {
		kind = StaticRuleQualifierCantActivateNonManaAbilities
		clauseEnd += 4
	}
	return StaticRuleQualifier{
		Kind: kind,
		Span: shared.SpanOf(tokens[start:clauseEnd]),
	}, clauseEnd, true
}

type requiredAttackRuleSyntax struct {
	Constraint StaticRuleConstraint  `json:",omitzero"`
	Operation  StaticRuleOperation   `json:",omitzero"`
	Qualifiers []StaticRuleQualifier `json:",omitempty"`
}

func parseRequiredAttackRule(tokens []shared.Token, start int) (requiredAttackRuleSyntax, bool) {
	constraintStart := start
	operationStart := start
	explicit := staticRuleWordsAt(tokens, start, "must")
	if explicit {
		operationStart++
	}
	if (explicit && !staticRuleWordsAt(tokens, operationStart, "attack")) ||
		(!explicit && !staticRuleWordsAt(tokens, operationStart, "attacks")) {
		return requiredAttackRuleSyntax{}, false
	}
	qualifierStart := operationStart + 1
	if !staticRuleWordsAt(tokens, qualifierStart, "each", "combat", "if", "able") ||
		qualifierStart+4 != len(tokens)-1 {
		return requiredAttackRuleSyntax{}, false
	}
	constraintSpan := shared.SpanOf(tokens[constraintStart : qualifierStart+4])
	if operationStart != constraintStart {
		constraintSpan = tokens[constraintStart].Span
	}
	return requiredAttackRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
			Span: constraintSpan,
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationAttack,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[operationStart].Span,
		},
		Qualifiers: []StaticRuleQualifier{
			{
				Kind: StaticRuleQualifierEachCombat,
				Span: shared.SpanOf(tokens[qualifierStart : qualifierStart+2]),
			},
			{
				Kind: StaticRuleQualifierIfAble,
				Span: shared.SpanOf(tokens[qualifierStart+2 : qualifierStart+4]),
			},
		},
	}, true
}

func parseRequiredBlockRule(tokens []shared.Token, start int) (requiredAttackRuleSyntax, bool) {
	if !staticRuleWordsAt(tokens, start, "must", "be", "blocked", "if", "able") ||
		start+5 != len(tokens)-1 {
		return requiredAttackRuleSyntax{}, false
	}
	return requiredAttackRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
			Span: tokens[start].Span,
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[start+1 : start+3]),
		},
		Qualifiers: []StaticRuleQualifier{
			{
				Kind: StaticRuleQualifierIfAble,
				Span: shared.SpanOf(tokens[start+3 : start+5]),
			},
		},
	}, true
}

// parseCanBlockOnlyFlyingRule recognizes the blocker-side permission restriction
// "can block only creatures with flying" (Cloud Sprite, Gloomwidow): the subject
// creature may block only attackers that have flying. The phrasing is fixed and
// fully consumed; any other restriction wording fails closed so the recognizer
// stays narrow. It models the restriction as a requirement on the active block
// operation bounded by the flying-attacker qualifier.
func parseCanBlockOnlyFlyingRule(tokens []shared.Token, start int) (requiredAttackRuleSyntax, bool) {
	if !staticRuleWordsAt(tokens, start, "can", "block", "only", "creatures", "with", "flying") ||
		start+6 != len(tokens)-1 {
		return requiredAttackRuleSyntax{}, false
	}
	return requiredAttackRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
			Span: shared.SpanOf(tokens[start : start+3]),
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[start+1].Span,
		},
		Qualifiers: []StaticRuleQualifier{{
			Kind: StaticRuleQualifierBlockedAttackerFlying,
			Span: shared.SpanOf(tokens[start+2 : start+6]),
		}},
	}, true
}

// doesntUntapRuleSyntax bundles a parsed "doesn't untap" prohibition with the
// cursor just past its fixed "untap step" phrasing so the caller can detect a
// trailing guard clause.
type doesntUntapRuleSyntax struct {
	Constraint StaticRuleConstraint `json:",omitzero"`
	Operation  StaticRuleOperation  `json:",omitzero"`
	OpNext     int                  `json:",omitempty"`
}

// parseCanBlockAdditionalRule recognizes the blocker-side capability "can block
// an additional creature each combat" (Brave the Sands, Coastline Chimera): the
// subject creature may block one more attacker than the usual single-blocker
// limit. The phrasing is fixed; "each combat" is consumed as part of the fixed
// wording. It returns the cursor just past the recognized phrasing so the caller
// can detect a trailing guard clause ("... as long as you're the monarch.",
// Entourage of Trest).
func parseCanBlockAdditionalRule(tokens []shared.Token, start int) (requiredAttackRuleSyntax, int, bool) {
	if !staticRuleWordsAt(tokens, start, "can", "block", "an", "additional", "creature", "each", "combat") {
		return requiredAttackRuleSyntax{}, 0, false
	}
	return requiredAttackRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
			Span: shared.SpanOf(tokens[start : start+2]),
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[start+1].Span,
		},
		Qualifiers: []StaticRuleQualifier{{
			Kind: StaticRuleQualifierAdditionalCreature,
			Span: shared.SpanOf(tokens[start+2 : start+7]),
		}},
	}, start + 7, true
}

// parseStaticDoesntUntapRule recognizes "doesn't untap during your untap step"
// or "doesn't untap during its controller's untap step", modeling the frozen
// permanent as a prohibition on the untap operation. OpNext points just past the
// fixed "untap step" phrasing so the caller can detect a trailing guard clause
// ("... unless that player is the monarch.", Fall from Favor).
func parseStaticDoesntUntapRule(tokens []shared.Token, start int) (doesntUntapRuleSyntax, bool) {
	if !staticRuleWordsAt(tokens, start, "doesn't", "untap", "during") {
		return doesntUntapRuleSyntax{}, false
	}
	cursor := start + 3
	switch {
	case staticRuleWordsAt(tokens, cursor, "your"):
		cursor++
	case staticRuleWordsAt(tokens, cursor, "its", "controller's"):
		cursor += 2
	default:
		return doesntUntapRuleSyntax{}, false
	}
	if !staticRuleWordsAt(tokens, cursor, "untap", "step") {
		return doesntUntapRuleSyntax{}, false
	}
	return doesntUntapRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintProhibition,
			Span: shared.SpanOf(tokens[start : start+1]),
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationUntap,
			Voice: StaticRuleVoiceActive,
			Span:  shared.SpanOf(tokens[start+1 : cursor+2]),
		},
		OpNext: cursor + 2,
	}, true
}

// parseStaticBlockerRestrictionQualifier consumes the blocker-characteristic
// restriction "by creatures with flying", "by creatures with power N or less",
// or "by creatures with power N or greater" that bounds a passive "can't be
// blocked" prohibition to blockers matching that characteristic. The phrasing is
// fixed; any deviation fails closed. end is the exclusive bound (the period
// index) so the qualifier never consumes the terminating punctuation.
// parseStaticBlockerTypeQualifier consumes the blocker-color or
// artifact-creature restriction "by <color> creatures" ("by white creatures")
// or "by artifact creatures" that bounds a passive "can't be blocked"
// prohibition to blockers matching that characteristic. The phrasing is fixed;
// any deviation fails closed. end is the exclusive bound (the period index) so
// the qualifier never consumes the terminating punctuation.
func parseStaticBlockerTypeQualifier(tokens []shared.Token, start, end int) (StaticRuleQualifier, int, bool) {
	if !staticRuleWordsAt(tokens, start, "by") || start+3 > end {
		return StaticRuleQualifier{}, 0, false
	}
	if !staticRuleWordsAt(tokens, start+2, "creatures") {
		return StaticRuleQualifier{}, 0, false
	}
	if staticRuleWordsAt(tokens, start+1, "artifact") {
		return StaticRuleQualifier{
			Kind: StaticRuleQualifierBlockerArtifact,
			Span: shared.SpanOf(tokens[start : start+3]),
		}, start + 3, true
	}
	if tokens[start+1].Kind != shared.Word {
		return StaticRuleQualifier{}, 0, false
	}
	color, ok := recognizeColorWord(tokens[start+1].Text)
	if !ok {
		return StaticRuleQualifier{}, 0, false
	}
	return StaticRuleQualifier{
		Kind:  StaticRuleQualifierBlockerColor,
		Span:  shared.SpanOf(tokens[start : start+3]),
		Color: color,
	}, start + 3, true
}

// parseStaticBlockedExceptClause recognizes the "except by ..." tail that turns a
// passive "can't be blocked" prohibition into the restricted "can't be blocked
// except by ..." prohibition, where only blockers matching the trailing blocker
// characteristic may block the subject. It matches only after a passive block
// operation; on a match it retargets the operation to
// StaticRuleOperationBlockedExcept and returns the bounding blocker qualifier.
// Any "except by ..." wording whose characteristic is not recognized fails closed
// so the sentence stays unsupported rather than misparsing.
func parseStaticBlockedExceptClause(tokens []shared.Token, operation *StaticRuleOperation, start, end int) (StaticRuleQualifier, int, bool) {
	if operation.Kind != StaticRuleOperationBlock || operation.Voice != StaticRuleVoicePassive {
		return StaticRuleQualifier{}, 0, false
	}
	if !staticRuleWordsAt(tokens, start, "except") {
		return StaticRuleQualifier{}, 0, false
	}
	qualifier, next, ok := parseStaticExceptByQualifier(tokens, start+1, end)
	if !ok {
		return StaticRuleQualifier{}, 0, false
	}
	operation.Kind = StaticRuleOperationBlockedExcept
	operation.Span = shared.SpanOf(tokens[start-2 : start+1])
	return qualifier, next, true
}

// parseStaticExceptByQualifier consumes the blocker characteristic following
// "except by": "by creatures with flying", "by <color> creatures", "by artifact
// creatures", "by creatures with defender", or "by legendary creatures". The
// flying, color, and artifact forms reuse the shared blocker-restriction
// recognizer; defender and legendary add the keyword and supertype forms unique
// to the "except by" family. end is the exclusive bound (the period index).
func parseStaticExceptByQualifier(tokens []shared.Token, start, end int) (StaticRuleQualifier, int, bool) {
	if qualifier, next, ok := parseStaticBlockerRestrictionQualifier(tokens, start, end); ok {
		return qualifier, next, true
	}
	if staticRuleWordsAt(tokens, start, "by", "creatures", "with", "defender") {
		return StaticRuleQualifier{
			Kind: StaticRuleQualifierBlockerDefender,
			Span: shared.SpanOf(tokens[start : start+4]),
		}, start + 4, true
	}
	if staticRuleWordsAt(tokens, start, "by", "legendary", "creatures") {
		return StaticRuleQualifier{
			Kind: StaticRuleQualifierBlockerLegendary,
			Span: shared.SpanOf(tokens[start : start+3]),
		}, start + 3, true
	}
	return StaticRuleQualifier{}, 0, false
}

func parseStaticBlockerRestrictionQualifier(tokens []shared.Token, start, end int) (StaticRuleQualifier, int, bool) {
	if qualifier, next, ok := parseStaticBlockerTypeQualifier(tokens, start, end); ok {
		return qualifier, next, true
	}
	if staticRuleWordsAt(tokens, start, "by", "creatures", "the", "monarch", "controls") {
		return StaticRuleQualifier{
			Kind: StaticRuleQualifierBlockerControlledByMonarch,
			Span: shared.SpanOf(tokens[start : start+5]),
		}, start + 5, true
	}
	if !staticRuleWordsAt(tokens, start, "by", "creatures", "with") {
		return StaticRuleQualifier{}, 0, false
	}
	cursor := start + 3
	if staticRuleWordsAt(tokens, cursor, "flying") && cursor < end {
		return StaticRuleQualifier{
			Kind: StaticRuleQualifierBlockerFlying,
			Span: shared.SpanOf(tokens[start : cursor+1]),
		}, cursor + 1, true
	}
	if !staticRuleWordsAt(tokens, cursor, "power") || cursor+3 > end {
		return StaticRuleQualifier{}, 0, false
	}
	amount, ok := staticUnsignedInteger(tokens[cursor+1])
	if !ok || !staticRuleWordsAt(tokens, cursor+2, "or") {
		return StaticRuleQualifier{}, 0, false
	}
	var kind StaticRuleQualifierKind
	switch {
	case staticRuleWordsAt(tokens, cursor+3, "less"):
		kind = StaticRuleQualifierBlockerPowerOrLess
	case staticRuleWordsAt(tokens, cursor+3, "greater"):
		kind = StaticRuleQualifierBlockerPowerOrGreater
	default:
		return StaticRuleQualifier{}, 0, false
	}
	return StaticRuleQualifier{
		Kind:   kind,
		Span:   shared.SpanOf(tokens[start : cursor+4]),
		Amount: amount,
	}, cursor + 4, true
}

// staticRuleHasGuardClause reports whether a trailing condition clause gates a
// static rule, such as "unless you control seven or more lands." on Topiary
// Stomper. It returns true when start begins a recognized condition introducer
// with at least one body token following it, up to the terminating period. The
// guard's meaning is derived separately by the condition machinery; the
// static-rule parser only records its presence so the rule consumes the
// sentence.
func staticRuleHasGuardClause(tokens []shared.Token, start int) bool {
	end := len(tokens) - 1
	if start >= end {
		return false
	}
	kind, width := conditionIntroAt(tokens, start)
	if kind == ConditionIntroUnknown || start+width >= end {
		return false
	}
	return true
}

func validStaticRuleSyntax(rule StaticRuleSyntax) bool {
	switch rule.Subject.Kind {
	case StaticRuleSubjectSourceCreature, StaticRuleSubjectAttachedObject:
		return validCreatureStaticRuleOperation(rule)
	case StaticRuleSubjectSourcePermanent, StaticRuleSubjectAttachedPermanent, StaticRuleSubjectBattlefieldPermanents:
		return (rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationUntap &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
			validAttachedPermanentAttackBlockRule(rule)
	case StaticRuleSubjectSourceSpell:
		return rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationCounter &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			len(rule.Qualifiers) == 0
	case StaticRuleSubjectControlledCreatures:
		return (rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			len(rule.Qualifiers) == 0) ||
			(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
				rule.Operation.Kind == StaticRuleOperationTransform &&
				rule.Operation.Voice == StaticRuleVoiceActive &&
				len(rule.Qualifiers) == 0) ||
			validAssignDamageByToughnessRule(rule) ||
			validGroupMustAttackRule(rule)
	case StaticRuleSubjectBattlefieldCreatures:
		return (rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
			(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
				rule.Operation.Kind == StaticRuleOperationUntap &&
				rule.Operation.Voice == StaticRuleVoiceActive &&
				len(rule.Qualifiers) == 0) ||
			validAssignDamageByToughnessRule(rule) ||
			validGroupMustAttackRule(rule)
	case StaticRuleSubjectOpponentControlledCreatures:
		return validGroupMustAttackRule(rule)
	default:
		return false
	}
}

// validAttachedPermanentAttackBlockRule reports whether a rule on the permanent
// an Aura is attached to ("Enchanted permanent") is the Arrest-family pinning
// prohibition "can't attack or block, and its activated abilities can't be
// activated[ unless they're mana abilities]." Faith's Fetters and similar Auras
// name "enchanted permanent" rather than "enchanted creature", so they thread
// the attached-permanent subject here. The bare "can't attack or block" without
// the activation clause stays the attached-object creature subject's job.
func validAttachedPermanentAttackBlockRule(rule StaticRuleSyntax) bool {
	return rule.Subject.Kind == StaticRuleSubjectAttachedPermanent &&
		rule.Constraint.Kind == StaticRuleConstraintProhibition &&
		rule.Operation.Kind == StaticRuleOperationAttackOrBlock &&
		rule.Operation.Voice == StaticRuleVoiceActive &&
		(staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierCantActivateAbilities) ||
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierCantActivateNonManaAbilities))
}

// validGroupMustAttackRule reports whether a group-scoped static rule is the
// forced-attack requirement "<group> attack[s] each combat if able."
func validGroupMustAttackRule(rule StaticRuleSyntax) bool {
	return rule.Constraint.Kind == StaticRuleConstraintRequirement &&
		rule.Operation.Kind == StaticRuleOperationAttack &&
		rule.Operation.Voice == StaticRuleVoiceActive &&
		staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble)
}

// validAssignDamageByToughnessRule reports whether a static rule is the
// combat-damage replacement "<subject> assigns combat damage equal to its
// toughness rather than its power."
func validAssignDamageByToughnessRule(rule StaticRuleSyntax) bool {
	return rule.Constraint.Kind == StaticRuleConstraintRequirement &&
		rule.Operation.Kind == StaticRuleOperationAssignDamageByToughness &&
		rule.Operation.Voice == StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0
}

// validCreatureStaticRuleOperation reports whether a creature-scoped static rule
// (a creature source or the creature an Aura or Equipment is attached to) carries
// a recognized constraint, operation, voice, and qualifier set.
func validCreatureStaticRuleOperation(rule StaticRuleSyntax) bool {
	return (rule.Constraint.Kind == StaticRuleConstraintProhibition &&
		rule.Operation.Kind == StaticRuleOperationBlock &&
		rule.Operation.Voice == StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			(len(rule.Qualifiers) == 0 ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierByMoreThanOne) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerFlying) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerPowerOrLess) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerPowerOrGreater) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerColor) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerArtifact) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerControlledByMonarch))) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlockedExcept &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			(staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerFlying) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerColor) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerArtifact) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerDefender) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerLegendary))) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationAttack &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			(len(rule.Qualifiers) == 0 ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierDefenderYou) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierAlone))) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationAttackOrBlock &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			(len(rule.Qualifiers) == 0 ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierAlone) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierCantActivateAbilities) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierCantActivateNonManaAbilities))) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierAlone)) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlockAndBeBlocked &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationUntap &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationTransform &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationAttack &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble)) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierIfAble)) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockedAttackerFlying)) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierAdditionalCreature)) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationBlockedByAll &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationAssignDamageAsUnblocked &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationGoaded &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
		validAssignDamageByToughnessRule(rule)
}

func staticRuleQualifiersAre(qualifiers []StaticRuleQualifier, kinds ...StaticRuleQualifierKind) bool {
	actual := make([]StaticRuleQualifierKind, len(qualifiers))
	for i := range qualifiers {
		actual[i] = qualifiers[i].Kind
	}
	return slices.Equal(actual, kinds)
}

func staticRuleWordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		token := tokens[start+i]
		if token.Kind != shared.Word || !strings.EqualFold(token.Text, word) {
			return false
		}
	}
	return true
}
