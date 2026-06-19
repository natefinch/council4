package game

import (
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// CardDefIssueCode identifies a class of structural CardDef validation issue.
type CardDefIssueCode string

// Structural validation issue codes identify problems found purely from game
// data without any tooling or runtime policy.
const (
	CardDefIssueNilCard                CardDefIssueCode = "nil-card"
	CardDefIssueMissingName            CardDefIssueCode = "missing-name"
	CardDefIssueOracleWithoutAbilities CardDefIssueCode = "oracle-without-abilities"
	CardDefIssueTargetIndexOutOfRange  CardDefIssueCode = "target-index-out-of-range"
	CardDefIssueInvalidReference       CardDefIssueCode = "invalid-reference"
	CardDefIssueInvalidTargetSpec      CardDefIssueCode = "invalid-target-spec"
	CardDefIssueInvalidKeywordAbility  CardDefIssueCode = "invalid-keyword-ability"
	CardDefIssueInvalidAbilityBody     CardDefIssueCode = "invalid-ability-body"
	CardDefIssueInvalidSelection       CardDefIssueCode = "invalid-selection"
	CardDefIssueInvalidCondition       CardDefIssueCode = "invalid-condition"
	CardDefIssueInvalidRuleEffect      CardDefIssueCode = "invalid-rule-effect"
)

// CardDefIssue describes one structural problem found in a CardDef.
type CardDefIssue struct {
	// FaceName is the name of the face the issue was found in, or empty for
	// card-level issues.
	FaceName string `json:"face_name,omitempty"`

	// Path is the dot-separated field path within the card definition where
	// the issue was found, or empty for top-level issues.
	Path string `json:"path,omitempty"`

	// Code identifies the class of issue.
	Code CardDefIssueCode `json:"code"`

	// Message is a human-readable description of the issue.
	Message string `json:"message"`
}

// ValidateCardDef performs deep structural validation of a CardDef and returns
// all issues found. A nil card produces a single CardDefIssueNilCard issue.
// ValidateCardDef is a package function rather than a method so that nil
// CardDef values can be diagnosed without a valid receiver.
func ValidateCardDef(card *CardDef) []CardDefIssue {
	v := &cardDefValidator{card: card}
	v.validate()
	return v.issues
}

type cardDefValidator struct {
	card   *CardDef
	issues []CardDefIssue
}

func (v *cardDefValidator) validate() {
	if v.card == nil {
		v.add("", "", CardDefIssueNilCard, "card definition is nil")
		return
	}
	if strings.TrimSpace(v.card.Name) == "" {
		v.add("", "", CardDefIssueMissingName, "card definition has no name")
	}
	v.validateFace(v.card.Name, "", &v.card.CardFace)
	if v.card.Back.Exists {
		face := v.card.Back.Val
		name := face.Name
		if strings.TrimSpace(name) == "" {
			name = "back face"
		}
		v.validateFace(name, "Back", &face)
	}
	if v.card.Alternate.Exists {
		face := v.card.Alternate.Val
		name := face.Name
		if strings.TrimSpace(name) == "" {
			name = "alternate face"
		}
		v.validateFace(name, "Alternate", &face)
	}
}

func (v *cardDefValidator) validateFace(faceName, path string, face *CardFace) {
	hasAbilities := face.SpellAbility.Exists ||
		face.EntersPrepared ||
		len(face.ActivatedAbilities) > 0 ||
		len(face.ManaAbilities) > 0 ||
		len(face.LoyaltyAbilities) > 0 ||
		len(face.TriggeredAbilities) > 0 ||
		len(face.ChapterAbilities) > 0 ||
		len(face.ReplacementAbilities) > 0 ||
		len(face.StaticAbilities) > 0 ||
		len(face.AdditionalCosts) > 0
	if strings.TrimSpace(face.OracleText) != "" && !hasAbilities && face.ImplementationID == "" {
		v.add(faceName, path, CardDefIssueOracleWithoutAbilities, "oracle text is non-empty but no abilities or hand-written implementation are defined")
	}
	if face.SpellAbility.Exists {
		v.validateAbilityBody(faceName, appendPath(path, "SpellAbility"), face.SpellAbility.Val, nil)
	}
	for i := range face.ActivatedAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("ActivatedAbilities[%d]", i)), face.ActivatedAbilities[i], nil)
	}
	for i := range face.ManaAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("ManaAbilities[%d]", i)), face.ManaAbilities[i], nil)
	}
	for i := range face.LoyaltyAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("LoyaltyAbilities[%d]", i)), face.LoyaltyAbilities[i], nil)
	}
	for i := range face.TriggeredAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("TriggeredAbilities[%d]", i)), face.TriggeredAbilities[i], nil)
	}
	for i := range face.ChapterAbilities {
		chapterPath := appendPath(path, fmt.Sprintf("ChapterAbilities[%d]", i))
		if len(face.ChapterAbilities[i].Chapters) == 0 {
			v.add(faceName, appendPath(chapterPath, "Chapters"), CardDefIssueInvalidAbilityBody, "chapter ability has no chapter numbers")
		}
		for j, chapter := range face.ChapterAbilities[i].Chapters {
			if chapter <= 0 {
				v.add(faceName, appendPath(chapterPath, fmt.Sprintf("Chapters[%d]", j)), CardDefIssueInvalidAbilityBody, "chapter number must be positive")
			}
		}
		v.validateAbilityBody(faceName, chapterPath, face.ChapterAbilities[i], nil)
	}
	for i := range face.ReplacementAbilities {
		v.validateReplacementAbility(faceName, appendPath(path, fmt.Sprintf("ReplacementAbilities[%d]", i)), &face.ReplacementAbilities[i])
	}
	for i := range face.StaticAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("StaticAbilities[%d]", i)), face.StaticAbilities[i], nil)
	}
}

