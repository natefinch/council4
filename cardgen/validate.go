package cardgen

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

// ValidationCode identifies a class of card-definition validation issue.
type ValidationCode string

// Validation issue codes identify generated-card validation failures.
//
// Structural codes (nil-card through invalid-ability-body) are mapped from
// [game.CardDefIssueCode] values returned by [game.ValidateCardDef].
// Policy codes are evaluated here in the cardgen adapter layer.
const (
	IssueNilCard                    ValidationCode = "nil-card"
	IssueMissingName                ValidationCode = "missing-name"
	IssueOracleWithoutAbilities     ValidationCode = "oracle-without-abilities"
	IssueTargetIndexOutOfRange      ValidationCode = "target-index-out-of-range"
	IssueInvalidReference           ValidationCode = "invalid-reference"
	IssueInvalidTargetSpec          ValidationCode = "invalid-target-spec"
	IssueInvalidKeywordAbility      ValidationCode = "invalid-keyword-ability"
	IssueInvalidAbilityBody         ValidationCode = "invalid-ability-body"
	IssueInvalidSelection           ValidationCode = "invalid-selection"
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
// be checked statically. It delegates structural validation to
// [game.ValidateCardDef] and then applies policy checks from opts.
func ValidateCard(card *game.CardDef, opts ValidationOptions) []ValidationIssue {
	gameIssues := game.ValidateCardDef(card)

	cardName := ""
	if card != nil {
		cardName = card.Name
	}

	issues := make([]ValidationIssue, 0, len(gameIssues))
	for _, gi := range gameIssues {
		issues = append(issues, ValidationIssue{
			CardName: cardName,
			FaceName: gi.FaceName,
			Path:     gi.Path,
			Code:     mapCardDefIssueCode(gi.Code),
			Message:  gi.Message,
		})
	}

	// Policy checks are tooling/runtime concerns that are not part of game data
	// structural validation. They are evaluated here in the adapter.
	if card != nil && (len(opts.KnownImplementationIDs) > 0 || opts.ReportImplementationIDs) {
		issues = append(issues, validateFacePolicy(cardName, card.Name, "", &card.CardFace, opts)...)
		if card.Back.Exists {
			face := card.Back.Val
			name := face.Name
			if strings.TrimSpace(name) == "" {
				name = "back face"
			}
			issues = append(issues, validateFacePolicy(cardName, name, "Back", &face, opts)...)
		}
		if card.Alternate.Exists {
			face := card.Alternate.Val
			name := face.Name
			if strings.TrimSpace(name) == "" {
				name = "alternate face"
			}
			issues = append(issues, validateFacePolicy(cardName, name, "Alternate", &face, opts)...)
		}
	}

	return issues
}

// validateFacePolicy reports policy issues for a single face: unregistered or
// reported ImplementationID values. These checks depend on ValidationOptions
// and are never part of game.ValidateCardDef.
func validateFacePolicy(cardName, faceName, path string, face *game.CardFace, opts ValidationOptions) []ValidationIssue {
	var issues []ValidationIssue
	if face.ImplementationID != "" && len(opts.KnownImplementationIDs) > 0 && !opts.KnownImplementationIDs[face.ImplementationID] {
		issues = append(issues, ValidationIssue{
			CardName: cardName,
			FaceName: faceName,
			Path:     path,
			Code:     IssueUnregisteredImplementation,
			Message:  fmt.Sprintf("implementation ID %q is not registered", face.ImplementationID),
		})
	}
	if face.ImplementationID != "" && opts.ReportImplementationIDs {
		issues = append(issues, ValidationIssue{
			CardName: cardName,
			FaceName: faceName,
			Path:     path,
			Code:     IssueImplementationRequired,
			Message:  fmt.Sprintf("implementation ID %q requires hand-written rules support", face.ImplementationID),
		})
	}
	return issues
}

// mapCardDefIssueCode translates a game.CardDefIssueCode to the corresponding
// cardgen ValidationCode. The string values are intentionally identical so
// that reports and tests that compare code strings see no change.
func mapCardDefIssueCode(code game.CardDefIssueCode) ValidationCode {
	switch code {
	case game.CardDefIssueNilCard:
		return IssueNilCard
	case game.CardDefIssueMissingName:
		return IssueMissingName
	case game.CardDefIssueOracleWithoutAbilities:
		return IssueOracleWithoutAbilities
	case game.CardDefIssueTargetIndexOutOfRange:
		return IssueTargetIndexOutOfRange
	case game.CardDefIssueInvalidReference:
		return IssueInvalidReference
	case game.CardDefIssueInvalidTargetSpec:
		return IssueInvalidTargetSpec
	case game.CardDefIssueInvalidKeywordAbility:
		return IssueInvalidKeywordAbility
	case game.CardDefIssueInvalidAbilityBody:
		return IssueInvalidAbilityBody
	case game.CardDefIssueInvalidSelection:
		return IssueInvalidSelection
	default:
		return ValidationCode(code)
	}
}
