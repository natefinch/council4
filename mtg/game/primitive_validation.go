package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func validateTargetReference(index int, targets []TargetSpec, checkTargets bool) error {
	if !checkTargets || index < 0 {
		return nil
	}
	if index >= targetSlotCapacity(targets) {
		return fmt.Errorf("target index %d has no matching target specification", index)
	}
	return nil
}

// targetSlotCapacity returns the total number of target slots the specs admit.
// Object, player, and stack-object references address chosen targets by a flat
// slot index across all specs, and a single spec with MaxTargets > 1 admits that
// many consecutive slots, so a reference index is in range when it is below this
// capacity. Every spec contributes at least one slot, so the capacity never
// drops below len(targets) and every index that was valid under the previous
// one-slot-per-spec rule stays valid.
func targetSlotCapacity(targets []TargetSpec) int {
	total := 0
	for i := range targets {
		width := max(targets[i].MaxTargets, 1)
		total += width
	}
	return total
}

// targetSpecForSlot maps a flat target slot index to the spec that admits it,
// mirroring how the rules resolve chosen targets across multi-target specs. It
// reports false when the index is negative or beyond the combined capacity.
func targetSpecForSlot(targets []TargetSpec, index int) (int, bool) {
	if index < 0 {
		return 0, false
	}
	cumulative := 0
	for i := range targets {
		width := max(targets[i].MaxTargets, 1)
		if index < cumulative+width {
			return i, true
		}
		cumulative += width
	}
	return 0, false
}

func validateTargetAllows(index int, allow TargetAllow, targets []TargetSpec, checkTargets bool) error {
	if err := validateTargetReference(index, targets, checkTargets); err != nil {
		return err
	}
	if !checkTargets {
		return nil
	}
	specIndex, ok := targetSpecForSlot(targets, index)
	if !ok {
		return fmt.Errorf("target index %d has no matching target specification", index)
	}
	if targetSpecAllowedKinds(&targets[specIndex]) != allow {
		return errors.New("target specification allows an incompatible target kind")
	}
	return nil
}

func targetSpecAllowedKinds(target *TargetSpec) TargetAllow {
	if target.Allow != TargetAllowUnspecified {
		return target.Allow
	}
	constraint := strings.ToLower(strings.TrimSpace(target.Constraint))
	constraint = strings.TrimPrefix(constraint, "target ")
	constraint = strings.Join(strings.Fields(constraint), " ")
	if constraint == "any target" {
		return TargetAllowPermanent | TargetAllowPlayer
	}
	switch constraint {
	case "player", "opponent":
		return TargetAllowPlayer
	}
	if strings.Contains(constraint, "permanent") ||
		strings.Contains(constraint, "creature") ||
		strings.Contains(constraint, "artifact") ||
		strings.Contains(constraint, "enchantment") ||
		strings.Contains(constraint, "land") ||
		strings.Contains(constraint, "planeswalker") ||
		strings.Contains(constraint, "battle") {
		return TargetAllowPermanent
	}
	return TargetAllowUnspecified
}

// firstProblem adapts the structural []string problem list returned by the
// reference Validate methods to the error-returning sequence validator.
func firstProblem(problems []string) error {
	if len(problems) == 0 {
		return nil
	}
	return errors.New(problems[0])
}

func validateObjectReference(ref ObjectReference, targets []TargetSpec, checkTargets bool) error {
	if err := firstProblem(ref.Validate()); err != nil {
		return err
	}
	return validateObjectReferenceTargetBounds(ref, targets, checkTargets)
}

// validateObjectReferenceTargetBounds performs the contextual target-slot bounds
// check for the target-derived object reference kinds. Structural consistency is
// owned by ObjectReference.Validate.
func validateObjectReferenceTargetBounds(ref ObjectReference, targets []TargetSpec, checkTargets bool) error {
	switch ref.Kind() {
	case ObjectReferenceTargetPermanent, ObjectReferenceTargetStackObject, ObjectReferenceTargetAttachedPermanent, ObjectReferenceTargetObject:
		return validateTargetReference(ref.TargetIndex(), targets, checkTargets)
	}
	return nil
}

func validatePlayerReference(ref PlayerReference, targets []TargetSpec, checkTargets bool) error {
	if err := firstProblem(ref.Validate()); err != nil {
		return err
	}
	switch ref.Kind() {
	case PlayerReferenceTargetPlayer:
		return validateTargetReference(ref.TargetIndex(), targets, checkTargets)
	case PlayerReferenceObjectController, PlayerReferenceObjectOwner:
		object, _ := ref.Object()
		return validateObjectReferenceTargetBounds(object, targets, checkTargets)
	}
	return nil
}

type capturedTargetControllerReferenceValidator interface {
	validateCapturedTargetControllerReferences([]TargetSpec, bool) error
}

func validateCapturedTargetControllerReference(ref PlayerReference, targets []TargetSpec, checkTargets bool) error {
	if ref.Kind() != PlayerReferenceCapturedTargetController {
		return nil
	}
	return validateTargetAllows(ref.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
}

func validateCapturedTargetControllerReferenceList(
	targets []TargetSpec,
	checkTargets bool,
	references ...PlayerReference,
) error {
	for _, ref := range references {
		if err := validateCapturedTargetControllerReference(ref, targets, checkTargets); err != nil {
			return err
		}
	}
	return nil
}

func validateCapturedTargetControllerQuantity(quantity Quantity, targets []TargetSpec, checkTargets bool) error {
	if !quantity.IsDynamic() {
		return nil
	}
	dynamic := quantity.DynamicAmount().Val
	if dynamic.Player != nil {
		if err := validateCapturedTargetControllerReference(*dynamic.Player, targets, checkTargets); err != nil {
			return err
		}
	}
	if dynamic.Object.Kind() == ObjectReferenceCapturedTargetStackObject {
		return validateTargetAllows(dynamic.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
	}
	return nil
}

func validateCapturedTargetControllerQuantities(
	targets []TargetSpec,
	checkTargets bool,
	quantities ...Quantity,
) error {
	for _, quantity := range quantities {
		if err := validateCapturedTargetControllerQuantity(quantity, targets, checkTargets); err != nil {
			return err
		}
	}
	return nil
}

func validateCapturedTargetControllerOptionalReference(
	reference opt.V[PlayerReference],
	targets []TargetSpec,
	checkTargets bool,
) error {
	if !reference.Exists {
		return nil
	}
	return validateCapturedTargetControllerReference(reference.Val, targets, checkTargets)
}

func (p Damage) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	var references []PlayerReference
	if ref, ok := p.Recipient.PlayerReference(); ok {
		references = append(references, ref)
	}
	if ref, ok := p.Recipient.AnyTargetPlayerReference(); ok {
		references = append(references, ref)
	}
	if err := validateCapturedTargetControllerReferenceList(targets, checkTargets, references...); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Draw) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p ReorderLibraryTop) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p LookAtLibraryTop) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p ShuffleLibrary) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p LookAtHand) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p Discard) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p AddMana) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p AddCounter) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p AddPlayerCounter) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p MoveCounters) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p ModifyPT) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantities(targets, checkTargets, p.PowerDelta, p.ToughnessDelta)
}

