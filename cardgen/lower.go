package cardgen

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
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
	return lowered, diagnostics
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
	if len(diagnostics) > 0 {
		return loweredFaceAbilities{}, diagnostics
	}

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
		return abilityLowering{
			staticAbilities: []loweredStaticAbility{{Body: staticBuff}},
			consumed: semanticConsumption{
				effects: 1,
			},
			sourceSpans: []oracle.Span{ability.Effects[0].Span},
		}, nil
	}
	switch ability.Kind {
	case oracle.AbilityStatic:
		bodies, diagnostic := lowerKeywordAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}

		spans := make([]oracle.Span, 0, len(ability.Keywords)+len(syntax.Reminders))
		for _, keyword := range ability.Keywords {
			spans = append(spans, keyword.Span)
		}
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
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			spellAbility: opt.Val(spellAbility),
			consumed: semanticConsumption{
				targets:    len(ability.Targets),
				effects:    len(ability.Effects),
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityTriggered:
		triggeredAbility, diagnostic := lowerEnterTrigger(cardName, ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []oracle.Span{ability.Trigger.Span}
		for _, effect := range ability.Effects {
			spans = append(spans, effect.Span)
		}
		for _, target := range ability.Targets {
			spans = append(spans, target.Span)
		}
		for _, reference := range ability.References {
			spans = append(spans, reference.Span)
		}
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			triggeredAbility: opt.Val(triggeredAbility),
			consumed: semanticConsumption{
				trigger:    true,
				optional:   ability.Optional,
				targets:    len(ability.Targets),
				effects:    len(ability.Effects),
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityChapter:
		return lowerChapterAbility(cardName, ability, syntax)
	case oracle.AbilityReplacement:
		replacementAbility, diagnostic := lowerEntersTappedReplacement(ability)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}

		return abilityLowering{
			replacementAbility: opt.Val(replacementAbility),
			consumed: semanticConsumption{
				effects:    1,
				conditions: len(ability.Conditions),
				references: len(ability.References),
			},
			sourceSpans: []oracle.Span{ability.Effects[0].Span},
		}, nil
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
	if dash < 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires an em dash after the chapter numbers",
		)
	}
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
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed: semanticConsumption{
				cost:    true,
				effects: 1,
			},
			sourceSpans: []oracle.Span{ability.Cost.Span, ability.Effects[0].Span},
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
	for _, effect := range ability.Effects {
		spans = append(spans, effect.Span)
	}
	for _, target := range ability.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.References {
		spans = append(spans, reference.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		activatedAbility: opt.Val(activatedAbility),
		consumed: semanticConsumption{
			cost:       true,
			targets:    len(ability.Targets),
			effects:    len(ability.Effects),
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
		len(ability.Keywords) != 0 ||
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
		len(ability.Cost.Components) > 2 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
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
			if i != len(ability.Cost.Components)-1 || len(additionalCosts) != 0 {
				return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
			}
			additionalCosts = cost.Tap
		default:
			return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
		}
	}

	colon := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.Colon
	})
	if colon < 0 || colon+1 >= len(syntax.Tokens) {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}
	body := ability
	body.Kind = oracle.AbilitySpell
	body.Cost = nil
	body.Span = oracle.Span{
		Start: syntax.Tokens[colon+1].Span.Start,
		End:   syntax.Span.End,
	}
	body.Text = strings.TrimSpace(ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset:])
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	bodySyntax.Tokens = syntax.Tokens[colon+1:]
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.ActivatedAbility{}, unsupportedActivatedAbilityDiagnostic(ability)
	}

	result := game.ActivatedAbility{
		Text:            ability.Text,
		AdditionalCosts: additionalCosts,
		ZoneOfFunction:  zone.Battlefield,
		Content:         content,
	}
	if manaCost != nil {
		result.ManaCost = opt.Val(manaCost)
	}
	return result, nil
}

