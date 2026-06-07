package cardgen

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

// ValidationCode identifies a class of card-definition validation issue.
type ValidationCode string

// Validation issue codes identify generated-card validation failures.
const (
	IssueNilCard                    ValidationCode = "nil-card"
	IssueMissingName                ValidationCode = "missing-name"
	IssueOracleWithoutAbilities     ValidationCode = "oracle-without-abilities"
	IssueTargetIndexOutOfRange      ValidationCode = "target-index-out-of-range"
	IssueInvalidReference           ValidationCode = "invalid-reference"
	IssueInvalidTargetSpec          ValidationCode = "invalid-target-spec"
	IssueInvalidKeywordAbility      ValidationCode = "invalid-keyword-ability"
	IssueInvalidAbilityBody         ValidationCode = "invalid-ability-body"
	IssueUnregisteredImplementation ValidationCode = "unregistered-implementation"
	IssueImplementationRequired     ValidationCode = "implementation-required"
	IssueGeneratedCardNotFound      ValidationCode = "generated-card-not-found"
	IssueValidationRunFailed        ValidationCode = "validation-run-failed"
)

// ValidationIssue describes one problem found in a generated card definition.
type ValidationIssue struct {
	CardName string         `json:"card_name"`
	FaceName string         `json:"face_name,omitempty"`
	Path     string         `json:"path,omitempty"`
	Code     ValidationCode `json:"code"`
	Message  string         `json:"message"`
}

// ValidationOptions configures generated-card validation.
type ValidationOptions struct {
	// KnownImplementationIDs is the optional set of hand-written implementation
	// IDs registered by the runtime. When non-empty, any card or face
	// ImplementationID outside this set is reported.
	KnownImplementationIDs map[string]bool

	// ReportImplementationIDs reports every hand-written ImplementationID as an
	// unsupported-card issue. Batch rollout uses this because ImplementationID is
	// an escape hatch for behavior that is not represented by generated data.
	ReportImplementationIDs bool
}

// ValidateCards validates a collection of generated CardDef values.
func ValidateCards(cards []*game.CardDef, opts ValidationOptions) []ValidationIssue {
	var issues []ValidationIssue
	for _, card := range cards {
		issues = append(issues, ValidateCard(card, opts)...)
	}
	return issues
}

// ValidateCard validates one generated CardDef against rules support that can
// be checked statically.
func ValidateCard(card *game.CardDef, opts ValidationOptions) []ValidationIssue {
	v := cardValidator{card: card, opts: opts}
	v.validate()
	return v.issues
}

type cardValidator struct {
	card   *game.CardDef
	opts   ValidationOptions
	issues []ValidationIssue
}

func (v *cardValidator) validate() {
	if v.card == nil {
		v.add("", "", IssueNilCard, "card definition is nil")
		return
	}
	if strings.TrimSpace(v.card.Name) == "" {
		v.add("", "", IssueMissingName, "card definition has no name")
	}
	v.validateFace(v.card.Name, "", &v.card.CardFace, true)
	if v.card.Back.Exists {
		face := v.card.Back.Val
		name := face.Name
		if strings.TrimSpace(name) == "" {
			name = "back face"
		}
		v.validateFace(name, "Back", &face, true)
	}
	if v.card.Alternate.Exists {
		face := v.card.Alternate.Val
		name := face.Name
		if strings.TrimSpace(name) == "" {
			name = "alternate face"
		}
		v.validateFace(name, "Alternate", &face, true)
	}
}

