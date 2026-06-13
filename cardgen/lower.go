package cardgen

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// loweredStaticAbility holds a typed StaticAbility with optional rendering
// metadata. VarName, when set, is a package-level variable reference like
// "game.FlyingStaticBody" that the Renderer emits instead of a struct literal.
type loweredStaticAbility struct {
	Body    game.StaticAbility
	VarName string
}

// loweredFaceAbilities holds the categorized typed game ability values
// produced by strict executable lowering for one card face, in Oracle order.
type loweredFaceAbilities struct {
	StaticAbilities      []loweredStaticAbility
	ActivatedAbilities   []game.ActivatedAbility
	ManaAbilities        []game.ManaAbility
	LoyaltyAbilities     []game.LoyaltyAbility
	TriggeredAbilities   []game.TriggeredAbility
	ChapterAbilities     []game.ChapterAbility
	ReplacementAbilities []game.ReplacementAbility
	SpellAbility         opt.V[game.AbilityContent]
	EntersPrepared       bool
}

// empty reports whether the face produced no abilities.
func (f loweredFaceAbilities) empty() bool {
	return len(f.StaticAbilities) == 0 &&
		len(f.ActivatedAbilities) == 0 &&
		len(f.ManaAbilities) == 0 &&
		len(f.LoyaltyAbilities) == 0 &&
		len(f.TriggeredAbilities) == 0 &&
		len(f.ChapterAbilities) == 0 &&
		len(f.ReplacementAbilities) == 0 &&
		!f.SpellAbility.Exists &&
		!f.EntersPrepared
}

// abilityLowering holds the typed result of lowering one CompiledAbility.
// Fields are set according to which ability kind was matched.
type abilityLowering struct {
	staticAbilities    []loweredStaticAbility
	activatedAbility   opt.V[game.ActivatedAbility]
	manaAbility        opt.V[game.ManaAbility]
	loyaltyAbility     opt.V[game.LoyaltyAbility]
	triggeredAbility   opt.V[game.TriggeredAbility]
	chapterAbility     opt.V[game.ChapterAbility]
	replacementAbility opt.V[game.ReplacementAbility]
	spellAbility       opt.V[game.AbilityContent]
	entersPrepared     bool
	consumed           semanticConsumption
	sourceSpans        []shared.Span
}

type semanticConsumption struct {
	cost         bool
	trigger      bool
	optional     bool
	modes        int
	targets      int
	conditions   int
	effects      int
	keywords     int
	references   int
	declarations int
}

// lowerExecutableFaces lowers every face of a card into typed ability values.
// It returns the face abilities in the same positional order as
// executableFaces and any diagnostics that prevented full lowering.
func lowerExecutableFaces(card *ScryfallCard) ([]loweredFaceAbilities, []shared.Diagnostic) {
	faces := executableFaces(card)
	lowered := make([]loweredFaceAbilities, len(faces))
	var diagnostics []shared.Diagnostic
	for i, face := range faces {
		faceAbilities, faceDiagnostics := lowerFaceAbilities(face)
		diagnostics = append(diagnostics, faceDiagnostics...)
		lowered[i] = faceAbilities
	}
	if card.Layout != "adventure" && hasAdventureCastPermission(lowered) {
		diagnostics = append(diagnostics, shared.Diagnostic{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported Adventure cast permission",
			Detail:   "an Adventure graveyard-cast permission requires an Adventure card layout",
		})
	}
	return lowered, diagnostics
}

func hasAdventureCastPermission(faces []loweredFaceAbilities) bool {
	for faceIndex := range faces {
		for abilityIndex := range faces[faceIndex].TriggeredAbilities {
			ability := &faces[faceIndex].TriggeredAbilities[abilityIndex]
			for modeIndex := range ability.Content.Modes {
				mode := &ability.Content.Modes[modeIndex]
				for instructionIndex := range mode.Sequence {
					instruction := &mode.Sequence[instructionIndex]
					if instruction.Primitive != nil &&
						instruction.Primitive.Kind() == game.PrimitiveGrantCastPermission {
						return true
					}
				}
			}
		}
	}
	return false
}

func lowerFaceAbilities(
	face scryfallFaceFields,
) (loweredFaceAbilities, []shared.Diagnostic) {
	parsedType := ParseTypeLine(face.TypeLine)
	if len(parsedType.Types) == 0 {
		return loweredFaceAbilities{}, []shared.Diagnostic{{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported type line",
			Detail:   fmt.Sprintf("type line %q has no supported card type", face.TypeLine),
		}}
	}
	if face.OracleText == "" {
		return loweredFaceAbilities{}, nil
	}
	document, diagnostics := parser.Parse(face.OracleText, parser.Context{
		InstantOrSorcery: slices.Contains(parsedType.Types, "Instant") || slices.Contains(parsedType.Types, "Sorcery"),
		Planeswalker:     slices.Contains(parsedType.Types, "Planeswalker"),
		Saga:             slices.Contains(parsedType.Subtypes, "Saga"),
		CardName:         face.Name,
	})
	compilation, compilerDiagnostics := compiler.Compile(document, compiler.Context{})
	diagnostics = append(diagnostics, compilerDiagnostics...)

	var result loweredFaceAbilities
	var unsupported []shared.Diagnostic
	for i, ability := range compilation.Abilities {
		syntax := compilation.Syntax.Abilities[i]
		lowered, diagnostic := lowerExecutableAbility(
			face.Name,
			slices.Contains(parsedType.Subtypes, "Saga"),
			ability,
			syntax,
		)
		if diagnostic != nil {
			unsupported = append(unsupported, *diagnostic)
			continue
		}
		if !lowered.complete(ability, syntax) {
			unsupported = append(unsupported, *executableDiagnostic(
				ability,
				"incomplete executable lowering",
				"the executable source backend did not consume every semantic element and source token",
			))
			continue
		}
		result.StaticAbilities = append(result.StaticAbilities, lowered.staticAbilities...)
		if lowered.activatedAbility.Exists {
			result.ActivatedAbilities = append(result.ActivatedAbilities, lowered.activatedAbility.Val)
		}
		if lowered.manaAbility.Exists {
			result.ManaAbilities = append(result.ManaAbilities, lowered.manaAbility.Val)
		}
		if lowered.loyaltyAbility.Exists {
			result.LoyaltyAbilities = append(result.LoyaltyAbilities, lowered.loyaltyAbility.Val)
		}
		if lowered.triggeredAbility.Exists {
			result.TriggeredAbilities = append(result.TriggeredAbilities, lowered.triggeredAbility.Val)
		}
		if lowered.chapterAbility.Exists {
			result.ChapterAbilities = append(result.ChapterAbilities, lowered.chapterAbility.Val)
		}
		if lowered.replacementAbility.Exists {
			result.ReplacementAbilities = append(result.ReplacementAbilities, lowered.replacementAbility.Val)
		}
		result.EntersPrepared = result.EntersPrepared || lowered.entersPrepared
		if lowered.spellAbility.Exists {
			if result.SpellAbility.Exists {
				unsupported = append(unsupported, *executableDiagnostic(
					ability,
					"unsupported multiple spell abilities",
					"the executable source backend supports only one spell ability per card face",
				))
				continue
			}
			result.SpellAbility = lowered.spellAbility
		}
	}
	for _, ability := range compilation.Abilities {
		for _, keyword := range ability.Content.Keywords {
			if keyword.Name != "Read ahead" {
				continue
			}
			sacrificeChapter, ok := readAheadSacrificeChapter(ability.Text)
			if !ok || sacrificeChapter == 0 {
				continue
			}
			finalChapter := 0
			for _, chapter := range result.ChapterAbilities {
				for _, number := range chapter.Chapters {
					finalChapter = max(finalChapter, number)
				}
			}
			if sacrificeChapter != finalChapter {
				unsupported = append(unsupported, *executableDiagnostic(
					ability,
					"unsupported Read ahead ability",
					fmt.Sprintf("the reminder sacrifice chapter %d does not match final chapter %d", sacrificeChapter, finalChapter),
				))
			}
		}
	}
	if len(unsupported) > 0 {
		return loweredFaceAbilities{}, append(diagnostics, unsupported...)
	}
	return result, diagnostics
}

func lowerExecutableAbility(
	cardName string,
	saga bool,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if lowered, handled, diagnostic := lowerExecutableAbilitySpecialCase(cardName, ability, syntax); handled {
		return lowered, diagnostic
	}
	switch ability.Kind {
	case compiler.AbilityStatic:
		bodies, diagnostic := lowerKeywordAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}

		spans := make([]shared.Span, 0, len(ability.Content.Keywords)+len(syntax.Reminders))
		if syntax.AbilityWord != nil && len(ability.Content.Keywords) > 0 {
			spans = append(spans, shared.Span{
				Start: ability.Span.Start,
				End:   ability.Content.Keywords[0].Span.Start,
			})
		}
		spans = appendKeywordSpans(spans, ability.Content.Keywords)
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			staticAbilities: bodies,
			consumed: semanticConsumption{
				keywords: len(ability.Content.Keywords),
			},
			sourceSpans: spans,
		}, nil
	case compiler.AbilityActivated:
		return lowerActivatedAbilityKind(cardName, ability, syntax)
	case compiler.AbilityLoyalty:
		return lowerLoyaltyAbility(cardName, ability, syntax)
	case compiler.AbilitySpell:
		body, bodySyntax, ok := spellBodyWithoutAbilityWord(ability, syntax)
		if !ok {
			return abilityLowering{}, executableDiagnostic(
				ability,
				"unsupported ability word",
				fmt.Sprintf("the executable source backend does not yet lower the %q ability word", ability.AbilityWord),
			)
		}
		spellAbility, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := make(
			[]shared.Span,
			0,
			len(ability.Content.Effects)+len(ability.Content.Targets)+len(ability.Content.Conditions)+len(ability.Content.References)+len(syntax.Reminders),
		)
		for _, effect := range ability.Content.Effects {
			spans = append(spans, effect.Span)
		}
		for _, target := range ability.Content.Targets {
			spans = append(spans, target.Span)
		}
		for _, condition := range ability.Content.Conditions {
			spans = append(spans, condition.Span)
		}
		for _, reference := range ability.Content.References {
			spans = append(spans, reference.Span)
		}
		spans = appendKeywordSpans(spans, ability.Content.Keywords)
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			spellAbility: opt.Val(spellAbility),
			consumed: semanticConsumption{
				targets:    len(ability.Content.Targets),
				conditions: len(ability.Content.Conditions),
				effects:    len(ability.Content.Effects),
				keywords:   len(ability.Content.Keywords),
				references: len(ability.Content.References),
			},
			sourceSpans: spans,
		}, nil
	case compiler.AbilityTriggered:
		return lowerTriggeredAbilityKind(cardName, ability, syntax)
	case compiler.AbilityChapter:
		return lowerChapterAbility(cardName, ability, syntax)
	case compiler.AbilityReplacement:
		return lowerReplacementAbility(ability)
	case compiler.AbilityReminder:
		if saga && isOrdinarySagaReminder(ability.Text) {
			return abilityLowering{sourceSpans: []shared.Span{ability.Span}}, nil
		}

		return lowerReminderManaAbility(ability)
	default:
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported "+ability.Kind.String()+" ability",
			"the executable source backend does not yet lower "+ability.Kind.String()+" abilities",
		)
	}
}

func lowerExecutableAbilitySpecialCase(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if len(ability.Content.Modes) > 0 && ability.Kind != compiler.AbilityActivated {
		lowered, diagnostic := lowerModalAbility(cardName, ability, syntax)
		return lowered, true, diagnostic
	}
	if lowered, ok := lowerEntersPrepared(ability, syntax); ok {
		return lowered, true, nil
	}
	if lowered, ok, diagnostic := lowerStaticDeclarations(ability); ok {
		return lowered, true, diagnostic
	}
	if diagnostic := lowerStaticDeclarationBlocker(ability); diagnostic != nil {
		return abilityLowering{}, true, diagnostic
	}
	if lowered, ok, diagnostic := lowerKeywordDispatch(ability, syntax); ok {
		return lowered, true, diagnostic
	}
	return abilityLowering{}, false, nil
}

func lowerReplacementAbility(ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if replacementAbility, handled, diagnostic := lowerDamageReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerCounterPlacementReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerTokenCreationReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerSelfZoneDestinationReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerEntersWithCountersReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	replacementAbility, diagnostic := lowerEntersTappedReplacement(ability)
	return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
}

func replacementAbilityLowering(ability compiler.CompiledAbility, replacementAbility *game.ReplacementAbility, diagnostic *shared.Diagnostic) (abilityLowering, *shared.Diagnostic) {
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	return abilityLowering{
		replacementAbility: opt.Val(*replacementAbility),
		consumed: semanticConsumption{
			effects:    len(ability.Content.Effects),
			conditions: len(ability.Content.Conditions),
			references: len(ability.Content.References),
		},
		sourceSpans: replacementSourceSpans(ability),
	}, nil
}

func appendKeywordSpans(spans []shared.Span, keywords []compiler.CompiledKeyword) []shared.Span {
	for _, keyword := range keywords {
		spans = append(spans, keyword.Span)
	}
	return spans
}

func replacementSourceSpans(ability compiler.CompiledAbility) []shared.Span {
	spans := make([]shared.Span, 0, len(ability.Content.Effects))
	for _, effect := range ability.Content.Effects {
		spans = append(spans, effect.Span)
	}
	return spans
}

func isOrdinarySagaReminder(text string) bool {
	const (
		withComma    = "(As this Saga enters and after your draw step, add a lore counter."
		withoutComma = "(As this Saga enters and after your draw step add a lore counter."
	)
	remainder, ok := strings.CutPrefix(text, withComma)
	if !ok {
		remainder, ok = strings.CutPrefix(text, withoutComma)
	}
	if !ok {
		return false
	}
	if remainder == ")" {
		return true
	}
	const sacrificePrefix = " Sacrifice after "
	chapter, ok := strings.CutPrefix(remainder, sacrificePrefix)
	if !ok || !strings.HasSuffix(chapter, ".)") {
		return false
	}
	chapter = strings.TrimSuffix(chapter, ".)")
	switch chapter {
	case "I", "II", "III", "IV", "V", "VI":
		return true
	default:
		return false
	}
}

func lowerChapterAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if len(ability.Chapters) == 0 || ability.ChapterSpan == (shared.Span{}) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires one or more chapter numbers",
		)
	}
	dash := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
		return token.Kind == shared.EmDash
	})
	if dash < 0 || dash+1 >= len(syntax.Tokens) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires an em dash after the chapter numbers",
		)
	}
	bodySpan := shared.Span{
		Start: syntax.Tokens[dash+1].Span.Start,
		End:   syntax.Span.End,
	}
	bodyText := strings.TrimSpace(
		ability.Text[bodySpan.Start.Offset-ability.Span.Start.Offset:],
	)
	bodyContent := ability.Content
	bodyContent.Keywords = keywordsWithinSpan(ability.Content.Keywords, bodySpan)
	if len(bodyContent.Keywords) != len(ability.Content.Keywords) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires chapter keywords to belong to a supported effect",
		)
	}
	bodySyntax := parser.Ability{
		Span:      bodySpan,
		Text:      bodyText,
		Tokens:    slices.Clone(syntax.Tokens[dash+1:]),
		Reminders: syntax.Reminders,
		Quoted:    syntax.Quoted,
		Atoms:     syntax.Atoms,
	}
	content, diagnostic := lowerAbilityContent(cardName, bodyContent, false, bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := []shared.Span{ability.ChapterSpan, syntax.Tokens[dash].Span}
	for _, effect := range ability.Content.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	for _, keyword := range ability.Content.Keywords {
		spans = append(spans, keyword.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		chapterAbility: opt.Val(game.ChapterAbility{
			Text:     ability.Text,
			Chapters: slices.Clone(ability.Chapters),
			Content:  content,
		}),
		consumed: semanticConsumption{
			targets:    len(ability.Content.Targets),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

func lowerEntersPrepared(ability compiler.CompiledAbility, syntax parser.Ability) (abilityLowering, bool) {
	const text = "This creature enters prepared."
	if ability.Kind != compiler.AbilityStatic ||
		(ability.Text != text && !strings.HasPrefix(ability.Text, text+" (")) ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterPrepared ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != compiler.ReferenceBindingSource ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil {
		return abilityLowering{}, false
	}
	return abilityLowering{
		entersPrepared: true,
		consumed: semanticConsumption{
			effects:    1,
			references: 1,
		},
		sourceSpans: []shared.Span{syntax.Span},
	}, true
}

func lowerActivatedAbilityKind(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if isSemanticManaAbility(ability) {
		manaAbility, diagnostic := lowerManaAbility(cardName, ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []shared.Span{ability.Cost.Span}
		for _, effect := range ability.Content.Effects {
			spans = append(spans, effect.Span)
		}
		spans = append(spans, activationConditionSourceSpans(ability, syntax)...)
		if ability.ActivationTiming != compiler.ActivationTimingNone {
			spans = append(spans, ability.ActivationTimingSpan)
		}
		for _, reference := range ability.Content.References {
			spans = append(spans, reference.Span)
		}
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed: semanticConsumption{
				cost:       true,
				conditions: len(ability.Content.Conditions),
				effects:    len(ability.Content.Effects),
				references: len(ability.Content.References),
			},
			sourceSpans: spans,
		}, nil
	}
	activatedAbility, diagnostic := lowerActivatedAbility(cardName, ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := make(
		[]shared.Span,
		0,
		1+len(ability.Content.Effects)+len(ability.Content.Targets)+len(ability.Content.References)+len(syntax.Reminders),
	)
	spans = append(spans, ability.Cost.Span)
	if ability.ActivationTiming != compiler.ActivationTimingNone {
		spans = append(spans, ability.ActivationTimingSpan)
	}
	for _, effect := range ability.Content.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	spans = append(spans, activationConditionSourceSpans(ability, syntax)...)
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	if len(ability.Content.Modes) > 0 {
		spans = append(spans, ability.Span)
	}
	return abilityLowering{
		activatedAbility: opt.Val(activatedAbility),
		consumed: semanticConsumption{
			cost:       true,
			modes:      len(ability.Content.Modes),
			targets:    len(ability.Content.Targets),
			conditions: len(ability.Content.Conditions),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

func isSemanticManaAbility(ability compiler.CompiledAbility) bool {
	return !abilityContentHasTargets(ability.Content) && abilityContentHasAddManaEffect(ability.Content)
}

func abilityContentHasAddManaEffect(content compiler.AbilityContent) bool {
	if slices.ContainsFunc(content.Effects, func(effect compiler.CompiledEffect) bool {
		return effect.Kind == compiler.EffectAddMana
	}) {
		return true
	}
	return slices.ContainsFunc(content.Modes, func(mode compiler.CompiledMode) bool {
		return abilityContentHasAddManaEffect(mode.Content)
	})
}

func abilityContentHasTargets(content compiler.AbilityContent) bool {
	if len(content.Targets) != 0 {
		return true
	}
	return slices.ContainsFunc(content.Modes, func(mode compiler.CompiledMode) bool {
		return abilityContentHasTargets(mode.Content)
	})
}

// lowerLoyaltyAbility lowers an AbilityLoyalty into a game.LoyaltyAbility.
// It accepts only exact signed integer loyalty costs and supported single or
// ordered effect bodies. Variable costs (X) are rejected.
func lowerLoyaltyAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	const unsupportedDetail = "the executable source backend supports only exact signed loyalty costs with a supported effect body"
	if ability.Cost == nil ||
		len(ability.Cost.Components) != 1 ||
		ability.Cost.Components[0].Kind != compiler.CostLoyalty ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.AbilityWord != "" {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	loyaltyCost, ok := parseLoyaltyCostAmount(ability.Cost.Components[0].Amount)
	if !ok {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", "the executable source backend supports only fixed integer loyalty costs, not variable costs")
	}

	colon := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
		return token.Kind == shared.Colon
	})
	if colon < 0 || colon+1 >= len(syntax.Tokens) {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	bodySpan := shared.Span{
		Start: syntax.Tokens[colon+1].Span.Start,
		End:   syntax.Span.End,
	}
	bodyText := strings.TrimSpace(ability.Text[bodySpan.Start.Offset-ability.Span.Start.Offset:])
	bodyContent := ability.Content
	bodyContent.Keywords = keywordsWithinSpan(ability.Content.Keywords, bodySpan)
	if len(bodyContent.Keywords) != len(ability.Content.Keywords) {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	bodySyntax := parser.Ability{
		Span:      bodySpan,
		Text:      bodyText,
		Tokens:    syntax.Tokens[colon+1:],
		Reminders: syntax.Reminders,
		Quoted:    syntax.Quoted,
		Atoms:     syntax.Atoms,
	}
	content, diagnostic := lowerAbilityContent(cardName, bodyContent, false, bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}

	spans := make(
		[]shared.Span,
		0,
		1+len(ability.Content.Effects)+len(ability.Content.Targets)+len(ability.Content.References)+len(syntax.Reminders),
	)
	spans = append(spans, ability.Cost.Span)
	for _, effect := range ability.Content.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		loyaltyAbility: opt.Val(game.LoyaltyAbility{
			Text:        ability.Text,
			LoyaltyCost: loyaltyCost,
			Content:     content,
		}),
		consumed: semanticConsumption{
			cost:       true,
			targets:    len(ability.Content.Targets),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

// parseLoyaltyCostAmount converts a loyalty cost amount string such as "+1",
// "−2", or "0" into a signed integer. It returns false for variable costs
// (e.g. "+X") or malformed input.
func parseLoyaltyCostAmount(amount string) (int, bool) {
	if amount == "" {
		return 0, false
	}
	rest := amount
	sign := 1
	switch {
	case strings.HasPrefix(rest, "+"):
		rest = rest[1:]
	case strings.HasPrefix(rest, "\u2212"):
		// Unicode minus sign U+2212 (3 bytes in UTF-8)
		sign = -1
		rest = rest[len("\u2212"):]
	case strings.HasPrefix(rest, "-"):
		sign = -1
		rest = rest[1:]
	default:
		// no sign prefix — treat as positive (e.g., "0")
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return sign * n, true
}

// lowerModalAbility lowers a modal spell/static shell. The modal body itself is
// lowered exclusively through lowerAbilityContent.
func lowerModalAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		ability.AbilityWord != "" {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported ability modes",
			"the executable source backend cannot lower this modal ability shell",
		)
	}
	switch ability.Kind {
	case compiler.AbilitySpell, compiler.AbilityStatic:
	default:
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported ability modes",
			"the executable source backend supports only spell or static modal abilities",
		)
	}
	content, diagnostic := lowerAbilityContent(cardName, ability.Content, ability.Optional, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	return abilityLowering{
		spellAbility: opt.Val(content),
		consumed: semanticConsumption{
			modes: len(ability.Content.Modes),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// parseChooseHeader inspects a modal header phrase and returns (minModes,
// maxModes, ok). It accepts "Choose <word> —" where <word> is a cardinal
// number spelled out as a single word ("one", "two", etc.), plus exact
// "Choose one or both —" headers.
func parseChooseHeader(header parser.Phrase, atoms parser.Atoms) (minModes, maxModes int, ok bool) {
	tokens := header.Tokens
	if len(tokens) == 5 &&
		tokens[0].Kind == shared.Word && strings.EqualFold(tokens[0].Text, "choose") &&
		tokens[1].Kind == shared.Word && strings.EqualFold(tokens[1].Text, "one") &&
		tokens[2].Kind == shared.Word && strings.EqualFold(tokens[2].Text, "or") &&
		tokens[3].Kind == shared.Word && strings.EqualFold(tokens[3].Text, "both") &&
		tokens[4].Kind == shared.EmDash {
		return 1, 2, true
	}
	// Expected: [Word("Choose"), Word(<number>), EmDash]
	if len(tokens) != 3 ||
		tokens[0].Kind != shared.Word || !strings.EqualFold(tokens[0].Text, "choose") ||
		tokens[1].Kind != shared.Word ||
		tokens[2].Kind != shared.EmDash {
		return 0, 0, false
	}
	n, numOK := atoms.CardinalAt(tokens[1].Span)
	if !numOK {
		return 0, 0, false
	}
	return n, n, true
}

func lowerModalContent(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported ability modes", detail)
	}
	if syntax.Modal == nil {
		return unsupported("the semantic modal content has no matching modal syntax")
	}
	minModes, maxModes, ok := parseChooseHeader(syntax.Modal.Header, syntax.Modal.Atoms)
	if !ok {
		return unsupported("the executable source backend supports only exact \"Choose N\" and \"Choose one or both\" modal headers")
	}
	if minModes < 1 || maxModes < minModes || maxModes > len(ctx.content.Modes) ||
		(minModes == 1 && maxModes == 2 && len(ctx.content.Modes) != 2) {
		return unsupported("the modal choice range does not match the number of modes")
	}
	if ctx.optional ||
		len(ctx.content.Effects) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported("the executable source backend does not support shared targets, effects, keywords, conditions, or references across modes")
	}
	if len(ctx.content.Modes) != len(syntax.Modal.Options) {
		return unsupported("semantic mode count does not match syntax mode count")
	}

	modes := make([]game.Mode, 0, len(ctx.content.Modes))
	for i, mode := range ctx.content.Modes {
		syntaxMode := syntax.Modal.Options[i]
		bodySyntax := parser.Ability{
			Span:      syntaxMode.Span,
			Text:      syntaxMode.Text,
			Tokens:    syntaxMode.Tokens,
			Reminders: syntaxMode.Reminders,
			Quoted:    syntaxMode.Quoted,
			Atoms:     syntaxMode.Atoms,
		}
		content, diagnostic := lowerAbilityContent(cardName, mode.Content, false, bodySyntax)
		if diagnostic != nil {
			return game.AbilityContent{}, diagnostic
		}
		if content.IsModal() || len(content.Modes) != 1 {
			return unsupported("mode lowering produced unexpected modal content")
		}
		if !modalOptionCompletelyRecognized(mode.Content, syntaxMode) {
			return unsupported("a modal option contains rules text without complete executable semantics")
		}
		modes = append(modes, content.Modes[0])
	}
	return game.AbilityContent{
		Modes:    modes,
		MinModes: minModes,
		MaxModes: maxModes,
	}, nil
}

func modalOptionCompletelyRecognized(content compiler.AbilityContent, syntax parser.Mode) bool {
	var spans []shared.Span
	for _, effect := range content.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range content.Targets {
		spans = append(spans, target.Span)
	}
	for _, condition := range content.Conditions {
		spans = append(spans, condition.Span)
	}
	for _, reference := range content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	for _, token := range syntax.Tokens {
		if token.Kind == shared.Comma ||
			token.Kind == shared.Period ||
			spanCovered(token.Span, spans) {
			continue
		}
		return false
	}
	return true
}

func lowerActivatedAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.ActivatedAbility, *shared.Diagnostic) {
	shell, diagnostic := lowerActivationShell(cardName, ability, syntax)
	if diagnostic != nil {
		return game.ActivatedAbility{}, diagnostic
	}

	result := game.ActivatedAbility{
		Text:                shell.text,
		ManaCost:            shell.manaCost,
		AdditionalCosts:     shell.additionalCosts,
		ZoneOfFunction:      shell.zoneOfFunction,
		Timing:              shell.timing,
		ActivationCondition: shell.activationCondition,
		Content:             shell.content,
	}
	return result, nil
}

func prepareActivationCondition(ability *compiler.CompiledAbility, syntax *parser.Ability) (opt.V[game.Condition], bool) {
	if len(ability.Content.Conditions) == 0 {
		*syntax = syntaxWithoutAbilityWord(*syntax)
		return opt.V[game.Condition]{}, true
	}
	if len(ability.Content.Conditions) != 1 {
		return opt.V[game.Condition]{}, false
	}
	condition, ok := lowerCondition(ability.Content.Conditions[0], conditionContextActivation)
	if !ok {
		return opt.V[game.Condition]{}, false
	}
	conditionSpan := []shared.Span{ability.Content.Conditions[0].Span}
	effects := slices.DeleteFunc(append([]compiler.CompiledEffect(nil), ability.Content.Effects...), func(effect compiler.CompiledEffect) bool {
		return spanCovered(effect.VerbSpan, conditionSpan)
	})
	bodyEffects := append([]compiler.CompiledEffect(nil), effects...)
	bodyEffects = appendModeEffects(bodyEffects, ability.Content.Modes)
	if len(bodyEffects) == 0 || slices.ContainsFunc(bodyEffects, func(effect compiler.CompiledEffect) bool {
		return effect.Span.End.Offset > ability.Content.Conditions[0].Span.Start.Offset
	}) {
		return opt.V[game.Condition]{}, false
	}
	ability.Content.Effects = effects
	ability.Content.Conditions = nil
	*syntax = syntaxWithoutAbilityWord(*syntax)
	lastEffectEnd := bodyEffects[0].Span.End.Offset
	for _, effect := range bodyEffects[1:] {
		lastEffectEnd = max(lastEffectEnd, effect.Span.End.Offset)
	}
	syntax.Tokens = slices.DeleteFunc(append([]shared.Token(nil), syntax.Tokens...), func(token shared.Token) bool {
		return token.Span.Start.Offset >= lastEffectEnd
	})
	return opt.Val(condition), true
}

func appendModeEffects(effects []compiler.CompiledEffect, modes []compiler.CompiledMode) []compiler.CompiledEffect {
	for _, mode := range modes {
		effects = append(effects, mode.Content.Effects...)
		effects = appendModeEffects(effects, mode.Content.Modes)
	}
	return effects
}

func activationConditionSourceSpans(ability compiler.CompiledAbility, syntax parser.Ability) []shared.Span {
	spans := make([]shared.Span, 0, len(ability.Content.Conditions)+1)
	for _, condition := range ability.Content.Conditions {
		spans = append(spans, condition.Span)
		firstConditionToken := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
			return spanCovered(token.Span, []shared.Span{condition.Span})
		})
		if condition.Kind == compiler.ConditionOnlyIf &&
			firstConditionToken > 0 &&
			equalTokenWord(syntax.Tokens[firstConditionToken-1], "activate") {
			spans = append(spans, syntax.Tokens[firstConditionToken-1].Span)
		}
	}
	return spans
}

func lowerActivationTiming(timing compiler.ActivationTimingKind) (game.TimingRestriction, bool) {
	switch timing {
	case compiler.ActivationTimingNone:
		return game.NoTimingRestriction, true
	case compiler.ActivationTimingSorcery:
		return game.SorceryOnly, true
	case compiler.ActivationTimingOncePerTurn:
		return game.OncePerTurn, true
	case compiler.ActivationTimingSorceryOncePerTurn:
		return game.SorceryOncePerTurn, true
	case compiler.ActivationTimingDuringCombat:
		return game.DuringCombat, true
	case compiler.ActivationTimingDuringUpkeep:
		return game.DuringUpkeep, true
	default:
		return game.NoTimingRestriction, false
	}
}

func lowerActivatedAdditionalCost(cardName string, component compiler.CostComponent) (cost.Additional, bool) {
	switch component.Kind {
	case compiler.CostSacrifice:
		return lowerSacrificeCost(cardName, component)
	case compiler.CostDiscard:
		return lowerDiscardCost(component)
	case compiler.CostPayLife:
		if !component.AmountKnown || component.AmountValue <= 0 {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalPayLife,
			Text:   component.Text,
			Amount: component.AmountValue,
		}, true
	case compiler.CostExile:
		if component.SourceSelf {
			return cost.Additional{
				Kind:   cost.AdditionalExileSource,
				Text:   component.Text,
				Amount: 1,
				Source: zone.Battlefield,
			}, true
		}
		return lowerExileCost(component)
	case compiler.CostReveal:
		return lowerRevealCost(component)
	case compiler.CostRemoveCounter:
		return lowerRemoveCounterCost(cardName, component)
	case compiler.CostTapPermanents:
		return lowerTapPermanentsCost(component)
	case compiler.CostEnergy:
		if !component.AmountKnown || component.AmountValue <= 0 {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalEnergy,
			Text:   component.Text,
			Amount: component.AmountValue,
		}, true
	case compiler.CostReturn:
		return lowerReturnToHandCost(component)
	case compiler.CostExert:
		return lowerExertCost(cardName, component)
	case compiler.CostMill:
		return lowerMillCost(component)
	case compiler.CostPutCounter:
		return lowerPutCounterCost(cardName, component)
	case compiler.CostCollectEvidence:
		return lowerCollectEvidenceCost(component)
	default:
		return cost.Additional{}, false
	}
}

// lowerActivationCostComponents is the shared cost-parsing kernel used by both
// lowerActivatedAbility and lowerManaAbility. It iterates the compiled cost
// components and produces (manaCost, additionalCosts):
//
//   - CostMana must be the first component and may appear at most once.
//   - CostTap and CostUntap each may appear at most once.
//   - All other cost kinds are delegated to lowerActivatedAdditionalCost,
//     which covers sacrifice, discard, pay-life, exile, reveal, remove-counter,
//     tap-permanents, energy, return, exert, mill, put-counter, and
//     collect-evidence.
//
// Returns nil, nil, false if any component is unsupported or ordering rules are
// violated. The caller must check that ability.Cost is non-nil and non-empty
// before calling.
func lowerActivationCostComponents(
	cardName string,
	compiled *compiler.CompiledCost,
) (cost.Mana, []cost.Additional, bool) {
	var manaCost cost.Mana
	var additionalCosts []cost.Additional
	for i, component := range compiled.Components {
		switch component.Kind {
		case compiler.CostMana:
			if i != 0 || manaCost != nil {
				return nil, nil, false
			}
			parsed, err := parseManaCostValue(component.Symbol)
			if err != nil || len(parsed) == 0 {
				return nil, nil, false
			}
			manaCost = parsed
		case compiler.CostTap:
			if slices.ContainsFunc(additionalCosts, func(a cost.Additional) bool {
				return a.Kind == cost.AdditionalTap
			}) {
				return nil, nil, false
			}
			additionalCosts = append(additionalCosts, cost.T)
		case compiler.CostUntap:
			if slices.ContainsFunc(additionalCosts, func(a cost.Additional) bool {
				return a.Kind == cost.AdditionalUntap
			}) {
				return nil, nil, false
			}
			additionalCosts = append(additionalCosts, cost.Additional{
				Kind: cost.AdditionalUntap,
				Text: component.Text,
			})
		default:
			additional, ok := lowerActivatedAdditionalCost(cardName, component)
			if !ok {
				return nil, nil, false
			}
			additionalCosts = append(additionalCosts, additional)
		}
	}
	return manaCost, additionalCosts, true
}

func lowerCollectEvidenceCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown || component.AmountValue <= 0 {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:   cost.AdditionalCollectEvidence,
		Text:   component.Text,
		Amount: component.AmountValue,
		Source: zone.Graveyard,
	}, true
}

func lowerExertCost(_ string, component compiler.CostComponent) (cost.Additional, bool) {
	if !component.SourceSelf {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind: cost.AdditionalExert,
		Text: component.Text,
	}, true
}

func lowerMillCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown || component.ObjectKind != compiler.SelectorCard {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:   cost.AdditionalMill,
		Text:   component.Text,
		Amount: component.AmountValue,
	}, true
}

func lowerPutCounterCost(_ string, component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown || !component.CounterKindKnown || !component.SourceSelf {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:        cost.AdditionalPutCounter,
		Text:        component.Text,
		Amount:      component.AmountValue,
		CounterKind: component.CounterKind,
	}, true
}

func lowerRevealCost(component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceZone != zone.Hand || component.ObjectKind != compiler.SelectorCard {
		return cost.Additional{}, false
	}
	if !component.AmountKnown && !component.AmountFromX {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalReveal,
		Text:   component.Text,
		Source: zone.Hand,
	}
	if component.AmountFromX {
		additional.AmountFromX = true
	} else {
		additional.Amount = component.AmountValue
	}
	if component.ObjectColorKnown {
		additional.MatchCardColor = true
		additional.CardColor = component.ObjectColor
	}
	return additional, true
}

func lowerReturnToHandCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown ||
		component.ObjectController != compiler.ControllerYou ||
		component.ToZone != zone.Hand {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:          cost.AdditionalReturnToHand,
		Text:          component.Text,
		Amount:        component.AmountValue,
		RequireTapped: component.RequireTapped,
	}
	if lowerCostPermanentObject(component, &additional, true) {
		return additional, true
	}
	return cost.Additional{}, false
}

func lowerCostPermanentObject(component compiler.CostComponent, additional *cost.Additional, allowSnowLand bool) bool {
	if component.ObjectColorKnown || component.ObjectNonToken {
		return false
	}
	switch component.ObjectKind {
	case compiler.SelectorPermanent:
		return true
	case compiler.SelectorArtifact, compiler.SelectorCreature, compiler.SelectorEnchantment, compiler.SelectorLand:
		additional.MatchPermanentType = true
		additional.PermanentType = component.ObjectType
		if allowSnowLand && component.SupertypeKnown && component.ObjectSupertype == types.Snow && component.ObjectType == types.Land {
			additional.RequireSupertype = types.Snow
		}
		return true
	default:
	}
	if len(component.SubtypesAny) == 1 {
		additional.SubtypesAny = cost.SubtypeSet{component.SubtypesAny[0]}
		return true
	}
	return false
}

func lowerTapPermanentsCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown ||
		!component.RequireUntapped ||
		component.ObjectController != compiler.ControllerYou {
		return cost.Additional{}, false
	}
	if len(component.SubtypesAny) == 1 && component.SubtypesAny[0] == types.Zombie && component.AmountValue >= 2 {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalTapPermanents,
		Text:   component.Text,
		Amount: component.AmountValue,
	}
	if lowerCostPermanentObject(component, &additional, false) {
		return additional, true
	}
	return cost.Additional{}, false
}

func lowerRemoveCounterCost(
	_ string,
	component compiler.CostComponent,
) (cost.Additional, bool) {
	if !component.AmountKnown || !component.CounterKindKnown || !component.SourceSelf {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:        cost.AdditionalRemoveCounter,
		Text:        component.Text,
		Amount:      component.AmountValue,
		CounterKind: component.CounterKind,
	}, true
}

func lowerExileCost(component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceZone != zone.Graveyard ||
		component.ObjectKind != compiler.SelectorCard ||
		!component.AmountKnown {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalExile,
		Text:   component.Text,
		Amount: component.AmountValue,
		Source: zone.Graveyard,
	}
	if component.ObjectTypeKnown {
		additional.MatchCardType = true
		additional.CardType = component.ObjectType
	}
	return additional, true
}

func lowerSacrificeCost(_ string, component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceSelf {
		return cost.Additional{
			Kind:   cost.AdditionalSacrificeSource,
			Text:   component.Text,
			Amount: 1,
		}, true
	}
	if !component.AmountKnown {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalSacrifice,
		Text:   component.Text,
		Amount: component.AmountValue,
	}
	if !lowerCostPermanentObject(component, &additional, false) {
		return cost.Additional{}, false
	}
	return additional, true
}

func lowerDiscardCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown ||
		component.ObjectKind != compiler.SelectorCard ||
		component.ObjectColorKnown ||
		component.ObjectNonToken ||
		component.PermanentModifier {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalDiscard,
		Text:   component.Text,
		Amount: component.AmountValue,
		Source: zone.Hand,
	}
	if component.ObjectTypeKnown {
		additional.MatchCardType = true
		additional.CardType = component.ObjectType
	}
	return additional, true
}

func lowerEnchantAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Name != "Enchant" {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	target, ok := enchantTargetSpec(keyword.Parameter)
	if !ok ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Enchant ability",
			"the executable source backend supports only exact Enchant with a supported target kind",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Enchant ability",
			"the executable source backend supports only exact Enchant with a supported target kind",
		)
	}
	return game.EnchantStaticAbility(&target), true, nil
}

func lowerProtectionAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Name != "Protection" {
		return game.StaticAbility{}, false, nil
	}
	// If the ability has effects, it is a grant (e.g., "Enchanted creature has
	// protection from X") — defer to Static Declaration lowering instead.
	if len(ability.Content.Effects) > 0 {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]

	// Common structural checks for all protection variants.
	structureOK := ability.Kind == compiler.AbilityStatic &&
		ability.Cost == nil &&
		ability.Trigger == nil &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.Effects) == 0 &&
		len(ability.Content.References) == 0 &&
		ability.AbilityWord == ""

	unsupported := func() (game.StaticAbility, bool, *shared.Diagnostic) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Protection ability",
			"the executable source backend supports only exact fixed-predicate protection",
		)
	}

	if !structureOK {
		return unsupported()
	}

	// Validate that the syntax tokens are fully covered by the keyword span.
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return unsupported()
	}

	if !keyword.ProtectionKnown || !protectionKeywordRuntimeSupported(keyword.Protection) {
		return unsupported()
	}
	return staticAbilityFromProtectionKeyword(keyword.Protection, ability.Text), true, nil
}

func protectionKeywordRuntimeSupported(prot game.ProtectionKeyword) bool {
	for _, sub := range prot.FromSubtypes {
		if !parser.SubtypeMatchesAnyRuntimeCardType(sub, []types.Card{types.Creature, types.Land}) {
			return false
		}
	}
	return true
}

// lowerKeywordDispatch tries Enchant, Protection, Equip, Cycling, Ninjutsu, and
// Mutate — the
// single-keyword special cases that each produce a full abilityLowering.
// Returns (lowering, true, nil) on success, (lowering, true, diag) on a
// recognized-but-rejected attempt, and ({}, false, nil) when no attempt matches.
func lowerKeywordDispatch(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if enchantAbility, ok, diag := lowerEnchantAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&enchantAbility, ability, syntax), true, nil
	}
	if protectionAbility, ok, diag := lowerProtectionAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&protectionAbility, ability, syntax), true, nil
	}
	if equipAbility, ok, diag := lowerEquipAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&equipAbility, ability, syntax), true, nil
	}
	if cyclingAbility, ok, diag := lowerCyclingAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&cyclingAbility, ability, syntax), true, nil
	}
	if ninjutsuAbility, ok, diag := lowerNinjutsuAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&ninjutsuAbility, ability, syntax), true, nil
	}
	if mutateAbility, ok, diag := lowerMutateAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&mutateAbility, ability, syntax), true, nil
	}
	return abilityLowering{}, false, nil
}

func keywordStaticLowering(
	body *game.StaticAbility,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) abilityLowering {
	spans := keywordSpans(ability, syntax)
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{Body: *body}},
		consumed:        semanticConsumption{keywords: 1},
		sourceSpans:     spans,
	}
}

func keywordActivatedLowering(
	body *game.ActivatedAbility,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) abilityLowering {
	spans := keywordSpans(ability, syntax)
	return abilityLowering{
		activatedAbility: opt.Val(*body),
		consumed:         semanticConsumption{keywords: 1},
		sourceSpans:      spans,
	}
}

func keywordSpans(ability compiler.CompiledAbility, syntax parser.Ability) []shared.Span {
	spans := []shared.Span{ability.Content.Keywords[0].Span}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

// staticAbilityFromProtectionKeyword builds a StaticAbility from a
// ProtectionKeyword using the appropriate factory function.
func staticAbilityFromProtectionKeyword(prot game.ProtectionKeyword, text string) game.StaticAbility {
	switch {
	case prot.Everything:
		return game.ProtectionFromEverythingStaticAbility()
	case prot.EachColor:
		return game.ProtectionFromEachColorStaticAbility()
	case prot.Multicolored:
		return game.ProtectionFromMulticoloredStaticAbility()
	case prot.Monocolored:
		return game.ProtectionFromMonocoloredStaticAbility()
	case len(prot.FromTypes) > 0:
		return game.ProtectionFromTypesStaticAbility(prot.FromTypes...)
	case len(prot.FromSubtypes) > 0:
		return game.ProtectionFromSubtypesStaticAbility(prot.FromSubtypes...)
	case len(prot.FromColors) > 0:
		return game.ProtectionFromColorsStaticAbility(prot.FromColors...)
	default:
		panic(fmt.Sprintf("lower: empty ProtectionKeyword for %q", text))
	}
}

func enchantTargetSpec(parameter string) (game.TargetSpec, bool) {
	target := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: parameter,
	}
	if parameter == "player" {
		target.Allow = game.TargetAllowPlayer
		return target, true
	}
	target.Allow = game.TargetAllowPermanent
	switch parameter {
	case "artifact":
		target.Predicate.PermanentTypes = []types.Card{types.Artifact}
	case "creature":
		target.Predicate.PermanentTypes = []types.Card{types.Creature}
	case "enchantment":
		target.Predicate.PermanentTypes = []types.Card{types.Enchantment}
	case "land":
		target.Predicate.PermanentTypes = []types.Card{types.Land}
	case "permanent":
	case "planeswalker":
		target.Predicate.PermanentTypes = []types.Card{types.Planeswalker}
	default:
		return game.TargetSpec{}, false
	}
	return target, true
}

func lowerEquipAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Name != "Equip" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.Parameter == "" ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	return game.EquipActivatedAbility(manaCost), true, nil
}

func lowerCyclingAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Name != "Cycling" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.Parameter == "" && (len(ability.Content.Targets) != 0 || len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0) {
		return game.ActivatedAbility{}, false, nil
	}
	if keyword.Parameter == "" ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	return game.CyclingActivatedAbility(manaCost), true, nil
}

func lowerNinjutsuAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Name != "Ninjutsu" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.Parameter == "" ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	return game.NinjutsuActivatedAbility(manaCost), true, nil
}

func lowerMutateAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Name != "Mutate" {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.Parameter == "" ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	return game.MutateStaticAbility(manaCost), true, nil
}

func mixedStaticKeywords(keywords []compiler.CompiledKeyword) ([]game.Keyword, bool) {
	result := make([]game.Keyword, 0, len(keywords))
	for _, keyword := range keywords {
		if keyword.Parameter != "" {
			return nil, false
		}
		body, ok := keywordStaticBodies[keyword.Name]
		if !ok || len(body.Body.KeywordAbilities) != 1 {
			return nil, false
		}
		simple, ok := body.Body.KeywordAbilities[0].(game.SimpleKeyword)
		if !ok || !mixedStaticKeywordImplemented(simple.Kind) {
			return nil, false
		}
		result = append(result, simple.Kind)
	}
	return result, true
}

func abilityKeywordsExcludingSelectorPredicates(content compiler.AbilityContent) []compiler.CompiledKeyword {
	if !abilityUsesCyclingSelectorPredicate(content) {
		return content.Keywords
	}
	filtered := make([]compiler.CompiledKeyword, 0, len(content.Keywords))
	for _, keyword := range content.Keywords {
		if keyword.Name == "Cycling" && keyword.Parameter == "" {
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

func abilityUsesCyclingSelectorPredicate(content compiler.AbilityContent) bool {
	for _, target := range content.Targets {
		if strings.EqualFold(target.Selector.Keyword, "Cycling") {
			return true
		}
	}
	for _, effect := range content.Effects {
		if strings.EqualFold(effect.Selector.Keyword, "Cycling") ||
			strings.EqualFold(effect.Amount.Selector().Keyword, "Cycling") {
			return true
		}
	}
	return false
}

func mixedStaticKeywordImplemented(keyword game.Keyword) bool {
	switch keyword {
	case game.Deathtouch,
		game.Defender,
		game.DoubleStrike,
		game.FirstStrike,
		game.Flying,
		game.Haste,
		game.Hexproof,
		game.Indestructible,
		game.Lifelink,
		game.Menace,
		game.Reach,
		game.Shroud,
		game.Trample,
		game.Vigilance,
		game.Wither:
		return true
	default:
		return false
	}
}

func resolvingStaticSubjectGroup(effect compiler.CompiledEffect) (game.GroupReference, bool) {
	selection := game.Selection{Controller: game.ControllerYou}
	switch effect.StaticSubject {
	case compiler.StaticSubjectControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
	case compiler.StaticSubjectControlledWalls:
		selection.SubtypesAny = []types.Sub{types.Wall}
	case compiler.StaticSubjectControlledArtifacts:
		selection.RequiredTypes = []types.Card{types.Artifact}
	case compiler.StaticSubjectControlledTokens:
		selection.TokenOnly = true
	case compiler.StaticSubjectOpponentControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
		selection.Controller = game.ControllerOpponent
	case compiler.StaticSubjectControlledCreatureSubtype:
		if !effect.StaticSubjectSubKnown() {
			return game.GroupReference{}, false
		}
		selection.SubtypesAny = []types.Sub{effect.StaticSubjectSub()}
	default:
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

func matchesExactKeywordList(tokens []shared.Token, keywords []compiler.CompiledKeyword) bool {
	elements := make([]string, 0, len(tokens))
	lastKeyword := -1
	for _, token := range tokens {
		keywordIndex := -1
		for i, keyword := range keywords {
			if spanCovered(token.Span, []shared.Span{keyword.Span}) {
				keywordIndex = i
				break
			}
		}
		if keywordIndex >= 0 {
			if keywordIndex != lastKeyword {
				elements = append(elements, "keyword")
				lastKeyword = keywordIndex
			}
			continue
		}
		lastKeyword = -1
		switch {
		case token.Kind == shared.Comma:
			elements = append(elements, "comma")
		case equalTokenWord(token, "and"):
			elements = append(elements, "and")
		default:
			return false
		}
	}
	if len(keywords) == 1 {
		return slices.Equal(elements, []string{"keyword"})
	}
	if len(keywords) == 2 {
		return slices.Equal(elements, []string{"keyword", "and", "keyword"})
	}
	position := 0
	for keywordIndex := range keywords {
		if position >= len(elements) || elements[position] != "keyword" {
			return false
		}
		position++
		if keywordIndex == len(keywords)-1 {
			return position == len(elements)
		}
		if keywordIndex == len(keywords)-2 {
			if position < len(elements) && elements[position] == "comma" {
				position++
			}
			if position >= len(elements) || elements[position] != "and" {
				return false
			}
			position++
			continue
		}
		if position >= len(elements) || elements[position] != "comma" {
			return false
		}
		position++
	}
	return false
}

func matchesStaticPTBuffPrefix(
	tokens []shared.Token,
	effect compiler.CompiledEffect,
) (int, bool) {
	switch effect.StaticSubject {
	case compiler.StaticSubjectAttachedObject:
		return 8, len(tokens) >= 8 &&
			(equalTokenWord(tokens[0], "enchanted") || equalTokenWord(tokens[0], "equipped")) &&
			equalTokenWord(tokens[1], "creature") &&
			equalTokenWord(tokens[2], "gets") &&
			tokensMatchSignedAmount(tokens[3], tokens[4], effect.PowerDelta) &&
			tokens[5].Kind == shared.Slash &&
			tokensMatchSignedAmount(tokens[6], tokens[7], effect.ToughnessDelta)
	case compiler.StaticSubjectControlledCreatures:
		return 9, len(tokens) >= 9 &&
			equalTokenWord(tokens[0], "creatures") &&
			equalTokenWord(tokens[1], "you") &&
			equalTokenWord(tokens[2], "control") &&
			equalTokenWord(tokens[3], "get") &&
			tokensMatchSignedAmount(tokens[4], tokens[5], effect.PowerDelta) &&
			tokens[6].Kind == shared.Slash &&
			tokensMatchSignedAmount(tokens[7], tokens[8], effect.ToughnessDelta)
	case compiler.StaticSubjectOtherControlledCreatures:
		return 10, len(tokens) >= 10 &&
			equalTokenWord(tokens[0], "other") &&
			equalTokenWord(tokens[1], "creatures") &&
			equalTokenWord(tokens[2], "you") &&
			equalTokenWord(tokens[3], "control") &&
			equalTokenWord(tokens[4], "get") &&
			tokensMatchSignedAmount(tokens[5], tokens[6], effect.PowerDelta) &&
			tokens[7].Kind == shared.Slash &&
			tokensMatchSignedAmount(tokens[8], tokens[9], effect.ToughnessDelta)
	case compiler.StaticSubjectControlledWalls:
		offset := 0
		noun := "walls"
		verb := "get"
		if len(tokens) > 0 && equalTokenWord(tokens[0], "each") {
			offset = 1
			noun = "wall"
			verb = "gets"
		}
		return 9 + offset, len(tokens) >= 9+offset &&
			equalTokenWord(tokens[offset], noun) &&
			equalTokenWord(tokens[offset+1], "you") &&
			equalTokenWord(tokens[offset+2], "control") &&
			equalTokenWord(tokens[offset+3], verb) &&
			tokensMatchSignedAmount(tokens[offset+4], tokens[offset+5], effect.PowerDelta) &&
			tokens[offset+6].Kind == shared.Slash &&
			tokensMatchSignedAmount(tokens[offset+7], tokens[offset+8], effect.ToughnessDelta)
	case compiler.StaticSubjectControlledArtifacts, compiler.StaticSubjectControlledTokens:
		noun := "artifacts"
		if effect.StaticSubject == compiler.StaticSubjectControlledTokens {
			noun = "tokens"
		}
		return 9, len(tokens) >= 9 &&
			equalTokenWord(tokens[0], noun) &&
			equalTokenWord(tokens[1], "you") &&
			equalTokenWord(tokens[2], "control") &&
			equalTokenWord(tokens[3], "get") &&
			tokensMatchSignedAmount(tokens[4], tokens[5], effect.PowerDelta) &&
			tokens[6].Kind == shared.Slash &&
			tokensMatchSignedAmount(tokens[7], tokens[8], effect.ToughnessDelta)
	case compiler.StaticSubjectOpponentControlledCreatures:
		return 10, len(tokens) >= 10 &&
			equalTokenWord(tokens[0], "creatures") &&
			equalTokenWord(tokens[1], "your") &&
			equalTokenWord(tokens[2], "opponents") &&
			equalTokenWord(tokens[3], "control") &&
			equalTokenWord(tokens[4], "get") &&
			tokensMatchSignedAmount(tokens[5], tokens[6], effect.PowerDelta) &&
			tokens[7].Kind == shared.Slash &&
			tokensMatchSignedAmount(tokens[8], tokens[9], effect.ToughnessDelta)
	default:
		return 0, false
	}
}

func syntaxSemanticTokens(syntax parser.Ability) []shared.Token {
	tokens := make([]shared.Token, 0, len(syntax.Tokens))
	for _, token := range syntax.Tokens {
		if spanCoveredByDelimited(token.Span, syntax.Reminders) ||
			spanCoveredByDelimited(token.Span, syntax.Quoted) {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func tokensMatchSignedAmount(sign, amount shared.Token, want compiler.CompiledSignedAmount) bool {
	expectedSign := shared.Plus
	if want.Negative {
		expectedSign = shared.Minus
	}
	return sign.Kind == expectedSign &&
		amount.Kind == shared.Integer &&
		amount.Text == strconv.Itoa(want.Value)
}

// lowerReminderManaAbility preserves a parenthesized reminder mana ability such
// as "({T}: Add {R} or {G}.)" and consumes other rules-free reminder abilities.
func lowerReminderManaAbility(
	ability compiler.CompiledAbility,
) (abilityLowering, *shared.Diagnostic) {
	unsupported := func() *shared.Diagnostic {
		return executableDiagnostic(
			ability,
			"unsupported reminder ability",
			"the executable source backend does not yet lower reminder abilities",
		)
	}
	if len(ability.Text) < 2 ||
		ability.Text[0] != '(' ||
		ability.Text[len(ability.Text)-1] != ')' {
		return abilityLowering{}, unsupported()
	}
	inner := strings.TrimSpace(ability.Text[1 : len(ability.Text)-1])
	innerDocument, innerDiags := parser.Parse(inner, parser.Context{})
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	innerDiags = append(innerDiags, compilerDiags...)
	if len(innerComp.Abilities) == 1 && isSemanticManaAbility(innerComp.Abilities[0]) {
		if len(innerDiags) != 0 ||
			len(innerComp.Syntax.Abilities) != 1 ||
			innerComp.Abilities[0].Kind != compiler.AbilityActivated {
			return abilityLowering{}, unsupported()
		}
		manaAbility, diagnostic := lowerManaAbility(
			"",
			innerComp.Abilities[0],
			innerComp.Syntax.Abilities[0],
		)
		if diagnostic != nil {
			return abilityLowering{}, unsupported()
		}
		// The compiled reminder ability has no independent semantic elements;
		// all content is filtered as parenthesized. The consumed counts are all
		// zero, matching the empty CompiledAbility fields.
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed:    semanticConsumption{},
			sourceSpans: []shared.Span{ability.Span},
		}, nil
	}

	// Non-mana reminder abilities carry no semantic content beyond their
	// parenthesized explanation.
	return abilityLowering{
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func lowerEntersTappedReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, *shared.Diagnostic) {
	if replacement, ok := lowerOptionalEntryPayment(ability); ok {
		return replacement, nil
	}
	if !entersTappedReplacementEffectsSupported(ability) ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterTapped ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != compiler.ReferenceBindingSource {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	if len(ability.Content.Conditions) == 1 {
		return lowerConditionalEntersTappedReplacement(ability)
	}
	if len(ability.Content.Conditions) != 0 {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only zero or one condition for self enters-tapped replacements",
		)
	}
	switch ability.Text {
	case "This land enters tapped.",
		"This artifact enters tapped.",
		"This creature enters tapped.",
		"This permanent enters tapped.":
	default:
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	return game.EntersTappedReplacement(ability.Text), nil
}

func lowerSelfZoneDestinationReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	event, eventOK := selfZoneDestinationReplacedEvent(ability)
	if !eventOK {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported self zone-destination replacement",
			detail,
		)
	}
	if len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!selfZoneDestinationReferencesSupported(ability) {
		return unsupported("the executable source backend supports only exact self graveyard-destination replacements")
	}
	destination, ok := selfZoneReplacementDestination(ability.Text)
	if !ok {
		return unsupported("the executable source backend supports only exile or shuffle-into-library self zone-destination replacements")
	}
	return game.ReplacementAbility{
		Text: ability.Text,
		Replacement: game.ReplacementEffect{
			MatchEvent:         game.EventZoneChanged,
			MatchFromZone:      event.matchFromZone,
			FromZone:           event.fromZone,
			MatchToZone:        true,
			ToZone:             zone.Graveyard,
			ReplaceToZone:      destination,
			ShuffleIntoLibrary: destination == zone.Library,
			RevealSource:       destination == zone.Library,
			Duration:           game.DurationPermanent,
		},
	}, true, nil
}

func lowerCounterPlacementReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !counterPlacementReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported counter-placement replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf ||
		len(ability.Content.Effects) != 2 ||
		ability.Content.Effects[0].Kind != compiler.EffectPut ||
		ability.Content.Effects[1].Kind != compiler.EffectPut ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact counter-doubling replacements")
	}
	switch ability.Content.Conditions[0].Predicate {
	case compiler.ConditionPredicateCounterPlacementOnControlledCreature:
		if !plusOneCounterDoublingEffects(ability.Content.Effects) {
			return unsupported("the executable source backend supports only +1/+1 counter-doubling replacement amounts")
		}
		return game.CounterPlacementReplacement(ability.Text, 2, counter.PlusOnePlusOne, game.TriggerControllerYou), true, nil
	case compiler.ConditionPredicateControllerCounterPlacement:
		if !anyCounterDoublingEffects(ability.Content.Effects) {
			return unsupported("the executable source backend supports only all-counter-doubling replacement amounts")
		}
		return game.AnyCounterPlacementReplacement(ability.Text, 2, game.TriggerControllerYou), true, nil
	default:
		return unsupported("the executable source backend supports only controlled-creature +1/+1 or broad permanent/player counter-doubling replacements")
	}
}

func lowerDamageReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !damageReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported damage replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact additive or multiplicative damage replacements")
	}
	raw := damageReplacementRawEffects(ability.Content.Effects)
	condition := ability.Content.Conditions[0]
	if condition.Predicate != compiler.ConditionPredicateDamageByControlledSource {
		return unsupported("the executable source backend supports only controlled-source red +1 damage or controlled-source double-damage replacements")
	}
	if len(condition.Selection.ColorsAny) == 1 &&
		condition.Selection.ColorsAny[0] == compiler.ConditionColorRed {
		if !strings.Contains(raw, "that much damage plus 1 to that permanent or player instead.") {
			return unsupported("the executable source backend supports only +1 red-source damage replacements")
		}
		if condition.Selection.ExcludeSource {
			return game.DamageReplacementExcludingSource(ability.Text, 0, 1, []color.Color{color.Red}, game.TriggerControllerYou), true, nil
		}
		return game.DamageReplacement(ability.Text, 0, 1, []color.Color{color.Red}, game.TriggerControllerYou), true, nil
	}
	if len(condition.Selection.ColorsAny) == 0 && !condition.Selection.ExcludeSource {
		if !strings.Contains(raw, "double that damage to that permanent or player instead.") &&
			!strings.Contains(raw, "twice that damage to that permanent or player instead.") {
			return unsupported("the executable source backend supports only double-damage replacements")
		}
		return game.DamageReplacement(ability.Text, 2, 0, nil, game.TriggerControllerYou), true, nil
	}
	return unsupported("the executable source backend supports only controlled-source red +1 damage or controlled-source double-damage replacements")
}

func damageReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	return ability.Content.Conditions[0].Predicate == compiler.ConditionPredicateDamageByControlledSource
}

func damageReplacementRawEffects(effects []compiler.CompiledEffect) string {
	raw := make([]string, 0, len(effects))
	for i := range effects {
		raw = append(raw, effects[i].Selector.Raw)
	}
	return strings.Join(raw, " ")
}

func counterPlacementReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	condition := ability.Content.Conditions[0]
	return condition.Predicate == compiler.ConditionPredicateControllerCounterPlacement ||
		condition.Predicate == compiler.ConditionPredicateCounterPlacementOnControlledCreature &&
			condition.Counter == compiler.ConditionCounterPlusOnePlusOne
}

func plusOneCounterDoublingEffects(effects []compiler.CompiledEffect) bool {
	first, second := effects[0], effects[1]
	if first.PowerDelta.Value != 1 || !first.PowerDelta.Known ||
		first.ToughnessDelta.Value != 1 || !first.ToughnessDelta.Known {
		return false
	}
	raw := first.Selector.Raw + " " + second.Selector.Raw
	if !strings.Contains(raw, "twice that many +1/+1 counters") {
		return false
	}
	return strings.Contains(raw, "on that creature instead.") ||
		strings.Contains(raw, "on it instead.")
}

func anyCounterDoublingEffects(effects []compiler.CompiledEffect) bool {
	raw := effects[0].Selector.Raw + " " + effects[1].Selector.Raw
	return strings.Contains(raw, "twice that many of each of those kinds of counters") &&
		strings.Contains(raw, "on that permanent or player instead.")
}

func lowerTokenCreationReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !tokenCreationReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported token-creation replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateTokenCreationUnderController ||
		len(ability.Content.Effects) != 2 ||
		ability.Content.Effects[0].Kind != compiler.EffectCreate ||
		ability.Content.Effects[1].Kind != compiler.EffectCreate ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact token-doubling replacements under your control")
	}
	switch ability.Content.Effects[1].Selector.Raw {
	case "twice that many of those tokens instead.", "twice that many tokens instead.":
	default:
		return unsupported("the executable source backend supports only token-doubling replacement amounts")
	}
	return game.TokenCreationReplacement(ability.Text, 2, game.TriggerControllerYou), true, nil
}

func tokenCreationReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	return ability.Content.Conditions[0].Predicate == compiler.ConditionPredicateTokenCreationUnderController
}

type selfZoneDestinationEvent struct {
	fromZone      zone.Type
	matchFromZone bool
}

func selfZoneDestinationReplacedEvent(ability compiler.CompiledAbility) (selfZoneDestinationEvent, bool) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf {
		return selfZoneDestinationEvent{}, false
	}
	switch ability.Content.Conditions[0].Predicate {
	case compiler.ConditionPredicateSourceWouldGoToGraveyard:
		if !referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
			return selfZoneDestinationEvent{}, false
		}
		return selfZoneDestinationEvent{}, true
	case compiler.ConditionPredicateSourceWouldDie:
		if !referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
			return selfZoneDestinationEvent{}, false
		}
		return selfZoneDestinationEvent{fromZone: zone.Battlefield, matchFromZone: true}, true
	default:
		return selfZoneDestinationEvent{}, false
	}
}

func selfZoneDestinationReferencesSupported(ability compiler.CompiledAbility) bool {
	return referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0)
}

func selfZoneReplacementDestination(text string) (zone.Type, bool) {
	if strings.Contains(text, " exile it instead.") {
		return zone.Exile, true
	}
	if strings.Contains(text, " shuffle it into its owner's library instead.") {
		return zone.Library, true
	}
	return zone.None, false
}

func lowerEntersWithCountersReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !isEntersWithCountersReplacement(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported enters-with-counters replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 0 {
		return unsupported("the executable source backend does not yet support conditional enters-with-counters replacements")
	}
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!selfEntersWithCountersReferences(ability.Content.References) {
		return unsupported("the executable source backend supports only exact unconditional self enters-with-counters replacements")
	}
	effect := ability.Content.Effects[0]
	if effect.Duration != compiler.DurationNone || effect.Negated {
		return unsupported("the executable source backend supports only exact unconditional self enters-with-counters replacements")
	}
	if strings.Contains(effect.Selector.Raw, " X ") ||
		strings.Contains(effect.Selector.Raw, " for each ") ||
		!effect.Amount.Known ||
		effect.Amount.Value <= 0 {
		return unsupported("the executable source backend does not yet support dynamic enters-with-counters quantities")
	}
	if !effect.CounterKindKnown {
		return unsupported("the executable source backend does not support this enters-with-counters counter kind")
	}
	return game.EntersWithCountersReplacement(ability.Text, game.CounterPlacement{
		Kind:   effect.CounterKind,
		Amount: effect.Amount.Value,
	}), true, nil
}

