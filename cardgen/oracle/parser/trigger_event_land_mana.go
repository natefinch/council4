package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// landAbilityAddsManaSuffixWords is the fixed tail of the Caged Sun mana-doubler
// trigger, following the possessive subject: "<subject>'s ability causes you to
// add one or more mana of the chosen color". The produced color is the source
// permanent's entry-time chosen color, resolved at match time.
var landAbilityAddsManaSuffixWords = []string{
	"ability", "causes", "you", "to", "add",
	"one", "or", "more", "mana", "of", "the", "chosen", "color",
}

// parseLandAbilityAddsManaTriggerEventClause recognizes "Whenever a land's
// ability causes you to add one or more mana of the chosen color" (Caged Sun).
// The engine models mana production through the authoritative mana-produced
// event, so the trigger fires whenever a land's ability adds the source's
// entry-time chosen color — by tapping, sacrificing, or paying life (CR 106.1 /
// 605). The subject's controller is forced to "you" because the produced mana is
// added to your pool. Any deviation from the exact wording returns nil so the
// shared dispatcher records no match.
func parseLandAbilityAddsManaTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	prefix, ok := stripTokenSuffix(tokens, landAbilityAddsManaSuffixWords...)
	if !ok || len(prefix) == 0 {
		return nil
	}
	// The possessive marker "'s" is attached to the final subject token
	// ("land's"); split it off so the subject parses as a plain permanent noun.
	last := prefix[len(prefix)-1]
	noun, ok := strings.CutSuffix(strings.ToLower(last.Text), "'s")
	if !ok || noun == "" {
		return nil
	}
	subjectTokens := append([]shared.Token(nil), prefix...)
	subjectTokens[len(subjectTokens)-1] = shared.Token{
		Kind: last.Kind,
		Text: noun,
		Span: last.Span,
	}
	subject := parsePermanentEventSubject(subjectTokens, false, atoms)
	if !subject.ok || subject.oneOrMore || subject.subject.Kind == TriggerEventSubjectSelf {
		return nil
	}
	controller := subject.controller
	if !mergeTriggerController(&controller, ControllerYou) {
		return nil
	}
	// The land subject becomes a ManaProducedByLand flag rather than a live
	// subject selection so the trigger still fires when the land has left the
	// battlefield (sacrifice-for-mana): the emitted event snapshots that its
	// source was a land.
	return &TriggerEventClause{
		Kind:                     TriggerEventKindManaProduced,
		Controller:               controller,
		ManaProducedByLand:       true,
		TappedForManaChosenColor: true,
	}
}
