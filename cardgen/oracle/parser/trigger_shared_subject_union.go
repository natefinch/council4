package parser

import (
	"slices"
	"strings"
)

// sharedSubjectUnionVerbs lists the bare combat and targeting verb phrases that a
// shared-subject comma-list trigger union distributes a single subject across, as
// in "Whenever this creature attacks, blocks, or becomes the target of a spell,
// <effect>" (Giggling Skitterspike). Each verb phrase names an event that the
// standard trigger pipeline already lowers on its own; the union is a shared
// subject applied to two or more of them (CR 603.1: each condition is its own
// triggered ability). Longer phrases precede their prefixes so subject extraction
// matches the longest verb.
var sharedSubjectUnionVerbs = []string{
	"becomes the target of a spell or ability",
	"becomes the target of a spell",
	"becomes the target of an ability",
	"becomes blocked",
	"attacks",
	"blocks",
}

// expandSharedSubjectTriggerUnion rewrites a triggered ability whose single shared
// effect responds to a shared-subject comma list of combat or targeting events —
// "(When|Whenever) <subject> <verbA>, <verbB>[, <verbC>...], or <verbN>, <effect>"
// — into one independent triggered ability per event, each naming the shared
// subject and sharing the effect text: "Whenever <subject> <verbA>, <effect>",
// "Whenever <subject> <verbB>, <effect>", and so on. This complements
// expandDisjunctiveTrigger, which handles the ", or "-joined form whose
// alternatives are each self-contained; here the alternatives are bare verb
// phrases that inherit one subject. The rewrite is parser-owned because it is a
// wording substitution; downstream stages see ordinary triggered abilities.
func expandSharedSubjectTriggerUnion(source string) string {
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		expanded, ok := splitSharedSubjectTriggerUnionLine(line)
		if !ok {
			out = append(out, line)
			continue
		}
		out = append(out, expanded...)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(out, "\n")
}

// splitSharedSubjectTriggerUnionLine splits a "(When|Whenever) <subject> <verbA>,
// <verbB>[, <verbC>], or <verbN>, <effect>" line into one triggered-ability line
// per event, each naming the shared subject and the shared effect body. It reports
// ok only when the line is a "When"/"Whenever" trigger (after an optional ability
// word prefix) whose first condition ends with a recognized verb phrase, whose
// remaining conditions are each a bare recognized verb phrase, whose final
// condition carries the "or" of the comma list, and which is followed by a
// non-empty comma-delimited shared effect body.
func splitSharedSubjectTriggerUnionLine(line string) (lines []string, ok bool) {
	prefix, introduction := splitAbilityWordPrefix(line)
	var keyword string
	switch {
	case strings.HasPrefix(introduction, "Whenever "):
		keyword = "Whenever "
	case strings.HasPrefix(introduction, "When "):
		keyword = "When "
	default:
		return nil, false
	}
	rest := introduction[len(keyword):]
	segments := strings.Split(rest, ", ")
	if len(segments) < 4 {
		// A shared-subject union needs at least the subject-bearing first
		// condition, one bare middle condition, the "or"-marked final condition,
		// and the effect body ("<subject> A, B, or C, <effect>").
		return nil, false
	}
	subject, firstVerb, ok := splitSubjectVerb(segments[0])
	if !ok {
		return nil, false
	}
	conditions := []string{firstVerb}
	sawFinal := false
	index := 1
	for ; index < len(segments); index++ {
		segment := segments[index]
		verb := strings.TrimPrefix(segment, "or ")
		final := verb != segment
		if !isSharedSubjectUnionVerb(verb) {
			break
		}
		conditions = append(conditions, verb)
		if final {
			sawFinal = true
			index++
			break
		}
	}
	if !sawFinal || len(conditions) < 3 {
		return nil, false
	}
	body := strings.TrimSpace(strings.Join(segments[index:], ", "))
	if body == "" {
		return nil, false
	}
	for _, condition := range conditions {
		lines = append(lines, prefix+"Whenever "+subject+" "+condition+", "+body)
	}
	return lines, true
}

// splitSubjectVerb splits a shared-subject union's first condition into its shared
// subject and its bare verb phrase, choosing the longest recognized verb phrase
// that the condition ends with ("this creature attacks" -> "this creature",
// "attacks"). It reports ok only when a non-empty subject precedes the verb.
func splitSubjectVerb(condition string) (subject, verb string, ok bool) {
	for _, candidate := range sharedSubjectUnionVerbs {
		if trimmed, found := strings.CutSuffix(condition, " "+candidate); found {
			subject = strings.TrimSpace(trimmed)
			if subject == "" {
				return "", "", false
			}
			return subject, candidate, true
		}
	}
	return "", "", false
}

// isSharedSubjectUnionVerb reports whether phrase is exactly one of the bare verb
// phrases a shared-subject union distributes a subject across.
func isSharedSubjectUnionVerb(phrase string) bool {
	return slices.Contains(sharedSubjectUnionVerbs, phrase)
}
