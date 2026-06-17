package cardgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func lowerEnchantAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordEnchant {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	target, ok := enchantTargetSpec(keyword.EnchantTarget)
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
	if !keywordOnlyCovered(syntax, keyword) {
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
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordProtection {
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
	if !keywordOnlyCovered(syntax, keyword) {
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
	syntax *parser.Ability,
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
	syntax *parser.Ability,
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
	syntax *parser.Ability,
) abilityLowering {
	spans := keywordSpans(ability, syntax)
	return abilityLowering{
		activatedAbility: opt.Val(*body),
		consumed:         semanticConsumption{keywords: 1},
		sourceSpans:      spans,
	}
}

func keywordSpans(ability compiler.CompiledAbility, syntax *parser.Ability) []shared.Span {
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

func enchantTargetSpec(targetKind parser.ObjectNoun) (game.TargetSpec, bool) {
	target := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
	}
	if targetKind == parser.ObjectNounPlayer {
		target.Constraint = "player"
		target.Allow = game.TargetAllowPlayer
		return target, true
	}
	target.Allow = game.TargetAllowPermanent
	switch targetKind {
	case parser.ObjectNounArtifact:
		target.Constraint = "artifact"
		target.Predicate.PermanentTypes = []types.Card{types.Artifact}
	case parser.ObjectNounCreature:
		target.Constraint = "creature"
		target.Predicate.PermanentTypes = []types.Card{types.Creature}
	case parser.ObjectNounEnchantment:
		target.Constraint = "enchantment"
		target.Predicate.PermanentTypes = []types.Card{types.Enchantment}
	case parser.ObjectNounLand:
		target.Constraint = "land"
		target.Predicate.PermanentTypes = []types.Card{types.Land}
	case parser.ObjectNounPermanent:
		target.Constraint = "permanent"
	case parser.ObjectNounPlaneswalker:
		target.Constraint = "planeswalker"
		target.Predicate.PermanentTypes = []types.Card{types.Planeswalker}
	default:
		return game.TargetSpec{}, false
	}
	return target, true
}

func lowerEquipAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordEquip {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
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
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	return game.EquipActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

func lowerCyclingAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordCycling {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind == parser.KeywordParameterNone &&
		(len(ability.Content.Targets) != 0 || len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0) {
		return game.ActivatedAbility{}, false, nil
	}
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
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
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	return game.CyclingActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

func lowerNinjutsuAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordNinjutsu {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
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
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	return game.NinjutsuActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

func lowerMutateAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordMutate {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
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
	if len(keyword.ManaCost) == 0 {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	return game.MutateStaticAbility(slices.Clone(keyword.ManaCost)), true, nil
}

func mixedStaticKeywords(keywords []compiler.CompiledKeyword) ([]game.Keyword, bool) {
	result := make([]game.Keyword, 0, len(keywords))
	for _, keyword := range keywords {
		if keyword.ParameterKind != parser.KeywordParameterNone {
			return nil, false
		}
		body, ok := keywordStaticBodies[keyword.Kind]
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
		if keyword.Kind == parser.KeywordCycling && keyword.ParameterKind == parser.KeywordParameterNone {
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

func abilityUsesCyclingSelectorPredicate(content compiler.AbilityContent) bool {
	for _, target := range content.Targets {
		if target.Selector.Keyword == parser.KeywordCycling {
			return true
		}
	}
	for i := range content.Effects {
		if content.Effects[i].Selector.Keyword == parser.KeywordCycling ||
			content.Effects[i].Amount.Selector().Keyword == parser.KeywordCycling {
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

func resolvingStaticSubjectGroup(effect *compiler.CompiledEffect) (game.GroupReference, bool) {
	selection := game.Selection{Controller: game.ControllerYou}
	switch effect.StaticSubject {
	case compiler.StaticSubjectAllCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}), true
	case compiler.StaticSubjectAllOtherCreatures:
		return game.BattlefieldGroupExcluding(
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.SourcePermanentReference(),
		), true
	case compiler.StaticSubjectAttackingCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		}), true
	case compiler.StaticSubjectBlockingCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateBlocking,
		}), true
	case compiler.StaticSubjectControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
	case compiler.StaticSubjectOtherControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
		return game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()), true
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
	case compiler.StaticSubjectOtherControlledCreatureSubtype:
		if !effect.StaticSubjectSubKnown() {
			return game.GroupReference{}, false
		}
		selection.SubtypesAny = []types.Sub{effect.StaticSubjectSub()}
		return game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()), true
	default:
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

func lowerKeywordAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) ([]loweredStaticAbility, *shared.Diagnostic) {
	for _, keyword := range ability.Content.Keywords {
		if keyword.Kind == parser.KeywordDevoid && !syntax.DevoidRecognized {
			return nil, executableDiagnostic(
				ability,
				"unsupported Devoid ability",
				"the executable source backend supports only exact \"Devoid (This card has no color.)\" abilities",
			)
		}
		if keyword.Kind == parser.KeywordReadAhead && !syntax.ReadAheadRecognized {
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
		if keyword.ParameterKind != parser.KeywordParameterNone {
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
		body, ok := keywordStaticBodies[keyword.Kind]
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
	for _, span := range syntax.CoverageSpans() {
		if (syntax.AbilityWord != nil && span == syntax.AbilityWord.SeparatorSpan) ||
			spanCoveredByAbilityWord(span, syntax.AbilityWord) ||
			spanCoveredByKeyword(span, ability.Content.Keywords) ||
			spanCoveredByDelimited(span, syntax.Reminders) {
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

func syntaxWithoutAbilityWord(syntax *parser.Ability) parser.Ability {
	result := *syntax
	if result.AbilityWord == nil {
		return result
	}
	result.Tokens = parser.TokensFrom(result.Tokens, result.AbilityWord.SeparatorSpan.End.Offset)
	return result
}

func spellBodyWithoutAbilityWord(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (compiler.CompiledAbility, parser.Ability, bool) {
	if ability.AbilityWord == "" {
		return ability, *syntax, true
	}
	if !rulesFreeAbilityWordLabel(ability.AbilityWord) || syntax.AbilityWord == nil {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	body := syntaxWithoutAbilityWord(syntax)
	if len(body.Tokens) == 0 {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	start := body.Tokens[0].Span.Start
	offset := start.Offset - ability.Span.Start.Offset
	if offset < 0 || offset >= len(ability.Text) {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	ability.Text = strings.TrimSpace(ability.Text[offset:])
	ability.Span.Start = start
	ability.AbilityWord = ""
	body.Span.Start = start
	body.Text = ability.Text
	body.AbilityWord = nil
	return ability, body, true
}

func tokensWithoutSpans(tokens []shared.Token, spans ...shared.Span) []shared.Token {
	return slices.DeleteFunc(append([]shared.Token(nil), tokens...), func(token shared.Token) bool {
		return spanCovered(token.Span, spans)
	})
}

// lowerParameterizedKeywordToStaticAbility handles lowering of a single
// parameterized keyword (Ward, Protection, and others) to a static ability.
// Returns (body, true, nil) on success, ({}, true, diag) on a recognised but
// unsupported form, and ({}, false, nil) when no handler matches.
func lowerParameterizedKeywordToStaticAbility(
	ability compiler.CompiledAbility,
	keyword compiler.CompiledKeyword,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	switch keyword.Kind {
	case parser.KeywordWard:
		if keyword.ParameterKind == parser.KeywordParameterManaCost && len(keyword.ManaCost) > 0 {
			return game.WardStaticAbility(slices.Clone(keyword.ManaCost)), true, nil
		}
	case parser.KeywordProtection:
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
	switch keyword.Kind {
	case parser.KeywordKicker:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.KickerKeyword{Cost: manaCost}}
	case parser.KeywordMadness:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.MadnessKeyword{Cost: manaCost}}
	case parser.KeywordMorph:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.MorphKeyword{Cost: manaCost}}
	case parser.KeywordDisguise:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.DisguiseKeyword{Cost: manaCost}}
	case parser.KeywordToxic:
		if keyword.ParameterKind != parser.KeywordParameterInteger || keyword.Integer <= 0 {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.ToxicKeyword{Amount: keyword.Integer}}
	default:
		return game.StaticAbility{}, false
	}
	return body, true
}

func fixedKeywordManaCost(keyword compiler.CompiledKeyword) (cost.Mana, bool) {
	if keyword.ParameterKind != parser.KeywordParameterManaCost || len(keyword.ManaCost) == 0 {
		return nil, false
	}
	for _, symbol := range keyword.ManaCost {
		if symbol.Kind == cost.VariableSymbol {
			return nil, false
		}
	}
	return slices.Clone(keyword.ManaCost), true
}