func isEntersWithCountersReplacement(ability compiler.CompiledAbility) bool {
	if len(ability.Content.Effects) == 0 ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterTapped {
		return false
	}
	raw := ability.Content.Effects[0].Selector.Raw
	return strings.HasPrefix(raw, "with ") &&
		strings.Contains(raw, " counter") &&
		strings.HasSuffix(raw, " on it.")
}

func selfEntersWithCountersReferences(references []compiler.CompiledReference) bool {
	return len(references) == 2 &&
		referencesBindTo(references, compiler.ReferenceBindingSource, 0)
}

func lowerOptionalEntryPayment(ability compiler.CompiledAbility) (game.ReplacementAbility, bool) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return game.ReplacementAbility{}, false
	}
	const payLifeText = "As this land enters, you may pay 2 life. If you don't, it enters tapped."
	if ability.Text == payLifeText &&
		len(ability.Content.Effects) == 2 &&
		ability.Content.Effects[0].Kind == compiler.EffectEnterTapped &&
		ability.Content.Effects[0].Amount.Known &&
		ability.Content.Effects[0].Amount.Value == 2 &&
		!ability.Content.Effects[0].Selector.Tapped &&
		ability.Content.Effects[1].Kind == compiler.EffectEnterTapped &&
		ability.Content.Effects[1].Selector.Tapped &&
		len(ability.Content.References) == 2 &&
		referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
		return game.EntersTappedUnlessPaidReplacement(ability.Text, game.ResolutionPayment{
			Prompt: "Pay 2 life?",
			AdditionalCosts: []cost.Additional{{
				Kind:   cost.AdditionalPayLife,
				Amount: 2,
			}},
		}), true
	}
	if !revealEntrySyntax(ability.Text) ||
		len(ability.Content.Effects) != 3 ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterTapped ||
		ability.Content.Effects[0].Selector.Tapped ||
		ability.Content.Effects[1].Kind != compiler.EffectReveal ||
		ability.Content.Effects[1].Amount.Value != 1 ||
		!ability.Content.Effects[1].Amount.Known ||
		len(ability.Content.Effects[1].Selector.SubtypesAny()) == 0 ||
		len(ability.Content.Effects[1].Selector.SubtypesAny()) > 2 ||
		ability.Content.Effects[2].Kind != compiler.EffectEnterTapped ||
		!ability.Content.Effects[2].Selector.Tapped ||
		len(ability.Content.References) != 2 ||
		!referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
		return game.ReplacementAbility{}, false
	}
	var subtypeSet cost.SubtypeSet
	copy(subtypeSet[:], ability.Content.Effects[1].Selector.SubtypesAny())
	return game.EntersTappedUnlessPaidReplacement(ability.Text, game.ResolutionPayment{
		Prompt: "Reveal a matching card?",
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalReveal,
			Amount:      1,
			SubtypesAny: subtypeSet,
			Source:      zone.Hand,
		}},
	}), true
}

func revealEntrySyntax(text string) bool {
	const prefix = "As this land enters, you may reveal "
	const suffix = " card from your hand. If you don't, this land enters tapped."
	if !strings.HasPrefix(text, prefix) || !strings.HasSuffix(text, suffix) {
		return false
	}
	return len(strings.Split(strings.TrimSuffix(strings.TrimPrefix(text, prefix), suffix), " or ")) <= 2
}

func entersTappedReplacementEffectsSupported(ability compiler.CompiledAbility) bool {
	if len(ability.Content.Effects) == 0 {
		return false
	}
	if len(ability.Content.Effects) == 1 {
		return true
	}
	if len(ability.Content.Conditions) != 1 {
		return false
	}
	conditionSpans := []shared.Span{ability.Content.Conditions[0].Span}
	for _, effect := range ability.Content.Effects[1:] {
		if !spanCovered(effect.VerbSpan, conditionSpans) {
			return false
		}
	}
	return true
}

func lowerConditionalEntersTappedReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, *shared.Diagnostic) {
	condition := ability.Content.Conditions[0]
	replacementCondition, ok := lowerCondition(condition, conditionContextReplacement)
	if !ok {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported conditional enters-tapped replacement",
			"the executable source backend does not support this enters-tapped condition",
		)
	}
	return game.EntersTappedIfReplacement(ability.Text, &replacementCondition), nil
}

func lowerAtTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported phase/step trigger phrase"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the executable source backend requires a semantic step trigger pattern",
		)
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok || pattern.Event != game.EventBeginningOfStep {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			triggerPatternCapabilityDiagnostic(ability.Trigger),
		)
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the executable source backend does not support this intervening-if condition",
		)
	}
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"modes and ability words are not supported in phase/step triggers",
		)
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"the executable source backend does not support this phase/step trigger body",
		)
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 game.TriggerAt,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func lowerAtInterveningCondition(trigger *compiler.CompiledTrigger) (opt.V[game.Condition], bool) {
	if trigger == nil || trigger.Condition == nil {
		return opt.V[game.Condition]{}, true
	}
	condition := *trigger.Condition
	if lowered, ok := lowerCondition(condition, conditionContextInterveningTrigger); ok {
		return opt.Val(lowered), true
	}
	return opt.V[game.Condition]{}, false
}

func lowerTriggeredAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern",
		)
	}
	pattern := ability.Trigger.Pattern
	if pattern.Kind == compiler.TriggerAt {
		return lowerAtTrigger(cardName, ability, syntax)
	}
	switch pattern.Event {
	case compiler.TriggerEventCardDrawn, compiler.TriggerEventCardDiscarded, compiler.TriggerEventCycled:
		return lowerDrawDiscardTrigger(cardName, ability, syntax)
	case compiler.TriggerEventLifeGained, compiler.TriggerEventLifeLost, compiler.TriggerEventDamageDealt:
		return lowerLifeDamageTrigger(cardName, ability, syntax)
	case compiler.TriggerEventPermanentEnteredBattlefield,
		compiler.TriggerEventPermanentDied,
		compiler.TriggerEventZoneChanged:
		return lowerPermanentZoneChangeTrigger(cardName, ability, syntax)
	case compiler.TriggerEventSpellCast:
		return lowerCastTrigger(cardName, ability, syntax)
	default:
		if unknownLifeDamageTrigger(&pattern, ability.Trigger.Event) {
			return lowerLifeDamageTrigger(cardName, ability, syntax)
		}
		if pattern.Source == compiler.TriggerSourceSelf {
			return lowerEnterTrigger(cardName, ability, syntax)
		}
		return lowerGenericPatternTrigger(cardName, ability, syntax)
	}
}

func unknownLifeDamageTrigger(pattern *compiler.TriggerPattern, event string) bool {
	if pattern.Event != compiler.TriggerEventUnknown {
		return false
	}
	event = strings.ToLower(event)
	return strings.HasPrefix(event, "a spell ") && strings.Contains(event, "damage") ||
		strings.Contains(event, "combat damage") && strings.Contains(event, " or dies")
}

func lowerDrawDiscardTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported draw/discard trigger"
	const effectSummary = "unsupported draw/discard trigger effect"
	if ability.Trigger == nil || ability.Trigger.Pattern.Kind != compiler.TriggerWhenever {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend supports only TriggerWhenever draw and discard triggers")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok ||
		(pattern.Event != game.EventCardDrawn &&
			pattern.Event != game.EventCardDiscarded &&
			pattern.Event != game.EventCycled) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"unrecognized semantic draw, discard, or cycling trigger pattern")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic draw/discard trigger condition")
	}
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this draw/discard trigger body")
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this draw/discard trigger body")
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 game.TriggerWhenever,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func lowerGenericPatternTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok {
		if ability.Trigger.Pattern.OneOrMore {
			if diagnostic := triggerBodyDiagnostic(cardName, ability, syntax); diagnostic != nil {
				return game.TriggeredAbility{}, diagnostic
			}
		}
		if strings.EqualFold(ability.Trigger.Event, "one or more outlaws you control deal combat damage to a player") {
			if diagnostic := triggerBodyDiagnostic(cardName, ability, syntax); diagnostic != nil {
				return game.TriggeredAbility{}, diagnostic
			}
		}
		switch ability.Text {
		case "Whenever an equipped creature you control attacks, exile the top card of your library. You may play that card this turn. You may cast Equipment spells this way without paying their mana costs.",
			"Whenever an equipped creature you control attacks, you may tap target creature defending player controls.",
			"Whenever an equipped creature you control attacks, you draw a card and you lose 1 life.":
			return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
				"the semantic Trigger Pattern contains a field with no runtime lowering adapter")
		}
		detail := triggerPatternCapabilityDiagnostic(ability.Trigger)
		if detail == "the executable source backend does not support this semantic permanent zone-change trigger pattern" {
			return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported permanent zone-change trigger", detail)
		}
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			detail)
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || triggerType == game.TriggerAt {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger kind")
	}

	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger condition")
	}
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this trigger body")
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this trigger body")
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 triggerType,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func triggerBodyDiagnostic(cardName string, ability compiler.CompiledAbility, syntax parser.Ability) *shared.Diagnostic {
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return nil
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return nil
	}
	_, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	return diagnostic
}

func triggerPatternCapabilityDiagnostic(trigger *compiler.CompiledTrigger) string {
	if trigger == nil {
		return "the trigger shell is missing a semantic Trigger Pattern"
	}
	if trigger.Pattern.Event == compiler.TriggerEventAbilityActivated && !trigger.Pattern.ExcludeManaAbility {
		return "the runtime ability-activated event stream omits payment-time mana abilities, so unrestricted activated-ability triggers require a missing runtime capability"
	}
	if trigger.Pattern.Event != compiler.TriggerEventUnknown {
		return "the semantic Trigger Pattern contains a field with no runtime lowering adapter"
	}
	event := strings.ToLower(trigger.Event)
	for _, boundary := range []string{
		"declare attackers step",
		"declare blockers step",
		"first strike damage step",
		"combat damage step",
		"cleanup step",
	} {
		if strings.Contains(event, boundary) {
			return fmt.Sprintf("the runtime does not emit a beginning-of-%s event", boundary)
		}
	}
	if strings.Contains(event, " dies") && strings.Contains(event, "blocking this") {
		return "the executable source backend does not support this semantic permanent zone-change trigger pattern"
	}
	switch event {
	case "an enchanted creature dies",
		"an equipped creature you control dies":
		return "the executable source backend does not support this semantic permanent zone-change trigger pattern"
	case "a renowned creature you control deals combat damage to a player",
		"an enchanted creature you control deals combat damage to a player",
		"a goaded creature deals combat damage to one of your opponents",
		"a noncreature source you control deals damage":
		return "the executable source backend does not support this semantic life or damage trigger pattern"
	case "an enchanted creature attacks one of your opponents",
		"a goaded creature attacks",
		"one or more suspected creatures you control attack":
		return "the semantic Trigger Pattern contains a field with no runtime lowering adapter"
	}
	if strings.Contains(event, "attack") ||
		strings.Contains(event, "block") ||
		strings.Contains(event, "damage") ||
		strings.Contains(event, "combat") ||
		strings.Contains(event, "upkeep") ||
		strings.Contains(event, "draw step") ||
		strings.Contains(event, "end step") ||
		strings.Contains(event, "main phase") {
		return "the runtime event exists, but this combat, phase, or step relation requires a missing runtime capability"
	}
	if strings.Contains(event, " or ") {
		return "the runtime events exist, but this trigger requires a missing event-or-subject-union semantic slot"
	}
	if strings.Contains(event, "first time") ||
		strings.Contains(event, "second time") ||
		strings.Contains(event, "third time") ||
		strings.Contains(event, "during your turn") ||
		strings.Contains(event, "during their turn") ||
		strings.Contains(event, "once each turn") {
		return "the runtime event exists, but this trigger requires a missing ordinal, active-turn, or temporal semantic slot"
	}
	if strings.Contains(event, "target") {
		return "the object-became-target event exists, but this trigger requires a missing target-subject, targeting-cause, or source relation slot"
	}
	if unrestrictedAbilityActivatedEvent(event) {
		if trigger.Condition != nil && strings.Contains(strings.ToLower(trigger.Condition.Text), "mana ability") {
			return "the ability-activated event exists, but non-mana exclusion in an intervening condition requires a missing semantic condition slot"
		}
		if !strings.Contains(event, "isn't a mana ability") {
			return "the runtime ability-activated event stream omits payment-time mana abilities, so unrestricted activated-ability triggers require a missing runtime capability"
		}
		return "the ability-activated event exists, but this trigger requires a missing source, activation-cost, or ability-provenance semantic slot"
	}
	if strings.Contains(event, "ability") {
		return "the ability-activated event exists, but this trigger requires a missing source, activation-cost, or ability-provenance semantic slot"
	}
	if strings.Contains(event, "cast") || strings.Contains(event, "spell") || strings.Contains(event, "copied") {
		return "the spell event exists, but this trigger requires a missing spell-event relation, copy, or provenance semantic slot"
	}
	if strings.Contains(event, "sacrific") {
		return "the permanent-sacrificed event exists, but this trigger requires a missing subject, actor, or sacrifice-provenance semantic slot"
	}
	if strings.Contains(event, "scry") || strings.Contains(event, "surveil") {
		return "the player-action event exists, but this trigger requires a missing action amount, provenance, or temporal semantic slot"
	}
	if strings.Contains(event, "tap") || strings.Contains(event, "untap") {
		if strings.Contains(event, "for mana") {
			return "the permanent-tapped event exists, but the runtime event lacks tapped-for-mana provenance"
		}
		return "the permanent-state event exists, but this trigger requires a missing subject, source, or turn-provenance semantic slot"
	}
	if strings.Contains(event, "counter") {
		return "the counter event exists, but this trigger requires a missing counter-kind, subject, controller, or removal semantic slot"
	}
	if strings.Contains(event, "draw") || strings.Contains(event, "discard") || strings.Contains(event, "cycl") {
		return "the player-card event exists, but this trigger requires a missing count, card-selection, source, or turn-provenance semantic slot"
	}
	if strings.Contains(event, "turned face up") {
		return "the permanent-turned-face-up event exists, but this trigger requires a missing subject, source, or Selection semantic slot"
	}
	if strings.Contains(event, "turned face down") {
		return "the runtime does not emit an authoritative permanent-turned-face-down event"
	}
	if strings.Contains(event, " enters") ||
		strings.Contains(event, " dies") ||
		strings.Contains(event, " leaves") ||
		strings.Contains(event, "graveyard") ||
		strings.Contains(event, "exiled") {
		return "the zone-change event exists, but this trigger requires a missing subject, zone, source, or Selection semantic slot"
	}
	if strings.Contains(event, "token") {
		return "the token-created event exists, but this trigger requires a missing creator, subject, or Selection semantic slot"
	}
	if strings.Contains(event, "transform") ||
		strings.Contains(event, "investigate") ||
		strings.Contains(event, "proliferate") ||
		strings.Contains(event, "explore") ||
		strings.Contains(event, "monstrous") ||
		strings.Contains(event, "venture") ||
		strings.Contains(event, "roll") ||
		strings.Contains(event, "vote") ||
		strings.Contains(event, "clash") {
		return "the runtime does not emit an authoritative event for this game action"
	}
	return "the runtime does not emit an authoritative event for this trigger action"
}

func unrestrictedAbilityActivatedEvent(event string) bool {
	for _, prefix := range []string{
		"you activate ",
		"an opponent activates ",
		"a player activates ",
	} {
		ability, ok := strings.CutPrefix(event, prefix)
		if !ok {
			continue
		}
		return ability == "an ability" || strings.HasPrefix(ability, "an ability of ")
	}
	return false
}

func lowerTriggeredAbilityKind(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	triggeredAbility, diagnostic := lowerTriggeredAbility(cardName, ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := []shared.Span{ability.Trigger.Span}
	if syntax.AbilityWord != nil {
		spans = append(spans, shared.Span{
			Start: ability.Span.Start,
			End:   ability.Trigger.Span.Start,
		})
	}
	for _, effect := range ability.Content.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	for _, condition := range ability.Content.Conditions {
		spans = append(spans, condition.Span)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		triggeredAbility: opt.Val(triggeredAbility),
		consumed: semanticConsumption{
			trigger:    true,
			optional:   ability.Optional,
			targets:    len(ability.Content.Targets),
			conditions: len(ability.Content.Conditions),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

func (lowering *abilityLowering) complete(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) bool {
	staticDeclarations := 0
	if ability.Static != nil {
		staticDeclarations = len(ability.Static.Declarations)
	}
	if lowering.consumed.cost != (ability.Cost != nil) ||
		lowering.consumed.trigger != (ability.Trigger != nil) ||
		lowering.consumed.optional != ability.Optional ||
		lowering.consumed.modes != len(ability.Content.Modes) ||
		lowering.consumed.targets != len(ability.Content.Targets) ||
		lowering.consumed.conditions != len(ability.Content.Conditions) ||
		lowering.consumed.effects != len(ability.Content.Effects) ||
		lowering.consumed.keywords != len(ability.Content.Keywords) ||
		lowering.consumed.references != len(ability.Content.References) ||
		lowering.consumed.declarations != staticDeclarations {
		return false
	}
	for _, token := range syntax.Tokens {
		if token.Kind == shared.Comma ||
			token.Kind == shared.Colon ||
			token.Kind == shared.Period ||
			(syntax.AbilityWord != nil && rulesFreeAbilityWordLabel(ability.AbilityWord) &&
				(token.Kind == shared.EmDash || spanCoveredByAbilityWord(token.Span, syntax.AbilityWord))) ||
			spanCovered(token.Span, lowering.sourceSpans) {
			continue
		}
		return false
	}
	return true
}

func lowerEnterTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported triggered ability",
			"the executable source backend requires a semantic self trigger pattern",
		)
	}
	pattern, supportedEvent := lowerTriggerPattern(&ability.Trigger.Pattern)
	eventKind := pattern.Event
	summary := "unsupported triggered ability"
	effectSummary := "unsupported triggered ability effect"
	detail := "the executable source backend supports only recognized semantic self triggers with supported effects"
	switch ability.Trigger.Pattern.Event {
	case compiler.TriggerEventPermanentEnteredBattlefield:
		summary = "unsupported enter trigger"
		effectSummary = "unsupported enter trigger effect"
		detail = "the executable source backend supports only recognized semantic self-enter triggers with supported effects"
	case compiler.TriggerEventPermanentDied:
		summary = "unsupported dies trigger"
		effectSummary = "unsupported dies trigger effect"
		detail = "the executable source backend supports only recognized semantic self-dies triggers with supported effects"
	default:
	}
	intervening, supportedCondition := lowerSelfInterveningCondition(eventKind, ability.Trigger)
	if !supportedSelfTriggerKind(eventKind, ability.Trigger.Pattern.Kind) ||
		!supportedEvent ||
		!supportedCondition {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary, detail)
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary, detail)
	}
	selfDamage := eventKind == game.EventPermanentDied &&
		normalizeSelfDamageReference(cardName, &body)
	if selfDamage {
		bodySyntax.Text = body.Text
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                                   triggerType,
			Pattern:                                pattern,
			InterveningIf:                          interveningIfText(ability.Trigger),
			InterveningCondition:                   intervening.condition,
			InterveningIfEventPermanentHadCounters: intervening.hadCounters,
			InterveningIfEventPermanentHadNoCounterKind: intervening.hadNoCounterKind,
			InterveningIfEventPermanentWasKicked:        intervening.wasKicked,
			InterveningIfEventPermanentWasCast:          intervening.wasCast,
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func lowerLifeDamageTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic life or damage trigger")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok ||
		(pattern.Event != game.EventLifeGained &&
			pattern.Event != game.EventLifeLost &&
			pattern.Event != game.EventDamageDealt) {
		if ability.Trigger.Pattern.OneOrMore {
			if diagnostic := triggerBodyDiagnostic(cardName, ability, syntax); diagnostic != nil {
				return game.TriggeredAbility{}, diagnostic
			}
		}
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic life or damage trigger pattern")
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || (triggerType != game.TriggerWhen && triggerType != game.TriggerWhenever) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"life and damage triggers require When or Whenever")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic life or damage trigger condition")
	}
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this life or damage trigger body")
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this life or damage trigger body")
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 triggerType,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func lowerEventCardEffect(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingEventCard, 0) {
		return game.AbilityContent{}, false
	}
	eventCard, ok := lowerCardReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectReturn:
		if ctx.text != "Return it to its owner's hand." &&
			ctx.text != "Return that card to its owner's hand." {
			return game.AbilityContent{}, false
		}
	case compiler.EffectExile:
		if ctx.text != "Exile it." && ctx.text != "Exile that card." {
			return game.AbilityContent{}, false
		}
	case compiler.EffectCast:
		if ctx.text != "Cast it from your graveyard as an Adventure until the end of your next turn." ||
			len(ctx.content.References) != 1 {
			return game.AbilityContent{}, false
		}
	default:
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectReturn:
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				Card:        eventCard,
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			},
		}}}.Ability(), true
	case compiler.EffectCast:
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GrantCastPermission{
				Card:     eventCard,
				FromZone: zone.Graveyard,
				Face:     game.FaceAlternate,
				Duration: game.DurationUntilEndOfYourNextTurn,
			},
		}}}.Ability(), true
	case compiler.EffectExile:
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				Card:        eventCard,
				FromZone:    zone.Graveyard,
				Destination: zone.Exile,
			},
		}}}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

type enterInterveningCondition struct {
	condition        opt.V[game.Condition]
	hadCounters      bool
	hadNoCounterKind opt.V[counter.Kind]
	wasKicked        bool
	wasCast          bool
}

func lowerSelfInterveningCondition(
	eventKind game.EventKind,
	trigger *compiler.CompiledTrigger,
) (enterInterveningCondition, bool) {
	if trigger != nil && trigger.Condition != nil {
		if condition, ok := lowerCondition(*trigger.Condition, conditionContextInterveningTrigger); ok {
			return enterInterveningCondition{condition: opt.Val(condition)}, true
		}
		if trigger.Condition.Predicate == compiler.ConditionPredicateEventSubjectHadCounters {
			if trigger.Condition.ObjectBinding != compiler.ReferenceBindingEventPermanent {
				return enterInterveningCondition{}, false
			}
			return enterInterveningCondition{hadCounters: true}, true
		}
	}
	switch eventKind {
	case game.EventPermanentEnteredBattlefield:
		return lowerEnterInterveningCondition(trigger)
	case game.EventPermanentDied:
		return lowerDiesInterveningCondition(trigger)
	default:
		return enterInterveningCondition{}, trigger == nil || trigger.Condition == nil
	}
}

func supportedSelfTriggerKind(eventKind game.EventKind, kind compiler.TriggerKind) bool {
	switch eventKind {
	case game.EventPermanentEnteredBattlefield,
		game.EventPermanentDied,
		game.EventZoneChanged,
		game.EventPermanentTurnedFaceUp,
		game.EventPermanentSacrificed,
		game.EventObjectBecameTarget:
		return kind == compiler.TriggerWhen || kind == compiler.TriggerWhenever
	case game.EventPermanentMutated,
		game.EventAttackerBecameBlocked,
		game.EventAttackerDeclared,
		game.EventBlockerDeclared,
		game.EventDamageDealt,
		game.EventPermanentTapped,
		game.EventPermanentUntapped,
		game.EventCountersAdded:
		return kind == compiler.TriggerWhenever
	default:
		return kind == compiler.TriggerWhen
	}
}

func lowerEnterInterveningCondition(trigger *compiler.CompiledTrigger) (enterInterveningCondition, bool) {
	if trigger == nil || trigger.Condition == nil {
		return enterInterveningCondition{}, true
	}
	condition := trigger.Condition
	if condition.Kind != compiler.ConditionIf || !condition.Intervening {
		return enterInterveningCondition{}, false
	}
	switch condition.Predicate {
	case compiler.ConditionPredicateEventSubjectWasKicked:
		return enterInterveningCondition{wasKicked: true}, true
	case compiler.ConditionPredicateEventSubjectWasCast:
		return enterInterveningCondition{wasCast: true}, true
	case compiler.ConditionPredicateEventSubjectWasCastByController:
		return enterInterveningCondition{}, false
	default:
	}
	lowered, ok := lowerCondition(*condition, conditionContextInterveningTrigger)
	if !ok {
		return enterInterveningCondition{}, false
	}
	return enterInterveningCondition{
		condition: opt.Val(lowered),
	}, true
}

func lowerDiesInterveningCondition(trigger *compiler.CompiledTrigger) (enterInterveningCondition, bool) {
	if trigger == nil || trigger.Condition == nil {
		return enterInterveningCondition{}, true
	}
	condition := trigger.Condition
	if condition.Kind != compiler.ConditionIf || !condition.Intervening {
		return enterInterveningCondition{}, false
	}
	if condition.Predicate != compiler.ConditionPredicateEventSubjectHadNoCounter {
		return enterInterveningCondition{}, false
	}
	switch condition.Counter {
	case compiler.ConditionCounterPlusOnePlusOne:
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.PlusOnePlusOne)}, true
	case compiler.ConditionCounterMinusOneMinusOne:
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.MinusOneMinusOne)}, true
	default:
		return enterInterveningCondition{}, false
	}
}

func normalizeSelfDamageReference(cardName string, ability *compiler.CompiledAbility) bool {
	if ability == nil ||
		len(ability.Content.Effects) != 1 ||
		(len(ability.Content.References) != 1 && len(ability.Content.References) != 2) ||
		ability.Content.References[0].Binding != compiler.ReferenceBindingEventPermanent ||
		!strings.HasPrefix(ability.Text, "It deals ") ||
		!strings.HasPrefix(strings.ToLower(ability.Content.Effects[0].Text), "it deals ") {
		return false
	}
	if len(ability.Content.References) == 2 &&
		(ability.Content.Effects[0].Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
			ability.Content.References[1].Binding != compiler.ReferenceBindingEventPermanent ||
			ability.Content.References[1].Span != ability.Content.Effects[0].Amount.ReferenceSpan) {
		return false
	}
	ability.Text = cardName + ability.Text[len("It"):]
	ability.Content.Effects[0].Text = cardName + ability.Content.Effects[0].Text[len("It"):]
	return true
}

func bodyReferences(
	references []compiler.CompiledReference,
	excludedSpans ...shared.Span,
) []compiler.CompiledReference {
	var body []compiler.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, excludedSpans) {
			continue
		}
		body = append(body, reference)
	}
	return body
}

func interveningIfText(trigger *compiler.CompiledTrigger) string {
	if trigger == nil || trigger.Condition == nil {
		return ""
	}
	return trigger.Condition.Text
}

// prepareTriggerBody builds the body CompiledAbility and syntax for a
// supported triggered ability. It handles condition consistency, effect
// filtering for intervening conditions, body span/text construction, reference
// exclusion, and optional "you may" stripping. Callers must have already
// verified that ability.Trigger is non-nil.
func prepareTriggerBody(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (compiler.CompiledAbility, parser.Ability, bool) {
	if ability.Trigger == nil {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	hasInterveningCondition := ability.Trigger.Condition != nil
	if (len(ability.Content.Conditions) != 0 && !hasInterveningCondition) ||
		(hasInterveningCondition && (len(ability.Content.Conditions) != 1 ||
			ability.Content.Conditions[0].Span != ability.Trigger.Condition.Span ||
			ability.Optional)) {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	resolvingEffects := ability.Content.Effects
	if hasInterveningCondition {
		conditionSpan := []shared.Span{ability.Trigger.Condition.Span}
		resolvingEffects = slices.DeleteFunc(
			append([]compiler.CompiledEffect(nil), ability.Content.Effects...),
			func(effect compiler.CompiledEffect) bool {
				return spanCovered(effect.VerbSpan, conditionSpan)
			},
		)
	}
	if len(resolvingEffects) == 0 {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	body := ability
	body.Content.Effects = resolvingEffects
	body.Kind = compiler.AbilitySpell
	body.Span = shared.Span{
		Start: resolvingEffects[0].Span.Start,
		End:   resolvingEffects[len(resolvingEffects)-1].Span.End,
	}
	body.Text = titleFirst(
		ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
	)
	body.Trigger = nil
	body.Optional = false
	body.OptionalSpan = shared.Span{}
	excludedReferenceSpans := []shared.Span{ability.Trigger.Span}
	if hasInterveningCondition {
		excludedReferenceSpans = append(excludedReferenceSpans, ability.Trigger.Condition.Span)
		body.Content.Conditions = nil
		bodyStart := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
			return token.Kind != shared.Comma &&
				token.Span.Start.Offset >= ability.Trigger.Condition.Span.End.Offset
		})
		if bodyStart < 0 {
			return compiler.CompiledAbility{}, parser.Ability{}, false
		}
		effect := body.Content.Effects[0]
		effect.Span.Start = syntax.Tokens[bodyStart].Span.Start
		effect.Text = ability.Text[effect.Span.Start.Offset-ability.Span.Start.Offset : effect.Span.End.Offset-ability.Span.Start.Offset]
		body.Content.Effects[0] = effect
		body.Span.Start = effect.Span.Start
		body.Text = titleFirst(
			ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
		)
	}
	body.Content.References = bodyReferences(ability.Content.References, excludedReferenceSpans...)
	bodyTokenStart := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
		return token.Span.Start.Offset >= body.Span.Start.Offset
	})
	if bodyTokenStart < 0 {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	bodySyntax := syntax
	bodySyntax.Kind = parser.AbilitySpell
	bodySyntax.Tokens = syntax.Tokens[bodyTokenStart:]
	if ability.Optional {
		if len(ability.Content.Effects) != 1 ||
			len(bodySyntax.Tokens) < 3 ||
			!equalTokenWord(bodySyntax.Tokens[0], "you") ||
			!equalTokenWord(bodySyntax.Tokens[1], "may") ||
			ability.OptionalSpan.Start != ability.Content.Effects[0].Span.Start {
			return compiler.CompiledAbility{}, parser.Ability{}, false
		}
		effect := body.Content.Effects[0]
		effect.Text = effect.Text[effect.VerbSpan.Start.Offset-effect.Span.Start.Offset:]
		effect.Span.Start = effect.VerbSpan.Start
		body.Content.Effects = []compiler.CompiledEffect{effect}
		body.Span.Start = effect.Span.Start
		body.Text = titleFirst(
			ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
		)
		bodySyntax.Tokens = bodySyntax.Tokens[2:]
	}
	body.Content.Keywords = keywordsWithinSpan(ability.Content.Keywords, body.Span)
	if len(body.Content.Keywords) != len(ability.Content.Keywords) {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	return body, bodySyntax, true
}

func lowerPermanentZoneChangeTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported permanent zone-change trigger"
	const effectSummary = "unsupported permanent zone-change trigger effect"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend requires a semantic permanent zone-change trigger")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok ||
		(pattern.Event != game.EventPermanentEnteredBattlefield &&
			pattern.Event != game.EventPermanentDied &&
			pattern.Event != game.EventZoneChanged) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic permanent zone-change trigger pattern")
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || (triggerType != game.TriggerWhen && triggerType != game.TriggerWhenever) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"permanent zone-change triggers require When or Whenever")
	}
	intervening, ok := lowerPermanentZoneChangeInterveningCondition(&pattern, ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic permanent zone-change trigger condition")
	}
	if len(ability.Content.Effects) == 0 ||
		len(ability.Content.Modes) != 0 ||
		(pattern.Event != game.EventPermanentEnteredBattlefield &&
			!rulesFreeAbilityWordLabel(ability.AbilityWord)) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this permanent zone-change trigger body")
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this permanent zone-change trigger body")
	}
	if (pattern.Event == game.EventPermanentDied || pattern.Event == game.EventZoneChanged) &&
		normalizeSelfDamageReference(cardName, &body) {
		bodySyntax.Text = body.Text
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return permanentZoneChangeTriggeredAbility(ability, triggerType, &pattern, &intervening, content), nil
}