func (p Search) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if err := validateCapturedTargetControllerOptionalReference(p.Controller, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Reveal) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if p.Card.Kind != CardReferenceNone {
		return nil
	}
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if err := validateCapturedTargetControllerOptionalReference(p.Recipient, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p PutOnBattlefield) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerOptionalReference(p.Recipient, targets, checkTargets)
}

func (p CreateToken) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerOptionalReference(p.Recipient, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p StartEngines) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p SetClassLevel) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Monstrosity) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p DiscoverCards) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Amass) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Renown) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Pay) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerOptionalReference(p.Payment.Payer, targets, checkTargets)
}

func (p Choose) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if p.Choice.PlayerReference == nil {
		return nil
	}
	return validateCapturedTargetControllerReference(*p.Choice.PlayerReference, targets, checkTargets)
}

func (p GainLife) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p LoseLife) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p MoveCard) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p SacrificePermanents) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Mill) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p ExileTopOfLibrary) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p PutHandOnLibraryThenDraw) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p RevealUntil) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p Scry) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Surveil) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Dig) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantities(targets, checkTargets, p.Look, p.Take)
}

func (p Investigate) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerOptionalReference(p.Recipient, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Proliferate) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p SkipStep) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p PreventDamage) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func validatePlayerGroupReference(ref PlayerGroupReference) error {
	return firstProblem(ref.Validate())
}

func validateCounterSourceSpec(source CounterSourceSpec, targets []TargetSpec, checkTargets bool) error {
	switch source.Kind {
	case CounterSourceNone, CounterSourceEventPermanent, CounterSourceSelf:
		return nil
	case CounterSourceTarget:
		return validateObjectReference(source.Object, targets, checkTargets)
	default:
		return errors.New("counter source has unknown kind")
	}
}

func validateMassObjectOrGroup(object ObjectReference, group GroupReference, targets []TargetSpec, checkTargets bool) error {
	hasObject := object.Kind() != ObjectReferenceNone
	hasGroup := group.Domain() != groupDomainNone
	if hasObject == hasGroup {
		return errors.New("mass-effect primitive requires exactly one of Object or Group")
	}
	if hasObject {
		return validateObjectReference(object, targets, checkTargets)
	}
	return validateGroupReference(group, targets, checkTargets)
}

func validateCardReference(ref CardReference) error {
	switch ref.Kind {
	case CardReferenceLinked:
		if ref.LinkID == "" {
			return errors.New("linked card reference requires LinkID")
		}
	case CardReferenceSource, CardReferenceEvent, CardReferenceTarget:
		if ref.LinkID != "" {
			return errors.New("source/event/target card reference must not set LinkID")
		}
		if ref.Kind != CardReferenceTarget && ref.TargetIndex != 0 {
			return errors.New("source/event card reference must not set TargetIndex")
		}
		if ref.TargetIndex < 0 {
			return errors.New("target card reference must not use a negative TargetIndex")
		}
	case CardReferenceNone:
		return errors.New("card reference has no kind")
	default:
		return fmt.Errorf("unknown card reference kind %d", ref.Kind)
	}
	return nil
}

// validateGroupReference checks structural validity of a GroupReference and
// recursively checks target-slot bounds for its anchor and exclusion
// ObjectReferences. group.Validate() already checks structural consistency
// of nested references; validateGroupReference adds the contextual bounds check.
func validateGroupReference(group GroupReference, targets []TargetSpec, checkTargets bool) error {
	if err := firstProblem(group.Validate()); err != nil {
		return err
	}
	if anchor, ok := group.Anchor(); ok {
		if err := validateObjectReferenceTargetBounds(anchor, targets, checkTargets); err != nil {
			return fmt.Errorf("anchor: %w", err)
		}
	}
	if exclude, ok := group.Exclusion(); ok {
		if err := validateObjectReferenceTargetBounds(exclude, targets, checkTargets); err != nil {
			return fmt.Errorf("exclusion: %w", err)
		}
	}
	return nil
}

