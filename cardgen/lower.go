package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
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
		syntax := &compilation.Syntax.Abilities[i]
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
	for i, ability := range compilation.Abilities {
		syntax := &compilation.Syntax.Abilities[i]
		for _, keyword := range ability.Content.Keywords {
			if keyword.Kind != parser.KeywordReadAhead {
				continue
			}
			if !syntax.ReadAheadRecognized || syntax.ReadAheadSacrificeChapter == 0 {
				continue
			}
			sacrificeChapter := syntax.ReadAheadSacrificeChapter
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
	syntax *parser.Ability,
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
		if len(body.Content.Effects) == 1 &&
			body.Content.Effects[0].Kind == compiler.EffectAddMana &&
			body.Content.Effects[0].Mana.AnyColor {
			return abilityLowering{}, executableDiagnostic(
				ability,
				"unsupported mana symbol",
				"the executable source backend cannot lower this add-mana content",
			)
		}
		spellAbility, diagnostic := lowerAbilityContent(cardName, body.Content, body.Optional, &bodySyntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := make(
			[]shared.Span,
			0,
			len(ability.Content.Effects)+len(ability.Content.Targets)+len(ability.Content.Conditions)+len(ability.Content.References)+len(syntax.Reminders),
		)
		for i := range ability.Content.Effects {
			spans = append(spans, ability.Content.Effects[i].Span)
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
		if saga && syntax.SagaReminder {
			return abilityLowering{sourceSpans: []shared.Span{ability.Span}}, nil
		}

		return lowerReminderManaAbility(ability, syntax)
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
	syntax *parser.Ability,
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