func (v *cardDefValidator) validateAbilityBody(faceName, path string, body Ability, targets []TargetSpec) {
	switch abilityBody := body.(type) {
	case AbilityContent:
		v.validateAbilityContent(faceName, path, abilityBody, targets)
	case ActivatedAbility:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case ManaAbility:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		if len(abilityBody.Content.Modes) > 0 {
			v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
		}
	case LoyaltyAbility:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case TriggeredAbility:
		v.validateTriggerPattern(faceName, appendPath(path, "Trigger.Pattern"), &abilityBody.Trigger.Pattern)
		if abilityBody.Trigger.InterveningCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "Trigger.InterveningCondition"), &abilityBody.Trigger.InterveningCondition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case ChapterAbility:
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case StaticAbility:
		if abilityBody.Condition.Exists {
			v.validateCondition(faceName, appendPath(path, "Condition"), &abilityBody.Condition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		for i := range abilityBody.ContinuousEffects {
			v.validateContinuousEffect(faceName, appendPath(path, fmt.Sprintf("ContinuousEffects[%d]", i)), &abilityBody.ContinuousEffects[i], targets)
		}
		for i := range abilityBody.RuleEffects {
			v.validateRuleEffect(faceName, appendPath(path, fmt.Sprintf("RuleEffects[%d]", i)), &abilityBody.RuleEffects[i])
		}
	case nil:
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, "ability body is nil")
	default:
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, fmt.Sprintf("unknown ability body %T", body))
	}
}

func (v *cardDefValidator) validateReplacementAbility(faceName, path string, ability *ReplacementAbility) {
	if ability == nil {
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, "replacement ability is nil")
		return
	}
	if ability.Replacement.Condition.Exists {
		v.validateCondition(faceName, appendPath(path, "Replacement.Condition"), &ability.Replacement.Condition.Val, nil)
	}
}

func (v *cardDefValidator) validateAbilityContent(faceName, path string, content AbilityContent, fallbackTargets []TargetSpec) {
	v.validateAbilityContentWithLinked(faceName, path, content, fallbackTargets, nil)
}

func (v *cardDefValidator) validateAbilityContentWithLinked(
	faceName, path string,
	content AbilityContent,
	fallbackTargets []TargetSpec,
	inheritedLinked map[LinkedKey]int,
) {
	if len(content.Modes) == 0 {
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, "ability content has no modes")
		return
	}
	for i := range content.SharedTargets {
		v.validateTargetSpec(faceName, appendPath(path, fmt.Sprintf("SharedTargets[%d]", i)), &content.SharedTargets[i])
	}
	for i := range content.Modes {
		mode := &content.Modes[i]
		modePath := appendPath(path, fmt.Sprintf("Modes[%d]", i))
		for j := range mode.Targets {
			v.validateTargetSpec(faceName, appendPath(modePath, fmt.Sprintf("Targets[%d]", j)), &mode.Targets[j])
		}
		targets := append([]TargetSpec(nil), content.SharedTargets...)
		targets = append(targets, mode.Targets...)
		if len(targets) == 0 {
			targets = fallbackTargets
		}
		v.validateInstructionSequence(faceName, appendPath(modePath, "Sequence"), mode.Sequence, targets, inheritedLinked)
	}
}

func (v *cardDefValidator) validateKeywordAbility(faceName, path string, ability KeywordAbility, targets []TargetSpec) {
	switch keyword := ability.(type) {
	case SimpleKeyword:
		if keyword.Kind == KeywordNone {
			v.add(faceName, path, CardDefIssueInvalidKeywordAbility, "simple keyword must set Kind")
		}
	case WardKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case EquipKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case EnchantKeyword:
		v.validateTargetSpec(faceName, appendPath(path, "Target"), &keyword.Target)
	case CyclingKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case NinjutsuKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case MutateKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case KickerKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
		if len(keyword.BonusContent.Modes) > 0 {
			v.validateAbilityContent(faceName, appendPath(path, "BonusContent"), keyword.BonusContent, targets)
		}
	case MadnessKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case MorphKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case DisguiseKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case SuspendKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
		if keyword.TimeCounters <= 0 {
			v.add(faceName, appendPath(path, "TimeCounters"), CardDefIssueInvalidKeywordAbility, "suspend time counters must be positive")
		}
	case ProtectionKeyword:
		// Count how many mutually exclusive predicate groups are set.
		predicateCount := 0
		if len(keyword.FromColors) > 0 {
			predicateCount++
		}
		if len(keyword.FromTypes) > 0 {
			predicateCount++
		}
		if len(keyword.FromSubtypes) > 0 {
			predicateCount++
		}
		if keyword.Multicolored {
			predicateCount++
		}
		if keyword.Monocolored {
			predicateCount++
		}
		if keyword.Everything {
			predicateCount++
		}
		if keyword.EachColor {
			predicateCount++
		}
		if predicateCount == 0 {
			v.add(faceName, appendPath(path, "FromColors"), CardDefIssueInvalidKeywordAbility, "protection needs at least one protected predicate")
		} else if predicateCount > 1 {
			v.add(faceName, path, CardDefIssueInvalidKeywordAbility, "protection must use exactly one predicate group (mixed predicates are not supported)")
		}
		// Validate that FromSubtypes values are known creature or land subtypes.
		for _, sub := range keyword.FromSubtypes {
			if !isKnownProtectionSubtype(sub) {
				v.add(faceName, appendPath(path, "FromSubtypes"), CardDefIssueInvalidKeywordAbility,
					fmt.Sprintf("unknown protection subtype %q", string(sub)))
			}
		}
		// Validate that FromTypes values are known renderable card types.
		for _, t := range keyword.FromTypes {
			if !isKnownProtectionCardType(t) {
				v.add(faceName, appendPath(path, "FromTypes"), CardDefIssueInvalidKeywordAbility,
					fmt.Sprintf("unknown protection card type %q", string(t)))
			}
		}
		// Validate that FromColors values are known magic colors.
		for _, c := range keyword.FromColors {
			if !isKnownProtectionColor(c) {
				v.add(faceName, appendPath(path, "FromColors"), CardDefIssueInvalidKeywordAbility,
					fmt.Sprintf("unknown protection color %q", string(c)))
			}
		}
	case ToxicKeyword:
		if keyword.Amount <= 0 {
			v.add(faceName, appendPath(path, "Amount"), CardDefIssueInvalidKeywordAbility, "toxic amount must be positive")
		}
	case nil:
		v.add(faceName, path, CardDefIssueInvalidKeywordAbility, "keyword ability is nil")
	default:
		v.add(faceName, path, CardDefIssueInvalidKeywordAbility, fmt.Sprintf("unknown keyword ability %T", ability))
	}
}

