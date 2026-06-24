package parser

import "strings"

// diesOrExileMarker is the verb disjunction shared by reanimation Auras whose
// recursion trigger fires when the enchanted permanent leaves the battlefield by
// death or by exile, as in Kaya's Ghostform.
const diesOrExileMarker = " dies or is put into exile"

// expandDiesOrExileTrigger rewrites a triggered ability whose condition is the
// verb disjunction "<subject> dies or is put into exile[ from the battlefield],
// <effect>" into two independent triggered abilities sharing the effect text:
// "<subject> dies, <effect>" and "Whenever <subject> is put into exile[ from the
// battlefield], <effect>". Each leave-the-battlefield event is its own trigger
// condition (CR 603.1), so emitting one ability per event lets the standard
// trigger pipeline lower each constituent independently. The rewrite is
// parser-owned because it is a wording substitution; downstream stages see
// ordinary triggered abilities.
func expandDiesOrExileTrigger(source string) string {
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		expanded, ok := splitDiesOrExileTriggerLine(line)
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

// splitDiesOrExileTriggerLine splits a "(When|Whenever) <subject> dies or is put
// into exile[ from the battlefield], <effect>" line into a death trigger and an
// exile trigger that share the post-condition effect text. It reports ok only
// when the line is a "When"/"Whenever" trigger (after an optional "<word> — "
// ability word) whose condition is exactly the dies-or-exile disjunction and
// whose effect body follows a single delimiting comma.
func splitDiesOrExileTriggerLine(line string) (lines []string, ok bool) {
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
	subject, remainder, found := strings.Cut(strings.TrimPrefix(introduction, introWord), diesOrExileMarker)
	if !found {
		return nil, false
	}
	if subject == "" || strings.Contains(subject, ",") {
		return nil, false
	}
	zone := ""
	if after, cut := strings.CutPrefix(remainder, " from the battlefield"); cut {
		zone = " from the battlefield"
		remainder = after
	}
	body, hasBody := strings.CutPrefix(remainder, ", ")
	if !hasBody || body == "" || strings.HasPrefix(body, " ") {
		return nil, false
	}
	lines = append(lines, prefix+introWord+subject+" dies, "+body)
	lines = append(lines, prefix+"Whenever "+subject+" is put into exile"+zone+", "+body)
	return lines, true
}
