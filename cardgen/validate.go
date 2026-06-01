package cardgen

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// ValidationCode identifies a class of card-definition validation issue.
type ValidationCode string

const (
	IssueNilCard                    ValidationCode = "nil-card"
	IssueMissingName                ValidationCode = "missing-name"
	IssueOracleWithoutAbilities     ValidationCode = "oracle-without-abilities"
	IssueUnexecutedEffect           ValidationCode = "unexecuted-effect"
	IssueMissingSearchSpec          ValidationCode = "missing-search-spec"
	IssueUnsupportedSearchSpec      ValidationCode = "unsupported-search-spec"
	IssueTargetIndexOutOfRange      ValidationCode = "target-index-out-of-range"
	IssueInvalidReference           ValidationCode = "invalid-reference"
	IssueInvalidTargetSpec          ValidationCode = "invalid-target-spec"
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
	if len(v.card.Faces) == 0 {
		v.validateFace(v.card.Name, "", v.card.OracleText, v.card.ImplementationID, v.card.Abilities, true)
		return
	}
	v.validateFace(v.card.Name, "", v.card.OracleText, v.card.ImplementationID, v.card.Abilities, false)
	for i, face := range v.card.Faces {
		name := face.Name
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("face %d", i)
		}
		v.validateFace(name, fmt.Sprintf("Faces[%d]", i), face.OracleText, face.ImplementationID, face.Abilities, true)
	}
}

func (v *cardValidator) validateFace(faceName string, path string, oracleText string, implementationID string, abilities []game.AbilityDef, walkAbilities bool) {
	if strings.TrimSpace(oracleText) != "" && len(abilities) == 0 && implementationID == "" {
		v.add(faceName, path, IssueOracleWithoutAbilities, "oracle text is non-empty but no abilities or hand-written implementation are defined")
	}
	if implementationID != "" && len(v.opts.KnownImplementationIDs) > 0 && !v.opts.KnownImplementationIDs[implementationID] {
		v.add(faceName, path, IssueUnregisteredImplementation, fmt.Sprintf("implementation ID %q is not registered", implementationID))
	}
	if implementationID != "" && v.opts.ReportImplementationIDs {
		v.add(faceName, path, IssueImplementationRequired, fmt.Sprintf("implementation ID %q requires hand-written rules support", implementationID))
	}
	if !walkAbilities {
		return
	}
	for i := range abilities {
		abilityPath := appendPath(path, fmt.Sprintf("Abilities[%d]", i))
		v.validateAbility(faceName, abilityPath, &abilities[i])
	}
}

func (v *cardValidator) validateAbility(faceName string, path string, ability *game.AbilityDef) {
	if ability.EnchantTarget.Exists {
		v.validateTargetSpec(faceName, appendPath(path, "EnchantTarget"), ability.EnchantTarget.Val)
	}
	for i, target := range ability.Targets {
		v.validateTargetSpec(faceName, appendPath(path, fmt.Sprintf("Targets[%d]", i)), target)
	}
	for i, effect := range ability.Effects {
		v.validateEffect(faceName, appendPath(path, fmt.Sprintf("Effects[%d]", i)), effect, ability.Targets)
	}
	for i, effect := range ability.KickerEffects {
		v.validateEffect(faceName, appendPath(path, fmt.Sprintf("KickerEffects[%d]", i)), effect, ability.Targets)
	}
	for i, mode := range ability.Modes {
		modePath := appendPath(path, fmt.Sprintf("Modes[%d]", i))
		for j, target := range mode.Targets {
			v.validateTargetSpec(faceName, appendPath(modePath, fmt.Sprintf("Targets[%d]", j)), target)
		}
		targets := mode.Targets
		if len(targets) == 0 {
			targets = ability.Targets
		}
		for j, effect := range mode.Effects {
			v.validateEffect(faceName, appendPath(modePath, fmt.Sprintf("Effects[%d]", j)), effect, targets)
		}
	}
}

func (v *cardValidator) validateTargetSpec(faceName string, path string, target game.TargetSpec) {
	if target.MinTargets < 0 || target.MaxTargets < 0 {
		v.add(faceName, path, IssueInvalidTargetSpec, "target counts must be non-negative")
		return
	}
	if target.MaxTargets < target.MinTargets {
		v.add(faceName, path, IssueInvalidTargetSpec, "max targets is less than min targets")
	}
}