func validateQuantity(quantity Quantity, targets []TargetSpec, checkTargets bool) error {
	if !quantity.IsDynamic() {
		return nil
	}
	dynamic := quantity.DynamicAmount().Val
	switch dynamic.Kind {
	case DynamicAmountNone:
		return errors.New("dynamic quantity has no kind")
	case DynamicAmountTargetPower, DynamicAmountTargetToughness, DynamicAmountTargetManaValue, DynamicAmountTargetCounters, DynamicAmountObjectPower, DynamicAmountObjectToughness, DynamicAmountObjectManaValue:
		if dynamic.Object.Kind() == ObjectReferenceCapturedTargetStackObject {
			return errors.New("captured target stack object reference requires a captured target mana value amount")
		}
		return validateObjectReference(dynamic.Object, targets, checkTargets)
	case DynamicAmountCapturedTargetManaValue:
		if dynamic.Object.Kind() != ObjectReferenceCapturedTargetStackObject {
			return errors.New("captured target mana value requires a captured target stack object reference")
		}
		return validateObjectReference(dynamic.Object, targets, checkTargets)
	case DynamicAmountObjectCounters:
		if !dynamic.CounterKind.Valid() {
			return errors.New("object-counter count requires a valid counter kind")
		}
		return validateObjectReference(dynamic.Object, targets, checkTargets)
	case DynamicAmountCountSelector, DynamicAmountGreatestPowerInGroup, DynamicAmountGreatestToughnessInGroup, DynamicAmountGreatestManaValueInGroup,
		DynamicAmountTotalPowerInGroup, DynamicAmountTotalToughnessInGroup, DynamicAmountColorCountInGroup,
		DynamicAmountSharedCreatureTypeCountInGroup:
		return validateGroupReference(dynamic.Group, targets, checkTargets)
	case DynamicAmountCountCardsInZone:
		if dynamic.CardZone == zone.None || dynamic.CardZone == zone.Battlefield || dynamic.CardZone == zone.Stack {
			return errors.New("card-zone count requires a non-battlefield zone")
		}
		if dynamic.Selection == nil {
			return errors.New("card-zone count requires a selection")
		}
		if dynamic.Player == nil {
			return errors.New("card-zone count requires a player")
		}
		return validatePlayerReference(*dynamic.Player, targets, checkTargets)
	case DynamicAmountPreviousEffectResult, DynamicAmountPreviousEffectExcessDamage:
		if dynamic.ResultKey == "" {
			return errors.New("previous-result quantity requires a result key")
		}
	case DynamicAmountChosenNumber:
		if dynamic.ResultKey == "" {
			return errors.New("chosen-number quantity requires a choice key")
		}
	case DynamicAmountDevotion:
		if len(dynamic.Colors) == 0 && dynamic.ColorFrom == "" {
			return errors.New("devotion quantity requires at least one color or a chosen-color source")
		}
	case DynamicAmountMaxOf:
		if len(dynamic.Operands) < 2 {
			return errors.New("max-of quantity requires at least two operands")
		}
		for i := range dynamic.Operands {
			if err := validateQuantity(Dynamic(dynamic.Operands[i]), targets, checkTargets); err != nil {
				return err
			}
		}
	default:
	}
	return nil
}

func validatePositiveQuantity(quantity Quantity, targets []TargetSpec, checkTargets bool) error {
	if !quantity.IsDynamic() && quantity.Value() <= 0 {
		return errors.New("counter amount must be positive")
	}
	return validateQuantity(quantity, targets, checkTargets)
}

