package parser

import "strings"

// disjunctiveTriggerJoin separates the alternative trigger conditions of a
// multi-event triggered ability, as in "Whenever <A>, or <B>, or <C>, <effect>".
const disjunctiveTriggerJoin = ", or "

// expandDisjunctiveTrigger rewrites a triggered ability whose single shared
// effect responds to two or more alternative events — "(When|Whenever) <A>, or
// <B>[, or <C>...], <effect>" — into one independent triggered ability per
// condition, each sharing the effect text: "(When|Whenever) <A>, <effect>",
// "Whenever <B>, <effect>", and so on. Each trigger condition is its own
// triggered ability (CR 603.1), so emitting one ability per condition lets the
// standard trigger pipeline lower each constituent independently (Syr Konrad,
// the Grim and other multi-event graveyard payoffs). The rewrite is parser-owned
// because it is a wording substitution; downstream stages see ordinary
// triggered abilities.
func expandDisjunctiveTrigger(source string) string {
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		expanded, ok := splitDisjunctiveTriggerLine(line)
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

// splitDisjunctiveTriggerLine splits a "(When|Whenever) <A>, or <B>[, or <C>],
// <effect>" line into one triggered-ability line per condition, all sharing the
// post-final-condition effect text. It reports ok only when the line is a
// "When"/"Whenever" trigger (after an optional "<word> — " ability word) whose
// alternative conditions are comma-free and whose final condition is followed by
// a comma-delimited shared effect body.
func splitDisjunctiveTriggerLine(line string) (lines []string, ok bool) {
	segments := strings.Split(line, disjunctiveTriggerJoin)
	if len(segments) < 2 {
		return nil, false
	}
	prefix, introduction := splitAbilityWordPrefix(segments[0])
	if !strings.HasPrefix(introduction, "When ") && !strings.HasPrefix(introduction, "Whenever ") {
		return nil, false
	}
	if strings.Contains(introduction, ",") {
		return nil, false
	}
	middle := segments[1 : len(segments)-1]
	for _, condition := range middle {
		if condition == "" || strings.Contains(condition, ",") {
			return nil, false
		}
	}
	last := segments[len(segments)-1]
	finalCondition, body, ok := strings.Cut(last, ",")
	if !ok {
		return nil, false
	}
	finalCondition = strings.TrimSpace(finalCondition)
	body = strings.TrimSpace(body)
	if finalCondition == "" || body == "" {
		return nil, false
	}
	lines = append(lines, segments[0]+", "+body)
	for _, condition := range middle {
		lines = append(lines, prefix+"Whenever "+strings.TrimSpace(condition)+", "+body)
	}
	lines = append(lines, prefix+"Whenever "+finalCondition+", "+body)
	return lines, true
}
