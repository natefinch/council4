package game

import (
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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
	CardDefIssueInvalidAlternativeCost CardDefIssueCode = "invalid-alternative-cost"
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
		face.Overload.Exists ||
		face.EntersPrepared ||
		len(face.ActivatedAbilities) > 0 ||
		len(face.ManaAbilities) > 0 ||
		len(face.LoyaltyAbilities) > 0 ||
		len(face.TriggeredAbilities) > 0 ||
		len(face.ChapterAbilities) > 0 ||
		len(face.ReplacementAbilities) > 0 ||
		len(face.StaticAbilities) > 0 ||
		len(face.AdditionalCosts) > 0 ||
		len(face.AlternativeCosts) > 0 ||
		face.DynamicPower.Exists ||
		face.DynamicToughness.Exists
	if strings.TrimSpace(face.OracleText) != "" && !hasAbilities && face.ImplementationID == "" {
		v.add(faceName, path, CardDefIssueOracleWithoutAbilities, "oracle text is non-empty but no abilities or hand-written implementation are defined")
	}
	v.validateEntryChoiceDependencies(faceName, path, face)
	v.validateLinkedExileColorDependencies(faceName, path, face)
	if face.SpellAbility.Exists {
		v.validateAbilityBody(faceName, appendPath(path, "SpellAbility"), &face.SpellAbility.Val, nil)
	}

	if face.Overload.Exists {
		if !face.SpellAbility.Exists {
			v.add(faceName, appendPath(path, "Overload"), CardDefIssueInvalidAlternativeCost, "overload requires a normal spell ability")
		}
		if len(face.Overload.Val.Cost) == 0 {
			v.add(faceName, appendPath(path, "Overload.Cost"), CardDefIssueInvalidAlternativeCost, "overload cost is empty")
		}
		if abilityContentHasTargets(face.Overload.Val.SpellAbility) {
			v.add(faceName, appendPath(path, "Overload.SpellAbility"), CardDefIssueInvalidAlternativeCost, "overload spell ability must not target")
		}
		v.validateAbilityBody(faceName, appendPath(path, "Overload.SpellAbility"), &face.Overload.Val.SpellAbility, nil)
	}
	for i := range face.ActivatedAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("ActivatedAbilities[%d]", i)), &face.ActivatedAbilities[i], nil)
	}
	for i := range face.ManaAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("ManaAbilities[%d]", i)), &face.ManaAbilities[i], nil)
	}
	for i := range face.LoyaltyAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("LoyaltyAbilities[%d]", i)), &face.LoyaltyAbilities[i], nil)
	}
	for i := range face.TriggeredAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("TriggeredAbilities[%d]", i)), &face.TriggeredAbilities[i], nil)
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

		v.validateAbilityBody(faceName, chapterPath, &face.ChapterAbilities[i], nil)
	}
	for i := range face.ReplacementAbilities {
		v.validateReplacementAbility(faceName, appendPath(path, fmt.Sprintf("ReplacementAbilities[%d]", i)), &face.ReplacementAbilities[i])
	}
	for i := range face.StaticAbilities {
		v.validateAbilityBody(faceName, appendPath(path, fmt.Sprintf("StaticAbilities[%d]", i)), &face.StaticAbilities[i], nil)
	}
	for i, alternative := range face.AlternativeCosts {
		if alternative.Condition != cost.AlternativeConditionNone &&
			alternative.Condition != cost.AlternativeConditionControlsCommander &&
			alternative.Condition != cost.AlternativeConditionNotYourTurn {
			v.add(
				faceName,
				appendPath(path, fmt.Sprintf("AlternativeCosts[%d].Condition", i)),
				CardDefIssueInvalidAlternativeCost,
				"alternative cost has an unknown condition",
			)
		}
	}
}

func (v *cardDefValidator) validateEntryChoiceDependencies(faceName, path string, face *CardFace) {
	if faceProvidesEntryTypeChoice(face) {
		return
	}
	for i := range face.ManaAbilities {
		for _, mode := range face.ManaAbilities[i].Content.Modes {
			for j := range mode.Sequence {
				addMana, ok := mode.Sequence[j].Primitive.(AddMana)
				if !ok || !addMana.SpendRider.Exists ||
					addMana.SpendRider.Val.ChosenSubtypeFrom != EntryTypeChoiceKey {
					continue
				}
				v.add(
					faceName,
					appendPath(path, fmt.Sprintf("ManaAbilities[%d]", i)),
					CardDefIssueInvalidAbilityBody,
					"chosen-type mana spend rider requires an entry-time creature-type choice",
				)
				return
			}
		}
	}
	for i := range face.StaticAbilities {
		for j := range face.StaticAbilities[i].RuleEffects {
			effect := &face.StaticAbilities[i].RuleEffects[j]
			if effect.Kind != RuleEffectCostModifier ||
				!effect.CostModifier.ChosenSubtypeFromEntryChoice {
				continue
			}
			v.add(
				faceName,
				appendPath(path, fmt.Sprintf("StaticAbilities[%d]", i)),
				CardDefIssueInvalidAbilityBody,
				"chosen-type cost modifier requires an entry-time creature-type choice",
			)
			return
		}
	}
}

func faceProvidesEntryTypeChoice(face *CardFace) bool {
	for i := range face.ReplacementAbilities {
		if face.ReplacementAbilities[i].Replacement.EntryTypeChoice {
			return true
		}
	}
	return false
}

