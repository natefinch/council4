package parser

import (
	"fmt"
	"strings"
)

// createTokenLeadingPlayerSetForEachUnmodeled reports whether a create-token
// effect opens with a "For each <player-set>[ who ...]," distributive quantifier
// (opponent or player) that the token-count model does not represent. The only
// modeled leading player-set distributive is the unconditional per-opponent
// count "For each opponent, [you ]create <one token>.", which
// parseCreateForEachAmount types as an EffectDynamicAmountFormForEach /
// EffectDynamicAmountOpponentCount amount so the runtime creates one token per
// opponent (Endless Ranks of HYDRA, Stampede Surfer). Every other leading
// player-set distributive — a conditional "For each opponent who <condition>,"
// (Faerie Slumber Party), a per-opponent count greater than one, or a per-player
// "For each player," — leaves the amount a fixed count that silently ignores the
// distributive, so this reports it unmodeled and the create dispatch fails closed
// rather than emitting a wrong fixed token count. It reads only the effect's own
// clause tokens and typed amount, so no downstream stage compares Oracle text.
func createTokenLeadingPlayerSetForEachUnmodeled(effect *EffectSyntax) bool {
	tokens := effect.Tokens
	if len(tokens) < 3 || !equalWord(tokens[0], "for") || !equalWord(tokens[1], "each") {
		return false
	}
	switch {
	case equalWord(tokens[2], "opponent"), equalWord(tokens[2], "opponents"),
		equalWord(tokens[2], "player"), equalWord(tokens[2], "players"):
	default:
		return false
	}
	return effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.DynamicKind != EffectDynamicAmountOpponentCount
}

// createTokenControllerClauseMatches reports whether a controller-form
// create-token clause matches its canonical reconstruction, accepting both the
// bare imperative "Create <body>" wording and the "You create <body>" subject
// wording that embedded trigger and ability bodies use ("..., you create a
// Treasure token.", "Whenever ..., you create a 1/1 ... token."). The two
// wordings describe the identical controller effect; the subject "You" is a
// surface variant the byte-exact reconstruction would otherwise reject. A
// trailing-"instead" escalation clause is normalized by exactEffectClauseText
// (the " instead" suffix is stripped for the plain EffectReplacementInstead
// form), so it reaches here as the bare body.
func createTokenControllerClauseMatches(clause, body string) bool {
	return strings.EqualFold(clause, "Create "+body) ||
		strings.EqualFold(clause, "You create "+body)
}

// createTokenSingleRecipientClauseMatches reconstructs the subject of a
// single-recipient token creation from its typed effect context. Besides the
// controller form, it admits the event-player wording used by attack triggers
// ("that attacking player creates ..."). The parser owns these surface variants;
// compiler and card generation consume only the typed context.
func createTokenSingleRecipientClauseMatches(effect *EffectSyntax, body string) bool {
	switch effect.Context {
	case EffectContextController:
		return createTokenControllerClauseMatches(exactEffectClauseText(effect), body)
	case EffectContextEventPlayer:
		clause := exactEffectClauseText(effect)
		return strings.EqualFold(clause, "That attacking player creates "+body) ||
			strings.EqualFold(clause, "The attacking player creates "+body)
	default:
		return false
	}
}

// createTokenControllerForEachClauseMatches reports whether a "for each"
// controller-form create-token clause matches either word order, in both the
// bare imperative and "you create" subject wordings: the leading-iterator form
// ("<iter>, create <spec>." / "<iter>, you create <spec>.") and the trailing
// form ("Create <spec> <iter>." / "You create <spec> <iter>.").
func createTokenControllerForEachClauseMatches(full, spec, iter string) bool {
	return strings.EqualFold(full, iter+", create "+spec+".") ||
		strings.EqualFold(full, iter+", you create "+spec+".") ||
		strings.EqualFold(full, "Create "+spec+" "+iter+".") ||
		strings.EqualFold(full, "You create "+spec+" "+iter+".")
}

// eachPlayerRecipientSubject returns the canonical clause subject for a
// token-creation effect whose recipient is a player group ("Each player",
// "Each opponent"), or ok=false for any other recipient context. The subject is
// reconstructed from the classified context rather than the raw text, so any
// qualified group ("Each player who controls the fewest creatures", "Each player
// other than target player") fails the byte-exact comparison and stays
// unsupported. The one supported group qualifier is the "who controls
// <selection>" per-member conditional ("Each player who controls an artifact or
// enchantment"), whose reconstructed clause recipientControlsQualifierText
// appends; every other qualifier leaves recipientControlsQualifierText ok=false
// and fails closed here.
func eachPlayerRecipientSubject(effect *EffectSyntax) (string, bool) {
	var subject string
	switch effect.Context {
	case EffectContextEachPlayer:
		subject = "Each player"
	case EffectContextEachOpponent:
		subject = "Each opponent"
	default:
		return "", false
	}
	qualifier, ok := recipientControlsQualifierText(effect.RecipientControlsSelection)
	if !ok {
		return "", false
	}
	if qualifier != "" {
		subject += " " + qualifier
	}
	return subject, true
}