func (v *cardDefValidator) validateInstructionSequence(
	faceName, path string,
	seq []Instruction,
	targets []TargetSpec,
	inheritedLinked map[LinkedKey]int,
) {
	if err := validateInstructionSequenceWithLinked(seq, targets, true, inheritedLinked); err != nil {
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, err.Error())
	}
	publishedLinked := make(map[LinkedKey]int, len(inheritedLinked))
	maps.Copy(publishedLinked, inheritedLinked)
	for i := range seq {
		instructionPath := appendPath(path, fmt.Sprintf("Instructions[%d]", i))
		effectCondition := seq[i].Condition
		if effectCondition.Exists && effectCondition.Val.Condition.Exists {
			condition := effectCondition.Val.Condition.Val
			v.validateCondition(
				faceName,
				appendPath(instructionPath, "Condition.Condition"),
				&condition,
				targets,
			)
		}
		if delayed, ok := seq[i].Primitive.(CreateDelayedTrigger); ok {
			v.validateAbilityContentWithLinked(
				faceName,
				appendPath(instructionPath, "Primitive.Trigger.Content"),
				delayed.Trigger.Content,
				nil,
				publishedLinked,
			)
		}
		if emblem, ok := seq[i].Primitive.(CreateEmblem); ok {
			for j, ability := range emblem.EmblemAbilities {
				v.validateAbilityBody(
					faceName,
					appendPath(instructionPath, fmt.Sprintf("Primitive.EmblemAbilities[%d]", j)),
					ability,
					nil,
				)
			}
		}
		if replacement, ok := seq[i].Primitive.(CreateReplacement); ok && replacement.Replacement != nil {
			v.validateReplacementEffect(
				faceName,
				appendPath(instructionPath, "Primitive.Replacement"),
				replacement.Replacement,
			)
		}
		if seq[i].Primitive != nil {
			if key := seq[i].Primitive.instructionRefs().publishesLinked; key != "" {
				publishedLinked[key] = i
			}
		}
	}
}

func (v *cardDefValidator) validateManaKeywordCost(faceName, path string, manaCost cost.Mana) {
	if len(manaCost) == 0 {
		v.add(faceName, appendPath(path, "Cost"), CardDefIssueInvalidKeywordAbility, "mana-valued keyword cost must be explicit")
	}
}

const knownTargetAllows = TargetAllowPermanent | TargetAllowPlayer | TargetAllowStackObject | TargetAllowCard

func (v *cardDefValidator) validateTargetSpec(faceName, path string, target *TargetSpec) {
	if target.MinTargets < 0 || target.MaxTargets < 0 {
		v.add(faceName, path, CardDefIssueInvalidTargetSpec, "target counts must be non-negative")
		return
	}
	if target.MaxTargets < target.MinTargets {
		v.add(faceName, path, CardDefIssueInvalidTargetSpec, "max targets is less than min targets")
	}
	if target.Allow&^knownTargetAllows != 0 {
		v.add(faceName, appendPath(path, "Allow"), CardDefIssueInvalidTargetSpec, "unknown target allow category")
	}
	v.validateStackObjectTargetPredicate(faceName, path, target)
	if target.Selection.Exists {
		selection := target.Selection.Val
		v.validateSelection(faceName, appendPath(path, "Selection"), selection)
		if !target.Predicate.Selection().Empty() {
			v.add(faceName, path, CardDefIssueInvalidSelection, "TargetSpec sets both Predicate and Selection")
		}
		if target.Allow == TargetAllowUnspecified {
			v.add(faceName, path, CardDefIssueInvalidTargetSpec, "Selection-based TargetSpec must set Allow")
		}
		allowsPermanents := target.Allow&TargetAllowPermanent != 0
		allowsPlayers := target.Allow&TargetAllowPlayer != 0
		allowsCards := target.Allow&TargetAllowCard != 0
		if allowsPlayers && selectionHasPermanentPredicates(selection) {
			v.add(faceName, appendPath(path, "Selection"), CardDefIssueInvalidSelection, "player targets cannot use permanent Selection predicates")
		}
		if !allowsPlayers && selection.Player != PlayerAny {
			v.add(faceName, appendPath(path, "Selection.Player"), CardDefIssueInvalidSelection, "non-player targets cannot use a player relation")
		}
		if !allowsPermanents && !allowsPlayers && !allowsCards && !selection.Empty() {
			v.add(faceName, appendPath(path, "Selection"), CardDefIssueInvalidSelection, "Selection requires permanent, card, or player targets")
		}
	}

	switch target.Chooser {
	case TargetChooserController:
	case TargetChooserOpponent:
		if target.MinTargets != 1 || target.MaxTargets != 1 {
			v.add(faceName, path, CardDefIssueInvalidTargetSpec, "non-controller target chooser requires exactly one target")
		}
		controller := target.Predicate.Controller
		if target.Selection.Exists {
			controller = target.Selection.Val.Controller
		}
		if controller != ControllerAny && controller != ControllerYou {
			field := "Predicate.Controller"
			if target.Selection.Exists {
				field = "Selection.Controller"
			}
			v.add(faceName, appendPath(path, field), CardDefIssueInvalidTargetSpec, "opponent target chooser only supports controller-any or controller-you predicates")
		}
	default:
		v.add(faceName, appendPath(path, "Chooser"), CardDefIssueInvalidTargetSpec, "unknown target chooser")
	}
}

