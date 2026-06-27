package parser

import "strings"

// entersOrLeavesBattlefieldMarker is the verb disjunction shared by permanents
// whose ability fires both when the named subject enters the battlefield and
// when it leaves the battlefield, as in Brandywine Farmer and Sundering Titan.
const entersOrLeavesBattlefieldMarker = " enters or leaves the battlefield"

// expandEntersOrLeavesBattlefieldTrigger rewrites a triggered ability whose
// condition is the verb disjunction "<subject> enters or leaves the
// battlefield, <effect>" into two independent triggered abilities sharing the
// effect text: "<subject> enters, <effect>" and "Whenever <subject> leaves the
// battlefield, <effect>". Entering the battlefield and leaving the battlefield
// are distinct events, and each trigger condition is its own triggered ability
// (CR 603.1), so emitting one ability per event lets the standard trigger
// pipeline lower each constituent independently. The rewrite is parser-owned
// because it is a wording substitution; downstream stages see ordinary
// triggered abilities.
func expandEntersOrLeavesBattlefieldTrigger(source string) string {
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		expanded, ok := splitEntersOrLeavesBattlefieldTriggerLine(line)
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

// splitEntersOrLeavesBattlefieldTriggerLine splits a "(When|Whenever) <subject>
// enters or leaves the battlefield, <effect>" line into an enters-the-
// battlefield trigger and a leaves-the-battlefield trigger that share the
// post-condition effect text. It reports ok only when the line is a
// "When"/"Whenever" trigger (after an optional "<word> — " ability word) whose
// condition is exactly the enters-or-leaves disjunction and whose effect body
// follows a single delimiting comma.
func splitEntersOrLeavesBattlefieldTriggerLine(line string) (lines []string, ok bool) {
	prefix, introduction := splitAbilityWordPrefix(line)
	introWord := ""
	switch {
	case strings.HasPrefix(introduction, "When "):
		introWord = "When "
	case strings.HasPrefix(introduction, "Whenever "):
		introWord = "Whenever "
	default:
		return nil, false
	}
	subject, remainder, found := strings.Cut(
		strings.TrimPrefix(introduction, introWord), entersOrLeavesBattlefieldMarker)
	if !found {
		return nil, false
	}
	if subject == "" || strings.Contains(subject, ",") {
		return nil, false
	}
	body, hasBody := strings.CutPrefix(remainder, ", ")
	if !hasBody || body == "" || strings.HasPrefix(body, " ") {
		return nil, false
	}
	lines = append(lines, prefix+introWord+subject+" enters, "+body)
	lines = append(lines, prefix+"Whenever "+subject+" leaves the battlefield, "+body)
	return lines, true
}
