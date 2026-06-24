package cardgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
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
	Overload             opt.V[game.OverloadAbility]
	AdditionalCosts      []cost.Additional
	AlternativeCosts     []cost.Alternative
	EntersPrepared       bool
	DynamicPower         opt.V[game.DynamicValue]
	DynamicToughness     opt.V[game.DynamicValue]
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
		!f.Overload.Exists &&
		len(f.AdditionalCosts) == 0 &&
		len(f.AlternativeCosts) == 0 &&
		!f.DynamicPower.Exists &&
		!f.DynamicToughness.Exists &&
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
	overloadCost       opt.V[cost.Mana]
	additionalCosts    []cost.Additional
	alternativeCosts   []cost.Alternative
	entersPrepared     bool
	dynamicPower       opt.V[game.DynamicValue]
	dynamicToughness   opt.V[game.DynamicValue]
	consumed           semanticConsumption
	sourceSpans        []shared.Span
}

type semanticConsumption struct {
	cost            bool
	alternativeCost bool
	trigger         bool
	optional        bool
	modes           int
	targets         int
	conditions      int
	effects         int
	keywords        int
	references      int
	declarations    int
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
					if instruction.Primitive == nil ||
						instruction.Primitive.Kind() != game.PrimitiveGrantCastPermission {
						continue
					}
					// Only an alternate-face cast permission is Adventure-specific
					// and requires the Adventure layout. A front-face graveyard
					// cast permission (Norika Yamazaki, the Poet) casts the card
					// normally and is valid on any layout.
					permission, ok := instruction.Primitive.(game.GrantCastPermission)
					if ok && permission.Face == game.FaceAlternate {
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
		Class:            slices.Contains(parsedType.Subtypes, "Class"),
		Leveler:          face.Layout == "leveler",
		CardName:         face.Name,
	})
	compilation, compilerDiagnostics := compiler.Compile(document, compiler.Context{})
	diagnostics = append(diagnostics, compilerDiagnostics...)

	var result loweredFaceAbilities
	if spell, ok := lowerSpellFaceCombiner(face.Name, compilation); ok {
		result.SpellAbility = opt.Val(spell)
		return result, diagnostics
	}
	var unsupported []shared.Diagnostic
	var pendingPonderPrefix *compiler.CompiledAbility
	creatureSubtypes := eternalizeFamilyCreatureSubtypes(parsedType.Subtypes)
	saga := slices.Contains(parsedType.Subtypes, "Saga")
	isClass := slices.Contains(parsedType.Subtypes, "Class")
	isLeveler := face.Layout == "leveler"
	classLevel := 1
	var currentBand *compiler.CompiledLevelBand
	for i, ability := range compilation.Abilities {
		syntax := &compilation.Syntax.Abilities[i]
		if isLeveler && ability.LevelUpRecognized {
			activated, diagnostic := lowerLevelUpAbility(face.Name, ability)
			if diagnostic != nil {
				unsupported = append(unsupported, *diagnostic)
				continue
			}
			result.ActivatedAbilities = append(result.ActivatedAbilities, activated)
			continue
		}
		if isLeveler && ability.Kind == compiler.AbilityLevelBand {
			currentBand = ability.LevelBand
			static, diagnostic, emit := lowerLevelBandPowerToughness(ability)
			if diagnostic != nil {
				unsupported = append(unsupported, *diagnostic)
				continue
			}
			if emit {
				result.StaticAbilities = append(result.StaticAbilities, static)
			}
			continue
		}
		var lowered abilityLowering
		var diagnostic *shared.Diagnostic
		levelGain := ability.ClassLevelGain
		if levelGain > 0 {
			lowered, diagnostic = lowerClassLevelGain(face.Name, ability, syntax, classLevel)
		} else {
			lowered, diagnostic = lowerExecutableAbility(
				face.Name,
				saga,
				creatureSubtypes,
				ability,
				syntax,
			)
		}
		if diagnostic != nil {
			unsupported = append(unsupported, *diagnostic)
			continue
		}
		if !lowered.complete(ability, syntax) {
			unsupported = append(unsupported, *incompleteLoweringDiagnostic(ability))
			continue
		}
		if isClass && levelGain == 0 && classLevel >= 2 {
			if diagnostic := gateLoweredAbilityByClassLevel(&lowered, ability, classLevel); diagnostic != nil {
				unsupported = append(unsupported, *diagnostic)
				continue
			}
		}
		if isLeveler && currentBand != nil {
			if diagnostic := gateLoweredAbilityByLevelBand(&lowered, ability, currentBand); diagnostic != nil {
				unsupported = append(unsupported, *diagnostic)
				continue
			}
		}
		if levelGain > 0 {
			classLevel = levelGain
		}
		result.StaticAbilities = append(result.StaticAbilities, lowered.staticAbilities...)
		appendSimpleLoweredAbilities(&result, &lowered)
		if diagnostic := mergeDynamicPowerToughness(&result, &lowered, ability); diagnostic != nil {
			unsupported = append(unsupported, *diagnostic)
			continue
		}
		if lowered.spellAbility.Exists {
			if result.SpellAbility.Exists {
				clearPonder, merged := mergeTrailingSpellAbility(&result.SpellAbility.Val, lowered.spellAbility.Val)
				if merged {
					if clearPonder {
						pendingPonderPrefix = nil
					}
					continue
				}
				unsupported = append(unsupported, *executableDiagnostic(
					ability,
					"unsupported multiple spell abilities",
					"the executable source backend supports only one spell ability per card face",
				))
				continue
			}
			result.SpellAbility = lowered.spellAbility
			if isPonderPrefixAbility(lowered.spellAbility.Val) {
				pending := ability
				pendingPonderPrefix = &pending
			}
		}
		if lowered.overloadCost.Exists {
			if result.Overload.Exists {
				unsupported = append(unsupported, *executableDiagnostic(
					ability,
					"unsupported multiple overload costs",
					"the executable source backend supports only one overload cost per card face",
				))
				continue
			}
			result.Overload = opt.Val(game.OverloadAbility{Cost: lowered.overloadCost.Val})
		}
	}
	unsupported = appendPendingPonderDiagnostic(unsupported, pendingPonderPrefix)
	if diagnostic := finalizeOverload(&result, compilation); diagnostic != nil {
		unsupported = append(unsupported, *diagnostic)
	}
	if len(unsupported) == 0 &&
		faceHasVariableXGroupEffect(result) &&
		!faceHasVariableLifeCost(result) {
		for _, ability := range compilation.Abilities {
			if ability.Kind != compiler.AbilitySpell {
				continue
			}
			unsupported = append(unsupported, *executableDiagnostic(
				ability,
				"unsupported linked X group effect",
				"the executable source backend requires an exact pay-X-life additional cost for a resolving group -X/-X effect",
			))
			break
		}
	}
	if len(unsupported) == 0 && faceHasVariableLifeCost(result) && faceManaCostHasX(face.ManaCost) {
		for _, ability := range compilation.Abilities {
			if ability.Kind != compiler.AbilitySpellAdditionalCost {
				continue
			}
			unsupported = append(unsupported, *executableDiagnostic(
				ability,
				"unsupported linked X spell cost",
				"the executable source backend does not combine pay-X-life additional costs with variable mana costs",
			))
			break
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
	linkExplicitExileReturns(&result)
	synthesizeExileUntilLeavesReturns(&result)
	if len(unsupported) > 0 {
		return loweredFaceAbilities{}, append(diagnostics, unsupported...)
	}
	return result, diagnostics
}

func appendSimpleLoweredAbilities(result *loweredFaceAbilities, lowered *abilityLowering) {
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
	result.AdditionalCosts = append(result.AdditionalCosts, lowered.additionalCosts...)
	result.AlternativeCosts = append(result.AlternativeCosts, lowered.alternativeCosts...)
}

func mergeDynamicPowerToughness(
	result *loweredFaceAbilities,
	lowered *abilityLowering,
	ability compiler.CompiledAbility,
) *shared.Diagnostic {
	if lowered.dynamicPower.Exists {
		if result.DynamicPower.Exists {
			return executableDiagnostic(
				ability,
				"unsupported multiple characteristic-defining power",
				"the executable source backend supports only one characteristic-defining power per card face",
			)
		}
		result.DynamicPower = lowered.dynamicPower
	}
	if lowered.dynamicToughness.Exists {
		if result.DynamicToughness.Exists {
			return executableDiagnostic(
				ability,
				"unsupported multiple characteristic-defining toughness",
				"the executable source backend supports only one characteristic-defining toughness per card face",
			)
		}
		result.DynamicToughness = lowered.dynamicToughness
	}
	return nil
}

func appendTrailingPonderDraw(content *game.AbilityContent, suffix game.AbilityContent) bool {
	if content == nil ||
		!isPonderPrefixAbility(*content) ||
		!isBareMandatoryControllerDrawOne(suffix) {
		return false
	}
	content.Modes[0].Sequence = append(content.Modes[0].Sequence, suffix.Modes[0].Sequence[0])
	return true
}

func isBareMandatoryControllerDrawOne(content game.AbilityContent) bool {
	if len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		content.MinModes != 1 ||
		content.MaxModes != 1 ||
		content.ModeChoiceBonus.Condition != game.ModeChoiceConditionNone ||
		content.ModeChoiceBonus.AdditionalMaxModes != 0 ||
		content.AllowDuplicateModes {
		return false
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		return false
	}
	instruction := mode.Sequence[0]
	if instruction.Optional ||
		instruction.Condition.Exists ||
		instruction.CardCondition.Exists ||
		instruction.ResultGate.Exists ||
		instruction.OptionalActor.Exists ||
		instruction.PublishResult != "" ||
		instruction.Description != "" {
		return false
	}
	draw, ok := instruction.Primitive.(game.Draw)
	return ok &&
		draw.Player.Kind() == game.PlayerReferenceController &&
		draw.PlayerGroup.Kind == game.PlayerGroupReferenceNone &&
		!draw.Amount.IsDynamic() &&
		draw.Amount.Value() == 1
}

func appendPendingPonderDiagnostic(
	diagnostics []shared.Diagnostic,
	pending *compiler.CompiledAbility,
) []shared.Diagnostic {
	if pending == nil {
		return diagnostics
	}
	return append(diagnostics, *executableDiagnostic(
		*pending,
		"unsupported spell ability",
		"the executable source backend requires the Ponder ordering and optional shuffle paragraph to be followed by an exact draw-one spell paragraph",
	))
}

func isPonderPrefixAbility(content game.AbilityContent) bool {
	if len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) != 2 {
		return false
	}
	reorder, reorderOK := content.Modes[0].Sequence[0].Primitive.(game.ReorderLibraryTop)
	shuffle, shuffleOK := content.Modes[0].Sequence[1].Primitive.(game.ShuffleLibrary)
	return reorderOK &&
		shuffleOK &&
		reorder.Player.Kind() == game.PlayerReferenceController &&
		!reorder.Amount.IsDynamic() &&
		reorder.Amount.Value() > 0 &&
		shuffle.Player.Kind() == game.PlayerReferenceController &&
		content.Modes[0].Sequence[1].Optional
}