func lowerPermanentZoneChangeInterveningCondition(
	pattern *game.TriggerPattern,
	trigger *compiler.CompiledTrigger,
) (enterInterveningCondition, bool) {
	if pattern.Source == game.TriggerSourceSelf {
		return lowerSelfInterveningCondition(pattern.Event, trigger)
	}
	if trigger != nil && trigger.Condition != nil {
		switch trigger.Condition.Predicate {
		case compiler.ConditionPredicateObjectMatches, compiler.ConditionPredicateObjectExists:
			if condition, ok := lowerCondition(*trigger.Condition, conditionContextInterveningTrigger); ok {
				return enterInterveningCondition{condition: opt.Val(condition)}, true
			}
		default:
		}
		if trigger.Condition.Predicate == compiler.ConditionPredicateEventSubjectHadCounters {
			if trigger.Condition.ObjectBinding != compiler.ReferenceBindingEventPermanent {
				return enterInterveningCondition{}, false
			}
			return enterInterveningCondition{hadCounters: true}, true
		}
	}
	if pattern.Event == game.EventPermanentEnteredBattlefield {
		intervening, ok := lowerEnterInterveningCondition(trigger)
		if !ok ||
			(trigger.Condition != nil &&
				trigger.Condition.Predicate == compiler.ConditionPredicateEventSubjectWasCastByController) {
			return enterInterveningCondition{}, false
		}
		return intervening, true
	}
	return enterInterveningCondition{}, trigger == nil || trigger.Condition == nil
}

func permanentZoneChangeTriggeredAbility(
	ability compiler.CompiledAbility,
	triggerType game.TriggerType,
	pattern *game.TriggerPattern,
	intervening *enterInterveningCondition,
	content game.AbilityContent,
) game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                                   triggerType,
			Pattern:                                *pattern,
			InterveningIf:                          interveningIfText(ability.Trigger),
			InterveningCondition:                   intervening.condition,
			InterveningIfEventPermanentHadCounters: intervening.hadCounters,
			InterveningIfEventPermanentHadNoCounterKind: intervening.hadNoCounterKind,
			InterveningIfEventPermanentWasKicked:        intervening.wasKicked,
			InterveningIfEventPermanentWasCast:          intervening.wasCast,
		},
		Optional: ability.Optional,
		Content:  content,
	}
}

// lowerCastTrigger lowers a recognized semantic spell-cast trigger into a
// game.TriggeredAbility with EventSpellCast.
func lowerCastTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Kind != compiler.TriggerWhenever {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic whenever spell-cast trigger")
	}

	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok || pattern.Event != game.EventSpellCast {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic spell-cast trigger pattern")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic spell-cast trigger condition")
	}
	if len(ability.Content.Modes) != 0 || !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this spell-cast trigger body")
	}

	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this spell-cast trigger body")
	}
	content, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}

	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 game.TriggerWhenever,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func spanCovered(span shared.Span, covering []shared.Span) bool {
	for _, candidate := range covering {
		if candidate.Start.Offset <= span.Start.Offset &&
			candidate.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func lowerKeywordAbility(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) ([]loweredStaticAbility, *shared.Diagnostic) {
	for _, keyword := range ability.Content.Keywords {
		if keyword.Name == "Devoid" && ability.Text != "Devoid (This card has no color.)" {
			return nil, executableDiagnostic(
				ability,
				"unsupported Devoid ability",
				"the executable source backend supports only exact \"Devoid (This card has no color.)\" abilities",
			)
		}
		if keyword.Name == "Read ahead" && !isReadAheadAbility(ability.Text) {
			return nil, executableDiagnostic(
				ability,
				"unsupported Read ahead ability",
				"the executable source backend supports only the canonical Read ahead ability and reminder text",
			)
		}
	}
	if len(ability.Content.Modes) > 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not yet lower modal abilities",
		)
	}
	if !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return nil, executableDiagnostic(
			ability,
			"unsupported ability word",
			fmt.Sprintf("the executable source backend does not yet lower the %q ability word", ability.AbilityWord),
		)
	}
	if len(ability.Content.Keywords) == 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend does not yet lower non-keyword static rules text",
		)
	}
	bodies := make([]loweredStaticAbility, 0, len(ability.Content.Keywords))
	for _, keyword := range ability.Content.Keywords {
		if keyword.Parameter != "" {
			if body, ok, diag := lowerParameterizedKeywordToStaticAbility(ability, keyword); ok {
				if diag != nil {
					return nil, diag
				}
				bodies = append(bodies, loweredStaticAbility{Body: body})
				continue
			}
			return nil, executableDiagnostic(
				ability,
				"unsupported parameterized keyword",
				fmt.Sprintf(
					"the executable source backend does not yet lower %s with parameter %q",
					keyword.Name,
					keyword.Parameter,
				),
			)
		}
		body, ok := keywordStaticBodies[keyword.Name]
		if !ok {
			return nil, executableDiagnostic(
				ability,
				"unsupported keyword ability",
				fmt.Sprintf(
					"the executable source backend has no reusable game template for %s",
					keyword.Name,
				),
			)
		}
		bodies = append(bodies, body)
	}
	if len(ability.Content.Targets) > 0 ||
		len(ability.Content.Conditions) > 0 ||
		len(ability.Content.Effects) > 0 ||
		len(ability.Content.References) > 0 {
		return nil, mixedKeywordDiagnostic(contentCtx{span: ability.Span, content: ability.Content})
	}
	for _, token := range syntax.Tokens {
		if token.Kind == shared.Comma ||
			(syntax.AbilityWord != nil && token.Kind == shared.EmDash) ||
			spanCoveredByAbilityWord(token.Span, syntax.AbilityWord) ||
			spanCoveredByKeyword(token.Span, ability.Content.Keywords) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return nil, mixedKeywordDiagnostic(contentCtx{span: ability.Span, content: ability.Content})
	}
	return bodies, nil
}

func rulesFreeAbilityWordLabel(label string) bool {
	switch label {
	case "",
		"Coven",
		"Delirium",
		"Domain",
		"Ferocious",
		"Hellbent",
		"Metalcraft",
		"Threshold":
		return true
	default:
		return false
	}
}

func syntaxWithoutAbilityWord(syntax parser.Ability) parser.Ability {
	if syntax.AbilityWord == nil {
		return syntax
	}
	dash := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
		return token.Kind == shared.EmDash
	})
	if dash >= 0 {
		syntax.Tokens = syntax.Tokens[dash+1:]
	}
	return syntax
}

func spellBodyWithoutAbilityWord(
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (compiler.CompiledAbility, parser.Ability, bool) {
	if ability.AbilityWord == "" {
		return ability, syntax, true
	}
	if !rulesFreeAbilityWordLabel(ability.AbilityWord) || syntax.AbilityWord == nil {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	syntax = syntaxWithoutAbilityWord(syntax)
	if len(syntax.Tokens) == 0 {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	start := syntax.Tokens[0].Span.Start
	offset := start.Offset - ability.Span.Start.Offset
	if offset < 0 || offset >= len(ability.Text) {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	ability.Text = strings.TrimSpace(ability.Text[offset:])
	ability.Span.Start = start
	ability.AbilityWord = ""
	syntax.Span.Start = start
	syntax.Text = ability.Text
	syntax.AbilityWord = nil
	return ability, syntax, true
}

func tokensWithoutSpans(tokens []shared.Token, spans ...shared.Span) []shared.Token {
	return slices.DeleteFunc(append([]shared.Token(nil), tokens...), func(token shared.Token) bool {
		return spanCovered(token.Span, spans)
	})
}

func isReadAheadAbility(text string) bool {
	_, ok := readAheadSacrificeChapter(text)
	return ok
}

func readAheadSacrificeChapter(text string) (int, bool) {
	const prefix = "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger."
	remainder, ok := strings.CutPrefix(text, prefix)
	if !ok {
		return 0, false
	}
	if remainder == ")" {
		return 0, true
	}
	chapter, ok := strings.CutPrefix(remainder, " Sacrifice after ")
	if !ok || !strings.HasSuffix(chapter, ".)") {
		return 0, false
	}
	chapter = strings.TrimSuffix(chapter, ".)")
	switch chapter {
	case "I":
		return 1, true
	case "II":
		return 2, true
	case "III":
		return 3, true
	case "IV":
		return 4, true
	case "V":
		return 5, true
	case "VI":
		return 6, true
	default:
		return 0, false
	}
}

// lowerParameterizedKeywordToStaticAbility handles lowering of a single
// parameterized keyword (Ward, Protection, and others) to a static ability.
// Returns (body, true, nil) on success, ({}, true, diag) on a recognised but
// unsupported form, and ({}, false, nil) when no handler matches.
func lowerParameterizedKeywordToStaticAbility(
	ability compiler.CompiledAbility,
	keyword compiler.CompiledKeyword,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	switch keyword.Name {
	case "Ward":
		manaCost, err := parseManaCostValue(keyword.Parameter)
		if err == nil && len(manaCost) > 0 {
			return game.WardStaticAbility(manaCost), true, nil
		}
	case "Protection":
		if keyword.ProtectionKnown {
			return staticAbilityFromProtectionKeyword(keyword.Protection, ""), true, nil
		}
	default:
	}
	if body, ok := lowerParameterizedStaticKeyword(keyword); ok {
		return body, true, nil
	}
	return game.StaticAbility{}, false, nil
}

func lowerParameterizedStaticKeyword(keyword compiler.CompiledKeyword) (game.StaticAbility, bool) {
	body := game.StaticAbility{Text: keyword.Name + " " + keyword.Parameter}
	switch keyword.Name {
	case "Kicker":
		manaCost, ok := parseFixedKeywordManaCost(keyword.Parameter)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.KickerKeyword{Cost: manaCost}}
	case "Madness":
		manaCost, ok := parseFixedKeywordManaCost(keyword.Parameter)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.MadnessKeyword{Cost: manaCost}}
	case "Morph":
		manaCost, ok := parseFixedKeywordManaCost(keyword.Parameter)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.MorphKeyword{Cost: manaCost}}
	case "Disguise":
		manaCost, ok := parseFixedKeywordManaCost(keyword.Parameter)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.DisguiseKeyword{Cost: manaCost}}
	case "Toxic":
		amount, err := strconv.Atoi(keyword.Parameter)
		if err != nil || amount <= 0 {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.ToxicKeyword{Amount: amount}}
	default:
		return game.StaticAbility{}, false
	}
	return body, true
}

func parseFixedKeywordManaCost(parameter string) (cost.Mana, bool) {
	manaCost, err := parseManaCostValue(parameter)
	if err != nil || len(manaCost) == 0 {
		return nil, false
	}
	for _, symbol := range manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return nil, false
		}
	}
	return manaCost, true
}

// lowerManaAbility lowers an activated mana ability into a game.ManaAbility.
// It accepts the same supported cost shapes as ordinary activated abilities,
// plus supported fixed-symbol, choice, and any-color mana output bodies.
// Unrecognised costs and bodies remain fail-closed.
func lowerManaAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax parser.Ability,
) (game.ManaAbility, *shared.Diagnostic) {
	if len(ability.Content.Modes) != 0 {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activation modes",
			"the Payment Planner cannot safely choose modes for a mana ability",
		)
	}
	shell, diagnostic := lowerActivationShell(cardName, ability, syntax)
	if diagnostic != nil {
		return game.ManaAbility{}, diagnostic
	}
	if len(shell.semanticContent.Effects) != 1 ||
		shell.semanticContent.Effects[0].Kind != compiler.EffectAddMana ||
		shell.semanticContent.Effects[0].Negated ||
		len(shell.semanticContent.Keywords) != 0 ||
		len(shell.semanticContent.Targets) != 0 {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana effect",
			"the executable source backend supports only exact non-targeting add-mana content in mana abilities",
		)
	}
	if shell.zoneOfFunction != zone.Battlefield {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activation zone",
			"the Payment Planner supports mana abilities only on the battlefield",
		)
	}

	return game.ManaAbility{
		Text:                shell.text,
		ManaCost:            shell.manaCost,
		Content:             shell.content,
		Timing:              shell.timing,
		ActivationCondition: shell.activationCondition,
		AdditionalCosts:     shell.additionalCosts,
	}, nil
}

func lowerAddManaContent(ctx contentCtx, bodyTokens []shared.Token) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if ctx.optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana effect",
			"the executable source backend supports only exact unconditional add-mana content",
		)
	}
	content, ok := manaBodyContent(bodyTokens)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	return content, nil
}

// manaBodyContent builds game.AbilityContent for exact add-mana content. It
// matches three patterns:
//
//   - "Add one mana of any color." → choice of all five colors
//   - "Add {X} or {Y}." (two or more mana symbols with separators) → choice
//   - "Add {X}." or "Add {X}{Y}." (one or more consecutive mana symbols) → fixed
//
// Any other body returns false.
func manaBodyContent(bodyTokens []shared.Token) (game.AbilityContent, bool) {
	if manaBodyIsAnyColor(bodyTokens) {
		return game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G).Content, true
	}
	if colorNames, ok := manaBodyChoiceColors(bodyTokens); ok {
		colors := make([]mana.Color, 0, len(colorNames))
		for _, name := range colorNames {
			c, ok := manaColorValue(name)
			if !ok {
				return game.AbilityContent{}, false
			}
			colors = append(colors, c)
		}
		if len(colors) < 2 {
			return game.AbilityContent{}, false
		}
		return game.TapManaChoiceAbility(colors...).Content, true
	}
	if colorNames, ok := manaBodyFixedColors(bodyTokens); ok {
		colors := make([]mana.Color, 0, len(colorNames))
		for _, name := range colorNames {
			c, ok := manaColorValue(name)
			if !ok {
				return game.AbilityContent{}, false
			}
			colors = append(colors, c)
		}
		return manaFixedContent(colors), true
	}
	return game.AbilityContent{}, false
}

// manaFixedContent builds AbilityContent that adds one mana of each color in
// the given order. For a single color this produces a single AddMana
// instruction identical to game.TapManaAbility.
func manaFixedContent(colors []mana.Color) game.AbilityContent {
	seq := make([]game.Instruction, 0, len(colors))
	for _, c := range colors {
		seq = append(seq, game.Instruction{
			Primitive: game.AddMana{
				Amount:    game.Fixed(1),
				ManaColor: c,
			},
		})
	}
	return game.Mode{Sequence: seq}.Ability()
}

// manaBodyIsAnyColor reports whether bodyTokens matches the pattern
// "Add one mana of any color.".
func manaBodyIsAnyColor(tokens []shared.Token) bool {
	return len(tokens) == 7 &&
		equalTokenWord(tokens[0], "add") &&
		equalTokenWord(tokens[1], "one") &&
		equalTokenWord(tokens[2], "mana") &&
		equalTokenWord(tokens[3], "of") &&
		equalTokenWord(tokens[4], "any") &&
		equalTokenWord(tokens[5], "color") &&
		tokens[6].Kind == shared.Period
}

// manaBodyChoiceColors extracts the mana color names from a body like
// "Add {R} or {G}." or "Add {W}, {U}, or {B}." Returns false if the pattern
// does not match or fewer than two colors are present.
func manaBodyChoiceColors(tokens []shared.Token) ([]string, bool) {
	if len(tokens) < 4 ||
		!equalTokenWord(tokens[0], "add") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	inner := tokens[1 : len(tokens)-1]
	var colors []string
	for i := 0; i < len(inner); {
		token := inner[i]
		manaColor, ok := manaColorName(token.Text)
		if token.Kind != shared.Symbol || !ok {
			return nil, false
		}
		colors = append(colors, manaColor)
		i++
		if i == len(inner) {
			break
		}
		if inner[i].Kind == shared.Comma {
			i++
			if i < len(inner) && equalTokenWord(inner[i], "or") {
				i++
			}
			continue
		}
		if !equalTokenWord(inner[i], "or") {
			return nil, false
		}
		i++
	}
	return colors, len(colors) >= 2
}

// manaBodyFixedColors extracts the mana color names from a body like "Add {G}."
// or "Add {G}{W}." (one or more consecutive mana symbols). Returns false if
// any inner token is not a recognised mana symbol.
func manaBodyFixedColors(tokens []shared.Token) ([]string, bool) {
	if len(tokens) < 3 ||
		!equalTokenWord(tokens[0], "add") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	inner := tokens[1 : len(tokens)-1]
	if len(inner) == 0 {
		return nil, false
	}
	var colors []string
	for _, token := range inner {
		name, ok := manaColorName(token.Text)
		if token.Kind != shared.Symbol || !ok {
			return nil, false
		}
		colors = append(colors, name)
	}
	return colors, len(colors) >= 1
}

func choiceTapManaAbility(colorNames []string) game.ManaAbility {
	colors := make([]mana.Color, 0, len(colorNames))
	for _, name := range colorNames {
		if manaColor, ok := manaColorValue(name); ok {
			colors = append(colors, manaColor)
		}
	}
	return game.TapManaChoiceAbility(colors...)
}

// contentCtx is the internal lowering context for ability body content.
// It holds the normalized body text (for exact-pattern matching), the source
// span (for diagnostic attribution), the optional flag, and the compiled
// semantic content. It is NOT an compiler.CompiledAbility and carries no shell
// semantics.
type contentCtx struct {
	text     string
	span     shared.Span
	optional bool
	content  compiler.AbilityContent
}

// contentDiagnostic creates a content-level diagnostic attributed to ctx.span.
func contentDiagnostic(ctx contentCtx, summary, detail string) *shared.Diagnostic {
	if summary == "unsupported damage spell" && len(ctx.content.Effects) > 0 {
		lowerText := strings.ToLower(ctx.text)
		if detail == "the executable source backend supports only exact fixed group damage amounts" &&
			ctx.content.Effects[0].Amount.Known &&
			strings.Contains(lowerText, "player or planeswalker it's attacking") {
			detail = "the executable source backend does not support this group recipient"
		} else if detail == "the executable source backend does not support this group recipient" &&
			(strings.Contains(lowerText, "aura deals") && strings.Contains(lowerText, "controller")) {
			detail = "the executable source backend supports only exact fixed group damage amounts"
		}
	}
	return &shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  summary,
		Detail:   detail,
		Span:     ctx.span,
	}
}

// lowerAbilityContent is the single entry point for lowering oracle semantic
// content (targets, conditions, effects, keywords, references) into a
// game.AbilityContent value. All ability shells (spell, activated body,
// triggered body, loyalty body, chapter body, and modal option) route their
// body content through this function. Shell lowerers do not create fake
// AbilitySpell wrappers; they build the adjusted content and body syntax
// directly and call this function.
func lowerAbilityContent(
	cardName string,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:     bodySyntax.Text,
		span:     bodySyntax.Span,
		optional: optional,
		content:  content,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

func lowerContent(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Modes) > 0 {
		return lowerModalContent(cardName, ctx, syntax)
	}
	if content, ok := lowerEventCardEffect(ctx); ok {
		return content, nil
	}
	if exactManifestDreadLongFormPattern(syntax.Tokens) &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 {
		return manifestDreadAbility(), nil
	}
	if len(ctx.content.Effects) > 0 && ctx.content.Effects[0].Kind == compiler.EffectSearch {
		return lowerSearchSpell(ctx)
	}
	if len(ctx.content.Effects) > 1 {
		if ctx.content.Effects[0].Kind == compiler.EffectGainControl ||
			(ctx.content.Effects[0].Kind == compiler.EffectUntap &&
				len(ctx.content.Effects) >= 2 &&
				ctx.content.Effects[1].Kind == compiler.EffectGainControl) {
			return lowerControlSpellSequence(cardName, ctx, syntax)
		}
		return lowerOrderedEffectSequence(cardName, ctx, syntax)
	}
	if len(ctx.content.Effects) == 1 {
		if ctx.content.Effects[0].Kind == compiler.EffectAddMana {
			return lowerAddManaContent(ctx, syntax.Tokens)
		}
		return lowerSingleEffectSpell(cardName, ctx, syntax)
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported ability content",
		"the executable source backend does not yet lower this ability content",
	)
}

func lowerSearchSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported search effect",
			detail,
		)
	}
	// Search is one runtime primitive, but each reference still binds to the
	// prior semantic search/reveal instruction that produced the found card.
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return unsupported("unexpected non-result reference in search effect")
		}
	}
	consumed := ctx
	consumed.content.References = nil
	if ctx.optional ||
		consumed.content.Unconsumed() ||
		!exactSearchEffectSequence(ctx.content.Effects) {
		return unsupported("the executable source backend supports only exact unconditional library-search sequences")
	}
	search := ctx.content.Effects[0]
	if !search.Amount.Known || search.Amount.Value != 1 {
		return unsupported("the executable source backend supports only searches for exactly one card")
	}
	text := search.Text
	for _, effect := range ctx.content.Effects {
		if effect.Text != text ||
			effect.DelayedTiming != 0 ||
			effect.Duration != compiler.DurationNone ||
			effect.Negated {
			return unsupported("the executable source backend supports only exact same-sentence library-search sequences")
		}
	}
	if !strings.HasPrefix(text, "Search your library for ") || !strings.HasSuffix(text, ", then shuffle.") {
		return unsupported("the executable source backend supports only searches of your library ending with \"then shuffle\"")
	}

	filter, ok := searchFilterPhrase(text)
	if !ok {
		return unsupported("the executable source backend supports only exact singular-card search wording")
	}
	spec, ok := searchSpecForFilter(filter)
	if !ok {
		return unsupported(fmt.Sprintf("unsupported library-search filter %q", filter))
	}
	spec.SourceZone = zone.Library

	spec.Reveal = len(ctx.content.Effects) == 4
	destination, entersTapped, ok := searchDestination(text, spec.Reveal)
	if !ok {
		return unsupported("the executable source backend supports only exact hand or battlefield search destinations")
	}
	spec.Destination = destination
	spec.EntersTapped = entersTapped

	return game.Mode{Sequence: []game.Instruction{{Primitive: game.Search{
		Player: game.ControllerReference(),
		Spec:   spec,
		Amount: game.Fixed(1),
	}}}}.Ability(), nil
}

func exactSearchEffectSequence(effects []compiler.CompiledEffect) bool {
	if len(effects) == 3 {
		return effects[0].Kind == compiler.EffectSearch &&
			effects[1].Kind == compiler.EffectPut &&
			effects[2].Kind == compiler.EffectShuffle
	}
	return len(effects) == 4 &&
		effects[0].Kind == compiler.EffectSearch &&
		effects[1].Kind == compiler.EffectReveal &&
		effects[2].Kind == compiler.EffectPut &&
		effects[3].Kind == compiler.EffectShuffle
}

func searchFilterPhrase(text string) (string, bool) {
	for _, prefix := range []string{
		"Search your library for a ",
		"Search your library for an ",
	} {
		rest, ok := strings.CutPrefix(text, prefix)
		if !ok {
			continue
		}
		if strings.HasPrefix(rest, "card,") {
			return "", true
		}
		filter, _, ok := strings.Cut(rest, " card,")
		return filter, ok
	}
	return "", false
}

func searchSpecForFilter(filter string) (game.SearchSpec, bool) {
	var spec game.SearchSpec
	switch filter {
	case "":
	case "basic land":
		spec.CardType = opt.Val(types.Land)
		spec.Supertype = opt.Val(types.Basic)
	case "land":
		spec.CardType = opt.Val(types.Land)
	case "creature":
		spec.CardType = opt.Val(types.Creature)
	case "artifact":
		spec.CardType = opt.Val(types.Artifact)
	case "enchantment":
		spec.CardType = opt.Val(types.Enchantment)
	case "Forest":
		spec.SubtypesAny = []types.Sub{types.Forest}
	case "Plains":
		spec.SubtypesAny = []types.Sub{types.Plains}
	case "Island":
		spec.SubtypesAny = []types.Sub{types.Island}
	case "Swamp":
		spec.SubtypesAny = []types.Sub{types.Swamp}
	case "Mountain":
		spec.SubtypesAny = []types.Sub{types.Mountain}
	case "Forest or Plains":
		spec.SubtypesAny = []types.Sub{types.Forest, types.Plains}
	case "Plains, Island, Swamp, or Mountain":
		spec.SubtypesAny = []types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain}
	default:
		return game.SearchSpec{}, false
	}
	return spec, true
}

func searchDestination(text string, reveal bool) (destination zone.Type, entersTapped, ok bool) {
	type searchDestinationPattern struct {
		suffix        string
		destination   zone.Type
		entersTapped  bool
		revealsSearch bool
	}
	for _, pattern := range []searchDestinationPattern{
		{suffix: ", put it into your hand, then shuffle.", destination: zone.Hand},
		{suffix: ", put that card into your hand, then shuffle.", destination: zone.Hand},
		{suffix: ", reveal it, put it into your hand, then shuffle.", destination: zone.Hand, revealsSearch: true},
		{suffix: ", reveal that card, put it into your hand, then shuffle.", destination: zone.Hand, revealsSearch: true},
		{suffix: ", put it onto the battlefield, then shuffle.", destination: zone.Battlefield},
		{suffix: ", put that card onto the battlefield, then shuffle.", destination: zone.Battlefield},
		{suffix: ", put it onto the battlefield tapped, then shuffle.", destination: zone.Battlefield, entersTapped: true},
		{suffix: ", put that card onto the battlefield tapped, then shuffle.", destination: zone.Battlefield, entersTapped: true},
	} {
		if reveal == pattern.revealsSearch && strings.HasSuffix(text, pattern.suffix) {
			return pattern.destination, pattern.entersTapped, true
		}
	}
	return zone.None, false, false
}

func lowerSingleEffectSpell(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Effects) == 1 && ctx.content.Effects[0].DelayedTiming != 0 {
		return lowerDelayedSingleEffectSpell(cardName, ctx, syntax)
	}
	return lowerImmediateSingleEffectSpell(cardName, ctx, syntax)
}

func lowerDelayedSingleEffectSpell(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	ctx.text = textWithoutDelimited(ctx.text, ctx.span, syntax.Reminders)
	syntax.Tokens = slices.DeleteFunc(
		append([]shared.Token(nil), syntax.Tokens...),
		func(token shared.Token) bool {
			return spanCoveredByDelimited(token.Span, syntax.Reminders)
		},
	)
	syntax.Reminders = nil
	text, textOK := stripDelayedTimingText(ctx.text, effect.DelayedTiming)
	tokens, tokensOK := stripDelayedTimingTokens(syntax.Tokens, effect.DelayedTiming)
	if !textOK || !tokensOK {
		return game.AbilityContent{}, unsupportedDelayedEffectDiagnostic(ctx)
	}
	ctx.text = text
	ctx.content.Effects[0].Text = text
	ctx.content.Effects[0].DelayedTiming = 0
	syntax.Tokens = tokens

	var content game.AbilityContent
	if primitive, ok := lowerDelayedSelfPrimitive(ctx); ok {
		content = game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability()
	} else {
		var diagnostic *shared.Diagnostic
		content, diagnostic = lowerImmediateSingleEffectSpell(cardName, ctx, syntax)
		if diagnostic != nil {
			return game.AbilityContent{}, unsupportedDelayedEffectDiagnostic(ctx)
		}
	}
	if len(content.SharedTargets) != 0 ||
		content.IsModal() ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) == 0 {
		return game.AbilityContent{}, unsupportedDelayedEffectDiagnostic(ctx)
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing:  effect.DelayedTiming,
			Content: content,
		},
	}}}}.Ability(), nil
}

func lowerDelayedSelfPrimitive(ctx contentCtx) (game.Primitive, bool) {
	if ctx.content.Effects[0].Negated {
		return nil, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingSource, 0) {
		return nil, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return nil, false
	}
	sourcePermanent, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource:      true,
		SourceCardObject: true,
	})
	if !ok {
		return nil, false
	}
	switch ctx.text {
	case "Exile it.":
		return game.Exile{Object: sourcePermanent}, true
	case "Sacrifice it.":
		return game.Sacrifice{Object: sourcePermanent}, true
	case "Return it to its owner's hand.":
		sourceCard, ok := lowerCardReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if !ok {
			return nil, false
		}
		return game.MoveCard{
			Card:        sourceCard,
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}, true
	default:
		return nil, false
	}
}

func stripDelayedTimingText(text string, timing game.DelayedTriggerTiming) (string, bool) {
	var suffix string
	switch timing {
	case game.DelayedAtBeginningOfNextEndStep:
		suffix = " at the beginning of the next end step."
	case game.DelayedAtBeginningOfNextUpkeep:
		suffix = " at the beginning of the next turn's upkeep."
	default:
		return "", false
	}
	base, ok := strings.CutSuffix(text, suffix)
	if !ok || base == "" {
		return "", false
	}
	return base + ".", true
}

func stripDelayedTimingTokens(tokens []shared.Token, timing game.DelayedTriggerTiming) ([]shared.Token, bool) {
	var suffix []string
	switch timing {
	case game.DelayedAtBeginningOfNextEndStep:
		suffix = []string{"at", "the", "beginning", "of", "the", "next", "end", "step"}
	case game.DelayedAtBeginningOfNextUpkeep:
		suffix = []string{"at", "the", "beginning", "of", "the", "next", "turn's", "upkeep"}
	default:
		return nil, false
	}
	if len(tokens) < len(suffix)+1 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	start := len(tokens) - len(suffix) - 1
	for i, text := range suffix {
		if !strings.EqualFold(tokens[start+i].Text, text) {
			return nil, false
		}
	}
	stripped := append([]shared.Token(nil), tokens[:start]...)
	stripped = append(stripped, tokens[len(tokens)-1])
	return stripped, true
}

func unsupportedDelayedEffectDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported delayed effect",
		"the executable source backend supports only exact non-target delayed one-shot effects",
	)
}

// lowerEventPermanentPronounEffect lowers an immediate single-effect body
// whose sole non-target subject is a no-target ReferenceBindingEventPermanent
// pronoun ("it"/"its"). All references must bind to EventPermanent; no targets,
// conditions, keywords, or modes are permitted. Accepted bodies are:
//
//   - "Destroy it."      → game.Destroy{Object: EventPermanentReference()}
//   - "Exile it."        → game.Exile{Object: EventPermanentReference()}
//   - "Tap it."          → game.Tap{Object: EventPermanentReference()}
//   - "Untap it."        → game.Untap{Object: EventPermanentReference()}
//   - "Sacrifice it."    → game.Sacrifice{Object: EventPermanentReference()}
//   - "Return it to its owner's hand." → game.Bounce{Object: EventPermanentReference()}
func lowerEventPermanentPronounEffect(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) == 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Effects[0].Negated {
		return game.AbilityContent{}, false
	}
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingEventPermanent {
			return game.AbilityContent{}, false
		}
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	var primitive game.Primitive
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectDestroy:
		if ctx.text != "Destroy it." {
			return game.AbilityContent{}, false
		}
		primitive = game.Destroy{Object: object}
	case compiler.EffectExile:
		if ctx.text != "Exile it." {
			return game.AbilityContent{}, false
		}
		primitive = game.Exile{Object: object}
	case compiler.EffectTap:
		if ctx.text != "Tap it." {
			return game.AbilityContent{}, false
		}
		primitive = game.Tap{Object: object}
	case compiler.EffectUntap:
		if ctx.text != "Untap it." {
			return game.AbilityContent{}, false
		}
		primitive = game.Untap{Object: object}
	case compiler.EffectSacrifice:
		if ctx.text != "Sacrifice it." {
			return game.AbilityContent{}, false
		}
		primitive = game.Sacrifice{Object: object}
	case compiler.EffectReturn:
		if ctx.text != "Return it to its owner's hand." {
			return game.AbilityContent{}, false
		}
		primitive = game.Bounce{Object: object}
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(), true
}

