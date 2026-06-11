package cardgen

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle"
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
	sourceSpans        []oracle.Span
}

type semanticConsumption struct {
	cost       bool
	trigger    bool
	optional   bool
	modes      int
	targets    int
	conditions int
	effects    int
	keywords   int
	references int
}

// lowerExecutableFaces lowers every face of a card into typed ability values.
// It returns the face abilities in the same positional order as
// executableFaces and any diagnostics that prevented full lowering.
func lowerExecutableFaces(card *ScryfallCard) ([]loweredFaceAbilities, []oracle.Diagnostic) {
	faces := executableFaces(card)
	lowered := make([]loweredFaceAbilities, len(faces))
	var diagnostics []oracle.Diagnostic
	for i, face := range faces {
		faceAbilities, faceDiagnostics := lowerFaceAbilities(face)
		diagnostics = append(diagnostics, faceDiagnostics...)
		lowered[i] = faceAbilities
	}
	if card.Layout != "adventure" && hasAdventureCastPermission(lowered) {
		diagnostics = append(diagnostics, oracle.Diagnostic{
			Severity: oracle.SeverityWarning,
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
) (loweredFaceAbilities, []oracle.Diagnostic) {
	parsedType := ParseTypeLine(face.TypeLine)
	if len(parsedType.Types) == 0 {
		return loweredFaceAbilities{}, []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported type line",
			Detail:   fmt.Sprintf("type line %q has no supported card type", face.TypeLine),
		}}
	}
	if face.OracleText == "" {
		return loweredFaceAbilities{}, nil
	}
	compilation, diagnostics := oracle.Compile(face.OracleText, oracle.ParseContext{
		CardName:         face.Name,
		InstantOrSorcery: slices.Contains(parsedType.Types, "Instant") || slices.Contains(parsedType.Types, "Sorcery"),
		Planeswalker:     slices.Contains(parsedType.Types, "Planeswalker"),
		Saga:             slices.Contains(parsedType.Subtypes, "Saga"),
	})

	var result loweredFaceAbilities
	var unsupported []oracle.Diagnostic
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
		for _, keyword := range ability.Keywords {
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	if len(ability.Modes) > 0 {
		return lowerModalAbility(cardName, ability, syntax)
	}
	if lowered, ok := lowerEntersPrepared(ability, syntax); ok {
		return lowered, nil
	}
	if handGrant, ok, diagnostic := lowerHandCyclingGrant(ability, syntax); ok {
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		return abilityLowering{
			staticAbilities: []loweredStaticAbility{{Body: handGrant}},
			consumed: semanticConsumption{
				effects:  1,
				keywords: len(ability.Keywords),
			},
			sourceSpans: handCyclingGrantSourceSpans(ability, syntax),
		}, nil
	}
	if lowered, ok, diagnostic := lowerKeywordDispatch(ability, syntax); ok {
		return lowered, diagnostic
	}
	if lowered, ok, diagnostic := lowerStaticRuleDeclaration(ability); ok {
		return lowered, diagnostic
	}
	if staticBuff, ok, diagnostic := lowerStaticPTBuff(ability, syntax); ok {
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		consumedReferences := 0
		if len(ability.References) == 1 &&
			(ability.References[0].Kind == oracle.ReferenceSelfName ||
				ability.References[0].Kind == oracle.ReferenceThisObject) {
			consumedReferences = 1
		}
		return abilityLowering{
			staticAbilities: []loweredStaticAbility{{Body: staticBuff}},
			consumed: semanticConsumption{
				effects:    1,
				keywords:   len(ability.Keywords),
				references: consumedReferences,
			},
			sourceSpans: staticPTBuffSourceSpans(ability, syntax),
		}, nil
	}
	if keywordGrant, ok, diagnostic := lowerStaticKeywordGrant(ability, syntax); ok {
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		return abilityLowering{
			staticAbilities: []loweredStaticAbility{{Body: keywordGrant}},
			consumed: semanticConsumption{
				effects:  1,
				keywords: len(ability.Keywords),
			},
			sourceSpans: staticKeywordGrantSourceSpans(ability, syntax),
		}, nil
	}
	if keywordGrant, ok := lowerSourceConditionalKeywordGrant(ability, syntax); ok {
		return abilityLowering{
			staticAbilities: []loweredStaticAbility{{Body: keywordGrant}},
			consumed: semanticConsumption{
				conditions: 1,
				effects:    1,
				keywords:   len(ability.Keywords),
				references: 1,
			},
			sourceSpans: sourceConditionalKeywordGrantSpans(ability, syntax),
		}, nil
	}
	switch ability.Kind {
	case oracle.AbilityStatic:
		bodies, diagnostic := lowerKeywordAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}

		spans := make([]oracle.Span, 0, len(ability.Keywords)+len(syntax.Reminders))
		if syntax.AbilityWord != nil && len(ability.Keywords) > 0 {
			spans = append(spans, oracle.Span{
				Start: ability.Span.Start,
				End:   ability.Keywords[0].Span.Start,
			})
		}
		spans = appendKeywordSpans(spans, ability.Keywords)
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			staticAbilities: bodies,
			consumed: semanticConsumption{
				keywords: len(ability.Keywords),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityActivated:
		return lowerActivatedAbilityKind(cardName, ability, syntax)
	case oracle.AbilityLoyalty:
		return lowerLoyaltyAbility(cardName, ability, syntax)
	case oracle.AbilitySpell:
		spellAbility, diagnostic := lowerSpell(cardName, ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := make(
			[]oracle.Span,
			0,
			len(ability.Effects)+len(ability.Targets)+len(ability.References)+len(syntax.Reminders),
		)
		for _, effect := range ability.Effects {
			spans = append(spans, effect.Span)
		}
		for _, target := range ability.Targets {
			spans = append(spans, target.Span)
		}
		for _, reference := range ability.References {
			spans = append(spans, reference.Span)
		}
		spans = appendKeywordSpans(spans, ability.Keywords)
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			spellAbility: opt.Val(spellAbility),
			consumed: semanticConsumption{
				targets:    len(ability.Targets),
				effects:    len(ability.Effects),
				keywords:   len(ability.Keywords),
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityTriggered:
		return lowerTriggeredAbilityKind(cardName, ability, syntax)
	case oracle.AbilityChapter:
		return lowerChapterAbility(cardName, ability, syntax)
	case oracle.AbilityReplacement:
		return lowerReplacementAbility(ability)
	case oracle.AbilityReminder:
		if saga && isOrdinarySagaReminder(ability.Text) {
			return abilityLowering{sourceSpans: []oracle.Span{ability.Span}}, nil
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

func lowerReplacementAbility(ability oracle.CompiledAbility) (abilityLowering, *oracle.Diagnostic) {
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

func replacementAbilityLowering(ability oracle.CompiledAbility, replacementAbility *game.ReplacementAbility, diagnostic *oracle.Diagnostic) (abilityLowering, *oracle.Diagnostic) {
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	return abilityLowering{
		replacementAbility: opt.Val(*replacementAbility),
		consumed: semanticConsumption{
			effects:    len(ability.Effects),
			conditions: len(ability.Conditions),
			references: len(ability.References),
		},
		sourceSpans: replacementSourceSpans(ability),
	}, nil
}

func appendKeywordSpans(spans []oracle.Span, keywords []oracle.CompiledKeyword) []oracle.Span {
	for _, keyword := range keywords {
		spans = append(spans, keyword.Span)
	}
	return spans
}

func replacementSourceSpans(ability oracle.CompiledAbility) []oracle.Span {
	spans := make([]oracle.Span, 0, len(ability.Effects))
	for _, effect := range ability.Effects {
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	if len(ability.Chapters) == 0 || ability.ChapterSpan == (oracle.Span{}) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires one or more chapter numbers",
		)
	}
	bodyAbility := ability
	bodyAbility.Kind = oracle.AbilitySpell
	bodyAbility.Chapters = nil
	bodyAbility.ChapterSpan = oracle.Span{}
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Chapters = nil
	bodySyntax.ChapterSpan = oracle.Span{}
	dash := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.EmDash
	})
	if dash < 0 || dash+1 >= len(syntax.Tokens) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires an em dash after the chapter numbers",
		)
	}
	bodyAbility.Span = oracle.Span{
		Start: syntax.Tokens[dash+1].Span.Start,
		End:   syntax.Span.End,
	}
	bodyAbility.Text = strings.TrimSpace(
		ability.Text[bodyAbility.Span.Start.Offset-ability.Span.Start.Offset:],
	)
	bodyAbility.Keywords = keywordsWithinSpan(ability.Keywords, bodyAbility.Span)
	if len(bodyAbility.Keywords) != len(ability.Keywords) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires chapter keywords to belong to a supported effect",
		)
	}
	bodySyntax.Span = bodyAbility.Span
	bodySyntax.Text = bodyAbility.Text
	bodySyntax.Tokens = slices.Clone(syntax.Tokens[dash+1:])
	content, diagnostic := lowerSpell(cardName, bodyAbility, bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			diagnostic.Detail,
		)
	}
	spans := []oracle.Span{ability.ChapterSpan, syntax.Tokens[dash].Span}
	for _, effect := range ability.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.References {
		spans = append(spans, reference.Span)
	}
	for _, keyword := range ability.Keywords {
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
			targets:    len(ability.Targets),
			effects:    len(ability.Effects),
			keywords:   len(ability.Keywords),
			references: len(ability.References),
		},
		sourceSpans: spans,
	}, nil
}

func lowerEntersPrepared(ability oracle.CompiledAbility, syntax oracle.Ability) (abilityLowering, bool) {
	const text = "This creature enters prepared."
	if ability.Kind != oracle.AbilityStatic ||
		(ability.Text != text && !strings.HasPrefix(ability.Text, text+" (")) ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectEnterPrepared ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferenceThisObject ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
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
		sourceSpans: []oracle.Span{syntax.Span},
	}, true
}

func lowerActivatedAbilityKind(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	if hasAddManaEffect(ability) {
		manaAbility, diagnostic := lowerTapManaAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []oracle.Span{ability.Cost.Span, ability.Effects[0].Span}
		if ability.ActivationTiming != oracle.ActivationTimingNone {
			spans = append(spans, ability.ActivationTimingSpan)
		}
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed: semanticConsumption{
				cost:    true,
				effects: 1,
			},
			sourceSpans: spans,
		}, nil
	}
	activatedAbility, diagnostic := lowerActivatedAbility(cardName, ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := make(
		[]oracle.Span,
		0,
		1+len(ability.Effects)+len(ability.Targets)+len(ability.References)+len(syntax.Reminders),
	)
	spans = append(spans, ability.Cost.Span)
	if ability.ActivationTiming != oracle.ActivationTimingNone {
		spans = append(spans, ability.ActivationTimingSpan)
	}
	for _, effect := range ability.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		activatedAbility: opt.Val(activatedAbility),
		consumed: semanticConsumption{
			cost:       true,
			targets:    len(ability.Targets),
			effects:    len(ability.Effects),
			keywords:   len(ability.Keywords),
			references: len(ability.References),
		},
		sourceSpans: spans,
	}, nil
}

func hasAddManaEffect(ability oracle.CompiledAbility) bool {
	return slices.ContainsFunc(ability.Effects, func(effect oracle.CompiledEffect) bool {
		return effect.Kind == oracle.EffectAddMana
	})
}

// lowerLoyaltyAbility lowers an AbilityLoyalty into a game.LoyaltyAbility.
// It accepts only exact signed integer loyalty costs and supported single or
// ordered effect bodies. Variable costs (X) are rejected.
func lowerLoyaltyAbility(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	const unsupportedDetail = "the executable source backend supports only exact signed loyalty costs with a supported effect body"
	if ability.Cost == nil ||
		len(ability.Cost.Components) != 1 ||
		ability.Cost.Components[0].Kind != oracle.CostLoyalty ||
		len(ability.Conditions) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	loyaltyCost, ok := parseLoyaltyCostAmount(ability.Cost.Components[0].Amount)
	if !ok {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", "the executable source backend supports only fixed integer loyalty costs, not variable costs")
	}

	colon := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.Colon
	})
	if colon < 0 || colon+1 >= len(syntax.Tokens) {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	body := ability
	body.Kind = oracle.AbilitySpell
	body.Cost = nil
	body.Span = oracle.Span{
		Start: syntax.Tokens[colon+1].Span.Start,
		End:   syntax.Span.End,
	}
	body.Text = strings.TrimSpace(ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset:])
	body.Keywords = keywordsWithinSpan(ability.Keywords, body.Span)
	if len(body.Keywords) != len(ability.Keywords) {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	bodySyntax.Tokens = syntax.Tokens[colon+1:]
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", diagnostic.Detail)
	}

	spans := make(
		[]oracle.Span,
		0,
		1+len(ability.Effects)+len(ability.Targets)+len(ability.References)+len(syntax.Reminders),
	)
	spans = append(spans, ability.Cost.Span)
	for _, effect := range ability.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Keywords)
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
			targets:    len(ability.Targets),
			effects:    len(ability.Effects),
			keywords:   len(ability.Keywords),
			references: len(ability.References),
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

// lowerModalAbility lowers a modal CompiledAbility into a spell or activated
// ability with multiple modes. Only "Choose one —" is supported; other
// cardinalities are rejected. Each mode is lowered independently through the
// shared spell-effect lowering path.
func lowerModalAbility(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	if syntax.Modal == nil {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not yet lower modal abilities",
		)
	}
	minModes, maxModes, ok := parseChooseHeader(syntax.Modal.Header)
	if !ok {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend supports only exact \"Choose N\" and \"Choose one or both\" modal abilities",
		)
	}
	if minModes < 1 || maxModes < minModes || maxModes > len(ability.Modes) ||
		(minModes == 1 && maxModes == 2 && len(ability.Modes) != 2) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the modal choice range does not match the number of modes",
		)
	}

	// Top-level semantic fields must be empty for a modal header: the header
	// "Choose one —" carries no targets, effects, keywords, or conditions of its own.
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		len(ability.Effects) != 0 ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.References) != 0 ||
		ability.AbilityWord != "" {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not support shared targets, costs, or conditions across modes",
		)
	}
	if len(ability.Modes) != len(syntax.Modal.Options) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"semantic mode count does not match syntax mode count",
		)
	}

	modes := make([]game.Mode, 0, len(ability.Modes))
	for i, compiledMode := range ability.Modes {
		syntaxMode := syntax.Modal.Options[i]
		mode, diagnostic := lowerModalMode(cardName, compiledMode, syntaxMode)
		if diagnostic != nil {
			return abilityLowering{}, executableDiagnostic(
				ability,
				"unsupported modal ability",
				diagnostic.Detail,
			)
		}
		modes = append(modes, mode)
	}

	content := game.AbilityContent{
		Modes:    modes,
		MinModes: minModes,
		MaxModes: maxModes,
	}
	switch ability.Kind {
	case oracle.AbilitySpell, oracle.AbilityStatic:
	default:
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend supports only spell or static modal abilities",
		)
	}
	return abilityLowering{
		spellAbility: opt.Val(content),
		consumed: semanticConsumption{
			modes: len(ability.Modes),
		},
		sourceSpans: []oracle.Span{syntax.Modal.Header.Span},
	}, nil
}

// parseChooseHeader inspects a modal header phrase and returns (minModes,
// maxModes, ok). It accepts "Choose <word> —" where <word> is a cardinal
// number spelled out as a single word ("one", "two", etc.), plus exact
// "Choose one or both —" headers.
func parseChooseHeader(header oracle.Phrase) (minModes, maxModes int, ok bool) {
	tokens := header.Tokens
	if len(tokens) == 5 &&
		tokens[0].Kind == oracle.Word && strings.EqualFold(tokens[0].Text, "choose") &&
		tokens[1].Kind == oracle.Word && strings.EqualFold(tokens[1].Text, "one") &&
		tokens[2].Kind == oracle.Word && strings.EqualFold(tokens[2].Text, "or") &&
		tokens[3].Kind == oracle.Word && strings.EqualFold(tokens[3].Text, "both") &&
		tokens[4].Kind == oracle.EmDash {
		return 1, 2, true
	}
	// Expected: [Word("Choose"), Word(<number>), EmDash]
	if len(tokens) != 3 ||
		tokens[0].Kind != oracle.Word || !strings.EqualFold(tokens[0].Text, "choose") ||
		tokens[1].Kind != oracle.Word ||
		tokens[2].Kind != oracle.EmDash {
		return 0, 0, false
	}
	n, numOK := parseCardinalWord(tokens[1].Text)
	if !numOK {
		return 0, 0, false
	}
	return n, n, true
}

// parseCardinalWord converts a lowercase English cardinal number word ("one",
// "two", … "ten") to an integer. Returns (0, false) for unrecognized words.
func parseCardinalWord(word string) (int, bool) {
	switch strings.ToLower(word) {
	case "one":
		return 1, true
	case "two":
		return 2, true
	case "three":
		return 3, true
	case "four":
		return 4, true
	case "five":
		return 5, true
	case "six":
		return 6, true
	case "seven":
		return 7, true
	case "eight":
		return 8, true
	case "nine":
		return 9, true
	case "ten":
		return 10, true
	default:
		return 0, false
	}
}

// lowerModalMode lowers one compiled mode into a game.Mode by routing through
// the shared spell-effect lowering path. Mode-local targets, effects, keywords,
// references, and source spans are all consumed independently.
func lowerModalMode(
	cardName string,
	mode oracle.CompiledMode,
	syntaxMode oracle.Mode,
) (game.Mode, *oracle.Diagnostic) {
	body := oracle.CompiledAbility{
		Kind:       oracle.AbilitySpell,
		Span:       mode.Span,
		Text:       mode.Text,
		Targets:    mode.Targets,
		Conditions: mode.Conditions,
		Effects:    mode.Effects,
		Keywords:   mode.Keywords,
		References: mode.References,
	}
	bodySyntax := oracle.Ability{
		Kind:      oracle.AbilitySpell,
		Span:      syntaxMode.Span,
		Text:      syntaxMode.Text,
		Tokens:    syntaxMode.Tokens,
		Reminders: syntaxMode.Reminders,
		Quoted:    syntaxMode.Quoted,
	}
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.Mode{}, diagnostic
	}
	if content.IsModal() || len(content.Modes) != 1 {
		return game.Mode{}, &oracle.Diagnostic{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported modal ability",
			Detail:   "mode lowering produced unexpected modal content",
			Span:     mode.Span,
		}
	}
	return content.Modes[0], nil
}

func lowerActivatedAbility(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ActivatedAbility, *oracle.Diagnostic) {
	if ability.Cost == nil ||
		len(ability.Cost.Components) == 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}
	var manaCost cost.Mana
	var additionalCosts []cost.Additional
	for i, component := range ability.Cost.Components {
		switch component.Kind {
		case oracle.CostMana:
			if i != 0 || manaCost != nil {
				return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
			}
			parsed, err := parseManaCostValue(component.Symbol)
			if err != nil || len(parsed) == 0 {
				return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
			}
			manaCost = parsed
		case oracle.CostTap:
			if slices.ContainsFunc(additionalCosts, func(additional cost.Additional) bool {
				return additional.Kind == cost.AdditionalTap
			}) {
				return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
			}
			additionalCosts = append(additionalCosts, cost.T)
		case oracle.CostUntap:
			if slices.ContainsFunc(additionalCosts, func(additional cost.Additional) bool {
				return additional.Kind == cost.AdditionalUntap
			}) {
				return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
			}
			additionalCosts = append(additionalCosts, cost.Additional{
				Kind: cost.AdditionalUntap,
				Text: component.Text,
			})
		default:
			additional, ok := lowerActivatedAdditionalCost(cardName, component)
			if !ok {
				return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
			}
			additionalCosts = append(additionalCosts, additional)
		}
	}

	colon := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.Colon
	})
	if colon < 0 || colon+1 >= len(syntax.Tokens) {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}
	bodyTokens := append([]oracle.Token(nil), syntax.Tokens[colon+1:]...)
	if ability.ActivationTiming != oracle.ActivationTimingNone {
		bodyTokens = slices.DeleteFunc(bodyTokens, func(token oracle.Token) bool {
			return spanCovered(token.Span, []oracle.Span{ability.ActivationTimingSpan})
		})
	}
	if len(bodyTokens) == 0 {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}
	body := ability
	body.Kind = oracle.AbilitySpell
	body.Cost = nil
	body.ActivationTiming = oracle.ActivationTimingNone
	body.ActivationTimingSpan = oracle.Span{}
	body.References = bodyReferences(ability.References, ability.Cost.Span)
	body.Span = oracle.Span{
		Start: bodyTokens[0].Span.Start,
		End:   bodyTokens[len(bodyTokens)-1].Span.End,
	}
	body.Text = strings.TrimSpace(ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset])
	body.Keywords = keywordsWithinSpan(ability.Keywords, body.Span)
	if len(body.Keywords) != len(ability.Keywords) {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	bodySyntax.Tokens = bodyTokens
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}

	result := game.ActivatedAbility{
		Text:            ability.Text,
		AdditionalCosts: additionalCosts,
		ZoneOfFunction:  lowerActivatedAbilityZoneOfFunction(body),
		Timing:          lowerActivationTiming(ability.ActivationTiming),
		Content:         content,
	}
	if manaCost != nil {
		result.ManaCost = opt.Val(manaCost)
	}
	return result, nil
}