func (v *cardValidator) validateFace(faceName, path string, face *game.CardFace, walkAbilities bool) {
	hasAbilities := face.SpellAbility.Exists ||
		len(face.ActivatedAbilities) > 0 ||
		len(face.ManaAbilities) > 0 ||
		len(face.LoyaltyAbilities) > 0 ||
		len(face.TriggeredAbilities) > 0 ||
		len(face.ReplacementAbilities) > 0 ||
		len(face.StaticAbilities) > 0
	if strings.TrimSpace(face.OracleText) != "" && !hasAbilities && face.ImplementationID == "" {
		v.add(faceName, path, IssueOracleWithoutAbilities, "oracle text is non-empty but no abilities or hand-written implementation are defined")
	}
	if face.ImplementationID != "" && len(v.opts.KnownImplementationIDs) > 0 && !v.opts.KnownImplementationIDs[face.ImplementationID] {
		v.add(faceName, path, IssueUnregisteredImplementation, fmt.Sprintf("implementation ID %q is not registered", face.ImplementationID))
	}
	if face.ImplementationID != "" && v.opts.ReportImplementationIDs {
		v.add(faceName, path, IssueImplementationRequired, fmt.Sprintf("implementation ID %q requires hand-written rules support", face.ImplementationID))
	}
	if !walkAbilities {
		return
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

func (v *cardValidator) validateAbilityBody(faceName, path string, body game.AbilityBody, targets []game.TargetSpec) {
	switch abilityBody := body.(type) {
	case game.SpellAbilityBody:
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case game.ActivatedAbilityBody:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case game.ManaAbilityBody:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		if abilityBody.Content != nil {
			v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
		}
	case game.LoyaltyAbilityBody:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case game.TriggeredAbilityBody:
		if abilityBody.Trigger.InterveningCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "Trigger.InterveningCondition"), &abilityBody.Trigger.InterveningCondition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case game.StaticAbilityBody:
		if abilityBody.Condition.Exists {
			v.validateCondition(faceName, appendPath(path, "Condition"), &abilityBody.Condition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		for i := range abilityBody.ContinuousEffects {
			v.validateContinuousEffect(faceName, appendPath(path, fmt.Sprintf("ContinuousEffects[%d]", i)), &abilityBody.ContinuousEffects[i])
		}
	case nil:
		v.add(faceName, path, IssueInvalidAbilityBody, "ability body is nil")
	default:
		v.add(faceName, path, IssueInvalidAbilityBody, fmt.Sprintf("unknown ability body %T", body))
	}
}

func (v *cardValidator) validateReplacementAbility(faceName, path string, ability *game.ReplacementAbilityBody) {
	if ability == nil {
		v.add(faceName, path, IssueInvalidAbilityBody, "replacement ability is nil")
		return
	}
	if ability.Replacement.Condition.Exists {
		v.validateCondition(faceName, appendPath(path, "Replacement.Condition"), &ability.Replacement.Condition.Val, nil)
	}
}

func (v *cardValidator) validateAbilityContent(faceName, path string, content game.AbilityContent, fallbackTargets []game.TargetSpec) {
	switch abilityContent := content.(type) {
	case game.PlainAbilityContent:
		for i := range abilityContent.Targets {
			v.validateTargetSpec(faceName, appendPath(path, fmt.Sprintf("Targets[%d]", i)), &abilityContent.Targets[i])
		}
		targets := abilityContent.Targets
		if len(targets) == 0 {
			targets = fallbackTargets
		}
		v.validateInstructionSequence(faceName, appendPath(path, "Sequence"), abilityContent.Sequence, targets)
	case game.ModalAbilityContent:
		for i := range abilityContent.SharedTargets {
			v.validateTargetSpec(faceName, appendPath(path, fmt.Sprintf("SharedTargets[%d]", i)), &abilityContent.SharedTargets[i])
		}
		for i := range abilityContent.Modes {
			mode := &abilityContent.Modes[i]
			modePath := appendPath(path, fmt.Sprintf("Modes[%d]", i))
			for j := range mode.Targets {
				v.validateTargetSpec(faceName, appendPath(modePath, fmt.Sprintf("Targets[%d]", j)), &mode.Targets[j])
			}
			targets := mode.Targets
			if len(targets) == 0 {
				targets = abilityContent.SharedTargets
			}
			v.validateInstructionSequence(faceName, appendPath(modePath, "Sequence"), mode.Sequence, targets)
		}
	case nil:
		v.add(faceName, path, IssueInvalidAbilityBody, "ability content is nil")
	default:
		v.add(faceName, path, IssueInvalidAbilityBody, fmt.Sprintf("unknown ability content %T", content))
	}
}

func (v *cardValidator) validateKeywordAbility(faceName, path string, ability game.KeywordAbility, targets []game.TargetSpec) {
	switch keyword := ability.(type) {
	case game.SimpleKeyword:
		if keyword.Kind == game.KeywordNone {
			v.add(faceName, path, IssueInvalidKeywordAbility, "simple keyword must set Kind")
		}
	case game.WardKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case game.EquipKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case game.EnchantKeyword:
		v.validateTargetSpec(faceName, appendPath(path, "Target"), &keyword.Target)
	case game.CyclingKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case game.KickerKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
		if keyword.BonusContent != nil {
			v.validateAbilityContent(faceName, appendPath(path, "BonusContent"), keyword.BonusContent, targets)
		}
	case game.MadnessKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case game.MorphKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case game.DisguiseKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case game.SuspendKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
		if keyword.TimeCounters <= 0 {
			v.add(faceName, appendPath(path, "TimeCounters"), IssueInvalidKeywordAbility, "suspend time counters must be positive")
		}
	case game.ProtectionKeyword:
		if len(keyword.FromColors) == 0 {
			v.add(faceName, appendPath(path, "FromColors"), IssueInvalidKeywordAbility, "protection needs at least one protected color")
		}
	case nil:
		v.add(faceName, path, IssueInvalidKeywordAbility, "keyword ability is nil")
	default:
		v.add(faceName, path, IssueInvalidKeywordAbility, fmt.Sprintf("unknown keyword ability %T", ability))
	}
}

func (v *cardValidator) validateInstructionSequence(faceName, path string, seq []game.Instruction, targets ...[]game.TargetSpec) {
	if err := game.ValidateInstructionSequence(seq, targets...); err != nil {
		v.add(faceName, path, IssueInvalidAbilityBody, err.Error())
	}
}

func (v *cardValidator) validateManaKeywordCost(faceName, path string, manaCost cost.Mana) {
	if len(manaCost) == 0 {
		v.add(faceName, appendPath(path, "Cost"), IssueInvalidKeywordAbility, "mana-valued keyword cost must be explicit")
	}
}

func (v *cardValidator) validateTargetSpec(faceName, path string, target *game.TargetSpec) {
	if target.MinTargets < 0 || target.MaxTargets < 0 {
		v.add(faceName, path, IssueInvalidTargetSpec, "target counts must be non-negative")
		return
	}
	if target.MaxTargets < target.MinTargets {
		v.add(faceName, path, IssueInvalidTargetSpec, "max targets is less than min targets")
	}
	switch target.Chooser {
	case game.TargetChooserController:
	case game.TargetChooserOpponent:
		if target.MinTargets != 1 || target.MaxTargets != 1 {
			v.add(faceName, path, IssueInvalidTargetSpec, "non-controller target chooser requires exactly one target")
		}
		if target.Predicate.Controller != game.ControllerAny && target.Predicate.Controller != game.ControllerYou {
			v.add(faceName, appendPath(path, "Predicate.Controller"), IssueInvalidTargetSpec, "opponent target chooser only supports controller-any or controller-you predicates")
		}
	default:
		v.add(faceName, appendPath(path, "Chooser"), IssueInvalidTargetSpec, "unknown target chooser")
	}
}

func (v *cardValidator) validateContinuousEffect(faceName, path string, continuous *game.ContinuousEffect) {
	for i := range continuous.AddAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("AddAbilities[%d]", i)), continuous.AddAbilities[i], nil)
	}
}