func lowerImmediateSingleEffectSpell(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx.text = textWithoutDelimited(ctx.text, ctx.span, syntax.Reminders)
	syntax.Tokens = slices.DeleteFunc(
		append([]shared.Token(nil), syntax.Tokens...),
		func(token shared.Token) bool {
			return spanCoveredByDelimited(token.Span, syntax.Reminders)
		},
	)
	// Route no-target EventPermanent pronoun bodies through the shared path
	// before individual effect dispatch so all compatible trigger shells
	// benefit from the same lowering.
	if content, ok := lowerEventPermanentPronounEffect(ctx); ok {
		return content, nil
	}
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectDealDamage:
		if len(ctx.content.Targets) == 0 {
			return lowerGroupDamageSpell(cardName, ctx)
		}
		return lowerFixedDamageSpell(cardName, ctx)
	case compiler.EffectDraw:
		return lowerFixedDrawSpell(ctx, syntax)
	case compiler.EffectDestroy:
		return lowerFixedDestroySpell(ctx)
	case compiler.EffectGain:
		if len(ctx.content.Keywords) != 0 &&
			ctx.content.Effects[0].Duration == compiler.DurationUntilEndOfTurn {
			return lowerTemporaryKeywordSpell(ctx, syntax)
		}
		return lowerFixedLifeSpell(ctx, "gain", func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.GainLife{Amount: amount, Player: player}
		}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
			return game.GainLife{Amount: amount, PlayerGroup: group}
		})
	case compiler.EffectGainControl:
		return lowerSingleControlSpell(ctx)
	case compiler.EffectLose:
		return lowerFixedLifeSpell(ctx, "lose", func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.LoseLife{Amount: amount, Player: player}
		}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
			return game.LoseLife{Amount: amount, PlayerGroup: group}
		})
	case compiler.EffectScry:
		return lowerFixedControllerSpell(ctx, syntax, "scry", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.Scry{Amount: amount, Player: player}
		})
	case compiler.EffectSurveil:
		return lowerFixedControllerSpell(ctx, syntax, "surveil", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.Surveil{Amount: amount, Player: player}
		})
	case compiler.EffectInvestigate:
		return lowerInvestigateSpell(ctx, syntax)
	case compiler.EffectProliferate:
		return lowerExactPrimitiveSpell(ctx, syntax, "proliferate", func(amount game.Quantity) game.Primitive {
			return game.Proliferate{Amount: amount}
		})
	case compiler.EffectExplore:
		return lowerExploreSpell(ctx, syntax)
	case compiler.EffectManifest, compiler.EffectManifestDread:
		return lowerManifestSpell(ctx, syntax)
	case compiler.EffectRegenerate:
		return lowerFixedPermanentTargetSpell(ctx, "Regenerate", func(object game.ObjectReference) game.Primitive {
			return game.Regenerate{Object: object}
		})
	case compiler.EffectFight:
		return lowerFightSpell(ctx)
	case compiler.EffectDiscard:
		return lowerFixedCardCountPlayerSpell(
			ctx, syntax, "discard", "discards", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Discard{Amount: amount, Player: player}
			},
		)
	case compiler.EffectMill:
		return lowerFixedCardCountPlayerSpell(
			ctx, syntax, "mill", "mills", true, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Mill{Amount: amount, Player: player}
			},
		)
	case compiler.EffectTap:
		return lowerFixedPermanentTargetSpell(ctx, "Tap", func(object game.ObjectReference) game.Primitive {
			return game.Tap{Object: object}
		})
	case compiler.EffectUntap:
		return lowerFixedPermanentTargetSpell(ctx, "Untap", func(object game.ObjectReference) game.Primitive {
			return game.Untap{Object: object}
		})
	case compiler.EffectExile:
		return lowerFixedExileSpell(ctx)
	case compiler.EffectReturn:
		if content, ok := lowerSelfCardGraveyardReturn(ctx); ok {
			return content, nil
		}
		if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
			return content, nil
		}
		return lowerFixedBounceSpell(ctx)
	case compiler.EffectPut:
		if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
			return content, nil
		}
		return lowerCounterPlacementSpell(ctx)
	case compiler.EffectModifyPT:
		return lowerFixedModifyPTSpell(ctx, syntax)
	case compiler.EffectCounter:
		return lowerCounterSpell(ctx)
	case compiler.EffectSacrifice:
		return lowerSacrificeSpell(ctx, syntax)
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported ability content",
			"the executable source backend does not yet lower this ability content",
		)
	}
}

func lowerSelfCardGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectReturn ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		!selfCardGraveyardReturnReferences(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	sourceCard, ok := lowerCardReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	switch {
	case ctx.text == "Return this card from your graveyard to your hand.":
		return game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
			Card:        sourceCard,
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}}}}.Ability(), true
	case ctx.text == "Return this card from your graveyard to the battlefield.":
		return game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(sourceCard),
		}}}}.Ability(), true
	case strings.HasPrefix(ctx.text, "Return this card from your graveyard to the battlefield"):
		tapped, counters, ok := selfCardBattlefieldReturnModifiers(ctx.text, ctx.content.Effects[0])
		if !ok {
			return game.AbilityContent{}, false
		}
		put := game.PutOnBattlefield{
			Source:      game.CardBattlefieldSource(sourceCard),
			EntryTapped: tapped,
		}
		if counters > 0 {
			put.EntryCounters = []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: counters}}
		}
		return game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

func selfCardGraveyardReturnReferences(references []compiler.CompiledReference) bool {
	return referencesBindTo(references, compiler.ReferenceBindingSource, 0)
}

func selfCardBattlefieldReturnModifiers(text string, effect compiler.CompiledEffect) (tapped bool, counters int, ok bool) {
	const prefix = "Return this card from your graveyard to the battlefield"
	suffix := strings.TrimPrefix(text, prefix)
	switch suffix {
	case ".":
		return false, 0, true
	case " tapped.":
		return true, 0, true
	default:
	}
	if strings.HasPrefix(suffix, " tapped with ") {
		tapped = true
		suffix = strings.TrimPrefix(suffix, " tapped")
	}
	if strings.HasPrefix(suffix, " with ") && strings.HasSuffix(suffix, " +1/+1 counters on it.") {
		return tapped, effect.Amount.Value, effect.Amount.Known &&
			effect.CounterKindKnown &&
			effect.CounterKind == counter.PlusOnePlusOne
	}
	return false, 0, false
}

func lowerTargetedGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		ctx.content.Effects[0].FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	switch ctx.content.Effects[0].ToZone {
	case zone.Hand:
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, game.Instruction{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}})
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	case zone.Library:
		destinationBottom, ok := graveyardReturnLibraryBottom(ctx.content.Targets[0].Text)
		if !ok {
			return game.AbilityContent{}, false
		}
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, game.Instruction{Primitive: game.MoveCard{
				Card:              game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i},
				FromZone:          zone.Graveyard,
				Destination:       zone.Library,
				DestinationBottom: destinationBottom,
			}})
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	case zone.Battlefield:
		for i := range targetSpec.MaxTargets {
			put, ok := targetedGraveyardBattlefieldPut(ctx.text, game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i})
			if !ok {
				return game.AbilityContent{}, false
			}
			sequence = append(sequence, game.Instruction{Primitive: put})
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

func targetedGraveyardBattlefieldPut(text string, targetCard game.CardReference) (game.PutOnBattlefield, bool) {
	put := game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(targetCard),
	}
	text = strings.TrimSuffix(text, ".")
	for {
		switch {
		case strings.HasSuffix(text, " under your control"):
			text = strings.TrimSuffix(text, " under your control")
			put.Recipient = opt.Val(game.ControllerReference())
		case strings.HasSuffix(text, " tapped"):
			text = strings.TrimSuffix(text, " tapped")
			put.EntryTapped = true
		default:
			if strings.HasSuffix(text, " to the battlefield") || strings.HasSuffix(text, " onto the battlefield") {
				return put, true
			}
			return game.PutOnBattlefield{}, false
		}
	}
}

func graveyardReturnLibraryBottom(text string) (destinationBottom, recognized bool) {
	switch {
	case strings.HasSuffix(text, " on top of your library") ||
		strings.HasSuffix(text, " on the top of your library"):
		return false, true
	case strings.HasSuffix(text, " on bottom of your library") ||
		strings.HasSuffix(text, " on the bottom of your library"):
		return true, true
	default:
		return false, false
	}
}

func cardInZoneTargetSpec(target compiler.CompiledTarget, targetZone zone.Type) (game.TargetSpec, bool) {
	if target.Cardinality.Min < 0 || target.Cardinality.Max < target.Cardinality.Min ||
		target.Cardinality.Max == 0 ||
		target.Selector.Another || target.Selector.Other ||
		target.Selector.Attacking || target.Selector.Blocking {
		return game.TargetSpec{}, false
	}
	targetText := graveyardCardTargetText(target.Text)
	const targetPrefix = "target "
	targetIndex := strings.Index(strings.ToLower(targetText), targetPrefix)
	if targetIndex < 0 {
		return game.TargetSpec{}, false
	}
	targetText = targetText[targetIndex:]
	targetBody := targetText[len(targetPrefix):]
	controller := game.ControllerAny
	switch {
	case strings.HasSuffix(targetBody, " from your graveyard"):
		targetBody = strings.TrimSuffix(targetBody, " from your graveyard")
		controller = game.ControllerYou
	case strings.HasSuffix(targetBody, " from a graveyard"):
		targetBody = strings.TrimSuffix(targetBody, " from a graveyard")
	case strings.HasSuffix(targetBody, " from an opponent's graveyard"):
		targetBody = strings.TrimSuffix(targetBody, " from an opponent's graveyard")
		controller = game.ControllerOpponent
	default:
		return game.TargetSpec{}, false
	}
	targetBody = strings.ToLower(targetBody)
	spec := game.TargetSpec{
		MinTargets: target.Cardinality.Min,
		MaxTargets: target.Cardinality.Max,
		Constraint: lowerFirst(targetText),
		Allow:      game.TargetAllowCard,
		TargetZone: targetZone,
	}
	var selection game.Selection
	switch targetBody {
	case "card":
	case "card with cycling", "cards with cycling", "card with a cycling ability", "cards with a cycling ability":
		selection.Keyword = game.Cycling
	case "instant or sorcery card":
		selection.RequiredTypesAny = []types.Card{types.Instant, types.Sorcery}
	case "instant or sorcery card with cycling", "instant or sorcery cards with cycling", "instant or sorcery card with a cycling ability", "instant or sorcery cards with a cycling ability":
		selection.RequiredTypesAny = []types.Card{types.Instant, types.Sorcery}
		selection.Keyword = game.Cycling
	case "artifact card":
		selection.RequiredTypes = []types.Card{types.Artifact}
	case "artifact card with cycling", "artifact cards with cycling", "artifact card with a cycling ability", "artifact cards with a cycling ability":
		selection.RequiredTypes = []types.Card{types.Artifact}
		selection.Keyword = game.Cycling
	case "creature card":
		selection.RequiredTypes = []types.Card{types.Creature}
	case "creature card with cycling", "creature cards with cycling", "creature card with a cycling ability", "creature cards with a cycling ability":
		selection.RequiredTypes = []types.Card{types.Creature}
		selection.Keyword = game.Cycling
	case "enchantment card":
		selection.RequiredTypes = []types.Card{types.Enchantment}
	case "enchantment card with cycling", "enchantment cards with cycling", "enchantment card with a cycling ability", "enchantment cards with a cycling ability":
		selection.RequiredTypes = []types.Card{types.Enchantment}
		selection.Keyword = game.Cycling
	case "land card":
		selection.RequiredTypes = []types.Card{types.Land}
	case "land card with cycling", "land cards with cycling", "land card with a cycling ability", "land cards with a cycling ability":
		selection.RequiredTypes = []types.Card{types.Land}
		selection.Keyword = game.Cycling
	case "planeswalker card":
		selection.RequiredTypes = []types.Card{types.Planeswalker}
	case "planeswalker card with cycling", "planeswalker cards with cycling", "planeswalker card with a cycling ability", "planeswalker cards with a cycling ability":
		selection.RequiredTypes = []types.Card{types.Planeswalker}
		selection.Keyword = game.Cycling
	case "permanent card with cycling", "permanent cards with cycling", "permanent card with a cycling ability", "permanent cards with a cycling ability":
		selection.RequiredTypesAny = []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}
		selection.Keyword = game.Cycling
	case "vehicle card":
		selection.SubtypesAny = []types.Sub{types.Vehicle}
	case "vehicle card with cycling", "vehicle cards with cycling", "vehicle card with a cycling ability", "vehicle cards with a cycling ability":
		selection.SubtypesAny = []types.Sub{types.Vehicle}
		selection.Keyword = game.Cycling
	default:
		if !lowerCardTargetManaValuePredicate(targetBody, &selection) {
			return game.TargetSpec{}, false
		}
	}
	selection.Controller = controller
	spec.Selection = opt.Val(selection)
	return spec, true
}

func lowerCardTargetManaValuePredicate(targetBody string, selection *game.Selection) bool {
	const predicate = " with mana value "
	cardType, comparisonText, ok := strings.Cut(targetBody, predicate)
	if !ok {
		return false
	}
	switch cardType {
	case "artifact card":
		selection.RequiredTypes = []types.Card{types.Artifact}
	case "creature card":
		selection.RequiredTypes = []types.Card{types.Creature}
	case "enchantment card":
		selection.RequiredTypes = []types.Card{types.Enchantment}
	case "instant or sorcery card":
		selection.RequiredTypesAny = []types.Card{types.Instant, types.Sorcery}
	case "land card":
		selection.RequiredTypes = []types.Card{types.Land}
	case "planeswalker card":
		selection.RequiredTypes = []types.Card{types.Planeswalker}
	case "vehicle card":
		selection.SubtypesAny = []types.Sub{types.Vehicle}
	default:
		return false
	}
	parts := strings.Fields(comparisonText)
	if len(parts) != 3 {
		return false
	}
	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	switch strings.Join(parts[1:], " ") {
	case "or less":
		selection.ManaValue = opt.Val(compare.Int{Op: compare.LessOrEqual, Value: value})
	case "or greater", "or more":
		selection.ManaValue = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: value})
	default:
		return false
	}
	return true
}

func graveyardCardTargetText(text string) string {
	for _, suffix := range []string{
		" to your hand",
		" to their hand",
		" to the battlefield under your control",
		" onto the battlefield under your control",
		" to the battlefield tapped under your control",
		" onto the battlefield tapped under your control",
		" to the battlefield tapped",
		" onto the battlefield tapped",
		" to the battlefield",
		" onto the battlefield",
		" on top of your library",
		" on the top of your library",
		" on bottom of your library",
		" on the bottom of your library",
		" into your library",
	} {
		if trimmed, ok := strings.CutSuffix(text, suffix); ok {
			return trimmed
		}
	}
	return text
}

func lowerCounterPlacementSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		(effect.Amount.Known && effect.Amount.Value <= 0) ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}

	kind := effect.CounterKind
	counterName := kind.String()
	var target game.TargetSpec
	var primitive game.Primitive
	if kind.PlayerOnly() {
		var ok bool
		target, ok = playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
	} else {
		var ok bool
		target, ok = permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
	}

	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	exactText := exactXCounterText(ctx, counterName) && len(ctx.content.References) == 0
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
		exactText = isExactPutCounterText(
			ctx.text,
			ctx.content.Targets[0].Text,
			effect.Amount.Value,
			counterName,
		) && len(ctx.content.References) == 0
	} else if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		dynamic, supported := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !supported ||
			!exactDynamicCounterText(ctx, counterName) ||
			!exactDynamicAmountReference(effect.Amount, ctx.content.References) {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		amount = game.Dynamic(dynamic)
		exactText = true
	}
	if !exactText {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	if kind.PlayerOnly() {
		primitive = game.AddPlayerCounter{
			Amount:      amount,
			Player:      game.TargetPlayerReference(0),
			CounterKind: kind,
		}
	} else {
		primitive = game.AddCounter{
			Amount:      amount,
			Object:      game.TargetPermanentReference(0),
			CounterKind: kind,
		}
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: primitive,
		}},
	}.Ability(), nil
}

func unsupportedCounterPlacementDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported counter placement",
		"the executable source backend supports exact recognized counter placement on one valid target",
	)
}

func exactXCounterText(ctx contentCtx, counterName string) bool {
	return ctx.text == fmt.Sprintf(
		"Put X %s counters on %s.",
		counterName,
		ctx.content.Targets[0].Text,
	)
}

func exactDynamicCounterText(ctx contentCtx, counterName string) bool {
	amount := ctx.content.Effects[0].Amount
	return amount.DynamicForm == compiler.DynamicAmountWhereX &&
		ctx.text == fmt.Sprintf(
			"Put X %s counters on %s, %s.",
			counterName,
			ctx.content.Targets[0].Text,
			amount.Text,
		)
}

func exactDynamicAmountReference(
	amount compiler.CompiledAmount,
	references []compiler.CompiledReference,
) bool {
	if amount.DynamicKind != compiler.DynamicAmountSourcePower {
		return len(references) == 0
	}
	if len(references) != 1 || references[0].Span != amount.ReferenceSpan {
		return false
	}
	return references[0].Binding == compiler.ReferenceBindingSource
}

func isExactPutCounterText(text, targetText string, amount int, counterName string) bool {
	amountWords := []string{strconv.Itoa(amount)}
	switch amount {
	case 1:
		amountWords = append(amountWords, "a", "an", "one")
	case 2:
		amountWords = append(amountWords, "two")
	case 3:
		amountWords = append(amountWords, "three")
	case 4:
		amountWords = append(amountWords, "four")
	case 5:
		amountWords = append(amountWords, "five")
	default:
	}
	noun := "counters"
	if amount == 1 {
		noun = "counter"
	}
	for _, amountWord := range amountWords {
		if text == fmt.Sprintf("Put %s %s %s on %s.", amountWord, counterName, noun, targetText) {
			return true
		}
	}
	return false
}

func textWithoutDelimited(text string, span shared.Span, groups []parser.Delimited) string {
	var result strings.Builder
	cursor := span.Start.Offset
	for _, group := range groups {
		if group.Span.Start.Offset < cursor ||
			group.Span.End.Offset > span.End.Offset {
			continue
		}
		start := group.Span.Start.Offset - span.Start.Offset
		end := cursor - span.Start.Offset
		_, _ = result.WriteString(text[end:start])
		cursor = group.Span.End.Offset
	}
	_, _ = result.WriteString(text[cursor-span.Start.Offset:])
	return strings.TrimSpace(result.String())
}

func lowerFightSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) != 2 ||
		ctx.content.Targets[0].Cardinality != (compiler.TargetCardinality{Min: 1, Max: 1}) ||
		ctx.content.Targets[1].Cardinality != (compiler.TargetCardinality{Min: 1, Max: 1}) ||
		ctx.content.Effects[0].Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.text != titleFirst(ctx.content.Targets[0].Text)+" fights "+ctx.content.Targets[1].Text+"." {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported fight spell",
			"the executable source backend supports only exact fights between two target creatures",
		)
	}
	first, firstOK := fightCreatureTargetSpec(ctx.content.Targets[0])
	second, secondOK := fightCreatureTargetSpec(ctx.content.Targets[1])
	if !firstOK || !secondOK {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported fight spell",
			"the executable source backend supports only exact fights between two target creatures",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{first, second},
		Sequence: []game.Instruction{{
			Primitive: game.Fight{
				Object:        game.TargetPermanentReference(0),
				RelatedObject: game.TargetPermanentReference(1),
			},
		}},
	}.Ability(), nil
}

func fightCreatureTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if target.Selector.Kind != compiler.SelectorCreature ||
		target.Selector.Another ||
		target.Selector.Other ||
		target.Selector.Attacking ||
		target.Selector.Blocking ||
		target.Selector.Tapped ||
		target.Selector.Untapped {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPermanent,
		Predicate: game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
		},
	}
	var expected string
	switch target.Selector.Controller {
	case compiler.ControllerAny:
		expected = "target creature"
	case compiler.ControllerYou:
		expected = "target creature you control"
		spec.Predicate.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		expected = "target creature an opponent controls"
		spec.Predicate.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		expected = "target creature you don't control"
		spec.Predicate.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	return spec, strings.EqualFold(target.Text, expected)
}

func lowerInvestigateSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	return lowerExactPrimitiveSpell(
		ctx,
		syntax,
		"investigate",
		func(amount game.Quantity) game.Primitive {
			return game.Investigate{Amount: amount}
		},
	)
}

func lowerExploreSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	tokens := syntax.Tokens
	unsupportedExplore := contentDiagnostic(
		ctx,
		"unsupported explore spell",
		"the executable source backend supports only the source permanent pattern \"it explores\"",
	)
	if ctx.content.Effects[0].Negated ||
		len(tokens) != 3 ||
		!equalTokenWord(tokens[0], "it") ||
		!equalTokenWord(tokens[1], "explores") ||
		tokens[2].Kind != shared.Period ||
		len(ctx.content.References) != 1 ||
		(ctx.content.References[0].Binding != compiler.ReferenceBindingSource &&
			ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent) {
		return game.AbilityContent{}, unsupportedExplore
	}
	// Reference validated as "it" pronoun — clear before the fail-closed check.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, unsupportedExplore
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
	if !ok {
		return game.AbilityContent{}, unsupportedExplore
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Explore{Creature: object},
	}}}.Ability(), nil
}

func lowerManifestSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	tokens := syntax.Tokens
	if ctx.content.Effects[0].Negated ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 ||
		!exactManifestTopLibraryPattern(tokens) &&
			!exactManifestDreadShorthandPattern(tokens) &&
			!exactManifestDreadLongFormPattern(tokens) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported manifest spell",
			"the executable source backend supports only \"manifest the top card of your library\" and manifest dread",
		)
	}
	dread := exactManifestDreadShorthandPattern(tokens) || exactManifestDreadLongFormPattern(tokens)
	if dread {
		return manifestDreadAbility(), nil
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Manifest{},
	}}}.Ability(), nil
}

func manifestDreadAbility() game.AbilityContent {
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Manifest{Dread: true},
	}}}.Ability()
}

func exactManifestTopLibraryPattern(tokens []shared.Token) bool {
	return len(tokens) == 8 &&
		equalTokenWord(tokens[0], "manifest") &&
		equalTokenWord(tokens[1], "the") &&
		equalTokenWord(tokens[2], "top") &&
		equalTokenWord(tokens[3], "card") &&
		equalTokenWord(tokens[4], "of") &&
		equalTokenWord(tokens[5], "your") &&
		equalTokenWord(tokens[6], "library") &&
		tokens[7].Kind == shared.Period
}

func exactManifestDreadShorthandPattern(tokens []shared.Token) bool {
	return len(tokens) == 3 &&
		equalTokenWord(tokens[0], "manifest") &&
		equalTokenWord(tokens[1], "dread") &&
		tokens[2].Kind == shared.Period
}

func exactManifestDreadLongFormPattern(tokens []shared.Token) bool {
	return len(tokens) == 33 &&
		equalTokenWord(tokens[0], "look") &&
		equalTokenWord(tokens[1], "at") &&
		equalTokenWord(tokens[2], "the") &&
		equalTokenWord(tokens[3], "top") &&
		equalTokenWord(tokens[4], "two") &&
		equalTokenWord(tokens[5], "cards") &&
		equalTokenWord(tokens[6], "of") &&
		equalTokenWord(tokens[7], "your") &&
		equalTokenWord(tokens[8], "library") &&
		tokens[9].Kind == shared.Period &&
		equalTokenWord(tokens[10], "put") &&
		equalTokenWord(tokens[11], "one") &&
		equalTokenWord(tokens[12], "of") &&
		equalTokenWord(tokens[13], "them") &&
		equalTokenWord(tokens[14], "onto") &&
		equalTokenWord(tokens[15], "the") &&
		equalTokenWord(tokens[16], "battlefield") &&
		equalTokenWord(tokens[17], "face") &&
		equalTokenWord(tokens[18], "down") &&
		equalTokenWord(tokens[19], "as") &&
		equalTokenWord(tokens[20], "a") &&
		tokens[21].Kind == shared.Integer &&
		tokens[21].Text == "2" &&
		tokens[22].Kind == shared.Slash &&
		tokens[23].Kind == shared.Integer &&
		tokens[23].Text == "2" &&
		equalTokenWord(tokens[24], "creature") &&
		tokens[25].Kind == shared.Period &&
		equalTokenWord(tokens[26], "put") &&
		equalTokenWord(tokens[27], "the") &&
		equalTokenWord(tokens[28], "other") &&
		equalTokenWord(tokens[29], "into") &&
		equalTokenWord(tokens[30], "your") &&
		equalTokenWord(tokens[31], "graveyard") &&
		tokens[32].Kind == shared.Period
}

func lowerExactPrimitiveSpell(
	ctx contentCtx,
	syntax parser.Ability,
	verb string,
	primitiveFactory func(game.Quantity) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	amount, ok := standaloneActionAmount(syntax.Tokens, syntax.Atoms, verb, ctx.content.Effects[0].Amount)
	if ctx.content.Effects[0].Negated ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 ||
		!ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact "+verb,
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: primitiveFactory(game.Fixed(amount)),
	}}}.Ability(), nil
}

func standaloneActionAmount(tokens []shared.Token, atoms parser.Atoms, verb string, compiled compiler.CompiledAmount) (int, bool) {
	if len(tokens) == 2 &&
		equalTokenWord(tokens[0], verb) &&
		tokens[1].Kind == shared.Period {
		return 1, true
	}
	if len(tokens) == 3 &&
		equalTokenWord(tokens[0], verb) &&
		tokens[2].Kind == shared.Period {
		if compiled.Known && fixedNumberSyntax(tokens[1], atoms, compiled.Value) {
			return compiled.Value, true
		}
	}
	if len(tokens) == 4 &&
		equalTokenWord(tokens[0], verb) &&
		equalTokenWord(tokens[2], "times") &&
		tokens[3].Kind == shared.Period {
		if compiled.Known && fixedNumberSyntax(tokens[1], atoms, compiled.Value) {
			return compiled.Value, true
		}
	}
	return 0, false
}

// lowerControlSpellSequence lowers an ordered effect sequence whose first
// effect (or second, after an initial Untap) is EffectGainControl.  It handles
// two oracle text patterns atomically:
//
//	Pattern A (effects[0] = GainControl):
//	  "Gain control of target X until end of turn. [Untap that X.] [It gains KW.] [Scry N.]"
//
//	Pattern B (effects[0] = Untap, effects[1] = GainControl, same sentence):
//	  "Untap target X and gain control of it until end of turn. [That X gains KW.]"
//
// Subsequent effects (Untap back-ref, keyword grant, counter placement, or
// standalone effects like Scry) are consumed in order.
func lowerControlSpellSequence(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported gain-control spell",
			"the executable source backend supports only exact gain-control sequences targeting one permanent",
		)
	}

	if len(ctx.content.Conditions) != 0 || len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if len(ctx.content.Targets) != 1 {
		return unsupported()
	}

	// Detect Pattern B: Untap first, GainControl second (same sentence span).
	isPatternB := len(ctx.content.Effects) >= 2 &&
		ctx.content.Effects[0].Kind == compiler.EffectUntap &&
		ctx.content.Effects[1].Kind == compiler.EffectGainControl

	gainControlIdx := 0
	if isPatternB {
		gainControlIdx = 1
	}
	controlEffect := ctx.content.Effects[gainControlIdx]

	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	// Gaining control of something you already control is a no-op for the
	// control layer, but we do allow ControllerAny (e.g. Threaten) so the
	// effect can still untap and grant keywords.
	if ctx.content.Targets[0].Selector.Controller == compiler.ControllerYou {
		return unsupported()
	}

	var duration game.EffectDuration
	switch controlEffect.Duration {
	case compiler.DurationUntilEndOfTurn:
		duration = game.DurationUntilEndOfTurn
	case compiler.DurationNone:
		duration = game.DurationPermanent
	case compiler.DurationForAsLongAsSourceOnBattlefield:
		duration = game.DurationForAsLongAsSourceOnBattlefield
	case compiler.DurationForAsLongAsYouControlSource:
		duration = game.DurationForAsLongAsYouControlSource
	default:
		return unsupported()
	}

	gainControlPrim := game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:         game.LayerControl,
			NewController: opt.Val(game.Player1),
		}},
		Duration: duration,
	}

	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	consumedTargets := 0
	// Use span-keyed sets to count each reference/keyword exactly once, even
	// when multiple same-sentence effects share the same reference spans.
	consumedRefSpans := make(map[shared.Span]bool)
	consumedKwSpans := make(map[shared.Span]bool)
	var sequence []game.Instruction

	if isPatternB {
		// Pattern B: effects[0] (Untap) and effects[1] (GainControl) share the
		// same sentence span.  Count targets and references from the shared span
		// once rather than per-effect to avoid double-counting.
		sharedSpan := ctx.content.Effects[0].Span
		consumedTargets += len(targetsWithinSpan(ctx.content.Targets, sharedSpan))
		for _, r := range referencesWithinSpan(ctx.content.References, sharedSpan) {
			consumedRefSpans[r.Span] = true
		}
		sequence = append(sequence,
			game.Instruction{Primitive: game.Untap{Object: game.TargetPermanentReference(0)}},
			game.Instruction{Primitive: gainControlPrim},
		)
		for i := 2; i < len(ctx.content.Effects); i++ {
			effAbility := contextForEffect(ctx, ctx.content.Effects[i])
			prim, ok := lowerControlSequenceFollowOn(cardName, effAbility, clauseSyntaxes[i])
			if !ok {
				return unsupported()
			}
			sequence = append(sequence, game.Instruction{Primitive: prim})
			for _, r := range effAbility.content.References {
				consumedRefSpans[r.Span] = true
			}
			for _, k := range effAbility.content.Keywords {
				consumedKwSpans[k.Span] = true
			}
		}
	} else {
		// Pattern A: effects[0] is GainControl; subsequent effects are follow-ons.
		effAbility0 := contextForEffect(ctx, ctx.content.Effects[0])
		consumedTargets += len(effAbility0.content.Targets)
		for _, r := range effAbility0.content.References {
			consumedRefSpans[r.Span] = true
		}
		sequence = append(sequence, game.Instruction{Primitive: gainControlPrim})
		for i := 1; i < len(ctx.content.Effects); i++ {
			effAbility := contextForEffect(ctx, ctx.content.Effects[i])
			prim, ok := lowerControlSequenceFollowOn(cardName, effAbility, clauseSyntaxes[i])
			if !ok {
				return unsupported()
			}
			sequence = append(sequence, game.Instruction{Primitive: prim})
			for _, r := range effAbility.content.References {
				consumedRefSpans[r.Span] = true
			}
			for _, k := range effAbility.content.Keywords {
				consumedKwSpans[k.Span] = true
			}
		}
	}

	if consumedTargets != len(ctx.content.Targets) ||
		len(consumedKwSpans) != len(ctx.content.Keywords) ||
		len(consumedRefSpans) != len(ctx.content.References) ||
		len(sequence) != len(ctx.content.Effects) {
		return unsupported()
	}

	return game.Mode{Targets: []game.TargetSpec{targetSpec}, Sequence: sequence}.Ability(), nil
}

// lowerControlSequenceFollowOn lowers a single follow-on effect in a
// gain-control sequence: an Untap back-reference, a keyword grant, a counter
// placement, or a standalone effect (e.g. Scry) with no back-references.
func lowerControlSequenceFollowOn(
	cardName string,
	ctx contentCtx,
	clauseSyntax parser.Ability,
) (game.Primitive, bool) {
	effect := ctx.content.Effects[0]

	switch effect.Kind {
	case compiler.EffectUntap:
		// Back-reference untap: "Untap that creature." — no new targets.
		if len(ctx.content.Targets) != 0 {
			return nil, false
		}
		return game.Untap{Object: game.TargetPermanentReference(0)}, true

	case compiler.EffectGain:
		// Keyword grant: "It gains haste until end of turn." — back-ref, no new targets.
		if len(ctx.content.Targets) != 0 || len(ctx.content.Keywords) == 0 {
			return nil, false
		}
		if effect.Duration != compiler.DurationUntilEndOfTurn {
			return nil, false
		}
		keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
		if !ok {
			return nil, false
		}
		return game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(0)),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				AddKeywords: keywords,
			}},
			Duration: game.DurationUntilEndOfTurn,
		}, true

	case compiler.EffectPut:
		// Counter placement: "Put a +1/+1 counter on it." — back-ref, no new targets.
		if len(ctx.content.Targets) != 0 {
			return nil, false
		}
		if !effect.CounterKindKnown || !compiler.CounterKindPlacementSupported(effect.CounterKind) {
			return nil, false
		}
		if !effect.Amount.Known || effect.Amount.Value < 1 {
			return nil, false
		}
		return game.AddCounter{
			Amount:      game.Fixed(effect.Amount.Value),
			Object:      game.TargetPermanentReference(0),
			CounterKind: effect.CounterKind,
		}, true

	default:
		// Standalone effect with no back-references (e.g. Scry).
		if len(ctx.content.References) != 0 || len(ctx.content.Targets) != 0 {
			return nil, false
		}
		content, diag := lowerSingleEffectSpell(cardName, ctx, clauseSyntax)
		if diag != nil {
			return nil, false
		}
		if len(content.SharedTargets) != 0 ||
			content.IsModal() ||
			len(content.Modes) != 1 ||
			len(content.Modes[0].Targets) != 0 ||
			len(content.Modes[0].Sequence) != 1 {
			return nil, false
		}
		return content.Modes[0].Sequence[0].Primitive, true
	}
}

// lowerSingleControlSpell lowers a single EffectGainControl spell with no
// Untap or keyword grant (e.g. "Gain control of target permanent." or the
// DurationUntilEndOfTurn variant).
func lowerSingleControlSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported gain-control spell",
			"the executable source backend supports only exact gain-control of one target permanent",
		)
	}
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Effects[0].Negated {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	if ctx.content.Targets[0].Selector.Controller == compiler.ControllerYou {
		return unsupported()
	}
	var duration game.EffectDuration
	switch ctx.content.Effects[0].Duration {
	case compiler.DurationUntilEndOfTurn:
		duration = game.DurationUntilEndOfTurn
	case compiler.DurationNone:
		duration = game.DurationPermanent
	case compiler.DurationForAsLongAsSourceOnBattlefield:
		duration = game.DurationForAsLongAsSourceOnBattlefield
	case compiler.DurationForAsLongAsYouControlSource:
		duration = game.DurationForAsLongAsYouControlSource
	default:
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:         game.LayerControl,
					NewController: opt.Val(game.Player1),
				}},
				Duration: duration,
			},
		}},
	}.Ability(), nil
}