func (p GroupSourceDamage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateGroupReference(p.Group, targets, checkTargets); err != nil {
		return err
	}
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (p Damage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Recipient.Valid() {
		return errors.New("damage requires a valid recipient")
	}
	if p.Divided {
		object, ok := p.Recipient.AnyTargetObjectReference()
		if !ok {
			return errors.New("divided damage requires an any-target recipient")
		}
		if checkTargets {
			specIndex := object.TargetIndex()
			if specIndex < 0 || specIndex >= len(targets) {
				return errors.New("divided damage references an out-of-range target spec")
			}
			if targets[specIndex].MaxTargets < 1 {
				return errors.New("divided damage requires a target spec that admits at least one target")
			}
		}
	}
	if object, ok := p.Recipient.ObjectReference(); ok {
		if err := validateObjectReference(object, targets, checkTargets); err != nil {
			return err
		}
	}
	if player, ok := p.Recipient.PlayerReference(); ok {
		if err := validatePlayerReference(player, targets, checkTargets); err != nil {
			return err
		}
	}
	if object, ok := p.Recipient.AnyTargetObjectReference(); ok {
		if err := validateObjectReference(object, targets, checkTargets); err != nil {
			return err
		}
	}
	if player, ok := p.Recipient.AnyTargetPlayerReference(); ok {
		if err := validatePlayerReference(player, targets, checkTargets); err != nil {
			return err
		}
	}
	if group, ok := p.Recipient.GroupReference(); ok {
		if err := firstProblem(group.Validate()); err != nil {
			return err
		}
	}
	if group, ok := p.Recipient.PlayerGroupReference(); ok {
		if err := validatePlayerGroupReference(group); err != nil {
			return err
		}
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.DamageSource.Exists {
		return validateObjectReference(p.DamageSource.Val, targets, checkTargets)
	}
	return nil
}

func (p Draw) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("Draw requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ReorderLibraryTop) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() < 1 {
		return errors.New("ReorderLibraryTop requires a positive number of cards")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p LookAtLibraryTop) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.PublishLinked == "" {
		return errors.New("LookAtLibraryTop requires PublishLinked")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ShuffleLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p LookAtHand) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Discard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.EntireHand && (p.Amount.IsDynamic() || p.Amount.Value() != 0) {
		return errors.New("Discard with EntireHand must not set Amount")
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("Discard requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Destroy) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p AddMana) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Player.Exists {
		if err := validatePlayerReference(p.Player.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	if p.SpendRider.Exists {
		if err := validateManaSpendRider(p.SpendRider.Val); err != nil {
			return err
		}
	}
	return nil
}

func validateManaSpendRider(rider ManaSpendRider) error {
	if len(rider.Effect.Targets) > 0 {
		return errors.New("add mana spend rider effect must not declare targets")
	}
	switch rider.Condition {
	case ManaSpendCastCommanderCreatureType:
		if rider.Restriction != ManaSpendUnrestricted ||
			rider.ChosenSubtypeFrom != "" ||
			rider.SpellRuleEffect != RuleEffectNone ||
			len(rider.SpellGainsKeywords) != 0 ||
			len(rider.Effect.Sequence) == 0 {
			return errors.New("commander-type mana spend rider has unsupported fields")
		}
	case ManaSpendCastChosenCreatureType:
		if rider.Restriction != ManaSpendRestrictedToCondition ||
			rider.ChosenSubtypeFrom != EntryTypeChoiceKey ||
			(rider.SpellRuleEffect != RuleEffectCantBeCountered && rider.SpellRuleEffect != RuleEffectNone) ||
			len(rider.SpellGainsKeywords) != 0 ||
			len(rider.Effect.Sequence) != 0 {
			return errors.New("chosen-type mana spend rider has unsupported fields")
		}
		return nil
	case ManaSpendCastOrActivateChosenCreatureType:
		if rider.Restriction != ManaSpendRestrictedToCondition ||
			rider.ChosenSubtypeFrom != EntryTypeChoiceKey ||
			rider.SpellRuleEffect != RuleEffectNone ||
			len(rider.SpellGainsKeywords) != 0 ||
			len(rider.Effect.Sequence) != 0 {
			return errors.New("chosen-type cast-or-activate mana spend rider has unsupported fields")
		}
		return nil
	case ManaSpendCastLegendarySpell:
		if rider.Restriction != ManaSpendRestrictedToCondition ||
			rider.ChosenSubtypeFrom != "" ||
			(rider.SpellRuleEffect != RuleEffectCantBeCountered && rider.SpellRuleEffect != RuleEffectNone) ||
			len(rider.SpellGainsKeywords) != 0 ||
			len(rider.Effect.Sequence) != 0 {
			return errors.New("legendary-spell mana spend rider has unsupported fields")
		}
		return nil
	case ManaSpendCastCreatureSpell:
		if rider.Restriction != ManaSpendUnrestricted ||
			rider.ChosenSubtypeFrom != "" ||
			rider.SpellRuleEffect != RuleEffectNone ||
			len(rider.Effect.Sequence) != 0 ||
			len(rider.SpellGainsKeywords) == 0 {
			return errors.New("creature-spell mana spend rider has unsupported fields")
		}
		return nil
	case ManaSpendCastArtifactSpell:
		if rider.Restriction != ManaSpendRestrictedToCondition ||
			rider.ChosenSubtypeFrom != "" ||
			rider.SpellRuleEffect != RuleEffectNone ||
			len(rider.SpellGainsKeywords) != 0 ||
			len(rider.Effect.Sequence) != 0 {
			return errors.New("artifact-spell mana spend rider has unsupported fields")
		}
		return nil
	default:
		return errors.New("add mana spend rider requires a recognized condition")
	}
	return ValidateInstructionSequence(rider.Effect.Sequence, rider.Effect.Targets)
}

func (p AddCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.AllKinds {
		if p.Group.Domain() != groupDomainNone {
			return errors.New("add counter doubling every kind requires a single object, not a group")
		}
		if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
			return err
		}
		if p.Object.Kind() == ObjectReferenceTargetPermanent {
			return validateTargetAllows(p.Object.TargetIndex(), TargetAllowPermanent, targets, checkTargets)
		}
		return nil
	}
	if !p.CounterKind.Valid() {
		return errors.New("add counter requires a recognized counter kind")
	}
	if p.CounterKind.PlayerOnly() {
		return errors.New("player-only counter kind cannot be placed on a permanent")
	}
	if err := validatePositiveQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Group.Domain() != groupDomainNone {
		if p.Object.Kind() != ObjectReferenceNone {
			return errors.New("add counter requires exactly one of Object or Group")
		}
		return validateGroupReference(p.Group, targets, checkTargets)
	}
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	if p.Object.Kind() == ObjectReferenceTargetPermanent {
		return validateTargetAllows(p.Object.TargetIndex(), TargetAllowPermanent, targets, checkTargets)
	}
	return nil
}

func (p AddPlayerCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.CounterKind.Valid() {
		return errors.New("add player counter requires a recognized counter kind")
	}
	if !p.CounterKind.PlayerOnly() {
		return errors.New("permanent-only counter kind cannot be placed on a player")
	}
	if err := validatePositiveQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.Player.Kind() == PlayerReferenceTargetPlayer {
		return validateTargetAllows(p.Player.TargetIndex(), TargetAllowPlayer, targets, checkTargets)
	}
	return nil
}

func (p MoveCounters) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Distribute {
		if p.Group == nil {
			return errors.New("distributed move counters requires a destination group")
		}
		if p.Object.Kind() != ObjectReferenceNone {
			return errors.New("distributed move counters cannot also target a single object")
		}
		if err := validateGroupReference(*p.Group, targets, checkTargets); err != nil {
			return err
		}
		return validateCounterSourceSpec(p.Source, targets, checkTargets)
	}
	if p.Group != nil {
		return errors.New("move counters requires a single Object unless Distribute is set")
	}
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	return validateCounterSourceSpec(p.Source, targets, checkTargets)
}

func (p ApplyContinuous) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if len(p.ContinuousEffects) == 0 {
		return errors.New("continuous effect instruction has no declarations")
	}
	if p.PublishLinked != "" &&
		(!p.Object.Exists || p.Object.Val.Kind() != ObjectReferenceTargetPermanent) {
		return errors.New("linked continuous effect must target a permanent")
	}
	if p.Object.Exists {
		return validateObjectReference(p.Object.Val, targets, checkTargets)
	}
	return nil
}

func (p ApplyRule) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if len(p.RuleEffects) == 0 {
		return errors.New("rule effect instruction has no declarations")
	}
	for i := range p.RuleEffects {
		effect := &p.RuleEffects[i]
		if !effect.Kind.Valid() {
			return errors.New("rule effect has an unsupported kind")
		}
		switch effect.Kind {
		case RuleEffectLifeTotalCantChange, RuleEffectNoMaximumHandSize:
			if effect.AffectedPlayer == PlayerAny {
				return errors.New("player rule effect requires an affected player")
			}
			if p.Object.Exists || effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
				return errors.New("player rule effect cannot affect a permanent")
			}
		case RuleEffectPlayerProtection:
			if p.Object.Exists || effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
				return errors.New("player protection cannot affect a permanent")
			}
			if effect.AffectedPlayer == PlayerAny {
				return errors.New("player protection requires an affected player")
			}
			if !effect.Protection.Everything ||
				len(effect.Protection.FromColors) != 0 ||
				len(effect.Protection.FromTypes) != 0 ||
				len(effect.Protection.FromSubtypes) != 0 ||
				effect.Protection.Multicolored ||
				effect.Protection.Monocolored ||
				effect.Protection.EachColor {
				return errors.New("player protection supports only protection from everything")
			}
		case RuleEffectAttackTax:
			if err := validateApplyRuleAttackTax(effect, p.Object.Exists); err != nil {
				return err
			}
		case RuleEffectPlayFromZone:
			if err := validatePlayFromZoneRuleEffect(effect, p.Object.Exists, false); err != nil {
				return err
			}
		default:
		}
	}
	if p.Object.Exists {
		return validateObjectReference(p.Object.Val, targets, checkTargets)
	}
	return nil
}