func (v *cardValidator) validateNestedCard(faceName, path string, card *game.CardDef) {
	if card == nil {
		return
	}
	v.validateFace(faceName, path, &card.CardFace, true)
	if card.Back.Exists {
		face := card.Back.Val
		v.validateFace(faceName, appendPath(path, "Back"), &face, true)
	}
}

func (v *cardValidator) validateTargetIndex(faceName, path string, targetIndex int, targets []game.TargetSpec, label string) {
	// Negative target indexes are rules-owned sentinels such as -1 for the
	// controller and -2 for the source/event object.
	if targetIndex < 0 {
		return
	}
	if targetIndex >= len(targets) {
		v.add(faceName, path, IssueTargetIndexOutOfRange, fmt.Sprintf("%s index %d has no matching TargetSpec", label, targetIndex))
	}
}

func (v *cardValidator) validateCondition(faceName, path string, condition *game.Condition, targets []game.TargetSpec) {
	if condition.Object.Exists {
		v.validateObjectReference(faceName, appendPath(path, "Object"), condition.Object.Val, targets)
	}
}

func (v *cardValidator) validateObjectReference(faceName, path string, ref game.ObjectReference, targets []game.TargetSpec) {
	switch ref.Kind {
	case game.ObjectReferenceTargetPermanent:
		v.validateTargetIndex(faceName, path, ref.TargetIndex, targets, "object reference target")
	case game.ObjectReferenceSourcePermanent:
		if ref.TargetIndex != 0 || ref.LinkID != "" {
			v.add(faceName, path, IssueInvalidReference, "source permanent reference must not set TargetIndex or LinkID")
		}
	case game.ObjectReferenceAttachedPermanent:
		if ref.TargetIndex >= 0 {
			v.validateTargetIndex(faceName, path, ref.TargetIndex, targets, "attached permanent reference target")
		}
	case game.ObjectReferenceLinkedObject:
		if ref.LinkID == "" {
			v.add(faceName, path, IssueInvalidReference, "linked object reference requires LinkID")
		}
	case game.ObjectReferenceEventPermanent:
		if ref.TargetIndex != 0 || ref.LinkID != "" {
			v.add(faceName, path, IssueInvalidReference, "event permanent reference must not set TargetIndex or LinkID")
		}
	case game.ObjectReferenceNone:
		v.add(faceName, path, IssueInvalidReference, "object reference has no kind")
	default:
		v.add(faceName, path, IssueInvalidReference, fmt.Sprintf("unknown object reference kind %d", ref.Kind))
	}
}