// validateLinkedExileColorDependencies enforces that a mana ability whose colors
// come from a card imprinted by an exile-from-hand effect (Chrome Mox) only
// appears alongside such an effect on the same face. The imprint mana ability is
// useless without an ExileFromHand publishing the same link, so a face that
// declares one without the other is a lowering error rather than a silently
// dead ability.
func (v *cardDefValidator) validateLinkedExileColorDependencies(faceName, path string, face *CardFace) {
	published := map[string]bool{}
	collectExileFromHandLinks(face, published)
	for i := range face.ManaAbilities {
		for _, mode := range face.ManaAbilities[i].Content.Modes {
			for j := range mode.Sequence {
				choose, ok := mode.Sequence[j].Primitive.(Choose)
				if !ok || choose.Choice.Kind != ResolutionChoiceMana ||
					choose.Choice.ColorSource != ResolutionChoiceColorSourceLinkedExileColors {
					continue
				}
				if choose.Choice.LinkID == "" || !published[choose.Choice.LinkID] {
					v.add(
						faceName,
						appendPath(path, fmt.Sprintf("ManaAbilities[%d]", i)),
						CardDefIssueInvalidAbilityBody,
						"linked-exile-color mana ability requires an exile-from-hand effect publishing its link on the same face",
					)
				}
			}
		}
	}
}

// collectExileFromHandLinks records the link keys published by every
// ExileFromHand primitive across the face's ability contents.
func collectExileFromHandLinks(face *CardFace, into map[string]bool) {
	collect := func(content AbilityContent) {
		for _, mode := range content.Modes {
			for i := range mode.Sequence {
				exile, ok := mode.Sequence[i].Primitive.(ExileFromHand)
				if ok && exile.PublishLinked != "" {
					into[string(exile.PublishLinked)] = true
				}
			}
		}
	}
	if face.SpellAbility.Exists {
		collect(face.SpellAbility.Val)
	}
	for i := range face.ActivatedAbilities {
		collect(face.ActivatedAbilities[i].Content)
	}
	for i := range face.TriggeredAbilities {
		collect(face.TriggeredAbilities[i].Content)
	}
	for i := range face.ChapterAbilities {
		collect(face.ChapterAbilities[i].Content)
	}
	for i := range face.LoyaltyAbilities {
		collect(face.LoyaltyAbilities[i].Content)
	}
}

func abilityContentHasTargets(content AbilityContent) bool {
	if len(content.SharedTargets) > 0 {
		return true
	}
	for _, mode := range content.Modes {
		if len(mode.Targets) > 0 {
			return true
		}
	}
	return false
}

