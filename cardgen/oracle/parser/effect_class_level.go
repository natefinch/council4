package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// classLevelReminder is the fixed reminder line every Class enchantment carries.
const classLevelReminder = "(Gain the next level as a sorcery to add its ability.)"

// recognizeClassLevelReminder reports whether the ability text is the canonical
// Class level-up reminder line.
func recognizeClassLevelReminder(text string) bool {
	return text == classLevelReminder
}

// recognizeClassLevelGain reads the target level from a Class enchantment's
// level-up activated-ability body ("Level N"), returning 0 when the body is not
// exactly the two-token "Level <number>" form.
func recognizeClassLevelGain(body []shared.Token) int {
	if len(body) != 2 {
		return 0
	}
	if body[0].Kind != shared.Word || body[0].Text != "Level" {
		return 0
	}
	if body[1].Kind != shared.Integer {
		return 0
	}
	level, err := strconv.Atoi(body[1].Text)
	if err != nil || level < 2 {
		return 0
	}
	return level
}

// parseClassBecameLevelTriggerEventClause recognizes the self-source trigger
// "this Class becomes level N" on a Class enchantment (CR 716). The clause's
// subject is the ability's own source and ClassBecameLevel carries the level
// reached.
func parseClassBecameLevelTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if len(tokens) < 5 {
		return nil
	}
	last := tokens[len(tokens)-1]
	if last.Kind != shared.Integer {
		return nil
	}
	if !equalWord(tokens[len(tokens)-3], "becomes") || !equalWord(tokens[len(tokens)-2], "level") {
		return nil
	}
	subjectTokens := tokens[:len(tokens)-3]
	if !classBecameLevelSubjectIsSelf(subjectTokens, atoms) {
		return nil
	}
	level, err := strconv.Atoi(last.Text)
	if err != nil || level < 2 {
		return nil
	}
	return &TriggerEventClause{
		Span:             shared.SpanOf(tokens),
		Kind:             TriggerEventKindClassBecameLevel,
		Subject:          TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: shared.SpanOf(subjectTokens)},
		ClassBecameLevel: level,
	}
}

// classBecameLevelSubjectIsSelf reports whether the subject tokens name the
// ability's own source, either the literal "this Class" wording or a recognized
// self-name/source marker span.
func classBecameLevelSubjectIsSelf(tokens []shared.Token, atoms Atoms) bool {
	if syntaxWordsEqual(tokens, "this", "Class") {
		return true
	}
	if _, count, ok := parseSelfSubject(tokens, atoms); ok && count == len(tokens) {
		return true
	}
	return false
}
