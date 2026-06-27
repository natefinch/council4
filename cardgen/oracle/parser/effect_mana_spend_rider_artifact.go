package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// artifactSpendClauseKind classifies one comma- or "or"-separated clause of an
// artifact mana-spend restriction sentence.
type artifactSpendClauseKind int

const (
	artifactSpendClauseNone artifactSpendClauseKind = iota
	// artifactSpendClauseCastArtifact is "cast an artifact spell" / "cast
	// artifact spells".
	artifactSpendClauseCastArtifact
	// artifactSpendClauseActivateArtifact is "activate an ability of an artifact
	// [source]" / "activate abilities of artifacts [sources]".
	artifactSpendClauseActivateArtifact
	// artifactSpendClauseActivateAny is the unqualified "activate an ability" /
	// "activate abilities".
	artifactSpendClauseActivateAny
)

// artifactSpendPrefixWords is the shared restriction prefix every artifact
// mana-spend restriction begins with.
var artifactSpendPrefixWords = []string{"spend", "this", "mana", "only"}

// recognizeArtifactManaSpendRider reports whether the sentence tokens are an
// artifact-restricted mana-spend restriction and, if so, returns its typed
// syntax. It recognizes a restriction sentence "Spend this mana only <clause>
// [or <clause>]." whose clauses are drawn from casting an artifact spell,
// activating an ability of an artifact, and activating any ability, then maps
// the recognized clause set to one closed condition:
//
//   - cast an artifact spell only → ManaSpendCastArtifactSpell (Castle Doom,
//     Mishra's Workshop).
//   - activate an ability of an artifact only → ManaSpendActivateArtifactAbility
//     (Soldevi Machinist).
//   - cast an artifact spell + activate an ability of an artifact →
//     ManaSpendCastOrActivateArtifact (Power Depot, Cargo Ship).
//   - cast an artifact spell + activate any ability →
//     ManaSpendCastArtifactOrActivateAbility (Guidelight Optimizer, Automated
//     Artificer).
//
// Any other clause, clause combination, or trailing content fails closed.
func recognizeArtifactManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	prefix := len(artifactSpendPrefixWords)
	if len(tokens) <= prefix || !effectWordsAt(tokens, 0, artifactSpendPrefixWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	index := prefix
	var castArtifact, activateArtifact, activateAny bool
	clauses := 0
	for {
		kind, next, ok := artifactSpendClauseAt(tokens, index)
		if !ok {
			return ManaSpendRiderSyntax{}, false
		}
		switch kind {
		case artifactSpendClauseCastArtifact:
			if castArtifact {
				return ManaSpendRiderSyntax{}, false
			}
			castArtifact = true
		case artifactSpendClauseActivateArtifact:
			if activateArtifact {
				return ManaSpendRiderSyntax{}, false
			}
			activateArtifact = true
		case artifactSpendClauseActivateAny:
			if activateAny {
				return ManaSpendRiderSyntax{}, false
			}
			activateAny = true
		default:
			return ManaSpendRiderSyntax{}, false
		}
		index = next
		clauses++
		if clauses > 2 {
			return ManaSpendRiderSyntax{}, false
		}
		if index < len(tokens) && equalWord(tokens[index], "or") {
			index++
			if index < len(tokens) && equalWord(tokens[index], "to") {
				index++
			}
			continue
		}
		break
	}
	for i := index; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	condition, ok := artifactSpendCondition(castArtifact, activateArtifact, activateAny)
	if !ok {
		return ManaSpendRiderSyntax{}, false
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:index]),
		Condition:     condition,
		Effect:        ManaSpendRiderEffectUnknown,
		Restricted:    true,
	}, true
}

// artifactSpendCondition maps a recognized clause set to its closed condition,
// failing closed on any unmodeled combination (no clauses, activate-any alone,
// both activate forms together).
func artifactSpendCondition(castArtifact, activateArtifact, activateAny bool) (ManaSpendConditionKind, bool) {
	switch {
	case castArtifact && !activateArtifact && !activateAny:
		return ManaSpendCastArtifactSpell, true
	case !castArtifact && activateArtifact && !activateAny:
		return ManaSpendActivateArtifactAbility, true
	case castArtifact && activateArtifact && !activateAny:
		return ManaSpendCastOrActivateArtifact, true
	case castArtifact && !activateArtifact && activateAny:
		return ManaSpendCastArtifactOrActivateAbility, true
	default:
		return ManaSpendConditionUnknown, false
	}
}

// artifactSpendClauseAt parses one restriction clause beginning at start,
// skipping an optional leading "to". It returns the clause kind and the index
// just past the clause. Unknown clauses fail closed.
func artifactSpendClauseAt(tokens []shared.Token, start int) (artifactSpendClauseKind, int, bool) {
	i := start
	if i < len(tokens) && equalWord(tokens[i], "to") {
		i++
	}
	if i >= len(tokens) {
		return artifactSpendClauseNone, 0, false
	}
	switch {
	case equalWord(tokens[i], "cast"):
		return artifactSpendCastClause(tokens, i+1)
	case equalWord(tokens[i], "activate"):
		return artifactSpendActivateClause(tokens, i+1)
	default:
		return artifactSpendClauseNone, 0, false
	}
}

// artifactSpendCastClause parses "an? artifact spell(s)?" after "cast".
func artifactSpendCastClause(tokens []shared.Token, i int) (artifactSpendClauseKind, int, bool) {
	if i < len(tokens) && equalWord(tokens[i], "an") {
		i++
	}
	if i >= len(tokens) || !equalWord(tokens[i], "artifact") {
		return artifactSpendClauseNone, 0, false
	}
	i++
	if i >= len(tokens) || (!equalWord(tokens[i], "spell") && !equalWord(tokens[i], "spells")) {
		return artifactSpendClauseNone, 0, false
	}
	return artifactSpendClauseCastArtifact, i + 1, true
}

// artifactSpendActivateClause parses "(an ability|abilities) (of an? artifact
// source(s)?)?" after "activate".
func artifactSpendActivateClause(tokens []shared.Token, i int) (artifactSpendClauseKind, int, bool) {
	switch {
	case effectWordsAt(tokens, i, "an", "ability"):
		i += 2
	case i < len(tokens) && equalWord(tokens[i], "abilities"):
		i++
	default:
		return artifactSpendClauseNone, 0, false
	}
	if i >= len(tokens) || !equalWord(tokens[i], "of") {
		return artifactSpendClauseActivateAny, i, true
	}
	i++
	if i < len(tokens) && equalWord(tokens[i], "an") {
		i++
	}
	if i >= len(tokens) || (!equalWord(tokens[i], "artifact") && !equalWord(tokens[i], "artifacts")) {
		return artifactSpendClauseNone, 0, false
	}
	i++
	if i < len(tokens) && (equalWord(tokens[i], "source") || equalWord(tokens[i], "sources")) {
		i++
	}
	return artifactSpendClauseActivateArtifact, i, true
}