func lowerOrderedEffectSequence(
	cardName string,
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Conditions) != 0 || len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
	}
	for _, target := range ctx.content.Targets {
		if _, ok := counterAbilityTargetSpec(target); ok {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
		}
	}
	for _, effect := range ctx.content.Effects {
		if effect.Kind == compiler.EffectSacrifice {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
		}
	}
	if content, ok := lowerTemporaryPTKeywordSpell(ctx, syntax); ok {
		return content, nil
	}
	if content, ok := lowerCyclingCountDamageAndGain(cardName, ctx); ok {
		return content, nil
	}
	if content, ok := lowerGroupLinkedLifeSpell(ctx); ok {
		return content, nil
	}
	var targets []game.TargetSpec
	var sequence []game.Instruction
	consumedTargets := 0
	consumedKeywords := 0
	consumedReferences := 0
	// oracleSpanToGameIdx maps each oracle target's Span to its first index in
	// the accumulated targets slice, recorded when the target is owned (i.e.
	// added as a new game.TargetSpec by a non-shared clause). This index is
	// looked up when an inherited shared-target clause needs to rebase its
	// sequence: the rebase offset equals the start index of the inherited
	// target rather than always 0, which is wrong when earlier effects already
	// contributed target specs before the then-joined group.
	oracleSpanToGameIdx := make(map[shared.Span]int)
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	// clauseRefSpans gives the per-clause "owned" sentence region for reference
	// and target accounting. For then-joined effects sharing a sentence Span,
	// each effect owns a distinct non-overlapping sub-region so that
	// CompiledTargets and CompiledReferences are attributed exactly once.
	// For effects NOT in a then-joined pair the region defaults to effect.Span.
	// subjectPrefixRefSpans carries the span of any subject phrase that precedes
	// the first verb in the sentence group; it is used to propagate shared
	// targets/references (e.g. "Target player" or "CardName") to implied-subject
	// clauses whose own clause span contains none.
	clauseRefSpans, subjectPrefixRefSpans := splitEffectRefSpans(syntax, ctx.content.Effects)
	for i, effect := range ctx.content.Effects {
		effectAbility := contextForEffect(ctx, effect)
		// Build the clause parser.Ability for routing through lowerAbilityContent.
		// syntaxWithinSpan always sets Text = ""; restore it from the effect text
		// for independent effects (same span), or capitalise the joined token text
		// for then-joined sub-clauses (split span) so exact-template lowerers see
		// the canonical sentence-start form.
		clauseAbility := clauseSyntaxes[i]
		if clauseAbility.Span != effect.Span {
			if clauseText := joinedTokenText(clauseAbility.Tokens); clauseText != "" {
				clauseAbility.Text = upperFirst(clauseText)
			}
		} else {
			clauseAbility.Text = effect.Text
		}
		// Per-clause target and reference scoping. Each effect owns only the
		// targets and references whose spans fall within its clause ref span.
		// Effects with an implied subject also inherit the subject-prefix targets
		// so the lowerer can match exact-text patterns ("Target creature fights…").
		// Inherited targets whose spans already appear in clauseTargets are pruned
		// to avoid duplicates (this can happen for the first effect in a group
		// whose clause ref span contains its own subject prefix).
		clauseTargets := targetsWithinSpan(ctx.content.Targets, clauseRefSpans[i])
		clauseRefs := referencesWithinSpan(ctx.content.References, clauseRefSpans[i])
		var inheritedTargets []compiler.CompiledTarget
		if subjectPrefixRefSpans[i] != (shared.Span{}) {
			for _, t := range targetsWithinSpan(ctx.content.Targets, subjectPrefixRefSpans[i]) {
				if !oracleTargetSpanIn(t.Span, clauseTargets) {
					inheritedTargets = append(inheritedTargets, t)
				}
			}
		}
		inheritedTargets = appendReferenceAntecedentTargets(
			inheritedTargets,
			clauseRefs,
			ctx.content.Targets,
			clauseTargets,
		)
		if len(clauseRefs) == 0 && subjectPrefixRefSpans[i] != (shared.Span{}) {
			clauseRefs = referencesWithinSpan(ctx.content.References, subjectPrefixRefSpans[i])
		}
		// Three target-handling modes:
		//   allSharedTargets: only inherited, no own — compound-mill "then draws".
		//   mixedTargets:     inherited + own — "then fights target creature" where
		//                     the inherited subject and a new object both appear.
		//   otherwise:        only own (or none) — normal independent effects.
		allSharedTargets := len(inheritedTargets) > 0 && len(clauseTargets) == 0
		mixedTargets := len(inheritedTargets) > 0 && len(clauseTargets) > 0
		switch {
		case allSharedTargets:
			effectAbility.content.Targets = inheritedTargets
		case mixedTargets:
			combined := make([]compiler.CompiledTarget, 0, len(inheritedTargets)+len(clauseTargets))
			combined = append(combined, inheritedTargets...)
			combined = append(combined, clauseTargets...)
			effectAbility.content.Targets = combined
		default:
			effectAbility.content.Targets = clauseTargets
		}
		effectAbility.content.References = clauseRefs
		localReferences, ok := localizeTargetReferences(
			effectAbility.content.References,
			ctx.content.Targets,
			effectAbility.content.Targets,
		)
		if !ok {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
		}
		effectAbility.content.References = localReferences
		effectAbility.content.Keywords = keywordsWithinSpan(ctx.content.Keywords, clauseRefSpans[i])
		consumedTargets += len(clauseTargets)
		consumedKeywords += len(effectAbility.content.Keywords)
		consumedReferences += len(referencesWithinSpan(ctx.content.References, clauseRefSpans[i]))
		// Lower the effect through the shared lowerAbilityContent entry point.
		// allSharedTargets: try with inherited targets; if that fails, retry
		//   with targets cleared (e.g. "then proliferate" rejects any target).
		// mixedTargets: inherited+own combined — no fallback (fail-closed).
		// default: straightforward lowering with own targets only.
		var content game.AbilityContent
		var diagnostic *shared.Diagnostic
		if linkedModify, delayedContent, ok := lowerDelayedTargetReturn(i, effectAbility, sequence); ok {
			sequence[len(sequence)-1].Primitive = linkedModify
			content = delayedContent
		} else if linkedExile, delayedContent, ok := lowerDelayedBlinkReturn(ctx.content.Effects, i, effectAbility, sequence); ok {
			sequence[len(sequence)-1].Primitive = linkedExile
			content = delayedContent
		} else if allSharedTargets {
			content, diagnostic = lowerAbilityContent(cardName, effectAbility.content, effectAbility.optional, clauseAbility)
			if diagnostic != nil {
				effectAbilityNoTarget := effectAbility
				effectAbilityNoTarget.content.Targets = nil
				content, diagnostic = lowerAbilityContent(cardName, effectAbilityNoTarget.content, effectAbilityNoTarget.optional, clauseAbility)
			}
		} else {
			content, diagnostic = lowerAbilityContent(cardName, effectAbility.content, effectAbility.optional, clauseAbility)
		}
		if diagnostic != nil ||
			len(content.SharedTargets) != 0 ||
			content.IsModal() ||
			len(content.Modes) != 1 {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
		}
		mode := content.Modes[0]
		newTargets, ok := applyTargetRemapping(
			mode, allSharedTargets, mixedTargets,
			inheritedTargets, clauseTargets,
			targets, oracleSpanToGameIdx,
		)
		if !ok {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
		}
		targets = newTargets
		sequence = append(sequence, mode.Sequence...)
	}
	if consumedTargets != len(ctx.content.Targets) ||
		consumedKeywords != len(ctx.content.Keywords) ||
		consumedReferences != len(ctx.content.References) ||
		len(sequence) != len(ctx.content.Effects) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx)
	}
	return game.Mode{Targets: targets, Sequence: sequence}.Ability(), nil
}

func localizeTargetReferences(
	references []compiler.CompiledReference,
	allTargets []compiler.CompiledTarget,
	localTargets []compiler.CompiledTarget,
) ([]compiler.CompiledReference, bool) {
	localized := append([]compiler.CompiledReference(nil), references...)
	for i := range localized {
		if localized[i].Binding != compiler.ReferenceBindingTarget {
			continue
		}
		if localized[i].Occurrence < 0 || localized[i].Occurrence >= len(allTargets) {
			return nil, false
		}
		targetSpan := allTargets[localized[i].Occurrence].Span
		local := slices.IndexFunc(localTargets, func(target compiler.CompiledTarget) bool {
			return target.Span == targetSpan
		})
		if local < 0 {
			return nil, false
		}
		localized[i].Occurrence = local
	}
	return localized, true
}

func appendReferenceAntecedentTargets(
	inherited []compiler.CompiledTarget,
	references []compiler.CompiledReference,
	allTargets []compiler.CompiledTarget,
	clauseTargets []compiler.CompiledTarget,
) []compiler.CompiledTarget {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingTarget ||
			reference.Occurrence < 0 ||
			reference.Occurrence >= len(allTargets) {
			continue
		}
		target := allTargets[reference.Occurrence]
		if !oracleTargetSpanIn(target.Span, clauseTargets) &&
			!oracleTargetSpanIn(target.Span, inherited) {
			inherited = append(inherited, target)
		}
	}
	return inherited
}

func lowerDelayedTargetReturn(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.ModifyPT, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectReturn ||
		ctx.content.Effects[0].DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		ctx.content.Effects[0].Negated ||
		ctx.text != "Return it to its owner's hand at the beginning of the next end step." ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	previous := sequence[effectIndex-1].Primitive
	if previous.Kind() != game.PrimitiveModifyPT {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify, ok := previous.(game.ModifyPT)
	if !ok ||
		modify.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		modify.PublishLinked != "" {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-target-%d", effectIndex))
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify.PublishLinked = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Bounce{
			Object: object,
		}}}}.Ability(),
	}}
	return modify, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

func lowerDelayedBlinkReturn(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Exile, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		effects[effectIndex-1].Kind != compiler.EffectExile ||
		effects[effectIndex-1].DelayedTiming != 0 ||
		len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectReturn ||
		ctx.content.Effects[0].DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		ctx.content.Effects[0].Negated ||
		(ctx.text != "Return it to the battlefield under its owner's control at the beginning of the next end step." &&
			ctx.text != "Return that card to the battlefield under its owner's control at the beginning of the next end step.") {
		return game.Exile{}, game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, effectIndex-1) {
		return game.Exile{}, game.AbilityContent{}, false
	}
	// References validated — clear before fail-closed check.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile, ok := sequence[effectIndex-1].Primitive.(game.Exile)
	if !ok ||
		exile.Group.Valid() ||
		exile.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		exile.ExileLinkedKey != "" {
		return game.Exile{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-blink-%d", effectIndex))
	if _, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		PriorInstruction: effectIndex - 1,
		PriorLinkedKey:   key,
	}); !ok {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile.ExileLinkedKey = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource(key),
		}}}}.Ability(),
	}}
	return exile, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// joinedTokenText reconstructs the source text from a token slice, inserting
// spaces between tokens where appropriate (following oracle punctuation rules).
// This mirrors the unexported compiler.joinedSourceText function.
func joinedTokenText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 && joinedTokenNeedsSpace(tokens[i-1], tok) { //nolint:gosec // i>0 guarantees valid index
			_ = b.WriteByte(' ')
		}
		_, _ = b.WriteString(tok.Text)
	}
	return b.String()
}

// upperFirst returns s with its first byte uppercased. It is safe for ASCII
// oracle text where the first character is always a plain letter.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// sharedTargetRebaseOffset returns the accumulated-targets start index for the
// first inherited target oracle span, by looking it up in oracleSpanToGameIdx.
// The offset is used to rebase the sequence of an inherited shared-target
// clause (e.g. the "then draws" in "mills …, then draws …") so that its
// local target index 0 maps to the correct position in the already-accumulated
// targets slice, even when an earlier unrelated effect already contributed
// target specs at indices 0, 1, etc.
//
// Returns (0, false) if inherited is empty or the first span has no entry in
// the map (caller should treat this as fail-closed).
// oracleTargetSpanIn reports whether any target in list has the given span.
func oracleTargetSpanIn(span shared.Span, list []compiler.CompiledTarget) bool {
	for _, t := range list {
		if t.Span == span {
			return true
		}
	}
	return false
}

func sharedTargetRebaseOffset(inherited []compiler.CompiledTarget, spanToIdx map[shared.Span]int) (int, bool) {
	if len(inherited) == 0 {
		return 0, false
	}
	idx, ok := spanToIdx[inherited[0].Span]
	return idx, ok
}

// applyTargetRemapping sequences mode's target references to the correct
// accumulated game indices and updates the targets slice and oracleSpanToGameIdx
// map accordingly. It handles three cases:
//   - allSharedTargets: uniform rebase to the inherited target's recorded index.
//   - mixedTargets: non-uniform per-local-index remap for inherited+owned targets.
//   - default: uniform rebase starting at len(accum).
//
// Returns the updated accum slice (false if any remapping step fails).
func applyTargetRemapping(
	mode game.Mode,
	allSharedTargets, mixedTargets bool,
	inherited, owned []compiler.CompiledTarget,
	accum []game.TargetSpec,
	spanToIdx map[shared.Span]int,
) ([]game.TargetSpec, bool) {
	m := mode
	switch {
	case len(m.Targets) > 0 && allSharedTargets:
		rebaseOffset, ok := sharedTargetRebaseOffset(inherited, spanToIdx)
		if !ok || !rebaseTargetedSequence(m.Sequence, rebaseOffset) {
			return nil, false
		}
	case len(m.Targets) > 0 && mixedTargets:
		if len(m.Targets) != len(inherited)+len(owned) {
			return nil, false
		}
		localToGame := make([]int, len(m.Targets))
		for j, t := range inherited {
			idx, ok := spanToIdx[t.Span]
			if !ok {
				return nil, false
			}
			localToGame[j] = idx
		}
		gameStartForOwn := len(accum)
		for j, ot := range owned {
			localToGame[len(inherited)+j] = gameStartForOwn + j
			spanToIdx[ot.Span] = gameStartForOwn + j
		}
		if !remapTargetedSequence(m.Sequence, localToGame) {
			return nil, false
		}
		accum = append(accum, m.Targets[len(inherited):]...)
	case len(m.Targets) > 0:
		gameStartIdx := len(accum)
		if !rebaseTargetedSequence(m.Sequence, gameStartIdx) {
			return nil, false
		}
		for j, ot := range owned {
			if j < len(m.Targets) {
				spanToIdx[ot.Span] = gameStartIdx + j
			}
		}
		accum = append(accum, m.Targets...)
	default:
	}
	return accum, true
}

func joinedTokenNeedsSpace(prev, cur shared.Token) bool {
	if cur.Kind == shared.Comma || cur.Kind == shared.Period || cur.Kind == shared.Colon ||
		cur.Kind == shared.Semicolon || cur.Kind == shared.RightParen ||
		cur.Kind == shared.Apostrophe || prev.Kind == shared.Apostrophe ||
		prev.Kind == shared.LeftParen || prev.Kind == shared.Quote || cur.Kind == shared.Quote {
		return false
	}
	if prev.Kind == shared.Plus || prev.Kind == shared.Minus || prev.Kind == shared.Slash ||
		cur.Kind == shared.Plus || cur.Kind == shared.Minus || cur.Kind == shared.Slash ||
		prev.Kind == shared.Asterisk || cur.Kind == shared.Asterisk {
		return false
	}
	return true
}

func lowerCyclingCountDamageAndGain(cardName string, ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectDealDamage ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[1].Negated ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!singleSelfReference(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	amountEffect := ctx.content.Effects[1].Amount
	if amountEffect.DynamicKind == compiler.DynamicAmountNone ||
		amountEffect.DynamicForm != compiler.DynamicAmountWhereX {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(amountEffect, game.SourcePermanentReference())
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	if ctx.text != fmt.Sprintf(
		"%s deals X damage to %s and you gain X life, %s.",
		cardName,
		ctx.content.Targets[0].Text,
		amountEffect.Text,
	) {
		return game.AbilityContent{}, false
	}
	amount := game.Dynamic(dynamic)
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{Primitive: game.Damage{
				Amount:    amount,
				Recipient: game.AnyTargetDamageRecipient(0),
			}},
			{Primitive: game.GainLife{
				Amount: amount,
				Player: game.ControllerReference(),
			}},
		},
	}.Ability(), true
}

// lowerGroupLinkedLifeSpell handles linked two-effect patterns of the form
// "Each opponent loses N life and you gain [N | that much] life."
// It emits LoseLife with PublishResult "life-change" followed by GainLife.
// For "that much", the GainLife amount uses DynamicAmountPreviousEffectResult.
func lowerGroupLinkedLifeSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectLose ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[1].Negated ||
		!ctx.content.Effects[0].Amount.Known ||
		ctx.content.Effects[0].Amount.Value < 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	loseAmount := game.Fixed(ctx.content.Effects[0].Amount.Value)
	amountText := fmt.Sprint(ctx.content.Effects[0].Amount.Value)

	// Determine player group from full sentence text.
	var group game.PlayerGroupReference
	switch {
	case ctx.text == fmt.Sprintf("Each opponent loses %s life and you gain %s life.", amountText, amountText) ||
		ctx.text == fmt.Sprintf("Each opponent loses %s life and you gain that much life.", amountText):
		group = game.OpponentsReference()
	default:
		return game.AbilityContent{}, false
	}

	// Determine the gain amount: fixed if effects[1] has a known value, dynamic ("that much") otherwise.
	var gainAmount game.Quantity
	switch {
	case ctx.content.Effects[1].Amount.Known && ctx.content.Effects[1].Amount.Value > 0:
		gainAmount = game.Fixed(ctx.content.Effects[1].Amount.Value)
	case !ctx.content.Effects[1].Amount.Known:
		gainAmount = game.Dynamic(game.DynamicAmount{
			Kind:      game.DynamicAmountPreviousEffectResult,
			ResultKey: "life-change",
		})
	default:
		return game.AbilityContent{}, false
	}

	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive:     game.LoseLife{PlayerGroup: group, Amount: loseAmount},
				PublishResult: "life-change",
			},
			{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: gainAmount},
			},
		},
	}.Ability(), true
}

// remapping to all target references in sequence. Unlike rebaseTargetedSequence
// which adds a uniform offset, this function looks up each local target index
// in localToGame and replaces it with the corresponding accumulated game index.
// This is needed for mixed inherited+owned target clauses where inherited
// targets live at their original accumulated indices while newly-owned targets
// start at a later position.
func remapTargetedSequence(sequence []game.Instruction, localToGame []int) bool {
	for i := range sequence {
		primitive, ok := remapTargetedPrimitive(sequence[i].Primitive, localToGame)
		if !ok {
			return false
		}
		sequence[i].Primitive = primitive
	}
	return true
}

func remapTargetedPrimitive(primitive game.Primitive, localToGame []int) (game.Primitive, bool) {
	// Explicit allowlist — same set as rebaseTargetedPrimitive.
	if value, ok := primitive.(game.Damage); ok {
		recipient, ok := remapDamageRecipient(value.Recipient, localToGame)
		if !ok {
			return nil, false
		}
		value.Recipient = recipient
		if value.DamageSource.Exists {
			source, ok := remapObjectReference(value.DamageSource.Val, localToGame)
			if !ok {
				return nil, false
			}
			value.DamageSource = opt.Val(source)
		}
		return value, true
	}
	if value, ok := primitive.(game.Destroy); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.AddCounter); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.AddPlayerCounter); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.ModifyPT); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Fight); ok {
		var ok bool
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		if !ok {
			return nil, false
		}
		value.RelatedObject, ok = remapObjectReference(value.RelatedObject, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Tap); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Untap); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Exile); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Bounce); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.CounterObject); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Regenerate); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Draw); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Discard); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Mill); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.GainLife); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.LoseLife); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.CreateDelayedTrigger); ok {
		return value, true
	}
	return nil, false
}

func remapDamageRecipient(recipient game.DamageRecipient, localToGame []int) (game.DamageRecipient, bool) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		idx := object.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.DamageRecipient{}, false
		}
		return game.AnyTargetDamageRecipient(localToGame[idx]), true
	}
	if object, ok := recipient.ObjectReference(); ok {
		remapped, valid := remapObjectReference(object, localToGame)
		return game.ObjectDamageRecipient(remapped), valid
	}
	if player, ok := recipient.PlayerReference(); ok {
		remapped, valid := remapPlayerReference(player, localToGame)
		return game.PlayerDamageRecipient(remapped), valid
	}
	return game.DamageRecipient{}, false
}

func remapObjectReference(reference game.ObjectReference, localToGame []int) (game.ObjectReference, bool) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.ObjectReference{}, false
		}
		return game.TargetPermanentReference(localToGame[idx]), true
	case game.ObjectReferenceTargetStackObject:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.ObjectReference{}, false
		}
		return game.TargetStackObjectReference(localToGame[idx]), true
	case game.ObjectReferenceTargetAttachedPermanent:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.ObjectReference{}, false
		}
		return game.TargetAttachedPermanentReference(localToGame[idx]), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func remapPlayerReference(reference game.PlayerReference, localToGame []int) (game.PlayerReference, bool) {
	switch reference.Kind() {
	case game.PlayerReferenceTargetPlayer:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.PlayerReference{}, false
		}
		return game.TargetPlayerReference(localToGame[idx]), true
	case game.PlayerReferenceObjectController, game.PlayerReferenceObjectOwner:
		object, ok := reference.Object()
		if !ok {
			return game.PlayerReference{}, false
		}
		object, ok = remapObjectReference(object, localToGame)
		if !ok {
			return game.PlayerReference{}, false
		}
		if reference.Kind() == game.PlayerReferenceObjectController {
			return game.ObjectControllerReference(object), true
		}
		return game.ObjectOwnerReference(object), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func rebaseTargetedSequence(sequence []game.Instruction, offset int) bool {
	for i := range sequence {
		primitive, ok := rebaseTargetedPrimitive(sequence[i].Primitive, offset)
		if !ok {
			return false
		}
		sequence[i].Primitive = primitive
	}
	return true
}

func rebaseTargetedPrimitive(primitive game.Primitive, offset int) (game.Primitive, bool) {
	// Keep this as an explicit allowlist so a new target-bearing primitive cannot
	// silently retain a clause-local target index.
	if value, ok := primitive.(game.Damage); ok {
		recipient, ok := rebaseDamageRecipient(value.Recipient, offset)
		if !ok {
			return nil, false
		}
		value.Recipient = recipient
		if value.DamageSource.Exists {
			source, ok := rebaseObjectReference(value.DamageSource.Val, offset)
			if !ok {
				return nil, false
			}
			value.DamageSource = opt.Val(source)
		}
		return value, true
	}
	if value, ok := primitive.(game.Destroy); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.AddCounter); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.AddPlayerCounter); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.ModifyPT); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Fight); ok {
		var ok bool
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		if !ok {
			return nil, false
		}
		value.RelatedObject, ok = rebaseObjectReference(value.RelatedObject, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Tap); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Untap); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Exile); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Bounce); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.CounterObject); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Regenerate); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Draw); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Discard); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Mill); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.GainLife); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.LoseLife); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.CreateDelayedTrigger); ok {
		return value, true
	}
	if value, ok := primitive.(game.ApplyContinuous); ok {
		if value.Object.Exists {
			rebased, ok := rebaseObjectReference(value.Object.Val, offset)
			if !ok {
				return nil, false
			}
			value.Object = opt.Val(rebased)
		}
		return value, true
	}
	return nil, false
}

func rebaseDamageRecipient(recipient game.DamageRecipient, offset int) (game.DamageRecipient, bool) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		return game.AnyTargetDamageRecipient(object.TargetIndex() + offset), true
	}
	if object, ok := recipient.ObjectReference(); ok {
		rebased, valid := rebaseObjectReference(object, offset)
		return game.ObjectDamageRecipient(rebased), valid
	}
	if player, ok := recipient.PlayerReference(); ok {
		rebased, valid := rebasePlayerReference(player, offset)
		return game.PlayerDamageRecipient(rebased), valid
	}
	return game.DamageRecipient{}, false
}

func rebaseObjectReference(reference game.ObjectReference, offset int) (game.ObjectReference, bool) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		return game.TargetPermanentReference(reference.TargetIndex() + offset), true
	case game.ObjectReferenceTargetStackObject:
		return game.TargetStackObjectReference(reference.TargetIndex() + offset), true
	case game.ObjectReferenceTargetAttachedPermanent:
		return game.TargetAttachedPermanentReference(reference.TargetIndex() + offset), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func rebasePlayerReference(reference game.PlayerReference, offset int) (game.PlayerReference, bool) {
	switch reference.Kind() {
	case game.PlayerReferenceTargetPlayer:
		return game.TargetPlayerReference(reference.TargetIndex() + offset), true
	case game.PlayerReferenceObjectController, game.PlayerReferenceObjectOwner:
		object, ok := reference.Object()
		if !ok {
			return game.PlayerReference{}, false
		}
		object, ok = rebaseObjectReference(object, offset)
		if !ok {
			return game.PlayerReference{}, false
		}
		if reference.Kind() == game.PlayerReferenceObjectController {
			return game.ObjectControllerReference(object), true
		}
		return game.ObjectOwnerReference(object), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func contextForEffect(
	ctx contentCtx,
	effect compiler.CompiledEffect,
) contentCtx {
	ctx.text = effect.Text
	ctx.span = effect.Span
	ctx.content.Effects = []compiler.CompiledEffect{effect}
	ctx.content.Targets = targetsWithinSpan(ctx.content.Targets, effect.Span)
	ctx.content.Keywords = keywordsWithinSpan(ctx.content.Keywords, effect.Span)
	ctx.content.References = referencesWithinSpan(ctx.content.References, effect.Span)
	return ctx
}

func targetsWithinSpan(targets []compiler.CompiledTarget, span shared.Span) []compiler.CompiledTarget {
	var within []compiler.CompiledTarget
	for _, target := range targets {
		if spanCovered(target.Span, []shared.Span{span}) {
			within = append(within, target)
		}
	}
	return within
}

func keywordsWithinSpan(keywords []compiler.CompiledKeyword, span shared.Span) []compiler.CompiledKeyword {
	var within []compiler.CompiledKeyword
	for _, keyword := range keywords {
		if spanCovered(keyword.Span, []shared.Span{span}) {
			within = append(within, keyword)
		}
	}
	return within
}

func referencesWithinSpan(references []compiler.CompiledReference, span shared.Span) []compiler.CompiledReference {
	var within []compiler.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, []shared.Span{span}) {
			within = append(within, reference)
		}
	}
	return within
}

func syntaxWithinSpan(syntax parser.Ability, span shared.Span) parser.Ability {
	syntax.Span = span
	syntax.Text = ""
	syntax.Tokens = slices.DeleteFunc(
		append([]shared.Token(nil), syntax.Tokens...),
		func(token shared.Token) bool {
			return !spanCovered(token.Span, []shared.Span{span})
		},
	)
	return syntax
}

// splitEffectSyntaxes returns per-clause syntax for each effect in an ordered
// sequence. For effects sharing the same sentence Span, the entire same-span
// group is processed in a single pass so that each clause's subject ownership
// is stable and never overwritten by a later pair.
//
// For each clause k in a then-joined group of n effects:
//   - Subject tokens:
//     k == 0: tokens[sentenceStart .. verb[0]] (any subject phrase before the first verb).
//     k  > 0: tokens immediately after the preceding "then" up to verb[k].
//     If that post-then region is non-empty it is the explicit subject
//     (e.g. "you" in "then you gain 2 life.").
//     If it is empty the subject is implied; it is inherited from the first
//     clause when the verb form ends in 's' (third-person singular, e.g.
//     "draws", "mills"), indicating the same grammatical subject continues.
//     Otherwise (imperative/controller verb, e.g. "draw", "proliferate")
//     no subject prefix is prepended.
//   - Verb clause tokens: from verb[k] to just before the next comma-then
//     connector for non-final clauses, or through the sentence period for the
//     final clause.
//   - The terminal period is appended to each non-final clause.
//
// The function is fail-closed: any invalid boundary for any pair in the group
// causes the entire group to fall back to syntaxWithinSpan(syntax, effect.Span).
func splitEffectSyntaxes(syntax parser.Ability, effects []compiler.CompiledEffect) []parser.Ability {
	clauses := make([]parser.Ability, len(effects))
	for i, effect := range effects {
		clauses[i] = syntaxWithinSpan(syntax, effect.Span)
	}
	tokens := syntax.Tokens

	// Process each same-span then-joined group in one pass. Groups are
	// contiguous runs of effects that share the same sentence Span.
	for i := 0; i < len(effects); {
		sentenceSpan := effects[i].Span
		j := i + 1
		for j < len(effects) && effects[j].Span == sentenceSpan {
			j++
		}
		n := j - i
		if n < 2 {
			i = j
			continue
		}

		// Find sentenceStart: first token index within the sentence span.
		sentenceStart := -1
		for k, tok := range tokens {
			if spanCovered(tok.Span, []shared.Span{sentenceSpan}) {
				sentenceStart = k
				break
			}
		}
		if sentenceStart < 0 {
			i = j
			continue
		}

		// Find terminal period for the sentence.
		period, hasPeriod := lastPeriodTokenInSpan(tokens, sentenceSpan)
		if !hasPeriod {
			i = j
			continue
		}
		periodIdx := -1
		for k := len(tokens) - 1; k >= sentenceStart; k-- {
			if tokens[k].Span == period.Span {
				periodIdx = k
				break
			}
		}
		if periodIdx < 0 {
			i = j
			continue
		}

		// Collect verb token indices for each effect in the group.
		verbs := make([]int, n)
		valid := true
		for k := range n {
			v := findVerbTokenIndex(tokens, effects[i+k].VerbSpan)
			if v < 0 || v < sentenceStart {
				valid = false
				break
			}
			verbs[k] = v
		}
		if !valid {
			i = j
			continue
		}

		// Collect "then" positions and clause-end positions for each pair.
		thens := make([]int, n-1)
		ends := make([]int, n-1) // tokens[ends[k]] is the first token NOT in clause k
		for k := 0; k < n-1; k++ {
			thenIdx := -1
			for m := verbs[k] + 1; m < verbs[k+1]; m++ {
				if tokens[m].Kind == shared.Word && strings.EqualFold(tokens[m].Text, "then") {
					thenIdx = m
					break
				}
			}
			if thenIdx < 0 {
				valid = false
				break
			}
			thens[k] = thenIdx
			end := thenIdx
			if end > sentenceStart && tokens[end-1].Kind == shared.Comma {
				end--
			}
			if end <= sentenceStart {
				valid = false
				break
			}
			ends[k] = end
		}
		if !valid {
			i = j
			continue
		}

		// Compute the subject token slice for each clause.
		// Clause 0: tokens[sentenceStart:verbs[0]] — any sentence-opening subject.
		// Clause k>0: post-then pre-verb tokens; if empty, either inherit the
		// first clause's subject (third-person 's' verb) or use no prefix.
		firstSubject := append([]shared.Token(nil), tokens[sentenceStart:verbs[0]]...)
		subjects := make([][]shared.Token, n)
		subjects[0] = firstSubject
		for k := 1; k < n; k++ {
			postThen := tokens[thens[k-1]+1 : verbs[k]]
			switch {
			case len(postThen) > 0:
				// Explicit subject in the post-then region (e.g. "you", "Test Bolt").
				subjects[k] = append([]shared.Token(nil), postThen...)
			case len(firstSubject) > 0 && verbImpliesInheritedSubject(tokens[verbs[k]]):
				// Implied subject: verb is third-person ('s'-ending) and the
				// first clause has a subject prefix — inherit it.
				// Example: "Target player mills …, then draws …"
				subjects[k] = firstSubject
			default:
				// Implied subject with controller verb (imperative, no 's'):
				// e.g. "then draw a card" or "then proliferate".
				subjects[k] = nil
			}
		}

		// Build clause tokens and spans for each effect in the group.
		for k := range n {
			var clauseTokens []shared.Token
			clauseTokens = append(clauseTokens, subjects[k]...)

			if k < n-1 {
				clauseTokens = append(clauseTokens, tokens[verbs[k]:ends[k]]...)
				clauseTokens = append(clauseTokens, period)
			} else {
				clauseTokens = append(clauseTokens, tokens[verbs[k]:periodIdx+1]...)
			}

			// Span.Start: use sentenceStart for the first clause (to cover the
			// subject phrase), verbs[k].Start for subsequent clauses (ensuring
			// clause.Span != sentence.Span even when subject tokens are prepended
			// from the sentence start).
			var spanStart shared.Position
			if k == 0 {
				spanStart = tokens[sentenceStart].Span.Start
			} else {
				spanStart = tokens[verbs[k]].Span.Start
			}
			var spanEnd shared.Position
			if k < n-1 {
				spanEnd = tokens[ends[k]-1].Span.End
			} else {
				spanEnd = period.Span.End
			}

			clauses[i+k] = parser.Ability{
				Span:      shared.Span{Start: spanStart, End: spanEnd},
				Tokens:    clauseTokens,
				Reminders: syntax.Reminders,
				Atoms:     syntax.Atoms,
			}
		}

		i = j
	}
	return clauses
}