func lowerActivatedAbilityZoneOfFunction(body oracle.CompiledAbility) zone.Type {
	if selfCardGraveyardReturnReferences(body.References) &&
		strings.HasPrefix(body.Text, "Return this card from your graveyard ") {
		return zone.Graveyard
	}
	return zone.Battlefield
}

func lowerActivationTiming(timing oracle.ActivationTimingKind) game.TimingRestriction {
	switch timing {
	case oracle.ActivationTimingNone:
		return game.NoTimingRestriction
	case oracle.ActivationTimingSorcery:
		return game.SorceryOnly
	case oracle.ActivationTimingOncePerTurn:
		return game.OncePerTurn
	case oracle.ActivationTimingSorceryOncePerTurn:
		return game.SorceryOncePerTurn
	case oracle.ActivationTimingDuringCombat:
		return game.DuringCombat
	case oracle.ActivationTimingDuringUpkeep:
		return game.DuringUpkeep
	default:
		panic(fmt.Sprintf("unknown activation timing %d", timing))
	}
}

func lowerActivatedAdditionalCost(cardName string, component oracle.CostComponent) (cost.Additional, bool) {
	switch component.Kind {
	case oracle.CostSacrifice:
		return lowerSacrificeCost(cardName, component)
	case oracle.CostDiscard:
		return lowerDiscardCost(component)
	case oracle.CostPayLife:
		amount, err := strconv.Atoi(component.Amount)
		if err != nil || amount <= 0 {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalPayLife,
			Text:   component.Text,
			Amount: amount,
		}, true
	case oracle.CostExile:
		if isSelfCostObject(cardName, component.Object) {
			return cost.Additional{
				Kind:   cost.AdditionalExileSource,
				Text:   component.Text,
				Amount: 1,
				Source: zone.Battlefield,
			}, true
		}
		return lowerExileCost(component)
	case oracle.CostRemoveCounter:
		return lowerRemoveCounterCost(cardName, component)
	case oracle.CostTapPermanents:
		return lowerTapPermanentsCost(component)
	case oracle.CostEnergy:
		amount, err := strconv.Atoi(component.Amount)
		if err != nil || amount <= 0 {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalEnergy,
			Text:   component.Text,
			Amount: amount,
		}, true
	default:
		return cost.Additional{}, false
	}
}

func lowerTapPermanentsCost(component oracle.CostComponent) (cost.Additional, bool) {
	words := strings.Fields(component.Object)
	if len(words) < 5 {
		return cost.Additional{}, false
	}
	amount, ok := exactCostAmount(strings.ToLower(words[0]))
	if !ok {
		return cost.Additional{}, false
	}
	if !strings.EqualFold(words[1], "untapped") ||
		!strings.EqualFold(words[len(words)-2], "you") ||
		!strings.EqualFold(words[len(words)-1], "control") {
		return cost.Additional{}, false
	}
	object := strings.Join(words[2:len(words)-2], " ")
	additional := cost.Additional{
		Kind:   cost.AdditionalTapPermanents,
		Text:   component.Text,
		Amount: amount,
	}
	if lowerTapPermanentsObject(object, &additional) {
		return additional, true
	}
	return cost.Additional{}, false
}

func lowerTapPermanentsObject(object string, additional *cost.Additional) bool {
	normalized := strings.ToLower(strings.TrimSpace(object))
	switch strings.TrimSuffix(normalized, "s") {
	case "permanent":
		return true
	case "artifact":
		additional.MatchPermanentType = true
		additional.PermanentType = types.Artifact
		return true
	case "creature":
		additional.MatchPermanentType = true
		additional.PermanentType = types.Creature
		return true
	case "enchantment":
		additional.MatchPermanentType = true
		additional.PermanentType = types.Enchantment
		return true
	case "land":
		additional.MatchPermanentType = true
		additional.PermanentType = types.Land
		return true
	default:
	}
	subtype, ok := tapPermanentsSubtype(object)
	if !ok {
		return false
	}
	additional.SubtypesAny = cost.SubtypeSet{subtype}
	return true
}

func tapPermanentsSubtype(object string) (types.Sub, bool) {
	candidates := []string{
		strings.TrimSpace(object),
		singularCostNoun(object),
	}
	for _, candidate := range candidates {
		subtype := types.Sub(candidate)
		if types.KnownSubtypeForType(types.Creature, subtype) ||
			types.KnownSubtypeForType(types.Artifact, subtype) {
			return subtype, true
		}
	}
	return "", false
}

func singularCostNoun(noun string) string {
	noun = strings.TrimSpace(noun)
	switch {
	case strings.HasSuffix(noun, "ies") && len(noun) > 3:
		return noun[:len(noun)-3] + "y"
	case strings.HasSuffix(noun, "ves") && len(noun) > 3:
		return noun[:len(noun)-3] + "f"
	case strings.HasSuffix(noun, "s") && len(noun) > 1:
		return noun[:len(noun)-1]
	default:
		return noun
	}
}

func lowerRemoveCounterCost(
	cardName string,
	component oracle.CostComponent,
) (cost.Additional, bool) {
	object := strings.ToLower(strings.TrimSpace(component.Object))
	amount, rest, ok := removeCounterCostAmount(object)
	if !ok {
		return cost.Additional{}, false
	}
	kind, source, ok := removeCounterCostKindAndSource(rest)
	if !ok || source != "it" && !isSelfCostObject(cardName, source) {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:        cost.AdditionalRemoveCounter,
		Text:        component.Text,
		Amount:      amount,
		CounterKind: kind,
	}, true
}

func removeCounterCostAmount(object string) (amount int, rest string, ok bool) {
	object = strings.TrimSpace(object)
	if strings.HasPrefix(object, "any number of ") || strings.HasPrefix(object, "x ") {
		return 0, "", false
	}
	amountWord, rest, ok := strings.Cut(object, " ")
	if !ok {
		return 0, "", false
	}
	amount, ok = exactCostAmount(amountWord)
	if !ok {
		return 0, "", false
	}
	return amount, strings.TrimSpace(rest), true
}

func removeCounterCostKindAndSource(rest string) (counter.Kind, string, bool) {
	for _, candidate := range removeCounterCostKinds() {
		for _, counterWord := range []string{" counter from ", " counters from "} {
			prefix := candidate.name + counterWord
			if !strings.HasPrefix(rest, prefix) {
				continue
			}
			return candidate.kind, strings.TrimSpace(strings.TrimPrefix(rest, prefix)), true
		}
	}
	return 0, "", false
}

func removeCounterCostKinds() []struct {
	name string
	kind counter.Kind
} {
	return []struct {
		name string
		kind counter.Kind
	}{
		{"+1/+1", counter.PlusOnePlusOne},
		{"-1/-1", counter.MinusOneMinusOne},
		{"loyalty", counter.Loyalty},
		{"charge", counter.Charge},
		{"storage", counter.Charge},
		{"fuse", counter.Charge},
		{"time", counter.Time},
		{"defense", counter.Defense},
		{"lore", counter.Lore},
		{"verse", counter.Verse},
		{"shield", counter.Shield},
		{"stun", counter.Stun},
		{"finality", counter.Finality},
		{"brick", counter.Brick},
		{"page", counter.Page},
		{"enlightened", counter.Enlightened},
		{"oil", counter.Oil},
		{"blood", counter.Blood},
		{"indestructible", counter.Indestructible},
		{"deathtouch", counter.Deathtouch},
		{"flying", counter.Flying},
		{"first strike", counter.FirstStrike},
		{"hexproof", counter.Hexproof},
		{"lifelink", counter.Lifelink},
		{"menace", counter.Menace},
		{"reach", counter.Reach},
		{"trample", counter.Trample},
		{"vigilance", counter.Vigilance},
	}
}

func lowerExileCost(component oracle.CostComponent) (cost.Additional, bool) {
	additional := cost.Additional{
		Kind:   cost.AdditionalExile,
		Text:   component.Text,
		Amount: 1,
		Source: zone.Graveyard,
	}
	switch strings.ToLower(strings.TrimSpace(component.Object)) {
	case "a card from your graveyard":
		return additional, true
	case "a creature card from your graveyard":
		additional.MatchCardType = true
		additional.CardType = types.Creature
		return additional, true
	case "an artifact card from your graveyard":
		additional.MatchCardType = true
		additional.CardType = types.Artifact
		return additional, true
	case "a land card from your graveyard":
		additional.MatchCardType = true
		additional.CardType = types.Land
		return additional, true
	case "two cards from your graveyard":
		additional.Amount = 2
		return additional, true
	default:
		return cost.Additional{}, false
	}
}

func lowerSacrificeCost(cardName string, component oracle.CostComponent) (cost.Additional, bool) {
	if isSelfCostObject(cardName, component.Object) {
		return cost.Additional{
			Kind:   cost.AdditionalSacrificeSource,
			Text:   component.Text,
			Amount: 1,
		}, true
	}
	spec, ok := exactCostObject(component.Object, false)
	if !ok {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:               cost.AdditionalSacrifice,
		Text:               component.Text,
		Amount:             spec.amount,
		MatchPermanentType: spec.matchesType,
		PermanentType:      spec.objectType,
	}, true
}

func lowerDiscardCost(component oracle.CostComponent) (cost.Additional, bool) {
	spec, ok := exactCostObject(component.Object, true)
	if !ok {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:          cost.AdditionalDiscard,
		Text:          component.Text,
		Amount:        spec.amount,
		MatchCardType: spec.matchesType,
		CardType:      spec.objectType,
		Source:        zone.Hand,
	}, true
}

func isSelfCostObject(cardName, object string) bool {
	normalized := strings.ToLower(strings.TrimSpace(object))
	switch normalized {
	case "this artifact", "this creature", "this enchantment", "this land", "this permanent", "this token":
		return true
	default:
		return strings.EqualFold(strings.TrimSpace(object), cardName)
	}
}

type costObjectSpec struct {
	amount      int
	objectType  types.Card
	matchesType bool
}

func exactCostObject(object string, cardObject bool) (costObjectSpec, bool) {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(object)))
	if cardObject {
		if len(words) < 2 || words[len(words)-1] != "card" && words[len(words)-1] != "cards" {
			return costObjectSpec{}, false
		}
		words = words[:len(words)-1]
	}
	if len(words) == 4 {
		if words[2] != "you" || words[3] != "control" {
			return costObjectSpec{}, false
		}
		words = words[:2]
	}
	if cardObject && len(words) == 1 {
		parsedAmount, valid := exactCostAmount(words[0])
		return costObjectSpec{amount: parsedAmount}, valid
	}
	if len(words) != 2 {
		return costObjectSpec{}, false
	}
	parsedAmount, valid := exactCostAmount(words[0])
	if !valid {
		return costObjectSpec{}, false
	}
	spec := costObjectSpec{amount: parsedAmount, matchesType: true}
	noun := strings.TrimSuffix(words[1], "s")
	switch noun {
	case "permanent":
		if cardObject {
			return costObjectSpec{}, false
		}
		spec.matchesType = false
	case "artifact":
		spec.objectType = types.Artifact
	case "creature":
		spec.objectType = types.Creature
	case "enchantment":
		spec.objectType = types.Enchantment
	case "land":
		spec.objectType = types.Land
	default:
		return costObjectSpec{}, false
	}
	return spec, true
}

func exactCostAmount(word string) (int, bool) {
	switch word {
	case "a", "an", "one":
		return 1, true
	default:
		if amount, ok := parseCardinalWord(word); ok {
			return amount, true
		}
		amount, err := strconv.Atoi(word)
		return amount, err == nil && amount > 0
	}
}

func unsupportedActivatedAbilityDiagnostic(ability oracle.CompiledAbility) *oracle.Diagnostic {
	return executableDiagnostic(
		ability,
		"unsupported activated ability",
		"the executable source backend supports only exact typed costs with a supported effect",
	)
}

func lowerEnchantAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Enchant" {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	target, ok := enchantTargetSpec(keyword.Parameter)
	if !ok ||
		ability.Kind != oracle.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Enchant ability",
			"the executable source backend supports only exact Enchant with a supported target kind",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Protection" {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	protectedColors, ok := oracleColors(keyword.Parameter)
	if !ok ||
		ability.Kind != oracle.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Protection ability",
			"the executable source backend supports only exact protection from colors",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Protection ability",
			"the executable source backend supports only exact protection from colors",
		)
	}
	return game.ProtectionFromColorsStaticAbility(protectedColors...), true, nil
}

// lowerKeywordDispatch tries Enchant, Protection, Equip, Cycling, Ninjutsu, and
// Mutate — the
// single-keyword special cases that each produce a full abilityLowering.
// Returns (lowering, true, nil) on success, (lowering, true, diag) on a
// recognized-but-rejected attempt, and ({}, false, nil) when no attempt matches.
func lowerKeywordDispatch(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, bool, *oracle.Diagnostic) {
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) abilityLowering {
	spans := keywordSpans(ability, syntax)
	return abilityLowering{
		activatedAbility: opt.Val(*body),
		consumed:         semanticConsumption{keywords: 1},
		sourceSpans:      spans,
	}
}

func keywordSpans(ability oracle.CompiledAbility, syntax oracle.Ability) []oracle.Span {
	spans := []oracle.Span{ability.Keywords[0].Span}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

func oracleColors(parameter string) ([]color.Color, bool) {
	names := strings.Split(parameter, ",")
	colors := make([]color.Color, 0, len(names))
	seen := make(map[color.Color]struct{}, len(names))
	for _, name := range names {
		oracleColor, ok := oracleColor(name)
		if !ok {
			return nil, false
		}
		if _, ok := seen[oracleColor]; ok {
			return nil, false
		}
		seen[oracleColor] = struct{}{}
		colors = append(colors, oracleColor)
	}
	return colors, len(colors) > 0
}

func oracleColor(name string) (color.Color, bool) {
	switch name {
	case "white":
		return color.White, true
	case "blue":
		return color.Blue, true
	case "black":
		return color.Black, true
	case "red":
		return color.Red, true
	case "green":
		return color.Green, true
	default:
		return "", false
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ActivatedAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Equip" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	if keyword.Parameter == "" ||
		ability.Kind != oracle.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
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
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ActivatedAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Cycling" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	if keyword.Parameter == "" && (len(ability.Targets) != 0 || len(ability.Effects) != 0 || len(ability.References) != 0) {
		return game.ActivatedAbility{}, false, nil
	}
	if keyword.Parameter == "" ||
		(ability.Kind != oracle.AbilityStatic && ability.Kind != oracle.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
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
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
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

func lowerHandCyclingGrant(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if ability.Kind != oracle.AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectGrantKeyword ||
		ability.Effects[0].Duration != oracle.DurationNone ||
		len(ability.Keywords) != 1 ||
		ability.Keywords[0].Name != "Cycling" ||
		ability.Keywords[0].Parameter == "" ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported hand Cycling grant",
			"the executable source backend supports only exact hand-card Cycling grants with a mana cost",
		)
	}
	semanticText := abilityTextWithoutReminders(syntax)
	selection, ok := handCyclingGrantSelection(semanticText, keyword.Parameter)
	if !ok {
		if semanticText == fmt.Sprintf("Each historic card in your hand has cycling %s.", keyword.Parameter) {
			return game.StaticAbility{}, true, executableDiagnostic(
				ability,
				"unsupported hand Cycling grant",
				"historic card predicates are not supported by the executable source backend",
			)
		}
		return game.StaticAbility{}, false, nil
	}
	return game.StaticAbility{
		Text: semanticText,
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectGrantHandCardAbility,
			AffectedPlayer: game.PlayerYou,
			CardSelection:  selection,
			GrantedAbility: game.CyclingActivatedAbility(manaCost),
		}},
	}, true, nil
}

func handCyclingGrantSelection(text, parameter string) (game.Selection, bool) {
	switch text {
	case fmt.Sprintf("Each land card in your hand has cycling %s.", parameter):
		return game.Selection{RequiredTypes: []types.Card{types.Land}}, true
	case fmt.Sprintf("Each creature card in your hand has cycling %s.", parameter):
		return game.Selection{RequiredTypes: []types.Card{types.Creature}}, true
	default:
		return game.Selection{}, false
	}
}

func abilityTextWithoutReminders(syntax oracle.Ability) string {
	var b strings.Builder
	prev := ""
	for _, token := range syntax.Tokens {
		if spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		text := token.Text
		if text == "." || text == "," || text == ";" || text == ":" {
			_, _ = b.WriteString(text)
			prev = text
			continue
		}
		if b.Len() > 0 && (!strings.HasPrefix(prev, "{") || !strings.HasPrefix(text, "{")) {
			_ = b.WriteByte(' ')
		}
		_, _ = b.WriteString(text)
		prev = text
	}
	return b.String()
}

func handCyclingGrantSourceSpans(ability oracle.CompiledAbility, syntax oracle.Ability) []oracle.Span {
	spans := make([]oracle.Span, 0, 1+len(ability.Keywords)+len(syntax.Reminders))
	if len(ability.Effects) > 0 {
		spans = append(spans, ability.Effects[0].Span)
	}
	for _, keyword := range ability.Keywords {
		spans = append(spans, keyword.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

func lowerNinjutsuAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ActivatedAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Ninjutsu" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	if keyword.Parameter == "" ||
		(ability.Kind != oracle.AbilityStatic && ability.Kind != oracle.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
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
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Mutate" {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	if keyword.Parameter == "" ||
		(ability.Kind != oracle.AbilityStatic && ability.Kind != oracle.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
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
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
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

func lowerStaticPTBuff(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if ability.Kind != oracle.AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectModifyPT ||
		ability.Effects[0].Duration != oracle.DurationNone ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, false, nil
	}
	effect := ability.Effects[0]
	sourceSelf := effect.StaticSubject == oracle.StaticSubjectNone &&
		len(ability.References) == 1 &&
		(ability.References[0].Kind == oracle.ReferenceSelfName ||
			ability.References[0].Kind == oracle.ReferenceThisObject)
	if (len(ability.References) != 0 && !sourceSelf) ||
		(effect.StaticSubject == oracle.StaticSubjectNone && !sourceSelf) {
		return game.StaticAbility{}, false, nil
	}
	dynamicPT := effect.Amount.DynamicKind != oracle.DynamicAmountNone
	if (!dynamicPT && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) ||
		(dynamicPT && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known ||
			!dynamicPTMultiplierMatches(effect.Amount.Multiplier, effect.PowerDelta, effect.ToughnessDelta))) {
		return game.StaticAbility{}, false, nil
	}
	keywordsForBuff := abilityKeywordsExcludingSelectorPredicates(ability)
	keywords, keywordsOK := mixedStaticKeywords(keywordsForBuff)
	syntaxOK := matchesExactStaticPTBuffSyntax(syntax, effect)
	if sourceSelf {
		syntaxOK = matchesExactSourceStaticPTBuffSyntax(syntax, effect, ability.References[0])
	}
	if !keywordsOK ||
		(len(keywords) == 0 && !syntaxOK) ||
		(len(keywords) > 0 && !matchesExactStaticPTBuffWithKeywordsSyntax(syntax, effect, keywordsForBuff)) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend supports only exact fixed static creature power/toughness buffs, optionally granting supported keywords",
		)
	}
	group := game.GroupReference{}
	if !sourceSelf {
		var ok bool
		group, ok = staticSubjectGroup(effect.StaticSubject, effect.StaticSubjectSubtype)
		if !ok {
			return game.StaticAbility{}, false, nil
		}
	}
	continuousEffects := []game.ContinuousEffect{{
		Layer:          game.LayerPowerToughnessModify,
		AffectedSource: sourceSelf,
		Group:          group,
		PowerDelta:     compiledSignedAmountValue(effect.PowerDelta),
		ToughnessDelta: compiledSignedAmountValue(effect.ToughnessDelta),
	}}
	if dynamicPT {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == oracle.DynamicAmountSourcePower {
			return game.StaticAbility{}, true, executableDiagnostic(
				ability,
				"unsupported static ability",
				"the executable source backend supports only exact supported static creature power/toughness buffs",
			)
		}
		continuousEffects[0].PowerDelta = 0
		continuousEffects[0].ToughnessDelta = 0
		if powerDelta := dynamicSignedQuantity(dynamic, effect.PowerDelta); powerDelta.IsDynamic() {
			continuousEffects[0].PowerDeltaDynamic = powerDelta.DynamicAmount()
		}
		if toughnessDelta := dynamicSignedQuantity(dynamic, effect.ToughnessDelta); toughnessDelta.IsDynamic() {
			continuousEffects[0].ToughnessDeltaDynamic = toughnessDelta.DynamicAmount()
		}
	}
	if len(keywords) > 0 {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:       game.LayerAbility,
			Group:       group,
			AddKeywords: keywords,
		})
	}
	return game.StaticAbility{
		Text:              ability.Text,
		ContinuousEffects: continuousEffects,
	}, true, nil
}