func (v *cardDefValidator) validateAbilityBody(faceName, path string, body Ability, targets []TargetSpec) {
	switch abilityBody := body.(type) {
	case *AbilityContent:
		v.validateAbilityContent(faceName, path, *abilityBody, targets)
	case *ActivatedAbility:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		for i := range abilityBody.CostModifiers {
			v.validateCostModifier(faceName, appendPath(path, fmt.Sprintf("CostModifiers[%d]", i)), abilityBody.CostModifiers[i], true)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case *ManaAbility:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		if len(abilityBody.Content.Modes) > 0 {
			v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
		}
	case *LoyaltyAbility:
		if abilityBody.ActivationCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "ActivationCondition"), &abilityBody.ActivationCondition.Val, targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case *TriggeredAbility:
		v.validateTriggerPattern(faceName, appendPath(path, "Trigger.Pattern"), &abilityBody.Trigger.Pattern)
		if abilityBody.Trigger.InterveningCondition.Exists {
			v.validateCondition(faceName, appendPath(path, "Trigger.InterveningCondition"), &abilityBody.Trigger.InterveningCondition.Val, targets)
		}
		for i := range abilityBody.KeywordAbilities {
			v.validateKeywordAbility(faceName, appendPath(path, fmt.Sprintf("KeywordAbilities[%d]", i)), abilityBody.KeywordAbilities[i], targets)
		}
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case *ChapterAbility:
		v.validateAbilityContent(faceName, appendPath(path, "Content"), abilityBody.Content, targets)
	case *StaticAbility:
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
	if ability.UnlessPaid.Exists {
		paymentPath := appendPath(path, "UnlessPaid")
		if err := validateResolutionPayment(ability.UnlessPaid.Val, nil, true); err != nil {
			v.add(faceName, paymentPath, CardDefIssueInvalidAbilityBody, err.Error())
		} else if err := validateEnterBattlefieldResolutionPayment(ability.UnlessPaid.Val); err != nil {
			v.add(faceName, paymentPath, CardDefIssueInvalidAbilityBody, err.Error())
		}
	}
	if ability.Replacement.Condition.Exists {
		v.validateCondition(faceName, appendPath(path, "Replacement.Condition"), &ability.Replacement.Condition.Val, nil)
	}
}

func validateEnterBattlefieldResolutionPayment(payment ResolutionPayment) error {
	if payment.DynamicGenericManaCost.Exists && payment.DynamicGenericManaCost.Val != nil {
		if err := validateEnterBattlefieldResolutionPaymentDynamic("dynamic generic mana cost", payment.DynamicGenericManaCost.Val); err != nil {
			return err
		}
	}
	if payment.ManaCostMultiplier.Exists && payment.ManaCostMultiplier.Val != nil {
		if err := validateEnterBattlefieldResolutionPaymentDynamic("mana cost multiplier", payment.ManaCostMultiplier.Val); err != nil {
			return err
		}
	}
	return nil
}

func validateEnterBattlefieldResolutionPaymentDynamic(name string, dynamic *DynamicAmount) error {
	switch dynamic.Kind {
	case DynamicAmountConstant,
		DynamicAmountX,
		DynamicAmountControllerLife,
		DynamicAmountControllerHandSize,
		DynamicAmountControllerGraveyardSize,
		DynamicAmountControllerBasicLandTypeCount,
		DynamicAmountOpponentCount:
		return nil
	case DynamicAmountObjectManaValue,
		DynamicAmountObjectCounters:
		if dynamic.Object != SourcePermanentReference() {
			return fmt.Errorf("enter-the-battlefield replacement %s must reference the source permanent", name)
		}
		return nil
	default:
		return fmt.Errorf("enter-the-battlefield replacement cannot safely evaluate %s kind %d", name, dynamic.Kind)
	}
}

func (v *cardDefValidator) validateAbilityContent(faceName, path string, content AbilityContent, fallbackTargets []TargetSpec) {
	v.validateAbilityContentWithLinked(faceName, path, content, fallbackTargets, nil, nil)
}

func (v *cardDefValidator) validateAbilityContentWithLinked(
	faceName, path string,
	content AbilityContent,
	fallbackTargets []TargetSpec,
	inheritedLinked map[LinkedKey]int,
	capturedTargets []TargetSpec,
) {
	if len(content.Modes) == 0 {
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, "ability content has no modes")
		return
	}
	minModes, maxModes := content.MinModes, content.MaxModes
	if minModes == 0 && maxModes == 0 {
		minModes, maxModes = 1, 1
	}
	if minModes < 0 {
		v.add(faceName, appendPath(path, "MinModes"), CardDefIssueInvalidAbilityBody, "minimum modes must not be negative")
	}
	if maxModes < 1 {
		v.add(faceName, appendPath(path, "MaxModes"), CardDefIssueInvalidAbilityBody, "maximum modes must be at least one")
	}
	if maxModes < minModes {
		v.add(faceName, appendPath(path, "MaxModes"), CardDefIssueInvalidAbilityBody, "maximum modes must not be less than minimum modes")
	}
	if !content.AllowDuplicateModes && maxModes > len(content.Modes) {
		v.add(faceName, appendPath(path, "MaxModes"), CardDefIssueInvalidAbilityBody, "maximum modes exceeds available distinct modes")
	}
	if bonus := content.ModeChoiceBonus; bonus.Condition != ModeChoiceConditionNone || bonus.AdditionalMaxModes != 0 {
		if bonus.Condition != ModeChoiceConditionControlsCommander {
			v.add(faceName, appendPath(path, "ModeChoiceBonus"), CardDefIssueInvalidAbilityBody, "mode choice bonus has unsupported condition")
		}
		if bonus.AdditionalMaxModes < 1 {
			v.add(faceName, appendPath(path, "ModeChoiceBonus"), CardDefIssueInvalidAbilityBody, "mode choice bonus must add at least one maximum mode")
		}
		if !content.AllowDuplicateModes && maxModes+bonus.AdditionalMaxModes > len(content.Modes) {
			v.add(faceName, appendPath(path, "ModeChoiceBonus"), CardDefIssueInvalidAbilityBody, "mode choice bonus exceeds available modes")
		}
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
		enclosingTargets := capturedTargets
		if enclosingTargets == nil {
			enclosingTargets = targets
		}
		v.validateInstructionSequence(
			faceName,
			appendPath(modePath, "Sequence"),
			mode.Sequence,
			targets,
			enclosingTargets,
			inheritedLinked,
		)
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
	case CumulativeUpkeepKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case EquipKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case EnchantKeyword:
		v.validateTargetSpec(faceName, appendPath(path, "Target"), &keyword.Target)
	case CyclingKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case ScavengeKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case UnearthKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case NinjutsuKeyword:
		v.validateManaKeywordCost(faceName, path, keyword.Cost)
	case OutlastKeyword:
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
	case FlashbackKeyword:
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
		if keyword.ChosenColor {
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
	case FabricateKeyword:
		if keyword.Count <= 0 {
			v.add(faceName, appendPath(path, "Count"), CardDefIssueInvalidKeywordAbility, "fabricate count must be positive")
		}
	case SoulshiftKeyword:
		if keyword.Count <= 0 {
			v.add(faceName, appendPath(path, "Count"), CardDefIssueInvalidKeywordAbility, "soulshift count must be positive")
		}
	case DredgeKeyword:
		if keyword.Count <= 0 {
			v.add(faceName, appendPath(path, "Count"), CardDefIssueInvalidKeywordAbility, "dredge count must be positive")
		}
	case RampageKeyword:
		if keyword.Count <= 0 {
			v.add(faceName, appendPath(path, "Count"), CardDefIssueInvalidKeywordAbility, "rampage count must be positive")
		}
	case LandwalkKeyword:
		if !keyword.AnyLand && !keyword.Nonbasic && keyword.Subtype == "" {
			v.add(faceName, appendPath(path, "Subtype"), CardDefIssueInvalidKeywordAbility, "landwalk needs a land subtype, AnyLand, or Nonbasic")
		}
		if keyword.AnyLand && keyword.Subtype != "" {
			v.add(faceName, path, CardDefIssueInvalidKeywordAbility, "landwalk must not combine AnyLand with a subtype")
		}
		if keyword.Nonbasic && (keyword.AnyLand || keyword.Subtype != "") {
			v.add(faceName, path, CardDefIssueInvalidKeywordAbility, "landwalk must not combine Nonbasic with AnyLand or a subtype")
		}
	case SaddleKeyword:
		if keyword.Power <= 0 {
			v.add(faceName, appendPath(path, "Power"), CardDefIssueInvalidKeywordAbility, "saddle power must be positive")
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
	capturedTargets []TargetSpec,
	inheritedLinked map[LinkedKey]int,
) {
	if err := validateInstructionSequenceWithLinked(seq, targets, true, inheritedLinked, capturedTargets, true); err != nil {
		v.add(faceName, path, CardDefIssueInvalidAbilityBody, err.Error())
	}
	publishedLinked := make(map[LinkedKey]int, len(inheritedLinked))
	maps.Copy(publishedLinked, inheritedLinked)
	for i := range seq {
		instructionPath := appendPath(path, fmt.Sprintf("Instructions[%d]", i))
		if seq[i].OptionalActor.Exists {
			if !seq[i].Optional {
				v.add(faceName, instructionPath, CardDefIssueInvalidAbilityBody, "OptionalActor set on a non-optional instruction")
			}
			referenceTargets := targets
			if seq[i].OptionalActor.Val.Kind() == PlayerReferenceCapturedTargetController {
				referenceTargets = capturedTargets
			}
			v.validatePlayerRef(faceName, appendPath(instructionPath, "OptionalActor"), seq[i].OptionalActor.Val, referenceTargets)
		}
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
				targets,
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
	allowsPermanents := knownAllows&TargetAllowPermanent != 0
	stackSelection := target.Predicate.Selection()
	// Controller restrictions are supported for stack-object targets (e.g.
	// "target activated ability you don't control"), so they do not count as an
	// unsupported permanent predicate here.
	stackSelection.Controller = ControllerAny
	// A mana-value comparison is a supported stack-spell qualifier ("counter
	// target spell with mana value N"); the runtime matcher applies it to the
	// spell choice, so it does not count as an unsupported permanent predicate.
	stackSelection.ManaValue = opt.V[compare.Int]{}
	// A combined "spell or permanent" target carries permanent predicates that
	// constrain only its permanent alternative; the stack-object side is gated
	// by StackObjectKinds and spell qualifiers, so they are not unsupported.
	if allowsStackObjects && !allowsPermanents && !stackSelection.Empty() {
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
	hasSpellTypePredicate := len(target.Predicate.SpellCardTypes) > 0 ||
		len(target.Predicate.SpellCardTypesAny) > 0 ||
		len(target.Predicate.ExcludedSpellCardTypes) > 0
	if hasSpellTypePredicate && (!allowsSpells || allowsAbilities) {
		v.add(faceName, appendPath(path, "Predicate"), CardDefIssueInvalidTargetSpec, "spell type predicates require spell-only stack-object targets")
	}
	if len(target.Predicate.SpellCardTypes) > 0 && len(target.Predicate.SpellCardTypesAny) > 0 {
		v.add(faceName, appendPath(path, "Predicate"), CardDefIssueInvalidTargetSpec, "spell type target cannot combine all-of and any-of predicates")
	}
	if len(target.Predicate.SpellCardTypesAny) == 1 {
		v.add(faceName, appendPath(path, "Predicate.SpellCardTypesAny"), CardDefIssueInvalidTargetSpec, "spell type union requires at least two card types")
	}
	seenTypes := make(map[types.Card]bool, len(target.Predicate.SpellCardTypesAny))
	for i, cardType := range target.Predicate.SpellCardTypesAny {
		if seenTypes[cardType] {
			v.add(faceName, appendPath(path, fmt.Sprintf("Predicate.SpellCardTypesAny[%d]", i)), CardDefIssueInvalidTargetSpec, "duplicate spell card type")
		}
		seenTypes[cardType] = true
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
		selection.ExcludedSubtype != "" ||
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
		abilityPath := appendPath(path, fmt.Sprintf("AddAbilities[%d]", i))
		v.validateAbilityBody(faceName, abilityPath, continuous.AddAbilities[i], nil)
		if manaAbility, ok := continuous.AddAbilities[i].(*ManaAbility); ok &&
			!IsTapAnyColorManaAbility(manaAbility) &&
			!IsTapColorlessManaAbility(manaAbility) &&
			!IsTapOneColorManaAbility(manaAbility) &&
			!IsTapSacrificeAnyOneColorManaAbility(manaAbility) {
			v.add(faceName, abilityPath, CardDefIssueInvalidAbilityBody, "continuous effects support only the standard tap-for-one-mana-of-any-color granted mana ability, the bare tap-for-one-mana ability, or the Treasure-style sacrifice mana ability")
		}
	}
	if len(continuous.AddAbilities) > 0 && continuous.Layer != LayerAbility {
		v.add(faceName, appendPath(path, "Layer"), CardDefIssueInvalidAbilityBody, "granted abilities require the ability layer")
	}
	if continuous.AffectedSource && !continuous.Group.Empty() {
		v.add(faceName, path, CardDefIssueInvalidReference, "continuous effect sets both AffectedSource and Group")
	}
	if !continuous.Group.Empty() {
		v.validateGroupRef(faceName, appendPath(path, "Group"), continuous.Group, targets)
	}
	if continuous.AddSubtypeFromEntryChoice != "" {
		if continuous.AddSubtypeFromEntryChoice != EntryTypeChoiceKey {
			v.add(faceName, appendPath(path, "AddSubtypeFromEntryChoice"), CardDefIssueInvalidReference, "entry-choice subtype reference must use EntryTypeChoiceKey")
		}
		if continuous.Layer != LayerType {
			v.add(faceName, appendPath(path, "Layer"), CardDefIssueInvalidAbilityBody, "entry-choice subtype reference requires the type layer")
		}
		if !continuous.AffectedSource {
			v.add(faceName, appendPath(path, "AffectedSource"), CardDefIssueInvalidReference, "entry-choice subtype reference must affect its source")
		}
	}
}

func (v *cardDefValidator) validateRuleEffect(faceName, path string, effect *RuleEffect) {
	if effect == nil {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "rule effect is nil")
		return
	}
	if !effect.Kind.Valid() {
		v.add(faceName, appendPath(path, "Kind"), CardDefIssueInvalidRuleEffect, "rule effect has an unsupported kind")
		return
	}
	switch effect.Kind {
	case RuleEffectCostModifier:
		v.validateCostModifier(faceName, appendPath(path, "CostModifier"), effect.CostModifier, false)
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
	case RuleEffectNoMaximumHandSize, RuleEffectLifeTotalCantChange, RuleEffectCastSpellsAsThoughFlash, RuleEffectPlayWithTopCardRevealed, RuleEffectLookAtTopCardAnyTime:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "player rule effects must set affected player")
		}
		if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "player rule effects cannot affect a permanent")
		}
	case RuleEffectAdditionalLandPlays:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "additional land plays must set affected player")
		}
		if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "additional land plays cannot affect a permanent")
		}
		if effect.AdditionalLandPlays < 1 {
			v.add(faceName, appendPath(path, "AdditionalLandPlays"), CardDefIssueInvalidRuleEffect, "additional land plays must grant at least one extra land play")
		}
	case RuleEffectAttackTax:
		v.validateAttackTaxRuleEffect(faceName, path, effect)
	case RuleEffectPlayFromZone:
		if err := validatePlayFromZoneRuleEffect(effect, false, true); err != nil {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, err.Error())
		}
	case RuleEffectPlayLandsFromZone:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "play-from-zone permission must set affected player")
		}
		if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "play-from-zone permission cannot affect a permanent")
		}
		if effect.CastFromZone == zone.None {
			v.add(faceName, appendPath(path, "CastFromZone"), CardDefIssueInvalidRuleEffect, "play-from-zone permission must set a source zone")
		}
		if effect.TopCardOnly && effect.CastFromZone != zone.Library {
			v.add(faceName, appendPath(path, "TopCardOnly"), CardDefIssueInvalidRuleEffect, "top-card-only play permission requires the library source zone")
		}
	case RuleEffectCastSpellsFromZone:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "cast-from-zone permission must set affected player")
		}
		if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "cast-from-zone permission cannot affect a permanent")
		}
		if effect.CastFromZone == zone.None {
			v.add(faceName, appendPath(path, "CastFromZone"), CardDefIssueInvalidRuleEffect, "cast-from-zone permission must set a source zone")
		}
		if effect.TopCardOnly && effect.CastFromZone != zone.Library {
			v.add(faceName, appendPath(path, "TopCardOnly"), CardDefIssueInvalidRuleEffect, "top-card-only cast permission requires the library source zone")
		}
	case RuleEffectPlayerProtection:
		if effect.AffectedPlayer == PlayerAny {
			v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "player protection must set affected player")
		}
		if effect.AffectedSource || effect.AffectedAttached {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "player protection cannot affect a permanent")
		}
		if !effect.Protection.Everything ||
			len(effect.Protection.FromColors) != 0 ||
			len(effect.Protection.FromTypes) != 0 ||
			len(effect.Protection.FromSubtypes) != 0 ||
			effect.Protection.Multicolored ||
			effect.Protection.Monocolored ||
			effect.Protection.EachColor {
			v.add(faceName, appendPath(path, "Protection"), CardDefIssueInvalidRuleEffect, "player protection currently supports only protection from everything")
		}
	case RuleEffectAdditionalTriggerForChosenCreatureType:
		payload := *effect
		payload.Kind = RuleEffectNone
		if !reflect.DeepEqual(payload, RuleEffect{}) {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "chosen-type trigger multiplier does not accept additional payload")
		}
	case RuleEffectCantCastSpells, RuleEffectCantActivateAbilities:
		v.validateActionRestrictionRuleEffect(faceName, path, effect)
	case RuleEffectCantCastFromZones:
		v.validateCastZoneRestrictionRuleEffect(faceName, path, effect)
	case RuleEffectCantEnterFromZones:
		v.validateEnterZoneRestrictionRuleEffect(faceName, path, effect)
	case RuleEffectAdditionalTriggerForEnteringPermanent:
		payload := *effect
		payload.Kind = RuleEffectNone
		payload.PermanentTypes = nil
		if !reflect.DeepEqual(payload, RuleEffect{}) {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "entering-permanent trigger multiplier accepts only a permanent-type filter")
		}
	default:
	}
}

