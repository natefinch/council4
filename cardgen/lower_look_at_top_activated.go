package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerActivatedBodyContent lowers an activated ability's resolving body. It
// first routes the parser-recognized conditional look-at-top battlefield exact
// sequence — "{cost}: Look at the top card of your library. If it's a <type>
// card, you may put it onto the battlefield[ tapped]. If you don't put the card
// onto the battlefield, you may put it on the bottom of your library." (Lantern
// of Revealing, Parcelbeast) — to the same fixed instruction template as the
// triggered form, then defers to the generic ability-content lowerer for every
// other body. It fails closed unless the compiler recorded at least one card
// type, so a partial recognition never lowers to a silently-wrong body.
func lowerActivatedBodyContent(
	cardName string,
	ability compiler.CompiledAbility,
	bodyContent compiler.AbilityContent,
	bodySyntax *parser.Ability,
	bodyText string,
) (game.AbilityContent, *shared.Diagnostic) {
	if ability.ExactSequence == compiler.ExactSequenceConditionalLookAtTopBattlefield &&
		len(ability.ExactSequenceLookAtTopTypes) > 0 {
		return game.Mode{
			Text:     bodyText,
			Sequence: conditionalLookAtTopBattlefieldSequence(ability),
		}.Ability(), nil
	}
	return lowerAbilityContent(cardName, ability.Kind, bodyContent, false, bodySyntax)
}