func staticPTBuffSourceSpans(ability oracle.CompiledAbility, syntax oracle.Ability) []oracle.Span {
	spans := make([]oracle.Span, 0, 1+len(ability.Keywords)+len(syntax.Reminders))
	spans = append(spans, ability.Effects[0].Span)
	for _, keyword := range ability.Keywords {
		spans = append(spans, keyword.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

func lowerStaticKeywordGrant(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if ability.Kind != oracle.AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectGrantKeyword ||
		ability.Effects[0].Duration != oracle.DurationNone ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].StaticSubject == oracle.StaticSubjectNone ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, false, nil
	}
	effect := ability.Effects[0]
	keywords, keywordsOK := mixedStaticKeywords(ability.Keywords)
	if !keywordsOK || len(keywords) == 0 || !matchesExactStaticKeywordGrantSyntax(syntax, effect, ability.Keywords) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend supports only exact standalone grants of runtime-supported keywords",
		)
	}
	group, ok := staticSubjectGroup(effect.StaticSubject, effect.StaticSubjectSubtype)
	if !ok {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend supports only known creature subtypes in standalone keyword grants",
		)
	}
	return game.StaticAbility{
		Text: ability.Text,
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			Group:       group,
			AddKeywords: keywords,
		}},
	}, true, nil
}

func staticKeywordGrantSourceSpans(ability oracle.CompiledAbility, syntax oracle.Ability) []oracle.Span {
	spans := make([]oracle.Span, 0, 1+len(ability.Keywords)+len(syntax.Reminders))
	spans = append(spans, ability.Effects[0].Span)
	for _, keyword := range ability.Keywords {
		spans = append(spans, keyword.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

func lowerSourceConditionalKeywordGrant(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool) {
	if ability.Kind != oracle.AbilityStatic ||
		len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != oracle.ConditionAsLongAs ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectGrantKeyword ||
		ability.Effects[0].Duration != oracle.DurationNone ||
		ability.Effects[0].StaticSubject != oracle.StaticSubjectNone ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferenceThisObject ||
		len(ability.Targets) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, false
	}
	condition, ok := lowerSourceAsLongAsCondition(ability.Conditions[0])
	if !ok {
		return game.StaticAbility{}, false
	}
	keywords, ok := mixedStaticKeywords(ability.Keywords)
	if !ok || len(keywords) == 0 ||
		!matchesExactSourceConditionalKeywordGrantSyntax(syntax, ability.Conditions[0], ability.Keywords) {
		return game.StaticAbility{}, false
	}
	return game.StaticAbility{
		Text:      ability.Text,
		Condition: opt.Val(condition),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:          game.LayerAbility,
			AffectedSource: true,
			AddKeywords:    keywords,
		}},
	}, true
}

func lowerSourceAsLongAsCondition(condition oracle.CompiledCondition) (game.Condition, bool) {
	const (
		aPrefix       = "as long as you control a "
		anPrefix      = "as long as you control an "
		anotherPrefix = "as long as you control another "
	)
	text := strings.ToLower(condition.Text)
	var noun string
	excludeSource := false
	switch {
	case strings.HasPrefix(text, aPrefix):
		noun = condition.Text[len(aPrefix):]
	case strings.HasPrefix(text, anPrefix):
		noun = condition.Text[len(anPrefix):]
	case strings.HasPrefix(text, anotherPrefix):
		noun = condition.Text[len(anotherPrefix):]
		excludeSource = true
	default:
		return game.Condition{}, false
	}
	if noun == "" || strings.TrimSpace(noun) != noun {
		return game.Condition{}, false
	}
	filter := game.PermanentFilter{ExcludeSource: excludeSource}
	switch strings.ToLower(noun) {
	case "artifact":
		filter.Types = []types.Card{types.Artifact}
	case "artifact creature":
		filter.Types = []types.Card{types.Artifact, types.Creature}
	case "battle":
		filter.Types = []types.Card{types.Battle}
	case "creature":
		filter.Types = []types.Card{types.Creature}
	case "enchantment":
		filter.Types = []types.Card{types.Enchantment}
	case "land":
		filter.Types = []types.Card{types.Land}
	case "planeswalker":
		filter.Types = []types.Card{types.Planeswalker}
	case "snow land":
		filter.Types = []types.Card{types.Land}
		filter.Supertypes = []types.Super{types.Snow}
	default:
		if lowerColorQualifiedPermanentFilter(noun, &filter) {
			break
		}
		subtype := types.Sub(noun)
		switch {
		case types.KnownSubtypeForType(types.Creature, subtype),
			types.KnownSubtypeForType(types.Land, subtype):
			filter.SubtypesAny = []types.Sub{subtype}
		case strings.HasSuffix(noun, " planeswalker"):
			subtype = types.Sub(strings.TrimSuffix(noun, " planeswalker"))
			if !types.KnownSubtypeForType(types.Planeswalker, subtype) {
				return game.Condition{}, false
			}
			filter.Types = []types.Card{types.Planeswalker}
			filter.SubtypesAny = []types.Sub{subtype}
		default:
			return game.Condition{}, false
		}
	}
	return game.Condition{
		Text:               condition.Text,
		ControllerControls: filter,
	}, true
}

func lowerColorQualifiedPermanentFilter(noun string, filter *game.PermanentFilter) bool {
	const (
		creatureSuffix  = " creature"
		permanentSuffix = " permanent"
	)
	var colorsText string
	switch {
	case strings.HasSuffix(noun, creatureSuffix):
		filter.Types = []types.Card{types.Creature}
		colorsText = strings.TrimSuffix(noun, creatureSuffix)
	case strings.HasSuffix(noun, permanentSuffix):
		colorsText = strings.TrimSuffix(noun, permanentSuffix)
	default:
		return false
	}
	if colorsText == "colorless" {
		filter.ExcludedColors = []color.Color{
			color.White,
			color.Blue,
			color.Black,
			color.Red,
			color.Green,
		}
		return true
	}
	parts := strings.Split(colorsText, " or ")
	filter.ColorsAny = make([]color.Color, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "white":
			filter.ColorsAny = append(filter.ColorsAny, color.White)
		case "blue":
			filter.ColorsAny = append(filter.ColorsAny, color.Blue)
		case "black":
			filter.ColorsAny = append(filter.ColorsAny, color.Black)
		case "red":
			filter.ColorsAny = append(filter.ColorsAny, color.Red)
		case "green":
			filter.ColorsAny = append(filter.ColorsAny, color.Green)
		default:
			filter.Types = nil
			filter.ColorsAny = nil
			return false
		}
	}
	return len(filter.ColorsAny) > 0
}

func matchesExactSourceConditionalKeywordGrantSyntax(
	syntax oracle.Ability,
	condition oracle.CompiledCondition,
	keywords []oracle.CompiledKeyword,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	return matchesPrefixSourceConditionalKeywordGrant(tokens, condition, keywords) ||
		matchesPostfixSourceConditionalKeywordGrant(tokens, condition, keywords)
}

func matchesPrefixSourceConditionalKeywordGrant(
	tokens []oracle.Token,
	condition oracle.CompiledCondition,
	keywords []oracle.CompiledKeyword,
) bool {
	conditionLength := 0
	for conditionLength < len(tokens) &&
		spanCovered(tokens[conditionLength].Span, []oracle.Span{condition.Span}) {
		conditionLength++
	}
	if conditionLength == 0 ||
		len(tokens) < conditionLength+6 ||
		tokens[conditionLength].Kind != oracle.Comma ||
		!equalTokenWord(tokens[conditionLength+1], "this") ||
		!equalTokenWord(tokens[conditionLength+2], "creature") ||
		!equalTokenWord(tokens[conditionLength+3], "has") ||
		tokens[len(tokens)-1].Kind != oracle.Period {
		return false
	}
	return matchesExactKeywordList(tokens[conditionLength+4:len(tokens)-1], keywords)
}

func matchesPostfixSourceConditionalKeywordGrant(
	tokens []oracle.Token,
	condition oracle.CompiledCondition,
	keywords []oracle.CompiledKeyword,
) bool {
	if len(tokens) < 8 ||
		!equalTokenWord(tokens[0], "this") ||
		!equalTokenWord(tokens[1], "creature") ||
		!equalTokenWord(tokens[2], "has") ||
		tokens[len(tokens)-1].Kind != oracle.Period {
		return false
	}
	conditionStart := slices.IndexFunc(tokens, func(token oracle.Token) bool {
		return spanCovered(token.Span, []oracle.Span{condition.Span})
	})
	if conditionStart <= 3 {
		return false
	}
	for _, token := range tokens[conditionStart : len(tokens)-1] {
		if !spanCovered(token.Span, []oracle.Span{condition.Span}) {
			return false
		}
	}
	return matchesExactKeywordList(tokens[3:conditionStart], keywords)
}

func sourceConditionalKeywordGrantSpans(ability oracle.CompiledAbility, syntax oracle.Ability) []oracle.Span {
	spans := make([]oracle.Span, 0, 3+len(ability.Keywords)+len(syntax.Reminders))
	spans = append(
		spans,
		ability.Conditions[0].Span,
		ability.Effects[0].Span,
		ability.References[0].Span,
	)
	for _, keyword := range ability.Keywords {
		spans = append(spans, keyword.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

func mixedStaticKeywords(keywords []oracle.CompiledKeyword) ([]game.Keyword, bool) {
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

func abilityKeywordsExcludingSelectorPredicates(ability oracle.CompiledAbility) []oracle.CompiledKeyword {
	if !abilityUsesCyclingSelectorPredicate(ability) {
		return ability.Keywords
	}
	filtered := make([]oracle.CompiledKeyword, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		if keyword.Name == "Cycling" && keyword.Parameter == "" {
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

func abilityUsesCyclingSelectorPredicate(ability oracle.CompiledAbility) bool {
	for _, target := range ability.Targets {
		if strings.EqualFold(target.Selector.Keyword, "Cycling") {
			return true
		}
	}
	for _, effect := range ability.Effects {
		if strings.EqualFold(effect.Selector.Keyword, "Cycling") ||
			strings.EqualFold(effect.Amount.Selector.Keyword, "Cycling") {
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

func staticSubjectGroup(subject oracle.StaticSubjectKind, subtypeText string) (game.GroupReference, bool) {
	switch subject {
	case oracle.StaticSubjectAttachedObject:
		return game.AttachedObjectGroup(game.SourcePermanentReference()), true
	case oracle.StaticSubjectControlledCreatures:
		return game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
		), true
	case oracle.StaticSubjectOtherControlledCreatures:
		return game.ObjectControlledGroupExcluding(
			game.SourcePermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.SourcePermanentReference(),
		), true
	case oracle.StaticSubjectControlledWalls:
		return game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{SubtypesAny: []types.Sub{types.Wall}},
		), true
	case oracle.StaticSubjectControlledArtifacts:
		return game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Artifact}},
		), true
	case oracle.StaticSubjectControlledTokens:
		return game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{TokenOnly: true},
		), true
	case oracle.StaticSubjectOpponentControlledCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerOpponent,
		}), true
	case oracle.StaticSubjectControlledCreatureSubtype:
		subtype, ok := knownCreatureSubtypeFromPlural(subtypeText)
		if !ok {
			return game.GroupReference{}, false
		}
		return game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{SubtypesAny: []types.Sub{subtype}},
		), true
	case oracle.StaticSubjectOtherControlledCreatureSubtype:
		subtype, ok := knownCreatureSubtypeFromPlural(subtypeText)
		if !ok {
			return game.GroupReference{}, false
		}
		return game.ObjectControlledGroupExcluding(
			game.SourcePermanentReference(),
			game.Selection{SubtypesAny: []types.Sub{subtype}},
			game.SourcePermanentReference(),
		), true
	default:
		return game.GroupReference{}, false
	}
}

func knownCreatureSubtypeFromPlural(text string) (types.Sub, bool) {
	candidates := []string{text}
	if singular, ok := strings.CutSuffix(text, "s"); ok {
		candidates = append(candidates, singular)
	}
	if stem, ok := strings.CutSuffix(text, "ies"); ok {
		candidates = append(candidates, stem+"y")
	}
	if stem, ok := strings.CutSuffix(text, "ves"); ok {
		candidates = append(candidates, stem+"f", stem+"fe")
	}
	if singular, ok := strings.CutSuffix(text, "es"); ok {
		candidates = append(candidates, singular)
	}
	switch text {
	case "Children":
		candidates = append(candidates, "Child")
	case "Mice":
		candidates = append(candidates, "Mouse")
	default:
	}
	for _, candidate := range candidates {
		subtype := types.Sub(candidate)
		if types.KnownSubtypeForType(types.Creature, subtype) {
			return subtype, true
		}
	}
	return "", false
}

func matchesExactStaticKeywordGrantSyntax(
	syntax oracle.Ability,
	effect oracle.CompiledEffect,
	keywords []oracle.CompiledKeyword,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	subjectLength := 0
	for subjectLength < len(tokens) && spanCovered(tokens[subjectLength].Span, []oracle.Span{effect.StaticSubjectSpan}) {
		subjectLength++
	}
	if subjectLength == 0 ||
		len(tokens) < subjectLength+3 ||
		(!equalTokenWord(tokens[subjectLength], "has") && !equalTokenWord(tokens[subjectLength], "have")) ||
		tokens[len(tokens)-1].Kind != oracle.Period {
		return false
	}
	return matchesExactKeywordList(tokens[subjectLength+1:len(tokens)-1], keywords)
}

func matchesExactStaticPTBuffSyntax(
	syntax oracle.Ability,
	effect oracle.CompiledEffect,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	prefixLength, ok := matchesStaticPTBuffPrefix(tokens, effect)
	if !ok {
		return false
	}
	if effect.Amount.DynamicKind != oracle.DynamicAmountNone {
		return len(tokens) > prefixLength+1 &&
			tokens[len(tokens)-1].Kind == oracle.Period &&
			lowerTokenText(tokens[prefixLength:len(tokens)-1]) == effect.Amount.Text
	}
	return len(tokens) == prefixLength+1 &&
		tokens[prefixLength].Kind == oracle.Period
}

func matchesExactSourceStaticPTBuffSyntax(
	syntax oracle.Ability,
	effect oracle.CompiledEffect,
	reference oracle.CompiledReference,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	subjectLength := 0
	for subjectLength < len(tokens) && spanCovered(tokens[subjectLength].Span, []oracle.Span{reference.Span}) {
		subjectLength++
	}
	prefixLength := subjectLength + 6
	if subjectLength == 0 ||
		len(tokens) < prefixLength ||
		!equalTokenWord(tokens[subjectLength], "gets") ||
		!tokensMatchSignedAmount(tokens[subjectLength+1], tokens[subjectLength+2], effect.PowerDelta) ||
		tokens[subjectLength+3].Kind != oracle.Slash ||
		!tokensMatchSignedAmount(tokens[subjectLength+4], tokens[subjectLength+5], effect.ToughnessDelta) {
		return false
	}
	if effect.Amount.DynamicKind != oracle.DynamicAmountNone {
		return len(tokens) > prefixLength+1 &&
			tokens[len(tokens)-1].Kind == oracle.Period &&
			lowerTokenText(tokens[prefixLength:len(tokens)-1]) == effect.Amount.Text
	}
	return len(tokens) == prefixLength+1 &&
		tokens[prefixLength].Kind == oracle.Period
}

func lowerTokenText(tokens []oracle.Token) string {
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, token.Text)
	}
	return strings.Join(parts, " ")
}

func matchesExactStaticPTBuffWithKeywordsSyntax(
	syntax oracle.Ability,
	effect oracle.CompiledEffect,
	keywords []oracle.CompiledKeyword,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	prefixLength, ok := matchesStaticPTBuffPrefix(tokens, effect)
	if !ok ||
		len(tokens) < prefixLength+4 ||
		!equalTokenWord(tokens[prefixLength], "and") ||
		!equalTokenWord(tokens[prefixLength+1], staticPTBuffKeywordVerb(tokens, effect)) ||
		tokens[len(tokens)-1].Kind != oracle.Period {
		return false
	}
	return matchesExactKeywordList(tokens[prefixLength+2:len(tokens)-1], keywords)
}

func staticPTBuffKeywordVerb(tokens []oracle.Token, effect oracle.CompiledEffect) string {
	if effect.StaticSubject == oracle.StaticSubjectAttachedObject ||
		(effect.StaticSubject == oracle.StaticSubjectControlledWalls &&
			len(tokens) > 0 &&
			equalTokenWord(tokens[0], "each")) {
		return "has"
	}
	return "have"
}