func appendTrailingSourceSpellExile(content *game.AbilityContent, suffix game.AbilityContent) bool {
	if content == nil ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(suffix.SharedTargets) != 0 ||
		len(suffix.Modes) != 1 ||
		len(suffix.Modes[0].Targets) != 0 ||
		len(suffix.Modes[0].Sequence) != 1 {
		return false
	}
	exile, ok := suffix.Modes[0].Sequence[0].Primitive.(game.Exile)
	if !ok || !exile.SourceSpell {
		return false
	}
	content.Modes[0].Sequence = append(content.Modes[0].Sequence, suffix.Modes[0].Sequence[0])
	return true
}

// mergeTrailingSpellAbility folds a second lowered spell-ability paragraph into
// the running spell ability, trying each supported merge shape in turn. It
// reports whether the paragraph was absorbed and whether a pending Ponder-prefix
// expectation should be cleared (the trailing draw it was waiting for has now
// been consumed by a merge).
func mergeTrailingSpellAbility(content *game.AbilityContent, suffix game.AbilityContent) (clearPonder, merged bool) {
	if appendTrailingPonderDraw(content, suffix) {
		return true, true
	}
	if appendTrailingSourceSpellExile(content, suffix) {
		return false, true
	}
	if appendSequentialSpellParagraph(content, suffix) {
		return true, true
	}
	return false, false
}