func validateApplyRuleAttackTax(effect *RuleEffect, hasObject bool) error {
	if !effect.AffectedPlayer.Valid() || effect.AffectedPlayer == PlayerAny {
		return errors.New("attack tax requires a recognized affected player")
	}
	if effect.AttackTaxGeneric <= 0 {
		return errors.New("attack tax requires a positive generic mana amount")
	}
	if hasObject || effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
		return errors.New("attack tax cannot affect a permanent")
	}
	return nil
}

func validatePlayFromZoneRuleEffect(effect *RuleEffect, objectScoped, cardDef bool) error {
	if !effect.AffectedPlayer.Valid() {
		return errors.New("play-from-zone rule effect requires a recognized affected player")
	}
	if effect.CastFromZone != zone.Exile {
		return errors.New("play-from-zone rule effect requires exile as its source zone")
	}
	if !cardDef && effect.AffectedCardID == 0 {
		return errors.New("play-from-zone rule effect requires a specific card")
	}
	if cardDef && effect.AffectedCardID != 0 {
		return errors.New("play-from-zone card definitions cannot embed a runtime card ID")
	}
	if effect.CastFace.Exists {
		return errors.New("play-from-zone rule effect cannot restrict the card face")
	}
	if objectScoped || effect.AffectedSource || effect.AffectedAttached || effect.AffectedObjectID != 0 {
		return errors.New("play-from-zone rule effect cannot affect a permanent")
	}
	return nil
}

func (p ModifyPT) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.PowerDelta, targets, checkTargets); err != nil {
		return err
	}
	if err := validateQuantity(p.ToughnessDelta, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Fight) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.RelatedObject, targets, checkTargets)
}