func (v *cardDefValidator) validateStackObjectTargetPredicate(faceName, path string, target *TargetSpec) {
	kinds := target.Predicate.StackObjectKinds
	knownAllows := target.Allow & knownTargetAllows
	allowsStackObjects := knownAllows&TargetAllowStackObject != 0
	stackSelection := target.Predicate.Selection()
	// Controller restrictions are supported for stack-object targets (e.g.
	// "target activated ability you don't control"), so they do not count as an
	// unsupported permanent predicate here.
	stackSelection.Controller = ControllerAny
	if allowsStackObjects && !stackSelection.Empty() {
		v.add(faceName, appendPath(path, "Predicate"), CardDefIssueInvalidTargetSpec, "stack-object target uses unsupported predicates")
	}
	if allowsStackObjects && target.Selection.Exists {
		v.add(faceName, appendPath(path, "Selection"), CardDefIssueInvalidTargetSpec, "stack-object target cannot use Selection")
	}
	if allowsStackObjects && len(kinds) == 0 {
		v.add(faceName, appendPath(path, "Predicate.StackObjectKinds"), CardDefIssueInvalidTargetSpec, "stack-object target must allow at least one stack-object kind")
		return
	}
	if len(kinds) > 0 && !allowsStackObjects {
		v.add(faceName, appendPath(path, "Predicate.StackObjectKinds"), CardDefIssueInvalidTargetSpec, "stack-object kinds require stack-object targets")
	}
	seen := make(map[StackObjectKind]bool, len(kinds))
	allowsSpells := false
	allowsAbilities := false
	for i, kind := range kinds {
		switch kind {
		case StackSpell:
			allowsSpells = true
		case StackActivatedAbility, StackTriggeredAbility:
			allowsAbilities = true
		default:
			v.add(faceName, appendPath(path, fmt.Sprintf("Predicate.StackObjectKinds[%d]", i)), CardDefIssueInvalidTargetSpec, "unknown stack-object kind")
		}
		if seen[kind] {
			v.add(faceName, appendPath(path, fmt.Sprintf("Predicate.StackObjectKinds[%d]", i)), CardDefIssueInvalidTargetSpec, "duplicate stack-object kind")
		}
		seen[kind] = true
	}
	hasSpellTypePredicate := len(target.Predicate.SpellCardTypes) > 0 || len(target.Predicate.ExcludedSpellCardTypes) > 0
	if hasSpellTypePredicate && (!allowsSpells || allowsAbilities) {
		v.add(faceName, appendPath(path, "Predicate"), CardDefIssueInvalidTargetSpec, "spell type predicates require spell-only stack-object targets")
	}
	// SpellSupertypes, SpellColorless, SpellColors, SpellExcludedColors, and
	// SpellMulticolored qualify only matched spells, so they may accompany ability
	// kinds in a mixed target but require that spells be allowed.
	hasSpellShapePredicate := len(target.Predicate.SpellSupertypes) > 0 ||
		target.Predicate.SpellColorless ||
		len(target.Predicate.SpellColors) > 0 ||
		len(target.Predicate.SpellExcludedColors) > 0 ||
		target.Predicate.SpellMulticolored
	if hasSpellShapePredicate && !allowsSpells {
		v.add(faceName, appendPath(path, "Predicate"), CardDefIssueInvalidTargetSpec, "spell shape predicates require a stack-object target that allows spells")
	}
	if len(target.Predicate.StackObjectSourceTypes) > 0 && !allowsStackObjects {
		v.add(faceName, appendPath(path, "Predicate.StackObjectSourceTypes"), CardDefIssueInvalidTargetSpec, "stack-object source types require stack-object targets")
	}
}

func (v *cardDefValidator) validateSelection(faceName, path string, selection Selection) {
	for _, problem := range selection.Validate() {
		v.add(faceName, path, CardDefIssueInvalidSelection, problem)
	}
}

func selectionHasPermanentPredicates(selection Selection) bool {
	return len(selection.RequiredTypes) > 0 ||
		len(selection.RequiredTypesAny) > 0 ||
		len(selection.ExcludedTypes) > 0 ||
		len(selection.Supertypes) > 0 ||
		selection.ExcludedSupertype != "" ||
		len(selection.SubtypesAny) > 0 ||
		len(selection.ColorsAny) > 0 ||
		len(selection.ExcludedColors) > 0 ||
		selection.Colorless ||
		selection.Multicolored ||
		selection.Controller != ControllerAny ||
		selection.Tapped != TriAny ||
		selection.CombatState != CombatStateAny ||
		selection.Keyword != KeywordNone ||
		selection.ExcludedKeyword != KeywordNone ||
		selection.ManaValue.Exists ||
		selection.Power.Exists ||
		selection.Toughness.Exists ||
		selection.ExcludeSource ||
		selection.NonToken ||
		selection.TokenOnly
}

func (v *cardDefValidator) validateContinuousEffect(faceName, path string, continuous *ContinuousEffect, targets []TargetSpec) {
	for i := range continuous.AddAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("AddAbilities[%d]", i)), continuous.AddAbilities[i], nil)
	}
	if continuous.AffectedSource && !continuous.Group.Empty() {
		v.add(faceName, path, CardDefIssueInvalidReference, "continuous effect sets both AffectedSource and Group")
	}
	if !continuous.Group.Empty() {
		v.validateGroupRef(faceName, appendPath(path, "Group"), continuous.Group, targets)
	}
}

func (v *cardDefValidator) validateRuleEffect(faceName, path string, effect *RuleEffect) {
	if effect == nil {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "rule effect is nil")
		return
	}
	switch effect.Kind {
	case RuleEffectCostModifier:
		v.validateCostModifier(faceName, appendPath(path, "CostModifier"), effect.CostModifier)
	case RuleEffectGrantHandCardAbility:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "hand-card ability grants must set affected player")
		}
		v.validateSelection(faceName, appendPath(path, "CardSelection"), effect.CardSelection)
		if effect.CardSelection.Empty() {
			v.add(faceName, appendPath(path, "CardSelection"), CardDefIssueInvalidSelection, "hand-card ability grants require a card selection")
		}
		if handCardSelectionHasUnsupportedPredicates(effect.CardSelection) {
			v.add(faceName, appendPath(path, "CardSelection"), CardDefIssueInvalidSelection, "hand-card ability grants support only printed card characteristics")
		}
		cyclingCost, ok := ActivatedBodyCyclingCost(&effect.GrantedAbility)
		if !ok {
			v.add(faceName, appendPath(path, "GrantedAbility"), CardDefIssueInvalidRuleEffect, "hand-card ability grant must grant Cycling")
			return
		}
		if effect.GrantedAbility.ZoneOfFunction != zone.Hand {
			v.add(faceName, appendPath(path, "GrantedAbility.ZoneOfFunction"), CardDefIssueInvalidRuleEffect, "hand-card granted ability must function from hand")
		}
		if !reflect.DeepEqual(effect.GrantedAbility, CyclingActivatedAbility(cyclingCost)) {
			v.add(faceName, appendPath(path, "GrantedAbility"), CardDefIssueInvalidRuleEffect, "hand-card ability grant must use the standard Cycling ability template")
		}
	case RuleEffectNoMaximumHandSize:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "no-maximum-hand-size effects must set affected player")
		}
		if effect.AffectedSource || effect.AffectedAttached {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "no-maximum-hand-size effects are player-scoped and cannot affect a permanent")
		}
	default:
	}
}