func unsupportedActivatedAbilityDiagnostic(ability oracle.CompiledAbility) *oracle.Diagnostic {
	return executableDiagnostic(
		ability,
		"unsupported activated ability",
		"the executable source backend supports only exact mana and tap costs with a supported effect",
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

// lowerKeywordDispatch tries Enchant, Protection, Equip, and Cycling — the
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

func lowerStaticPTBuff(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.StaticAbility, bool, *oracle.Diagnostic) {
	if ability.Kind != oracle.AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectModifyPT ||
		ability.Effects[0].Duration != oracle.DurationNone ||
		!ability.Effects[0].PowerDelta.Known ||
		!ability.Effects[0].ToughnessDelta.Known ||
		len(ability.Keywords) != 0 ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].StaticSubject == oracle.StaticSubjectNone ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, false, nil
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []oracle.Span{ability.Effects[0].Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend supports only exact fixed static creature power/toughness buffs",
		)
	}
	effect := ability.Effects[0]
	if !matchesExactStaticPTBuffSyntax(syntax, effect) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend supports only exact fixed static creature power/toughness buffs",
		)
	}
	group, ok := staticSubjectGroup(effect.StaticSubject)
	if !ok {
		return game.StaticAbility{}, false, nil
	}
	return game.StaticAbility{
		Text: ability.Text,
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:          game.LayerPowerToughnessModify,
			Group:          group,
			PowerDelta:     compiledSignedAmountValue(effect.PowerDelta),
			ToughnessDelta: compiledSignedAmountValue(effect.ToughnessDelta),
		}},
	}, true, nil
}

func staticSubjectGroup(subject oracle.StaticSubjectKind) (game.GroupReference, bool) {
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
	default:
		return game.GroupReference{}, false
	}
}

func matchesExactStaticPTBuffSyntax(
	syntax oracle.Ability,
	effect oracle.CompiledEffect,
) bool {
	tokens := syntaxSemanticTokens(syntax)
	switch effect.StaticSubject {
	case oracle.StaticSubjectAttachedObject:
		return len(tokens) == 9 &&
			(equalTokenWord(tokens[0], "enchanted") || equalTokenWord(tokens[0], "equipped")) &&
			equalTokenWord(tokens[1], "creature") &&
			equalTokenWord(tokens[2], "gets") &&
			tokensMatchSignedAmount(tokens[3], tokens[4], effect.PowerDelta) &&
			tokens[5].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[6], tokens[7], effect.ToughnessDelta) &&
			tokens[8].Kind == oracle.Period
	case oracle.StaticSubjectControlledCreatures:
		return len(tokens) == 10 &&
			equalTokenWord(tokens[0], "creatures") &&
			equalTokenWord(tokens[1], "you") &&
			equalTokenWord(tokens[2], "control") &&
			equalTokenWord(tokens[3], "get") &&
			tokensMatchSignedAmount(tokens[4], tokens[5], effect.PowerDelta) &&
			tokens[6].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[7], tokens[8], effect.ToughnessDelta) &&
			tokens[9].Kind == oracle.Period
	case oracle.StaticSubjectOtherControlledCreatures:
		return len(tokens) == 11 &&
			equalTokenWord(tokens[0], "other") &&
			equalTokenWord(tokens[1], "creatures") &&
			equalTokenWord(tokens[2], "you") &&
			equalTokenWord(tokens[3], "control") &&
			equalTokenWord(tokens[4], "get") &&
			tokensMatchSignedAmount(tokens[5], tokens[6], effect.PowerDelta) &&
			tokens[7].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[8], tokens[9], effect.ToughnessDelta) &&
			tokens[10].Kind == oracle.Period
	case oracle.StaticSubjectControlledWalls:
		offset := 0
		noun := "walls"
		verb := "get"
		if len(tokens) == 11 && equalTokenWord(tokens[0], "each") {
			offset = 1
			noun = "wall"
			verb = "gets"
		}
		return len(tokens) == 10+offset &&
			equalTokenWord(tokens[offset], noun) &&
			equalTokenWord(tokens[offset+1], "you") &&
			equalTokenWord(tokens[offset+2], "control") &&
			equalTokenWord(tokens[offset+3], verb) &&
			tokensMatchSignedAmount(tokens[offset+4], tokens[offset+5], effect.PowerDelta) &&
			tokens[offset+6].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[offset+7], tokens[offset+8], effect.ToughnessDelta) &&
			tokens[offset+9].Kind == oracle.Period
	case oracle.StaticSubjectControlledArtifacts, oracle.StaticSubjectControlledTokens:
		noun := "artifacts"
		if effect.StaticSubject == oracle.StaticSubjectControlledTokens {
			noun = "tokens"
		}
		return len(tokens) == 10 &&
			equalTokenWord(tokens[0], noun) &&
			equalTokenWord(tokens[1], "you") &&
			equalTokenWord(tokens[2], "control") &&
			equalTokenWord(tokens[3], "get") &&
			tokensMatchSignedAmount(tokens[4], tokens[5], effect.PowerDelta) &&
			tokens[6].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[7], tokens[8], effect.ToughnessDelta) &&
			tokens[9].Kind == oracle.Period
	case oracle.StaticSubjectOpponentControlledCreatures:
		return len(tokens) == 11 &&
			equalTokenWord(tokens[0], "creatures") &&
			equalTokenWord(tokens[1], "your") &&
			equalTokenWord(tokens[2], "opponents") &&
			equalTokenWord(tokens[3], "control") &&
			equalTokenWord(tokens[4], "get") &&
			tokensMatchSignedAmount(tokens[5], tokens[6], effect.PowerDelta) &&
			tokens[7].Kind == oracle.Slash &&
			tokensMatchSignedAmount(tokens[8], tokens[9], effect.ToughnessDelta) &&
			tokens[10].Kind == oracle.Period
	default:
		return false
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