// appendSequentialSpellParagraph merges a trailing spell-ability paragraph into
// the running spell ability so that a face with several "do X. do Y." paragraphs
// (each already an individually supported spell effect) lowers to one resolving
// instruction sequence rather than failing closed on more than one spell
// ability. It is deliberately conservative: both paragraphs must be ordinary
// non-modal content with no shared targets, and the trailing paragraph must take
// no targets of its own (so existing target indices stay valid without
// reindexing) and must not publish or gate on a result. The accumulated content
// may itself publish and gate on a result internally (e.g. a "counter target
// spell unless its controller pays" paragraph), because appending a plain,
// key-free suffix after it runs unconditionally and cannot rebind or collide
// with those paragraph-local keys.
func appendSequentialSpellParagraph(content *game.AbilityContent, suffix game.AbilityContent) bool {
	if content == nil ||
		content.IsModal() ||
		suffix.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(suffix.SharedTargets) != 0 ||
		len(suffix.Modes[0].Targets) != 0 ||
		!plainResolvingSequence(suffix.Modes[0].Sequence) {
		return false
	}
	content.Modes[0].Sequence = append(content.Modes[0].Sequence, suffix.Modes[0].Sequence...)
	return true
}

// plainResolvingSequence reports whether every instruction is safe to resequence
// across paragraph boundaries — none publishes a result key or gates on one, so
// merging cannot rebind a paragraph-local "if you do" reference.
func plainResolvingSequence(sequence []game.Instruction) bool {
	for i := range sequence {
		if sequence[i].PublishResult != "" || sequence[i].ResultGate.Exists {
			return false
		}
	}
	return true
}