func (p Tap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p Search) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Spec.SourceZone == zone.None || p.Spec.Destination == zone.None {
		return errors.New("search requires source and destination zones")
	}
	if p.Spec.SourceZone != zone.Library || !validSearchDestination(SearchDestination{
		Zone:         p.Spec.Destination,
		Position:     p.Spec.DestinationPosition,
		EntersTapped: p.Spec.EntersTapped,
	}) {
		return errors.New("search has unsupported source or destination")
	}
	if p.Spec.Destination == zone.Library &&
		(p.Amount.IsDynamic() || p.Amount.Value() != 1 || p.Spec.SplitDestination.Exists) {
		return errors.New("library-top search requires exactly one card and no split destination")
	}
	switch p.Spec.FailToFindPolicy {
	case SearchFailToFindDefault:
	case SearchMayFailToFind:
		if !p.Amount.IsDynamic() && p.Amount.Value() == 1 && p.Spec.IsUnrestricted() {
			return errors.New("singular unrestricted library search cannot allow fail-to-find")
		}
	case SearchMustFindIfAvailable:
		if p.Amount.IsDynamic() || p.Amount.Value() != 1 || !p.Spec.IsUnrestricted() {
			return errors.New("required library search must find exactly one unrestricted card")
		}
	default:
		return errors.New("search has unsupported fail-to-find policy")
	}
	if p.Spec.SplitDestination.Exists &&
		(!validSearchDestination(p.Spec.SplitDestination.Val) ||
			p.Spec.SplitDestination.Val.Zone == zone.Library) {
		return errors.New("search has unsupported split destination")
	}
	if p.Spec.CardType.Exists && len(p.Spec.CardTypesAny) != 0 {
		return errors.New("search cannot combine one required card type with a card-type union")
	}
	if len(p.Spec.CardTypesAny) == 1 {
		return errors.New("search card-type union requires at least two card types")
	}
	if slices.Contains(p.Spec.CardTypesAny, "") {
		return errors.New("search card-type union cannot contain an empty type")
	}
	if p.Spec.Supertype.Exists && p.Spec.Supertype.Val == "" {
		return errors.New("search supertype cannot be empty")
	}
	if p.PublishLinked != "" &&
		(p.Amount.IsDynamic() ||
			p.Amount.Value() != 1 ||
			p.Spec.Destination != zone.Battlefield ||
			p.Spec.SplitDestination.Exists) {
		return errors.New("linked search requires exactly one card moved to the battlefield")
	}
	if p.Controller.Exists {
		if p.Spec.Destination != zone.Battlefield {
			return errors.New("search controller applies only to a battlefield destination")
		}
		if err := validatePlayerReference(p.Controller.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("Search requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		if p.Controller.Exists {
			return errors.New("Search with PlayerGroup cannot set a controller")
		}
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func validSearchDestination(destination SearchDestination) bool {
	switch destination.Zone {
	case zone.Hand, zone.Graveyard:
		return destination.Position == SearchPositionUnspecified && !destination.EntersTapped
	case zone.Battlefield:
		return destination.Position == SearchPositionUnspecified
	case zone.Library:
		return destination.Position == SearchPositionTop && !destination.EntersTapped
	default:
		return false
	}
}

func (p Reveal) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Card.Kind != CardReferenceNone {
		if p.Player.Kind() != PlayerReferenceNone ||
			p.Recipient.Exists ||
			p.PublishLinked != "" ||
			p.Amount.IsDynamic() ||
			p.Amount.Value() != 0 {
			return errors.New("card reveal cannot set player, recipient, amount, or publish a link")
		}
		return validateCardReference(p.Card)
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.Recipient.Exists {
		if err := validatePlayerReference(p.Recipient.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	if p.PublishLinked != "" && !p.Amount.IsDynamic() && p.Amount.Value() != 1 {
		return errors.New("linked reveal must reveal exactly one card")
	}
	return nil
}

func (p ExileFromHand) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Amount.IsDynamic() || p.Amount.Value() < 1 {
		return errors.New("exile from hand requires a fixed positive amount")
	}
	if p.PublishLinked != "" && p.Amount.Value() != 1 {
		return errors.New("linked exile from hand must exile exactly one card")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p PutFromHand) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Amount.IsDynamic() || p.Amount.Value() < 1 {
		return errors.New("put from hand requires a fixed positive amount")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p CastForFree) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Zone == zone.None {
		return errors.New("cast for free requires a source zone")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ReturnFromGraveyard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Amount.IsDynamic() || p.Amount.Value() < 1 {
		return errors.New("return from graveyard requires a fixed positive amount")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p MassReturnFromGraveyard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Destination != zone.Hand && p.Destination != zone.Battlefield {
		return errors.New("mass return from graveyard requires a hand or battlefield destination")
	}
	if p.EntryTapped && p.Destination != zone.Battlefield {
		return errors.New("mass return from graveyard tapped entry requires a battlefield destination")
	}
	if p.ControlledByOwner && p.Destination != zone.Battlefield {
		return errors.New("mass return from graveyard owner control requires a battlefield destination")
	}
	if p.SourceGroup.Kind != PlayerGroupReferenceNone {
		if problems := p.SourceGroup.Validate(); len(problems) != 0 {
			return errors.New(problems[0])
		}
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p MassReanimationExchange) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if len(p.Selection.RequiredTypes) != 1 {
		return errors.New("mass reanimation exchange requires a single card-type filter")
	}
	return nil
}

func (p PutOnBattlefield) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Source.Valid() == (len(p.Sources) > 0) {
		return errors.New("put on battlefield requires a valid source")
	}
	sources := p.Sources
	if p.Source.Valid() {
		sources = []BattlefieldSource{p.Source}
	}
	for _, source := range sources {
		if !source.Valid() {
			return errors.New("put on battlefield requires valid sources")
		}
		ref, ok := source.CardRef()
		if !ok {
			if len(p.Sources) > 0 {
				return errors.New("simultaneous put on battlefield requires referenced-card sources")
			}
			if p.PublishLinked != "" {
				return errors.New("put on battlefield can publish only a referenced card")
			}
			continue
		}
		if err := validateCardReference(ref); err != nil {
			return err
		}
		if err := validateTargetCardReference(ref, targets, checkTargets); err != nil {
			return err
		}
	}
	if len(p.Sources) > 0 && p.PublishLinked != "" {
		return errors.New("simultaneous put on battlefield cannot publish one linked permanent")
	}
	if len(p.Sources) > 0 && len(p.ContinuousEffects) > 0 {
		return errors.New("simultaneous put on battlefield does not support continuous effects")
	}
	if p.Recipient.Exists {
		if err := validatePlayerReference(p.Recipient.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	for _, placement := range p.EntryCounters {
		if placement.Amount <= 0 {
			return errors.New("put on battlefield entry counters require a positive amount")
		}
	}
	return nil
}

func (p CreateToken) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Source.Valid() {
		return errors.New("create token requires a valid source")
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if spec, ok := p.Source.TokenCopy(); ok && spec.Source == TokenCopySourceObject {
		if err := validateObjectReference(spec.Object, targets, checkTargets); err != nil {
			return err
		}
	}
	if p.Recipient.Exists {
		return validatePlayerReference(p.Recipient.Val, targets, checkTargets)
	}
	return nil
}

func (p ShufflePermanentIntoLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p PutPermanentOnLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p StartEngines) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p SetClassLevel) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Monstrosity) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p DiscoverCards) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (p Amass) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (p Renown) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p BecomeSaddled) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (ShuffleSpellIntoLibrary) validatePrimitive(_ []TargetSpec, _ bool) error {
	return nil
}

func (p Pay) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateResolutionPayment(p.Payment, targets, checkTargets)
}

func validateResolutionPayment(payment ResolutionPayment, targets []TargetSpec, checkTargets bool) error {
	if payment.Payer.Exists {
		if err := validatePlayerReference(payment.Payer.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	if payment.DynamicGenericManaCost.Exists {
		if payment.ManaCost.Exists || payment.ManaCostMultiplier.Exists {
			return errors.New("resolution payment cannot combine fixed and dynamic generic or multiplied mana costs")
		}
		if payment.DynamicGenericManaCost.Val == nil {
			return errors.New("resolution payment has nil dynamic generic mana cost")
		}
		if err := validateQuantity(Dynamic(*payment.DynamicGenericManaCost.Val), targets, checkTargets); err != nil {
			return fmt.Errorf("dynamic generic mana cost: %w", err)
		}
	}
	if payment.ManaCostMultiplier.Exists {
		if !payment.ManaCost.Exists {
			return errors.New("resolution payment mana multiplier requires a fixed mana cost")
		}
		if payment.ManaCostMultiplier.Val == nil {
			return errors.New("resolution payment has nil mana cost multiplier")
		}
		if err := validateQuantity(Dynamic(*payment.ManaCostMultiplier.Val), targets, checkTargets); err != nil {
			return fmt.Errorf("mana cost multiplier: %w", err)
		}
	}
	if !payment.ManaCost.Exists && !payment.DynamicGenericManaCost.Exists && !payment.ManaCostMultiplier.Exists && len(payment.AdditionalCosts) == 0 {
		return errors.New("resolution payment has no cost")
	}
	return nil
}

func (p Choose) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Choice.Kind == ResolutionChoiceNone {
		return errors.New("choose instruction has no choice kind")
	}
	if p.Choice.UsePlayer && p.Choice.PlayerReference != nil {
		return errors.New("resolution choice cannot set both Player and PlayerReference")
	}
	if p.Choice.PlayerReference != nil {
		if err := validatePlayerReference(*p.Choice.PlayerReference, targets, checkTargets); err != nil {
			return fmt.Errorf("resolution choice player: %w", err)
		}
	}
	if p.Choice.Kind == ResolutionChoiceNumber &&
		(p.Choice.MinNumber < 0 || p.Choice.MaxNumber < p.Choice.MinNumber) {
		return errors.New("number choice requires a nonnegative inclusive range")
	}
	return nil
}

func (p GainLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("GainLife requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p LoseLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("LoseLife requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p PlayerLosesGame) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Player.Kind() == PlayerReferenceNone {
		return errors.New("PlayerLosesGame requires a Player reference")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p PlayerWinsGame) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Player.Kind() == PlayerReferenceNone {
		return errors.New("PlayerWinsGame requires a Player reference")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p PunisherEachLoseLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.PlayerGroup.Kind == PlayerGroupReferenceNone {
		return errors.New("PunisherEachLoseLife requires a PlayerGroup reference")
	}
	if !p.AllowSacrifice && !p.AllowDiscard {
		return errors.New("PunisherEachLoseLife requires at least one alternative cost")
	}
	if err := firstProblem(p.SacrificeSelection.Validate()); err != nil {
		return err
	}
	return validatePlayerGroupReference(p.PlayerGroup)
}

func (p RepeatProcess) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Times, targets, checkTargets); err != nil {
		return err
	}
	if len(p.Body.Modes) == 0 {
		return errors.New("RepeatProcess requires a body")
	}
	return validateNestedAbilityContent(p.Body, nil, targets, checkTargets)
}

func (p Exile) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.SourceSpell {
		if p.Object.Kind() != ObjectReferenceNone || p.Group.Valid() || p.ExileLinkedKey != "" {
			return errors.New("source-spell exile cannot set an object, group, or linked key")
		}
		return nil
	}
	if p.ExileLinkedKey != "" && p.Group.Valid() {
		return errors.New("linked exile requires one object")
	}
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p Bounce) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.ControlledChoice {
		if p.Object.Kind() != ObjectReferenceNone {
			return errors.New("controlled-choice bounce must not set Object")
		}
		if !p.Group.Valid() {
			return errors.New("controlled-choice bounce requires a candidate Group")
		}
		return nil
	}
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p MoveCard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	hasCard := p.Card.Kind != CardReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	set := 0
	for _, present := range []bool{hasCard, hasPlayer, hasGroup} {
		if present {
			set++
		}
	}
	if set != 1 {
		return errors.New("move card requires exactly one of Card, Player, or PlayerGroup")
	}
	if hasGroup {
		if err := validatePlayerGroupReference(p.PlayerGroup); err != nil {
			return err
		}
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("player-group move must not set Amount")
		}
		if p.DestinationBottom {
			return errors.New("player-group move must not request bottom placement")
		}
	} else if err := p.validateMoveReference(hasCard, targets, checkTargets); err != nil {
		return err
	}
	if p.FromZone == zone.None || p.FromZone == zone.Battlefield || p.FromZone == zone.Stack {
		return errors.New("move card requires a non-battlefield source zone")
	}
	if p.Destination == zone.None || p.Destination == zone.Battlefield || p.Destination == zone.Stack {
		return errors.New("move card requires a non-battlefield destination zone")
	}
	if p.FromZone == p.Destination {
		return errors.New("move card requires different source and destination zones")
	}
	if p.DestinationBottom && p.Destination != zone.Library {
		return errors.New("bottom placement requires library as destination zone")
	}
	return nil
}

