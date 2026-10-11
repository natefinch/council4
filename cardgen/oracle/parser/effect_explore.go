package parser

import "strings"

func exactExploreEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != EffectDurationNone {
		return false
	}
	var subject string
	var ok bool
	switch effect.Context {
	case EffectContextSource:
		if len(effect.Targets) == 0 {
			subject, ok = exactSelfSubjectReferenceText(effect.SubjectReferences)
		}
	case EffectContextReferencedObject:
		if len(effect.Targets) == 0 && len(effect.SubjectReferences) == 1 &&
			effect.SubjectReferences[0].Kind == ReferencePronoun &&
			effect.SubjectReferences[0].Pronoun == PronounIt {
			subject, ok = effect.SubjectReferences[0].Text, true
		}
	case EffectContextTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact {
			subject, ok = effect.Targets[0].Text, true
		}
	default:
		return false
	}
	if !ok {
		return false
	}
	action := subject + " explores"
	return strings.EqualFold(exactEffectClauseText(effect), action+".") ||
		exactRepeatedActionSyntax(effect, action)
}