func matchesExactKeywordList(tokens []oracle.Token, keywords []oracle.CompiledKeyword) bool {
	elements := make([]string, 0, len(tokens))
	lastKeyword := -1
	for _, token := range tokens {
		keywordIndex := -1
		for i, keyword := range keywords {
			if spanCovered(token.Span, []oracle.Span{keyword.Span}) {
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
		case token.Kind == oracle.Comma:
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
	tokens []oracle.Token,
	effect oracle.CompiledEffect,
) (int, bool) {
	switch effect.StaticSubject {
	case oracle.StaticSubjectAttachedObject:
		return 8, len(tokens) >= 8 &&
			(equalTokenWord(tokens[0], "enchanted") || equalTokenWord(tokens[0], "equipped")) &&
			equalTokenWord(tokens[1], "creature") &&
			equalTokenWord(tokens[2], "gets") &&
			tokensMatchSignedAmount(tokens[3], tokens[4], effect.PowerDelta) &&
			tokens[5].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[6], tokens[7], effect.ToughnessDelta)
	case oracle.StaticSubjectControlledCreatures:
		return 9, len(tokens) >= 9 &&
			equalTokenWord(tokens[0], "creatures") &&
			equalTokenWord(tokens[1], "you") &&
			equalTokenWord(tokens[2], "control") &&
			equalTokenWord(tokens[3], "get") &&
			tokensMatchSignedAmount(tokens[4], tokens[5], effect.PowerDelta) &&
			tokens[6].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[7], tokens[8], effect.ToughnessDelta)
	case oracle.StaticSubjectOtherControlledCreatures:
		return 10, len(tokens) >= 10 &&
			equalTokenWord(tokens[0], "other") &&
			equalTokenWord(tokens[1], "creatures") &&
			equalTokenWord(tokens[2], "you") &&
			equalTokenWord(tokens[3], "control") &&
			equalTokenWord(tokens[4], "get") &&
			tokensMatchSignedAmount(tokens[5], tokens[6], effect.PowerDelta) &&
			tokens[7].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[8], tokens[9], effect.ToughnessDelta)
	case oracle.StaticSubjectControlledWalls:
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
			tokens[offset+6].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[offset+7], tokens[offset+8], effect.ToughnessDelta)
	case oracle.StaticSubjectControlledArtifacts, oracle.StaticSubjectControlledTokens:
		noun := "artifacts"
		if effect.StaticSubject == oracle.StaticSubjectControlledTokens {
			noun = "tokens"
		}
		return 9, len(tokens) >= 9 &&
			equalTokenWord(tokens[0], noun) &&
			equalTokenWord(tokens[1], "you") &&
			equalTokenWord(tokens[2], "control") &&
			equalTokenWord(tokens[3], "get") &&
			tokensMatchSignedAmount(tokens[4], tokens[5], effect.PowerDelta) &&
			tokens[6].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[7], tokens[8], effect.ToughnessDelta)
	case oracle.StaticSubjectOpponentControlledCreatures:
		return 10, len(tokens) >= 10 &&
			equalTokenWord(tokens[0], "creatures") &&
			equalTokenWord(tokens[1], "your") &&
			equalTokenWord(tokens[2], "opponents") &&
			equalTokenWord(tokens[3], "control") &&
			equalTokenWord(tokens[4], "get") &&
			tokensMatchSignedAmount(tokens[5], tokens[6], effect.PowerDelta) &&
			tokens[7].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[8], tokens[9], effect.ToughnessDelta)
	default:
		return 0, false
	}
}

func syntaxSemanticTokens(syntax oracle.Ability) []oracle.Token {
	tokens := make([]oracle.Token, 0, len(syntax.Tokens))
	for _, token := range syntax.Tokens {
		if spanCoveredByDelimited(token.Span, syntax.Reminders) ||
			spanCoveredByDelimited(token.Span, syntax.Quoted) {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func tokensMatchSignedAmount(sign, amount oracle.Token, want oracle.CompiledSignedAmount) bool {
	expectedSign := oracle.Plus
	if want.Negative {
		expectedSign = oracle.Minus
	}
	return sign.Kind == expectedSign &&
		amount.Kind == oracle.Integer &&
		amount.Text == strconv.Itoa(want.Value)
}

// lowerReminderManaAbility preserves a parenthesized reminder mana ability such
// as "({T}: Add {R} or {G}.)" and consumes other rules-free reminder abilities.
func lowerReminderManaAbility(
	ability oracle.CompiledAbility,
) (abilityLowering, *oracle.Diagnostic) {
	unsupported := func() *oracle.Diagnostic {
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
	innerComp, innerDiags := oracle.Compile(inner, oracle.ParseContext{})
	if len(innerComp.Abilities) == 1 && hasAddManaEffect(innerComp.Abilities[0]) {
		if len(innerDiags) != 0 ||
			len(innerComp.Syntax.Abilities) != 1 ||
			innerComp.Abilities[0].Kind != oracle.AbilityActivated {
			return abilityLowering{}, unsupported()
		}
		manaAbility, diagnostic := lowerTapManaAbility(
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
			sourceSpans: []oracle.Span{ability.Span},
		}, nil
	}

	// Non-mana reminder abilities carry no semantic content beyond their
	// parenthesized explanation.
	return abilityLowering{
		sourceSpans: []oracle.Span{ability.Span},
	}, nil
}

func lowerEntersTappedReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, *oracle.Diagnostic) {
	if replacement, ok := lowerOptionalEntryPayment(ability); ok {
		return replacement, nil
	}
	if !entersTappedReplacementEffectsSupported(ability) ||
		ability.Effects[0].Kind != oracle.EffectEnterTapped ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferenceThisObject {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	if len(ability.Conditions) == 1 {
		return lowerConditionalEntersTappedReplacement(ability)
	}
	if len(ability.Conditions) != 0 {
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
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
	event, eventOK := selfZoneDestinationReplacedEvent(ability)
	if !eventOK {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported self zone-destination replacement",
			detail,
		)
	}
	if len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
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
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
	if !counterPlacementReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported counter-placement replacement",
			detail,
		)
	}
	if len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != oracle.ConditionIf ||
		len(ability.Effects) != 2 ||
		ability.Effects[0].Kind != oracle.EffectPut ||
		ability.Effects[1].Kind != oracle.EffectPut ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact counter-doubling replacements")
	}
	condition := ability.Conditions[0].Text
	switch condition {
	case "If one or more +1/+1 counters would be put on a creature you control":
		if !plusOneCounterDoublingEffects(ability.Effects) {
			return unsupported("the executable source backend supports only +1/+1 counter-doubling replacement amounts")
		}
		return game.CounterPlacementReplacement(ability.Text, 2, counter.PlusOnePlusOne, game.TriggerControllerYou), true, nil
	case "If you would put one or more counters on a permanent or player":
		if !anyCounterDoublingEffects(ability.Effects) {
			return unsupported("the executable source backend supports only all-counter-doubling replacement amounts")
		}
		return game.AnyCounterPlacementReplacement(ability.Text, 2, game.TriggerControllerYou), true, nil
	default:
		return unsupported("the executable source backend supports only controlled-creature +1/+1 or broad permanent/player counter-doubling replacements")
	}
}

func lowerDamageReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
	if !damageReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported damage replacement",
			detail,
		)
	}
	if len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != oracle.ConditionIf ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact additive or multiplicative damage replacements")
	}
	condition := ability.Conditions[0].Text
	raw := damageReplacementRawEffects(ability.Effects)
	switch condition {
	case "If another red source you control would deal damage to a permanent or player",
		"If a red source you control would deal damage to a permanent or player":
		if !strings.Contains(raw, "that much damage plus 1 to that permanent or player instead.") {
			return unsupported("the executable source backend supports only +1 red-source damage replacements")
		}
		if strings.Contains(condition, "another red source") {
			return game.DamageReplacementExcludingSource(ability.Text, 0, 1, []color.Color{color.Red}, game.TriggerControllerYou), true, nil
		}
		return game.DamageReplacement(ability.Text, 0, 1, []color.Color{color.Red}, game.TriggerControllerYou), true, nil
	case "If a source you control would deal damage to a permanent or player":
		if !strings.Contains(raw, "double that damage to that permanent or player instead.") &&
			!strings.Contains(raw, "twice that damage to that permanent or player instead.") {
			return unsupported("the executable source backend supports only double-damage replacements")
		}
		return game.DamageReplacement(ability.Text, 2, 0, nil, game.TriggerControllerYou), true, nil
	default:
		return unsupported("the executable source backend supports only controlled-source red +1 damage or controlled-source double-damage replacements")
	}
}

func damageReplacementCandidate(ability oracle.CompiledAbility) bool {
	if ability.Kind != oracle.AbilityReplacement || len(ability.Conditions) == 0 {
		return false
	}
	return strings.Contains(ability.Conditions[0].Text, "would deal damage")
}

func damageReplacementRawEffects(effects []oracle.CompiledEffect) string {
	raw := make([]string, 0, len(effects))
	for i := range effects {
		raw = append(raw, effects[i].Selector.Raw)
	}
	return strings.Join(raw, " ")
}

func counterPlacementReplacementCandidate(ability oracle.CompiledAbility) bool {
	if ability.Kind != oracle.AbilityReplacement || len(ability.Conditions) == 0 {
		return false
	}
	condition := ability.Conditions[0].Text
	return strings.Contains(condition, "counters would be put") ||
		strings.Contains(condition, "would put one or more counters")
}

func plusOneCounterDoublingEffects(effects []oracle.CompiledEffect) bool {
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

func anyCounterDoublingEffects(effects []oracle.CompiledEffect) bool {
	raw := effects[0].Selector.Raw + " " + effects[1].Selector.Raw
	return strings.Contains(raw, "twice that many of each of those kinds of counters") &&
		strings.Contains(raw, "on that permanent or player instead.")
}

func lowerTokenCreationReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
	if !tokenCreationReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported token-creation replacement",
			detail,
		)
	}
	if len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != oracle.ConditionIf ||
		ability.Conditions[0].Text != "If an effect would create one or more tokens under your control" ||
		len(ability.Effects) != 2 ||
		ability.Effects[0].Kind != oracle.EffectCreate ||
		ability.Effects[1].Kind != oracle.EffectCreate ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact token-doubling replacements under your control")
	}
	switch ability.Effects[1].Selector.Raw {
	case "twice that many of those tokens instead.", "twice that many tokens instead.":
	default:
		return unsupported("the executable source backend supports only token-doubling replacement amounts")
	}
	return game.TokenCreationReplacement(ability.Text, 2, game.TriggerControllerYou), true, nil
}

func tokenCreationReplacementCandidate(ability oracle.CompiledAbility) bool {
	if ability.Kind != oracle.AbilityReplacement || len(ability.Conditions) == 0 {
		return false
	}
	condition := ability.Conditions[0].Text
	return strings.Contains(condition, "would create") && strings.Contains(condition, "tokens")
}

type selfZoneDestinationEvent struct {
	fromZone      zone.Type
	matchFromZone bool
}

func selfZoneDestinationReplacedEvent(ability oracle.CompiledAbility) (selfZoneDestinationEvent, bool) {
	if len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != oracle.ConditionIf {
		return selfZoneDestinationEvent{}, false
	}
	condition := ability.Conditions[0].Text
	subject, ok := strings.CutPrefix(condition, "If ")
	if !ok {
		return selfZoneDestinationEvent{}, false
	}
	subject, ok = strings.CutSuffix(subject, " would be put into a graveyard from anywhere")
	if ok && selfReferenceSubjectSupported(ability.References, subject) {
		return selfZoneDestinationEvent{}, true
	}
	subject, ok = strings.CutSuffix(strings.TrimPrefix(condition, "If "), " would die")
	if ok && selfReferenceSubjectSupported(ability.References, subject) {
		return selfZoneDestinationEvent{fromZone: zone.Battlefield, matchFromZone: true}, true
	}
	return selfZoneDestinationEvent{}, false
}

func selfReferenceSubjectSupported(references []oracle.CompiledReference, subject string) bool {
	for _, reference := range references {
		if reference.Kind != oracle.ReferenceThisObject &&
			reference.Kind != oracle.ReferenceSelfName {
			continue
		}
		if strings.EqualFold(reference.Text, subject) {
			return true
		}
	}
	return false
}

func selfZoneDestinationReferencesSupported(ability oracle.CompiledAbility) bool {
	for _, reference := range ability.References {
		switch reference.Kind {
		case oracle.ReferenceThisObject, oracle.ReferenceSelfName, oracle.ReferencePronoun:
		default:
			return false
		}
	}
	return len(ability.References) > 0
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
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
	if !isEntersWithCountersReplacement(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *oracle.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported enters-with-counters replacement",
			detail,
		)
	}
	if len(ability.Conditions) != 0 {
		return unsupported("the executable source backend does not yet support conditional enters-with-counters replacements")
	}
	if len(ability.Effects) != 1 ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!selfEntersWithCountersReferences(ability.References) {
		return unsupported("the executable source backend supports only exact unconditional self enters-with-counters replacements")
	}
	effect := ability.Effects[0]
	if effect.Duration != oracle.DurationNone || effect.Negated {
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

func isEntersWithCountersReplacement(ability oracle.CompiledAbility) bool {
	if len(ability.Effects) == 0 ||
		ability.Effects[0].Kind != oracle.EffectEnterTapped {
		return false
	}
	raw := ability.Effects[0].Selector.Raw
	return strings.HasPrefix(raw, "with ") &&
		strings.Contains(raw, " counter") &&
		strings.HasSuffix(raw, " on it.")
}

func selfEntersWithCountersReferences(references []oracle.CompiledReference) bool {
	return len(references) == 2 &&
		references[0].Kind == oracle.ReferenceThisObject &&
		references[1].Kind == oracle.ReferencePronoun &&
		strings.EqualFold(references[1].Text, "it")
}

func lowerOptionalEntryPayment(ability oracle.CompiledAbility) (game.ReplacementAbility, bool) {
	if len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != oracle.ConditionIf ||
		ability.Conditions[0].Text != "If you don't" ||
		len(ability.Targets) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return game.ReplacementAbility{}, false
	}
	const payLifeText = "As this land enters, you may pay 2 life. If you don't, it enters tapped."
	if ability.Text == payLifeText &&
		len(ability.Effects) == 2 &&
		ability.Effects[0].Kind == oracle.EffectEnterTapped &&
		ability.Effects[0].Amount.Known &&
		ability.Effects[0].Amount.Value == 2 &&
		!ability.Effects[0].Selector.Tapped &&
		ability.Effects[1].Kind == oracle.EffectEnterTapped &&
		ability.Effects[1].Selector.Tapped &&
		len(ability.References) == 2 &&
		ability.References[0].Kind == oracle.ReferenceThisObject &&
		ability.References[1].Kind == oracle.ReferencePronoun {
		return game.EntersTappedUnlessPaidReplacement(ability.Text, game.ResolutionPayment{
			Prompt: "Pay 2 life?",
			AdditionalCosts: []cost.Additional{{
				Kind:   cost.AdditionalPayLife,
				Amount: 2,
			}},
		}), true
	}
	subtypes, ok := revealEntrySubtypes(ability.Text)
	if !ok ||
		len(ability.Effects) != 3 ||
		ability.Effects[0].Kind != oracle.EffectEnterTapped ||
		ability.Effects[0].Selector.Tapped ||
		ability.Effects[1].Kind != oracle.EffectReveal ||
		ability.Effects[1].Amount.Value != 1 ||
		!ability.Effects[1].Amount.Known ||
		ability.Effects[2].Kind != oracle.EffectEnterTapped ||
		!ability.Effects[2].Selector.Tapped ||
		len(ability.References) != 2 ||
		ability.References[0].Kind != oracle.ReferenceThisObject ||
		ability.References[1].Kind != oracle.ReferenceThisObject {
		return game.ReplacementAbility{}, false
	}
	var subtypeSet cost.SubtypeSet
	copy(subtypeSet[:], subtypes)
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

func revealEntrySubtypes(text string) ([]types.Sub, bool) {
	const prefix = "As this land enters, you may reveal "
	const suffix = " card from your hand. If you don't, this land enters tapped."
	if !strings.HasPrefix(text, prefix) || !strings.HasSuffix(text, suffix) {
		return nil, false
	}
	names := strings.Split(strings.TrimSuffix(strings.TrimPrefix(text, prefix), suffix), " or ")
	if len(names) > 2 {
		return nil, false
	}
	subtypes := make([]types.Sub, 0, len(names))
	for _, name := range names {
		subtype := types.Sub(strings.TrimPrefix(strings.TrimPrefix(name, "a "), "an "))
		if !types.KnownSubtypeForType(types.Land, subtype) &&
			!types.KnownSubtypeForType(types.Creature, subtype) {
			return nil, false
		}
		subtypes = append(subtypes, subtype)
	}
	return subtypes, len(subtypes) > 0
}

func entersTappedReplacementEffectsSupported(ability oracle.CompiledAbility) bool {
	if len(ability.Effects) == 0 {
		return false
	}
	if len(ability.Effects) == 1 {
		return true
	}
	if len(ability.Conditions) != 1 {
		return false
	}
	conditionSpans := []oracle.Span{ability.Conditions[0].Span}
	for _, effect := range ability.Effects[1:] {
		if !spanCovered(effect.VerbSpan, conditionSpans) {
			return false
		}
	}
	return true
}

func lowerConditionalEntersTappedReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, *oracle.Diagnostic) {
	condition := ability.Conditions[0]
	if condition.Kind != oracle.ConditionUnless ||
		ability.Text != "This land enters tapped "+condition.Text+"." {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported conditional enters-tapped replacement",
			"the executable source backend does not support this enters-tapped condition",
		)
	}
	var replacementCondition game.Condition
	switch condition.Text {
	case "unless you have 10 or more life":
		replacementCondition.Negate = true
		replacementCondition.ControllerLifeAtLeast = 10
	case "unless you have 20 or more life":
		replacementCondition.Negate = true
		replacementCondition.ControllerLifeAtLeast = 20
	case "unless a player has 13 or less life":
		replacementCondition.Negate = true
		replacementCondition.AnyPlayerLifeAtMost = 13
	case "unless you have two or more opponents":
		replacementCondition.Negate = true
		replacementCondition.OpponentCountAtLeast = 2
	case "unless an opponent controls two or more lands":
		replacementCondition.Negate = true
		replacementCondition.AnyOpponentControls = opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
			MinCount:  2,
		})
	case "unless your opponents control eight or more lands":
		replacementCondition.Negate = true
		replacementCondition.OpponentsControl = opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
			MinCount:  8,
		})
	case "unless you control two or more basic lands":
		replacementCondition.Negate = true
		replacementCondition.ControllerControls = game.PermanentFilter{
			Types:      []types.Card{types.Land},
			Supertypes: []types.Super{types.Basic},
			MinCount:   2,
		}
	case "unless you control two or more other lands":
		replacementCondition.Negate = true
		replacementCondition.ControllerControls = game.PermanentFilter{
			Types:         []types.Card{types.Land},
			MinCount:      2,
			ExcludeSource: true,
		}
	case "unless you control two or fewer other lands":
		replacementCondition.ControllerControls = game.PermanentFilter{
			Types:         []types.Card{types.Land},
			MinCount:      3,
			ExcludeSource: true,
		}
	default:
		subtypes, ok := entersTappedLandSubtypes(condition.Text)
		if !ok {
			return game.ReplacementAbility{}, executableDiagnostic(
				ability,
				"unsupported conditional enters-tapped replacement",
				"the executable source backend does not support this enters-tapped condition",
			)
		}
		replacementCondition.Negate = true
		replacementCondition.ControllerControls = game.PermanentFilter{
			Types:       []types.Card{types.Land},
			SubtypesAny: subtypes,
		}
	}
	return game.EntersTappedIfReplacement(ability.Text, &replacementCondition), nil
}

func entersTappedLandSubtypes(condition string) ([]types.Sub, bool) {
	const prefix = "unless you control "
	if !strings.HasPrefix(condition, prefix) {
		return nil, false
	}
	parts := strings.Split(strings.TrimPrefix(condition, prefix), " or ")
	if len(parts) > 2 {
		return nil, false
	}
	subtypes := make([]types.Sub, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimPrefix(strings.TrimPrefix(part, "a "), "an ")
		subtype := types.Sub(name)
		if !types.KnownSubtypeForType(types.Land, subtype) {
			return nil, false
		}
		subtypes = append(subtypes, subtype)
	}
	return subtypes, len(subtypes) > 0
}

type atTriggerParams struct {
	step       game.Step
	controller game.TriggerControllerFilter
}

var atTriggerPhrases = map[string]atTriggerParams{
	"the beginning of your upkeep":            {game.StepUpkeep, game.TriggerControllerYou},
	"the beginning of each upkeep":            {game.StepUpkeep, game.TriggerControllerAny},
	"the beginning of each player's upkeep":   {game.StepUpkeep, game.TriggerControllerAny},
	"the beginning of each opponent's upkeep": {game.StepUpkeep, game.TriggerControllerOpponent},
	"the beginning of your end step":          {game.StepEnd, game.TriggerControllerYou},
	"the beginning of each end step":          {game.StepEnd, game.TriggerControllerAny},
	"the beginning of each player's end step": {game.StepEnd, game.TriggerControllerAny},
	"the beginning of combat on your turn":    {game.StepBeginningOfCombat, game.TriggerControllerYou},
	"the beginning of each combat":            {game.StepBeginningOfCombat, game.TriggerControllerAny},
	"the beginning of your draw step":         {game.StepDraw, game.TriggerControllerYou},
}

func lowerAtTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, *oracle.Diagnostic) {
	const summary = "unsupported phase/step trigger phrase"
	params, ok := atTriggerPhrases[ability.Trigger.Event]
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			fmt.Sprintf("the executable source backend does not support step trigger phrase %q", ability.Trigger.Event),
		)
	}
	if ability.Trigger.Condition != nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"intervening-if conditions are not supported for phase/step triggers",
		)
	}
	if len(ability.Modes) != 0 || ability.AbilityWord != "" {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"modes and ability words are not supported in phase/step triggers",
		)
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the executable source backend does not support this phase/step trigger body",
		)
	}
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary+" effect",
			diagnostic.Detail,
		)
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type: game.TriggerAt,
			Pattern: game.TriggerPattern{
				Event:      game.EventBeginningOfStep,
				Controller: params.controller,
				Step:       params.step,
			},
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func lowerTriggeredAbility(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, *oracle.Diagnostic) {
	if ability.Trigger != nil && ability.Trigger.Kind == oracle.TriggerAt {
		return lowerAtTrigger(cardName, ability, syntax)
	}
	if cyclingAbility, ok, diagnostic := lowerCyclingTrigger(cardName, ability, syntax); ok {
		return cyclingAbility, diagnostic
	}
	if triggeredAbility, ok := lowerLifeDamageTrigger(cardName, ability, syntax); ok {
		return triggeredAbility, nil
	}
	triggeredAbility, diagnostic := lowerEnterTrigger(cardName, ability, syntax)
	if diagnostic == nil ||
		ability.Trigger == nil ||
		!strings.Contains(ability.Trigger.Event, " enter") {
		if diagnostic != nil && ability.Trigger != nil {
			if castAbility, ok := lowerCastTrigger(cardName, ability, syntax); ok {
				return castAbility, nil
			}
		}
		return triggeredAbility, diagnostic
	}
	nonSelf, ok := lowerNonSelfEnterTrigger(cardName, ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, diagnostic
	}
	return nonSelf, nil
}

type cyclingTriggerPattern struct {
	event       game.EventKind
	excludeSelf bool
}

var cyclingTriggerPhrases = map[string]cyclingTriggerPattern{
	"you cycle a card":                  {event: game.EventCycled},
	"you cycle another card":            {event: game.EventCycled, excludeSelf: true},
	"you cycle or discard a card":       {event: game.EventCardDiscarded},
	"you cycle or discard another card": {event: game.EventCardDiscarded, excludeSelf: true},
}

func lowerCyclingTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, bool, *oracle.Diagnostic) {
	if ability.Trigger == nil || ability.Trigger.Kind != oracle.TriggerWhenever {
		return game.TriggeredAbility{}, false, nil
	}
	params, ok := cyclingTriggerPhrases[ability.Trigger.Event]
	if !ok {
		return game.TriggeredAbility{}, false, nil
	}
	const summary = "unsupported cycling trigger"
	if ability.Trigger.Condition != nil ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			summary,
			"the executable source backend supports only exact cycling trigger phrases without modes, ability words, or intervening conditions",
		)
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			summary,
			"the executable source backend does not support this cycling trigger body",
		)
	}
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			summary+" effect",
			diagnostic.Detail,
		)
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:       params.event,
				Player:      game.TriggerPlayerYou,
				ExcludeSelf: params.excludeSelf,
			},
		},
		Optional: ability.Optional,
		Content:  content,
	}, true, nil
}

func lowerTriggeredAbilityKind(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	triggeredAbility, diagnostic := lowerTriggeredAbility(cardName, ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := []oracle.Span{ability.Trigger.Span}
	if syntax.AbilityWord != nil {
		spans = append(spans, oracle.Span{
			Start: ability.Span.Start,
			End:   ability.Trigger.Span.Start,
		})
	}
	for _, effect := range ability.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Targets {
		spans = append(spans, target.Span)
	}
	for _, condition := range ability.Conditions {
		spans = append(spans, condition.Span)
	}
	for _, reference := range ability.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		triggeredAbility: opt.Val(triggeredAbility),
		consumed: semanticConsumption{
			trigger:    true,
			optional:   ability.Optional,
			targets:    len(ability.Targets),
			conditions: len(ability.Conditions),
			effects:    len(ability.Effects),
			keywords:   len(ability.Keywords),
			references: len(ability.References),
		},
		sourceSpans: spans,
	}, nil
}

func (lowering *abilityLowering) complete(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) bool {
	if lowering.consumed.cost != (ability.Cost != nil) ||
		lowering.consumed.trigger != (ability.Trigger != nil) ||
		lowering.consumed.optional != ability.Optional ||
		lowering.consumed.modes != len(ability.Modes) ||
		lowering.consumed.targets != len(ability.Targets) ||
		lowering.consumed.conditions != len(ability.Conditions) ||
		lowering.consumed.effects != len(ability.Effects) ||
		lowering.consumed.keywords != len(ability.Keywords) ||
		lowering.consumed.references != len(ability.References) {
		return false
	}
	for _, token := range syntax.Tokens {
		if token.Kind == oracle.Comma ||
			token.Kind == oracle.Colon ||
			token.Kind == oracle.Period ||
			spanCovered(token.Span, lowering.sourceSpans) {
			continue
		}
		return false
	}
	return true
}

func lowerEnterTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, *oracle.Diagnostic) {
	pattern, supportedEvent := lowerSelfTriggerPattern(cardName, ability)
	eventKind := pattern.Event
	summary := "unsupported triggered ability"
	detail := "the executable source backend supports only exact self-enter, self-dies, self-mutate, and simple combat triggers with supported effects"
	if ability.Trigger != nil && strings.Contains(ability.Trigger.Event, " enters") {
		summary = "unsupported enter trigger"
		detail = "the executable source backend supports only exact self-enter triggers with supported effects"
	} else if ability.Trigger != nil && strings.HasSuffix(ability.Trigger.Event, " dies") {
		summary = "unsupported dies trigger"
		detail = "the executable source backend supports only exact self-dies triggers with supported effects"
	}
	intervening, supportedCondition := lowerSelfInterveningCondition(eventKind, ability.Trigger)
	if ability.Trigger == nil ||
		!supportedSelfTriggerKind(eventKind, ability.Trigger.Kind) ||
		!supportedEvent ||
		!supportedCondition ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	selfDamage := eventKind == game.EventPermanentDied &&
		normalizeSelfDamageReference(cardName, &body)
	content, diagnostic := lowerSelfTriggerBody(cardName, eventKind, body, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary+" effect",
			diagnostic.Detail,
		)
	}
	if selfDamage {
		damage, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
		if !ok {
			return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
		}
		damage.DamageSource = opt.Val(game.EventPermanentReference())
		if dynamic := damage.Amount.DynamicAmount(); dynamic.Exists &&
			dynamic.Val.Kind == game.DynamicAmountObjectPower {
			dynamic.Val.Object = game.EventPermanentReference()
			damage.Amount = game.Dynamic(dynamic.Val)
		}
		content.Modes[0].Sequence[0].Primitive = damage
	}
	triggerType := game.TriggerWhen
	if ability.Trigger.Kind == oracle.TriggerWhenever {
		triggerType = game.TriggerWhenever
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 triggerType,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening.condition,
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, bool) {
	if ability.Trigger == nil || ability.Trigger.Kind != oracle.TriggerWhenever {
		return game.TriggeredAbility{}, false
	}
	pattern, ok := lowerLifeDamageTriggerPattern(ability)
	if !ok ||
		ability.Trigger.Condition != nil ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.TriggeredAbility{}, false
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, false
	}
	content, diagnostic := lowerSelfTriggerBody(cardName, pattern.Event, body, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, false
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhenever,
			Pattern: pattern,
		},
		Optional: ability.Optional,
		Content:  content,
	}, true
}

func lowerLifeDamageTriggerPattern(ability oracle.CompiledAbility) (game.TriggerPattern, bool) {
	if ability.Trigger == nil {
		return game.TriggerPattern{}, false
	}
	switch ability.Trigger.Event {
	case "you gain life":
		return game.TriggerPattern{
			Event:  game.EventLifeGained,
			Player: game.TriggerPlayerYou,
		}, true
	case "an opponent gains life":
		return game.TriggerPattern{
			Event:  game.EventLifeGained,
			Player: game.TriggerPlayerOpponent,
		}, true
	case "you lose life":
		return game.TriggerPattern{
			Event:  game.EventLifeLost,
			Player: game.TriggerPlayerYou,
		}, true
	case "an opponent loses life":
		return game.TriggerPattern{
			Event:  game.EventLifeLost,
			Player: game.TriggerPlayerOpponent,
		}, true
	case "this creature is dealt damage",
		"this permanent is dealt damage":
		return game.TriggerPattern{
			Event:           game.EventDamageDealt,
			Source:          game.TriggerSourceSelf,
			Subject:         game.TriggerSubjectPermanent,
			DamageRecipient: game.DamageRecipientPermanent,
		}, true
	case "enchanted creature is dealt damage",
		"enchanted permanent is dealt damage",
		"equipped creature is dealt damage":
		return game.TriggerPattern{
			Event:           game.EventDamageDealt,
			Source:          game.TriggerSourceAttachedPermanent,
			DamageRecipient: game.DamageRecipientPermanent,
		}, true
	case "you're dealt damage", "you are dealt damage":
		return game.TriggerPattern{
			Event:           game.EventDamageDealt,
			Player:          game.TriggerPlayerYou,
			DamageRecipient: game.DamageRecipientPlayer,
		}, true
	default:
		return game.TriggerPattern{}, false
	}
}

func lowerSelfTriggerBody(
	cardName string,
	eventKind game.EventKind,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	if eventKind == game.EventPermanentDied {
		if content, ok := lowerDiesEventCardEffect(ability); ok {
			return content, nil
		}
	}
	return lowerSpell(cardName, ability, syntax)
}

func lowerDiesEventCardEffect(ability oracle.CompiledAbility) (game.AbilityContent, bool) {
	if len(ability.Effects) != 1 ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	eventCard := game.CardReference{Kind: game.CardReferenceEvent}
	switch ability.Effects[0].Kind {
	case oracle.EffectReturn:
		if ability.Text != "Return it to its owner's hand." ||
			!exactPronounReferences(ability.References, "it", "its") {
			return game.AbilityContent{}, false
		}
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				Card:        eventCard,
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			},
		}}}.Ability(), true
	case oracle.EffectCast:
		if ability.Text != "Cast it from your graveyard as an Adventure until the end of your next turn." ||
			!exactPronounReferences(ability.References, "it") {
			return game.AbilityContent{}, false
		}
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GrantCastPermission{
				Card:     eventCard,
				FromZone: zone.Graveyard,
				Face:     game.FaceAlternate,
				Duration: game.DurationUntilEndOfYourNextTurn,
			},
		}}}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

func exactPronounReferences(references []oracle.CompiledReference, texts ...string) bool {
	if len(references) != len(texts) {
		return false
	}
	for i, text := range texts {
		if references[i].Kind != oracle.ReferencePronoun ||
			!strings.EqualFold(references[i].Text, text) {
			return false
		}
	}
	return true
}

type enterInterveningCondition struct {
	condition        opt.V[game.Condition]
	hadNoCounterKind opt.V[counter.Kind]
	wasKicked        bool
	wasCast          bool
}

func lowerSelfInterveningCondition(
	eventKind game.EventKind,
	trigger *oracle.CompiledTrigger,
) (enterInterveningCondition, bool) {
	switch eventKind {
	case game.EventPermanentEnteredBattlefield:
		return lowerEnterInterveningCondition(trigger)
	case game.EventPermanentDied:
		return lowerDiesInterveningCondition(trigger)
	default:
		return enterInterveningCondition{}, trigger == nil || trigger.Condition == nil
	}
}

func supportedSelfTriggerKind(eventKind game.EventKind, kind oracle.TriggerKind) bool {
	switch eventKind {
	case game.EventPermanentMutated,
		game.EventAttackerBecameBlocked,
		game.EventAttackerDeclared,
		game.EventBlockerDeclared,
		game.EventDamageDealt:
		return kind == oracle.TriggerWhenever
	default:
		return kind == oracle.TriggerWhen
	}
}

func lowerEnterInterveningCondition(trigger *oracle.CompiledTrigger) (enterInterveningCondition, bool) {
	if trigger == nil || trigger.Condition == nil {
		return enterInterveningCondition{}, true
	}
	condition := trigger.Condition
	if condition.Kind != oracle.ConditionIf || !condition.Intervening {
		return enterInterveningCondition{}, false
	}
	switch condition.Text {
	case "if it was kicked":
		return enterInterveningCondition{wasKicked: true}, true
	case "if it was cast", "if you cast it":
		return enterInterveningCondition{wasCast: true}, true
	}
	cardType, ok := controlledPermanentConditionType(condition.Text)
	if !ok {
		return enterInterveningCondition{}, false
	}
	return enterInterveningCondition{
		condition: opt.Val(game.Condition{
			Text: condition.Text,
			ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{RequiredTypes: []types.Card{cardType}},
			}),
		}),
	}, true
}

func lowerDiesInterveningCondition(trigger *oracle.CompiledTrigger) (enterInterveningCondition, bool) {
	if trigger == nil || trigger.Condition == nil {
		return enterInterveningCondition{}, true
	}
	condition := trigger.Condition
	if condition.Kind != oracle.ConditionIf || !condition.Intervening {
		return enterInterveningCondition{}, false
	}
	switch condition.Text {
	case "if it had no +1/+1 counters", "if it had no +1/+1 counters on it":
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.PlusOnePlusOne)}, true
	case "if it had no -1/-1 counters", "if it had no -1/-1 counters on it":
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.MinusOneMinusOne)}, true
	default:
		return enterInterveningCondition{}, false
	}
}

func controlledPermanentConditionType(text string) (types.Card, bool) {
	switch text {
	case "if you control a battle":
		return types.Battle, true
	case "if you control a creature":
		return types.Creature, true
	case "if you control an artifact":
		return types.Artifact, true
	case "if you control an enchantment":
		return types.Enchantment, true
	case "if you control a land":
		return types.Land, true
	case "if you control a planeswalker":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

func normalizeSelfDamageReference(cardName string, ability *oracle.CompiledAbility) bool {
	if ability == nil ||
		len(ability.Effects) != 1 ||
		(len(ability.References) != 1 && len(ability.References) != 2) ||
		ability.References[0].Kind != oracle.ReferencePronoun ||
		!strings.EqualFold(ability.References[0].Text, "it") ||
		!strings.HasPrefix(ability.Text, "It deals ") ||
		!strings.HasPrefix(strings.ToLower(ability.Effects[0].Text), "it deals ") {
		return false
	}
	if len(ability.References) == 2 &&
		(ability.Effects[0].Amount.DynamicKind != oracle.DynamicAmountSourcePower ||
			ability.References[1].Kind != oracle.ReferencePronoun ||
			!strings.EqualFold(ability.References[1].Text, "its") ||
			ability.References[1].Span != ability.Effects[0].Amount.ReferenceSpan) {
		return false
	}
	ability.Text = cardName + ability.Text[len("It"):]
	ability.Effects[0].Text = cardName + ability.Effects[0].Text[len("It"):]
	ability.References[0].Kind = oracle.ReferenceSelfName
	ability.References[0].Text = cardName
	return true
}

func lowerSelfTriggerPattern(cardName string, ability oracle.CompiledAbility) (game.TriggerPattern, bool) {
	if ability.Trigger == nil {
		return game.TriggerPattern{}, false
	}
	switch ability.Trigger.Event {
	case "this creature enters",
		"this permanent enters",
		"this aura enters",
		"this artifact enters",
		"this equipment enters",
		"this land enters",
		"this vehicle enters",
		"this enchantment enters":
		return game.TriggerPattern{
			Event:  game.EventPermanentEnteredBattlefield,
			Source: game.TriggerSourceSelf,
		}, true
	case "this creature dies", "this permanent dies":
		return game.TriggerPattern{
			Event:  game.EventPermanentDied,
			Source: game.TriggerSourceSelf,
		}, true
	case "this creature mutates":
		return game.TriggerPattern{
			Event:  game.EventPermanentMutated,
			Source: game.TriggerSourceSelf,
		}, true
	case "this creature attacks":
		return game.TriggerPattern{
			Event:  game.EventAttackerDeclared,
			Source: game.TriggerSourceSelf,
		}, true
	case "this creature blocks":
		return game.TriggerPattern{
			Event:  game.EventBlockerDeclared,
			Source: game.TriggerSourceSelf,
		}, true
	case "this creature becomes blocked":
		return game.TriggerPattern{
			Event:  game.EventAttackerBecameBlocked,
			Source: game.TriggerSourceSelf,
		}, true
	case "this creature deals combat damage to a player":
		return game.TriggerPattern{
			Event:               game.EventDamageDealt,
			Source:              game.TriggerSourceSelf,
			Subject:             game.TriggerSubjectDamageSource,
			DamageRecipient:     game.DamageRecipientPlayer,
			RequireCombatDamage: true,
		}, true
	case "this creature deals combat damage to a creature":
		return game.TriggerPattern{
			Event:                game.EventDamageDealt,
			Source:               game.TriggerSourceSelf,
			Subject:              game.TriggerSubjectDamageSource,
			DamageRecipient:      game.DamageRecipientPermanent,
			DamageRecipientTypes: []types.Card{types.Creature},
			RequireCombatDamage:  true,
		}, true
	default:
		if strings.EqualFold(ability.Trigger.Event, cardName+" dies") {
			return game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			}, true
		}
		return game.TriggerPattern{}, false
	}
}

func bodyReferences(
	references []oracle.CompiledReference,
	excludedSpans ...oracle.Span,
) []oracle.CompiledReference {
	var body []oracle.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, excludedSpans) {
			continue
		}
		body = append(body, reference)
	}
	return body
}

func interveningIfText(trigger *oracle.CompiledTrigger) string {
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (oracle.CompiledAbility, oracle.Ability, bool) {
	if ability.Trigger == nil {
		return oracle.CompiledAbility{}, oracle.Ability{}, false
	}
	hasInterveningCondition := ability.Trigger.Condition != nil
	if (len(ability.Conditions) != 0 && !hasInterveningCondition) ||
		(hasInterveningCondition && (len(ability.Conditions) != 1 ||
			ability.Conditions[0] != *ability.Trigger.Condition ||
			ability.Optional)) {
		return oracle.CompiledAbility{}, oracle.Ability{}, false
	}
	resolvingEffects := ability.Effects
	if hasInterveningCondition {
		conditionSpan := []oracle.Span{ability.Trigger.Condition.Span}
		resolvingEffects = slices.DeleteFunc(
			append([]oracle.CompiledEffect(nil), ability.Effects...),
			func(effect oracle.CompiledEffect) bool {
				return spanCovered(effect.VerbSpan, conditionSpan)
			},
		)
	}
	if len(resolvingEffects) == 0 {
		return oracle.CompiledAbility{}, oracle.Ability{}, false
	}
	body := ability
	body.Effects = resolvingEffects
	body.Kind = oracle.AbilitySpell
	body.Span = oracle.Span{
		Start: resolvingEffects[0].Span.Start,
		End:   resolvingEffects[len(resolvingEffects)-1].Span.End,
	}
	body.Text = titleFirst(
		ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
	)
	body.Trigger = nil
	body.Optional = false
	body.OptionalSpan = oracle.Span{}
	excludedReferenceSpans := []oracle.Span{ability.Trigger.Span}
	if hasInterveningCondition {
		excludedReferenceSpans = append(excludedReferenceSpans, ability.Trigger.Condition.Span)
		body.Conditions = nil
		bodyStart := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
			return token.Kind != oracle.Comma &&
				token.Span.Start.Offset >= ability.Trigger.Condition.Span.End.Offset
		})
		if bodyStart < 0 {
			return oracle.CompiledAbility{}, oracle.Ability{}, false
		}
		effect := body.Effects[0]
		effect.Span.Start = syntax.Tokens[bodyStart].Span.Start
		effect.Text = ability.Text[effect.Span.Start.Offset-ability.Span.Start.Offset : effect.Span.End.Offset-ability.Span.Start.Offset]
		body.Effects[0] = effect
		body.Span.Start = effect.Span.Start
		body.Text = titleFirst(
			ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
		)
	}
	body.References = bodyReferences(ability.References, excludedReferenceSpans...)
	bodyTokenStart := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Span.Start.Offset >= body.Span.Start.Offset
	})
	if bodyTokenStart < 0 {
		return oracle.CompiledAbility{}, oracle.Ability{}, false
	}
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Tokens = syntax.Tokens[bodyTokenStart:]
	if ability.Optional {
		if len(ability.Effects) != 1 ||
			len(bodySyntax.Tokens) < 3 ||
			!equalTokenWord(bodySyntax.Tokens[0], "you") ||
			!equalTokenWord(bodySyntax.Tokens[1], "may") ||
			ability.OptionalSpan.Start != ability.Effects[0].Span.Start {
			return oracle.CompiledAbility{}, oracle.Ability{}, false
		}
		effect := body.Effects[0]
		effect.Text = effect.Text[effect.VerbSpan.Start.Offset-effect.Span.Start.Offset:]
		effect.Span.Start = effect.VerbSpan.Start
		body.Effects = []oracle.CompiledEffect{effect}
		body.Span.Start = effect.Span.Start
		body.Text = titleFirst(
			ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
		)
		bodySyntax.Tokens = bodySyntax.Tokens[2:]
	}
	body.Keywords = keywordsWithinSpan(ability.Keywords, body.Span)
	if len(body.Keywords) != len(ability.Keywords) {
		return oracle.CompiledAbility{}, oracle.Ability{}, false
	}
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	return body, bodySyntax, true
}

func lowerNonSelfEnterTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, bool) {
	if ability.Trigger == nil ||
		ability.Trigger.Kind != oracle.TriggerWhenever ||
		len(ability.Effects) == 0 ||
		len(ability.Modes) != 0 {
		return game.TriggeredAbility{}, false
	}

	event := ability.Trigger.Event
	pattern := game.TriggerPattern{
		Event: game.EventPermanentEnteredBattlefield,
	}

	if strings.HasPrefix(event, "one or more ") {
		pattern.OneOrMore = true
		rest := strings.TrimPrefix(event, "one or more ")
		cardType, controller, ok := parseOneOrMoreEnterSuffix(rest)
		if !ok {
			return game.TriggeredAbility{}, false
		}
		if cardType != "" {
			pattern.RequirePermanentTypes = []types.Card{cardType}
		}
		pattern.Controller = controller
	} else {
		switch {
		case strings.HasPrefix(event, "another "):
			pattern.ExcludeSelf = true
			event = strings.TrimPrefix(event, "another ")
		case strings.HasPrefix(event, "a "):
			event = strings.TrimPrefix(event, "a ")
		case strings.HasPrefix(event, "an "):
			event = strings.TrimPrefix(event, "an ")
		default:
			return game.TriggeredAbility{}, false
		}
		if strings.HasPrefix(event, "nontoken ") {
			pattern.RequireNonToken = true
			event = strings.TrimPrefix(event, "nontoken ")
		}
		cardType, controller, ok := parseSingleEnterSuffix(event)
		if !ok {
			return game.TriggeredAbility{}, false
		}
		if cardType != "" {
			pattern.RequirePermanentTypes = []types.Card{cardType}
		}
		pattern.Controller = controller
	}

	intervening, supportedCondition := lowerEnterInterveningCondition(ability.Trigger)
	if !supportedCondition ||
		(ability.Trigger.Condition != nil && ability.Trigger.Condition.Text == "if you cast it") {
		return game.TriggeredAbility{}, false
	}
	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, false
	}
	content, contentOK := lowerEventPermanentModifyPTBody(body)
	if !contentOK {
		var diagnostic *oracle.Diagnostic
		content, diagnostic = lowerSelfTriggerBody(cardName, game.EventPermanentEnteredBattlefield, body, bodySyntax)
		if diagnostic != nil {
			return game.TriggeredAbility{}, false
		}
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                                 game.TriggerWhenever,
			Pattern:                              pattern,
			InterveningIf:                        interveningIfText(ability.Trigger),
			InterveningCondition:                 intervening.condition,
			InterveningIfEventPermanentWasKicked: intervening.wasKicked,
			InterveningIfEventPermanentWasCast:   intervening.wasCast,
		},
		Optional: ability.Optional,
		Content:  content,
	}, true
}

// lowerCastTrigger lowers a "whenever ... casts ..." triggered ability into a
// game.TriggeredAbility with EventSpellCast. It returns false for any event
// string it cannot fully represent, including self-cast (TriggerWhen), all
// forms other than the three accepted player prefixes, unsupported spell
// phrases, intervening-if conditions, keywords, modes, and ability words.
func lowerCastTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, bool) {
	if ability.Trigger == nil ||
		ability.Trigger.Kind != oracle.TriggerWhenever ||
		ability.Trigger.Condition != nil ||
		len(ability.Effects) == 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.TriggeredAbility{}, false
	}

	event := ability.Trigger.Event
	var controller game.TriggerControllerFilter
	switch {
	case strings.HasPrefix(event, "you cast "):
		controller = game.TriggerControllerYou
		event = strings.TrimPrefix(event, "you cast ")
	case strings.HasPrefix(event, "a player casts "):
		controller = game.TriggerControllerAny
		event = strings.TrimPrefix(event, "a player casts ")
	case strings.HasPrefix(event, "an opponent casts "):
		controller = game.TriggerControllerOpponent
		event = strings.TrimPrefix(event, "an opponent casts ")
	default:
		return game.TriggeredAbility{}, false
	}

	cardSelection, ok := parseCastSpellSelection(event)
	if !ok {
		return game.TriggeredAbility{}, false
	}

	body, bodySyntax, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, false
	}
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, false
	}

	pattern := game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: controller,
	}
	if !cardSelection.Empty() {
		pattern.CardSelection = cardSelection
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhenever,
			Pattern: pattern,
		},
		Optional: ability.Optional,
		Content:  content,
	}, true
}

// castSpellPhrases maps an oracle spell-phrase fragment to the corresponding
// game.Selection. "a spell" maps to an empty selection (wildcard).
var castSpellPhrases = map[string]game.Selection{
	"a spell":                      {},
	"a noncreature spell":          {ExcludedTypes: []types.Card{types.Creature}},
	"a creature spell":             {RequiredTypes: []types.Card{types.Creature}},
	"an instant or sorcery spell":  {RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
	"an instant spell":             {RequiredTypes: []types.Card{types.Instant}},
	"an instant":                   {RequiredTypes: []types.Card{types.Instant}},
	"a sorcery spell":              {RequiredTypes: []types.Card{types.Sorcery}},
	"an artifact spell":            {RequiredTypes: []types.Card{types.Artifact}},
	"an enchantment spell":         {RequiredTypes: []types.Card{types.Enchantment}},
	"a land spell":                 {RequiredTypes: []types.Card{types.Land}},
	"a planeswalker spell":         {RequiredTypes: []types.Card{types.Planeswalker}},
	"a noncreature, nonland spell": {ExcludedTypes: []types.Card{types.Creature, types.Land}},
	"a white spell":                {ColorsAny: []color.Color{color.White}},
	"a blue spell":                 {ColorsAny: []color.Color{color.Blue}},
	"a black spell":                {ColorsAny: []color.Color{color.Black}},
	"a red spell":                  {ColorsAny: []color.Color{color.Red}},
	"a green spell":                {ColorsAny: []color.Color{color.Green}},
}

// parseCastSpellSelection maps the spell-phrase fragment (what follows the
// player+casts prefix) to a game.Selection. It returns false for any
// unrecognized or unsupported phrase.
func parseCastSpellSelection(phrase string) (game.Selection, bool) {
	sel, ok := castSpellPhrases[phrase]
	return sel, ok
}

// lowerEventPermanentModifyPTBody handles the narrow case of a triggered body
// that modifies the entering permanent via the pronoun "it", e.g.
// "It gets +2/+0 until end of turn." The pronoun resolves to
// game.EventPermanentReference(), which identifies the permanent named by the
// triggering event. Only exact fixed static P/T changes until end of turn are
// accepted.
func lowerEventPermanentModifyPTBody(body oracle.CompiledAbility) (game.AbilityContent, bool) {
	if len(body.Effects) != 1 ||
		body.Effects[0].Kind != oracle.EffectModifyPT ||
		len(body.Targets) != 0 ||
		len(body.References) != 1 ||
		body.References[0].Kind != oracle.ReferencePronoun ||
		!strings.EqualFold(body.References[0].Text, "it") ||
		len(body.Conditions) != 0 ||
		len(body.Keywords) != 0 ||
		len(body.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := body.Effects[0]
	if !effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		effect.Negated ||
		effect.Duration != oracle.DurationUntilEndOfTurn {
		return game.AbilityContent{}, false
	}
	want := fmt.Sprintf("It gets %s/%s until end of turn.",
		signedAmountText(effect.PowerDelta),
		signedAmountText(effect.ToughnessDelta))
	if body.Text != want {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         game.EventPermanentReference(),
				PowerDelta:     game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
				ToughnessDelta: game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)),
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

// parseSingleEnterSuffix parses "{type} {controller?} enters" from the event
// fragment after an article and optional "nontoken" have been stripped.
// An empty card type return signals "permanent" (no type filter).
func parseSingleEnterSuffix(event string) (types.Card, game.TriggerControllerFilter, bool) {
	controller := game.TriggerControllerAny
	if s, ok := strings.CutSuffix(event, " you control enters"); ok {
		event = s + " enters"
		controller = game.TriggerControllerYou
	} else if s, ok := strings.CutSuffix(event, " an opponent controls enters"); ok {
		event = s + " enters"
		controller = game.TriggerControllerOpponent
	}
	cardType, ok := permanentEnterTypeWord(event)
	return cardType, controller, ok
}

// parseOneOrMoreEnterSuffix parses "{type_plural} {controller?} enter" from
// the fragment after "one or more " has been stripped.
// An empty card type return signals "permanents" (no type filter).
func parseOneOrMoreEnterSuffix(event string) (types.Card, game.TriggerControllerFilter, bool) {
	controller := game.TriggerControllerAny
	if s, ok := strings.CutSuffix(event, " you control enter"); ok {
		event = s + " enter"
		controller = game.TriggerControllerYou
	} else if s, ok := strings.CutSuffix(event, " an opponent controls enter"); ok {
		event = s + " enter"
		controller = game.TriggerControllerOpponent
	}
	cardType, ok := permanentEnterTypePlural(event)
	return cardType, controller, ok
}

func permanentEnterTypeWord(event string) (types.Card, bool) {
	switch event {
	case "creature enters":
		return types.Creature, true
	case "artifact enters":
		return types.Artifact, true
	case "enchantment enters":
		return types.Enchantment, true
	case "land enters":
		return types.Land, true
	case "permanent enters":
		return "", true
	case "planeswalker enters":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

func permanentEnterTypePlural(event string) (types.Card, bool) {
	switch event {
	case "creatures enter":
		return types.Creature, true
	case "artifacts enter":
		return types.Artifact, true
	case "enchantments enter":
		return types.Enchantment, true
	case "lands enter":
		return types.Land, true
	case "permanents enter":
		return "", true
	case "planeswalkers enter":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

func spanCovered(span oracle.Span, covering []oracle.Span) bool {
	for _, candidate := range covering {
		if candidate.Start.Offset <= span.Start.Offset &&
			candidate.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func lowerKeywordAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) ([]loweredStaticAbility, *oracle.Diagnostic) {
	for _, keyword := range ability.Keywords {
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
	if len(ability.Modes) > 0 {
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
	if len(ability.Keywords) == 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend does not yet lower non-keyword static rules text",
		)
	}
	bodies := make([]loweredStaticAbility, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		if keyword.Parameter != "" {
			if keyword.Name == "Ward" {
				manaCost, err := parseManaCostValue(keyword.Parameter)
				if err == nil && len(manaCost) > 0 {
					bodies = append(bodies, loweredStaticAbility{
						Body: game.WardStaticAbility(manaCost),
					})
					continue
				}
			}
			if keyword.Name == "Protection" {
				protectedColors, ok := oracleColors(keyword.Parameter)
				if ok {
					bodies = append(bodies, loweredStaticAbility{
						Body: game.ProtectionFromColorsStaticAbility(protectedColors...),
					})
					continue
				}
			}
			if body, ok := lowerParameterizedStaticKeyword(keyword); ok {
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
	if len(ability.Targets) > 0 ||
		len(ability.Conditions) > 0 ||
		len(ability.Effects) > 0 ||
		len(ability.References) > 0 {
		return nil, mixedKeywordDiagnostic(ability)
	}
	for _, token := range syntax.Tokens {
		if token.Kind == oracle.Comma ||
			(syntax.AbilityWord != nil && token.Kind == oracle.EmDash) ||
			spanCoveredByAbilityWord(token.Span, syntax.AbilityWord) ||
			spanCoveredByKeyword(token.Span, ability.Keywords) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return nil, mixedKeywordDiagnostic(ability)
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

func lowerParameterizedStaticKeyword(keyword oracle.CompiledKeyword) (game.StaticAbility, bool) {
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

func lowerTapManaAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ManaAbility, *oracle.Diagnostic) {
	if ability.Cost == nil ||
		len(ability.Cost.Components) != 1 ||
		(ability.Cost.Components[0].Kind != oracle.CostTap &&
			ability.Cost.Components[0].Kind != oracle.CostUntap) ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectAddMana ||
		!ability.Effects[0].Amount.Known ||
		ability.Effects[0].Amount.Value != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" ||
		len(ability.Keywords) != 0 ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.References) != 0 {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activated ability",
			"the executable source backend supports only exact supported tap and untap mana abilities",
		)
	}
	costSymbol := "{T}"
	if ability.Cost.Components[0].Kind == oracle.CostUntap {
		costSymbol = "{Q}"
	}
	if ability.ActivationTiming != oracle.ActivationTimingNone {
		syntax.Tokens = slices.DeleteFunc(
			append([]oracle.Token(nil), syntax.Tokens...),
			func(token oracle.Token) bool {
				return spanCovered(token.Span, []oracle.Span{ability.ActivationTimingSpan})
			},
		)
	}
	if exactAnyColorManaSyntax(syntax.Tokens, costSymbol) {
		result := choiceTapManaAbility(
			[]string{"W", "U", "B", "R", "G"},
		)
		applyManaAbilityProperties(&result, ability)
		return result, nil
	}
	if colors, ok := exactChoiceManaSyntax(syntax.Tokens, costSymbol); ok {
		result := choiceTapManaAbility(colors)
		applyManaAbilityProperties(&result, ability)
		return result, nil
	}
	if !exactManaSyntax(syntax.Tokens, costSymbol) {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activated ability",
			"the executable source backend supports only exact supported tap and untap mana abilities",
		)
	}
	colorName, ok := manaColorName(ability.Effects[0].Symbol)
	if !ok {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana symbol",
			fmt.Sprintf("the executable source backend cannot emit mana symbol %q", ability.Effects[0].Symbol),
		)
	}
	manaColor, ok := manaColorValue(colorName)
	if !ok {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana symbol",
			fmt.Sprintf("the executable source backend cannot emit mana symbol %q", ability.Effects[0].Symbol),
		)
	}
	result := game.TapManaAbility(manaColor)
	applyManaAbilityProperties(&result, ability)
	return result, nil
}

func applyManaAbilityProperties(result *game.ManaAbility, ability oracle.CompiledAbility) {
	result.Text = ability.Text
	result.Timing = lowerActivationTiming(ability.ActivationTiming)
	if ability.Cost.Components[0].Kind == oracle.CostUntap {
		result.AdditionalCosts = []cost.Additional{{
			Kind: cost.AdditionalUntap,
			Text: ability.Cost.Components[0].Text,
		}}
	}
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

func lowerSpell(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	if exactManifestDreadLongFormPattern(syntax.Tokens) &&
		len(ability.Targets) == 0 &&
		len(ability.Conditions) == 0 &&
		len(ability.Keywords) == 0 &&
		len(ability.Modes) == 0 {
		return manifestDreadAbility(), nil
	}
	if len(ability.Effects) > 1 {
		return lowerOrderedEffectSequence(cardName, ability, syntax)
	}
	if len(ability.Effects) == 1 {
		return lowerSingleEffectSpell(cardName, ability, syntax)
	}
	return game.AbilityContent{}, executableDiagnostic(
		ability,
		"unsupported spell ability",
		"the executable source backend does not yet lower this spell ability",
	)
}

func lowerSingleEffectSpell(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	ability.Text = textWithoutDelimited(ability.Text, ability.Span, syntax.Reminders)
	syntax.Tokens = slices.DeleteFunc(
		append([]oracle.Token(nil), syntax.Tokens...),
		func(token oracle.Token) bool {
			return spanCoveredByDelimited(token.Span, syntax.Reminders)
		},
	)
	switch ability.Effects[0].Kind {
	case oracle.EffectDealDamage:
		return lowerFixedDamageSpell(cardName, ability)
	case oracle.EffectDraw:
		return lowerFixedDrawSpell(ability, syntax)
	case oracle.EffectDestroy:
		return lowerFixedDestroySpell(ability)
	case oracle.EffectGain:
		return lowerFixedLifeSpell(ability, "gain", func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.GainLife{Amount: amount, Player: player}
		})
	case oracle.EffectLose:
		return lowerFixedLifeSpell(ability, "lose", func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.LoseLife{Amount: amount, Player: player}
		})
	case oracle.EffectScry:
		return lowerFixedControllerSpell(ability, syntax, "scry", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.Scry{Amount: amount, Player: player}
		})
	case oracle.EffectSurveil:
		return lowerFixedControllerSpell(ability, syntax, "surveil", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.Surveil{Amount: amount, Player: player}
		})
	case oracle.EffectInvestigate:
		return lowerInvestigateSpell(ability, syntax)
	case oracle.EffectProliferate:
		return lowerExactPrimitiveSpell(ability, syntax, "proliferate", func(amount game.Quantity) game.Primitive {
			return game.Proliferate{Amount: amount}
		})
	case oracle.EffectExplore:
		return lowerExploreSpell(ability, syntax)
	case oracle.EffectManifest, oracle.EffectManifestDread:
		return lowerManifestSpell(ability, syntax)
	case oracle.EffectRegenerate:
		return lowerFixedPermanentTargetSpell(ability, "Regenerate", func(object game.ObjectReference) game.Primitive {
			return game.Regenerate{Object: object}
		})
	case oracle.EffectFight:
		return lowerFightSpell(ability)
	case oracle.EffectDiscard:
		return lowerFixedCardCountPlayerSpell(
			ability, syntax, "discard", "discards", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Discard{Amount: amount, Player: player}
			},
		)
	case oracle.EffectMill:
		return lowerFixedCardCountPlayerSpell(
			ability, syntax, "mill", "mills", true, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Mill{Amount: amount, Player: player}
			},
		)
	case oracle.EffectTap:
		return lowerFixedPermanentTargetSpell(ability, "Tap", func(object game.ObjectReference) game.Primitive {
			return game.Tap{Object: object}
		})
	case oracle.EffectUntap:
		return lowerFixedPermanentTargetSpell(ability, "Untap", func(object game.ObjectReference) game.Primitive {
			return game.Untap{Object: object}
		})
	case oracle.EffectExile:
		return lowerFixedPermanentTargetSpell(ability, "Exile", func(object game.ObjectReference) game.Primitive {
			return game.Exile{Object: object}
		})
	case oracle.EffectReturn:
		if content, ok := lowerSelfCardGraveyardReturn(ability); ok {
			return content, nil
		}
		if content, ok := lowerTargetedGraveyardReturn(ability); ok {
			return content, nil
		}
		return lowerFixedBounceSpell(ability)
	case oracle.EffectPut:
		if content, ok := lowerTargetedGraveyardReturn(ability); ok {
			return content, nil
		}
		return lowerCounterPlacementSpell(ability)
	case oracle.EffectModifyPT:
		return lowerFixedModifyPTSpell(ability)
	default:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported spell ability",
			"the executable source backend does not yet lower this spell ability",
		)
	}
}

func lowerSelfCardGraveyardReturn(ability oracle.CompiledAbility) (game.AbilityContent, bool) {
	if len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectReturn ||
		len(ability.Targets) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.Conditions) != 0 ||
		!selfCardGraveyardReturnReferences(ability.References) {
		return game.AbilityContent{}, false
	}
	sourceCard := game.CardReference{Kind: game.CardReferenceSource}
	switch {
	case ability.Text == "Return this card from your graveyard to your hand.":
		return game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
			Card:        sourceCard,
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}}}}.Ability(), true
	case ability.Text == "Return this card from your graveyard to the battlefield.":
		return game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(sourceCard),
		}}}}.Ability(), true
	case strings.HasPrefix(ability.Text, "Return this card from your graveyard to the battlefield"):
		tapped, counters, ok := selfCardBattlefieldReturnModifiers(ability.Text)
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

