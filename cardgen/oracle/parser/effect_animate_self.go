package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// AnimateSelfSyntax holds the structured fields of an EffectAnimateSelf effect
// (Faerie Conclave, the Keyrune mana rocks, Mutavault). Power and Toughness are
// the literal base power/toughness the source gains; Colors are the colors the
// source's color set becomes; AddArtifact records the stated "artifact" card
// type added alongside the creature card type; Subtypes are the creature
// subtype(s) added; EveryCreatureType records the "with all creature types"
// rider; and Keywords are the keyword(s) granted. The change lasts until end of
// turn while the source keeps its existing land or artifact types.
type AnimateSelfSyntax struct {
	Power             int
	Toughness         int
	Colors            []Color
	AddArtifact       bool
	Subtypes          []types.Sub
	EveryCreatureType bool
	Keywords          []KeywordKind
}

// parseAnimateSelfEffect recognizes the one-shot continuous self-animation "This
// <land|artifact|creature|permanent> becomes a N/N [<color>...] [artifact]
// [<subtype>...] creature [with <keyword>...|all creature types] until end of
// turn." (Faerie Conclave, Mishra's Factory, Stuffed Bear, the Keyrune and
// Monument mana rocks, Mutavault; CR 613). The trailing "until end of turn"
// duration is required, which is what distinguishes this temporary animation
// from a permanent type change. The base power/toughness is a literal N/N; the
// colors set the source's color set; an "artifact" card-type word adds the
// artifact type; the named subtypes (or the "all creature types" rider) are
// added creature types; and the "with" clause grants supported keyword(s). Any
// richer shape — an X/X or "with base power and toughness N/N" amount, a target
// or named selector ("becomes a ... named ..."), a quoted granted ability, a
// permanent duration, a non-{creature,artifact} card type, or an unsupported
// keyword — fails closed so those cards stay unsupported.
func parseAnimateSelfEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	inner, endOfTurn := becomeCopyTrimUntilEndOfTurn(body[:len(body)-1])
	if !endOfTurn {
		return nil, false
	}
	if len(inner) < 5 || !equalWord(inner[0], "this") || !animateSelfSubjectNoun(inner[1]) ||
		!equalWord(inner[2], "becomes") {
		return nil, false
	}
	cursor := 3
	if equalWord(inner[cursor], "a") || equalWord(inner[cursor], "an") {
		cursor++
	}
	pt, ok := parsePowerToughness(inner, cursor)
	if !ok {
		return nil, false
	}
	cursor = pt.Next

	colors, cursor := parseAnimateSelfColorRun(inner, cursor)

	characteristics, cursor, ok := parseAnimateSelfCharacteristicRun(inner, cursor, atoms)
	if !ok || !characteristics.HasCreature {
		return nil, false
	}

	keywords, everyCreatureType, ok := parseAnimateSelfRiders(inner, cursor)
	if !ok {
		return nil, false
	}
	if everyCreatureType && len(characteristics.Subtypes) != 0 {
		return nil, false
	}

	effect := EffectSyntax{
		Kind:       EffectAnimateSelf,
		Context:    EffectContextController,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), body...),
		Duration:   EffectDurationUntilEndOfTurn,
		AnimateSelf: &AnimateSelfSyntax{
			Power:             pt.Power,
			Toughness:         pt.Toughness,
			Colors:            colors,
			AddArtifact:       characteristics.AddArtifact,
			Subtypes:          characteristics.Subtypes,
			EveryCreatureType: everyCreatureType,
			Keywords:          keywords,
		},
	}
	return []EffectSyntax{effect}, true
}

// animateSelfSubjectNoun reports whether the token names a permanent type the
// self-animation subject may be ("This land/artifact/creature/permanent
// becomes ...").
func animateSelfSubjectNoun(token shared.Token) bool {
	return equalWord(token, "land") || equalWord(token, "artifact") ||
		equalWord(token, "creature") || equalWord(token, "permanent") ||
		equalWord(token, "token")
}

// parseAnimateSelfColorRun consumes the run of color words ("blue", "black and
// red") that follows the base power/toughness, returning the colors and the
// cursor past them. A connecting "and" is consumed only when another color word
// follows it, so the subtype run that follows is left untouched.
func parseAnimateSelfColorRun(tokens []shared.Token, cursor int) ([]Color, int) {
	colors := make([]Color, 0)
	for cursor < len(tokens) {
		parsedColor, ok := recognizeColorWord(tokens[cursor].Text)
		if !ok {
			break
		}
		colors = append(colors, parsedColor)
		cursor++
		if cursor+1 < len(tokens) && equalWord(tokens[cursor], "and") {
			if _, ok := recognizeColorWord(tokens[cursor+1].Text); ok {
				cursor++
			}
		}
	}
	return colors, cursor
}

// animateSelfCharacteristics holds the card types and subtypes the animation
// adds between the colors and the closing "with"/duration boundary.
type animateSelfCharacteristics struct {
	AddArtifact bool
	HasCreature bool
	Subtypes    []types.Sub
}

