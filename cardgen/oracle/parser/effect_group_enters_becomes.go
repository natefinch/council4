package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// parseGroupEntersBecomesEffect recognizes the static group ETB characteristic
// replacement "As a <subject> you control enters, it becomes a [N/N] [<color>...]
// [<subtype>...] <card type>... in addition to its other types." (Displaced
// Dinosaurs: "As a historic permanent you control enters, it becomes a 7/7
// Dinosaur creature in addition to its other types."). The subject names whose
// entering permanents are affected (ControllerScope) and restricts them to
// historic permanents (the "historic" qualifier) or a bare "permanent"; the
// predicate sets the entrant's added card types, creature subtypes, colors, and
// an optional fixed base power/toughness. The trailing "in addition to its other
// types" rider is required so the entrant keeps its other types, which is what
// distinguishes this additive characteristic replacement from a type-setting
// animation. Any richer shape fails closed so unrelated wordings stay
// unsupported.
func parseGroupEntersBecomesEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 12 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := body[:len(body)-1]
	if !equalWord(words[0], "as") {
		return nil, false
	}
	cursor := 1
	if cursor < len(words) && (equalWord(words[cursor], "a") || equalWord(words[cursor], "an")) {
		cursor++
	}
	historic := false
	if cursor < len(words) && equalWord(words[cursor], "historic") {
		historic = true
		cursor++
	}
	var subjectTypes []types.Card
	if cursor >= len(words) {
		return nil, false
	}
	if equalWord(words[cursor], "permanent") {
		cursor++
	} else if cardType, ok := groupEntersTappedPermanentType(words[cursor].Text); ok {
		subjectTypes = append(subjectTypes, cardType)
		cursor++
	} else {
		return nil, false
	}
	scope := EntersTappedGroupControllerEach
	switch {
	case cursor+1 < len(words) && equalWord(words[cursor], "you") && equalWord(words[cursor+1], "control"):
		scope = EntersTappedGroupControllerYou
		cursor += 2
	case cursor+2 < len(words) && equalWord(words[cursor], "your") &&
		equalWord(words[cursor+1], "opponents") && equalWord(words[cursor+2], "control"):
		scope = EntersTappedGroupControllerOpponents
		cursor += 3
	case cursor+2 < len(words) && equalWord(words[cursor], "an") &&
		equalWord(words[cursor+1], "opponent") && equalWord(words[cursor+2], "controls"):
		scope = EntersTappedGroupControllerOpponents
		cursor += 3
	default:
	}
	if cursor >= len(words) || !equalWord(words[cursor], "enters") {
		return nil, false
	}
	cursor++
	if cursor < len(words) && words[cursor].Kind == shared.Comma {
		cursor++
	}
	if cursor+1 >= len(words) || !equalWord(words[cursor], "it") || !equalWord(words[cursor+1], "becomes") {
		return nil, false
	}
	cursor += 2
	if cursor < len(words) && (equalWord(words[cursor], "a") || equalWord(words[cursor], "an")) {
		cursor++
	}
	var basePower, baseToughness opt.V[int]
	if pt, ok := parsePowerToughness(words, cursor); ok {
		basePower = opt.Val(pt.Power)
		baseToughness = opt.Val(pt.Toughness)
		cursor = pt.Next
	}
	colors, cursor := parseAnimateSelfColorRun(words, cursor)
	if len(colors) == 0 {
		colors = nil
	}
	characteristics, cursor, ok := parseAnimateSelfCharacteristicRun(words, cursor, atoms)
	if !ok {
		return nil, false
	}
	subtypes := characteristics.Subtypes
	if len(subtypes) == 0 {
		subtypes = nil
	}
	var addTypes []types.Card
	if characteristics.HasCreature {
		addTypes = append(addTypes, types.Creature)
	}
	if characteristics.AddArtifact {
		addTypes = append(addTypes, types.Artifact)
	}
	if len(addTypes) == 0 && len(subtypes) == 0 {
		return nil, false
	}
	// The additive "in addition to its other types" rider is required: it keeps
	// the entrant's printed types, the semantics this replacement models.
	if !equalWordSequence(words, cursor, "in", "addition", "to", "its", "other", "types") {
		return nil, false
	}
	cursor += 6
	if cursor != len(words) {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:       EffectEnterTapped,
		Context:    EffectContextController,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		GroupEntryModification: GroupEntryModificationSyntax{
			Kind:            GroupEntryModificationBecomes,
			ControllerScope: scope,
			Types:           subjectTypes,
			Historic:        historic,
			AddTypes:        addTypes,
			AddSubtypes:     subtypes,
			Colors:          colors,
			BasePower:       basePower,
			BaseToughness:   baseToughness,
		},
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

// abilityHasGroupEntersBecomes reports whether the ability carries a recognized
// enters-becomes-group characteristic replacement effect.
func abilityHasGroupEntersBecomes(ability *Ability) bool {
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].EntersBecomesGroup() {
				return true
			}
		}
	}
	return false
}

// stripGroupEntersBecomesSemantics clears the residual reference, keyword, and
// condition semantics the general scans re-derive for an ability whose resolving
// content is a single enters-becomes-group replacement. The "As <subject>
// enters, it becomes ..." clause mentions the "it"/"its" pronouns and type words
// those scans would otherwise surface as ability-level references, over-counting
// the ability and failing the lowering coverage gate. It mirrors
// stripAnimateSelfSemantics and runs after emitSemanticAccessors re-derives those
// fields.
func stripGroupEntersBecomesSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasGroupEntersBecomes(ability) {
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
