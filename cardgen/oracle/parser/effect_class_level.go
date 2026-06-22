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