// validateActionRestrictionRuleEffect checks a cast- or activation-prohibition
// rule effect. These prohibitions target players, never permanents, and the
// permanent-type filter only constrains the activation prohibition.
func (v *cardDefValidator) validateActionRestrictionRuleEffect(faceName, path string, effect *RuleEffect) {
	if !effect.AffectedPlayer.Valid() {
		v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "action restriction must set a recognized affected player")
	}
	if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "action restriction cannot affect a permanent")
	}
	if effect.Kind == RuleEffectCantCastSpells && len(effect.PermanentTypes) != 0 {
		v.add(faceName, appendPath(path, "PermanentTypes"), CardDefIssueInvalidRuleEffect, "cast prohibition does not constrain permanent types")
	}
	if effect.Kind == RuleEffectCantActivateAbilities && len(effect.SpellTypes) != 0 {
		v.add(faceName, appendPath(path, "SpellTypes"), CardDefIssueInvalidRuleEffect, "activation prohibition does not constrain spell types")
	}
	if effect.Kind == RuleEffectCantActivateAbilities && len(effect.ExcludedSpellTypes) != 0 {
		v.add(faceName, appendPath(path, "ExcludedSpellTypes"), CardDefIssueInvalidRuleEffect, "activation prohibition does not constrain spell types")
	}
}