func (p MoveCard) validateMoveReference(hasCard bool, targets []TargetSpec, checkTargets bool) error {
	if hasCard {
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("single-card move must not set Amount")
		}
		if err := validateCardReference(p.Card); err != nil {
			return err
		}
		return validateTargetCardReference(p.Card, targets, checkTargets)
	}
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
		if err := validatePositiveQuantity(p.Amount, targets, checkTargets); err != nil {
			return fmt.Errorf("chosen-card move amount: %w", err)
		}
		if p.FromZone != zone.Hand || p.Destination != zone.Library {
			return errors.New("chosen-card move requires hand to library")
		}
	}
	if p.DestinationBottom {
		return errors.New("player-zone move must not request bottom placement")
	}
	return nil
}

func (p MoveCommander) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Destination == zone.None || p.Destination == zone.Battlefield ||
		p.Destination == zone.Stack || p.Destination == zone.Command {
		return errors.New("move commander requires a non-battlefield destination zone")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func validateTargetCardReference(ref CardReference, targets []TargetSpec, checkTargets bool) error {
	if ref.Kind != CardReferenceTarget {
		return nil
	}
	if ref.TargetIndex < 0 {
		return errors.New("target card reference must not use a negative TargetIndex")
	}
	if !checkTargets || len(targets) == 0 {
		return errors.New("target card reference requires a target specification")
	}
	cardSlot := 0
	for i := range targets {
		target := targets[i]
		maxTargets := target.MaxTargets
		if maxTargets == 0 {
			continue
		}
		if targetSpecAllowedKinds(&target)&TargetAllowCard == 0 {
			continue
		}
		if ref.TargetIndex < cardSlot+maxTargets {
			return nil
		}
		cardSlot += maxTargets
	}
	if cardSlot == 0 {
		return errors.New("target card reference requires a card target specification")
	}
	return fmt.Errorf("target card reference index %d has no matching target specification", ref.TargetIndex)
}

func (p GrantCastPermission) validatePrimitive([]TargetSpec, bool) error {
	if err := validateCardReference(p.Card); err != nil {
		return err
	}
	if p.FromZone != zone.Graveyard {
		return errors.New("cast permission requires graveyard as its source zone")
	}
	if p.Duration != DurationUntilEndOfYourNextTurn {
		return errors.New("cast permission requires a supported bounded duration")
	}
	return nil
}

func (p Sacrifice) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Object.Kind() == ObjectReferenceNone {
		return nil
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p SacrificePermanents) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("SacrificePermanents requires exactly one of Player or PlayerGroup")
	}
	if err := p.Fallback.validate(targets, checkTargets); err != nil {
		return err
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p SacrificeFallback) validate(targets []TargetSpec, checkTargets bool) error {
	switch p.Kind {
	case SacrificeFallbackNone:
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("SacrificeFallbackNone requires a zero Amount")
		}
		return nil
	case SacrificeFallbackDiscard, SacrificeFallbackLoseLife:
		return validateQuantity(p.Amount, targets, checkTargets)
	default:
		return fmt.Errorf("unknown SacrificeFallbackKind %d", p.Kind)
	}
}

