package rules

import (
	"iter"
	"regexp"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

// abilityParagraphKind classifies an Oracle paragraph by the kind of ability
// its wording introduces.
type abilityParagraphKind int

const (
	paragraphOther abilityParagraphKind = iota
	paragraphTriggered
	paragraphActivated
	paragraphMana
	paragraphLoyalty
	paragraphChapter
)

var (
	chapterParagraphPattern = regexp.MustCompile(`^[IVX]+(, [IVX]+)* — `)
	loyaltyParagraphPattern = regexp.MustCompile(`^[+−-]?(\d+|X): `)
	abilityWordPattern      = regexp.MustCompile(`^[^:.,(—"]+? — `)
)

// printedAbilityParagraphs returns the Oracle paragraph of each untexted
// triggered, activated, loyalty and chapter ability of face whose paragraph can
// be identified, keyed by canonical ability index. Each paragraph of rules text
// is a separate ability (CR 113.2c), so an untexted ability's own text is one
// paragraph, never the face's whole Oracle text.
//
// Paragraphs that are some ability's own recorded text are set aside. When a
// single paragraph remains, it contains every remaining untexted ability. Past
// that, untexted abilities of one kind are paired in order with the paragraphs
// whose wording introduces that kind (a trigger word, an activation cost, a
// loyalty cost or chapter numerals) only when both counts agree; a triggered
// pairing must also agree on whether the trigger is an "At" trigger. Keyword
// abilities, which are written as their keyword rather than as their
// trigger or cost, and any kind whose paragraphs cannot be paired are left out:
// such an ability has no identifiable printed text, which is preferable to
// describing it with unrelated abilities.
func printedAbilityParagraphs(face *game.CardFace, oracleText string) map[int]string {
	if face == nil {
		return nil
	}
	texted := make(map[string]bool)
	var untexted []int
	for index := range face.AbilityCount() {
		body := face.BodyAt(index)
		if text := oracleRulesText(game.BodyText(body)); text != "" {
			texted[text] = true
			continue
		}
		switch body.(type) {
		case *game.TriggeredAbility, *game.ActivatedAbility, *game.LoyaltyAbility, *game.ChapterAbility:
			untexted = append(untexted, index)
		}
	}
	if len(untexted) == 0 {
		return nil
	}
	var paragraphs []string
	for paragraph := range oracleParagraphs(oracleText) {
		if !texted[paragraph] {
			paragraphs = append(paragraphs, paragraph)
		}
	}
	if len(paragraphs) == 0 {
		return nil
	}
	result := make(map[int]string, len(untexted))
	if len(paragraphs) == 1 {
		for _, index := range untexted {
			result[index] = paragraphs[0]
		}
		return result
	}
	byKind := make(map[abilityParagraphKind][]string)
	for _, paragraph := range paragraphs {
		kind := classifyAbilityParagraph(paragraph)
		byKind[kind] = append(byKind[kind], paragraph)
	}
	abilities := make(map[abilityParagraphKind][]int)
	for _, index := range untexted {
		if kind, ok := untextedAbilityParagraphKind(face.BodyAt(index)); ok {
			abilities[kind] = append(abilities[kind], index)
		} else {
			abilities[paragraphOther] = append(abilities[paragraphOther], index)
		}
	}
	for kind, indexes := range abilities {
		if kind == paragraphOther || len(indexes) != len(byKind[kind]) {
			continue
		}
		if kind == paragraphTriggered && !triggerWordsAgree(face, indexes, byKind[kind]) {
			continue
		}
		for i, index := range indexes {
			result[index] = byKind[kind][i]
		}
	}
	return result
}

// oracleParagraphs yields the paragraphs of Oracle text, each with the "•"
// mode lines that belong to it (CR 700.2).
func oracleParagraphs(oracleText string) iter.Seq[string] {
	return func(yield func(string) bool) {
		var current []string
		for line := range strings.Lines(oracleText) {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "•") && len(current) > 0 {
				current = append(current, line)
				continue
			}
			if len(current) > 0 && !yield(strings.Join(current, "\n")) {
				return
			}
			current = []string{line}
		}
		if len(current) > 0 {
			yield(strings.Join(current, "\n"))
		}
	}
}

// printedAbilityParagraph returns the Oracle paragraph of the untexted ability
// at index on face, or "" when it cannot be identified.
func printedAbilityParagraph(face *game.CardFace, oracleText string, index int) string {
	return printedAbilityParagraphs(face, oracleText)[index]
}