// recipientControlsQualifierText reconstructs the canonical "who controls
// <selection>" clause of a group recipient's per-member conditional ("Each
// player who controls an artifact or enchantment creates ..."). It returns an
// empty string with ok=true when the recipient carries no such qualifier, the
// reconstructed clause with ok=true when the selection is a bare union of
// permanent card types controlled by anyone, and ok=false (fail closed) for any
// selection carrying a further qualifier. The union renders with the singular
// "or" conjunction and an indefinite article ("an artifact or enchantment"); the
// canonical noun phrase from selectionPhrase must equal the plain "and"-joined
// type union, which rejects any hidden color, combat, controller, or numeric
// qualifier the "or"-swap would otherwise silently drop.
func recipientControlsQualifierText(selection *SelectionSyntax) (string, bool) {
	if selection == nil {
		return "", true
	}
	if selection.Controller != SelectionControllerAny {
		return "", false
	}
	nouns, ok := permanentTypeUnionNouns(selection)
	if !ok {
		return "", false
	}
	canonical, ok := selectionPhrase(*selection, selectionPhraseOptions{Number: numberSingular})
	if !ok || canonical != joinUnionNounsSep(nouns, "and") {
		return "", false
	}
	return "who controls " + indefiniteArticle(nouns[0]) + " " + joinUnionNounsSep(nouns, "or"), true
}

// permanentTypeUnionNouns returns the lowercase Oracle nouns for a selection that
// names one or more permanent card types, mirroring selectionPhraseNoun: a
// RequiredTypesAny union yields one noun per distinct type, and a single-type
// selection falls back to its Kind's permanent noun. It fails closed for a
// duplicate type or any non-permanent type.
func permanentTypeUnionNouns(selection *SelectionSyntax) ([]string, bool) {
	if len(selection.RequiredTypesAny) >= 1 {
		nouns := make([]string, 0, len(selection.RequiredTypesAny))
		seen := make(map[CardType]bool, len(selection.RequiredTypesAny))
		for _, cardType := range selection.RequiredTypesAny {
			if seen[cardType] {
				return nil, false
			}
			seen[cardType] = true
			noun, ok := permanentCardTypeNoun(cardType)
			if !ok {
				return nil, false
			}
			nouns = append(nouns, noun)
		}
		return nouns, true
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return nil, false
	}
	return []string{noun}, true
}

// exactCreateTokenEachPlayerEffectSyntax recognizes the player-group recipient
// form of a fixed-count token creation, "Each player creates <spec>." and "Each
// opponent creates <spec>.", for both vanilla creature tokens and predefined
// artifact tokens (Treasure, Food, ...). The runtime resolves the group and
// creates the token for each member; only the simple fixed-count shape with no
// targets, granted ability, copy, choice, or dynamic count is accepted,
// mirroring the controller paths. Every richer or qualified group context fails
// closed.
func exactCreateTokenEachPlayerEffectSyntax(effect *EffectSyntax) bool {
	subject, ok := eachPlayerRecipientSubject(effect)
	if !ok {
		return false
	}
	if effect.Negated || effect.TokenCopyOfTarget || effect.TokenChoice ||
		effect.TokenGrantedAbility != nil ||
		effect.TokenPTVariableX ||
		len(effect.Targets) != 0 ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		effect.Amount.VariableX ||
		!effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	countWord, noun := createTokenArticle(effect), "token"
	if effect.Amount.Value != 1 {
		countWord, noun = effectAmountSourceText(effect), "tokens"
	}
	if effect.TokenPTKnown {
		specBody, ok := creatureTokenSpecBody(effect)
		if !ok {
			return false
		}
		return strings.EqualFold(exactEffectClauseText(effect),
			subject+" creates "+specBody(countWord, noun)+".")
	}
	sel := effect.Selection
	if sel.Kind != SelectionUnknown ||
		len(sel.SubtypesAny) != 1 ||
		!namedArtifactTokenSubtype(sel.SubtypesAny[0]) ||
		sel.Keyword != KeywordUnknown ||
		len(sel.ColorsAny) != 0 || len(sel.ExcludedColors) != 0 ||
		len(sel.RequiredTypesAny) != 0 || len(sel.ExcludedTypes) != 0 ||
		len(sel.SourceTypes) != 0 || len(sel.Supertypes) != 0 ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Untapped || sel.Attacking || sel.Blocking ||
		sel.All || sel.Another || sel.Other ||
		sel.Colorless || sel.Multicolored {
		return false
	}
	tappedPart := ""
	if sel.Tapped {
		tappedPart = "tapped "
	}
	specBody := fmt.Sprintf("%s %s%s %s", countWord, tappedPart, string(sel.SubtypesAny[0]), noun)
	return strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody+".")
}