func (v *cardDefValidator) validateCostModifier(faceName, path string, modifier CostModifier) {
	if modifier.GenericIncrease < 0 {
		v.add(faceName, appendPath(path, "GenericIncrease"), CardDefIssueInvalidRuleEffect, "generic cost increase cannot be negative")
	}
	if modifier.GenericReduction < 0 {
		v.add(faceName, appendPath(path, "GenericReduction"), CardDefIssueInvalidRuleEffect, "generic cost reduction cannot be negative")
	}
	if modifier.SetGeneric.Exists && modifier.SetGeneric.Val < 0 {
		v.add(faceName, appendPath(path, "SetGeneric"), CardDefIssueInvalidRuleEffect, "generic cost replacement cannot be negative")
	}
	if modifier.MinimumGeneric < 0 {
		v.add(faceName, appendPath(path, "MinimumGeneric"), CardDefIssueInvalidRuleEffect, "minimum generic cost cannot be negative")
	}
	if modifier.FirstCycleEachTurn && modifier.AbilityKeyword != Cycling {
		v.add(faceName, appendPath(path, "FirstCycleEachTurn"), CardDefIssueInvalidRuleEffect, "first-cycle cost modifiers must match Cycling")
	}
	if modifier.Kind == CostModifierAbility && modifier.AbilityKeyword == KeywordNone {
		v.add(faceName, appendPath(path, "AbilityKeyword"), CardDefIssueInvalidRuleEffect, "ability cost modifiers must set AbilityKeyword")
	}
	if modifier.SetManaCost.Exists && modifier.SetGeneric.Exists {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "cost modifier cannot set both full mana cost and generic cost")
	}
}

func handCardSelectionHasUnsupportedPredicates(selection Selection) bool {
	return selection.Controller != ControllerAny ||
		selection.Player != PlayerAny ||
		selection.Tapped != TriAny ||
		selection.CombatState != CombatStateAny ||
		selection.Keyword != KeywordNone ||
		selection.ExcludedKeyword != KeywordNone ||
		selection.Power.Exists ||
		selection.Toughness.Exists ||
		selection.ExcludeSource ||
		selection.NonToken ||
		selection.TokenOnly
}

// validateGroupRef validates the structural consistency of a GroupReference and
// checks contextual target-slot bounds for its anchor and exclusion. Structural
// issues are reported only once: group.Validate() handles the nested references,
// and validateObjectRefBounds adds bounds without re-reporting structure.
func (v *cardDefValidator) validateGroupRef(faceName, path string, group GroupReference, targets []TargetSpec) {
	for _, problem := range group.Validate() {
		v.add(faceName, path, CardDefIssueInvalidReference, problem)
	}
	if anchor, ok := group.Anchor(); ok {
		v.validateObjectRefBounds(faceName, appendPath(path, "Anchor"), anchor, targets)
	}
	if exclude, ok := group.Exclusion(); ok {
		v.validateObjectRefBounds(faceName, appendPath(path, "Exclusion"), exclude, targets)
	}
}

func (v *cardDefValidator) validateNestedCard(faceName, path string, card *CardDef) {
	if card == nil {
		return
	}
	v.validateFace(faceName, path, &card.CardFace)
	if card.Back.Exists {
		face := card.Back.Val
		v.validateFace(faceName, appendPath(path, "Back"), &face)
	}
}

func (v *cardDefValidator) validateTargetIndex(faceName, path string, targetIndex int, targets []TargetSpec, label string) {
	// Negative target indexes are reserved for rules-owned internal bindings.
	if targetIndex < 0 {
		return
	}
	// Object references address chosen targets by a flat slot index across all
	// specs, so a single multi-target spec admits MaxTargets consecutive slots.
	if targetIndex >= targetSlotCapacity(targets) {
		v.add(faceName, path, CardDefIssueTargetIndexOutOfRange, fmt.Sprintf("%s index %d has no matching TargetSpec", label, targetIndex))
	}
}