func finalizeOverload(
	result *loweredFaceAbilities,
	compilation compiler.Compilation,
) *shared.Diagnostic {
	if !result.Overload.Exists {
		return nil
	}
	var spell *compiler.CompiledAbility
	for i := range compilation.Abilities {
		if compilation.Abilities[i].Kind != compiler.AbilitySpell {
			continue
		}
		if spell != nil {
			spell = nil
			break
		}
		spell = &compilation.Abilities[i]
	}
	overloaded, ok := lowerOverloadSpell(result.SpellAbility, spell)
	if ok {
		value := result.Overload.Val
		value.SpellAbility = overloaded
		result.Overload = opt.Val(value)
		return nil
	}
	for _, ability := range compilation.Abilities {
		if ability.Kind == compiler.AbilitySpellAlternativeCost &&
			ability.AlternativeCost != nil &&
			ability.AlternativeCost.Kind == compiler.AlternativeCostOverload {
			return executableDiagnostic(
				ability,
				"unsupported overload effect",
				"overload requires one exact permanent target and a supported target-only effect",
			)
		}
	}
	return nil
}

func faceHasVariableLifeCost(face loweredFaceAbilities) bool {
	for _, additional := range face.AdditionalCosts {
		if additional.Kind == cost.AdditionalPayLife && additional.AmountFromX {
			return true
		}
	}
	return false
}

func faceHasVariableXGroupEffect(face loweredFaceAbilities) bool {
	if !face.SpellAbility.Exists {
		return false
	}
	for _, mode := range face.SpellAbility.Val.Modes {
		for instructionIndex := range mode.Sequence {
			apply, ok := mode.Sequence[instructionIndex].Primitive.(game.ApplyContinuous)
			if !ok {
				continue
			}
			for effectIndex := range apply.ContinuousEffects {
				effect := &apply.ContinuousEffects[effectIndex]
				if effect.Group.Valid() &&
					effect.PowerDeltaDynamic.Exists &&
					effect.PowerDeltaDynamic.Val.Kind == game.DynamicAmountX &&
					effect.ToughnessDeltaDynamic.Exists &&
					effect.ToughnessDeltaDynamic.Val.Kind == game.DynamicAmountX {
					return true
				}
			}
		}
	}
	return false
}

