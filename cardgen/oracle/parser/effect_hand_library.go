package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func parseHandLibraryPut(effect *EffectSyntax) HandLibraryPutSyntax {
	if effect.Kind != EffectPut || effect.ToZone != zone.Library {
		return HandLibraryPutSyntax{}
	}
	refuse := func(parameter string, tokens []shared.Token) HandLibraryPutSyntax {
		span := shared.SpanOf(tokens)
		if len(tokens) == 0 {
			span = effect.ClauseSpan
		}
		return HandLibraryPutSyntax{UnsupportedParameter: parameter, ParameterSpan: span}
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return refuse("clause", nil)
	}
	tokens := effect.Tokens[verb+1:]
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	from := slices.IndexFunc(tokens, func(token shared.Token) bool { return equalWord(token, "from") })
	if from < 0 {
		return HandLibraryPutSyntax{}
	}
	hand := slices.IndexFunc(tokens[from:], func(token shared.Token) bool { return equalWord(token, "hand") })
	if hand < 0 {
		return HandLibraryPutSyntax{}
	}
	hand += from
	if from == 0 {
		return refuse("card selection", tokens[:hand+1])
	}
	amount := effect.Amount
	if amount.RangeKnown || amount.AnyNumber ||
		amount.DynamicKind != EffectDynamicAmountNone || amount.DynamicForm != EffectDynamicAmountFormNone ||
		amount.Multiplier != 0 || amount.Addend != 0 ||
		amount.Known == amount.VariableX || amount.Known && amount.Value <= 0 {
		return refuse("amount", tokens[:from])
	}
	countedText := "Put " + joinedEffectText(tokens[:from]) + "."
	if !exactCountedNounEffectText(countedText, "Put", "card", "cards", amount, effectAmountSourceText(effect), false) {
		noun := tokens[from-1]
		if !equalWord(noun, "card") && !equalWord(noun, "cards") {
			return refuse("card selection", tokens[:from])
		}
		if from > 2 && tokens[0].Span == amount.Span {
			return refuse("card selection", tokens[1:from-1])
		}
		return refuse("amount", tokens[:from])
	}
	if !strings.EqualFold(joinedEffectText(tokens[from:hand+1]), "from your hand") {
		return refuse("source owner", tokens[from:hand+1])
	}
	library := slices.IndexFunc(tokens[hand+1:], func(token shared.Token) bool { return equalWord(token, "library") })
	if library < 0 {
		return refuse("destination owner", tokens[hand+1:])
	}
	library += hand + 1
	position := tokens[hand+1 : library]
	top := strings.EqualFold(joinedEffectText(position), "on top of your")
	bottom := strings.EqualFold(joinedEffectText(position), "on the bottom of your")
	if !top && !bottom {
		if effectWordsAt(position, 0, "on", "top", "of") || effectWordsAt(position, 0, "on", "the", "bottom", "of") {
			return refuse("destination owner", position)
		}
		return refuse("library position", position)
	}
	order := tokens[library+1:]
	if len(order) != 0 && !strings.EqualFold(joinedEffectText(order), "in any order") {
		return refuse("library order", order)
	}
	if effect.Context != EffectContextController || len(effect.Targets) != 0 {
		return refuse("recipient", effect.Tokens[:verb])
	}
	return HandLibraryPutSyntax{Present: true, Bottom: bottom}
}

func exactHandLibraryPutEffectSyntax(effect *EffectSyntax) bool {
	if !effect.HandLibraryPut.Present {
		return false
	}
	noun := "cards"
	if effect.Amount.Known && effect.Amount.Value == 1 {
		noun = "card"
	}
	position := "on top of"
	if effect.HandLibraryPut.Bottom {
		position = "on the bottom of"
	}
	clause := "Put " + effectAmountSourceText(effect) + " " + noun + " from your hand " + position + " your library"
	text := exactEffectClauseText(effect)
	return strings.EqualFold(text, clause+".") || strings.EqualFold(text, clause+" in any order.")
}