// validateCastZoneRestrictionRuleEffect checks a cast-zone restriction rule
// effect. These prohibitions target players, never permanents, and list one or
// more real non-hand cast zones from which the affected players cannot cast.
func (v *cardDefValidator) validateCastZoneRestrictionRuleEffect(faceName, path string, effect *RuleEffect) {
	if !effect.AffectedPlayer.Valid() {
		v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "cast-zone restriction must set a recognized affected player")
	}
	if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "cast-zone restriction cannot affect a permanent")
	}
	if len(effect.CantCastFromZones) == 0 {
		v.add(faceName, appendPath(path, "CantCastFromZones"), CardDefIssueInvalidRuleEffect, "cast-zone restriction must list at least one zone")
	}
	for _, restricted := range effect.CantCastFromZones {
		if restricted == zone.None || restricted == zone.Hand {
			v.add(faceName, appendPath(path, "CantCastFromZones"), CardDefIssueInvalidRuleEffect, "cast-zone restriction must list real non-hand cast zones")
		}
	}
}

// validateEnterZoneRestrictionRuleEffect checks an enter-the-battlefield zone
// restriction rule effect. The restriction is global (it never targets a player
// or permanent) and lists one or more real zones cards cannot enter the
// battlefield out of.
func (v *cardDefValidator) validateEnterZoneRestrictionRuleEffect(faceName, path string, effect *RuleEffect) {
	if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "enter-zone restriction cannot affect a permanent")
	}
	if len(effect.EnterFromZones) == 0 {
		v.add(faceName, appendPath(path, "EnterFromZones"), CardDefIssueInvalidRuleEffect, "enter-zone restriction must list at least one zone")
	}
	for _, restricted := range effect.EnterFromZones {
		if restricted == zone.None || restricted == zone.Battlefield {
			v.add(faceName, appendPath(path, "EnterFromZones"), CardDefIssueInvalidRuleEffect, "enter-zone restriction must list real source zones")
		}
	}
}