func faceManaCostHasX(manaCost string) bool {
	parsed, err := parseManaCostValue(manaCost)
	if err != nil {
		return false
	}
	for _, symbol := range parsed {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

// incompleteLoweringDiagnostic reports that strict executable lowering left
// typed semantic elements or source spans unconsumed. When the typed effect
// family is one the executable backend names specifically, it restores the
// family-specific diagnostic so the support report records what the backend
// recognized but cannot yet lower, rather than the opaque generic reason. It
// reads only typed compiler content and never inspects Oracle wording.
func incompleteLoweringDiagnostic(ability compiler.CompiledAbility) *shared.Diagnostic {
	summary, detail := unsupportedEffectFamily(ability.Content)
	return executableDiagnostic(ability, summary, detail)
}

// unsupportedEffectFamily names the effect family of an unconsumed ability body
// from typed compiler signals alone. Delayed one-shot effects, add-mana content,
// and multi-effect ordered sequences each map to their established family
// diagnostic; every other shape keeps the generic incomplete-lowering reason
// because the backend cannot attribute it to a known family.
func unsupportedEffectFamily(content compiler.AbilityContent) (summary, detail string) {
	switch {
	case abilityContentHasDelayedEffect(content):
		return "unsupported delayed effect",
			"the executable source backend supports only exact non-target delayed one-shot effects"
	case abilityContentHasAddManaEffect(content):
		return "unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content"
	case abilityContentEffectCount(content) >= 2:
		return "unsupported ordered effect sequence",
			"structural — multi-effect body not lowered as a sequence"
	default:
		return "incomplete executable lowering",
			"the executable source backend did not consume every semantic element and source token"
	}
}

// abilityContentHasDelayedEffect reports whether any resolving effect, including
// those nested in modes, carries a delayed trigger timing.
func abilityContentHasDelayedEffect(content compiler.AbilityContent) bool {
	if slices.ContainsFunc(content.Effects, func(effect compiler.CompiledEffect) bool {
		return effect.DelayedTiming != 0
	}) {
		return true
	}
	return slices.ContainsFunc(content.Modes, func(mode compiler.CompiledMode) bool {
		return abilityContentHasDelayedEffect(mode.Content)
	})
}

// abilityContentEffectCount counts resolving effects across the body and any
// nested modes, identifying multi-effect bodies that require ordered lowering.
func abilityContentEffectCount(content compiler.AbilityContent) int {
	count := len(content.Effects)
	for i := range content.Modes {
		count += abilityContentEffectCount(content.Modes[i].Content)
	}
	return count
}

// eternalizeFamilyCreatureSubtypes converts a parsed type line's subtypes to the
// runtime creature subtypes an Eternalize/Embalm token copy re-adds, dropping any
// printed Zombie type so the keyword's granted Zombie type is not duplicated.
func eternalizeFamilyCreatureSubtypes(subtypes []string) []types.Sub {
	result := make([]types.Sub, 0, len(subtypes))
	for _, subtype := range subtypes {
		if types.Sub(subtype) == types.Zombie {
			continue
		}
		result = append(result, types.Sub(subtype))
	}
	return result
}

// keywordOnlyContent reports whether an ability's content is a bare keyword
// declaration carrying keywords but no effects, targets, conditions,
// references, or modes. Such a paragraph is a static keyword ability even when
// it compiles as a spell ability on an instant or sorcery.
func keywordOnlyContent(content compiler.AbilityContent) bool {
	return len(content.Keywords) > 0 &&
		len(content.Effects) == 0 &&
		len(content.Targets) == 0 &&
		len(content.Conditions) == 0 &&
		len(content.References) == 0 &&
		len(content.Modes) == 0
}

// lowerStaticKeywordLowering lowers a keyword-only ability to its reusable
// static keyword bodies, accumulating the keyword, ability-word, reminder, and
// keyword-list separator source spans the ability consumes.
// lowerCompanionAbility lowers a recognized companion keyword ability (CR
// 702.139) to the inert companion static keyword. The companion deckbuilding
// condition and the from-outside-the-game put-into-hand permission are sideboard
// and deck-construction mechanics the deterministic playtester does not simulate,
// so the keyword carries no in-game effect; the whole paragraph span is consumed.
func lowerCompanionAbility(ability compiler.CompiledAbility) abilityLowering {
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{
			{Body: game.CompanionStaticBody, VarName: "game.CompanionStaticBody"},
		},
		sourceSpans: []shared.Span{ability.Span},
	}
}

func lowerStaticKeywordLowering(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
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
	spans = appendKeywordListSemicolonSpans(spans, syntax.Tokens)
	return abilityLowering{
		staticAbilities: bodies,
		consumed: semanticConsumption{
			keywords: len(ability.Content.Keywords),
		},
		sourceSpans: spans,
	}, nil
}

func lowerExecutableAbility(
	cardName string,
	saga bool,
	creatureSubtypes []types.Sub,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if lowered, handled, diagnostic := lowerExecutableAbilitySpecialCase(cardName, creatureSubtypes, ability, syntax); handled {
		return lowered, diagnostic
	}
	switch ability.Kind {
	case compiler.AbilityStatic:
		return lowerStaticKeywordLowering(ability, syntax)
	case compiler.AbilityActivated:
		return lowerActivatedAbilityKind(cardName, ability, syntax)
	case compiler.AbilityLoyalty:
		return lowerLoyaltyAbility(cardName, ability, syntax)
	case compiler.AbilitySpell:
		// A spell keyword on its own paragraph (e.g. Delve, Convoke, or Storm on
		// an instant or sorcery) compiles as a keyword-only spell ability. It is
		// a static keyword ability, not a resolving spell effect, so lower it
		// through the same static keyword path the permanent keywords use.
		if keywordOnlyContent(ability.Content) {
			return lowerStaticKeywordLowering(ability, syntax)
		}
		body, bodySyntax, ok := spellBodyWithoutAbilityWord(ability, syntax)
		if !ok {
			return abilityLowering{}, executableDiagnostic(
				ability,
				"unsupported ability word",
				fmt.Sprintf("the executable source backend does not yet lower the %q ability word", ability.AbilityWord),
			)
		}
		if body.ExactSequence == compiler.ExactSequenceBottomHandThenDraw {
			spellAbility := lowerBottomHandThenDrawSequence(body)
			return abilityLowering{
				spellAbility: opt.Val(spellAbility),
				sourceSpans:  []shared.Span{body.Content.Span},
			}, nil
		}
		if body.ExactSequence == compiler.ExactSequenceDiscardHandThenDraw {
			spellAbility := lowerDiscardHandThenDrawSequence(body)
			return abilityLowering{
				spellAbility: opt.Val(spellAbility),
				sourceSpans:  []shared.Span{body.Content.Span},
			}, nil
		}
		if len(body.Content.Effects) == 1 &&
			body.Content.Effects[0].Kind == compiler.EffectAddMana &&
			(body.Content.Effects[0].Mana.AnyColor || body.Content.Effects[0].Mana.FilterPair) {
			return abilityLowering{}, executableDiagnostic(
				ability,
				"unsupported mana symbol",
				"the executable source backend cannot lower this add-mana content",
			)
		}
		spellAbility, diagnostic := lowerSpellAbilityContent(cardName, body.Content, body.Optional, &bodySyntax)
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
			if ability.Content.Effects[i].Payment.Span != (shared.Span{}) {
				spans = append(spans, ability.Content.Effects[i].Payment.Span)
			}
			if ability.Content.Effects[i].PreventRegeneration {
				spans = append(spans, ability.Content.Effects[i].RegenerationRiderSpan)
			}
			if ability.Content.Effects[i].HandChoiceDiscard.Present {
				spans = append(spans, ability.Content.Effects[i].HandChoiceDiscard.ChooseSpan)
			}
			if len(ability.Content.Effects[i].TokenCopyGrantKeywords) != 0 {
				spans = append(spans, ability.Content.Effects[i].TokenCopyGrantRiderSpan)
			}
			if ability.Content.Effects[i].ReturnAsEnchantment {
				spans = append(spans, ability.Content.Effects[i].ReturnAsEnchantmentRiderSpan)
			}
			if ability.Content.Effects[i].CopyMayChooseNewTargets {
				spans = append(spans, ability.Content.Effects[i].CopyChooseNewTargetsRiderSpan)
			}
			if ability.Content.Effects[i].PlayFromTopPayLife {
				spans = append(spans, ability.Content.Effects[i].PlayFromTopPayLifeRiderSpan)
			}
		}
		for _, target := range ability.Content.Targets {
			spans = append(spans, target.Span)
			if target.ChoiceSpan != (shared.Span{}) {
				spans = append(spans, target.ChoiceSpan)
			}
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
		lowered, diagnostic := lowerReplacementAbility(ability)
		if diagnostic == nil {
			for i := range syntax.Reminders {
				lowered.sourceSpans = append(lowered.sourceSpans, syntax.Reminders[i].Span)
			}
			if syntax.AbilityWord != nil && replacementAbilityWordConsumed(lowered) {
				lowered.sourceSpans = append(lowered.sourceSpans,
					syntax.AbilityWord.Span, syntax.AbilityWord.SeparatorSpan)
			}
		}
		return lowered, diagnostic
	case compiler.AbilitySpellAdditionalCost:
		return lowerSpellAdditionalCost(cardName, ability)
	case compiler.AbilitySpellAlternativeCost:
		return lowerSpellAlternativeCost(cardName, ability)
	case compiler.AbilityReminder:
		if saga && syntax.SagaReminder {
			return abilityLowering{sourceSpans: []shared.Span{ability.Span}}, nil
		}
		if syntax.ClassReminder {
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

func lowerSpellAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostFlashback {
		return lowerFlashbackAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostEscape {
		return lowerEscapeAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil &&
		ability.AlternativeCost.Kind == compiler.AlternativeCostOverload &&
		ability.AlternativeCost.ReplaceTargetWithEach &&
		len(ability.AlternativeCost.ManaCost) > 0 &&
		overloadManaCostSupported(ability.AlternativeCost.ManaCost) &&
		ability.Cost == nil &&
		len(ability.Content.Effects) == 0 &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.Keywords) == 0 &&
		len(ability.Content.Modes) == 0 {
		return abilityLowering{
			overloadCost: opt.Val(slices.Clone(ability.AlternativeCost.ManaCost)),
			consumed: semanticConsumption{
				alternativeCost: true,
				references:      len(ability.Content.References),
			},
			sourceSpans: []shared.Span{ability.Span},
		}, nil
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostPitch {
		return lowerPitchAlternativeCost(ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostDiscard {
		return lowerDiscardAlternativeCost(ability)
	}
	if ability.AlternativeCost == nil ||
		(ability.AlternativeCost.Kind != compiler.AlternativeCostUnknown &&
			ability.AlternativeCost.Kind != compiler.AlternativeCostCommander) ||
		ability.AlternativeCost.Condition != compiler.AlternativeCostConditionControlsCommander ||
		!ability.AlternativeCost.WithoutPayingManaCost ||
		ability.Cost != nil ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:     "Cast without paying mana cost",
			Condition: cost.AlternativeConditionControlsCommander,
		}},
		consumed: semanticConsumption{
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func overloadManaCostSupported(manaCost cost.Mana) bool {
	for _, symbol := range manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return false
		}
	}
	return true
}

// lowerFlashbackAlternativeCost lowers the em-dash Flashback form
// "Flashback—<cost>" into a SimpleKeyword(Flashback) grant plus a Flashback
// alternative cost carrying the non-mana (or compound) cost typed by the shared
// cost machinery. The runtime gates graveyard flashback casting on the keyword
// grant and pays the alternative's mana and additional costs, then exiles the
// spell. It fails closed when the cost is unrecognized.
func lowerFlashbackAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost == nil || len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the flashback cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this flashback cost",
		)
	}
	alternative := cost.Alternative{
		Label:           "Flashback",
		AdditionalCosts: additionalCosts,
	}
	if len(manaCost) > 0 {
		alternative.ManaCost = opt.Val(manaCost)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body: game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flashback}},
			},
		}},
		alternativeCosts: []cost.Alternative{alternative},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerEscapeAlternativeCost lowers the em-dash Escape form