func (p Untap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets); err != nil {
		return err
	}
	if !p.ChooseUpTo {
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("untap Amount requires ChooseUpTo")
		}
		return nil
	}
	if p.Object.Kind() != ObjectReferenceNone {
		return errors.New("bounded untap requires a Group rather than an Object")
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() <= 0 {
		return errors.New("bounded untap requires a positive Amount")
	}
	return nil
}

func (p SkipNextUntap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p CounterObject) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	if p.Object.Kind() != ObjectReferenceTargetStackObject {
		return errors.New("counter object requires a target stack object reference")
	}
	return validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
}

func (p ChooseNewTargets) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	if p.Object.Kind() != ObjectReferenceTargetStackObject {
		return errors.New("choose new targets requires a target stack object reference")
	}
	return validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
}

func (p CopyStackObject) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	switch p.Object.Kind() {
	case ObjectReferenceTargetStackObject:
		return validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
	case ObjectReferenceEventStackObject:
		return nil
	default:
		return errors.New("copy stack object requires a target or event stack object reference")
	}
}

func (p Mill) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("Mill requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExileTopOfLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("ExileTopOfLibrary requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p PutHandOnLibraryThenDraw) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.DrawOffset < 0 {
		return errors.New("PutHandOnLibraryThenDraw requires a non-negative DrawOffset")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p RevealUntil) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Destination != zone.Graveyard && p.Destination != zone.Hand {
		return errors.New("RevealUntil requires a Graveyard or Hand Destination")
	}
	if err := firstProblem(p.Until.Validate()); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return errors.New("RevealUntil requires exactly one of Player or PlayerGroup")
	}
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Scry) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Surveil) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Dig) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Look, targets, checkTargets); err != nil {
		return err
	}
	if err := validateQuantity(p.Take, targets, checkTargets); err != nil {
		return err
	}
	if !p.Look.IsDynamic() && p.Look.Value() < 1 {
		return errors.New("Dig requires looking at a positive number of cards")
	}
	if !p.Take.IsDynamic() && p.Take.Value() < 1 {
		return errors.New("Dig requires taking a positive number of cards")
	}
	if !p.Look.IsDynamic() && !p.Take.IsDynamic() && p.Take.Value() > p.Look.Value() {
		return errors.New("Dig cannot take more cards than it looks at")
	}
	switch p.Remainder {
	case DigRemainderGraveyard, DigRemainderLibraryBottom:
	default:
		return errors.New("Dig has an unknown remainder destination")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ImpulseExile) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() < 1 {
		return errors.New("ImpulseExile requires a positive number of cards")
	}
	if p.Duration != DurationThisTurn && p.Duration != DurationUntilEndOfTurn &&
		p.Duration != DurationUntilEndOfYourNextTurn {
		return errors.New("ImpulseExile requires a this-turn, until-end-of-turn, or until-end-of-your-next-turn play window")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Investigate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Recipient.Exists {
		return validatePlayerReference(p.Recipient.Val, targets, checkTargets)
	}
	return nil
}

func (p Proliferate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (p Explore) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Creature, targets, checkTargets)
}

func (p Manifest) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Player.Kind() != PlayerReferenceNone {
		return validatePlayerReference(p.Player, targets, checkTargets)
	}
	return nil
}

func (p Goad) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p RemoveCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p Transform) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p PhaseOut) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p Regenerate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p BecomeCopy) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Card.Kind != CardReferenceNone {
		if p.Object.Kind() != ObjectReferenceNone {
			return errors.New("become copy must set only one of Object or Card")
		}
		if err := validateCardReference(p.Card); err != nil {
			return err
		}
		return validateTargetCardReference(p.Card, targets, checkTargets)
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Attach) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Attachment, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Target, targets, checkTargets)
}

func (p SkipStep) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p AddExtraPhases) validatePrimitive([]TargetSpec, bool) error {
	if !p.Combat && !p.Main {
		return errors.New("add extra phases requires at least one phase")
	}
	if p.Main && !p.Combat {
		return errors.New("add extra phases main phase must follow an extra combat phase")
	}
	return nil
}

func (p CreateEmblem) validatePrimitive([]TargetSpec, bool) error {
	if len(p.EmblemAbilities) == 0 {
		return errors.New("create emblem requires at least one ability")
	}
	return nil
}

func (p CreateDelayedTrigger) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	switch p.Trigger.Timing {
	case DelayedAtBeginningOfNextEndStep, DelayedAtBeginningOfNextUpkeep, DelayedAtBeginningOfNextMainPhase:
	default:
		return errors.New("delayed trigger requires a recognized timing")
	}
	if len(p.Trigger.Content.Modes) == 0 {
		return errors.New("delayed trigger requires content")
	}
	return nil
}

func (p CreateReplacement) validatePrimitive([]TargetSpec, bool) error {
	if p.Replacement == nil {
		return errors.New("create replacement requires a replacement")
	}
	if p.Replacement.MatchEvent == EventUnknown {
		return errors.New("create replacement requires an event")
	}
	return nil
}

func (p PreventDamage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	hasObject := p.Object.Kind() != ObjectReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	if p.Global {
		if hasObject || hasPlayer || p.BySource {
			return errors.New("global prevent damage must not set Object, Player, or BySource")
		}
		return nil
	}
	if hasObject == hasPlayer {
		return errors.New("prevent damage requires exactly one of Object or Player")
	}
	if hasObject {
		return validateObjectReference(p.Object, targets, checkTargets)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func validateNestedAbilityContent(
	content AbilityContent,
	inheritedLinked map[LinkedKey]int,
	capturedTargets []TargetSpec,
	checkCapturedTargets bool,
) error {
	for i := range content.Modes {
		targets := append([]TargetSpec(nil), content.SharedTargets...)
		targets = append(targets, content.Modes[i].Targets...)
		if err := validateInstructionSequenceWithLinked(
			content.Modes[i].Sequence,
			targets,
			true,
			inheritedLinked,
			capturedTargets,
			checkCapturedTargets,
		); err != nil {
			return fmt.Errorf("mode %d: %w", i, err)
		}
	}
	return nil
}