// lowerReminderManaAbility handles a parenthesized reminder mana ability such
// as "({T}: Add {R} or {G}.)" — reminder text that describes the tap-for-mana
// behavior granted by basic-land subtypes. These are compiled with no semantic
// elements because their content is filtered as parenthesized. We re-compile
// the inner text and lower it as a mana ability.
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
	if len(innerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		innerComp.Abilities[0].Kind != oracle.AbilityActivated {
		return abilityLowering{}, unsupported()
	}
	innerAbility := innerComp.Abilities[0]
	innerSyntax := innerComp.Syntax.Abilities[0]
	if !hasAddManaEffect(innerAbility) {
		return abilityLowering{}, unsupported()
	}
	manaAbility, diagnostic := lowerTapManaAbility(innerAbility, innerSyntax)
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

func lowerEntersTappedReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, *oracle.Diagnostic) {
	if len(ability.Effects) != 1 ||
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

// lowerConditionalEntersTappedReplacement handles the exact wording family
// "This land enters tapped unless you control two or more basic lands.".
func lowerConditionalEntersTappedReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, *oracle.Diagnostic) {
	condition := ability.Conditions[0]
	const supportedConditionText = "unless you control two or more basic lands"
	const supportedAbilityText = "This land enters tapped unless you control two or more basic lands."
	if condition.Kind != oracle.ConditionUnless ||
		condition.Text != supportedConditionText ||
		ability.Text != supportedAbilityText {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported conditional enters-tapped replacement",
			"the executable source backend supports only the exact condition: unless you control two or more basic lands",
		)
	}
	return game.EntersTappedIfReplacement(ability.Text, &game.Condition{
		Negate: true,
		ControllerControls: game.PermanentFilter{
			Types:      []types.Card{types.Land},
			Supertypes: []types.Super{types.Basic},
			MinCount:   2,
		},
	}), nil
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
	eventKind, supportedEvent := lowerSelfTriggerEvent(ability)
	summary := "unsupported enter trigger"
	detail := "the executable source backend supports only exact self-enter triggers with supported effects"
	if ability.Trigger != nil && strings.HasSuffix(ability.Trigger.Event, " dies") {
		summary = "unsupported dies trigger"
		detail = "the executable source backend supports only exact self-dies triggers with supported effects"
	}
	if ability.Trigger == nil ||
		ability.Trigger.Kind != oracle.TriggerWhen ||
		!supportedEvent ||
		ability.Trigger.Condition != nil ||
		len(ability.Effects) == 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			detail,
		)
	}
	comma := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.Comma
	})
	if comma < 0 || comma+1 >= len(syntax.Tokens) {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			detail,
		)
	}
	body := ability
	body.Kind = oracle.AbilitySpell
	body.Span = oracle.Span{
		Start: ability.Effects[0].Span.Start,
		End:   ability.Effects[len(ability.Effects)-1].Span.End,
	}
	body.Text = titleFirst(
		ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
	)
	body.Trigger = nil
	body.Optional = false
	body.OptionalSpan = oracle.Span{}
	body.References = bodyReferences(ability.References, ability.Trigger.Span)
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Tokens = syntax.Tokens[comma+1:]
	if ability.Optional {
		if len(ability.Effects) != 1 ||
			len(bodySyntax.Tokens) < 3 ||
			!equalTokenWord(bodySyntax.Tokens[0], "you") ||
			!equalTokenWord(bodySyntax.Tokens[1], "may") ||
			ability.OptionalSpan.Start != ability.Effects[0].Span.Start {
			return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
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
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
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
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  eventKind,
				Source: game.TriggerSourceSelf,
			},
		},
		Optional: ability.Optional,
		Content:  content,
	}, nil
}

func lowerSelfTriggerEvent(ability oracle.CompiledAbility) (game.EventKind, bool) {
	if ability.Trigger == nil {
		return 0, false
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
		return game.EventPermanentEnteredBattlefield, true
	case "this creature dies", "this permanent dies":
		return game.EventPermanentDied, true
	default:
		return 0, false
	}
}