// verbImpliesInheritedSubject reports whether a verb token uses the third-
// person singular form (ends in 's', e.g. "draws", "mills", "discards").
// When a then-joined clause has no explicit post-then subject and the verb
// ends in 's', the first clause's subject prefix is inherited (e.g. "Target
// player mills …, then draws …"). An imperative verb ("draw", "mill",
// "proliferate") receives no subject prefix and is lowered as a controller
// action.
func verbImpliesInheritedSubject(tok shared.Token) bool {
	return strings.HasSuffix(strings.ToLower(tok.Text), "s")
}

// splitEffectRefSpans returns two parallel slices keyed by effect index.
// It uses the same single-pass group strategy as splitEffectSyntaxes.
//
//   - clauseRefSpans: the "owned" sentence region for reference and target
//     accounting. For a then-joined group, effect k's region is:
//     k == 0: sentenceStart .. just before the first comma-then connector
//     k  > 0: immediately after the preceding "then" .. just before the next
//     comma-then connector (or sentence end for the final effect)
//     This partitions the sentence so every CompiledTarget/Reference is
//     attributed to exactly one clause without overlap.
//
//   - subjectPrefixRefSpans: the span of the first-clause subject phrase
//     ({sentenceStart..before verb[0]}). Set only for clauses whose
//     clauseRefSpan will contain no targets/references but whose implied
//     subject carries shared ones — specifically, when the post-then pre-verb
//     region is empty AND the verb implies inheritance (third-person 's').
//     Callers use this to propagate subject-carried targets/references to the
//     lowerer without double-counting them in the accounting totals.
func splitEffectRefSpans(syntax parser.Ability, effects []compiler.CompiledEffect) (clauseRefSpans, subjectPrefixRefSpans []shared.Span) {
	clauseRefSpans = make([]shared.Span, len(effects))
	subjectPrefixRefSpans = make([]shared.Span, len(effects))
	for i, effect := range effects {
		clauseRefSpans[i] = effect.Span
	}
	tokens := syntax.Tokens

	for i := 0; i < len(effects); {
		sentenceSpan := effects[i].Span
		j := i + 1
		for j < len(effects) && effects[j].Span == sentenceSpan {
			j++
		}
		n := j - i
		if n < 2 {
			i = j
			continue
		}

		sentenceStart := -1
		for k, tok := range tokens {
			if spanCovered(tok.Span, []shared.Span{sentenceSpan}) {
				sentenceStart = k
				break
			}
		}
		if sentenceStart < 0 {
			i = j
			continue
		}

		verbs := make([]int, n)
		valid := true
		for k := range n {
			v := findVerbTokenIndex(tokens, effects[i+k].VerbSpan)
			if v < 0 {
				valid = false
				break
			}
			verbs[k] = v
		}
		if !valid {
			i = j
			continue
		}

		thens := make([]int, n-1)
		ends := make([]int, n-1)
		for k := 0; k < n-1; k++ {
			thenIdx := -1
			for m := verbs[k] + 1; m < verbs[k+1]; m++ {
				if tokens[m].Kind == shared.Word && strings.EqualFold(tokens[m].Text, "then") {
					thenIdx = m
					break
				}
			}
			if thenIdx < 0 {
				valid = false
				break
			}
			thens[k] = thenIdx
			end := thenIdx
			if end > sentenceStart && tokens[end-1].Kind == shared.Comma {
				end--
			}
			if end <= sentenceStart {
				valid = false
				break
			}
			ends[k] = end
		}
		if !valid {
			i = j
			continue
		}

		// Clause 0 ref span: from sentenceStart through just before the first then.
		clauseRefSpans[i] = shared.Span{
			Start: tokens[sentenceStart].Span.Start,
			End:   tokens[ends[0]-1].Span.End,
		}
		// Clause k (1..n-1) ref span: from immediately after the preceding "then"
		// through just before the next comma-then (or sentence end for the last).
		for k := 1; k < n; k++ {
			if thens[k-1]+1 >= len(tokens) {
				valid = false
				break
			}
			var end shared.Position
			if k < n-1 {
				end = tokens[ends[k]-1].Span.End
			} else {
				end = sentenceSpan.End
			}
			clauseRefSpans[i+k] = shared.Span{
				Start: tokens[thens[k-1]+1].Span.Start,
				End:   end,
			}
		}
		if !valid {
			i = j
			continue
		}

		// Subject prefix span: tokens[sentenceStart..verb[0]-1].
		// Set for all effects in the group that use implied-subject inheritance,
		// i.e. those whose post-then pre-verb region is empty AND whose verb
		// implies the same subject continues (third-person 's').
		if verbs[0] > sentenceStart {
			subjectSpan := shared.Span{
				Start: tokens[sentenceStart].Span.Start,
				End:   tokens[verbs[0]-1].Span.End,
			}
			// Clause 0 always gets the subject prefix span.
			subjectPrefixRefSpans[i] = subjectSpan
			// Subsequent clauses get it only when implied-subject inheritance applies.
			for k := 1; k < n; k++ {
				postThen := tokens[thens[k-1]+1 : verbs[k]]
				if len(postThen) == 0 && verbImpliesInheritedSubject(tokens[verbs[k]]) {
					subjectPrefixRefSpans[i+k] = subjectSpan
				}
			}
		}

		i = j
	}
	return clauseRefSpans, subjectPrefixRefSpans
}

// lastPeriodTokenInSpan returns the last Period token in tokens whose own span
// lies within span, or the zero Token and false if none is found.
func lastPeriodTokenInSpan(tokens []shared.Token, span shared.Span) (shared.Token, bool) {
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i].Kind == shared.Period && spanCovered(tokens[i].Span, []shared.Span{span}) {
			return tokens[i], true
		}
	}
	return shared.Token{}, false
}

// findVerbTokenIndex returns the index in tokens of the token whose span start
// matches verbSpan.Start, or -1 if not found.
func findVerbTokenIndex(tokens []shared.Token, verbSpan shared.Span) int {
	for i, token := range tokens {
		if token.Span.Start.Offset == verbSpan.Start.Offset {
			return i
		}
	}
	return -1
}

func unsupportedEffectSequenceDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported ordered effect sequence",
		"the executable source backend supports only exact ordered sequences of independently supported effects",
	)
}

func lowerGroupDamageSpell(
	cardName string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	damageSource, ok := lowerDamageSourceReference(ctx.content.References)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	amountText := fmt.Sprint(effect.Amount.Value)
	// When the source is bound to the triggering permanent, the body may use
	// "It deals" rather than the card name as the damage source subject.
	textSourceSubject := cardName
	if len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent &&
		strings.HasPrefix(ctx.text, "It deals ") {
		textSourceSubject = "It"
	}
	sel := effect.Selector
	var recipient game.DamageRecipient
	switch {
	case sel.Kind == compiler.SelectorOpponent && !sel.Other:
		if ctx.text != fmt.Sprintf("%s deals %s damage to each opponent.", textSourceSubject, amountText) {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact fixed group damage amounts",
			)
		}
		recipient = game.PlayerGroupDamageRecipient(game.OpponentsReference())
	case sel.Kind == compiler.SelectorPlayer && !sel.Other:
		if ctx.text != fmt.Sprintf("%s deals %s damage to each player.", textSourceSubject, amountText) {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact fixed group damage amounts",
			)
		}
		recipient = game.PlayerGroupDamageRecipient(game.AllPlayersReference())
	case sel.Kind == compiler.SelectorCreature && !sel.Other:
		if ctx.text != fmt.Sprintf("%s deals %s damage to each creature.", textSourceSubject, amountText) {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact fixed group damage amounts",
			)
		}
		recipient = game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}))
	case sel.Kind == compiler.SelectorCreature && sel.Other:
		if ctx.text != fmt.Sprintf("%s deals %s damage to each other creature.", textSourceSubject, amountText) {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact fixed group damage amounts",
			)
		}
		recipient = game.GroupDamageRecipient(game.BattlefieldGroupExcluding(
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.SourcePermanentReference(),
		))
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend does not support this group recipient",
		)
	}
	damage := game.Damage{
		Amount:    game.Fixed(effect.Amount.Value),
		Recipient: recipient,
	}
	if damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: damage,
			},
		},
	}.Ability(), nil
}

func lowerFixedDamageSpell(
	cardName string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		(effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	amountText := "X"
	var damageSource game.ObjectReference
	var sourceBound bool
	if len(ctx.content.References) > 0 {
		damageSource, sourceBound = lowerDamageSourceReference(ctx.content.References[:1])
	}
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
		amountText = fmt.Sprint(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		amountObject := game.SourcePermanentReference()
		if sourceBound {
			amountObject = damageSource
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	// When the source is the triggering permanent, the body may use "It deals"
	// rather than the card name. Accept that pronoun only when the source
	// reference is bound to the triggering event.
	textSourceSubject := cardName
	if len(ctx.content.References) > 0 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent &&
		strings.HasPrefix(ctx.text, "It deals ") {
		textSourceSubject = "It"
	}
	if !ok ||
		!exactDamageAmountSyntax(textSourceSubject, ctx, effect.Amount, amountText) ||
		!exactDamageAmountReferences(effect.Amount, ctx.content.References) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	damage := game.Damage{
		Amount:    amount,
		Recipient: game.AnyTargetDamageRecipient(0),
	}
	if sourceBound && damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{
				Primitive: damage,
			},
		},
	}.Ability(), nil
}

func lowerFixedModifyPTSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.StaticSubject != compiler.StaticSubjectNone {
		return lowerFixedGroupModifyPTSpell(ctx, syntax, effect)
	}
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent {
		return lowerEventPermanentFixedModifyPT(ctx)
	}
	dynamicPT := effect.Amount.DynamicKind != compiler.DynamicAmountNone
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Targets[0].Selector.Kind != compiler.SelectorCreature ||
		(!dynamicPT && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!exactModifyPTAmountSyntax(ctx, effect) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	powerDelta := game.Fixed(compiledSignedAmountValue(effect.PowerDelta))
	toughnessDelta := game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta))
	if dynamicPT {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
		switch effect.Amount.DynamicForm {
		case compiler.DynamicAmountWhereX:
			powerDelta = game.Dynamic(dynamic)
			toughnessDelta = game.Dynamic(dynamic)
		case compiler.DynamicAmountForEach:
			powerDelta = dynamicSignedQuantity(dynamic, effect.PowerDelta)
			toughnessDelta = dynamicSignedQuantity(dynamic, effect.ToughnessDelta)
		default:
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ModifyPT{
					Object:         game.TargetPermanentReference(0),
					PowerDelta:     powerDelta,
					ToughnessDelta: toughnessDelta,
					Duration:       game.DurationUntilEndOfTurn,
				},
			},
		},
	}.Ability(), nil
}

// lowerEventPermanentFixedModifyPT lowers an exact fixed until-end-of-turn
// ModifyPT body whose sole non-target subject reference is
// ReferenceBindingEventPermanent. The text must be exactly
// "It gets <power>/<toughness> until end of turn." The object lowers to
// game.EventPermanentReference(), which identifies the permanent named by the
// triggering event.
func lowerEventPermanentFixedModifyPT(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed until-end-of-turn power/toughness changes to the triggering permanent",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	want := fmt.Sprintf("It gets %s/%s until end of turn.",
		signedAmountText(effect.PowerDelta),
		signedAmountText(effect.ToughnessDelta))
	if ctx.text != want {
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         object,
				PowerDelta:     game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
				ToughnessDelta: game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)),
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

func lowerFixedGroupModifyPTSpell(
	ctx contentCtx,
	syntax parser.Ability,
	effect compiler.CompiledEffect,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported group power/toughness spell",
			"the executable source backend supports only exact fixed supported group power/toughness changes until end of turn",
		)
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		!matchesExactTemporaryGroupPTSyntax(syntax, effect) {
		return unsupported()
	}
	group, ok := resolvingStaticSubjectGroup(effect)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:          game.LayerPowerToughnessModify,
					Group:          group,
					PowerDelta:     compiledSignedAmountValue(effect.PowerDelta),
					ToughnessDelta: compiledSignedAmountValue(effect.ToughnessDelta),
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

func lowerTemporaryKeywordSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported temporary keyword spell",
			"the executable source backend supports only exact non-parameterized keyword grants to one target creature or permanent until end of turn",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Negated ||
		effect.StaticSubject != compiler.StaticSubjectNone ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		!temporaryKeywordTarget(ctx.content.Targets[0]) ||
		!matchesExactTemporaryKeywordSyntax(syntax, ctx.content.Targets[0], ctx.content.Keywords) {
		return unsupported()
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					AddKeywords: keywords,
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

func lowerTemporaryPTKeywordSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectModifyPT ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!temporaryKeywordTarget(ctx.content.Targets[0]) {
		return game.AbilityContent{}, false
	}
	modifyEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if modifyEffect.Span != keywordEffect.Span ||
		modifyEffect.Negated ||
		keywordEffect.Negated ||
		modifyEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		modifyEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		modifyEffect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!modifyEffect.PowerDelta.Known ||
		!modifyEffect.ToughnessDelta.Known ||
		!matchesExactTemporaryPTKeywordSyntax(syntax, ctx.content.Targets[0], modifyEffect, ctx.content.Keywords) {
		return game.AbilityContent{}, false
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:          game.LayerPowerToughnessModify,
						PowerDelta:     compiledSignedAmountValue(modifyEffect.PowerDelta),
						ToughnessDelta: compiledSignedAmountValue(modifyEffect.ToughnessDelta),
					},
					{
						Layer:       game.LayerAbility,
						AddKeywords: keywords,
					},
				},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

func temporaryKeywordTarget(target compiler.CompiledTarget) bool {
	return target.Selector.Kind == compiler.SelectorCreature ||
		target.Selector.Kind == compiler.SelectorPermanent
}

func lowerFixedBounceSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	target := ctx.content.Targets[0]
	target.Text = strings.TrimSuffix(target.Text, " to its owner's hand")
	targetSpec, ok := permanentTargetSpec(target)
	if !ok || ctx.text != "Return "+target.Text+" to its owner's hand." {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Bounce{
					Object: object,
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedPermanentTargetSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok || ctx.text != verb+" "+ctx.content.Targets[0].Text+"." {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(game.TargetPermanentReference(0)),
			},
		},
	}.Ability(), nil
}

func lowerFixedCardCountPlayerSpell(
	ctx contentCtx,
	syntax parser.Ability,
	controllerVerb string,
	targetVerb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	// Allow a single EventPlayer reference for "They {verb} N card(s)." bodies;
	// reject all other non-zero-reference forms.
	hasEventPlayerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPlayer
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !hasEventPlayerRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	amount, ok := cardCountQuantity(effect.Amount, allowDynamic)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case hasEventPlayerRef && len(ctx.content.Targets) == 0 &&
		exactEventPlayerCardCountSyntax(ctx.text, controllerVerb, effect.Amount):
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		!hasEventPlayerRef &&
		(exactCardCountPlayerSyntax(syntax.Tokens, syntax.Atoms, controllerVerb, effect.Amount) ||
			exactDynamicCardCountPlayerText(ctx.text, "", controllerVerb, effect.Amount)):
	case len(ctx.content.Targets) == 1 &&
		!hasEventPlayerRef &&
		(exactTargetCardCountPlayerSyntax(syntax.Tokens, syntax.Atoms, targetVerb, effect.Amount) ||
			exactDynamicCardCountPlayerText(ctx.text, titleFirst(ctx.content.Targets[0].Text), targetVerb, effect.Amount)) &&
		strings.EqualFold(syntax.Tokens[0].Text, "target") &&
		strings.EqualFold(syntax.Tokens[1].Text, "player"):
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{targetSpec}
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, playerRef),
			},
		},
	}.Ability(), nil
}

func lowerFixedControllerSpell(
	ctx contentCtx,
	syntax parser.Ability,
	verb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	amount, ok := controllerActionQuantity(effect.Amount, allowDynamic)
	if !ok || !exactControllerAmountSyntax(syntax.Tokens, syntax.Atoms, ctx.text, verb, effect.Amount) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, game.ControllerReference()),
			},
		},
	}.Ability(), nil
}

func cardCountQuantity(amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

func controllerActionQuantity(amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

func lowerFixedLifeSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
	groupPrimitiveFactory func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	amountText := "X"
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
		amountText = fmt.Sprint(effect.Amount.Value)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower ||
			len(ctx.content.References) != 0 {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact supported life changes",
			)
		}
		amount = game.Dynamic(dynamic)
	case len(ctx.content.References) != 0:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact supported life changes",
		)
	default:
	}
	// Group patterns: "Each opponent gains/loses N life." / "Each player gains/loses N life."
	// These require a known fixed amount and no targets.
	if len(ctx.content.Targets) == 0 && effect.Amount.Known {
		switch {
		case exactLifeAmountSyntax("Each opponent", verb+"s", ctx.text, effect.Amount, amountText):
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.OpponentsReference()),
				}},
			}.Ability(), nil
		case exactLifeAmountSyntax("Each player", verb+"s", ctx.text, effect.Amount, amountText):
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.AllPlayersReference()),
				}},
			}.Ability(), nil
		}
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ctx.content.Targets) == 0 &&
		exactLifeAmountSyntax("You", verb, ctx.text, effect.Amount, amountText):
	case len(ctx.content.Targets) == 0 &&
		exactLifeAmountSyntax("That player", verb+"s", ctx.text, effect.Amount, amountText):
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		exactLifeAmountSyntax("They", verb, ctx.text, effect.Amount, amountText):
		// "They" is a pronoun for the player who triggered the event (e.g. "they lose 2 life"
		// in "Whenever an opponent draws a card, they lose 2 life.").
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 1:
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok ||
			!exactLifeAmountSyntax(
				titleFirst(ctx.content.Targets[0].Text),
				verb+"s",
				ctx.text,
				effect.Amount,
				amountText,
			) {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		targets = []game.TargetSpec{targetSpec}
		playerRef = game.TargetPlayerReference(0)
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: primitiveFactory(amount, playerRef),
		}},
	}.Ability(), nil
}

func lowerFixedDestroySpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if group, ok := exactMassDestroyGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Destroy{
						Group: group,
					},
				},
			},
		}.Ability(), nil
	}
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.content.Effects[0].Negated {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok || ctx.text != "Destroy "+ctx.content.Targets[0].Text+"." {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Destroy{
					Object: game.TargetPermanentReference(0),
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedExileSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if group, ok := exactMassExileGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Exile{Group: group},
			}},
		}.Ability(), nil
	}
	return lowerFixedPermanentTargetSpell(ctx, "Exile", func(object game.ObjectReference) game.Primitive {
		return game.Exile{Object: object}
	})
}

func exactMassDestroyGroup(ctx contentCtx) (game.GroupReference, bool) {
	return exactMassGroup(ctx, "Destroy all ")
}

func exactMassExileGroup(ctx contentCtx) (game.GroupReference, bool) {
	return exactMassGroup(ctx, "Exile all ")
}

func exactMassGroup(ctx contentCtx, verbPrefix string) (game.GroupReference, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.content.Effects[0].Negated {
		return game.GroupReference{}, false
	}
	if !strings.HasPrefix(ctx.text, verbPrefix) || !strings.HasSuffix(ctx.text, ".") {
		return game.GroupReference{}, false
	}
	phrase := ctx.text[len(verbPrefix) : len(ctx.text)-1]
	selection, ok := massGroupSelection(ctx.content.Effects[0].Selector, ctx.content.Keywords)
	if !ok {
		return game.GroupReference{}, false
	}
	if !massGroupSyntaxAccepted(phrase) {
		return game.GroupReference{}, false
	}
	if !massGroupKeywordsMatch(ctx.content.Keywords, phrase, selection) {
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

func massGroupKeywordsMatch(keywords []compiler.CompiledKeyword, phrase string, selection game.Selection) bool {
	if selection.Keyword == game.KeywordNone {
		return len(keywords) == 0
	}
	keywordText, ok := strings.CutPrefix(strings.ToLower(phrase), "creatures with ")
	return ok &&
		len(keywords) == 1 &&
		keywords[0].Parameter == "" &&
		strings.EqualFold(keywords[0].Text, keywordText)
}

func massGroupSelection(selector compiler.CompiledSelector, keywords []compiler.CompiledKeyword) (game.Selection, bool) {
	selection := game.Selection{
		RequiredTypesAny: append([]types.Card(nil), selector.RequiredTypesAny()...),
		ExcludedTypes:    append([]types.Card(nil), selector.ExcludedTypes()...),
		ColorsAny:        append([]color.Color(nil), selector.ColorsAny()...),
		ExcludedColors:   append([]color.Color(nil), selector.ExcludedColors()...),
		ExcludeSource:    selector.Other,
	}
	if len(selection.RequiredTypesAny) == 0 {
		if requiredType, ok := massGroupRequiredType(selector.Kind); ok {
			selection.RequiredTypes = []types.Card{requiredType}
		} else if selector.Kind != compiler.SelectorPermanent {
			return game.Selection{}, false
		}
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent, compiler.ControllerNotYou:
		selection.Controller = game.ControllerOpponent
	default:
		return game.Selection{}, false
	}
	if selector.Tapped {
		selection.Tapped = game.TriTrue
	}
	if selector.MatchManaValue {
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.MatchPower {
		selection.Power = opt.Val(selector.Power)
	}
	if selector.MatchToughness {
		selection.Toughness = opt.Val(selector.Toughness)
	}
	if len(keywords) > 0 {
		if len(keywords) != 1 || keywords[0].Parameter != "" {
			return game.Selection{}, false
		}
		keyword, ok := oracleKeyword(keywords[0].Name)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	return selection, true
}

func massGroupRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorPlaneswalker:
		return types.Planeswalker, true
	default:
		return "", false
	}
}

// massGroupSyntaxAccepted parses the qualifier in "Verb all <qualifier>."
// sentences only to prove exact supported syntax. Selection values come from the
// compiler's typed selector IR.
func massGroupSyntaxAccepted(phrase string) bool {
	if phrase == "" || strings.TrimSpace(phrase) != phrase {
		return false
	}
	phrase = strings.ToLower(phrase)

	hadControllerSuffix := false
	for _, suffix := range []string{" you don't control", " your opponents control", " you control"} {
		if remainder, ok := strings.CutSuffix(phrase, suffix); ok {
			phrase = remainder
			hadControllerSuffix = true
			break
		}
	}

	if massGroupNumericSyntaxAccepted(phrase) {
		return true
	}
	if !hadControllerSuffix {
		if keywordText, ok := strings.CutPrefix(phrase, "creatures with "); ok {
			return massGroupKeywordSyntaxAccepted(keywordText)
		}
	}
	if massGroupBaseNounSyntaxAccepted(phrase) {
		return true
	}
	if remainder, ok := strings.CutPrefix(phrase, "other "); ok {
		return massGroupBaseNounSyntaxAccepted(remainder)
	}
	if remainder, ok := strings.CutPrefix(phrase, "tapped "); ok {
		return massGroupBaseNounSyntaxAccepted(remainder)
	}
	for _, prefix := range []string{"nonland ", "nonartifact ", "noncreature ", "nonenchantment "} {
		remainder, ok := strings.CutPrefix(phrase, prefix)
		if !ok {
			continue
		}
		return massGroupBaseNounSyntaxAccepted(remainder)
	}
	for _, prefix := range []string{"white ", "blue ", "black ", "red ", "green ", "nonwhite ", "nonblue ", "nonblack ", "nonred ", "nongreen "} {
		remainder, ok := strings.CutPrefix(phrase, prefix)
		if !ok {
			continue
		}
		if strings.Contains(remainder, " ") {
			return false
		}
		return massGroupBaseNounSyntaxAccepted(remainder)
	}
	return false
}

func massGroupBaseNounSyntaxAccepted(phrase string) bool {
	switch phrase {
	case "creatures", "artifacts", "enchantments", "lands", "planeswalkers", "permanents",
		"creatures and lands", "creatures and planeswalkers", "artifacts and enchantments",
		"artifacts and creatures", "artifacts, creatures, and enchantments",
		"artifacts, creatures, and lands":
		return true
	default:
		return false
	}
}

func massGroupKeywordSyntaxAccepted(text string) bool {
	switch text {
	case "flying", "reach", "trample", "lifelink", "deathtouch", "indestructible", "haste", "menace", "vigilance":
		return true
	default:
		return false
	}
}

func massGroupNumericSyntaxAccepted(phrase string) bool {
	for _, qualifier := range []string{"mana value", "power", "toughness"} {
		prefix := "creatures with " + qualifier + " "
		comparisonText, ok := strings.CutPrefix(phrase, prefix)
		if !ok {
			continue
		}
		return massGroupComparisonSyntaxAccepted(comparisonText)
	}
	return false
}

func massGroupComparisonSyntaxAccepted(text string) bool {
	parts := strings.Fields(text)
	switch {
	case len(parts) == 1:
		_, err := strconv.Atoi(parts[0])
		return err == nil
	case len(parts) == 3 && parts[0] == "equal" && parts[1] == "to":
		_, err := strconv.Atoi(parts[2])
		return err == nil
	case len(parts) == 3 && parts[1] == "or":
		_, err := strconv.Atoi(parts[0])
		if err != nil {
			return false
		}
		switch parts[2] {
		case "less", "greater":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func lowerFixedDrawSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	// Allow a single EventPlayer reference for "They draw N card(s)." bodies;
	// reject all other non-zero-reference forms.
	hasEventPlayerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPlayer
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !hasEventPlayerRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact supported card draw",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case hasEventPlayerRef && len(ctx.content.Targets) == 0 &&
		exactEventPlayerDrawSyntax(ctx.text, effect.Amount):
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		!hasEventPlayerRef &&
		(exactControllerDrawSyntax(syntax.Tokens, syntax.Atoms, effect.Amount.Value) ||
			(!effect.Amount.Known &&
				effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
				exactXControllerDrawSyntax(syntax.Tokens)) ||
			exactDynamicDrawSyntax(ctx.text, "", effect.Amount)):
	case len(ctx.content.Targets) == 1 &&
		!hasEventPlayerRef &&
		(exactTargetPlayerDrawSyntax(syntax.Tokens, syntax.Atoms, effect.Amount.Value) ||
			(!effect.Amount.Known &&
				effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
				exactXTargetPlayerDrawSyntax(syntax.Tokens)) ||
			exactDynamicDrawSyntax(ctx.text, titleFirst(ctx.content.Targets[0].Text), effect.Amount)) &&
		ctx.content.Targets[0].Cardinality.Min == 1 &&
		ctx.content.Targets[0].Cardinality.Max == 1 &&
		ctx.content.Targets[0].Selector.Kind == compiler.SelectorPlayer:
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{
			{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target player",
				Allow:      game.TargetAllowPlayer,
			},
		}
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: game.Draw{
					Amount: amount,
					Player: playerRef,
				},
			},
		},
	}.Ability(), nil
}

func lowerDynamicAmount(amount compiler.CompiledAmount, object game.ObjectReference) (game.DynamicAmount, bool) {
	if amount.Multiplier < 1 {
		return game.DynamicAmount{}, false
	}
	dynamic := game.DynamicAmount{Multiplier: amount.Multiplier}
	switch amount.DynamicKind {
	case compiler.DynamicAmountCount:
		if dynamic, ok := dynamicCardZoneAmount(amount.Selector(), amount.Multiplier); ok {
			return dynamic, true
		}
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountCountSelector
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountControllerLife:
		dynamic.Kind = game.DynamicAmountControllerLife
	case compiler.DynamicAmountOpponentCount:
		dynamic.Kind = game.DynamicAmountOpponentCount
	case compiler.DynamicAmountBasicLandTypes:
		dynamic.Kind = game.DynamicAmountControllerBasicLandTypeCount
	case compiler.DynamicAmountSourcePower:
		if len(object.Validate()) != 0 {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountObjectPower
		dynamic.Object = object
	default:
		return game.DynamicAmount{}, false
	}
	return dynamic, true
}

func dynamicAmountSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	var requiredType types.Card
	switch selector.Kind {
	case compiler.SelectorArtifact:
		requiredType = types.Artifact
	case compiler.SelectorCreature:
		requiredType = types.Creature
	case compiler.SelectorEnchantment:
		requiredType = types.Enchantment
	case compiler.SelectorLand:
		requiredType = types.Land
	case compiler.SelectorPermanent:
	default:
		return game.Selection{}, false
	}
	var controller game.ControllerRelation
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		controller = game.ControllerYou
	case compiler.ControllerOpponent:
		controller = game.ControllerOpponent
	default:
		return game.Selection{}, false
	}
	selection := game.Selection{Controller: controller}
	if requiredType != "" {
		selection.RequiredTypes = []types.Card{requiredType}
	}
	if selector.Keyword != "" {
		keyword, ok := oracleKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	return selection, true
}

func dynamicCardZoneAmount(selector compiler.CompiledSelector, multiplier int) (game.DynamicAmount, bool) {
	if selector.Kind != compiler.SelectorCard || selector.Zone == zone.None {
		return game.DynamicAmount{}, false
	}
	if selector.Zone != zone.Graveyard || selector.Controller != compiler.ControllerYou {
		return game.DynamicAmount{}, false
	}
	keyword, ok := oracleKeyword(selector.Keyword)
	if !ok || keyword != game.Cycling {
		return game.DynamicAmount{}, false
	}
	player := game.ControllerReference()
	return game.DynamicAmount{
		Kind:       game.DynamicAmountCountCardsInZone,
		Multiplier: multiplier,
		Player:     &player,
		CardZone:   selector.Zone,
		Selection:  &game.Selection{Keyword: keyword},
	}, true
}

func oracleKeyword(name string) (game.Keyword, bool) {
	switch name {
	case "Cycling":
		return game.Cycling, true
	case "Flying":
		return game.Flying, true
	case "Reach":
		return game.Reach, true
	case "Trample":
		return game.Trample, true
	case "Lifelink":
		return game.Lifelink, true
	case "Deathtouch":
		return game.Deathtouch, true
	case "Indestructible":
		return game.Indestructible, true
	case "Haste":
		return game.Haste, true
	case "Menace":
		return game.Menace, true
	case "Vigilance":
		return game.Vigilance, true
	default:
		return game.KeywordNone, false
	}
}

func exactDamageAmountSyntax(
	cardName string,
	ctx contentCtx,
	amount compiler.CompiledAmount,
	fixedText string,
) bool {
	target := ctx.content.Targets[0].Text
	switch amount.DynamicForm {
	case compiler.DynamicAmountFormNone:
		return ctx.text == fmt.Sprintf("%s deals %s damage to %s.", cardName, fixedText, target)
	case compiler.DynamicAmountEqual:
		return ctx.text == fmt.Sprintf("%s deals damage %s to %s.", cardName, amount.Text, target)
	case compiler.DynamicAmountForEach:
		return ctx.text == fmt.Sprintf(
			"%s deals %d damage %s to %s.",
			cardName,
			amount.Multiplier,
			amount.Text,
			target,
		)
	case compiler.DynamicAmountWhereX:
		return ctx.text == fmt.Sprintf(
			"%s deals X damage to %s, %s.",
			cardName,
			target,
			amount.Text,
		)
	default:
		return false
	}
}

func exactDamageAmountReferences(amount compiler.CompiledAmount, references []compiler.CompiledReference) bool {
	if amount.DynamicKind != compiler.DynamicAmountSourcePower {
		_, ok := lowerDamageSourceReference(references)
		return ok
	}
	if len(references) != 2 ||
		references[1].Span != amount.ReferenceSpan {
		return false
	}
	_, ok := lowerDamageSourceReference(references[:1])
	return ok && references[1].Binding == references[0].Binding
}

func lowerDamageSourceReference(references []compiler.CompiledReference) (game.ObjectReference, bool) {
	if len(references) != 1 {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(references[0], referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
}

func exactLifeAmountSyntax(
	subject, verb, text string,
	amount compiler.CompiledAmount,
	fixedText string,
) bool {
	switch amount.DynamicForm {
	case compiler.DynamicAmountFormNone:
		return text == fmt.Sprintf("%s %s %s life.", subject, verb, fixedText)
	case compiler.DynamicAmountEqual:
		return text == fmt.Sprintf("%s %s life %s.", subject, verb, amount.Text)
	case compiler.DynamicAmountForEach:
		return text == fmt.Sprintf("%s %s %d life %s.", subject, verb, amount.Multiplier, amount.Text)
	case compiler.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s %s X life, %s.", subject, verb, amount.Text)
	default:
		return false
	}
}

func exactDynamicDrawSyntax(text, subject string, amount compiler.CompiledAmount) bool {
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return false
	}
	prefix := "Draw"
	if subject != "" {
		prefix = subject + " draws"
	}
	switch amount.DynamicForm {
	case compiler.DynamicAmountEqual:
		return text == fmt.Sprintf("%s cards %s.", prefix, amount.Text)
	case compiler.DynamicAmountForEach:
		noun := "cards"
		if amount.Multiplier == 1 {
			return text == fmt.Sprintf("%s 1 card %s.", prefix, amount.Text) ||
				text == fmt.Sprintf("%s a card %s.", prefix, amount.Text)
		}
		return text == fmt.Sprintf("%s %d %s %s.", prefix, amount.Multiplier, noun, amount.Text)
	case compiler.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s X cards, %s.", prefix, amount.Text)
	default:
		return false
	}
}

func exactModifyPTAmountSyntax(ctx contentCtx, effect compiler.CompiledEffect) bool {
	subject := titleFirst(ctx.content.Targets[0].Text)
	amount := effect.Amount
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return len(ctx.content.References) == 0 &&
			ctx.text == fmt.Sprintf(
				"%s gets %s/%s until end of turn.",
				subject,
				signedAmountText(effect.PowerDelta),
				signedAmountText(effect.ToughnessDelta),
			)
	}
	if len(ctx.content.References) != 0 || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return false
	}
	switch amount.DynamicForm {
	case compiler.DynamicAmountForEach:
		if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known ||
			!dynamicPTMultiplierMatches(amount.Multiplier, effect.PowerDelta, effect.ToughnessDelta) {
			return false
		}
		return ctx.text == fmt.Sprintf(
			"%s gets %s/%s %s until end of turn.",
			subject,
			signedAmountText(effect.PowerDelta),
			signedAmountText(effect.ToughnessDelta),
			amount.Text,
		) || ctx.text == fmt.Sprintf(
			"%s gets %s/%s until end of turn %s.",
			subject,
			signedAmountText(effect.PowerDelta),
			signedAmountText(effect.ToughnessDelta),
			amount.Text,
		)
	case compiler.DynamicAmountWhereX:
		return !effect.PowerDelta.Known &&
			!effect.ToughnessDelta.Known &&
			ctx.text == fmt.Sprintf("%s gets +X/+X until end of turn, %s.", subject, amount.Text)
	default:
		return false
	}
}