// parseAnimateSelfCharacteristicRun consumes the creature subtype(s) and card
// type(s) between the colors and the closing "with"/duration boundary. It
// reports the added "artifact" card type, whether the required "creature" card
// type is present, the added subtypes, and the cursor past the run. Any card
// type other than creature or artifact fails closed.
func parseAnimateSelfCharacteristicRun(
	tokens []shared.Token,
	cursor int,
	atoms Atoms,
) (animateSelfCharacteristics, int, bool) {
	characteristics := animateSelfCharacteristics{Subtypes: make([]types.Sub, 0)}
	for cursor < len(tokens) {
		if subtype, width, found := staticSubtypeAt(tokens, cursor, len(tokens), atoms); found {
			characteristics.Subtypes = append(characteristics.Subtypes, subtype)
			cursor += width
			continue
		}
		cardType, found := atoms.CardTypeAt(tokens[cursor].Span)
		if !found {
			break
		}
		switch cardType {
		case CardTypeCreature:
			characteristics.HasCreature = true
		case CardTypeArtifact:
			characteristics.AddArtifact = true
		default:
			return animateSelfCharacteristics{}, cursor, false
		}
		cursor++
	}
	return characteristics, cursor, true
}

// parseAnimateSelfRiders consumes the optional "with <keyword>...|all creature
// types" clause that closes the animation, returning the granted keywords and
// whether the every-creature-type rider is present. An empty cursor (no "with"
// clause) is valid. A quoted granted ability or any unrecognized rider word
// fails closed.
func parseAnimateSelfRiders(tokens []shared.Token, cursor int) (keywords []KeywordKind, everyCreatureType bool, ok bool) {
	if cursor == len(tokens) {
		return nil, false, true
	}
	if !equalWord(tokens[cursor], "with") {
		return nil, false, false
	}
	cursor++
	keywords = make([]KeywordKind, 0)
	for cursor < len(tokens) {
		if tokens[cursor].Kind == shared.Comma || equalWord(tokens[cursor], "and") {
			cursor++
			continue
		}
		if staticWordsAt(tokens, cursor, "all", "creature", "types") {
			everyCreatureType = true
			cursor += 3
			continue
		}
		kind, width, found := recognizeKeywordNameAt(tokens, cursor)
		if !found {
			return nil, false, false
		}
		keywords = append(keywords, kind)
		cursor += width
	}
	if len(keywords) == 0 && !everyCreatureType {
		return nil, false, false
	}
	return keywords, everyCreatureType, true
}

// abilityHasAnimateSelf reports whether the ability carries a recognized
// EffectAnimateSelf effect.
func abilityHasAnimateSelf(ability *Ability) bool {
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectAnimateSelf {
				return true
			}
		}
	}
	return false
}

// stripAnimateSelfSemantics clears the residual reference, keyword, and
// condition semantics the general scans re-derive for an ability whose resolving
// content is a single EffectAnimateSelf. The animation clause mentions a keyword
// ("with flying") and a self subject that those scans would otherwise surface as
// ability-level keywords or references, over-counting the ability and failing
// the lowering coverage gate. It mirrors stripPayRepeatedlyAnimateSemantics and
// runs after emitSemanticAccessors re-derives those fields.
func stripAnimateSelfSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasAnimateSelf(ability) {
			continue
		}
		ability.SemanticReferences = nil
		ability.SemanticKeywords = nil
		ability.ConditionBoundaries = nil
		ability.EventHistoryConditions = nil
		ability.ConditionClauses = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
		ability.StaticDeclarations = nil
	}
}

// foldAnimateSelfStillSentence extends an EffectAnimateSelf effect's span to
// cover the immediately following reminder-equivalent sentence "It's still a
// land." / "It's still an artifact." that confirms the animated permanent keeps
// its original type (Faerie Conclave, Mishra's Factory, the manland cycle). That
// confirmation carries no new semantics — the type layer adds rather than sets,
// so the source already keeps its land or artifact type — but its tokens would
// otherwise be left uncovered and fail the lowering coverage gate. Folding the
// span onto the recognized effect accounts for those tokens without adding any
// resolving behavior. Abilities without the trailing sentence are unaffected.
func foldAnimateSelfStillSentence(ability *Ability) {
	for i := range ability.Sentences {
		if !sentenceHasAnimateSelf(&ability.Sentences[i]) || i+1 >= len(ability.Sentences) {
			continue
		}
		next := &ability.Sentences[i+1]
		if len(next.Effects) != 0 || !isStillSourceTypeSentence(next.Tokens) {
			continue
		}
		sentence := &ability.Sentences[i]
		for e := range sentence.Effects {
			if sentence.Effects[e].Kind != EffectAnimateSelf {
				continue
			}
			sentence.Effects[e].Span.End = next.Span.End
			sentence.Effects[e].ClauseSpan.End = next.Span.End
		}
	}
}

// sentenceHasAnimateSelf reports whether the sentence carries an EffectAnimateSelf
// effect.
func sentenceHasAnimateSelf(sentence *Sentence) bool {
	for j := range sentence.Effects {
		if sentence.Effects[j].Kind == EffectAnimateSelf {
			return true
		}
	}
	return false
}

// isStillSourceTypeSentence reports whether the sentence is the fixed "It's
// still a land." / "It's still an artifact." confirmation that follows a
// self-animation.
func isStillSourceTypeSentence(tokens []shared.Token) bool {
	words := normalizedWords(semanticEffectTokens(tokens))
	if len(words) != 4 {
		return false
	}
	if words[0] != "it's" || words[1] != "still" || (words[2] != "a" && words[2] != "an") {
		return false
	}
	return words[3] == "land" || words[3] == "artifact"
}