func (v *cardDefValidator) validateCondition(faceName, path string, condition *Condition, targets []TargetSpec) {
	if condition.ControllerLifeAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerLifeAtLeast"), CardDefIssueInvalidCondition, "life threshold cannot be negative")
	}
	if condition.ControllerHandSizeAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerHandSizeAtLeast"), CardDefIssueInvalidCondition, "hand-size threshold cannot be negative")
	}
	if condition.AnyPlayerLifeAtMost < 0 {
		v.add(faceName, appendPath(path, "AnyPlayerLifeAtMost"), CardDefIssueInvalidCondition, "life threshold cannot be negative")
	}
	if condition.OpponentCountAtLeast < 0 {
		v.add(faceName, appendPath(path, "OpponentCountAtLeast"), CardDefIssueInvalidCondition, "opponent-count threshold cannot be negative")
	}
	if condition.ControllerGraveyardCardCountAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerGraveyardCardCountAtLeast"), CardDefIssueInvalidCondition, "graveyard-card threshold cannot be negative")
	}
	if condition.ControllerGraveyardCardTypeCountAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerGraveyardCardTypeCountAtLeast"), CardDefIssueInvalidCondition, "graveyard-card-type threshold cannot be negative")
	}
	if condition.ControllerBasicLandTypeCountAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerBasicLandTypeCountAtLeast"), CardDefIssueInvalidCondition, "basic-land-type threshold cannot be negative")
	}
	if condition.ControllerCreaturePowerDiversityAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerCreaturePowerDiversityAtLeast"), CardDefIssueInvalidCondition, "creature-power-diversity threshold cannot be negative")
	}
	if condition.ControllerControls.MinCount < 0 {
		v.add(faceName, appendPath(path, "ControllerControls.MinCount"), CardDefIssueInvalidCondition, "permanent-count threshold cannot be negative")
	}
	if !condition.ControllerControls.Empty() {
		v.validateSelection(faceName, appendPath(path, "ControllerControls"), condition.ControllerControls.Selection())
	}
	if condition.ControlsMatching.Exists {
		v.validateConditionSelectionCount(faceName, appendPath(path, "ControlsMatching"), condition.ControlsMatching.Val)
		if !condition.ControllerControls.Empty() {
			v.add(faceName, path, CardDefIssueInvalidSelection, "Condition sets both ControllerControls and ControlsMatching")
		}
	}
	if condition.AnyOpponentControls.Exists {
		v.validateConditionSelectionCount(faceName, appendPath(path, "AnyOpponentControls"), condition.AnyOpponentControls.Val)
	}
	if condition.OpponentsControl.Exists {
		v.validateConditionSelectionCount(faceName, appendPath(path, "OpponentsControl"), condition.OpponentsControl.Val)
	}
	if condition.Object.Exists {
		v.validateObjectRef(faceName, appendPath(path, "Object"), condition.Object.Val, targets)
	}
	if condition.ObjectMatches.Exists {
		v.validateSelection(faceName, appendPath(path, "ObjectMatches"), condition.ObjectMatches.Val)
		if !condition.Object.Exists {
			v.add(faceName, appendPath(path, "ObjectMatches"), CardDefIssueInvalidCondition, "ObjectMatches requires an Object reference")
		}
		if len(condition.Types) > 0 {
			v.add(faceName, path, CardDefIssueInvalidSelection, "Condition sets both legacy Types and ObjectMatches")
		}
		if condition.ObjectMatches.Val.Player != PlayerAny {
			v.add(faceName, appendPath(path, "ObjectMatches.Player"), CardDefIssueInvalidSelection, "object Selection cannot use a player relation")
		}
	}
	if condition.EventHistory.Exists {
		v.validateEventHistoryCondition(faceName, appendPath(path, "EventHistory"), &condition.EventHistory.Val)
	}
}

func (v *cardDefValidator) validateConditionSelectionCount(faceName, path string, count SelectionCount) {
	if count.MinCount < 0 {
		v.add(faceName, appendPath(path, "MinCount"), CardDefIssueInvalidCondition, "permanent-count threshold cannot be negative")
	}
	selection := count.Selection
	v.validateSelection(faceName, appendPath(path, "Selection"), selection)
	if selection.Player != PlayerAny {
		v.add(faceName, appendPath(path, "Selection.Player"), CardDefIssueInvalidSelection, "controlled-permanent Selection cannot use a player relation")
	}
}

func (v *cardDefValidator) validateEventHistoryCondition(faceName, path string, hist *EventHistoryCondition) {
	if hist.Pattern.Event == EventUnknown {
		v.add(faceName, appendPath(path, "Pattern.Event"), CardDefIssueInvalidCondition, "EventHistoryCondition Pattern.Event must not be EventUnknown")
	}
	if !hist.Pattern.SubjectSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "Pattern.SubjectSelection"), hist.Pattern.SubjectSelection)
	}
}

func (v *cardDefValidator) validateReplacementEffect(faceName, path string, replacement *ReplacementEffect) {
	if replacement.Condition.Exists {
		condition := replacement.Condition.Val
		v.validateCondition(faceName, appendPath(path, "Condition"), &condition, nil)
	}
}