func matchesExactTemporaryGroupPTSyntax(syntax parser.Ability, effect compiler.CompiledEffect) bool {
	tokens := syntaxSemanticTokens(syntax)
	prefixLength, ok := matchesStaticPTBuffPrefix(tokens, effect)
	return ok &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		matchesUntilEndOfTurnSuffix(tokens, prefixLength)
}

func matchesExactTemporaryKeywordSyntax(
	syntax parser.Ability,
	target compiler.CompiledTarget,
	keywords []compiler.CompiledKeyword,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	targetLength := leadingSpanTokenCount(tokens, target.Span)
	suffixStart := len(tokens) - 5
	return targetLength > 0 &&
		suffixStart > targetLength+1 &&
		equalTokenWord(tokens[targetLength], "gains") &&
		matchesExactKeywordList(tokens[targetLength+1:suffixStart], keywords) &&
		matchesUntilEndOfTurnSuffix(tokens, suffixStart)
}

func matchesExactTemporaryPTKeywordSyntax(
	syntax parser.Ability,
	target compiler.CompiledTarget,
	effect compiler.CompiledEffect,
	keywords []compiler.CompiledKeyword,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	targetLength := leadingSpanTokenCount(tokens, target.Span)
	keywordStart := targetLength + 8
	suffixStart := len(tokens) - 5
	return targetLength > 0 &&
		suffixStart > keywordStart &&
		equalTokenWord(tokens[targetLength], "gets") &&
		tokensMatchSignedAmount(tokens[targetLength+1], tokens[targetLength+2], effect.PowerDelta) &&
		tokens[targetLength+3].Kind == shared.Slash &&
		tokensMatchSignedAmount(tokens[targetLength+4], tokens[targetLength+5], effect.ToughnessDelta) &&
		equalTokenWord(tokens[targetLength+6], "and") &&
		equalTokenWord(tokens[targetLength+7], "gains") &&
		matchesExactKeywordList(tokens[keywordStart:suffixStart], keywords) &&
		matchesUntilEndOfTurnSuffix(tokens, suffixStart)
}

func leadingSpanTokenCount(tokens []shared.Token, span shared.Span) int {
	length := 0
	for length < len(tokens) && spanCovered(tokens[length].Span, []shared.Span{span}) {
		length++
	}
	return length
}

func matchesUntilEndOfTurnSuffix(tokens []shared.Token, start int) bool {
	return start >= 0 &&
		len(tokens) == start+5 &&
		equalTokenWord(tokens[start], "until") &&
		equalTokenWord(tokens[start+1], "end") &&
		equalTokenWord(tokens[start+2], "of") &&
		equalTokenWord(tokens[start+3], "turn") &&
		tokens[start+4].Kind == shared.Period
}

func dynamicPTMultiplierMatches(
	multiplier int,
	power, toughness compiler.CompiledSignedAmount,
) bool {
	matches := func(amount compiler.CompiledSignedAmount) bool {
		return amount.Value == 0 || amount.Value == multiplier
	}
	return multiplier > 0 && matches(power) && matches(toughness)
}

func dynamicSignedQuantity(
	dynamic game.DynamicAmount,
	amount compiler.CompiledSignedAmount,
) game.Quantity {
	if amount.Value == 0 {
		return game.Fixed(0)
	}
	if amount.Negative {
		dynamic.Multiplier = -dynamic.Multiplier
	}
	return game.Dynamic(dynamic)
}

func exactXControllerDrawSyntax(tokens []shared.Token) bool {
	return len(tokens) == 4 &&
		equalTokenWord(tokens[0], "draw") &&
		equalTokenWord(tokens[1], "X") &&
		equalTokenWord(tokens[2], "cards") &&
		tokens[3].Kind == shared.Period
}

func exactXTargetPlayerDrawSyntax(tokens []shared.Token) bool {
	return len(tokens) == 6 &&
		equalTokenWord(tokens[0], "target") &&
		equalTokenWord(tokens[1], "player") &&
		equalTokenWord(tokens[2], "draws") &&
		equalTokenWord(tokens[3], "X") &&
		equalTokenWord(tokens[4], "cards") &&
		tokens[5].Kind == shared.Period
}

func exactControllerDrawSyntax(tokens []shared.Token, atoms parser.Atoms, amount int) bool {
	if len(tokens) != 4 ||
		tokens[0].Kind != shared.Word ||
		!strings.EqualFold(tokens[0].Text, "draw") ||
		tokens[2].Kind != shared.Word ||
		tokens[3].Kind != shared.Period {
		return false
	}
	if amount == 1 &&
		strings.EqualFold(tokens[1].Text, "a") &&
		strings.EqualFold(tokens[2].Text, "card") {
		return true
	}
	return fixedNumberSyntax(tokens[1], atoms, amount) &&
		strings.EqualFold(tokens[2].Text, "cards")
}

func exactTargetPlayerDrawSyntax(tokens []shared.Token, atoms parser.Atoms, amount int) bool {
	return len(tokens) == 6 &&
		tokens[0].Kind == shared.Word &&
		strings.EqualFold(tokens[0].Text, "target") &&
		tokens[1].Kind == shared.Word &&
		strings.EqualFold(tokens[1].Text, "player") &&
		tokens[2].Kind == shared.Word &&
		strings.EqualFold(tokens[2].Text, "draws") &&
		fixedCardCountSyntax(tokens[3], tokens[4], atoms, amount) &&
		tokens[5].Kind == shared.Period
}

func fixedCardCountSyntax(amountToken, cardToken shared.Token, atoms parser.Atoms, amount int) bool {
	if amount == 1 &&
		strings.EqualFold(amountToken.Text, "a") &&
		strings.EqualFold(cardToken.Text, "card") {
		return true
	}
	return fixedNumberSyntax(amountToken, atoms, amount) &&
		strings.EqualFold(cardToken.Text, "cards")
}

func exactCardCountPlayerSyntax(tokens []shared.Token, atoms parser.Atoms, verb string, amount compiler.CompiledAmount) bool {
	if len(tokens) != 4 ||
		!equalTokenWord(tokens[0], verb) ||
		tokens[3].Kind != shared.Period {
		return false
	}
	return cardCountAmountSyntax(tokens[1], tokens[2], atoms, amount)
}

func exactTargetCardCountPlayerSyntax(tokens []shared.Token, atoms parser.Atoms, verb string, amount compiler.CompiledAmount) bool {
	if len(tokens) != 6 ||
		!equalTokenWord(tokens[0], "target") ||
		!equalTokenWord(tokens[1], "player") ||
		!equalTokenWord(tokens[2], verb) ||
		tokens[5].Kind != shared.Period {
		return false
	}
	return cardCountAmountSyntax(tokens[3], tokens[4], atoms, amount)
}

func cardCountAmountSyntax(amountToken, cardToken shared.Token, atoms parser.Atoms, amount compiler.CompiledAmount) bool {
	if amount.Known {
		return fixedCardCountSyntax(amountToken, cardToken, atoms, amount.Value)
	}
	return equalTokenWord(amountToken, "X") &&
		strings.EqualFold(cardToken.Text, "cards")
}

func exactDynamicCardCountPlayerText(text, subject, verb string, amount compiler.CompiledAmount) bool {
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return false
	}
	prefix := titleFirst(verb)
	if subject != "" {
		prefix = subject + " " + verb
	}
	switch amount.DynamicForm {
	case compiler.DynamicAmountForEach:
		if amount.Multiplier == 1 {
			return text == fmt.Sprintf("%s 1 card %s.", prefix, amount.Text) ||
				text == fmt.Sprintf("%s a card %s.", prefix, amount.Text)
		}
		return text == fmt.Sprintf("%s %d cards %s.", prefix, amount.Multiplier, amount.Text)
	case compiler.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s X cards, %s.", prefix, amount.Text)
	default:
		return false
	}
}

// exactEventPlayerDrawSyntax reports whether text is the exact "They draw
// N card(s)." form expected for an event-player draw body. Only fixed known
// amounts are accepted.
func exactEventPlayerDrawSyntax(text string, amount compiler.CompiledAmount) bool {
	if !amount.Known || amount.Value < 1 {
		return false
	}
	if amount.Value == 1 {
		return text == "They draw a card."
	}
	return text == fmt.Sprintf("They draw %d cards.", amount.Value)
}

// exactEventPlayerCardCountSyntax reports whether text is the exact
// "They {verb} N card(s)." form expected for an event-player
// discard/mill/similar body. Only fixed known amounts are accepted.
func exactEventPlayerCardCountSyntax(text, verb string, amount compiler.CompiledAmount) bool {
	if !amount.Known || amount.Value < 1 {
		return false
	}
	if amount.Value == 1 {
		return text == fmt.Sprintf("They %s a card.", verb)
	}
	return text == fmt.Sprintf("They %s %d cards.", verb, amount.Value)
}

func exactControllerAmountSyntax(tokens []shared.Token, atoms parser.Atoms, text, verb string, amount compiler.CompiledAmount) bool {
	if amount.Known {
		return len(tokens) == 3 &&
			equalTokenWord(tokens[0], verb) &&
			fixedNumberSyntax(tokens[1], atoms, amount.Value) &&
			tokens[2].Kind == shared.Period
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return len(tokens) == 3 &&
			equalTokenWord(tokens[0], verb) &&
			equalTokenWord(tokens[1], "X") &&
			tokens[2].Kind == shared.Period
	}
	switch amount.DynamicForm {
	case compiler.DynamicAmountForEach:
		return text == fmt.Sprintf("%s %d %s.", titleFirst(verb), amount.Multiplier, amount.Text)
	case compiler.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s X, %s.", titleFirst(verb), amount.Text)
	default:
		return false
	}
}

func fixedNumberSyntax(token shared.Token, atoms parser.Atoms, amount int) bool {
	if token.Kind == shared.Integer {
		return token.Text == fmt.Sprint(amount)
	}
	value, ok := atoms.CardinalAt(token.Span)
	return ok && value == amount
}

func singleSelfReference(references []compiler.CompiledReference) bool {
	return len(references) == 1 && references[0].Binding == compiler.ReferenceBindingSource
}

func damageTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
	}
	switch target.Selector.Kind {
	case compiler.SelectorAny:
		if target.Text != "any target" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent | game.TargetAllowPlayer
	case compiler.SelectorCreature, compiler.SelectorPlaneswalker, compiler.SelectorBattle:
		permanent, ok := permanentTargetSpec(target)
		if !ok {
			return game.TargetSpec{}, false
		}
		return permanent, true
	case compiler.SelectorPlayer:
		if target.Text != "target player" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPlayer
	case compiler.SelectorOpponent:
		if target.Text != "target opponent" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPlayer
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func permanentTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
	}
	var noun string
	switch target.Selector.Kind {
	case compiler.SelectorArtifact:
		noun = "artifact"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Artifact}}
	case compiler.SelectorCreature:
		noun = "creature"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	case compiler.SelectorEnchantment:
		noun = "enchantment"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Enchantment}}
	case compiler.SelectorLand:
		noun = "land"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Land}}
	case compiler.SelectorPermanent:
		noun = "permanent"
	case compiler.SelectorPlaneswalker:
		noun = "planeswalker"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Planeswalker}}
	case compiler.SelectorBattle:
		noun = "battle"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Battle}}
	default:
		return game.TargetSpec{}, false
	}
	if target.Selector.Another || target.Selector.Other ||
		(target.Selector.Tapped && target.Selector.Untapped) ||
		((target.Selector.Tapped || target.Selector.Untapped) &&
			(target.Selector.Attacking || target.Selector.Blocking)) {
		return game.TargetSpec{}, false
	}

	expected := "target "
	switch {
	case target.Selector.Attacking && target.Selector.Blocking:
		expected += "attacking or blocking "
		spec.Predicate.CombatState = game.CombatStateAttackingOrBlocking
	case target.Selector.Attacking:
		expected += "attacking "
		spec.Predicate.CombatState = game.CombatStateAttacking
	case target.Selector.Blocking:
		expected += "blocking "
		spec.Predicate.CombatState = game.CombatStateBlocking
	case target.Selector.Tapped:
		expected += "tapped "
		spec.Predicate.Tapped = game.TriTrue
	case target.Selector.Untapped:
		expected += "untapped "
		spec.Predicate.Tapped = game.TriFalse
	default:
	}
	expected += noun
	switch target.Selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		expected += " you control"
		spec.Predicate.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		expected += " an opponent controls"
		spec.Predicate.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		expected += " you don't control"
		spec.Predicate.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	if !strings.EqualFold(target.Text, expected) {
		return game.TargetSpec{}, false
	}
	spec.Constraint = lowerFirst(target.Text)
	return spec, true
}

func stackSpellTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if target.Selector.Another || target.Selector.Other ||
		target.Selector.Controller != compiler.ControllerAny {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
		},
	}
	text := strings.ToLower(target.Text)
	switch target.Selector.Kind {
	case compiler.SelectorSpell:
		switch text {
		case "target spell":
		case "target instant spell":
			spec.Predicate.SpellCardTypes = []types.Card{types.Instant}
		case "target sorcery spell":
			spec.Predicate.SpellCardTypes = []types.Card{types.Sorcery}
		case "target noncreature spell":
			spec.Predicate.ExcludedSpellCardTypes = []types.Card{types.Creature}
		default:
			return game.TargetSpec{}, false
		}
	case compiler.SelectorCreature:
		if text != "target creature spell" {
			return game.TargetSpec{}, false
		}
		spec.Predicate.SpellCardTypes = []types.Card{types.Creature}
	case compiler.SelectorArtifact:
		if text != "target artifact spell" {
			return game.TargetSpec{}, false
		}
		spec.Predicate.SpellCardTypes = []types.Card{types.Artifact}
	default:
		return game.TargetSpec{}, false
	}
	spec.Constraint = lowerFirst(target.Text)
	return spec, true
}

func counterAbilityTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if target.Selector.Another || target.Selector.Other ||
		target.Selector.Controller != compiler.ControllerAny {
		return game.TargetSpec{}, false
	}
	var kinds []game.StackObjectKind
	switch {
	case target.Selector.Kind == compiler.SelectorActivatedAbility && target.Text == "target activated ability":
		kinds = []game.StackObjectKind{game.StackActivatedAbility}
	case target.Selector.Kind == compiler.SelectorTriggeredAbility && target.Text == "target triggered ability":
		kinds = []game.StackObjectKind{game.StackTriggeredAbility}
	case target.Selector.Kind == compiler.SelectorActivatedOrTriggeredAbility && target.Text == "target activated or triggered ability":
		kinds = []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility}
	case target.Selector.Kind == compiler.SelectorSpellActivatedOrTriggeredAbility && target.Text == "target spell, activated ability, or triggered ability":
		kinds = []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility}
	default:
		return game.TargetSpec{}, false
	}
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: lowerFirst(target.Text),
		Allow:      game.TargetAllowStackObject,
		Predicate:  game.TargetPredicate{StackObjectKinds: kinds},
	}, true
}

func counterTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if spec, ok := stackSpellTargetSpec(target); ok {
		return spec, true
	}
	return counterAbilityTargetSpec(target)
}

func lowerCounterSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported counter spell",
			"the executable source backend supports only exact counter of one target spell",
		)
	}
	if content, ok := lowerCounterUnlessPaysSpell(ctx); ok {
		return content, nil
	}
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := counterTargetSpec(ctx.content.Targets[0])
	if !ok || ctx.text != "Counter "+ctx.content.Targets[0].Text+"." {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
		}},
	}.Ability(), nil
}

func lowerSacrificeSpell(
	ctx contentCtx,
	syntax parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported sacrifice spell",
			"the executable source backend does not yet lower this sacrifice effect",
		)
	}

	effect := ctx.content.Effects[0]
	// Exact source-bound or event-permanent-bound "Sacrifice it.": the
	// controller sacrifices the named object. Only accepted when there are no
	// targets, no conditions/keywords/modes, and the text is exactly
	// "Sacrifice it." with a single source or event-permanent reference.
	if ctx.text == "Sacrifice it." &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 &&
		!effect.Negated {
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource:      true,
			SourceCardObject: true,
			AllowEvent:       true,
		})
		if ok {
			return game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Sacrifice{Object: object},
			}}}.Ability(), nil
		}
	}
	// Strict fail-closed: reject unsupported modifiers and dynamic amounts.
	if len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated {
		return unsupported()
	}

	// Map selector kind to game.Selection; fail-closed for unknown kinds.
	var selection game.Selection
	switch effect.Selector.Kind {
	case compiler.SelectorCreature:
		selection = game.Selection{RequiredTypes: []types.Card{types.Creature}}
	case compiler.SelectorArtifact:
		selection = game.Selection{RequiredTypes: []types.Card{types.Artifact}}
	case compiler.SelectorLand:
		selection = game.Selection{RequiredTypes: []types.Card{types.Land}}
	case compiler.SelectorEnchantment:
		selection = game.Selection{RequiredTypes: []types.Card{types.Enchantment}}
	case compiler.SelectorPermanent:
		// zero Selection = any permanent
	default:
		return unsupported()
	}

	amount := game.Fixed(effect.Amount.Value)

	switch {
	case len(ctx.content.Targets) == 1:
		// "Target player/opponent sacrifices <N> <type>."
		target := ctx.content.Targets[0]
		if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
			return unsupported()
		}
		targetSpec, ok := playerTargetSpec(target)
		if !ok {
			return unsupported()
		}
		var actor string
		switch target.Selector.Kind {
		case compiler.SelectorPlayer:
			actor = "player"
		case compiler.SelectorOpponent:
			actor = "opponent"
		default:
			return unsupported()
		}
		if !matchesExactSacrificeSyntax(syntax, "target", actor, effect) {
			return unsupported()
		}
		return game.Mode{
			Targets: []game.TargetSpec{targetSpec},
			Sequence: []game.Instruction{{
				Primitive: game.SacrificePermanents{
					Player:    game.TargetPlayerReference(0),
					Amount:    amount,
					Selection: selection,
				},
			}},
		}.Ability(), nil

	case len(ctx.content.Targets) == 0:
		// "Each opponent/player sacrifices <N> <type>."
		var group game.PlayerGroupReference
		var actor string
		switch {
		case strings.HasPrefix(ctx.text, "Each opponent "):
			group = game.OpponentsReference()
			actor = "opponent"
		case strings.HasPrefix(ctx.text, "Each player "):
			group = game.AllPlayersReference()
			actor = "player"
		default:
			return unsupported()
		}
		if !matchesExactSacrificeSyntax(syntax, "each", actor, effect) {
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.SacrificePermanents{
					PlayerGroup: group,
					Amount:      amount,
					Selection:   selection,
				},
			}},
		}.Ability(), nil

	default:
		return unsupported()
	}
}

func matchesExactSacrificeSyntax(
	syntax parser.Ability,
	actorQuantifier, actor string,
	effect compiler.CompiledEffect,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	singular, plural, ok := sacrificeSelectorNouns(effect.Selector.Kind)
	if !ok ||
		(len(tokens) != 6 && len(tokens) != 9) ||
		!equalTokenWord(tokens[0], actorQuantifier) ||
		!equalTokenWord(tokens[1], actor) ||
		!equalTokenWord(tokens[2], "sacrifices") ||
		!matchesExactSacrificeChoiceSuffix(tokens) {
		return false
	}
	if effect.Amount.Value == 1 {
		return (equalTokenWord(tokens[3], "a") ||
			equalTokenWord(tokens[3], "an") ||
			equalTokenWord(tokens[3], "one")) &&
			equalTokenWord(tokens[4], singular)
	}
	return fixedNumberSyntax(tokens[3], syntax.Atoms, effect.Amount.Value) &&
		equalTokenWord(tokens[4], plural)
}

func matchesExactSacrificeChoiceSuffix(tokens []shared.Token) bool {
	if len(tokens) == 6 {
		return tokens[5].Kind == shared.Period
	}
	return len(tokens) == 9 &&
		equalTokenWord(tokens[5], "of") &&
		equalTokenWord(tokens[6], "their") &&
		equalTokenWord(tokens[7], "choice") &&
		tokens[8].Kind == shared.Period
}

func sacrificeSelectorNouns(kind compiler.SelectorKind) (singular, plural string, ok bool) {
	switch kind {
	case compiler.SelectorCreature:
		return "creature", "creatures", true
	case compiler.SelectorArtifact:
		return "artifact", "artifacts", true
	case compiler.SelectorLand:
		return "land", "lands", true
	case compiler.SelectorEnchantment:
		return "enchantment", "enchantments", true
	case compiler.SelectorPermanent:
		return "permanent", "permanents", true
	default:
		return "", "", false
	}
}

func lowerCounterUnlessPaysSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	targetText, manaCost, ok := counterUnlessPaysParts(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	target.Text = targetText
	targetSpec, ok := stackSpellTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	const resultKey = game.ResultKey("unless-paid")
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Pay{Payment: game.ResolutionPayment{
					Prompt:   "Pay " + manaCost.String() + "?",
					Payer:    opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
					ManaCost: opt.Val(manaCost),
				}},
				PublishResult: resultKey,
			},
			{
				Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       resultKey,
					Succeeded: game.TriFalse,
				}),
			},
		},
	}.Ability(), true
}

func counterUnlessPaysParts(ctx contentCtx) (string, cost.Mana, bool) {
	const marker = " unless its controller pays "
	before, after, ok := strings.Cut(ctx.text, marker)
	if !ok || !strings.HasPrefix(before, "Counter ") || !strings.HasSuffix(after, ".") {
		return "", nil, false
	}
	targetText := strings.TrimPrefix(before, "Counter ")
	manaText := strings.TrimSuffix(after, ".")
	manaCost, err := parseManaCostValue(manaText)
	if err != nil || len(manaCost) == 0 || manaCost.String() != manaText || manaCostHasVariableSymbol(manaCost) {
		return "", nil, false
	}
	if ctx.content.Conditions[0].Predicate != compiler.ConditionPredicateTargetControllerDoesNotPay {
		return "", nil, false
	}
	return targetText, manaCost, true
}

func manaCostHasVariableSymbol(manaCost cost.Mana) bool {
	for _, symbol := range manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

func playerTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPlayer,
	}
	switch target.Selector.Kind {
	case compiler.SelectorPlayer:
		if !strings.EqualFold(target.Text, "target player") {
			return game.TargetSpec{}, false
		}
	case compiler.SelectorOpponent:
		if !strings.EqualFold(target.Text, "target opponent") {
			return game.TargetSpec{}, false
		}
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func signedAmountText(amount compiler.CompiledSignedAmount) string {
	if amount.Negative {
		return fmt.Sprintf("-%d", amount.Value)
	}
	return fmt.Sprintf("+%d", amount.Value)
}

func compiledSignedAmountValue(amount compiler.CompiledSignedAmount) int {
	if amount.Negative {
		return -amount.Value
	}
	return amount.Value
}

func titleFirst(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func lowerFirst(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToLower(text[:1]) + text[1:]
}

func manaColorName(symbol string) (string, bool) {
	switch strings.ToUpper(symbol) {
	case "{W}":
		return "W", true
	case "{U}":
		return "U", true
	case "{B}":
		return "B", true
	case "{R}":
		return "R", true
	case "{G}":
		return "G", true
	case "{C}":
		return "C", true
	default:
		return "", false
	}
}

func manaColorValue(name string) (mana.Color, bool) {
	switch name {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	case "C":
		return mana.C, true
	default:
		return "", false
	}
}

func equalTokenWord(token shared.Token, word string) bool {
	return token.Kind == shared.Word && strings.EqualFold(token.Text, word)
}

func spanCoveredByKeyword(span shared.Span, keywords []compiler.CompiledKeyword) bool {
	for _, keyword := range keywords {
		if keyword.Span.Start.Offset <= span.Start.Offset &&
			keyword.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func spanCoveredByAbilityWord(span shared.Span, abilityWord *parser.Phrase) bool {
	return abilityWord != nil &&
		abilityWord.Span.Start.Offset <= span.Start.Offset &&
		abilityWord.Span.End.Offset >= span.End.Offset
}

func spanCoveredByDelimited(span shared.Span, groups []parser.Delimited) bool {
	for _, group := range groups {
		if group.Span.Start.Offset <= span.Start.Offset &&
			group.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func executableDiagnostic(
	ability compiler.CompiledAbility,
	summary string,
	detail string,
) *shared.Diagnostic {
	return &shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  summary,
		Detail:   detail,
		Span:     ability.Span,
	}
}

func mixedKeywordDiagnostic(ctx contentCtx) *shared.Diagnostic {
	names := make([]string, 0, len(ctx.content.Keywords))
	for _, keyword := range ctx.content.Keywords {
		names = append(names, keyword.Name)
	}
	return contentDiagnostic(
		ctx,
		"unsupported mixed keyword ability",
		fmt.Sprintf(
			"the executable source backend recognized %s but does not yet lower the additional rules text",
			strings.Join(names, ", "),
		),
	)
}

// parseManaCostValue parses a Scryfall mana cost string (e.g., "{2}{W}") into a
// typed cost.Mana value. Empty input yields a nil cost.
func parseManaCostValue(s string) (cost.Mana, error) {
	if s == "" {
		return nil, nil
	}
	matches := manaSymbolRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	out := make(cost.Mana, 0, len(matches))
	for _, match := range matches {
		symbol, err := parseManaSymbolValue(match[1])
		if err != nil {
			return nil, fmt.Errorf("unsupported mana symbol {%s} in cost %q: %w", match[1], s, err)
		}
		out = append(out, symbol)
	}
	return out, nil
}

func parseManaSymbolValue(sym string) (cost.Symbol, error) {
	switch sym {
	case "X":
		return cost.X, nil
	case "C":
		return cost.C, nil
	case "S":
		return cost.S, nil
	case "W":
		return cost.W, nil
	case "U":
		return cost.U, nil
	case "B":
		return cost.B, nil
	case "R":
		return cost.R, nil
	case "G":
		return cost.G, nil
	default:
	}
	if before, ok := strings.CutSuffix(sym, "/P"); ok {
		manaColor, ok := manaColorValue(before)
		if !ok {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.PhyrexianMana(manaColor), nil
	}
	if strings.Contains(sym, "/") {
		parts := strings.SplitN(sym, "/", 2)
		if _, err := strconv.Atoi(parts[0]); err == nil {
			manaColor, ok := manaColorValue(parts[1])
			if !ok {
				return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
			}
			return cost.Twobrid(manaColor), nil
		}
		first, ok := manaColorValue(parts[0])
		second, ok2 := manaColorValue(parts[1])
		if !ok || !ok2 {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.HybridMana(first, second), nil
	}
	if n, err := strconv.Atoi(sym); err == nil {
		return cost.O(n), nil
	}
	return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
}

// keywordStaticBodies maps a keyword name to its reusable typed StaticAbility and
// the package-level variable reference the Renderer emits for it.
var keywordStaticBodies = map[string]loweredStaticAbility{
	"Devoid":         {Body: game.DevoidStaticBody, VarName: "game.DevoidStaticBody"},
	"Deathtouch":     {Body: game.DeathtouchStaticBody, VarName: "game.DeathtouchStaticBody"},
	"Defender":       {Body: game.DefenderStaticBody, VarName: "game.DefenderStaticBody"},
	"Delve":          {Body: game.DelveStaticBody, VarName: "game.DelveStaticBody"},
	"Double strike":  {Body: game.DoubleStrikeStaticBody, VarName: "game.DoubleStrikeStaticBody"},
	"Exalted":        {Body: game.ExaltedStaticBody, VarName: "game.ExaltedStaticBody"},
	"First strike":   {Body: game.FirstStrikeStaticBody, VarName: "game.FirstStrikeStaticBody"},
	"Flash":          {Body: game.FlashStaticBody, VarName: "game.FlashStaticBody"},
	"Flying":         {Body: game.FlyingStaticBody, VarName: "game.FlyingStaticBody"},
	"Haste":          {Body: game.HasteStaticBody, VarName: "game.HasteStaticBody"},
	"Hexproof":       {Body: game.HexproofStaticBody, VarName: "game.HexproofStaticBody"},
	"Improvise":      {Body: game.ImproviseStaticBody, VarName: "game.ImproviseStaticBody"},
	"Indestructible": {Body: game.IndestructibleStaticBody, VarName: "game.IndestructibleStaticBody"},
	"Infect":         {Body: game.InfectStaticBody, VarName: "game.InfectStaticBody"},
	"Lifelink":       {Body: game.LifelinkStaticBody, VarName: "game.LifelinkStaticBody"},
	"Menace":         {Body: game.MenaceStaticBody, VarName: "game.MenaceStaticBody"},
	"Persist":        {Body: game.PersistStaticBody, VarName: "game.PersistStaticBody"},
	"Prowess":        {Body: game.ProwessStaticBody, VarName: "game.ProwessStaticBody"},
	"Read ahead":     {Body: game.ReadAheadStaticBody, VarName: "game.ReadAheadStaticBody"},
	"Reach":          {Body: game.ReachStaticBody, VarName: "game.ReachStaticBody"},
	"Shroud":         {Body: game.ShroudStaticBody, VarName: "game.ShroudStaticBody"},
	"Split second":   {Body: game.SplitSecondStaticBody, VarName: "game.SplitSecondStaticBody"},
	"Storm":          {Body: game.StormStaticBody, VarName: "game.StormStaticBody"},
	"Trample":        {Body: game.TrampleStaticBody, VarName: "game.TrampleStaticBody"},
	"Undying":        {Body: game.UndyingStaticBody, VarName: "game.UndyingStaticBody"},
	"Vigilance":      {Body: game.VigilanceStaticBody, VarName: "game.VigilanceStaticBody"},
	"Wither":         {Body: game.WitherStaticBody, VarName: "game.WitherStaticBody"},
	"Cascade":        {Body: game.CascadeStaticBody, VarName: "game.CascadeStaticBody"},
	"Convoke":        {Body: game.ConvokeStaticBody, VarName: "game.ConvokeStaticBody"},
}