func (v *cardValidator) validateEffect(faceName string, path string, effect game.Effect, targets []game.TargetSpec) {
	if !rules.IsEffectTypeExecuted(effect.Type) {
		v.add(faceName, path, IssueUnexecutedEffect, fmt.Sprintf("effect type %d is not executed by rules", effect.Type))
	}
	if effect.Type == game.EffectSearch {
		if !effect.Search.Exists {
			v.add(faceName, path, IssueMissingSearchSpec, "search effect has no SearchSpec")
		} else if effect.Search.Val.SourceZone != game.ZoneLibrary || effect.Search.Val.Destination != game.ZoneHand {
			v.add(faceName, path, IssueUnsupportedSearchSpec, "only library-to-hand SearchSpec is currently supported")
		}
	}
	if effect.Selector != game.EffectSelectorNone && effect.PlayerSelector != game.PlayerSelectorNone {
		v.add(faceName, path, IssueInvalidReference, "Effect cannot set both Selector and PlayerSelector")
	}
	if effect.PlayerSelector != game.PlayerSelectorNone && effect.Type != game.EffectDamage {
		v.add(faceName, appendPath(path, "PlayerSelector"), IssueInvalidReference, "PlayerSelector is only supported on damage effects")
	}
	v.validateTargetIndex(faceName, path, effect.TargetIndex, targets, "effect target")
	if effect.DamageSource.Exists {
		if effect.Type != game.EffectDamage {
			v.add(faceName, appendPath(path, "DamageSource"), IssueInvalidReference, "DamageSource is only supported on damage effects")
		}
		v.validateObjectReference(faceName, appendPath(path, "DamageSource"), effect.DamageSource.Val, targets)
	}
	if effect.Recipient.Exists {
		switch effect.Type {
		case game.EffectCreateToken, game.EffectInvestigate:
		default:
			v.add(faceName, appendPath(path, "Recipient"), IssueInvalidReference, "Recipient is only supported on token-creating effects")
		}
		v.validatePlayerReference(faceName, appendPath(path, "Recipient"), effect.Recipient.Val, targets)
	}
	if effect.Condition.Exists {
		v.validateTargetIndex(faceName, appendPath(path, "Condition"), effect.Condition.Val.TargetIndex, targets, "condition target")
	}
	if effect.DynamicAmount.Exists && dynamicAmountUsesTarget(effect.DynamicAmount.Val) {
		v.validateTargetIndex(faceName, appendPath(path, "DynamicAmount"), effect.DynamicAmount.Val.TargetIndex, targets, "dynamic amount target")
	}
	if effect.CounterSource.Kind == game.CounterSourceTarget {
		v.validateTargetIndex(faceName, appendPath(path, "CounterSource"), effect.CounterSource.TargetIndex, targets, "counter source target")
	}
	if effect.DelayedTrigger.Exists {
		delayedPath := appendPath(path, "DelayedTrigger")
		for i, target := range effect.DelayedTrigger.Val.Targets {
			v.validateTargetSpec(faceName, appendPath(delayedPath, fmt.Sprintf("Targets[%d]", i)), target)
		}
		for i, delayedEffect := range effect.DelayedTrigger.Val.Effects {
			v.validateEffect(faceName, appendPath(delayedPath, fmt.Sprintf("Effects[%d]", i)), delayedEffect, effect.DelayedTrigger.Val.Targets)
		}
	}
	if effect.Token.Exists && effect.Token.Val != nil {
		v.validateNestedCard(faceName, appendPath(path, "Token"), effect.Token.Val)
	}
	for i, continuous := range effect.ContinuousEffects {
		continuousPath := appendPath(path, fmt.Sprintf("ContinuousEffects[%d]", i))
		for j := range continuous.AddAbilities {
			v.validateAbility(faceName, appendPath(continuousPath, fmt.Sprintf("AddAbilities[%d]", j)), &continuous.AddAbilities[j])
		}
	}
	for i := range effect.EmblemAbilities {
		v.validateAbility(faceName, appendPath(path, fmt.Sprintf("EmblemAbilities[%d]", i)), &effect.EmblemAbilities[i])
	}
}

func (v *cardValidator) validateNestedCard(faceName string, path string, card *game.CardDef) {
	if card == nil {
		return
	}
	if len(card.Faces) == 0 {
		v.validateFace(faceName, path, card.OracleText, card.ImplementationID, card.Abilities, true)
		return
	}
	v.validateFace(faceName, path, card.OracleText, card.ImplementationID, card.Abilities, false)
	for i, face := range card.Faces {
		v.validateFace(faceName, appendPath(path, fmt.Sprintf("Faces[%d]", i)), face.OracleText, face.ImplementationID, face.Abilities, true)
	}
}

func (v *cardValidator) validateTargetIndex(faceName string, path string, targetIndex int, targets []game.TargetSpec, label string) {
	// Negative target indexes are rules-owned sentinels such as -1 for the
	// controller and -2 for the source/event object.
	if targetIndex < 0 {
		return
	}
	if targetIndex >= len(targets) {
		v.add(faceName, path, IssueTargetIndexOutOfRange, fmt.Sprintf("%s index %d has no matching TargetSpec", label, targetIndex))
	}
}

func (v *cardValidator) validateObjectReference(faceName string, path string, ref game.ObjectReference, targets []game.TargetSpec) {
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
	case game.ObjectReferenceNone:
		v.add(faceName, path, IssueInvalidReference, "object reference has no kind")
	default:
		v.add(faceName, path, IssueInvalidReference, fmt.Sprintf("unknown object reference kind %d", ref.Kind))
	}
}

func (v *cardValidator) validatePlayerReference(faceName string, path string, ref game.PlayerReference, targets []game.TargetSpec) {
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

func dynamicAmountUsesTarget(dynamic game.DynamicAmount) bool {
	switch dynamic.Kind {
	case game.DynamicAmountTargetPower,
		game.DynamicAmountTargetToughness,
		game.DynamicAmountTargetManaValue,
		game.DynamicAmountTargetCounters:
		return true
	default:
		return false
	}
}

func (v *cardValidator) add(faceName string, path string, code ValidationCode, message string) {
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

func appendPath(parent string, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}
