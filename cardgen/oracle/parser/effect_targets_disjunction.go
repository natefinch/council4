package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseQualifiedDisjunctivePermanentTarget recognizes a single permanent target
// whose noun phrase is an Oxford-comma disjunction of permanent sub-selections
// where at least one member carries a qualifier the flattened card-type union
// cannot express: a tapped/combat state, a keyword, or a power/toughness/mana-
// value comparison ("artifact, enchantment, or creature with flying", "artifact,
// enchantment, or creature with power 4 or greater", "artifact, enchantment, or
// tapped creature an opponent controls"). Each member becomes its own
// Selection.Alternatives entry so the lowering builds a Selection.AnyOf, letting
// the qualified creature member keep its filter instead of being flattened into
// a lossy card-type union.
//
// It fires only for the mandatory single target and the "up to one" optional
// target, only when the reconstructed phrase round-trips the printed Oracle text
// byte for byte, and only when at least one member is qualified. A disjunction of
// bare card types ("artifact, enchantment, or creature") stays on the existing
// flattened-union path, so this production never changes an already-supported
// card. Every other shape fails closed.
func parseQualifiedDisjunctivePermanentTarget(
	tokens []shared.Token,
	atoms Atoms,
	start, targetIndex int,
	cardinality TargetCardinalitySyntax,
) (TargetSyntax, bool) {
	prefix, ok := disjunctiveTargetDeterminerPrefix(cardinality)
	if !ok {
		return TargetSyntax{}, false
	}
	selEnd := disjunctiveTargetSelectionEnd(tokens, targetIndex+1)
	segments, ok := splitTopLevelOxfordList(tokens[targetIndex+1 : selEnd])
	if !ok {
		return TargetSyntax{}, false
	}
	members := make([]SelectionSyntax, 0, len(segments))
	controller := SelectionControllerAny
	qualified := false
	for index, segment := range segments {
		member := parseSelection(segment, atoms)
		if member.Controller != SelectionControllerAny {
			// A controller clause ("an opponent controls", "you control") trails
			// the whole list, so it may appear only on the final member; lifting
			// it to the parent applies it to every alternative.
			if index != len(segments)-1 {
				return TargetSyntax{}, false
			}
			controller = member.Controller
			member.Controller = SelectionControllerAny
		}
		words, ok := disjunctMemberWords(member)
		if !ok {
			return TargetSyntax{}, false
		}
		if !memberIsBareCardType(member, words) {
			qualified = true
		}
		members = append(members, disjunctMemberSelection(member))
	}
	if !qualified {
		return TargetSyntax{}, false
	}
	targetTokens := tokens[start:selEnd]
	expected, ok := disjunctiveTargetExactText(prefix, members, controller)
	if !ok || !strings.EqualFold(joinedEffectText(targetTokens), expected) {
		return TargetSyntax{}, false
	}
	selectionTokens := tokens[targetIndex+1 : selEnd]
	return TargetSyntax{
		Span:        shared.SpanOf(targetTokens),
		Text:        joinedEffectText(targetTokens),
		Cardinality: cardinality,
		Selection: SelectionSyntax{
			Span:         shared.SpanOf(selectionTokens),
			Text:         joinedEffectText(selectionTokens),
			Kind:         SelectionPermanent,
			Controller:   controller,
			Alternatives: members,
		},
		Exact: true,
	}, true
}

// disjunctiveTargetDeterminerPrefix returns the determiner words that precede the
// disjunction for a supported cardinality: the mandatory single target ("target")
// and the optional "up to one" target ("up to one target"). Every other
// cardinality fails closed, so plural and unbounded slots never reach the
// single-target reconstruction.
func disjunctiveTargetDeterminerPrefix(cardinality TargetCardinalitySyntax) (string, bool) {
	switch cardinality {
	case TargetCardinalitySyntax{Min: 1, Max: 1}:
		return "target", true
	case TargetCardinalitySyntax{Min: 0, Max: 1}:
		return "up to one target", true
	default:
		return "", false
	}
}

