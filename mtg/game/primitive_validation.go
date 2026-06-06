package game

import (
	"errors"
	"fmt"
)

func validateTargetReference(index int, targets []TargetSpec, checkTargets bool) error {
	if !checkTargets || index < 0 {
		return nil
	}
	if index >= len(targets) {
		return fmt.Errorf("target index %d has no matching target specification", index)
	}
	return nil
}

func validateObjectReference(ref ObjectReference, targets []TargetSpec, checkTargets bool) error {
	switch ref.Kind {
	case ObjectReferenceTargetPermanent:
		return validateTargetReference(ref.TargetIndex, targets, checkTargets)
	case ObjectReferenceSourcePermanent, ObjectReferenceEventPermanent:
		if ref.TargetIndex != 0 || ref.LinkID != "" {
			return fmt.Errorf("object reference kind %d cannot set target index or link ID", ref.Kind)
		}
	case ObjectReferenceAttachedPermanent:
		if ref.TargetIndex >= 0 || ref.LinkID != "" {
			return errors.New("attached permanent reference requires a negative source target index")
		}
	case ObjectReferenceLinkedObject:
		if ref.LinkID == "" {
			return errors.New("linked object reference requires a link ID")
		}
	default:
		return errors.New("object reference has no kind")
	}
	return nil
}

func validatePlayerReference(ref PlayerReference, targets []TargetSpec, checkTargets bool) error {
	switch ref.Kind {
	case PlayerReferenceController:
		return nil
	case PlayerReferenceTargetPlayer:
		return validateTargetReference(ref.TargetIndex, targets, checkTargets)
	case PlayerReferenceObjectController, PlayerReferenceObjectOwner:
		if !ref.Object.Exists {
			return fmt.Errorf("player reference kind %d requires an object", ref.Kind)
		}
		return validateObjectReference(ref.Object.Val, targets, checkTargets)
	default:
		return errors.New("player reference has no kind")
	}
}

func validateQuantity(quantity Quantity, targets []TargetSpec, checkTargets bool) error {
	if !quantity.IsDynamic() {
		return nil
	}
	dynamic := quantity.DynamicAmount().Val
	switch dynamic.Kind {
	case DynamicAmountNone:
		return errors.New("dynamic quantity has no kind")
	case DynamicAmountTargetPower, DynamicAmountTargetToughness, DynamicAmountTargetManaValue, DynamicAmountTargetCounters:
		return validateTargetReference(dynamic.TargetIndex, targets, checkTargets)
	case DynamicAmountPreviousEffectResult, DynamicAmountPreviousEffectExcessDamage:
		if dynamic.ResultKey == "" && dynamic.LinkID == "" {
			return errors.New("previous-result quantity requires a result key")
		}
	case DynamicAmountObjectPower:
		return validateObjectReference(dynamic.Object, targets, checkTargets)
	default:
	}
	return nil
}

func (p Damage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Recipient.Valid() {
		return errors.New("damage requires a valid recipient")
	}
	if index, ok := p.Recipient.TargetIndex(); ok {
		if err := validateTargetReference(index, targets, checkTargets); err != nil {
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
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Discard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Destroy) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p AddMana) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (p AddCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p MoveCounters) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateTargetReference(p.TargetIndex, targets, checkTargets); err != nil {
		return err
	}
	if p.Source.Kind == CounterSourceTarget {
		return validateTargetReference(p.Source.TargetIndex, targets, checkTargets)
	}
	return nil
}

func (p ApplyContinuous) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if len(p.ContinuousEffects) == 0 {
		return errors.New("continuous effect instruction has no declarations")
	}
	if p.Object.Exists {
		return validateObjectReference(p.Object.Val, targets, checkTargets)
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p ApplyRule) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if len(p.RuleEffects) == 0 {
		return errors.New("rule effect instruction has no declarations")
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

//nolint:gocritic // Value receivers are required by the sealed Primitive interface.
func (p ModifyPT) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.PowerDelta, targets, checkTargets); err != nil {
		return err
	}
	if err := validateQuantity(p.ToughnessDelta, targets, checkTargets); err != nil {
		return err
	}
	if p.Object.Exists {
		return validateObjectReference(p.Object.Val, targets, checkTargets)
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Fight) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateTargetReference(p.TargetIndex, targets, checkTargets); err != nil {
		return err
	}
	if p.RelatedTargetIndex.Exists {
		return validateTargetReference(p.RelatedTargetIndex.Val, targets, checkTargets)
	}
	return nil
}

func (p Tap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Search) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Spec.SourceZone == ZoneNone || p.Spec.Destination == ZoneNone {
		return errors.New("search requires source and destination zones")
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Reveal) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateTargetReference(p.TargetIndex, targets, checkTargets); err != nil {
		return err
	}
	if p.Recipient.Exists {
		return validatePlayerReference(p.Recipient.Val, targets, checkTargets)
	}
	return nil
}

func (p PutOnBattlefield) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Source.Valid() {
		return errors.New("put on battlefield requires a valid source")
	}
	if err := validateTargetReference(p.TargetIndex, targets, checkTargets); err != nil {
		return err
	}
	if p.Recipient.Exists {
		return validatePlayerReference(p.Recipient.Val, targets, checkTargets)
	}
	return nil
}

//nolint:gocritic // Value receivers are required by the sealed Primitive interface.
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
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p StartEngines) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p SetClassLevel) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Monstrosity) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p DiscoverCards) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (Pay) validatePrimitive([]TargetSpec, bool) error {
	return nil
}

func (p Choose) validatePrimitive([]TargetSpec, bool) error {
	if p.Choice.Kind == ResolutionChoiceNone {
		return errors.New("choose instruction has no choice kind")
	}
	return nil
}
