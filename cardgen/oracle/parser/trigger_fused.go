package parser

import "strings"

// fusedTriggerJoin is the conjunction that joins a fused triggered ability's two
// trigger conditions, as in "When ~ enters and whenever you cast a spell, ...".
const fusedTriggerJoin = " and whenever "

// abilityWordSeparator is the " em dash " that follows an ability word, e.g.
// "Eerie — Whenever an enchantment you control enters and whenever ...".
const abilityWordSeparator = " — "

// expandFusedTrigger rewrites a fused triggered ability — one that joins two
// trigger conditions under a single shared effect, "(When|Whenever) <A> and
// whenever <B>, <effect>" — into two independent triggered abilities that share
// the effect text: "(When|Whenever) <A>, <effect>" and "Whenever <B>, <effect>".
// Each trigger condition is its own triggered ability (CR 603.1), so emitting one
// ability per condition lets the standard trigger pipeline lower each constituent
// independently. This is the broad "enters and whenever" family (Orcish
// Bowmasters, Up the Beanstalk, the Eerie Room cards, Titans' Vanguard, ...). The
// rewrite is parser-owned because it is a wording substitution; downstream stages
// see two ordinary triggered abilities.
func expandFusedTrigger(source string) string {
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		first, second, ok := splitFusedTriggerLine(line)
		if !ok {
			out = append(out, line)
			continue
		}
		out = append(out, first, second)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(out, "\n")
}

// splitFusedTriggerLine splits a fused two-condition triggered ability line into
// its two constituent triggered-ability lines. It reports ok only when the line
// is a trigger introduced by "When"/"Whenever" (after an optional "<word> — "
// ability word) whose " and whenever " join precedes the body comma, so both
// constituents share the post-comma effect text.
func splitFusedTriggerLine(line string) (first, second string, ok bool) {
	comma := strings.Index(line, ",")
	if comma < 0 {
		return "", "", false
	}
	join := strings.Index(line, fusedTriggerJoin)
	if join < 0 || join >= comma {
		return "", "", false
	}
	head := line[:join]
	condition := strings.TrimSpace(line[join+len(fusedTriggerJoin) : comma])
	body := strings.TrimSpace(line[comma+1:])
	if condition == "" || body == "" {
		return "", "", false
	}
	prefix, introduction := splitAbilityWordPrefix(head)
	if !strings.HasPrefix(introduction, "When ") && !strings.HasPrefix(introduction, "Whenever ") {
		return "", "", false
	}
	first = head + ", " + body
	second = prefix + "Whenever " + condition + ", " + body
	return first, second, true
}

// splitAbilityWordPrefix separates an optional leading "<word> — " ability word
// from a trigger introduction. The returned prefix retains the separator so the
// caller can prepend it to the second constituent ability; introduction is the
// remaining trigger text beginning with its introduction word.
func splitAbilityWordPrefix(head string) (prefix, introduction string) {
	if i := strings.Index(head, abilityWordSeparator); i >= 0 {
		boundary := i + len(abilityWordSeparator)
		return head[:boundary], head[boundary:]
	}
	return "", head
}