func bodyReferences(
	references []oracle.CompiledReference,
	triggerSpan oracle.Span,
) []oracle.CompiledReference {
	var body []oracle.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, []oracle.Span{triggerSpan}) {
			continue
		}
		body = append(body, reference)
	}
	return body
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
	}
	if len(ability.Modes) > 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not yet lower modal abilities",
		)
	}
	if ability.AbilityWord != "" {
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
			spanCoveredByKeyword(token.Span, ability.Keywords) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return nil, mixedKeywordDiagnostic(ability)
	}
	return bodies, nil
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
		ability.Cost.Components[0].Kind != oracle.CostTap ||
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
			"the executable source backend supports only exact supported tap mana abilities",
		)
	}
	if exactAnyColorTapManaSyntax(syntax.Tokens) {
		return choiceTapManaAbility(
			[]string{"W", "U", "B", "R", "G"},
		), nil
	}
	if colors, ok := exactChoiceTapManaSyntax(syntax.Tokens); ok {
		return choiceTapManaAbility(colors), nil
	}
	if !exactTapManaSyntax(syntax.Tokens) {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activated ability",
			"the executable source backend supports only exact supported tap mana abilities",
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
	return game.TapManaAbility(manaColor), nil
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
		return lowerFixedControllerSpell(ability, syntax, "scry", func(amount int, player game.PlayerReference) game.Primitive {
			return game.Scry{Amount: game.Fixed(amount), Player: player}
		})
	case oracle.EffectSurveil:
		return lowerFixedControllerSpell(ability, syntax, "surveil", func(amount int, player game.PlayerReference) game.Primitive {
			return game.Surveil{Amount: game.Fixed(amount), Player: player}
		})
	case oracle.EffectInvestigate:
		return lowerInvestigateSpell(ability, syntax)
	case oracle.EffectProliferate:
		return lowerExactPrimitiveSpell(ability, syntax, "proliferate", game.Proliferate{})
	case oracle.EffectRegenerate:
		return lowerFixedPermanentTargetSpell(ability, "Regenerate", func(object game.ObjectReference) game.Primitive {
			return game.Regenerate{Object: object}
		})
	case oracle.EffectFight:
		return lowerFightSpell(ability)
	case oracle.EffectDiscard:
		return lowerFixedCardCountPlayerSpell(
			ability, syntax, "discard", "discards", func(amount int, player game.PlayerReference) game.Primitive {
				return game.Discard{Amount: game.Fixed(amount), Player: player}
			},
		)
	case oracle.EffectMill:
		return lowerFixedCardCountPlayerSpell(
			ability, syntax, "mill", "mills", func(amount int, player game.PlayerReference) game.Primitive {
				return game.Mill{Amount: game.Fixed(amount), Player: player}
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
		return lowerFixedBounceSpell(ability)
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
		game.Investigate{Amount: game.Fixed(1)},
	)
}

func lowerExactPrimitiveSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	verb string,
	primitive game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	if ability.Effects[0].Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		len(syntax.Tokens) != 2 ||
		!equalTokenWord(syntax.Tokens[0], verb) ||
		syntax.Tokens[1].Kind != oracle.Period {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact "+verb,
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: primitive,
	}}}.Ability(), nil
}

func lowerOrderedEffectSequence(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Conditions) != 0 || len(ability.Keywords) != 0 || len(ability.Modes) != 0 {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
	}
	var targets []game.TargetSpec
	var sequence []game.Instruction
	consumedTargets := 0
	consumedReferences := 0
	for _, effect := range ability.Effects {
		effectAbility := abilityForEffect(ability, effect)
		consumedTargets += len(effectAbility.Targets)
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
			if len(targets) > 0 {
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
			}
			targets = mode.Targets
		}
		sequence = append(sequence, mode.Sequence...)
	}
	if consumedTargets != len(ability.Targets) ||
		consumedReferences != len(ability.References) ||
		len(sequence) != len(ability.Effects) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ability)
	}
	return game.Mode{Targets: targets, Sequence: sequence}.Ability(), nil
}

