package game

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/game/zone"
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
		if dynamic.ResultKey == "" {
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
	if p.Selector != EffectSelectorNone {
		return nil
	}
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
	if p.Spec.SourceZone == zone.None || p.Spec.Destination == zone.None {
		return errors.New("search requires source and destination zones")
	}
	if p.Spec.SourceZone != zone.Library ||
		p.Spec.Destination != zone.Hand && p.Spec.Destination != zone.Battlefield {
		return errors.New("search only supports library-to-hand and library-to-battlefield")
	}
	if p.Spec.Supertype.Exists && p.Spec.Supertype.Val == "" {
		return errors.New("search supertype cannot be empty")
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
		if err := validatePlayerReference(p.Recipient.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	if p.PublishLinked != "" && !p.Amount.IsDynamic() && p.Amount.Value() != 1 {
		return errors.New("linked reveal must reveal exactly one card")
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

func (p GainLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p LoseLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Exile) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Selector != EffectSelectorNone {
		return nil
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Bounce) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Selector != EffectSelectorNone {
		return nil
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Sacrifice) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Untap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Selector != EffectSelectorNone {
		return nil
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p CounterObject) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Mill) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Scry) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Surveil) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
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

func (Proliferate) validatePrimitive([]TargetSpec, bool) error {
	return nil
}

func (p Goad) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p RemoveCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Selector != EffectSelectorNone {
		return nil
	}
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Transform) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p PhaseOut) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p Regenerate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p SkipStep) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func (p CreateEmblem) validatePrimitive([]TargetSpec, bool) error {
	if len(p.EmblemAbilities) == 0 {
		return errors.New("create emblem requires at least one ability")
	}
	return nil
}

func (p CreateDelayedTrigger) validatePrimitive([]TargetSpec, bool) error {
	if p.Trigger.Timing == 0 {
		return errors.New("delayed trigger requires timing")
	}
	return validateNestedAbilityContent(p.Trigger.Content)
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
	return validateTargetReference(p.TargetIndex, targets, checkTargets)
}

func validateNestedAbilityContent(content AbilityContent) error {
	switch c := content.(type) {
	case PlainAbilityContent:
		return ValidateInstructionSequence(c.Sequence, c.Targets)
	case ModalAbilityContent:
		for i := range c.Modes {
			targets := c.Modes[i].Targets
			if len(targets) == 0 {
				targets = c.SharedTargets
			}
			if err := ValidateInstructionSequence(c.Modes[i].Sequence, targets); err != nil {
				return fmt.Errorf("mode %d: %w", i, err)
			}
		}
		return nil
	case nil:
		return errors.New("delayed trigger requires content")
	default:
		return fmt.Errorf("unsupported delayed-trigger content %T", content)
	}
}
