package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// recognizeStaticAssignCombatDamageByToughnessDeclarations maps a standalone
// combat-damage replacement rule onto a closed semantic declaration, e.g. "Each
// creature you control assigns combat damage equal to its toughness rather than
// its power." (Assault Formation), "Each creature assigns combat damage equal to
// its toughness rather than its power." (Doran, the Siege Tower), or "This
// creature assigns combat damage equal to its toughness rather than its power."
// The affected group derives entirely from the typed parser rule subject: the
// source subject affects the object itself, the controlled-creatures subject
// yields a controller-permanents group restricted to creatures, and the
// every-creature subject yields a battlefield group. Costs, triggers,
// conditions, or any resolving content fail closed because a continuous group
// rule carries none.
func recognizeStaticAssignCombatDamageByToughnessDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationRule) {
		return nil, false
	}
	ruleNode := &statics[0]
	switch ruleNode.Rule.Subject.Kind {
	case parser.StaticRuleSubjectSourceCreature,
		parser.StaticRuleSubjectControlledCreatures,
		parser.StaticRuleSubjectBattlefieldCreatures:
	default:
		return nil, false
	}
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok || rule != StaticRuleAssignsCombatDamageByToughness {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(ruleNode.Subject)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 {
		return nil, false
	}
	declaration := staticRuleDeclaration(ability.Span, group.Span, ruleNode.OperationSpan, rule, zone, group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil)
	declaration.Group = group
	return []StaticDeclaration{declaration}, true
}