// disjunctiveTargetSelectionEnd returns the index just past the disjunction's
// noun phrase: the first top-level clause terminator (period or semicolon) at or
// after from, or the end of the sentence tokens. parseTargets receives a single
// sentence, so the destroy/exile object runs to that terminator.
func disjunctiveTargetSelectionEnd(tokens []shared.Token, from int) int {
	depth := 0
	for i := from; i < len(tokens); i++ {
		switch tokens[i].Kind {
		case shared.LeftParen:
			depth++
		case shared.RightParen:
			if depth > 0 {
				depth--
			}
		case shared.Period, shared.Semicolon:
			if depth == 0 {
				return i
			}
		default:
		}
	}
	return len(tokens)
}

// splitTopLevelOxfordList splits a noun phrase into the members of an Oxford-comma
// disjunction ("artifact, enchantment, or creature with flying"). Members are
// separated by top-level commas, and the final member is introduced by "or",
// which the split strips. It returns ok=false unless the phrase holds at least
// two comma-separated members and the final member begins with "or", so a bare
// "X or Y" union (no comma) and a single noun keep their existing parses.
func splitTopLevelOxfordList(tokens []shared.Token) ([][]shared.Token, bool) {
	var segments [][]shared.Token
	depth := 0
	segStart := 0
	for i, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			depth++
		case shared.RightParen:
			if depth > 0 {
				depth--
			}
		case shared.Comma:
			if depth == 0 {
				segments = append(segments, tokens[segStart:i])
				segStart = i + 1
			}
		default:
		}
	}
	segments = append(segments, tokens[segStart:])
	if len(segments) < 2 {
		return nil, false
	}
	last := segments[len(segments)-1]
	if len(last) == 0 || !equalWord(last[0], "or") {
		return nil, false
	}
	segments[len(segments)-1] = last[1:]
	for _, segment := range segments {
		if len(segment) == 0 {
			return nil, false
		}
	}
	return segments, true
}

// disjunctMemberWords reconstructs the canonical Oracle words for one disjunction
// member ("artifact", "tapped creature", "creature with flying"). It fails closed
// unless the member is a permanent sub-selection the single-permanent qualifier
// machinery can express and names a non-empty noun phrase, so an inexpressible or
// player/spell member can never become an alternative.
func disjunctMemberWords(member SelectionSyntax) ([]string, bool) {
	if len(member.Alternatives) != 0 || !disjunctSideExpressible(member) {
		return nil, false
	}
	words, ok := permanentSelectionQualifierWords(member)
	if !ok || len(words) == 0 {
		return nil, false
	}
	return words, true
}

// memberIsBareCardType reports whether a member is an unqualified card-type noun
// ("artifact", "creature"), the signal that the whole disjunction is a plain
// card-type union the existing flattened-union path already supports. A bare
// member reconstructs to exactly its card-type noun.
func memberIsBareCardType(member SelectionSyntax, words []string) bool {
	if len(words) != 1 {
		return false
	}
	noun, ok := permanentSelectionNoun(member.Kind)
	return ok && words[0] == noun
}

// disjunctMemberSelection returns a clean alternative selection carrying only the
// member's type and qualifier dimensions, with the controller already lifted to
// the parent and the source span and text cleared so the alternative compiles to
// a controller-free sub-selector.
func disjunctMemberSelection(member SelectionSyntax) SelectionSyntax {
	member.Controller = SelectionControllerAny
	member.Span = shared.Span{}
	member.Text = ""
	return member
}

// disjunctiveTargetExactText reconstructs the canonical Oracle phrase for the
// whole disjunctive target from its determiner prefix, members, and shared
// controller clause ("target artifact, enchantment, or creature with flying",
// "up to one target artifact, enchantment, or tapped creature an opponent
// controls"). It fails closed if any member's words cannot be reconstructed.
func disjunctiveTargetExactText(prefix string, members []SelectionSyntax, controller SelectionController) (string, bool) {
	memberTexts := make([]string, 0, len(members))
	for i := range members {
		words, ok := permanentSelectionQualifierWords(members[i])
		if !ok || len(words) == 0 {
			return "", false
		}
		memberTexts = append(memberTexts, strings.Join(words, " "))
	}
	last := len(memberTexts) - 1
	body := strings.Join(memberTexts[:last], ", ") + ", or " + memberTexts[last]
	return targetControllerSuffix(prefix+" "+body, controller)
}