// untextedAbilityParagraphKind returns the paragraph kind that introduces an
// untexted ability's own wording. Keyword abilities are written as their
// keyword, so they have none.
func untextedAbilityParagraphKind(body game.Ability) (abilityParagraphKind, bool) {
	switch ability := body.(type) {
	case *game.TriggeredAbility:
		return paragraphTriggered, len(ability.KeywordAbilities) == 0
	case *game.ActivatedAbility:
		return paragraphActivated, len(ability.KeywordAbilities) == 0
	case *game.LoyaltyAbility:
		return paragraphLoyalty, true
	case *game.ChapterAbility:
		return paragraphChapter, true
	}
	return paragraphOther, false
}

func triggerWordsAgree(face *game.CardFace, indexes []int, paragraphs []string) bool {
	for i, index := range indexes {
		ability, ok := face.BodyAt(index).(*game.TriggeredAbility)
		if !ok {
			return false
		}
		at := strings.HasPrefix(withoutAbilityWord(paragraphs[i]), "At ")
		if at != (ability.Trigger.Type == game.TriggerAt) {
			return false
		}
	}
	return true
}

func classifyAbilityParagraph(paragraph string) abilityParagraphKind {
	switch {
	case chapterParagraphPattern.MatchString(paragraph):
		return paragraphChapter
	case loyaltyParagraphPattern.MatchString(paragraph):
		return paragraphLoyalty
	}
	body := withoutAbilityWord(paragraph)
	if startsWithTriggerWord(body) {
		return paragraphTriggered
	}
	if colon := topLevelIndex(body, ':'); colon >= 0 {
		if strings.HasPrefix(strings.TrimSpace(body[colon+1:]), "Add ") {
			return paragraphMana
		}
		return paragraphActivated
	}
	return paragraphOther
}

func startsWithTriggerWord(text string) bool {
	return strings.HasPrefix(text, "When ") || strings.HasPrefix(text, "Whenever ") || strings.HasPrefix(text, "At ")
}

// withoutAbilityWord strips an italicized ability word ("Landfall — ") that
// has no rules meaning of its own (CR 207.2c).
func withoutAbilityWord(paragraph string) string {
	if startsWithTriggerWord(paragraph) {
		return paragraph
	}
	if prefix := abilityWordPattern.FindString(paragraph); prefix != "" {
		return paragraph[len(prefix):]
	}
	return paragraph
}

// topLevelIndex returns the byte index of the first r outside quotation marks
// and parenthesized reminder text, or -1.
func topLevelIndex(text string, r rune) int {
	quoted := false
	depth := 0
	for i, c := range text {
		switch {
		case c == '"' || c == '“' || c == '”':
			quoted = !quoted
		case c == '(':
			depth++
		case c == ')' && depth > 0:
			depth--
		case c == r && !quoted && depth == 0:
			return i
		}
	}
	return -1
}

// delayedTriggerOwnText returns the sentence of a creating ability's text that
// is an event-based delayed triggered ability's own wording, through the end of
// its paragraph ("... put a +1/+1 counter on target creature. Whenever that
// creature attacks the monarch this turn, it gains double strike and trample
// until end of turn."). A delayed triggered ability has only its own text (CR
// 603.7a), not the clause of the ability that created it. It reports false
// unless exactly one later sentence of a paragraph begins with a trigger word,
// so wording that cannot be singled out is kept whole.
func delayedTriggerOwnText(text string) (string, bool) {
	var found string
	count := 0
	for line := range strings.Lines(text) {
		line = strings.TrimSpace(line)
		for start := range laterSentenceStarts(line) {
			sentence := line[start:]
			if strings.HasPrefix(sentence, "When ") || strings.HasPrefix(sentence, "Whenever ") {
				found = sentence
				count++
			}
		}
	}
	return found, count == 1
}

// laterSentenceStarts yields the byte offsets of the sentences after the first
// in paragraph, ignoring sentence ends inside quotation marks and reminder text.
func laterSentenceStarts(paragraph string) iter.Seq[int] {
	return func(yield func(int) bool) {
		quoted := false
		depth := 0
		for i, c := range paragraph {
			switch {
			case c == '"' || c == '“' || c == '”':
				quoted = !quoted
			case c == '(':
				depth++
			case c == ')' && depth > 0:
				depth--
			case c == '.' && !quoted && depth == 0 && strings.HasPrefix(paragraph[i:], ". ") && i+2 < len(paragraph):
				if !yield(i + 2) {
					return
				}
			}
		}
	}
}