func (v *cardValidator) validatePlayerReference(faceName, path string, ref game.PlayerReference, targets []game.TargetSpec) {
	switch ref.Kind {
	case game.PlayerReferenceController:
		if ref.TargetIndex != 0 || ref.Object.Exists {
			v.add(faceName, path, IssueInvalidReference, "controller reference must not set TargetIndex or Object")
		}
	case game.PlayerReferenceTargetPlayer:
		v.validateTargetIndex(faceName, path, ref.TargetIndex, targets, "player reference target")
	case game.PlayerReferenceObjectController, game.PlayerReferenceObjectOwner:
		if !ref.Object.Exists {
			v.add(faceName, path, IssueInvalidReference, "object controller/owner reference requires Object")
			return
		}
		v.validateObjectReference(faceName, appendPath(path, "Object"), ref.Object.Val, targets)
	case game.PlayerReferenceNone:
		v.add(faceName, path, IssueInvalidReference, "player reference has no kind")
	default:
		v.add(faceName, path, IssueInvalidReference, fmt.Sprintf("unknown player reference kind %d", ref.Kind))
	}
}

func (v *cardValidator) validateCardCondition(faceName, path string, condition game.CardCondition) {
	v.validateCardReference(faceName, appendPath(path, "Card"), condition.Card)
	if !condition.RequirePermanentCard && len(condition.Types) == 0 && len(condition.Supertypes) == 0 && len(condition.SubtypesAny) == 0 {
		v.add(faceName, path, IssueInvalidReference, "card condition has no filters")
	}
}

func (v *cardValidator) validateCardReference(faceName, path string, ref game.CardReference) bool {
	switch ref.Kind {
	case game.CardReferenceLinked:
		if ref.LinkID == "" {
			v.add(faceName, path, IssueInvalidReference, "linked card reference requires LinkID")
			return false
		}
	case game.CardReferenceSource, game.CardReferenceEvent:
		if ref.LinkID != "" {
			v.add(faceName, path, IssueInvalidReference, "source/event card reference must not set LinkID")
			return false
		}
	case game.CardReferenceNone:
		v.add(faceName, path, IssueInvalidReference, "card reference has no kind")
		return false
	default:
		v.add(faceName, path, IssueInvalidReference, fmt.Sprintf("unknown card reference kind %d", ref.Kind))
		return false
	}
	return true
}

func (v *cardValidator) validateTokenCopySpec(faceName, path string, spec game.TokenCopySpec, targets []game.TargetSpec) {
	switch spec.Source {
	case game.TokenCopySourceObject:
		v.validateObjectReference(faceName, appendPath(path, "Object"), spec.Object, targets)
	case game.TokenCopySourceSourceCard:
	case game.TokenCopySourceNone:
		v.add(faceName, appendPath(path, "Source"), IssueInvalidReference, "token copy source has no kind")
	default:
		v.add(faceName, appendPath(path, "Source"), IssueInvalidReference, fmt.Sprintf("unknown token copy source %d", spec.Source))
	}
}

func (v *cardValidator) add(faceName, path string, code ValidationCode, message string) {
	cardName := ""
	if v.card != nil {
		cardName = v.card.Name
	}
	v.issues = append(v.issues, ValidationIssue{
		CardName: cardName,
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
