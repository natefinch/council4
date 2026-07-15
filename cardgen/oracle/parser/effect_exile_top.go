package parser

import "strings"

// parseExileTopOfLibrary recognizes the closed "Exile the top [N] card(s) of
// <player>'s library." effect and marks it as a top-of-library card source so
// lowering emits the exile-top-of-library primitive instead of a permanent
// target exile. It normalizes the implicit singular count to one and, for the
// controller-actor "exile the top card of each player's/opponent's library."
// shapes, records the library-owner player scope in the effect context.
func parseExileTopOfLibrary(effect *EffectSyntax) {
	if parseExileThatManyTopOfLibrary(effect) {
		return
	}
	match, ok := exileTopOfLibraryAmount(effect)
	if !ok {
		return
	}
	effect.CardSource = EffectCardSourceTopOfPlayerLibrary
	effect.Context = match.ownerContext
	effect.FaceDown = match.faceDown
	effect.Amount = EffectAmountSyntax{Span: effect.Amount.Span, Value: match.amount, Known: true}
}

// parseExileThatManyTopOfLibrary recognizes the dynamic controller-scoped "exile
// that many cards from the top of your library[ face down]." effect, marking it
// as a top-of-library card source (like the fixed-count form) while preserving
// the "that many" triggering-event amount that parseEffectAmount typed onto the
// effect. The optional trailing "face down" sets FaceDown so lowering exiles the
// cards face down. The amount is only meaningful inside a measuring trigger; the
// parser records the generic triggering-event kind and lowering fails it closed
// elsewhere. It reports whether it recognized the clause.
func parseExileThatManyTopOfLibrary(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile ||
		effect.Negated ||
		effect.Optional ||
		effect.Additional ||
		len(effect.Targets) > 0 ||
		effect.Context != EffectContextController ||
		effect.CounterKnown ||
		effect.Amount.DynamicKind != EffectDynamicAmountTriggeringCombatDamage {
		return false
	}
	clause := exactEffectClauseText(effect)
	const base = "Exile that many cards from the top of your library"
	faceDown := false
	switch {
	case strings.EqualFold(clause, base+"."):
	case strings.EqualFold(clause, base+" face down."):
		faceDown = true
	default:
		return false
	}
	effect.CardSource = EffectCardSourceTopOfPlayerLibrary
	effect.FaceDown = faceDown
	return true
}

// exileTopCandidate is one recognized "Exile the top card of <possessive>
// library." spelling: the reconstructed subject prefix, the library possessive,
// and the player scope whose libraries are exiled.
type exileTopCandidate struct {
	subject      string
	possessive   string
	ownerContext EffectContextKind
}

// exileTopMatch is the normalized result of matching an exact "Exile the top [N]
// card(s) of <player>'s library[ face down]." clause: the card count, the
// resolved library-owner player scope, and whether the cards are exiled face
// down.
type exileTopMatch struct {
	amount       int
	ownerContext EffectContextKind
	faceDown     bool
}

// exileTopOfLibraryAmount reports the normalized card count, resolved
// library-owner player scope, and whether the cards are exiled face down for an
// exact "Exile the top [N] card(s) of <player>'s library[ face down]." clause,
// and whether the clause matches that closed shape. It accepts the controller,
// each-player, each-opponent, each-other-player, and single targeted-player
// subjects, including the controller-actor "exile the top card of each player's
// library." spelling, the player scopes the executable backend lowers. The
// trailing " face down" rider is recognized only for the controller's own
// library ("Exile the top card of your library face down.", Necropotence), where
// the exiled card is hidden until a later effect returns it; every other subject
// stays face up so a face-down rider on it falls through and fails closed.
func exileTopOfLibraryAmount(effect *EffectSyntax) (exileTopMatch, bool) {
	if effect.Kind != EffectExile ||
		effect.Negated ||
		effect.Optional ||
		effect.Additional ||
		len(effect.Targets) > 1 {
		return exileTopMatch{}, false
	}
	amount := 1
	noun := "card"
	if effect.Amount.Known {
		if effect.Amount.Value < 1 {
			return exileTopMatch{}, false
		}
		amount = effect.Amount.Value
		if amount > 1 {
			word, ok := cardinalNumberWord(amount)
			if !ok {
				return exileTopMatch{}, false
			}
			noun = word + " cards"
		}
	}
	candidates, ok := exileTopOfLibraryCandidates(effect)
	if !ok {
		return exileTopMatch{}, false
	}
	clause := exactEffectClauseText(effect)
	for _, candidate := range candidates {
		base := candidate.subject + "the top " + noun + " of " + candidate.possessive + " library"
		if strings.EqualFold(clause, base+".") {
			return exileTopMatch{amount: amount, ownerContext: candidate.ownerContext}, true
		}
		// The trailing "face down" rider (Necropotence's "Exile the top card of
		// your library face down.") exiles the controller's own top card hidden
		// so a later effect can return it without revealing it. It is recognized
		// only for the controller's own library; any other library keeps the
		// public exile and a face-down rider on it fails closed.
		if candidate.ownerContext == EffectContextController &&
			strings.EqualFold(clause, base+" face down.") {
			return exileTopMatch{amount: amount, ownerContext: candidate.ownerContext, faceDown: true}, true
		}
		// A trailing "with a <kind> counter on it/them" rider places one named
		// marker counter on each exiled card ("exile the top card of each
		// player's library with a collection counter on it.", Evelyn, the
		// Covetous). The counter kind is captured text-blind in the effect's
		// CounterKind/CounterKnown fields; lowering reads it onto the
		// ExileTopOfLibrary primitive. Only the single-counter placement form is
		// recognized, so multi-counter riders fall through and fail closed.
		if effect.CounterKnown {
			for _, suffix := range exileTopCounterSuffixes(effect.CounterKind.String()) {
				if strings.EqualFold(clause, base+suffix) {
					return exileTopMatch{amount: amount, ownerContext: candidate.ownerContext}, true
				}
			}
		}
	}
	return exileTopMatch{}, false
}

// exileTopCounterSuffixes reconstructs the recognized "with a <kind> counter on
// it/them" placement riders for an exile-top clause, covering both indefinite
// articles and both card-count pronouns so the recognizer generalizes across
// counter kinds and singular/plural exile counts.
func exileTopCounterSuffixes(kind string) []string {
	suffixes := make([]string, 0, 4)
	for _, article := range []string{"a", "an"} {
		for _, pronoun := range []string{"it", "them"} {
			suffixes = append(suffixes, " with "+article+" "+kind+" counter on "+pronoun+".")
		}
	}
	return suffixes
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