func (v *cardDefValidator) validateAttackTaxRuleEffect(faceName, path string, effect *RuleEffect) {
	if !effect.AffectedPlayer.Valid() || effect.AffectedPlayer == PlayerAny {
		v.add(faceName, appendPath(path, "AffectedPlayer"), CardDefIssueInvalidRuleEffect, "attack tax must set a recognized affected player")
	}
	if effect.AttackTaxGeneric <= 0 {
		v.add(faceName, appendPath(path, "AttackTaxGeneric"), CardDefIssueInvalidRuleEffect, "attack tax must have a positive generic mana amount")
	}
	if effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "attack tax cannot affect a permanent")
	}
}

func (v *cardDefValidator) validateCostModifier(faceName, path string, modifier CostModifier, sourceAbility bool) {
	if sourceAbility && modifier.Kind != CostModifierAbility {
		v.add(faceName, appendPath(path, "Kind"), CardDefIssueInvalidRuleEffect, "source ability cost modifiers must have ability kind")
	}
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
	if !sourceAbility && modifier.Kind == CostModifierAbility && modifier.AbilityKeyword == KeywordNone {
		v.add(faceName, appendPath(path, "AbilityKeyword"), CardDefIssueInvalidRuleEffect, "ability cost modifiers must set AbilityKeyword")
	}
	if modifier.SetManaCost.Exists && modifier.SetGeneric.Exists {
		v.add(faceName, path, CardDefIssueInvalidRuleEffect, "cost modifier cannot set both full mana cost and generic cost")
	}
	if len(modifier.MatchColors) != 0 {
		if modifier.Kind != CostModifierSpell {
			v.add(faceName, appendPath(path, "MatchColors"), CardDefIssueInvalidRuleEffect, "color-disjunction cost modifiers must be spell modifiers")
		}
		if modifier.MatchColor || modifier.MatchCardType {
			v.add(faceName, appendPath(path, "MatchColors"), CardDefIssueInvalidRuleEffect, "color-disjunction cost modifiers cannot also match a single color or card type")
		}
		if len(modifier.MatchColors) < 2 {
			v.add(faceName, appendPath(path, "MatchColors"), CardDefIssueInvalidRuleEffect, "color-disjunction cost modifiers require two or more colors")
		}
		for _, c := range modifier.MatchColors {
			if c == "" {
				v.add(faceName, appendPath(path, "MatchColors"), CardDefIssueInvalidRuleEffect, "color-disjunction cost modifiers require real colors")
			}
		}
	}
	if len(modifier.MatchSubtypes) != 0 {
		if modifier.Kind != CostModifierSpell {
			v.add(faceName, appendPath(path, "MatchSubtypes"), CardDefIssueInvalidRuleEffect, "subtype cost modifiers must be spell modifiers")
		}
		if modifier.MatchCardType || len(modifier.MatchColors) != 0 {
			v.add(faceName, appendPath(path, "MatchSubtypes"), CardDefIssueInvalidRuleEffect, "subtype cost modifiers cannot also match a card type or a color disjunction")
		}
		for _, sub := range modifier.MatchSubtypes {
			if sub == "" {
				v.add(faceName, appendPath(path, "MatchSubtypes"), CardDefIssueInvalidRuleEffect, "subtype cost modifiers require real subtypes")
			}
		}
	}
	if modifier.ChosenSubtypeFromEntryChoice &&
		(modifier.Kind != CostModifierSpell ||
			!modifier.MatchCardType ||
			modifier.CardType != types.Creature ||
			modifier.MatchColor) {
		v.add(faceName, appendPath(path, "ChosenSubtypeFromEntryChoice"), CardDefIssueInvalidRuleEffect, "chosen subtype cost modifier must match creature spells from the entry-time creature-type choice")
	}
	if modifier.SourceZone.Exists {
		if modifier.Kind != CostModifierSpell {
			v.add(faceName, appendPath(path, "SourceZone"), CardDefIssueInvalidRuleEffect, "source-zone cost modifiers must be spell modifiers")
		}
		if modifier.SourceZone.Val == zone.None {
			v.add(faceName, appendPath(path, "SourceZone"), CardDefIssueInvalidRuleEffect, "source-zone cost modifiers require a real zone")
		}
	}
	if modifier.PerObjectReduction < 0 {
		v.add(faceName, appendPath(path, "PerObjectReduction"), CardDefIssueInvalidRuleEffect, "per-object cost reduction cannot be negative")
	}
	if modifier.PerObjectReduction > 0 {
		if modifier.Kind != CostModifierSpell && (!sourceAbility || modifier.Kind != CostModifierAbility) {
			v.add(faceName, path, CardDefIssueInvalidRuleEffect, "per-object cost reduction requires a spell modifier or a source ability modifier")
		}
		if modifier.CountSelection == nil || modifier.CountSelection.Empty() {
			v.add(faceName, appendPath(path, "CountSelection"), CardDefIssueInvalidRuleEffect, "per-object cost reduction requires a count selection")
		} else {
			v.validateSelection(faceName, appendPath(path, "CountSelection"), *modifier.CountSelection)
		}
	} else if modifier.CountSelection != nil && !modifier.CountSelection.Empty() {
		v.add(faceName, appendPath(path, "CountSelection"), CardDefIssueInvalidRuleEffect, "count selection requires a per-object reduction")
	}
	if modifier.DynamicReduction != nil {
		if modifier.Kind != CostModifierSpell {
			v.add(faceName, appendPath(path, "DynamicReduction"), CardDefIssueInvalidRuleEffect, "dynamic cost reduction requires a spell modifier")
		}
		if modifier.PerObjectReduction > 0 {
			v.add(faceName, appendPath(path, "DynamicReduction"), CardDefIssueInvalidRuleEffect, "dynamic cost reduction cannot combine with a per-object reduction")
		}
		if !dynamicCostReductionKindSupported(modifier.DynamicReduction.Kind) {
			v.add(faceName, appendPath(path, "DynamicReduction"), CardDefIssueInvalidRuleEffect, "dynamic cost reduction amount kind is unsupported")
		}
	}
}

