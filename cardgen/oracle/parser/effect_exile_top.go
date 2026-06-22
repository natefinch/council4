package parser

import "strings"

// parseExileTopOfLibrary recognizes the closed "Exile the top [N] card(s) of
// <player>'s library." effect and marks it as a top-of-library card source so
// lowering emits the exile-top-of-library primitive instead of a permanent
// target exile. It normalizes the implicit singular count to one and, for the
// controller-actor "exile the top card of each player's/opponent's library."
// shapes, records the library-owner player scope in the effect context.
func parseExileTopOfLibrary(effect *EffectSyntax) {
	amount, ownerContext, ok := exileTopOfLibraryAmount(effect)
	if !ok {
		return
	}
	effect.CardSource = EffectCardSourceTopOfPlayerLibrary
	effect.Context = ownerContext
	effect.Amount = EffectAmountSyntax{Span: effect.Amount.Span, Value: amount, Known: true}
}

// exileTopCandidate is one recognized "Exile the top card of <possessive>
// library." spelling: the reconstructed subject prefix, the library possessive,
// and the player scope whose libraries are exiled.
type exileTopCandidate struct {
	subject      string
	possessive   string
	ownerContext EffectContextKind
}

// exileTopOfLibraryAmount reports the normalized card count and resolved
// library-owner player scope for an exact "Exile the top [N] card(s) of
// <player>'s library." clause and whether the clause matches that closed shape.
// It accepts the controller, each-player, each-opponent, each-other-player, and
// single targeted-player subjects, including the controller-actor "exile the
// top card of each player's library." spelling, the player scopes the
// executable backend lowers.
func exileTopOfLibraryAmount(effect *EffectSyntax) (int, EffectContextKind, bool) {
	if effect.Kind != EffectExile ||
		effect.Negated ||
		effect.Optional ||
		effect.Additional ||
		len(effect.Targets) > 1 {
		return 0, "", false
	}
	amount := 1
	noun := "card"
	if effect.Amount.Known {
		if effect.Amount.Value < 1 {
			return 0, "", false
		}
		amount = effect.Amount.Value
		if amount > 1 {
			word, ok := cardinalNumberWord(amount)
			if !ok {
				return 0, "", false
			}
			noun = word + " cards"
		}
	}
	candidates, ok := exileTopOfLibraryCandidates(effect)
	if !ok {
		return 0, "", false
	}
	clause := exactEffectClauseText(effect)
	for _, candidate := range candidates {
		want := candidate.subject + "the top " + noun + " of " + candidate.possessive + " library."
		if strings.EqualFold(clause, want) {
			return amount, candidate.ownerContext, true
		}
	}
	return 0, "", false
}

// exileTopOfLibraryCandidates returns the recognized "Exile the top card of
// <possessive> library." spellings for the effect's actor context and whether
// that context is one the exile-top recognizer supports. The controller actor
// accepts both its own library ("your") and the each-player/each-opponent/
// each-other-player library possessives, mapping each to the player scope whose
// libraries are exiled.
func exileTopOfLibraryCandidates(effect *EffectSyntax) ([]exileTopCandidate, bool) {
	switch effect.Context {
	case EffectContextController:
		return []exileTopCandidate{
			{"Exile ", "your", EffectContextController},
			{"Exile ", "each player's", EffectContextEachPlayer},
			{"Exile ", "each opponent's", EffectContextEachOpponent},
			{"Exile ", "each other player's", EffectContextEachOtherPlayer},
		}, true
	case EffectContextEachPlayer:
		return []exileTopCandidate{{"Each player exiles ", "their", EffectContextEachPlayer}}, true
	case EffectContextEachOpponent:
		return []exileTopCandidate{{"Each opponent exiles ", "their", EffectContextEachOpponent}}, true
	case EffectContextEachOtherPlayer:
		return []exileTopCandidate{{"Each other player exiles ", "their", EffectContextEachOtherPlayer}}, true
	case EffectContextTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			return []exileTopCandidate{{
				titleFirstEffectText(effect.Targets[0].Text) + " exiles ", "their", EffectContextTarget,
			}}, true
		}
	default:
	}
	return nil, false
}

// exactExileTopOfLibrarySyntax reports whether an exile effect was recognized as
// the closed top-of-library card-source form, which lowers to the exile-top
// primitive.
func exactExileTopOfLibrarySyntax(effect *EffectSyntax) bool {
	return effect.CardSource == EffectCardSourceTopOfPlayerLibrary
}