func abilityForEffect(
	ability oracle.CompiledAbility,
	effect oracle.CompiledEffect,
) oracle.CompiledAbility {
	ability.Text = effect.Text
	ability.Span = effect.Span
	ability.Effects = []oracle.CompiledEffect{effect}
	ability.Targets = targetsWithinSpan(ability.Targets, effect.Span)
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
		"the executable source backend supports only exact ordered sequences of independently supported effects with at most one targeted clause",
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
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		!singleSelfReference(ability.References) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed damage to one target",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	amountText := "X"
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
		amountText = fmt.Sprint(effect.Amount.Value)
	}
	target, ok := damageTargetSpec(ability.Targets[0])
	if !ok ||
		ability.Text != fmt.Sprintf(
			"%s deals %s damage to %s.",
			cardName,
			amountText,
			ability.Targets[0].Text,
		) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed damage to one target",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{
				Primitive: game.Damage{
					Amount:    amount,
					Recipient: game.AnyTargetDamageRecipient(0),
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedModifyPTSpell(
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Targets[0].Selector.Kind != oracle.SelectorCreature ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		effect.Negated ||
		effect.Duration != oracle.DurationUntilEndOfTurn ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Text != fmt.Sprintf(
			"%s gets %s/%s until end of turn.",
			titleFirst(ability.Targets[0].Text),
			signedAmountText(effect.PowerDelta),
			signedAmountText(effect.ToughnessDelta),
		) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed target-creature power/toughness changes until end of turn",
		)
	}
	targetSpec, ok := permanentTargetSpec(ability.Targets[0])
	if !ok {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed target-creature power/toughness changes until end of turn",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ModifyPT{
					Object:         game.TargetPermanentReference(0),
					PowerDelta:     game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
					ToughnessDelta: game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)),
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
	primitiveFactory func(amount int, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
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
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		len(syntax.Tokens) == 4 &&
		strings.EqualFold(syntax.Tokens[0].Text, controllerVerb) &&
		fixedCardCountSyntax(syntax.Tokens[1], syntax.Tokens[2], effect.Amount.Value) &&
		syntax.Tokens[3].Kind == oracle.Period:
	case len(ability.Targets) == 1 &&
		len(syntax.Tokens) == 6 &&
		strings.EqualFold(syntax.Tokens[0].Text, "target") &&
		strings.EqualFold(syntax.Tokens[1].Text, "player") &&
		strings.EqualFold(syntax.Tokens[2].Text, targetVerb) &&
		fixedCardCountSyntax(syntax.Tokens[3], syntax.Tokens[4], effect.Amount.Value) &&
		syntax.Tokens[5].Kind == oracle.Period:
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
				Primitive: primitiveFactory(effect.Amount.Value, playerRef),
			},
		},
	}.Ability(), nil
}

func lowerFixedControllerSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	verb string,
	primitiveFactory func(amount int, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		len(syntax.Tokens) != 3 ||
		!strings.EqualFold(syntax.Tokens[0].Text, verb) ||
		!fixedNumberToken(syntax.Tokens[1], effect.Amount.Value) ||
		syntax.Tokens[2].Kind != oracle.Period {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(effect.Amount.Value, game.ControllerReference()),
			},
		},
	}.Ability(), nil
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
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	amountText := "X"
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
		amountText = fmt.Sprint(effect.Amount.Value)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		ability.Text == fmt.Sprintf("You %s %s life.", verb, amountText):
	case len(ability.Targets) == 1:
		targetSpec, ok := playerTargetSpec(ability.Targets[0])
		if !ok ||
			ability.Text != fmt.Sprintf(
				"%s %ss %s life.",
				titleFirst(ability.Targets[0].Text),
				verb,
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
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		(exactControllerDrawSyntax(syntax.Tokens, effect.Amount.Value) ||
			(!effect.Amount.Known && exactXControllerDrawSyntax(syntax.Tokens))):
	case len(ability.Targets) == 1 &&
		(exactTargetPlayerDrawSyntax(syntax.Tokens, effect.Amount.Value) ||
			(!effect.Amount.Known && exactXTargetPlayerDrawSyntax(syntax.Tokens))) &&
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

func exactAnyColorTapManaSyntax(tokens []oracle.Token) bool {
	return len(tokens) == 9 &&
		tokens[0].Kind == oracle.Symbol &&
		strings.EqualFold(tokens[0].Text, "{T}") &&
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

func exactChoiceTapManaSyntax(tokens []oracle.Token) ([]string, bool) {
	if len(tokens) < 7 ||
		tokens[0].Kind != oracle.Symbol ||
		!strings.EqualFold(tokens[0].Text, "{T}") ||
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

func exactTapManaSyntax(tokens []oracle.Token) bool {
	return len(tokens) == 5 &&
		tokens[0].Kind == oracle.Symbol &&
		strings.EqualFold(tokens[0].Text, "{T}") &&
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