func selfCardGraveyardReturnReferences(references []oracle.CompiledReference) bool {
	if len(references) == 0 || references[0].Kind != oracle.ReferenceThisObject {
		return false
	}
	for _, reference := range references[1:] {
		if reference.Kind != oracle.ReferencePronoun || reference.Text != "it" {
			return false
		}
	}
	return true
}

func selfCardBattlefieldReturnModifiers(text string) (tapped bool, counters int, ok bool) {
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
		word := strings.TrimSuffix(strings.TrimPrefix(suffix, " with "), " +1/+1 counters on it.")
		amount, amountOK := smallNumberWord(word)
		return tapped, amount, amountOK
	}
	return false, 0, false
}

func smallNumberWord(word string) (int, bool) {
	switch word {
	case "one", "a", "an":
		return 1, true
	case "two":
		return 2, true
	case "three":
		return 3, true
	case "four":
		return 4, true
	case "five":
		return 5, true
	default:
		return 0, false
	}
}

func lowerTargetedGraveyardReturn(ability oracle.CompiledAbility) (game.AbilityContent, bool) {
	if len(ability.Targets) != 1 ||
		len(ability.Effects) != 1 ||
		len(ability.Modes) != 0 ||
		len(ability.Conditions) != 0 ||
		ability.Effects[0].FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := cardInZoneTargetSpec(ability.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	switch ability.Effects[0].ToZone {
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
		destinationBottom, ok := graveyardReturnLibraryBottom(ability.Targets[0].Text)
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
			put, ok := targetedGraveyardBattlefieldPut(ability.Text, game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i})
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

func cardInZoneTargetSpec(target oracle.CompiledTarget, targetZone zone.Type) (game.TargetSpec, bool) {
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
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		(effect.Amount.Known && effect.Amount.Value <= 0) ||
		!effect.CounterKindKnown ||
		!oracle.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Modes) != 0 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ability)
	}

	kind := effect.CounterKind
	counterName := kind.String()
	var target game.TargetSpec
	var primitive game.Primitive
	if kind.PlayerOnly() {
		var ok bool
		target, ok = playerTargetSpec(ability.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ability)
		}
	} else {
		var ok bool
		target, ok = permanentTargetSpec(ability.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ability)
		}
	}

	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	exactText := exactXCounterText(ability, counterName) && len(ability.References) == 0
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
		exactText = isExactPutCounterText(
			ability.Text,
			ability.Targets[0].Text,
			effect.Amount.Value,
			counterName,
		) && len(ability.References) == 0
	} else if effect.Amount.DynamicKind != oracle.DynamicAmountNone {
		dynamic, supported := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !supported ||
			!exactDynamicCounterText(ability, counterName) ||
			!exactDynamicAmountReference(effect.Amount, ability.References) {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ability)
		}
		amount = game.Dynamic(dynamic)
		exactText = true
	}
	if !exactText {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ability)
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

func unsupportedCounterPlacementDiagnostic(ability oracle.CompiledAbility) *oracle.Diagnostic {
	return executableDiagnostic(
		ability,
		"unsupported counter placement",
		"the executable source backend supports exact recognized counter placement on one valid target",
	)
}

func exactXCounterText(ability oracle.CompiledAbility, counterName string) bool {
	return ability.Text == fmt.Sprintf(
		"Put X %s counters on %s.",
		counterName,
		ability.Targets[0].Text,
	)
}

func exactDynamicCounterText(ability oracle.CompiledAbility, counterName string) bool {
	amount := ability.Effects[0].Amount
	return amount.DynamicForm == oracle.DynamicAmountWhereX &&
		ability.Text == fmt.Sprintf(
			"Put X %s counters on %s, %s.",
			counterName,
			ability.Targets[0].Text,
			amount.Text,
		)
}

func exactDynamicAmountReference(
	amount oracle.CompiledAmount,
	references []oracle.CompiledReference,
) bool {
	if amount.DynamicKind != oracle.DynamicAmountSourcePower {
		return len(references) == 0
	}
	if len(references) != 1 || references[0].Span != amount.ReferenceSpan {
		return false
	}
	switch references[0].Kind {
	case oracle.ReferenceSelfName, oracle.ReferenceThisObject:
		return true
	default:
		return false
	}
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

func textWithoutDelimited(text string, span oracle.Span, groups []oracle.Delimited) string {
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

func lowerFightSpell(ability oracle.CompiledAbility) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Targets) != 2 ||
		ability.Targets[0].Cardinality != (oracle.TargetCardinality{Min: 1, Max: 1}) ||
		ability.Targets[1].Cardinality != (oracle.TargetCardinality{Min: 1, Max: 1}) ||
		ability.Effects[0].Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Text != titleFirst(ability.Targets[0].Text)+" fights "+ability.Targets[1].Text+"." {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported fight spell",
			"the executable source backend supports only exact fights between two target creatures",
		)
	}
	first, firstOK := fightCreatureTargetSpec(ability.Targets[0])
	second, secondOK := fightCreatureTargetSpec(ability.Targets[1])
	if !firstOK || !secondOK {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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

func fightCreatureTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	if target.Selector.Kind != oracle.SelectorCreature ||
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
	case oracle.ControllerAny:
		expected = "target creature"
	case oracle.ControllerYou:
		expected = "target creature you control"
		spec.Predicate.Controller = game.ControllerYou
	case oracle.ControllerOpponent:
		expected = "target creature an opponent controls"
		spec.Predicate.Controller = game.ControllerOpponent
	case oracle.ControllerNotYou:
		expected = "target creature you don't control"
		spec.Predicate.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	return spec, strings.EqualFold(target.Text, expected)
}

func lowerInvestigateSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	return lowerExactPrimitiveSpell(
		ability,
		syntax,
		"investigate",
		func(amount game.Quantity) game.Primitive {
			return game.Investigate{Amount: amount}
		},
	)
}

func lowerExploreSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	tokens := syntax.Tokens
	if ability.Effects[0].Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(tokens) != 3 ||
		!equalTokenWord(tokens[0], "it") ||
		!equalTokenWord(tokens[1], "explores") ||
		tokens[2].Kind != oracle.Period ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferencePronoun {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported explore spell",
			"the executable source backend supports only the source permanent pattern \"it explores\"",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Explore{Creature: game.SourcePermanentReference()},
	}}}.Ability(), nil
}

func lowerManifestSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	tokens := syntax.Tokens
	if ability.Effects[0].Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		!exactManifestTopLibraryPattern(tokens) &&
			!exactManifestDreadShorthandPattern(tokens) &&
			!exactManifestDreadLongFormPattern(tokens) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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

func exactManifestTopLibraryPattern(tokens []oracle.Token) bool {
	return len(tokens) == 8 &&
		equalTokenWord(tokens[0], "manifest") &&
		equalTokenWord(tokens[1], "the") &&
		equalTokenWord(tokens[2], "top") &&
		equalTokenWord(tokens[3], "card") &&
		equalTokenWord(tokens[4], "of") &&
		equalTokenWord(tokens[5], "your") &&
		equalTokenWord(tokens[6], "library") &&
		tokens[7].Kind == oracle.Period
}

func exactManifestDreadShorthandPattern(tokens []oracle.Token) bool {
	return len(tokens) == 3 &&
		equalTokenWord(tokens[0], "manifest") &&
		equalTokenWord(tokens[1], "dread") &&
		tokens[2].Kind == oracle.Period
}

func exactManifestDreadLongFormPattern(tokens []oracle.Token) bool {
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
		tokens[9].Kind == oracle.Period &&
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
		tokens[21].Kind == oracle.Integer &&
		tokens[21].Text == "2" &&
		tokens[22].Kind == oracle.Slash &&
		tokens[23].Kind == oracle.Integer &&
		tokens[23].Text == "2" &&
		equalTokenWord(tokens[24], "creature") &&
		tokens[25].Kind == oracle.Period &&
		equalTokenWord(tokens[26], "put") &&
		equalTokenWord(tokens[27], "the") &&
		equalTokenWord(tokens[28], "other") &&
		equalTokenWord(tokens[29], "into") &&
		equalTokenWord(tokens[30], "your") &&
		equalTokenWord(tokens[31], "graveyard") &&
		tokens[32].Kind == oracle.Period
}

func lowerExactPrimitiveSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	verb string,
	primitiveFactory func(game.Quantity) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	amount, ok := standaloneActionAmount(syntax.Tokens, verb)
	if ability.Effects[0].Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		!ok {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact "+verb,
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: primitiveFactory(game.Fixed(amount)),
	}}}.Ability(), nil
}

func standaloneActionAmount(tokens []oracle.Token, verb string) (int, bool) {
	if len(tokens) == 2 &&
		equalTokenWord(tokens[0], verb) &&
		tokens[1].Kind == oracle.Period {
		return 1, true
	}
	if len(tokens) == 3 &&
		equalTokenWord(tokens[0], verb) &&
		tokens[2].Kind == oracle.Period {
		switch strings.ToLower(tokens[1].Text) {
		case "twice":
			return 2, true
		case "thrice":
			return 3, true
		}
	}
	if len(tokens) == 4 &&
		equalTokenWord(tokens[0], verb) &&
		equalTokenWord(tokens[2], "times") &&
		tokens[3].Kind == oracle.Period {
		for amount := 1; amount <= 4; amount++ {
			if fixedNumberToken(tokens[1], amount) {
				return amount, true
			}
		}
	}
	return 0, false
}

func lowerOrderedEffectSequence(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Conditions) != 0 || len(ability.Modes) != 0 {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
	}
	if content, ok := lowerCyclingCountDamageAndGain(cardName, ability); ok {
		return content, nil
	}
	var targets []game.TargetSpec
	var sequence []game.Instruction
	consumedTargets := 0
	consumedKeywords := 0
	consumedReferences := 0
	for _, effect := range ability.Effects {
		effectAbility := abilityForEffect(ability, effect)
		consumedTargets += len(effectAbility.Targets)
		consumedKeywords += len(effectAbility.Keywords)
		consumedReferences += len(effectAbility.References)
		content, diagnostic := lowerSingleEffectSpell(
			cardName,
			effectAbility,
			syntaxWithinSpan(syntax, effect.Span),
		)
		if diagnostic != nil ||
			len(content.SharedTargets) != 0 ||
			content.IsModal() ||
			len(content.Modes) != 1 {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
		}
		mode := content.Modes[0]
		if len(mode.Targets) > 0 {
			if !rebaseTargetedSequence(mode.Sequence, len(targets)) {
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
			}
			targets = append(targets, mode.Targets...)
		}
		sequence = append(sequence, mode.Sequence...)
	}
	if consumedTargets != len(ability.Targets) ||
		consumedKeywords != len(ability.Keywords) ||
		consumedReferences != len(ability.References) ||
		len(sequence) != len(ability.Effects) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
	}
	return game.Mode{Targets: targets, Sequence: sequence}.Ability(), nil
}

