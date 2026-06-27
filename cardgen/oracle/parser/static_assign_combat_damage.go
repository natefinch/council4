package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// staticGroupAssignsCombatDamageSubject recognizes the creature-group subject of
// a static combat-damage replacement ("Creatures you control assign combat damage
// ...", "Each creature you control assigns ...", "All creatures assign ...",
// "Each creature assigns ...") when it precedes the verb "assign"/"assigns".
// parseEffectStaticSubject only delimits these groups before a get/have/gain/lose
// verb, so the combat-damage boundary is recognized here. Only the unfiltered
// group forms are supported; any leading or trailing filter fails closed.
func staticGroupAssignsCombatDamageSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, int, bool) {
	forms := []struct {
		words []string
		kind  EffectStaticSubjectKind
	}{
		{[]string{"each", "creature", "you", "control"}, EffectStaticSubjectControlledCreatures},
		{[]string{"creatures", "you", "control"}, EffectStaticSubjectControlledCreatures},
		{[]string{"each", "creature"}, EffectStaticSubjectAllCreatures},
		{[]string{"all", "creatures"}, EffectStaticSubjectAllCreatures},
	}
	for _, form := range forms {
		if !staticWordsAt(tokens, 0, form.words...) {
			continue
		}
		width := len(form.words)
		if !staticWordsAt(tokens, width, "assigns") && !staticWordsAt(tokens, width, "assign") {
			return EffectStaticSubjectSyntax{}, 0, false
		}
		return EffectStaticSubjectSyntax{Kind: form.kind, Span: shared.SpanOf(tokens[:width])}, width, true
	}
	return EffectStaticSubjectSyntax{}, 0, false
}

// parseStaticAssignsCombatDamageRuleOperation recognizes the combat-damage
// replacement operation "assign[s] combat damage equal to <its> toughness rather
// than <its> power", the static rule that makes the subject creatures assign
// combat damage equal to their toughness instead of their power. The possessive
// pronouns ("its"/"his"/"her"/"their") agree with the printed creature. Any
// trailing rider or differing wording fails closed.
func parseStaticAssignsCombatDamageRuleOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	verb := index
	if !staticWordsAt(tokens, verb, "assigns") && !staticWordsAt(tokens, verb, "assign") {
		return StaticDeclarationSyntax{}, 0, false
	}
	next := verb + 1
	if !staticWordsAt(tokens, next, "combat", "damage", "equal", "to") {
		return StaticDeclarationSyntax{}, 0, false
	}
	next += 4
	if !staticPossessivePronounAt(tokens, next) {
		return StaticDeclarationSyntax{}, 0, false
	}
	next++
	if !staticWordsAt(tokens, next, "toughness", "rather", "than") {
		return StaticDeclarationSyntax{}, 0, false
	}
	next += 3
	if !staticPossessivePronounAt(tokens, next) {
		return StaticDeclarationSyntax{}, 0, false
	}
	next++
	if !staticWordsAt(tokens, next, "power") || next+1 != end {
		return StaticDeclarationSyntax{}, 0, false
	}
	next++
	return staticRuleOperation(tokens, index, next, subject,
		StaticRuleConstraint{Kind: StaticRuleConstraintRequirement, Span: tokens[verb].Span},
		StaticRuleOperation{
			Kind:  StaticRuleOperationAssignDamageByToughness,
			Voice: StaticRuleVoiceActive,
			Span:  shared.SpanOf(tokens[verb : verb+3]),
		},
		nil,
	)
}

// staticPossessivePronounAt reports whether the token at index is a third-person
// possessive pronoun ("its", "his", "her", "their").
func staticPossessivePronounAt(tokens []shared.Token, index int) bool {
	return staticWordsAt(tokens, index, "its") ||
		staticWordsAt(tokens, index, "his") ||
		staticWordsAt(tokens, index, "her") ||
		staticWordsAt(tokens, index, "their")
}
