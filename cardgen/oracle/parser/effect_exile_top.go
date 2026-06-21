package parser

import "strings"

// parseExileTopOfLibrary recognizes the closed "Exile the top [N] card(s) of
// <player>'s library." effect and marks it as a top-of-library card source so
// lowering emits the exile-top-of-library primitive instead of a permanent
// target exile. It normalizes the implicit singular count to one.
func parseExileTopOfLibrary(effect *EffectSyntax) {
	amount, ok := exileTopOfLibraryAmount(effect)
	if !ok {
		return
	}
	effect.CardSource = EffectCardSourceTopOfPlayerLibrary
	effect.Amount = EffectAmountSyntax{Span: effect.Amount.Span, Value: amount, Known: true}
}

// exileTopOfLibraryAmount reports the normalized card count for an exact "Exile
// the top [N] card(s) of <player>'s library." clause and whether the clause
// matches that closed shape. It accepts the controller, each-player,
// each-opponent, each-other-player, and single targeted-player subjects, the
// player scopes the executable backend lowers.
func exileTopOfLibraryAmount(effect *EffectSyntax) (int, bool) {
	if effect.Kind != EffectExile ||
		effect.Negated ||
		effect.Optional ||
		effect.Additional ||
		len(effect.Targets) > 1 {
		return 0, false
	}
	amount := 1
	noun := "card"
	if effect.Amount.Known {
		if effect.Amount.Value < 1 {
			return 0, false
		}
		amount = effect.Amount.Value
		if amount > 1 {
			word, ok := cardinalNumberWord(amount)
			if !ok {
				return 0, false
			}
			noun = word + " cards"
		}
	}
	subject, possessive, ok := exileTopOfLibrarySubject(effect)
	if !ok {
		return 0, false
	}
	want := subject + "the top " + noun + " of " + possessive + " library."
	if strings.EqualFold(exactEffectClauseText(effect), want) {
		return amount, true
	}
	return 0, false
}

// exileTopOfLibrarySubject returns the reconstructed subject prefix (ending in a
// space) and the library possessive ("your" or "their") for the effect's player
// scope, and whether that scope is one the exile-top recognizer supports.
func exileTopOfLibrarySubject(effect *EffectSyntax) (subject, possessive string, ok bool) {
	switch effect.Context {
	case EffectContextController:
		return "Exile ", "your", true
	case EffectContextEachPlayer:
		return "Each player exiles ", "their", true
	case EffectContextEachOpponent:
		return "Each opponent exiles ", "their", true
	case EffectContextEachOtherPlayer:
		return "Each other player exiles ", "their", true
	case EffectContextTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			return titleFirstEffectText(effect.Targets[0].Text) + " exiles ", "their", true
		}
	default:
	}
	return "", "", false
}

// exactExileTopOfLibrarySyntax reports whether an exile effect was recognized as
// the closed top-of-library card-source form, which lowers to the exile-top
// primitive.
func exactExileTopOfLibrarySyntax(effect *EffectSyntax) bool {
	return effect.CardSource == EffectCardSourceTopOfPlayerLibrary
}