func (v *cardDefValidator) validateTriggerPattern(faceName, path string, pattern *TriggerPattern) {
	if !pattern.SubjectSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "SubjectSelection"), pattern.SubjectSelection)
		unsupported := pattern.SubjectSelection
		unsupported.RequiredTypes = nil
		unsupported.RequiredTypesAny = nil
		unsupported.ExcludedTypes = nil
		unsupported.Supertypes = nil
		unsupported.SubtypesAny = nil
		unsupported.ColorsAny = nil
		unsupported.ExcludedColors = nil
		unsupported.Colorless = false
		unsupported.Multicolored = false
		unsupported.Controller = ControllerAny
		unsupported.Tapped = TriAny
		unsupported.CombatState = CombatStateAny
		unsupported.Keyword = KeywordNone
		unsupported.ExcludedKeyword = KeywordNone
		unsupported.ManaValue.Exists = false
		unsupported.Power.Exists = false
		unsupported.Toughness.Exists = false
		unsupported.NonToken = false
		unsupported.TokenOnly = false
		if !unsupported.Empty() {
			v.add(faceName, appendPath(path, "SubjectSelection"), CardDefIssueInvalidSelection, "trigger subject Selection uses predicates unavailable from event data")
		}
		if len(pattern.RequirePermanentTypes) > 0 || len(pattern.ExcludePermanentTypes) > 0 || pattern.RequireNonToken {
			v.add(faceName, path, CardDefIssueInvalidSelection, "TriggerPattern sets both permanent-type filters and SubjectSelection")
		}
	}
	if pattern.SubjectSelectionOrSelf {
		v.validateSubjectSelectionOrSelf(faceName, path, pattern)
	}
	if !pattern.RelatedSubjectSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "RelatedSubjectSelection"), pattern.RelatedSubjectSelection)
	}
	if !pattern.CardSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "CardSelection"), pattern.CardSelection)
		unsupported := pattern.CardSelection
		unsupported.RequiredTypes = nil
		unsupported.RequiredTypesAny = nil
		unsupported.ExcludedTypes = nil
		if pattern.Event == EventSpellCast {
			unsupported.Supertypes = nil
			unsupported.SubtypesAny = nil
			unsupported.ColorsAny = nil
			unsupported.Colorless = false
			unsupported.Multicolored = false
			unsupported.ManaValue.Exists = false
		}
		if !unsupported.Empty() {
			v.add(faceName, appendPath(path, "CardSelection"), CardDefIssueInvalidSelection, "trigger card Selection uses predicates unavailable from event data")
		}
		if len(pattern.RequireCardTypes) > 0 || len(pattern.ExcludeCardTypes) > 0 {
			v.add(faceName, path, CardDefIssueInvalidSelection, "TriggerPattern sets both card-type filters and CardSelection")
		}
	}
	if !pattern.DamageRecipientSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "DamageRecipientSelection"), pattern.DamageRecipientSelection)
		if len(pattern.DamageRecipientTypes) > 0 {
			v.add(faceName, path, CardDefIssueInvalidSelection, "TriggerPattern sets both damage-recipient type filters and DamageRecipientSelection")
		}
	}
	if !pattern.DamageSourceSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "DamageSourceSelection"), pattern.DamageSourceSelection)
	}
	if pattern.DamageRecipientIsSource && pattern.DamageRecipient&DamageRecipientPermanent == 0 {
		v.add(faceName, path, CardDefIssueInvalidSelection, "DamageRecipientIsSource requires a permanent damage recipient")
	}
	if !pattern.AttackRecipientSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "AttackRecipientSelection"), pattern.AttackRecipientSelection)
	}
	if pattern.RequireCombatDamage && pattern.RequireNonCombatDamage {
		v.add(faceName, path, CardDefIssueInvalidSelection, "trigger pattern cannot require both combat and noncombat damage")
	}
	if pattern.OneOrMorePerAttackTarget && (!pattern.OneOrMore || pattern.Event != EventAttackerDeclared) {
		v.add(faceName, path, CardDefIssueInvalidSelection, "OneOrMorePerAttackTarget requires a one-or-more attacker-declared pattern")
	}
	v.validateAttackerCountRelations(faceName, path, pattern)
	if !pattern.StepPlayerSourceAttachedSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "StepPlayerSourceAttachedSelection"), pattern.StepPlayerSourceAttachedSelection)
		if pattern.Event != EventBeginningOfStep {
			v.add(faceName, path, CardDefIssueInvalidSelection, "StepPlayerSourceAttachedSelection requires a beginning-of-step pattern")
		}
	}
	if pattern.RequireKickerPaid && pattern.Event != EventSpellCast {
		v.add(faceName, appendPath(path, "RequireKickerPaid"), CardDefIssueInvalidSelection, "kicker-paid trigger filter is only supported for spell-cast events")
	}
	if pattern.RequireHistoric && pattern.Event != EventSpellCast {
		v.add(faceName, appendPath(path, "RequireHistoric"), CardDefIssueInvalidSelection, "historic trigger filter is only supported for spell-cast events")
	}
	if pattern.MatchSpellCopy && pattern.Event != EventSpellCast {
		v.add(faceName, appendPath(path, "MatchSpellCopy"), CardDefIssueInvalidSelection, "spell-copy matching is only supported for spell-cast events")
	}
	if pattern.ExcludeManaAbility && pattern.Event != EventAbilityActivated {
		v.add(faceName, appendPath(path, "ExcludeManaAbility"), CardDefIssueInvalidSelection, "mana-ability exclusion is only supported for ability-activated events")
	}
	if pattern.Event == EventAbilityActivated && !pattern.ExcludeManaAbility {
		v.add(faceName, appendPath(path, "ExcludeManaAbility"), CardDefIssueInvalidSelection, "unrestricted ability-activated triggers are unavailable because the runtime event stream omits payment-time mana abilities")
	}
	if pattern.PlayerEventOrdinalThisTurn < 0 {
		v.add(faceName, appendPath(path, "PlayerEventOrdinalThisTurn"), CardDefIssueInvalidSelection, "player-event ordinal cannot be negative")
	}
	if pattern.PlayerEventOrdinalThisTurn > 0 &&
		pattern.Event != EventCardDrawn &&
		pattern.Event != EventLifeGained &&
		pattern.Event != EventLifeLost &&
		pattern.Event != EventScry &&
		pattern.Event != EventSurveil &&
		pattern.Event != EventSpellCast {
		v.add(faceName, appendPath(path, "PlayerEventOrdinalThisTurn"), CardDefIssueInvalidSelection, "player-event ordinal is unavailable for this event")
	}
	if pattern.MatchFromZone && pattern.FromZone == zone.None {
		v.add(faceName, appendPath(path, "FromZone"), CardDefIssueInvalidSelection, "from-zone trigger filter must set a source zone")
	}
	if pattern.MatchToZone && pattern.ToZone == zone.None {
		v.add(faceName, appendPath(path, "ToZone"), CardDefIssueInvalidSelection, "to-zone trigger filter must set a destination zone")
	}
	if pattern.ExcludeToZone && pattern.ToZone == zone.None {
		v.add(faceName, appendPath(path, "ToZone"), CardDefIssueInvalidSelection, "excluded to-zone trigger filter must set a destination zone")
	}
	if pattern.MatchToZone && pattern.ExcludeToZone {
		v.add(faceName, appendPath(path, "ToZone"), CardDefIssueInvalidSelection, "to-zone trigger filter cannot both require and exclude its destination")
	}
	if pattern.FaceDown && !pattern.MatchFaceDown {
		v.add(faceName, appendPath(path, "FaceDown"), CardDefIssueInvalidSelection, "face-down trigger filter must be enabled")
	}
}

// validateAttackerCountRelations checks the attacker-count combat relations.
// AttackAlone only applies to attacker-declared events; AttackerCountAtLeast
// must require at least two attackers via a one-or-more attacker-declared
// pattern that is not also attacks-alone.
func (v *cardDefValidator) validateAttackerCountRelations(faceName, path string, pattern *TriggerPattern) {
	if pattern.AttackAlone && pattern.Event != EventAttackerDeclared {
		v.add(faceName, appendPath(path, "AttackAlone"), CardDefIssueInvalidSelection, "attacks-alone trigger filter is only supported for attacker-declared events")
	}
	if pattern.AttackerCountAtLeast == 0 {
		return
	}
	if pattern.AttackerCountAtLeast < 2 {
		v.add(faceName, appendPath(path, "AttackerCountAtLeast"), CardDefIssueInvalidSelection, "attacker-count trigger filter must require at least two attackers")
	}
	if pattern.Event != EventAttackerDeclared || !pattern.OneOrMore || pattern.AttackAlone {
		v.add(faceName, appendPath(path, "AttackerCountAtLeast"), CardDefIssueInvalidSelection, "attacker-count trigger filter requires a one-or-more attacker-declared pattern without attacks-alone")
	}
}

