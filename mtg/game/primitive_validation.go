package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
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
	case ObjectReferenceTargetPermanent, ObjectReferenceTargetStackObject, ObjectReferenceTargetAttachedPermanent:
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

func validatePlayerGroupReference(ref PlayerGroupReference) error {
	return firstProblem(ref.Validate())
}

func validateCounterSourceSpec(source CounterSourceSpec, targets []TargetSpec, checkTargets bool) error {
	switch source.Kind {
	case CounterSourceNone, CounterSourceEventPermanent:
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
		return validateObjectReference(dynamic.Object, targets, checkTargets)
	case DynamicAmountCountSelector:
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

func (p Discard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
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
	if p.SpendRider.Exists {
		rider := p.SpendRider.Val
		// The condition enum is closed: reject every value other than the fully
		// modeled conditions so an unrecognized or out-of-range condition fails
		// validation rather than silently resolving as a no-op rider.
		switch rider.Condition {
		case ManaSpendCastCommanderCreatureType:
		default:
			return errors.New("add mana spend rider requires a recognized condition")
		}
		if len(rider.Effect.Sequence) == 0 {
			return errors.New("add mana spend rider requires a rider effect")
		}
		// A fired rider is put on the stack with no targets of its own, so a
		// targeted rider effect could never choose a legal target. Reject riders
		// that declare target specs, and validate the sequence against an empty
		// target set so any instruction referencing a target is rejected too.
		if len(rider.Effect.Targets) > 0 {
			return errors.New("add mana spend rider effect must not declare targets")
		}
		if err := ValidateInstructionSequence(rider.Effect.Sequence, rider.Effect.Targets); err != nil {
			return err
		}
	}
	return nil
}

func (p AddCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
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
	if p.Object.Exists {
		return validateObjectReference(p.Object.Val, targets, checkTargets)
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
	if p.Spec.SourceZone != zone.Library ||
		p.Spec.Destination != zone.Hand && p.Spec.Destination != zone.Battlefield {
		return errors.New("search only supports library-to-hand and library-to-battlefield")
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
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Reveal) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
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

func (p PutOnBattlefield) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Source.Valid() {
		return errors.New("put on battlefield requires a valid source")
	}
	if ref, ok := p.Source.CardRef(); ok {
		if err := validateCardReference(ref); err != nil {
			return err
		}
		if err := validateTargetCardReference(ref, targets, checkTargets); err != nil {
			return err
		}
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

func (p Pay) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateResolutionPayment(p.Payment, targets, checkTargets)
}

func validateResolutionPayment(payment ResolutionPayment, targets []TargetSpec, checkTargets bool) error {
	if payment.Payer.Exists {
		return validatePlayerReference(payment.Payer.Val, targets, checkTargets)
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

func (p Exile) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
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
	if hasCard == hasPlayer {
		return errors.New("move card requires exactly one of Card or Player")
	}
	if hasCard {
		if err := validateCardReference(p.Card); err != nil {
			return err
		}
		if err := validateTargetCardReference(p.Card, targets, checkTargets); err != nil {
			return err
		}
	} else {
		if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
			return err
		}
		if p.DestinationBottom {
			return errors.New("player-zone move must not request bottom placement")
		}
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
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Untap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
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

func (Manifest) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
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
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Regenerate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p SkipStep) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p CreateEmblem) validatePrimitive([]TargetSpec, bool) error {
	if len(p.EmblemAbilities) == 0 {
		return errors.New("create emblem requires at least one ability")
	}
	return nil
}

func (p CreateDelayedTrigger) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	switch p.Trigger.Timing {
	case DelayedAtBeginningOfNextEndStep, DelayedAtBeginningOfNextUpkeep:
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
	if hasObject == hasPlayer {
		return errors.New("prevent damage requires exactly one of Object or Player")
	}
	if hasObject {
		return validateObjectReference(p.Object, targets, checkTargets)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func validateNestedAbilityContent(content AbilityContent, inheritedLinked map[LinkedKey]int) error {
	for i := range content.Modes {
		targets := append([]TargetSpec(nil), content.SharedTargets...)
		targets = append(targets, content.Modes[i].Targets...)
		if err := validateInstructionSequenceWithLinked(content.Modes[i].Sequence, targets, true, inheritedLinked); err != nil {
			return fmt.Errorf("mode %d: %w", i, err)
		}
	}
	return nil
}