// "Escape—<cost>, Exile N cards from your graveyard." into a
// SimpleKeyword(Escape) grant plus an Escape alternative cost carrying the
// compound escape cost typed by the shared cost machinery (its mana cost plus
// the graveyard-exile additional cost). The runtime gates graveyard escape
// casting on the keyword grant and pays the alternative's mana and additional
// costs. Unlike Flashback the spell is not exiled, so it can be escaped again.
// It fails closed when the cost is unrecognized.
func lowerEscapeAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost == nil || len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the escape cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this escape cost",
		)
	}
	alternative := cost.Alternative{
		Label:           "Escape",
		AdditionalCosts: additionalCosts,
	}
	if len(manaCost) > 0 {
		alternative.ManaCost = opt.Val(manaCost)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body: game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Escape}},
			},
		}},
		alternativeCosts: []cost.Alternative{alternative},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// free (no-mana) alternative whose additional costs exile a colored card from
// hand and optionally pay life, gated by the optional not-your-turn condition.
func lowerPitchAlternativeCost(ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	alternative := ability.AlternativeCost
	unsupported := alternative == nil ||
		!alternative.PitchColorKnown ||
		alternative.PitchCount < 1 ||
		alternative.PitchLife < 0 ||
		alternative.WithoutPayingManaCost ||
		len(alternative.ManaCost) != 0 ||
		ability.Cost != nil ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0
	condition, conditionOK := lowerAlternativeCostCondition(alternative)
	if unsupported || !conditionOK {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}
	var additionalCosts []cost.Additional
	if alternative.PitchLife > 0 {
		additionalCosts = append(additionalCosts, cost.Additional{
			Kind:   cost.AdditionalPayLife,
			Amount: alternative.PitchLife,
		})
	}
	additionalCosts = append(additionalCosts, cost.Additional{
		Kind:           cost.AdditionalExile,
		Amount:         alternative.PitchCount,
		Source:         zone.Hand,
		MatchCardColor: true,
		CardColor:      alternative.PitchColor,
	})
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:           pitchAlternativeLabel(alternative.PitchColor),
			AdditionalCosts: additionalCosts,
			Condition:       condition,
		}},
		consumed: semanticConsumption{
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerDiscardAlternativeCost lowers the Foil/Outbreak family: a free (no-mana)
// alternative whose additional costs discard one or more cards from hand,
// optionally constrained by subtype, rather than paying the printed mana cost.
func lowerDiscardAlternativeCost(ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	alternative := ability.AlternativeCost
	unsupported := alternative == nil ||
		len(alternative.DiscardCards) == 0 ||
		alternative.WithoutPayingManaCost ||
		len(alternative.ManaCost) != 0 ||
		alternative.PitchColorKnown ||
		ability.Cost != nil ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0
	condition, conditionOK := lowerAlternativeCostCondition(alternative)
	if unsupported || !conditionOK {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}
	var additionalCosts []cost.Additional
	for _, card := range alternative.DiscardCards {
		additional := cost.Additional{
			Kind:   cost.AdditionalDiscard,
			Amount: 1,
			Source: zone.Hand,
		}
		if card.HasSubtype {
			additional.SubtypesAny = cost.SubtypeSet{card.Subtype}
		}
		additionalCosts = append(additionalCosts, additional)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:           discardAlternativeLabel(alternative.DiscardCards),
			AdditionalCosts: additionalCosts,
			Condition:       condition,
		}},
		consumed: semanticConsumption{
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func discardAlternativeLabel(cards []compiler.CompiledAlternativeDiscardCard) string {
	parts := make([]string, 0, len(cards))
	for i, card := range cards {
		switch {
		case card.HasSubtype:
			parts = append(parts, indefiniteArticle(string(card.Subtype))+" "+string(card.Subtype)+" card")
		case i > 0:
			parts = append(parts, "another card")
		default:
			parts = append(parts, "a card")
		}
	}
	return "Discard " + strings.Join(parts, " and ")
}

func indefiniteArticle(word string) string {
	if word == "" {
		return "a"
	}
	switch word[0] {
	case 'A', 'E', 'I', 'O', 'U', 'a', 'e', 'i', 'o', 'u':
		return "an"
	default:
		return "a"
	}
}

func lowerAlternativeCostCondition(alternative *compiler.CompiledAlternativeCost) (cost.AlternativeCondition, bool) {
	switch alternative.Condition {
	case compiler.AlternativeCostConditionUnknown:
		return cost.AlternativeConditionNone, true
	case compiler.AlternativeCostConditionNotYourTurn:
		return cost.AlternativeConditionNotYourTurn, true
	default:
		return cost.AlternativeConditionNone, false
	}
}

func pitchAlternativeLabel(c color.Color) string {
	if name, ok := colorDisplayName(c); ok {
		return "Exile a " + name + " card"
	}
	return "Exile a card"
}

func colorDisplayName(c color.Color) (string, bool) {
	switch c {
	case color.White:
		return "white", true
	case color.Blue:
		return "blue", true
	case color.Black:
		return "black", true
	case color.Red:
		return "red", true
	case color.Green:
		return "green", true
	default:
		return "", false
	}
}

func lowerOverloadSpell(
	normal opt.V[game.AbilityContent],
	spell *compiler.CompiledAbility,
) (game.AbilityContent, bool) {
	if !normal.Exists || spell == nil ||
		len(spell.Content.Targets) != 1 ||
		!spell.Content.Targets[0].Exact ||
		spell.Content.Targets[0].Cardinality.Min != 1 ||
		spell.Content.Targets[0].Cardinality.Max != 1 ||
		len(normal.Val.Modes) != 1 ||
		len(normal.Val.SharedTargets)+len(normal.Val.Modes[0].Targets) != 1 ||
		len(normal.Val.Modes[0].Sequence) != 1 {
		return game.AbilityContent{}, false
	}
	selection, ok := massGroupSelection(spell.Content.Targets[0].Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	group := game.BattlefieldGroup(selection)
	instruction := normal.Val.Modes[0].Sequence[0]
	switch primitive := instruction.Primitive.(type) {
	case game.Destroy:
		if primitive.Object != game.TargetPermanentReference(0) || primitive.Group.Valid() {
			return game.AbilityContent{}, false
		}
		primitive.Object = game.ObjectReference{}
		primitive.Group = group
		instruction.Primitive = primitive
	case game.Tap:
		if primitive.Object != game.TargetPermanentReference(0) || primitive.Group.Valid() {
			return game.AbilityContent{}, false
		}
		primitive.Object = game.ObjectReference{}
		primitive.Group = group
		instruction.Primitive = primitive
	case game.Untap:
		if primitive.Object != game.TargetPermanentReference(0) || primitive.Group.Valid() {
			return game.AbilityContent{}, false
		}
		primitive.Object = game.ObjectReference{}
		primitive.Group = group
		instruction.Primitive = primitive
	case game.Bounce:
		if primitive.Object != game.TargetPermanentReference(0) ||
			primitive.Group.Valid() ||
			primitive.ControlledChoice {
			return game.AbilityContent{}, false
		}
		primitive.Object = game.ObjectReference{}
		primitive.Group = group
		instruction.Primitive = primitive
	default:
		return game.AbilityContent{}, false
	}
	mode := normal.Val.Modes[0]
	mode.Targets = nil
	mode.Sequence = []game.Instruction{instruction}
	overloaded := normal.Val
	overloaded.SharedTargets = nil
	overloaded.Modes = []game.Mode{mode}
	return overloaded, true
}

// lowerSpellAdditionalCost lowers a spell additional-cost paragraph ("As an
// additional cost to cast this spell, <cost>.") into typed cost.Additional
// values, reusing the shared activated-ability cost lowering. The paragraph has
// no resolving body of its own; its only semantic element is the cost. It fails
// closed when any cost component is not a recognized additional cost.
func lowerSpellAdditionalCost(
	cardName string,
	ability compiler.CompiledAbility,
) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost == nil || len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported activation cost",
			"the executable source backend could not recognize the spell's additional cost",
		)
	}
	additional := make([]cost.Additional, 0, len(ability.Cost.Components))
	for _, component := range ability.Cost.Components {
		lowered, ok := lowerActivatedAdditionalCost(cardName, component)
		if !ok {
			return abilityLowering{}, executableDiagnostic(
				ability,
				"unsupported activation cost",
				"the executable source backend does not yet lower this additional cost to cast",
			)
		}
		lowered.ChoiceGroup = component.ChoiceGroup
		additional = append(additional, lowered)
	}
	return abilityLowering{
		additionalCosts: additional,
		consumed: semanticConsumption{
			cost:       true,
			references: len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func lowerExecutableAbilitySpecialCase(
	cardName string,
	creatureSubtypes []types.Sub,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if len(ability.Content.Modes) > 0 &&
		ability.Kind != compiler.AbilityActivated &&
		ability.Kind != compiler.AbilityTriggered &&
		ability.Kind != compiler.AbilityChapter {
		lowered, diagnostic := lowerModalAbility(cardName, ability, syntax)
		return lowered, true, diagnostic
	}
	if ability.Companion {
		return lowerCompanionAbility(ability), true, nil
	}
	if lowered, handled, diagnostic := lowerSourceSpellCostReduction(ability, syntax); handled {
		return lowered, true, diagnostic
	}
	if lowered, ok := lowerEntersPrepared(ability, syntax); ok {
		return lowered, true, nil
	}
	if lowered, ok, diagnostic := lowerStaticDeclarations(ability, syntax); ok {
		return lowered, true, diagnostic
	}
	if diagnostic := lowerStaticDeclarationBlocker(ability); diagnostic != nil {
		return abilityLowering{}, true, diagnostic
	}
	if lowered, ok, diagnostic := lowerKeywordDispatch(creatureSubtypes, ability, syntax); ok {
		return lowered, true, diagnostic
	}
	return abilityLowering{}, false, nil
}