func (v *cardDefValidator) validateSubjectSelectionOrSelf(faceName, path string, pattern *TriggerPattern) {
	subPath := appendPath(path, "SubjectSelectionOrSelf")
	if pattern.SubjectSelection.Empty() {
		v.add(faceName, subPath, CardDefIssueInvalidSelection, "SubjectSelectionOrSelf requires a SubjectSelection")
	}
	if pattern.Source != TriggerSourceAny {
		v.add(faceName, subPath, CardDefIssueInvalidSelection, "SubjectSelectionOrSelf cannot combine with a source filter")
	}
	if pattern.ExcludeSelf {
		v.add(faceName, subPath, CardDefIssueInvalidSelection, "SubjectSelectionOrSelf cannot combine with ExcludeSelf")
	}
	switch pattern.Event {
	case EventPermanentEnteredBattlefield, EventPermanentDied, EventZoneChanged:
	default:
		v.add(faceName, subPath, CardDefIssueInvalidSelection, "SubjectSelectionOrSelf is only supported for permanent zone-change events")
	}
}

func (v *cardDefValidator) validateObjectRef(faceName, path string, ref ObjectReference, targets []TargetSpec) {
	for _, problem := range ref.Validate() {
		v.add(faceName, path, CardDefIssueInvalidReference, problem)
	}
	v.validateObjectRefBounds(faceName, path, ref, targets)
}

// validateObjectRefBounds checks only the contextual target-slot bounds for an
// object reference. Structural consistency is reported by validateObjectRef so
// that nested references are not diagnosed twice.
func (v *cardDefValidator) validateObjectRefBounds(faceName, path string, ref ObjectReference, targets []TargetSpec) {
	switch ref.Kind() {
	case ObjectReferenceTargetPermanent, ObjectReferenceTargetStackObject:
		v.validateTargetIndex(faceName, path, ref.TargetIndex(), targets, "object reference target")
	case ObjectReferenceTargetAttachedPermanent:
		v.validateTargetIndex(faceName, path, ref.TargetIndex(), targets, "attached permanent reference target")
	default:
	}
}

func (v *cardDefValidator) validatePlayerRef(faceName, path string, ref PlayerReference, targets []TargetSpec) {
	for _, problem := range ref.Validate() {
		v.add(faceName, path, CardDefIssueInvalidReference, problem)
	}
	switch ref.Kind() {
	case PlayerReferenceTargetPlayer:
		v.validateTargetIndex(faceName, path, ref.TargetIndex(), targets, "player reference target")
	case PlayerReferenceObjectController, PlayerReferenceObjectOwner:
		if object, ok := ref.Object(); ok {
			v.validateObjectRefBounds(faceName, appendPath(path, "Object"), object, targets)
		}
	default:
	}
}

func (v *cardDefValidator) validateCardCondition(faceName, path string, condition CardCondition) {
	v.validateCardRef(faceName, appendPath(path, "Card"), condition.Card)
	if !condition.RequirePermanentCard && len(condition.Types) == 0 && len(condition.Supertypes) == 0 && len(condition.SubtypesAny) == 0 {
		v.add(faceName, path, CardDefIssueInvalidReference, "card condition has no filters")
	}
}

func (v *cardDefValidator) validateCardRef(faceName, path string, ref CardReference) bool {
	switch ref.Kind {
	case CardReferenceLinked:
		if ref.LinkID == "" {
			v.add(faceName, path, CardDefIssueInvalidReference, "linked card reference requires LinkID")
			return false
		}
	case CardReferenceSource, CardReferenceEvent, CardReferenceTarget:
		if ref.LinkID != "" {
			v.add(faceName, path, CardDefIssueInvalidReference, "source/event/target card reference must not set LinkID")
			return false
		}
		if ref.Kind != CardReferenceTarget && ref.TargetIndex != 0 {
			v.add(faceName, path, CardDefIssueInvalidReference, "source/event card reference must not set TargetIndex")
			return false
		}
		if ref.TargetIndex < 0 {
			v.add(faceName, path, CardDefIssueInvalidReference, "target card reference must not use a negative TargetIndex")
			return false
		}
	case CardReferenceNone:
		v.add(faceName, path, CardDefIssueInvalidReference, "card reference has no kind")
		return false
	default:
		v.add(faceName, path, CardDefIssueInvalidReference, fmt.Sprintf("unknown card reference kind %d", ref.Kind))
		return false
	}
	return true
}

func (v *cardDefValidator) validateTokenCopySpec(faceName, path string, spec TokenCopySpec, targets []TargetSpec) {
	switch spec.Source {
	case TokenCopySourceObject:
		v.validateObjectRef(faceName, appendPath(path, "Object"), spec.Object, targets)
	case TokenCopySourceSourceCard:
	case TokenCopySourceNone:
		v.add(faceName, appendPath(path, "Source"), CardDefIssueInvalidReference, "token copy source has no kind")
	default:
		v.add(faceName, appendPath(path, "Source"), CardDefIssueInvalidReference, fmt.Sprintf("unknown token copy source %d", spec.Source))
	}
}

func (v *cardDefValidator) add(faceName, path string, code CardDefIssueCode, message string) {
	v.issues = append(v.issues, CardDefIssue{
		FaceName: faceName,
		Path:     path,
		Code:     code,
		Message:  message,
	})
}

func appendPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

// isKnownProtectionSubtype reports whether sub is a creature or land subtype
// that the renderer can emit (via types.KnownSubtypeForType).
func isKnownProtectionSubtype(sub types.Sub) bool {
	return types.KnownSubtypeForType(types.Creature, sub) ||
		types.KnownSubtypeForType(types.Land, sub)
}

// isKnownProtectionCardType reports whether t is a card type the renderer can
// serialise. Mirrors the set supported by cardgen.cardTypeLiteral.
func isKnownProtectionCardType(t types.Card) bool {
	switch t {
	case types.Land, types.Creature, types.Artifact, types.Enchantment,
		types.Instant, types.Sorcery, types.Planeswalker, types.Battle,
		types.Kindred, types.Plane, types.Dungeon, types.Phenomenon,
		types.Scheme, types.Vanguard, types.Conspiracy:
		return true
	default:
		return false
	}
}

// isKnownProtectionColor reports whether c is one of the five Magic colors.
func isKnownProtectionColor(c color.Color) bool {
	switch c {
	case color.White, color.Blue, color.Black, color.Red, color.Green:
		return true
	default:
		return false
	}
}