// dynamicCostReductionKindSupported reports whether a dynamic amount kind can be
// evaluated at cost time without a resolving stack object, so it may scale a
// source-spell DynamicReduction. Only controller-aggregate and battlefield-group
// kinds qualify; object-referencing kinds (target/source power, counters) need a
// resolving effect and fail closed.
func dynamicCostReductionKindSupported(kind DynamicAmountKind) bool {
	switch kind {
	case DynamicAmountCountSelector,
		DynamicAmountGreatestPowerInGroup,
		DynamicAmountGreatestToughnessInGroup,
		DynamicAmountGreatestManaValueInGroup,
		DynamicAmountTotalPowerInGroup,
		DynamicAmountTotalToughnessInGroup,
		DynamicAmountTotalManaValueInGroup,
		DynamicAmountControllerLife,
		DynamicAmountControllerHandSize,
		DynamicAmountControllerGraveyardSize,
		DynamicAmountControllerBasicLandTypeCount,
		DynamicAmountOpponentCount,
		DynamicAmountDevotion:
		return true
	default:
		return false
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
	if condition.ControllerLifeAtMost.Exists && condition.ControllerLifeAtMost.Val < 0 {
		v.add(faceName, appendPath(path, "ControllerLifeAtMost"), CardDefIssueInvalidCondition, "life threshold cannot be negative")
	}
	if condition.ControllerLifeAtLeastAboveStarting < 0 {
		v.add(faceName, appendPath(path, "ControllerLifeAtLeastAboveStarting"), CardDefIssueInvalidCondition, "life threshold cannot be negative")
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
	if condition.ControllerLibrarySizeAtLeast < 0 {
		v.add(faceName, appendPath(path, "ControllerLibrarySizeAtLeast"), CardDefIssueInvalidCondition, "library-size threshold cannot be negative")
	}
	if condition.ControllerLifeExactly.Exists && condition.ControllerLifeExactly.Val < 0 {
		v.add(faceName, appendPath(path, "ControllerLifeExactly"), CardDefIssueInvalidCondition, "life threshold cannot be negative")
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
	if condition.ControlComparison.Exists {
		v.validateControlCountComparison(faceName, appendPath(path, "ControlComparison"), condition.ControlComparison.Val)
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

func (v *cardDefValidator) validateControlCountComparison(faceName, path string, cmp ControlCountComparison) {
	v.validateSelection(faceName, appendPath(path, "Selection"), cmp.Selection)
	if cmp.Selection.Player != PlayerAny {
		v.add(faceName, appendPath(path, "Selection.Player"), CardDefIssueInvalidSelection, "control-comparison Selection cannot use a player relation")
	}
	if (cmp.Left == ControlPlayerController) == (cmp.Right == ControlPlayerController) {
		v.add(faceName, path, CardDefIssueInvalidCondition, "control comparison must contrast the controller with an opponent scope")
	}
	if cmp.Op != compare.GreaterThan && cmp.Op != compare.LessThan {
		v.add(faceName, appendPath(path, "Op"), CardDefIssueInvalidCondition, "control comparison operator must be a strict greater/less comparison")
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
		unsupported.ExcludedSubtype = ""
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
		unsupported.SubtypeChoice = SubtypeChoiceWithoutEntry(unsupported.SubtypeChoice)
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
			unsupported.ExcludedSubtype = ""
			unsupported.SubtypeChoice = SubtypeChoiceWithoutEntry(unsupported.SubtypeChoice)
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
	if pattern.AttackedPlayerHasMostLife && pattern.Event != EventAttackerDeclared {
		v.add(faceName, appendPath(path, "AttackedPlayerHasMostLife"), CardDefIssueInvalidSelection, "attacked-player-has-most-life trigger filter is only supported for attacker-declared events")
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
	if pattern.ExcludeFirstDrawInDrawStep && pattern.Event != EventCardDrawn {
		v.add(faceName, appendPath(path, "ExcludeFirstDrawInDrawStep"), CardDefIssueInvalidSelection, "first-draw-in-draw-step exclusion is only supported for card-drawn events")
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
	if pattern.AttackWhileSaddled && pattern.Event != EventAttackerDeclared {
		v.add(faceName, appendPath(path, "AttackWhileSaddled"), CardDefIssueInvalidSelection, "attacks-while-saddled trigger filter is only supported for attacker-declared events")
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
	case ObjectReferenceTargetPermanent, ObjectReferenceTargetStackObject, ObjectReferenceTargetObject:
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
	case PlayerReferenceCapturedTargetController:
		v.validateTargetIndex(faceName, path, ref.TargetIndex(), targets, "captured target controller")
		if specIndex, ok := targetSpecForSlot(targets, ref.TargetIndex()); ok &&
			targetSpecAllowedKinds(&targets[specIndex]) != TargetAllowStackObject {
			v.add(faceName, path, CardDefIssueInvalidReference, "captured target controller requires a stack-object target")
		}
	case PlayerReferenceObjectController, PlayerReferenceObjectOwner:
		if object, ok := ref.Object(); ok {
			v.validateObjectRefBounds(faceName, appendPath(path, "Object"), object, targets)
		}
	default:
	}
}

func (v *cardDefValidator) validateCardCondition(faceName, path string, condition CardCondition) {
	v.validateCardRef(faceName, appendPath(path, "Card"), condition.Card)
	if condition.ChosenSubtypeFrom != "" && condition.ChosenSubtypeFrom != EntryTypeChoiceKey {
		v.add(faceName, appendPath(path, "ChosenSubtypeFrom"), CardDefIssueInvalidCondition, "chosen subtype condition must use the entry-time creature-type choice")
	}
	if !condition.RequirePermanentCard && len(condition.Types) == 0 && len(condition.Supertypes) == 0 && len(condition.SubtypesAny) == 0 && condition.ChosenSubtypeFrom == "" {
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
	case TokenCopySourceEachInGroup:
		if spec.Group == nil {
			v.add(faceName, appendPath(path, "Group"), CardDefIssueInvalidReference, "token copy for-each group is nil")
			return
		}
		v.validateGroupRef(faceName, appendPath(path, "Group"), *spec.Group, targets)
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