func lowerCyclingCountDamageAndGain(cardName string, ability oracle.CompiledAbility) (game.AbilityContent, bool) {
	if len(ability.Effects) != 2 ||
		ability.Effects[0].Kind != oracle.EffectDealDamage ||
		ability.Effects[1].Kind != oracle.EffectGain ||
		ability.Effects[0].Negated ||
		ability.Effects[1].Negated ||
		len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		len(ability.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ability)) != 0 ||
		len(ability.Modes) != 0 ||
		!singleSelfReference(ability.References) {
		return game.AbilityContent{}, false
	}
	amountEffect := ability.Effects[1].Amount
	if amountEffect.DynamicKind == oracle.DynamicAmountNone ||
		amountEffect.DynamicForm != oracle.DynamicAmountWhereX {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(amountEffect, game.SourcePermanentReference())
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := damageTargetSpec(ability.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	if ability.Text != fmt.Sprintf(
		"%s deals X damage to %s and you gain X life, %s.",
		cardName,
		ability.Targets[0].Text,
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

func abilityForEffect(
	ability oracle.CompiledAbility,
	effect oracle.CompiledEffect,
) oracle.CompiledAbility {
	ability.Text = effect.Text
	ability.Span = effect.Span
	ability.Effects = []oracle.CompiledEffect{effect}
	ability.Targets = targetsWithinSpan(ability.Targets, effect.Span)
	ability.Keywords = keywordsWithinSpan(ability.Keywords, effect.Span)
	ability.References = referencesWithinSpan(ability.References, effect.Span)
	return ability
}

func targetsWithinSpan(targets []oracle.CompiledTarget, span oracle.Span) []oracle.CompiledTarget {
	var within []oracle.CompiledTarget
	for _, target := range targets {
		if spanCovered(target.Span, []oracle.Span{span}) {
			within = append(within, target)
		}
	}
	return within
}

func keywordsWithinSpan(keywords []oracle.CompiledKeyword, span oracle.Span) []oracle.CompiledKeyword {
	var within []oracle.CompiledKeyword
	for _, keyword := range keywords {
		if spanCovered(keyword.Span, []oracle.Span{span}) {
			within = append(within, keyword)
		}
	}
	return within
}

func referencesWithinSpan(references []oracle.CompiledReference, span oracle.Span) []oracle.CompiledReference {
	var within []oracle.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, []oracle.Span{span}) {
			within = append(within, reference)
		}
	}
	return within
}

func syntaxWithinSpan(syntax oracle.Ability, span oracle.Span) oracle.Ability {
	syntax.Span = span
	syntax.Text = ""
	syntax.Tokens = slices.DeleteFunc(
		append([]oracle.Token(nil), syntax.Tokens...),
		func(token oracle.Token) bool {
			return !spanCovered(token.Span, []oracle.Span{span})
		},
	)
	return syntax
}

func unsupportedEffectSequenceDiagnostic(ability oracle.CompiledAbility) *oracle.Diagnostic {
	return executableDiagnostic(
		ability,
		"unsupported ordered effect sequence",
		"the executable source backend supports only exact ordered sequences of independently supported effects",
	)
}

func lowerFixedDamageSpell(
	cardName string,
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if len(ability.Effects) != 1 ||
		effect.Kind != oracle.EffectDealDamage ||
		(effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		len(ability.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ability)) != 0 ||
		len(ability.Modes) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	amountText := "X"
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
		amountText = fmt.Sprint(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != oracle.DynamicAmountNone {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	target, ok := damageTargetSpec(ability.Targets[0])
	if !ok ||
		!exactDamageAmountSyntax(cardName, ability, effect.Amount, amountText) ||
		!exactDamageAmountReferences(effect.Amount, ability.References) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	damage := game.Damage{
		Amount:    amount,
		Recipient: game.AnyTargetDamageRecipient(0),
	}
	if effect.Amount.DynamicKind == oracle.DynamicAmountSourcePower {
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
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	dynamicPT := effect.Amount.DynamicKind != oracle.DynamicAmountNone
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Targets[0].Selector.Kind != oracle.SelectorCreature ||
		(!dynamicPT && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) ||
		effect.Negated ||
		effect.Duration != oracle.DurationUntilEndOfTurn ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		!exactModifyPTAmountSyntax(ability, effect) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	targetSpec, ok := permanentTargetSpec(ability.Targets[0])
	if !ok {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	powerDelta := game.Fixed(compiledSignedAmountValue(effect.PowerDelta))
	toughnessDelta := game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta))
	if dynamicPT {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == oracle.DynamicAmountSourcePower {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
		switch effect.Amount.DynamicForm {
		case oracle.DynamicAmountWhereX:
			powerDelta = game.Dynamic(dynamic)
			toughnessDelta = game.Dynamic(dynamic)
		case oracle.DynamicAmountForEach:
			powerDelta = dynamicSignedQuantity(dynamic, effect.PowerDelta)
			toughnessDelta = dynamicSignedQuantity(dynamic, effect.ToughnessDelta)
		default:
			return game.AbilityContent{}, executableDiagnostic(
				ability,
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

func lowerFixedBounceSpell(
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferencePronoun ||
		!strings.EqualFold(ability.References[0].Text, "its") {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	target := ability.Targets[0]
	target.Text = strings.TrimSuffix(target.Text, " to its owner's hand")
	targetSpec, ok := permanentTargetSpec(target)
	if !ok || ability.Text != "Return "+target.Text+" to its owner's hand." {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Bounce{
					Object: game.TargetPermanentReference(0),
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedPermanentTargetSpell(
	ability oracle.CompiledAbility,
	verb string,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ability.Targets[0])
	if !ok || ability.Text != verb+" "+ability.Targets[0].Text+"." {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	controllerVerb string,
	targetVerb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	amount, ok := cardCountQuantity(effect.Amount, allowDynamic)
	if !ok {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		(exactCardCountPlayerSyntax(syntax.Tokens, controllerVerb, effect.Amount) ||
			exactDynamicCardCountPlayerText(ability.Text, "", controllerVerb, effect.Amount)):
	case len(ability.Targets) == 1 &&
		(exactTargetCardCountPlayerSyntax(syntax.Tokens, targetVerb, effect.Amount) ||
			exactDynamicCardCountPlayerText(ability.Text, titleFirst(ability.Targets[0].Text), targetVerb, effect.Amount)) &&
		strings.EqualFold(syntax.Tokens[0].Text, "target") &&
		strings.EqualFold(syntax.Tokens[1].Text, "player"):
		targetSpec, ok := playerTargetSpec(ability.Targets[0])
		if !ok {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{targetSpec}
	default:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	verb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	amount, ok := controllerActionQuantity(effect.Amount, allowDynamic)
	if !ok || !exactControllerAmountSyntax(syntax.Tokens, ability.Text, verb, effect.Amount) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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

func cardCountQuantity(amount oracle.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == oracle.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == oracle.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

func controllerActionQuantity(amount oracle.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == oracle.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == oracle.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

func lowerFixedLifeSpell(
	ability oracle.CompiledAbility,
	verb string,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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
	case effect.Amount.DynamicKind != oracle.DynamicAmountNone:
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == oracle.DynamicAmountSourcePower ||
			len(ability.References) != 0 {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported life spell",
				"the executable source backend supports only exact supported life changes",
			)
		}
		amount = game.Dynamic(dynamic)
	case len(ability.References) != 0:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported life spell",
			"the executable source backend supports only exact supported life changes",
		)
	default:
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		exactLifeAmountSyntax("You", verb, ability.Text, effect.Amount, amountText):
	case len(ability.Targets) == 1:
		targetSpec, ok := playerTargetSpec(ability.Targets[0])
		if !ok ||
			!exactLifeAmountSyntax(
				titleFirst(ability.Targets[0].Text),
				verb+"s",
				ability.Text,
				effect.Amount,
				amountText,
			) {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		targets = []game.TargetSpec{targetSpec}
		playerRef = game.TargetPlayerReference(0)
	default:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	if group, ok := exactMassDestroyGroup(ability); ok {
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
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].Negated {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ability.Targets[0])
	if !ok || ability.Text != "Destroy "+ability.Targets[0].Text+"." {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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

func exactMassDestroyGroup(ability oracle.CompiledAbility) (game.GroupReference, bool) {
	if len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].Negated {
		return game.GroupReference{}, false
	}
	switch ability.Text {
	case "Destroy all creatures.":
		return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}), true
	case "Destroy all artifacts.":
		return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}), true
	case "Destroy all enchantments.":
		return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}), true
	default:
		return game.GroupReference{}, false
	}
}

func lowerFixedDrawSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != oracle.DynamicAmountNone {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == oracle.DynamicAmountSourcePower {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported draw spell",
				"the executable source backend supports only exact supported card draw",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		(exactControllerDrawSyntax(syntax.Tokens, effect.Amount.Value) ||
			(!effect.Amount.Known &&
				effect.Amount.DynamicKind == oracle.DynamicAmountNone &&
				exactXControllerDrawSyntax(syntax.Tokens)) ||
			exactDynamicDrawSyntax(ability.Text, "", effect.Amount)):
	case len(ability.Targets) == 1 &&
		(exactTargetPlayerDrawSyntax(syntax.Tokens, effect.Amount.Value) ||
			(!effect.Amount.Known &&
				effect.Amount.DynamicKind == oracle.DynamicAmountNone &&
				exactXTargetPlayerDrawSyntax(syntax.Tokens)) ||
			exactDynamicDrawSyntax(ability.Text, titleFirst(ability.Targets[0].Text), effect.Amount)) &&
		ability.Targets[0].Cardinality.Min == 1 &&
		ability.Targets[0].Cardinality.Max == 1 &&
		ability.Targets[0].Selector.Kind == oracle.SelectorPlayer:
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
		return game.AbilityContent{}, executableDiagnostic(
			ability,
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

func lowerDynamicAmount(amount oracle.CompiledAmount, object game.ObjectReference) (game.DynamicAmount, bool) {
	if amount.Multiplier < 1 {
		return game.DynamicAmount{}, false
	}
	dynamic := game.DynamicAmount{Multiplier: amount.Multiplier}
	switch amount.DynamicKind {
	case oracle.DynamicAmountCount:
		if dynamic, ok := dynamicCardZoneAmount(amount.Selector, amount.Multiplier); ok {
			return dynamic, true
		}
		selection, ok := dynamicAmountSelection(amount.Selector)
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountCountSelector
		dynamic.Group = game.BattlefieldGroup(selection)
	case oracle.DynamicAmountControllerLife:
		dynamic.Kind = game.DynamicAmountControllerLife
	case oracle.DynamicAmountOpponentCount:
		dynamic.Kind = game.DynamicAmountOpponentCount
	case oracle.DynamicAmountSourcePower:
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

func dynamicAmountSelection(selector oracle.CompiledSelector) (game.Selection, bool) {
	var requiredType types.Card
	switch selector.Kind {
	case oracle.SelectorArtifact:
		requiredType = types.Artifact
	case oracle.SelectorCreature:
		requiredType = types.Creature
	case oracle.SelectorEnchantment:
		requiredType = types.Enchantment
	case oracle.SelectorLand:
		requiredType = types.Land
	case oracle.SelectorPermanent:
	default:
		return game.Selection{}, false
	}
	var controller game.ControllerRelation
	switch selector.Controller {
	case oracle.ControllerAny:
	case oracle.ControllerYou:
		controller = game.ControllerYou
	case oracle.ControllerOpponent:
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

func dynamicCardZoneAmount(selector oracle.CompiledSelector, multiplier int) (game.DynamicAmount, bool) {
	if selector.Kind != oracle.SelectorCard || selector.Zone == zone.None {
		return game.DynamicAmount{}, false
	}
	if selector.Zone != zone.Graveyard || selector.Controller != oracle.ControllerYou {
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
	default:
		return game.KeywordNone, false
	}
}

func exactDamageAmountSyntax(
	cardName string,
	ability oracle.CompiledAbility,
	amount oracle.CompiledAmount,
	fixedText string,
) bool {
	target := ability.Targets[0].Text
	switch amount.DynamicForm {
	case oracle.DynamicAmountFormNone:
		return ability.Text == fmt.Sprintf("%s deals %s damage to %s.", cardName, fixedText, target)
	case oracle.DynamicAmountEqual:
		return ability.Text == fmt.Sprintf("%s deals damage %s to %s.", cardName, amount.Text, target)
	case oracle.DynamicAmountForEach:
		return ability.Text == fmt.Sprintf(
			"%s deals %d damage %s to %s.",
			cardName,
			amount.Multiplier,
			amount.Text,
			target,
		)
	case oracle.DynamicAmountWhereX:
		return ability.Text == fmt.Sprintf(
			"%s deals X damage to %s, %s.",
			cardName,
			target,
			amount.Text,
		)
	default:
		return false
	}
}

func exactDamageAmountReferences(amount oracle.CompiledAmount, references []oracle.CompiledReference) bool {
	if amount.DynamicKind != oracle.DynamicAmountSourcePower {
		return singleSelfReference(references)
	}
	if len(references) != 2 ||
		references[0].Kind != oracle.ReferenceSelfName ||
		references[1].Span != amount.ReferenceSpan {
		return false
	}
	switch references[1].Kind {
	case oracle.ReferenceSelfName, oracle.ReferenceThisObject:
		return true
	case oracle.ReferencePronoun:
		return strings.EqualFold(references[1].Text, "its")
	default:
		return false
	}
}

func exactLifeAmountSyntax(
	subject, verb, text string,
	amount oracle.CompiledAmount,
	fixedText string,
) bool {
	switch amount.DynamicForm {
	case oracle.DynamicAmountFormNone:
		return text == fmt.Sprintf("%s %s %s life.", subject, verb, fixedText)
	case oracle.DynamicAmountEqual:
		return text == fmt.Sprintf("%s %s life %s.", subject, verb, amount.Text)
	case oracle.DynamicAmountForEach:
		return text == fmt.Sprintf("%s %s %d life %s.", subject, verb, amount.Multiplier, amount.Text)
	case oracle.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s %s X life, %s.", subject, verb, amount.Text)
	default:
		return false
	}
}

func exactDynamicDrawSyntax(text, subject string, amount oracle.CompiledAmount) bool {
	if amount.DynamicKind == oracle.DynamicAmountNone {
		return false
	}
	prefix := "Draw"
	if subject != "" {
		prefix = subject + " draws"
	}
	switch amount.DynamicForm {
	case oracle.DynamicAmountEqual:
		return text == fmt.Sprintf("%s cards %s.", prefix, amount.Text)
	case oracle.DynamicAmountForEach:
		noun := "cards"
		if amount.Multiplier == 1 {
			return text == fmt.Sprintf("%s 1 card %s.", prefix, amount.Text) ||
				text == fmt.Sprintf("%s a card %s.", prefix, amount.Text)
		}
		return text == fmt.Sprintf("%s %d %s %s.", prefix, amount.Multiplier, noun, amount.Text)
	case oracle.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s X cards, %s.", prefix, amount.Text)
	default:
		return false
	}
}

func exactModifyPTAmountSyntax(ability oracle.CompiledAbility, effect oracle.CompiledEffect) bool {
	subject := titleFirst(ability.Targets[0].Text)
	amount := effect.Amount
	if amount.DynamicKind == oracle.DynamicAmountNone {
		return len(ability.References) == 0 &&
			ability.Text == fmt.Sprintf(
				"%s gets %s/%s until end of turn.",
				subject,
				signedAmountText(effect.PowerDelta),
				signedAmountText(effect.ToughnessDelta),
			)
	}
	if len(ability.References) != 0 || amount.DynamicKind == oracle.DynamicAmountSourcePower {
		return false
	}
	switch amount.DynamicForm {
	case oracle.DynamicAmountForEach:
		if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known ||
			!dynamicPTMultiplierMatches(amount.Multiplier, effect.PowerDelta, effect.ToughnessDelta) {
			return false
		}
		return ability.Text == fmt.Sprintf(
			"%s gets %s/%s %s until end of turn.",
			subject,
			signedAmountText(effect.PowerDelta),
			signedAmountText(effect.ToughnessDelta),
			amount.Text,
		) || ability.Text == fmt.Sprintf(
			"%s gets %s/%s until end of turn %s.",
			subject,
			signedAmountText(effect.PowerDelta),
			signedAmountText(effect.ToughnessDelta),
			amount.Text,
		)
	case oracle.DynamicAmountWhereX:
		return !effect.PowerDelta.Known &&
			!effect.ToughnessDelta.Known &&
			ability.Text == fmt.Sprintf("%s gets +X/+X until end of turn, %s.", subject, amount.Text)
	default:
		return false
	}
}

func dynamicPTMultiplierMatches(
	multiplier int,
	power, toughness oracle.CompiledSignedAmount,
) bool {
	matches := func(amount oracle.CompiledSignedAmount) bool {
		return amount.Value == 0 || amount.Value == multiplier
	}
	return multiplier > 0 && matches(power) && matches(toughness)
}

func dynamicSignedQuantity(
	dynamic game.DynamicAmount,
	amount oracle.CompiledSignedAmount,
) game.Quantity {
	if amount.Value == 0 {
		return game.Fixed(0)
	}
	if amount.Negative {
		dynamic.Multiplier = -dynamic.Multiplier
	}
	return game.Dynamic(dynamic)
}

func exactXControllerDrawSyntax(tokens []oracle.Token) bool {
	return len(tokens) == 4 &&
		equalTokenWord(tokens[0], "draw") &&
		equalTokenWord(tokens[1], "X") &&
		equalTokenWord(tokens[2], "cards") &&
		tokens[3].Kind == oracle.Period
}

func exactXTargetPlayerDrawSyntax(tokens []oracle.Token) bool {
	return len(tokens) == 6 &&
		equalTokenWord(tokens[0], "target") &&
		equalTokenWord(tokens[1], "player") &&
		equalTokenWord(tokens[2], "draws") &&
		equalTokenWord(tokens[3], "X") &&
		equalTokenWord(tokens[4], "cards") &&
		tokens[5].Kind == oracle.Period
}

func exactControllerDrawSyntax(tokens []oracle.Token, amount int) bool {
	if len(tokens) != 4 ||
		tokens[0].Kind != oracle.Word ||
		!strings.EqualFold(tokens[0].Text, "draw") ||
		tokens[2].Kind != oracle.Word ||
		tokens[3].Kind != oracle.Period {
		return false
	}
	if amount == 1 &&
		strings.EqualFold(tokens[1].Text, "a") &&
		strings.EqualFold(tokens[2].Text, "card") {
		return true
	}
	return fixedNumberToken(tokens[1], amount) &&
		strings.EqualFold(tokens[2].Text, "cards")
}

func exactTargetPlayerDrawSyntax(tokens []oracle.Token, amount int) bool {
	return len(tokens) == 6 &&
		tokens[0].Kind == oracle.Word &&
		strings.EqualFold(tokens[0].Text, "target") &&
		tokens[1].Kind == oracle.Word &&
		strings.EqualFold(tokens[1].Text, "player") &&
		tokens[2].Kind == oracle.Word &&
		strings.EqualFold(tokens[2].Text, "draws") &&
		fixedNumberToken(tokens[3], amount) &&
		tokens[4].Kind == oracle.Word &&
		strings.EqualFold(tokens[4].Text, "cards") &&
		tokens[5].Kind == oracle.Period
}

func fixedCardCountSyntax(amountToken, cardToken oracle.Token, amount int) bool {
	if amount == 1 &&
		strings.EqualFold(amountToken.Text, "a") &&
		strings.EqualFold(cardToken.Text, "card") {
		return true
	}
	return fixedNumberToken(amountToken, amount) &&
		strings.EqualFold(cardToken.Text, "cards")
}

func exactCardCountPlayerSyntax(tokens []oracle.Token, verb string, amount oracle.CompiledAmount) bool {
	if len(tokens) != 4 ||
		!equalTokenWord(tokens[0], verb) ||
		tokens[3].Kind != oracle.Period {
		return false
	}
	return cardCountAmountSyntax(tokens[1], tokens[2], amount)
}

func exactTargetCardCountPlayerSyntax(tokens []oracle.Token, verb string, amount oracle.CompiledAmount) bool {
	if len(tokens) != 6 ||
		!equalTokenWord(tokens[0], "target") ||
		!equalTokenWord(tokens[1], "player") ||
		!equalTokenWord(tokens[2], verb) ||
		tokens[5].Kind != oracle.Period {
		return false
	}
	return cardCountAmountSyntax(tokens[3], tokens[4], amount)
}

func cardCountAmountSyntax(amountToken, cardToken oracle.Token, amount oracle.CompiledAmount) bool {
	if amount.Known {
		return fixedCardCountSyntax(amountToken, cardToken, amount.Value)
	}
	return equalTokenWord(amountToken, "X") &&
		strings.EqualFold(cardToken.Text, "cards")
}

func exactDynamicCardCountPlayerText(text, subject, verb string, amount oracle.CompiledAmount) bool {
	if amount.DynamicKind == oracle.DynamicAmountNone {
		return false
	}
	prefix := titleFirst(verb)
	if subject != "" {
		prefix = subject + " " + verb
	}
	switch amount.DynamicForm {
	case oracle.DynamicAmountForEach:
		if amount.Multiplier == 1 {
			return text == fmt.Sprintf("%s 1 card %s.", prefix, amount.Text) ||
				text == fmt.Sprintf("%s a card %s.", prefix, amount.Text)
		}
		return text == fmt.Sprintf("%s %d cards %s.", prefix, amount.Multiplier, amount.Text)
	case oracle.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s X cards, %s.", prefix, amount.Text)
	default:
		return false
	}
}

func exactControllerAmountSyntax(tokens []oracle.Token, text, verb string, amount oracle.CompiledAmount) bool {
	if amount.Known {
		return len(tokens) == 3 &&
			equalTokenWord(tokens[0], verb) &&
			fixedNumberToken(tokens[1], amount.Value) &&
			tokens[2].Kind == oracle.Period
	}
	if amount.DynamicKind == oracle.DynamicAmountNone {
		return len(tokens) == 3 &&
			equalTokenWord(tokens[0], verb) &&
			equalTokenWord(tokens[1], "X") &&
			tokens[2].Kind == oracle.Period
	}
	switch amount.DynamicForm {
	case oracle.DynamicAmountForEach:
		return text == fmt.Sprintf("%s %d %s.", titleFirst(verb), amount.Multiplier, amount.Text)
	case oracle.DynamicAmountWhereX:
		return text == fmt.Sprintf("%s X, %s.", titleFirst(verb), amount.Text)
	default:
		return false
	}
}

func fixedNumberToken(token oracle.Token, amount int) bool {
	switch strings.ToLower(token.Text) {
	case "one":
		return amount == 1
	case "two":
		return amount == 2
	case "three":
		return amount == 3
	case "four":
		return amount == 4
	default:
		return token.Kind == oracle.Integer && token.Text == fmt.Sprint(amount)
	}
}

func singleSelfReference(references []oracle.CompiledReference) bool {
	return len(references) == 1 && references[0].Kind == oracle.ReferenceSelfName
}

func damageTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
	}
	switch target.Selector.Kind {
	case oracle.SelectorAny:
		if target.Text != "any target" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent | game.TargetAllowPlayer
	case oracle.SelectorCreature, oracle.SelectorPlaneswalker, oracle.SelectorBattle:
		permanent, ok := permanentTargetSpec(target)
		if !ok {
			return game.TargetSpec{}, false
		}
		return permanent, true
	case oracle.SelectorPlayer:
		if target.Text != "target player" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPlayer
	case oracle.SelectorOpponent:
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

func permanentTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
	}
	var noun string
	switch target.Selector.Kind {
	case oracle.SelectorArtifact:
		noun = "artifact"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Artifact}}
	case oracle.SelectorCreature:
		noun = "creature"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	case oracle.SelectorEnchantment:
		noun = "enchantment"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Enchantment}}
	case oracle.SelectorLand:
		noun = "land"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Land}}
	case oracle.SelectorPermanent:
		noun = "permanent"
	case oracle.SelectorPlaneswalker:
		noun = "planeswalker"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Planeswalker}}
	case oracle.SelectorBattle:
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
	case oracle.ControllerAny:
	case oracle.ControllerYou:
		expected += " you control"
		spec.Predicate.Controller = game.ControllerYou
	case oracle.ControllerOpponent:
		expected += " an opponent controls"
		spec.Predicate.Controller = game.ControllerOpponent
	case oracle.ControllerNotYou:
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

func playerTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPlayer,
	}
	switch target.Selector.Kind {
	case oracle.SelectorPlayer:
		if !strings.EqualFold(target.Text, "target player") {
			return game.TargetSpec{}, false
		}
	case oracle.SelectorOpponent:
		if !strings.EqualFold(target.Text, "target opponent") {
			return game.TargetSpec{}, false
		}
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func signedAmountText(amount oracle.CompiledSignedAmount) string {
	if amount.Negative {
		return fmt.Sprintf("-%d", amount.Value)
	}
	return fmt.Sprintf("+%d", amount.Value)
}

func compiledSignedAmountValue(amount oracle.CompiledSignedAmount) int {
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

func exactAnyColorManaSyntax(tokens []oracle.Token, costSymbol string) bool {
	return len(tokens) == 9 &&
		tokens[0].Kind == oracle.Symbol &&
		strings.EqualFold(tokens[0].Text, costSymbol) &&
		tokens[1].Kind == oracle.Colon &&
		equalTokenWord(tokens[2], "add") &&
		equalTokenWord(tokens[3], "one") &&
		equalTokenWord(tokens[4], "mana") &&
		equalTokenWord(tokens[5], "of") &&
		equalTokenWord(tokens[6], "any") &&
		equalTokenWord(tokens[7], "color") &&
		tokens[8].Kind == oracle.Period
}

func equalTokenWord(token oracle.Token, word string) bool {
	return token.Kind == oracle.Word && strings.EqualFold(token.Text, word)
}

func exactChoiceManaSyntax(tokens []oracle.Token, costSymbol string) ([]string, bool) {
	if len(tokens) < 7 ||
		tokens[0].Kind != oracle.Symbol ||
		!strings.EqualFold(tokens[0].Text, costSymbol) ||
		tokens[1].Kind != oracle.Colon ||
		!equalTokenWord(tokens[2], "add") ||
		tokens[len(tokens)-1].Kind != oracle.Period {
		return nil, false
	}
	var colors []string
	for i := 3; i < len(tokens)-1; {
		token := tokens[i]
		manaColor, ok := manaColorName(token.Text)
		if token.Kind != oracle.Symbol || !ok {
			return nil, false
		}
		colors = append(colors, manaColor)
		i++
		if i == len(tokens)-1 {
			break
		}
		if tokens[i].Kind == oracle.Comma {
			i++
			if i < len(tokens)-1 && equalTokenWord(tokens[i], "or") {
				i++
			}
			continue
		}
		if !equalTokenWord(tokens[i], "or") {
			return nil, false
		}
		i++
	}
	return colors, len(colors) >= 2
}

func exactManaSyntax(tokens []oracle.Token, costSymbol string) bool {
	return len(tokens) == 5 &&
		tokens[0].Kind == oracle.Symbol &&
		strings.EqualFold(tokens[0].Text, costSymbol) &&
		tokens[1].Kind == oracle.Colon &&
		tokens[2].Kind == oracle.Word &&
		strings.EqualFold(tokens[2].Text, "Add") &&
		tokens[3].Kind == oracle.Symbol &&
		tokens[4].Kind == oracle.Period
}

func spanCoveredByKeyword(span oracle.Span, keywords []oracle.CompiledKeyword) bool {
	for _, keyword := range keywords {
		if keyword.Span.Start.Offset <= span.Start.Offset &&
			keyword.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func spanCoveredByAbilityWord(span oracle.Span, abilityWord *oracle.Phrase) bool {
	return abilityWord != nil &&
		abilityWord.Span.Start.Offset <= span.Start.Offset &&
		abilityWord.Span.End.Offset >= span.End.Offset
}

func spanCoveredByDelimited(span oracle.Span, groups []oracle.Delimited) bool {
	for _, group := range groups {
		if group.Span.Start.Offset <= span.Start.Offset &&
			group.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func executableDiagnostic(
	ability oracle.CompiledAbility,
	summary string,
	detail string,
) *oracle.Diagnostic {
	return &oracle.Diagnostic{
		Severity: oracle.SeverityWarning,
		Summary:  summary,
		Detail:   detail,
		Span:     ability.Span,
	}
}

func mixedKeywordDiagnostic(ability oracle.CompiledAbility) *oracle.Diagnostic {
	names := make([]string, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		names = append(names, keyword.Name)
	}
	return executableDiagnostic(
		ability,
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
