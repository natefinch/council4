package game

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game/cost"
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
		len(face.ActivatedAbilities) > 0 ||
		len(face.ManaAbilities) > 0 ||
		len(face.LoyaltyAbilities) > 0 ||
		len(face.TriggeredAbilities) > 0 ||
		len(face.ReplacementAbilities) > 0 ||
		len(face.StaticAbilities) > 0
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
		v.validateInstructionSequence(faceName, appendPath(modePath, "Sequence"), mode.Sequence, targets)
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
		if len(keyword.FromColors) == 0 {
			v.add(faceName, appendPath(path, "FromColors"), CardDefIssueInvalidKeywordAbility, "protection needs at least one protected color")
		}
	case nil:
		v.add(faceName, path, CardDefIssueInvalidKeywordAbility, "keyword ability is nil")
	default:
		v.add(faceName, path, CardDefIssueInvalidKeywordAbility, fmt.Sprintf("unknown keyword ability %T", ability))
	}
}

func (v *cardDefValidator) validateInstructionSequence(faceName, path string, seq []Instruction, targets ...[]TargetSpec) {
	if err := ValidateInstructionSequence(seq, targets...); err != nil {
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, err.Error())
	}
}

func (v *cardDefValidator) validateManaKeywordCost(faceName, path string, manaCost cost.Mana) {
	if len(manaCost) == 0 {
		v.add(faceName, appendPath(path, "Cost"), CardDefIssueInvalidKeywordAbility, "mana-valued keyword cost must be explicit")
	}
}

func (v *cardDefValidator) validateTargetSpec(faceName, path string, target *TargetSpec) {
	if target.MinTargets < 0 || target.MaxTargets < 0 {
		v.add(faceName, path, CardDefIssueInvalidTargetSpec, "target counts must be non-negative")
		return
	}
	if target.MaxTargets < target.MinTargets {
		v.add(faceName, path, CardDefIssueInvalidTargetSpec, "max targets is less than min targets")
	}
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
		if allowsPlayers && selectionHasPermanentPredicates(selection) {
			v.add(faceName, appendPath(path, "Selection"), CardDefIssueInvalidSelection, "player targets cannot use permanent Selection predicates")
		}
		if !allowsPlayers && selection.Player != PlayerAny {
			v.add(faceName, appendPath(path, "Selection.Player"), CardDefIssueInvalidSelection, "non-player targets cannot use a player relation")
		}
		if !allowsPermanents && !allowsPlayers && !selection.Empty() {
			v.add(faceName, appendPath(path, "Selection"), CardDefIssueInvalidSelection, "Selection requires permanent or player targets")
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
		len(selection.SubtypesAny) > 0 ||
		len(selection.ColorsAny) > 0 ||
		len(selection.ExcludedColors) > 0 ||
		selection.Controller != ControllerAny ||
		selection.Tapped != TriAny ||
		selection.CombatState != CombatStateAny ||
		selection.Keyword != KeywordNone ||
		selection.ExcludedKeyword != KeywordNone ||
		selection.ManaValue.Exists ||
		selection.Power.Exists ||
		selection.Toughness.Exists ||
		selection.ExcludeSource ||
		selection.NonToken
}

func (v *cardDefValidator) validateContinuousEffect(faceName, path string, continuous *ContinuousEffect, targets []TargetSpec) {
	for i := range continuous.AddAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("AddAbilities[%d]", i)), continuous.AddAbilities[i], nil)
	}
	if continuous.Group.Valid() {
		v.validateGroupRef(faceName, appendPath(path, "Group"), continuous.Group, targets)
	}
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
	if targetIndex >= len(targets) {
		v.add(faceName, path, CardDefIssueTargetIndexOutOfRange, fmt.Sprintf("%s index %d has no matching TargetSpec", label, targetIndex))
	}
}

func (v *cardDefValidator) validateCondition(faceName, path string, condition *Condition, targets []TargetSpec) {
	if condition.ControlsMatching.Exists {
		selection := condition.ControlsMatching.Val.Selection
		v.validateSelection(faceName, appendPath(path, "ControlsMatching.Selection"), selection)
		if selection.Player != PlayerAny {
			v.add(faceName, appendPath(path, "ControlsMatching.Selection.Player"), CardDefIssueInvalidSelection, "controlled-permanent Selection cannot use a player relation")
		}
		if !condition.ControllerControls.Empty() {
			v.add(faceName, path, CardDefIssueInvalidSelection, "Condition sets both ControllerControls and ControlsMatching")
		}
	}
	if condition.Object.Exists {
		v.validateObjectRef(faceName, appendPath(path, "Object"), condition.Object.Val, targets)
	}
}

func (v *cardDefValidator) validateTriggerPattern(faceName, path string, pattern *TriggerPattern) {
	if !pattern.SubjectSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "SubjectSelection"), pattern.SubjectSelection)
		unsupported := pattern.SubjectSelection
		unsupported.RequiredTypes = nil
		unsupported.RequiredTypesAny = nil
		unsupported.ExcludedTypes = nil
		unsupported.Controller = ControllerAny
		unsupported.NonToken = false
		if !unsupported.Empty() {
			v.add(faceName, appendPath(path, "SubjectSelection"), CardDefIssueInvalidSelection, "trigger subject Selection uses predicates unavailable from event data")
		}
		if len(pattern.RequirePermanentTypes) > 0 || len(pattern.ExcludePermanentTypes) > 0 || pattern.RequireNonToken {
			v.add(faceName, path, CardDefIssueInvalidSelection, "TriggerPattern sets both permanent-type filters and SubjectSelection")
		}
	}
	if !pattern.CardSelection.Empty() {
		v.validateSelection(faceName, appendPath(path, "CardSelection"), pattern.CardSelection)
		unsupported := pattern.CardSelection
		unsupported.RequiredTypes = nil
		unsupported.RequiredTypesAny = nil
		unsupported.ExcludedTypes = nil
		if !unsupported.Empty() {
			v.add(faceName, appendPath(path, "CardSelection"), CardDefIssueInvalidSelection, "trigger card Selection supports only card-type predicates")
		}
		if len(pattern.RequireCardTypes) > 0 || len(pattern.ExcludeCardTypes) > 0 {
			v.add(faceName, path, CardDefIssueInvalidSelection, "TriggerPattern sets both card-type filters and CardSelection")
		}
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
	case CardReferenceSource, CardReferenceEvent:
		if ref.LinkID != "" {
			v.add(faceName, path, CardDefIssueInvalidReference, "source/event card reference must not set LinkID")
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
