package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// entersAndSecondTriggerMarker joins the enters-the-battlefield condition of a
// triggered ability to a second trigger condition that shares the same effect,
// as in "When this creature enters and at the beginning of your upkeep,
// surveil 1" (Unshakable Tail) and "When this artifact enters and when you
// sacrifice it, ..." (Carrot Cake).
const entersAndSecondTriggerMarker = " enters and "
const entersBattlefieldAndSecondTriggerMarker = " enters the battlefield and "

// secondTriggerLeadIns are the recognized lead-ins of the second trigger
// condition in an enters-and conjunction. Restricting the split to these
// lead-ins keeps the rewrite from firing on event subjects ("... and another
// creature you control dies") whose "and" is not a trigger-condition join.
var secondTriggerLeadIns = []string{"at the beginning of ", "when ", "whenever "}

// expandEntersAndSecondTrigger rewrites a triggered ability whose condition is
// the conjunction "<subject> enters and <second condition>, <effect>" into two
// independent triggered abilities sharing the effect text: "<subject> enters,
// <effect>" and "<second condition>, <effect>". An ability that lists multiple
// trigger events triggers separately on each event (CR 603.1), so emitting one
// ability per condition lets the standard trigger pipeline lower each
// constituent independently. The rewrite is parser-owned because it is a
// wording substitution; downstream stages see ordinary triggered abilities.
func expandEntersAndSecondTrigger(source string, cardNames ...string) string {
	cardName := ""
	if len(cardNames) > 0 {
		cardName = cardNames[0]
	}
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		expanded, ok := splitEntersAndSecondTriggerLine(line, cardName)
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

// splitEntersAndSecondTriggerLine splits a "(When|Whenever) <subject> enters
// and <second condition>, <effect>" line into an enters-the-battlefield trigger
// and a second trigger that share the post-condition effect text. It reports ok
// only when the line is a "When"/"Whenever" trigger (after an optional
// "<word> — " ability word) whose enters condition has a comma-free subject,
// whose second condition begins with a recognized trigger lead-in and is
// comma-free, and whose effect body follows the single delimiting comma.
func splitEntersAndSecondTriggerLine(line, cardName string) (lines []string, ok bool) {
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
	rest := strings.TrimPrefix(introduction, introWord)
	subject, tail, found := strings.Cut(rest, entersAndSecondTriggerMarker)
	entersText := " enters"
	if !found {
		subject, tail, found = strings.Cut(rest, entersBattlefieldAndSecondTriggerMarker)
		entersText = " enters the battlefield"
	}
	if !found {
		return nil, false
	}
	if subject == "" ||
		(strings.Contains(subject, ",") && !strings.EqualFold(subject, strings.TrimSpace(cardName))) {
		return nil, false
	}
	secondCondition, body, hasBody := strings.Cut(tail, ", ")
	if !hasBody || secondCondition == "" || body == "" {
		return nil, false
	}
	if !hasSecondTriggerLeadIn(secondCondition) {
		return nil, false
	}
	lines = append(lines, prefix+introWord+subject+entersText+", "+body)
	lines = append(lines, prefix+capitalizeFirstRune(secondCondition)+", "+body)
	return lines, true
}

// hasSecondTriggerLeadIn reports whether the second condition begins with one of
// the recognized trigger lead-ins, ignoring leading-letter case.
func hasSecondTriggerLeadIn(condition string) bool {
	lower := strings.ToLower(condition)
	for _, lead := range secondTriggerLeadIns {
		if strings.HasPrefix(lower, lead) {
			return true
		}
	}
	return false
}

// capitalizeFirstRune returns the string with its first rune upper-cased so a
// mid-sentence trigger condition reads as a standalone triggered ability.
func capitalizeFirstRune(s string) string {
	if s == "" {
		return s
	}
	first, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(first)) + s[size:]
}
