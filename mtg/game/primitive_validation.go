package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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
	case ObjectReferenceAllTargetPermanents, ObjectReferenceAllTargetStackObjects:
		return validateTargetSpecReference(ref.TargetIndex(), targets, checkTargets)
	default:
		// Non-target-addressed references have no target bounds to check.
	}
	return nil
}

// validateTargetSpecReference bounds-checks a reference that addresses a whole
// target spec by its index (not a flat target slot), used by the group-blink
// all-target-permanents reference. The index must name a declared spec.
func validateTargetSpecReference(specIndex int, targets []TargetSpec, checkTargets bool) error {
	if specIndex < 0 {
		return fmt.Errorf("target spec index %d is negative", specIndex)
	}
	if checkTargets && specIndex >= len(targets) {
		return fmt.Errorf("target spec index %d has no matching target specification", specIndex)
	}
	return nil
}

// validateTargetSpecAllows bounds-checks a whole-spec reference by its spec index
// and confirms the named spec targets the required kind. It backs the all-target
// group references (e.g. "exile any number of target spells") that address an
// entire target spec at once rather than a single flat target slot.
func validateTargetSpecAllows(specIndex int, allow TargetAllow, targets []TargetSpec, checkTargets bool) error {
	if err := validateTargetSpecReference(specIndex, targets, checkTargets); err != nil {
		return err
	}
	if !checkTargets {
		return nil
	}
	if targetSpecAllowedKinds(&targets[specIndex]) != allow {
		return errors.New("target specification allows an incompatible target kind")
	}
	return nil
}

func validatePlayerReference(ref PlayerReference, targets []TargetSpec, checkTargets bool) error {
	if err := firstProblem(ref.Validate()); err != nil {
		return err
	}
	switch ref.Kind() {
	case PlayerReferenceTargetPlayer, PlayerReferenceAffectedTargetController:
		return validateTargetReference(ref.TargetIndex(), targets, checkTargets)
	case PlayerReferenceObjectController, PlayerReferenceObjectOwner:
		object, _ := ref.Object()
		return validateObjectReferenceTargetBounds(object, targets, checkTargets)
	default:
		// Non-target-addressed player references have no target bounds to check.
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

// validateCapturedTargetControllerPlayerAndQuantity is the shared body for the
// many single-player primitives whose only captured-target-controller checks are
// their Player reference followed by their Amount quantity. Collapsing the
// identical reference-then-quantity body keeps each primitive's method a single
// explicit call while still routing through the same per-shape helpers.
func validateCapturedTargetControllerPlayerAndQuantity(
	player PlayerReference,
	quantity Quantity,
	targets []TargetSpec,
	checkTargets bool,
) error {
	if err := validateCapturedTargetControllerReference(player, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(quantity, targets, checkTargets)
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
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p ReorderLibraryTop) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
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

func (p ChooseDiscardFromHand) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p Discard) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p AddMana) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p AddCounter) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p AddPlayerCounter) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
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
	if err := validateCapturedTargetControllerOptionalReference(p.EntryAttackingDefender, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p StartEngines) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p BecomeMonarch) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}
func (p CantBecomeMonarch) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}
func (GainCityBlessing) validateCapturedTargetControllerReferences([]TargetSpec, bool) error {
	return nil
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

func (p Incubate) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	if err := validateCapturedTargetControllerOptionalReference(p.Recipient, targets, checkTargets); err != nil {
		return err
	}
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Bolster) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Renown) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Adapt) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerQuantity(p.Amount, targets, checkTargets)
}

func (p Connive) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
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
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p LoseLife) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p MoveCard) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p SacrificePermanents) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p Mill) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p ExileTopOfLibrary) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p PutHandOnLibraryThenDraw) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p DiscardThenDraw) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p DiscardUnlessType) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p RevealUntil) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}

func (p Scry) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
}

func (p Surveil) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
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
	return validateCapturedTargetControllerPlayerAndQuantity(p.Player, p.Amount, targets, checkTargets)
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

// validateMassPlayerOrGroup is the player twin of validateMassObjectOrGroup: it
// enforces that exactly one of a primitive's Player or PlayerGroup reference is
// set. name is the owning primitive's name so each call site keeps its existing
// error message verbatim. Reference dispatch (validatePlayerReference /
// validatePlayerGroupReference) stays at the call sites because the surrounding
// per-primitive checks differ.
func validateMassPlayerOrGroup(name string, player PlayerReference, group PlayerGroupReference) error {
	hasGroup := group.Kind != PlayerGroupReferenceNone
	hasPlayer := player.Kind() != PlayerReferenceNone
	if hasGroup == hasPlayer {
		return fmt.Errorf("%s requires exactly one of Player or PlayerGroup", name)
	}
	return nil
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
	case CardReferenceCaptured:
		if ref.LinkID != "" {
			return errors.New("captured card reference must not set LinkID")
		}
		if ref.TargetIndex != 0 {
			return errors.New("captured card reference must not set TargetIndex")
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
	if player, ok := group.PlayerAnchor(); ok {
		if err := validatePlayerReference(player, targets, checkTargets); err != nil {
			return fmt.Errorf("player anchor: %w", err)
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
		if dynamic.Player != nil {
			if dynamic.Object.Kind() != ObjectReferenceLinkedObject {
				return errors.New("player-correlated object quantity requires a linked object reference")
			}
			if err := validatePlayerReference(*dynamic.Player, targets, checkTargets); err != nil {
				return err
			}
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
		DynamicAmountTotalPowerInGroup, DynamicAmountTotalToughnessInGroup, DynamicAmountTotalManaValueInGroup, DynamicAmountColorCountInGroup,
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
	case DynamicAmountReferencedCardsTotalManaValue:
		if dynamic.LinkedKey == "" {
			return errors.New("referenced-cards total mana value requires a linked key")
		}
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

func (p GroupSelfPowerDamage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateGroupReference(p.Group, targets, checkTargets)
}

func (p Damage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Recipient.Valid() {
		return errors.New("damage requires a valid recipient")
	}
	if p.Divided && p.EachTarget {
		return errors.New("damage cannot be both divided and dealt to each target")
	}
	if p.Divided || p.EachTarget {
		object, ok := p.Recipient.AnyTargetObjectReference()
		if !ok {
			return errors.New("divided or each-target damage requires an any-target recipient")
		}
		if checkTargets {
			specIndex := object.TargetIndex()
			if specIndex < 0 || specIndex >= len(targets) {
				return errors.New("divided or each-target damage references an out-of-range target spec")
			}
			if targets[specIndex].MaxTargets < 1 {
				return errors.New("divided or each-target damage requires a target spec that admits at least one target")
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
	return p.validateExcessRecipient(targets, checkTargets)
}

// validateExcessRecipient checks an excess-damage redirect: the redirect target
// must be a single player, and the primary recipient must be a single permanent
// (object or any-target), since excess damage is defined only relative to a
// creature's lethal damage. The zero-value (unset) recipient imposes no
// constraints.
func (p Damage) validateExcessRecipient(targets []TargetSpec, checkTargets bool) error {
	if !p.ExcessRecipient.set {
		return nil
	}
	player, ok := p.ExcessRecipient.PlayerReference()
	if !ok {
		return errors.New("excess damage redirect requires a player recipient")
	}
	if err := validatePlayerReference(player, targets, checkTargets); err != nil {
		return err
	}
	if _, ok := p.Recipient.ObjectReference(); ok {
		return nil
	}
	if _, ok := p.Recipient.AnyTargetObjectReference(); ok {
		return nil
	}
	return errors.New("excess damage redirect requires a single-permanent recipient")
}

func (p Draw) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("Draw", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
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

func (p ShuffleGraveyardIntoLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateMassPlayerOrGroup("ShuffleGraveyardIntoLibrary", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	if p.PlayerGroup.Kind != PlayerGroupReferenceNone {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p LookAtHand) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ChooseDiscardFromHand) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p Discard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.EntireHand && (p.Amount.IsDynamic() || p.Amount.Value() != 0) {
		return errors.New("Discard with EntireHand must not set Amount")
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("Discard", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	if p.PublishLinked != "" && (p.EntireHand || hasGroup) {
		return errors.New("Discard with PublishLinked must be a single-player, non-entire-hand discard")
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
	if len(p.CombinationColors) != 0 {
		if err := validateManaCombinationColors(p); err != nil {
			return err
		}
	}
	return nil
}

// validateManaCombinationColors enforces the invariants of an AddMana whose
// produced mana is split freely among a fixed color set. The colors must be two
// or more distinct basic colors, and the combination shape is mutually exclusive
// with every other color-selection mechanism (a fixed color, a linked or
// entry-time color choice, the each-controlled-color union, and spend riders).
func validateManaCombinationColors(p AddMana) error {
	if p.ManaColor != "" || p.ChoiceFrom != "" || p.EntryChoiceFrom != "" ||
		p.EachControlledColor != nil || p.SpendRider.Exists {
		return errors.New("combination add mana cannot combine with ManaColor, ChoiceFrom, EntryChoiceFrom, EachControlledColor, or SpendRider")
	}
	if len(p.CombinationColors) < 2 {
		return errors.New("combination add mana requires at least two colors")
	}
	seen := make(map[mana.Color]bool, len(p.CombinationColors))
	for _, c := range p.CombinationColors {
		switch c {
		case mana.W, mana.U, mana.B, mana.R, mana.G, mana.C:
		default:
			return errors.New("combination add mana has an invalid mana color")
		}
		if seen[c] {
			return errors.New("combination add mana has a duplicate mana color")
		}
		seen[c] = true
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
		// Two modeled shapes share this condition: the unrestricted haste bonus
		// rider (Arena of Glory: spendable on anything, a creature spell paid
		// with it gains haste) and the bare restricted spend rider (Beastcaller
		// Savant: spendable only on creature spells, no further effect).
		if rider.ChosenSubtypeFrom != "" ||
			rider.SpellRuleEffect != RuleEffectNone ||
			len(rider.Effect.Sequence) != 0 {
			return errors.New("creature-spell mana spend rider has unsupported fields")
		}
		switch rider.Restriction {
		case ManaSpendUnrestricted:
			if len(rider.SpellGainsKeywords) == 0 {
				return errors.New("creature-spell mana spend rider has unsupported fields")
			}
		case ManaSpendRestrictedToCondition:
			if len(rider.SpellGainsKeywords) != 0 {
				return errors.New("creature-spell mana spend rider has unsupported fields")
			}
		default:
			return errors.New("creature-spell mana spend rider has unsupported fields")
		}
		return nil
	case ManaSpendCastArtifactSpell,
		ManaSpendCastArtifactSpellOnly,
		ManaSpendCastOrActivateArtifact,
		ManaSpendActivateArtifactAbility,
		ManaSpendCastArtifactOrActivateAbility,
		ManaSpendCastInstantOrSorcerySpell,
		ManaSpendCastNoncreatureSpell,
		ManaSpendCastMulticoloredSpell,
		ManaSpendCastPlaneswalkerSpell,
		ManaSpendCastMonocoloredSpellOfChosenColor,
		ManaSpendCastOrActivateCreature:
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
	if p.Distribute {
		if p.Group.Domain() != groupDomainNone {
			return errors.New("distributed add counter requires a target spec, not a group")
		}
		if p.AllKinds || p.ChooseOne || len(p.KindChoices) != 0 {
			return errors.New("distributed add counter cannot combine with AllKinds, ChooseOne, or KindChoices")
		}
		if p.Object.Kind() != ObjectReferenceAllTargetPermanents {
			return errors.New("distributed add counter requires an all-target-permanents object reference")
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
		if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
			return err
		}
		return validateTargetAllows(p.Object.TargetIndex(), TargetAllowPermanent, targets, checkTargets)
	}
	if p.DoubleKind {
		if p.Group.Domain() == groupDomainNone {
			return errors.New("add counter doubling a kind requires a group")
		}
		if p.Object.Kind() != ObjectReferenceNone {
			return errors.New("add counter doubling a kind requires a group, not an object")
		}
		if p.AllKinds || p.ChooseOne || len(p.KindChoices) != 0 {
			return errors.New("add counter doubling a kind cannot combine with AllKinds, ChooseOne, or KindChoices")
		}
		if !p.CounterKind.Valid() {
			return errors.New("add counter requires a recognized counter kind")
		}
		if p.CounterKind.PlayerOnly() {
			return errors.New("player-only counter kind cannot be placed on a permanent")
		}
		return validateGroupReference(p.Group, targets, checkTargets)
	}
	if p.ChooseOne && p.Group.Domain() == groupDomainNone {
		return errors.New("add counter choosing one recipient requires a group")
	}
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
	if len(p.KindChoices) != 0 {
		if p.Group.Domain() != groupDomainNone {
			return errors.New("add counter with a kind choice requires a single object, not a group")
		}
		if len(p.KindChoices) < 2 {
			return errors.New("add counter kind choice requires two or more kinds")
		}
		seen := make(map[counter.Kind]bool, len(p.KindChoices))
		for _, kind := range p.KindChoices {
			if !kind.Valid() {
				return errors.New("add counter kind choice requires recognized counter kinds")
			}
			if kind.PlayerOnly() {
				return errors.New("player-only counter kind cannot be placed on a permanent")
			}
			if seen[kind] {
				return errors.New("add counter kind choice has duplicate kinds")
			}
			seen[kind] = true
		}
		if err := validatePositiveQuantity(p.Amount, targets, checkTargets); err != nil {
			return err
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
		if p.PublishLinked != "" {
			return errors.New("add counter PublishLinked requires a single object, not a group")
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
	if err := validateMassPlayerOrGroup("AddPlayerCounter", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
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
	for i := range p.ContinuousEffects {
		if err := validateContinuousEffectNewControllerRef(&p.ContinuousEffects[i], targets, checkTargets); err != nil {
			return err
		}
	}
	if p.ChooseFrom.Valid() {
		if p.Object.Exists {
			return errors.New("continuous effect instruction cannot both choose from a group and target an object")
		}
		if p.PublishLinked != "" {
			return errors.New("group-choosing continuous effect cannot publish a single linked object")
		}
		return nil
	}
	if p.PublishLinked != "" &&
		(!p.Object.Exists ||
			(p.Object.Val.Kind() != ObjectReferenceTargetPermanent &&
				p.Object.Val.Kind() != ObjectReferenceLinkedObject)) {
		return errors.New("linked continuous effect must target a permanent or linked object")
	}
	if p.Object.Exists {
		return validateObjectReference(p.Object.Val, targets, checkTargets)
	}
	return nil
}

// validateContinuousEffectNewControllerRef enforces the invariants of a
// LayerControl effect whose new controller is resolved from a player reference
// at application time: it is mutually exclusive with the NewController sentinel,
// requires the control layer, and must reference an in-bounds player.
func validateContinuousEffectNewControllerRef(continuous *ContinuousEffect, targets []TargetSpec, checkTargets bool) error {
	if !continuous.NewControllerRef.Exists {
		return nil
	}
	if continuous.NewController.Exists {
		return errors.New("continuous effect sets both NewController and NewControllerRef")
	}
	if continuous.Layer != LayerControl {
		return errors.New("NewControllerRef requires the control layer")
	}
	return validatePlayerReference(continuous.NewControllerRef.Val, targets, checkTargets)
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
		case RuleEffectLifeTotalCantChange, RuleEffectNoMaximumHandSize, RuleEffectPlayerHexproof, RuleEffectPlayerShroud, RuleEffectLegendRuleDoesNotApply:
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

func (p PlayerMayPayGenericOrRule) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Player.Kind() == PlayerReferenceNone {
		return errors.New("pay-or-rule instruction requires a payer")
	}
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if len(p.RuleEffects) == 0 {
		return errors.New("pay-or-rule instruction has no rule effects")
	}
	for i := range p.RuleEffects {
		if !p.RuleEffects[i].Kind.Valid() {
			return errors.New("pay-or-rule instruction has an unsupported rule effect kind")
		}
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

func (p CorrelatedFight) validatePrimitive([]TargetSpec, bool) error {
	if p.Subjects == "" || p.Objects == "" {
		return errors.New("correlated fight requires both subject and object linked keys")
	}
	if p.Subjects == p.Objects {
		return errors.New("correlated fight subject and object linked keys must differ")
	}
	return nil
}

func (p Tap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p TapOrUntap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Search) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Spec.Name != "" && p.Spec.NameFromChoice != "" {
		return errors.New("search cannot combine fixed and choice-derived names")
	}
	if p.Spec.RevealOnly {
		return p.validateRevealOnlySearch(targets, checkTargets)
	}
	if p.Spec.ExileFaceDown {
		return p.validateExileFaceDownSearch(targets, checkTargets)
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
	if p.Spec.AlsoGraveyard {
		if p.Spec.ConditionalShuffle {
			if err := p.validateConditionalShuffleSearch(); err != nil {
				return err
			}
		} else if p.Spec.Destination != zone.Hand ||
			!p.Spec.Reveal ||
			p.Spec.Name == "" ||
			p.Spec.RevealOnly ||
			p.Spec.MaxManaValueFromX ||
			p.Spec.MaxManaValueFromSacrificedCost.Exists ||
			p.Spec.SharedSubtype ||
			p.Spec.DifferentNames ||
			p.Spec.EntersTapped ||
			p.Spec.SplitDestination.Exists ||
			len(p.Spec.SlotFilters) != 0 ||
			p.Controller.Exists ||
			p.PlayerGroup.Kind != PlayerGroupReferenceNone {
			return errors.New("library-and-graveyard search requires a named reveal-to-hand search with no other riders")
		}
	} else if p.Spec.ConditionalShuffle {
		return errors.New("conditional-shuffle search requires AlsoGraveyard")
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
	if p.Spec.DifferentNames {
		if !p.Amount.IsDynamic() && p.Amount.Value() < 2 {
			return errors.New("different-names search must allow more than one card")
		}
		if p.Spec.SharedSubtype ||
			p.Spec.SplitDestination.Exists ||
			len(p.Spec.SlotFilters) != 0 ||
			p.Spec.Destination == zone.Library {
			return errors.New("different-names search cannot combine shared-subtype, split, slot, or library-top riders")
		}
	}
	if p.Spec.AnyNumber {
		// An "any number of" search finds none up to every matching card, so its
		// destination must accept an unbounded set of found cards (hand,
		// battlefield, or graveyard) and it cannot combine with any rider that
		// presupposes a bounded count or a single found card.
		if p.Spec.Destination != zone.Hand &&
			p.Spec.Destination != zone.Battlefield &&
			p.Spec.Destination != zone.Graveyard {
			return errors.New("any-number search requires a hand, battlefield, or graveyard destination")
		}
		if p.Spec.DestinationPosition != SearchPositionUnspecified ||
			p.Spec.SplitDestination.Exists ||
			p.Spec.SharedSubtype ||
			p.Spec.DifferentNames ||
			len(p.Spec.SlotFilters) != 0 ||
			p.Spec.AlsoGraveyard ||
			p.Spec.MaxManaValueFromX ||
			p.Spec.MaxManaValueFromSacrificedCost.Exists ||
			p.PlayerGroup.Kind != PlayerGroupReferenceNone {
			return errors.New("any-number search cannot combine bounded-count, multi-zone, or multi-searcher riders")
		}
		if p.Spec.FailToFindPolicy == SearchMustFindIfAvailable {
			return errors.New("any-number search cannot require finding a card")
		}
	}
	if len(p.Spec.Filter.RequiredTypes) != 0 && len(p.Spec.Filter.RequiredTypesAny) != 0 {
		return errors.New("search cannot combine one required card type with a card-type union")
	}
	if len(p.Spec.Filter.RequiredTypesAny) == 1 {
		return errors.New("search card-type union requires at least two card types")
	}
	if slices.Contains(p.Spec.Filter.RequiredTypesAny, "") {
		return errors.New("search card-type union cannot contain an empty type")
	}
	if slices.Contains(p.Spec.Filter.Supertypes, "") {
		return errors.New("search supertype cannot be empty")
	}
	if problems := p.Spec.Filter.Validate(); len(problems) != 0 {
		return errors.New("search filter: " + problems[0])
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
	if err := validateMassPlayerOrGroup("Search", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
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

// validateRevealOnlySearch validates the search-and-reveal-only slice that leaves
// the found card in the library for a following ConditionalDestinationPlace to
// route. It requires a single searching player, a library source, no destination,
// a reveal, exactly one found card, and none of the placement riders.
func (p Search) validateRevealOnlySearch(targets []TargetSpec, checkTargets bool) error {
	if p.Spec.SourceZone != zone.Library {
		return errors.New("reveal-only search must source from the library")
	}
	if p.Spec.Destination != zone.None {
		return errors.New("reveal-only search leaves the card in the library and has no destination")
	}
	if !p.Spec.Reveal {
		return errors.New("reveal-only search must reveal the found card")
	}
	if p.Amount.IsDynamic() || p.Amount.Value() != 1 {
		return errors.New("reveal-only search must find exactly one card")
	}
	if p.Spec.SplitDestination.Exists ||
		p.Spec.EntersTapped ||
		p.Spec.MaxManaValueFromX ||
		p.Spec.MaxManaValueFromSacrificedCost.Exists ||
		p.Spec.SharedSubtype ||
		p.Spec.DifferentNames ||
		p.Spec.DestinationPosition != SearchPositionUnspecified {
		return errors.New("reveal-only search does not support split destination, tapped entry, X bound, shared-subtype, or library-position riders")
	}
	if p.Controller.Exists {
		return errors.New("reveal-only search does not support a controller rider")
	}
	if p.PlayerGroup.Kind != PlayerGroupReferenceNone {
		return errors.New("reveal-only search requires a single searching player")
	}
	switch p.Spec.FailToFindPolicy {
	case SearchFailToFindDefault, SearchMayFailToFind:
	default:
		return errors.New("reveal-only search supports only the default or may-fail-to-find policy")
	}
	if len(p.Spec.Filter.RequiredTypes) != 0 && len(p.Spec.Filter.RequiredTypesAny) != 0 {
		return errors.New("search cannot combine one required card type with a card-type union")
	}
	if len(p.Spec.Filter.RequiredTypesAny) == 1 {
		return errors.New("search card-type union requires at least two card types")
	}
	if problems := p.Spec.Filter.Validate(); len(problems) != 0 {
		return errors.New("search filter: " + problems[0])
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

// validateExileFaceDownSearch validates the search-and-exile-face-down slice that
// finds a single card, exiles it face down, shuffles the library, and publishes
// the exiled card under a linked key for a following instruction to reference. It
// requires a single searching player, a library source, an Exile destination,
// exactly one found card, a published link, and none of the reveal, split,
// tapped-entry, controller, or ordered-position riders.
func (p Search) validateExileFaceDownSearch(targets []TargetSpec, checkTargets bool) error {
	if p.Spec.SourceZone != zone.Library {
		return errors.New("exile-face-down search must source from the library")
	}
	if p.Spec.Destination != zone.Exile {
		return errors.New("exile-face-down search must send the found card to exile")
	}
	if p.Amount.IsDynamic() || p.Amount.Value() != 1 {
		return errors.New("exile-face-down search must find exactly one card")
	}
	if p.PublishLinked == "" {
		return errors.New("exile-face-down search must publish the exiled card under a linked key")
	}
	if p.Spec.Reveal ||
		p.Spec.RevealOnly ||
		p.Spec.AlsoGraveyard ||
		p.Spec.SplitDestination.Exists ||
		p.Spec.EntersTapped ||
		p.Spec.MaxManaValueFromX ||
		p.Spec.MaxManaValueFromSacrificedCost.Exists ||
		p.Spec.SharedSubtype ||
		p.Spec.DifferentNames ||
		len(p.Spec.SlotFilters) != 0 ||
		(p.Spec.Name != "" || p.Spec.NameFromChoice != "") ||
		p.Spec.DestinationPosition != SearchPositionUnspecified {
		return errors.New("exile-face-down search does not support reveal, split destination, tapped entry, X bound, name, slot, or library-position riders")
	}
	if p.Controller.Exists {
		return errors.New("exile-face-down search does not support a controller rider")
	}
	if p.PlayerGroup.Kind != PlayerGroupReferenceNone {
		return errors.New("exile-face-down search requires a single searching player")
	}
	switch p.Spec.FailToFindPolicy {
	case SearchFailToFindDefault, SearchMayFailToFind, SearchMustFindIfAvailable:
	default:
		return errors.New("exile-face-down search has unsupported fail-to-find policy")
	}
	if problems := p.Spec.Filter.Validate(); len(problems) != 0 {
		return errors.New("search filter: " + problems[0])
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

// validateConditionalShuffleSearch checks the choose-your-zones, publish-searched
// form of the multi-zone "search your library and/or graveyard ... and put it
// onto the battlefield" search (Finale of Devastation). It requires a Hand or
// Battlefield destination and rejects the reveal, name, split, tapped-entry,
// slot, controller, and multi-player riders that form has no wording for; it
// composes with MaxManaValueFromX and a card-type/characteristic Filter.
func (p Search) validateConditionalShuffleSearch() error {
	if p.Spec.Destination != zone.Hand && p.Spec.Destination != zone.Battlefield {
		return errors.New("conditional-shuffle search must put the found card into the hand or onto the battlefield")
	}
	if p.Spec.Reveal ||
		p.Spec.RevealOnly ||
		p.Spec.ExileFaceDown ||
		(p.Spec.Name != "" || p.Spec.NameFromChoice != "") ||
		p.Spec.MaxManaValueFromSacrificedCost.Exists ||
		p.Spec.SharedSubtype ||
		p.Spec.DifferentNames ||
		p.Spec.EntersTapped ||
		p.Spec.SplitDestination.Exists ||
		len(p.Spec.SlotFilters) != 0 ||
		p.Spec.DestinationPosition != SearchPositionUnspecified ||
		p.Controller.Exists ||
		p.PlayerGroup.Kind != PlayerGroupReferenceNone {
		return errors.New("conditional-shuffle search does not support reveal, name, split, tapped entry, slot, controller, or multi-player riders")
	}
	if p.Amount.IsDynamic() || p.Amount.Value() != 1 {
		return errors.New("conditional-shuffle search must find exactly one card")
	}
	return nil
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

// validatePrimitive validates the canonical choose-from-zone envelope. It
// enforces the bound (a fixed positive Quantity, or the any-number form's empty
// Quantity and absent mana-value cap), the object-scoped single-card publish
// restriction (the imprint rule, Chrome Mox), and the destination/tapped/
// mana-value-cap rules, failing closed on any unsupported combination. It
// reproduces the accept/reject set of the retired per-family validators
// (ExileFromHand, ExileFromGraveyard, PutFromHand, ReturnFromGraveyard) that now
// lower to this envelope.
func (p ChooseFromZone) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.SourceZone == zone.None {
		return errors.New("choose from zone requires a source zone")
	}
	if p.Count == ChooseAnyNumber {
		if p.Quantity.IsDynamic() || p.Quantity.Value() != 0 {
			return errors.New("choose from zone any-number form takes no fixed amount")
		}
		if p.Riders.MaxTotalManaValue.Exists {
			return errors.New("choose from zone any-number form takes no total mana value cap")
		}
	} else {
		if err := validateQuantity(p.Quantity, targets, checkTargets); err != nil {
			return err
		}
		if p.Quantity.IsDynamic() || p.Quantity.Value() < 1 {
			return errors.New("choose from zone requires a fixed positive amount")
		}
	}
	if p.Riders.PublishObjectScoped && p.Riders.PublishLinked != "" && p.Quantity.Value() != 1 {
		return errors.New("linked choose from zone must move exactly one card")
	}
	switch p.Destination.Zone {
	case zone.Exile, zone.Hand, zone.Battlefield:
	default:
		return errors.New("choose from zone requires an exile, hand, or battlefield destination")
	}
	if p.Riders.EntersTapped && p.Destination.Zone != zone.Battlefield {
		return errors.New("choose from zone tapped entry requires a battlefield destination")
	}
	if p.Riders.EntersAttacking {
		if p.Destination.Zone != zone.Battlefield {
			return errors.New("choose from zone attacking entry requires a battlefield destination")
		}
		if p.Riders.FaceDown {
			return errors.New("choose from zone attacking entry cannot enter face down")
		}
	}
	if p.Riders.MaxTotalManaValue.Exists {
		if p.Destination.Zone != zone.Battlefield {
			return errors.New("choose from zone total mana value cap requires a battlefield destination")
		}
		if p.Riders.MaxTotalManaValue.Val < 0 {
			return errors.New("choose from zone total mana value cap must be non-negative")
		}
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExileEntireHand) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("exile entire hand requires a linked key")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ReturnExiledCardsToHand) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("return exiled cards to hand requires a linked key")
	}
	return nil
}

func (p ReturnExiledCardsWithCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if !p.Counter.Valid() {
		return errors.New("return exiled cards with counter requires a valid counter kind")
	}
	if p.Counter.PlayerOnly() {
		return errors.New("return exiled cards with counter requires a permanent-placeable counter kind")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExileForEachPlayer) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("exile for each player requires a linked key")
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	return validatePlayerReference(p.Chooser, targets, checkTargets)
}

func (p ChampionExile) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("champion exile requires a linked key")
	}
	return firstProblem(p.Selection.Validate())
}

func (p ReturnLinkedExiledCardsToBattlefield) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("return linked exiled cards to battlefield requires a linked key")
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Amount.IsDynamic() || p.Amount.Value() < 1 {
		return errors.New("return linked exiled cards to battlefield requires a fixed positive Amount")
	}
	return validatePlayerReference(p.Chooser, targets, checkTargets)
}

func (p DestroyForEachPlayer) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("destroy for each player requires a linked key")
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	return validatePlayerReference(p.Chooser, targets, checkTargets)
}

func (p EachPlayerChooseDestroy) validatePrimitive([]TargetSpec, bool) error {
	if p.Selection.Empty() {
		return errors.New("each player choose destroy requires a candidate selection")
	}
	return firstProblem(p.Selection.Validate())
}

func (p OptionalCounterForEachPlayer) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if problems := p.Players.Validate(); len(problems) != 0 {
		return fmt.Errorf("optional counter player group: %s", problems[0])
	}
	if p.Selection.Empty() {
		return errors.New("optional counter for each player requires a candidate selection")
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	if p.Amount.IsDynamic() {
		if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
			return err
		}
	} else if p.Amount.Value() <= 0 {
		return errors.New("optional counter for each player requires a positive amount")
	}
	if !p.CounterKind.Valid() || p.CounterKind.PlayerOnly() {
		return errors.New("optional counter for each player requires a permanent-placeable counter kind")
	}
	if p.PublishLinked == "" {
		return errors.New("optional counter for each player requires PublishLinked")
	}
	return nil
}

func (p CreateTokenForEachDestroyed) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("create token for each destroyed requires a linked key")
	}
	if !p.Source.Valid() {
		return errors.New("create token for each destroyed requires a valid source")
	}
	return nil
}

func (p ExileForEachOpponent) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("exile for each opponent requires a linked key")
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	switch p.Extremum {
	case PermanentChoiceExtremumNone, PermanentChoiceGreatestPower,
		PermanentChoiceGreatestToughness, PermanentChoiceGreatestManaValue:
	default:
		return errors.New("exile for each opponent has an invalid choice extremum")
	}
	return validatePlayerReference(p.Chooser, targets, checkTargets)
}

func (p DrawForEachExiled) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("draw for each exiled requires a linked key")
	}
	return nil
}

func (p ManifestForEachLinked) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("manifest for each linked requires a linked key")
	}
	if p.Dread && p.Cloak {
		return errors.New("manifest for each linked cannot be both dread and cloak")
	}
	return nil
}

func (p RemoveTargetsForToken) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("remove targets for token requires a linked key")
	}
	return nil
}

func (p CastForFree) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Zone == zone.None {
		return errors.New("cast for free requires a source zone")
	}
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.Card.Kind != CardReferenceNone {
		if p.MaxManaValueFromX {
			return errors.New("cast for free X mana-value bound requires a selection-driven cast")
		}
		if err := validateCardReference(p.Card); err != nil {
			return err
		}
		return validateTargetCardReference(p.Card, targets, checkTargets)
	}
	if p.PayManaCost {
		return errors.New("cast for free paying mana cost requires a referenced card")
	}
	return nil
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
	if p.FromTriggerBatch {
		if p.Destination != zone.Battlefield {
			return errors.New("mass return from graveyard trigger batch requires a battlefield destination")
		}
		if p.SourceGroup.Kind != PlayerGroupReferenceNone {
			return errors.New("mass return from graveyard trigger batch cannot widen source graveyards")
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
	attackModes := 0
	if p.EntryAttacking {
		attackModes++
	}
	if p.EntryAttackingDefender.Exists {
		attackModes++
	}
	if p.AttackEachOtherOpponent {
		attackModes++
	}
	if p.AttackSameAsSource {
		attackModes++
	}
	if p.AttackSameAsObject.Exists {
		attackModes++
	}
	if attackModes > 1 {
		return errors.New("create token attack-entry modes are mutually exclusive")
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.EntryAttachedTo.Exists {
		if err := validateObjectReference(p.EntryAttachedTo.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	if p.AttackSameAsObject.Exists {
		if err := validateObjectReference(p.AttackSameAsObject.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	if spec, ok := p.Source.TokenCopy(); ok {
		switch spec.Source {
		case TokenCopySourceObject:
			if err := validateObjectReference(spec.Object, targets, checkTargets); err != nil {
				return err
			}
		case TokenCopySourceEachInGroup, TokenCopySourceChosenFromGroup:
			if spec.Group == nil {
				return errors.New("create token copy group is nil")
			}
			if problems := spec.Group.Validate(); len(problems) != 0 {
				return errors.New(problems[0])
			}
		default:
		}
	}
	if p.Recipient.Exists {
		if p.RecipientGroup.Kind != PlayerGroupReferenceNone {
			return errors.New("create token cannot set both a recipient and a recipient group")
		}
		return validatePlayerReference(p.Recipient.Val, targets, checkTargets)
	}
	if p.RecipientGroup.Kind != PlayerGroupReferenceNone {
		if problems := p.RecipientGroup.Validate(); len(problems) != 0 {
			return errors.New(problems[0])
		}
	}
	if p.EntryAttackingDefender.Exists {
		return validatePlayerReference(p.EntryAttackingDefender.Val, targets, checkTargets)
	}
	return nil
}

func (p ShufflePermanentIntoLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p PutPermanentOnLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p PutLinkedExiledCardsInLibrary) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkedKey == "" {
		return errors.New("put linked exiled cards in library requires a linked key")
	}
	if p.RandomOrder && !p.Bottom {
		return errors.New("put linked exiled cards in library random order requires bottom placement")
	}
	return nil
}

func (p StartEngines) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p BecomeMonarch) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}
func (p CantBecomeMonarch) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

// validatePrimitive implements Primitive for VentureIntoDungeon,
// VentureIntoUndercity, and TakeInitiative, each of which acts on a single
// referenced player.
func (p VentureIntoDungeon) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}
func (p VentureIntoUndercity) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}
func (p TakeInitiative) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

// validatePrimitive implements Primitive for RevealPutOntoBattlefield. It reveals
// and puts onto the battlefield without consulting the ability's target specs.
func (p RevealPutOntoBattlefield) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Look, targets, checkTargets); err != nil {
		return err
	}
	return validateQuantity(p.Counters, targets, checkTargets)
}

// validatePrimitive implements Primitive for CastLinkedCardForFree. It requires a
// non-empty linked key naming the card group to cast from.
func (p CastLinkedCardForFree) validatePrimitive([]TargetSpec, bool) error {
	if p.LinkID == "" {
		return errors.New("cast linked card for free requires a link key")
	}
	return validatePlayerReference(p.Player, nil, false)
}

// validatePrimitive implements Primitive for RollDiceCreateTokens. It requires a
// positive die count and side count and a valid token source.
func (p RollDiceCreateTokens) validatePrimitive([]TargetSpec, bool) error {
	if p.Dice <= 0 || p.Sides <= 0 {
		return errors.New("roll dice create tokens requires positive dice and sides")
	}
	if !p.Source.Valid() {
		return errors.New("roll dice create tokens requires a valid token source")
	}
	return nil
}

// validatePrimitive implements Primitive for RevealToHandDrainManaValue.
func (p RevealToHandDrainManaValue) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

// validatePrimitive implements Primitive for GoadForEachOpponent and
// CreateCommanderCopyToken. Both act on the resolving controller and consult no
// target specs.
func (GoadForEachOpponent) validatePrimitive([]TargetSpec, bool) error      { return nil }
func (CreateCommanderCopyToken) validatePrimitive([]TargetSpec, bool) error { return nil }

// validatePrimitive implements Primitive for GainCityBlessing. It always acts on
// the resolving object's controller and consults no targets, so it has nothing
// to validate against the ability's target specs.
func (GainCityBlessing) validatePrimitive([]TargetSpec, bool) error {
	return nil
}

// validatePrimitive implements Primitive for PartitionExiledCostCards. It reads
// the resolving object's cost-exiled card IDs and consults no targets, so it has
// nothing to validate against the ability's target specs.
func (PartitionExiledCostCards) validatePrimitive([]TargetSpec, bool) error {
	return nil
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

func (p Incubate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.Recipient.Exists {
		return validatePlayerReference(p.Recipient.Val, targets, checkTargets)
	}
	return nil
}

func (p Bolster) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateQuantity(p.Amount, targets, checkTargets)
}

func (p Renown) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Adapt) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p Connive) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p BecomeSaddled) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p RecordEchoObligation) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (ShuffleSpellIntoLibrary) validatePrimitive(_ []TargetSpec, _ bool) error {
	return nil
}

func (p Pay) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateResolutionPayment(p.Payment, targets, checkTargets)
}

func (p PayRepeatedly) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.PublishCount == "" {
		return errors.New("PayRepeatedly requires a published count key")
	}
	if p.MaxCount.Exists && p.MaxCount.Val == nil {
		return errors.New("PayRepeatedly MaxCount is set but nil")
	}
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
	if p.Choice.AtRandom && p.Choice.Kind != ResolutionChoiceNumber {
		return errors.New("random resolution choice requires a number choice")
	}
	if p.Choice.Kind == ResolutionChoicePermanent {
		if p.Choice.PlayerReference == nil {
			return errors.New("permanent choice requires a player reference")
		}
		if p.Choice.Selection == nil {
			return errors.New("permanent choice requires a selection")
		}
		if problems := p.Choice.Selection.Validate(); len(problems) != 0 {
			return errors.New("permanent choice selection: " + problems[0])
		}
	}
	return nil
}

func (p GainLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("GainLife", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p LoseLife) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("LoseLife", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExchangeLifeTotalWithSourceCharacteristic) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Characteristic != SourcePower && p.Characteristic != SourceToughness {
		return errors.New("life-total exchange requires source power or toughness")
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
	if p.ContinueResult == "" {
		if err := validateQuantity(p.Times, targets, checkTargets); err != nil {
			return err
		}
	} else if p.Times.IsDynamic() || p.Times.Value() != 0 {
		return errors.New("conditional RepeatProcess cannot also set Times")
	}
	if len(p.Body.Modes) == 0 {
		return errors.New("RepeatProcess requires a body")
	}
	if err := validateNestedAbilityContent(p.Body, nil, targets, checkTargets, nil); err != nil {
		return err
	}
	if p.ContinueResult != "" && !abilityContentPublishesResult(p.Body, p.ContinueResult) {
		return errors.New("conditional RepeatProcess body must publish ContinueResult")
	}
	return nil
}

func abilityContentPublishesResult(content AbilityContent, key ResultKey) bool {
	for i := range content.Modes {
		for j := range content.Modes[i].Sequence {
			if content.Modes[i].Sequence[j].PublishResult == key {
				return true
			}
		}
	}
	return false
}

func (p Exile) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.SourceSpell {
		if p.Object.Kind() != ObjectReferenceNone || p.Group.Valid() || p.ExileLinkedKey != "" {
			return errors.New("source-spell exile cannot set an object, group, or linked key")
		}
		return nil
	}
	if p.ExileLinkedKey != "" && p.Group.Valid() && p.Object.Kind() != ObjectReferenceNone {
		return errors.New("linked exile must not set both an object and a group")
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
	if p.PublishLinked != "" && (!hasCard || p.Destination != zone.Exile) {
		return errors.New("linked move-card publication requires a single-card exile")
	}
	if p.PublishLinkedObjectScoped && p.PublishLinked == "" {
		return errors.New("object-scoped move-card publication requires a linked key")
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
	if p.Duration != DurationUntilEndOfTurn && p.Duration != DurationUntilEndOfYourNextTurn {
		return errors.New("cast permission requires a supported bounded duration")
	}
	return nil
}

func (p ExileForPlay) validatePrimitive([]TargetSpec, bool) error {
	if !p.SelectFromBatch {
		if err := validateCardReference(p.Card); err != nil {
			return err
		}
	}
	if p.FromZone != zone.Graveyard {
		return errors.New("ExileForPlay requires graveyard as its source zone")
	}
	if p.Duration != DurationThisTurn && p.Duration != DurationUntilEndOfTurn &&
		p.Duration != DurationUntilEndOfYourNextTurn && p.Duration != DurationUntilYourNextEndStep {
		return errors.New("ExileForPlay requires a this-turn, until-end-of-turn, until-end-of-your-next-turn, or until-your-next-end-step play window")
	}
	return nil
}

func (p ExilePermanentForPlay) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p PlayChosenExiledCard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.Zone != zone.Exile {
		return errors.New("PlayChosenExiledCard requires exile as its source zone")
	}
	if p.Duration != DurationThisTurn && p.Duration != DurationUntilEndOfTurn &&
		p.Duration != DurationUntilEndOfYourNextTurn && p.Duration != DurationUntilYourNextEndStep {
		return errors.New("PlayChosenExiledCard requires a this-turn, until-end-of-turn, until-end-of-your-next-turn, or until-your-next-end-step play window")
	}
	return nil
}

func (p CopyCard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.LinkID == "" {
		return errors.New("CopyCard requires a link ID")
	}
	return nil
}

func (p PlayLinkedExiledCard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validatePlayerReference(p.Player, targets, checkTargets); err != nil {
		return err
	}
	if p.LinkID == "" {
		return errors.New("PlayLinkedExiledCard requires a link ID")
	}
	return nil
}

func (p Sacrifice) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Group.Valid() {
		return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
	}
	if p.Object.Kind() == ObjectReferenceNone {
		return nil
	}
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p SacrificePermanents) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.All && p.AnyNumber {
		return errors.New("SacrificePermanents cannot set both All and AnyNumber")
	}
	if p.All {
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("SacrificePermanents with All set requires a zero Amount")
		}
		if p.Fallback.Kind != SacrificeFallbackNone {
			return errors.New("SacrificePermanents with All set cannot carry a fallback")
		}
	}
	if p.AnyNumber {
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("SacrificePermanents with AnyNumber set requires a zero Amount")
		}
		if p.Fallback.Kind != SacrificeFallbackNone {
			return errors.New("SacrificePermanents with AnyNumber set cannot carry a fallback")
		}
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("SacrificePermanents", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	if err := p.Fallback.validate(targets, checkTargets); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	if hasGroup {
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p KeepOnePerType) validatePrimitive(_ []TargetSpec, _ bool) error {
	if len(p.Types) == 0 {
		return errors.New("KeepOnePerType requires at least one type")
	}
	if err := firstProblem(p.AffectedSelection.Validate()); err != nil {
		return err
	}
	return validatePlayerGroupReference(p.Players)
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
	if p.ChooseUpTo && p.ChooseOne {
		return errors.New("untap cannot choose up to and choose exactly one")
	}
	if p.ChooseOne {
		if p.Object.Kind() != ObjectReferenceNone {
			return errors.New("choose-one untap requires a Group rather than an Object")
		}
		if p.Amount.IsDynamic() || p.Amount.Value() != 0 {
			return errors.New("choose-one untap cannot carry an Amount")
		}
		return nil
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
	if p.Chooser.Kind() != PlayerReferenceNone {
		if err := firstProblem(p.Chooser.Validate()); err != nil {
			return fmt.Errorf("untap chooser: %w", err)
		}
	}
	return nil
}

func (p SkipNextUntap) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p RemoveFromCombat) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p CounterObject) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	if p.Object.Kind() != ObjectReferenceTargetStackObject &&
		p.Object.Kind() != ObjectReferenceEventStackObject {
		return errors.New("counter object requires a target or event stack object reference")
	}
	if p.ExileInstead && p.Destination != CounteredSpellGraveyard {
		return errors.New("counter object cannot both exile and redirect a countered spell")
	}
	switch p.Destination {
	case CounteredSpellGraveyard, CounteredSpellLibraryTop, CounteredSpellHand:
	default:
		return errors.New("counter object has an unknown countered-spell destination")
	}
	if p.Object.Kind() == ObjectReferenceEventStackObject {
		return nil
	}
	return validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
}

func (p ExileTargetSpells) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	switch p.Object.Kind() {
	case ObjectReferenceTargetStackObject:
		return validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
	case ObjectReferenceAllTargetStackObjects:
		return validateTargetSpecAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
	default:
		return errors.New("exile target spells requires a target or all-target stack object reference")
	}
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

func (p ChangeStackObjectController) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	if p.Object.Kind() != ObjectReferenceTargetStackObject {
		return errors.New("change stack object controller requires a target stack object reference")
	}
	if err := validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets); err != nil {
		return err
	}
	return validatePlayerReference(p.Controller, targets, checkTargets)
}

func (p CopyStackObject) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateObjectReference(p.Object, targets, checkTargets); err != nil {
		return err
	}
	if p.Count < 0 {
		return errors.New("copy stack object count must be nonnegative")
	}
	if p.DynamicCount.IsDynamic() {
		if p.Count != 0 {
			return errors.New("copy stack object cannot combine fixed and dynamic counts")
		}
		if err := validateQuantity(p.DynamicCount, targets, checkTargets); err != nil {
			return err
		}
	} else if p.DynamicCount.Value() != 0 {
		return errors.New("copy stack object dynamic count must be dynamic")
	}
	if p.Chooser.Exists {
		if err := validatePlayerReference(p.Chooser.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	switch p.Object.Kind() {
	case ObjectReferenceTargetStackObject:
		return validateTargetAllows(p.Object.TargetIndex(), TargetAllowStackObject, targets, checkTargets)
	case ObjectReferenceEventStackObject, ObjectReferenceResolvingStackObject:
		return nil
	default:
		return errors.New("copy stack object requires a target, event, or resolving stack object reference")
	}
}

func (p Mill) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("Mill", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
	if hasGroup {
		if p.PublishLinked != "" {
			return errors.New("Mill cannot publish linked cards for the group form")
		}
		return validatePlayerGroupReference(p.PlayerGroup)
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExileTopOfLibrary) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("ExileTopOfLibrary", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
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

func (p DiscardThenDraw) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Max < 0 {
		return errors.New("DiscardThenDraw requires a non-negative Max")
	}
	if p.DrawOffset < 0 {
		return errors.New("DiscardThenDraw requires a non-negative DrawOffset")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p DiscardUnlessType) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Amount < 0 {
		return errors.New("DiscardUnlessType requires a non-negative Amount")
	}
	if len(p.ExemptTypes) == 0 {
		return errors.New("DiscardUnlessType requires at least one ExemptType")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p RevealUntil) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Destination != zone.Graveyard && p.Destination != zone.Hand {
		return errors.New("RevealUntil requires a Graveyard or Hand Destination")
	}
	if p.MatchToDestinationRestRandomBottom && p.Destination != zone.Hand {
		return errors.New("RevealUntil matching-card partition requires a Hand Destination")
	}
	if err := firstProblem(p.Until.Validate()); err != nil {
		return err
	}
	if err := validateMassPlayerOrGroup("RevealUntil", p.Player, p.PlayerGroup); err != nil {
		return err
	}
	hasGroup := p.PlayerGroup.Kind != PlayerGroupReferenceNone
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
	switch p.Destination {
	case zone.None, zone.Battlefield, zone.Library:
	default:
		return errors.New("Dig supports only a hand (zero), battlefield, or library-top destination")
	}
	if p.EntersTapped && p.Destination != zone.Battlefield {
		return errors.New("Dig EntersTapped requires a battlefield destination")
	}
	if p.Destination == zone.Library && p.Reveal {
		return errors.New("Dig with a library-top destination does not reveal the chosen cards")
	}
	for _, slot := range p.Slots {
		if err := validateQuantity(slot.Count, targets, checkTargets); err != nil {
			return err
		}
		if !slot.Count.IsDynamic() && slot.Count.Value() < 1 {
			return errors.New("Dig slot requires routing a positive number of cards")
		}
		switch slot.Destination {
		case zone.Hand, zone.Library, zone.Exile, zone.Graveyard:
		default:
			return errors.New("Dig slot has an unsupported destination")
		}
		if slot.Bottom && slot.Destination != zone.Library {
			return errors.New("Dig slot bottom placement requires a library destination")
		}
		if slot.Play.Exists && slot.Destination != zone.Exile {
			return errors.New("Dig slot play permission requires an exile destination")
		}
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p PileSplit) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() < 1 {
		return errors.New("PileSplit requires revealing a positive number of cards")
	}
	if p.Kept != zone.Hand {
		return errors.New("PileSplit requires a Hand kept destination")
	}
	if p.Other != zone.Graveyard && p.Other != zone.Library {
		return errors.New("PileSplit requires a Graveyard or Library other destination")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p RevealTopPartition) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() < 1 {
		return errors.New("RevealTopPartition requires revealing a positive number of cards")
	}
	if p.Remainder != DigRemainderGraveyard && p.Remainder != DigRemainderLibraryBottom {
		return errors.New("RevealTopPartition requires a Graveyard or LibraryBottom remainder")
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
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
		p.Duration != DurationUntilEndOfYourNextTurn && p.Duration != DurationUntilYourNextEndStep &&
		p.Duration != DurationPermanent {
		return errors.New("ImpulseExile requires a this-turn, until-end-of-turn, until-end-of-your-next-turn, until-your-next-end-step, or for-as-long-as-exiled play window")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExileLibraryUntilNonlandCast) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p IterativeLibraryProcess) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Stop >= iterativeLibraryStopCount {
		return errors.New("IterativeLibraryProcess has an unknown stop predicate")
	}
	if err := validateQuantity(p.PreExile, targets, checkTargets); err != nil {
		return err
	}
	if !p.PreExile.IsDynamic() && p.PreExile.Value() < 0 {
		return errors.New("IterativeLibraryProcess requires a non-negative pre-exile count")
	}
	if p.Stop == IterativeLibraryStopChosenName && !p.ChooseName {
		return errors.New("IterativeLibraryProcess chosen-name stop requires ChooseName")
	}
	if p.OptionalTake && p.Stop != IterativeLibraryStopDuplicateName {
		return errors.New("IterativeLibraryProcess OptionalTake requires the duplicate-name stop")
	}
	if p.AllowAbsentName && p.Stop != IterativeLibraryStopChosenName {
		return errors.New("IterativeLibraryProcess AllowAbsentName requires the chosen-name stop")
	}
	if p.Stop == IterativeLibraryStopDifferentNameNonland {
		if err := validateObjectReference(p.DifferentNameFrom, targets, checkTargets); err != nil {
			return err
		}
		if p.DifferentNameFrom.Kind() != ObjectReferenceTargetStackObject {
			return errors.New("IterativeLibraryProcess different-name-nonland stop requires a target stack object reference")
		}
	} else {
		if p.DifferentNameFrom.Kind() != ObjectReferenceNone {
			return errors.New("IterativeLibraryProcess DifferentNameFrom requires the different-name-nonland stop")
		}
		if p.PublishLinked != "" {
			return errors.New("IterativeLibraryProcess PublishLinked requires the different-name-nonland stop")
		}
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p ExileTopEachLibraryCastFree) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() < 1 {
		return errors.New("ExileTopEachLibraryCastFree requires a positive number of cards")
	}
	return nil
}

func (p HideawayExile) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if !p.Amount.IsDynamic() && p.Amount.Value() < 1 {
		return errors.New("HideawayExile requires a positive number of cards")
	}
	return nil
}

func (PlayHideawayCard) validatePrimitive([]TargetSpec, bool) error { return nil }

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
	if p.ConsumeLinked {
		key, linked := p.Group.LinkedKey()
		if !linked || key == "" || p.Object.Kind() != ObjectReferenceNone {
			return errors.New("consume-linked goad requires a linked-objects group")
		}
	}
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p RemoveCounter) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.AllKinds {
		if p.Amount != (Quantity{}) {
			return errors.New("remove all counters must not set an amount")
		}
		return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
	}
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p Transform) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p TurnFaceDown) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateObjectReference(p.Object, targets, checkTargets)
}

func (p PhaseOut) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
}

func (p Regenerate) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validateMassObjectOrGroup(p.Object, p.Group, targets, checkTargets)
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
	if p.CombatCount < 0 {
		return errors.New("add extra phases combat count must not be negative")
	}
	if p.Combat && p.CombatCount != 0 {
		return errors.New("add extra phases must set only one of combat or combat count")
	}
	hasCombat := p.Combat || p.CombatCount > 0
	if !hasCombat && !p.Main && !p.Beginning {
		return errors.New("add extra phases requires at least one phase")
	}
	if p.Main && !hasCombat {
		return errors.New("add extra phases main phase must follow an extra combat phase")
	}
	if p.Beginning && (hasCombat || p.Main) {
		return errors.New("add extra beginning phase does not combine with combat or main phases")
	}
	return nil
}

func (AddExtraUpkeepStep) validatePrimitive([]TargetSpec, bool) error {
	return nil
}

func (p RollDie) validatePrimitive([]TargetSpec, bool) error {
	if p.Sides < 2 {
		return errors.New("roll die requires at least two sides")
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
	if p.Trigger.EventPattern.Exists {
		if p.Trigger.Timing != 0 {
			return errors.New("event-based delayed trigger must not set a timing")
		}
		if p.Trigger.Window == DelayedWindowNone {
			return errors.New("event-based delayed trigger requires a window")
		}
	} else {
		switch p.Trigger.Timing {
		case DelayedAtBeginningOfNextEndStep, DelayedAtBeginningOfNextUpkeep, DelayedAtBeginningOfNextMainPhase, DelayedAtEndOfCombat, DelayedAtBeginningOfYourNextEndStep:
		default:
			return errors.New("delayed trigger requires a recognized timing")
		}
		if p.Trigger.Window != DelayedWindowNone {
			return errors.New("fixed-phase delayed trigger must not set a window")
		}
	}
	if len(p.Trigger.Content.Modes) == 0 {
		return errors.New("delayed trigger requires content")
	}
	if p.Trigger.DamageSourceObject.Exists {
		if !p.Trigger.EventPattern.Exists || !p.Trigger.EventPattern.Val.DamageSourceCaptured {
			return errors.New("delayed trigger DamageSourceObject requires a DamageSourceCaptured event pattern")
		}
		if err := validateObjectReference(p.Trigger.DamageSourceObject.Val, targets, checkTargets); err != nil {
			return err
		}
	} else if p.Trigger.EventPattern.Exists && p.Trigger.EventPattern.Val.DamageSourceCaptured {
		return errors.New("delayed trigger DamageSourceCaptured pattern requires a DamageSourceObject")
	}
	if p.Trigger.CapturedAttackerObject.Exists {
		if !p.Trigger.EventPattern.Exists || !p.Trigger.EventPattern.Val.AttackerCaptured {
			return errors.New("delayed trigger CapturedAttackerObject requires an AttackerCaptured event pattern")
		}
		if err := validateObjectReference(p.Trigger.CapturedAttackerObject.Val, targets, checkTargets); err != nil {
			return err
		}
	} else if p.Trigger.EventPattern.Exists && p.Trigger.EventPattern.Val.AttackerCaptured {
		return errors.New("delayed trigger AttackerCaptured pattern requires a CapturedAttackerObject")
	}
	if p.Trigger.CapturedDyingObject.Exists {
		if !p.Trigger.EventPattern.Exists || !p.Trigger.EventPattern.Val.DyingObjectCaptured {
			return errors.New("delayed trigger CapturedDyingObject requires a DyingObjectCaptured event pattern")
		}
		if err := validateObjectReference(p.Trigger.CapturedDyingObject.Val, targets, checkTargets); err != nil {
			return err
		}
	} else if p.Trigger.EventPattern.Exists && p.Trigger.EventPattern.Val.DyingObjectCaptured {
		return errors.New("delayed trigger DyingObjectCaptured pattern requires a CapturedDyingObject")
	}
	if p.Trigger.CapturedCard.Exists {
		if p.Trigger.EventPattern.Exists {
			return errors.New("delayed trigger CapturedCard requires a fixed-phase timing")
		}
		if p.Trigger.CapturedCard.Val.Kind() != ObjectReferenceLinkedObject {
			return errors.New("delayed trigger CapturedCard requires a linked-object reference")
		}
		if err := validateObjectReference(p.Trigger.CapturedCard.Val, targets, checkTargets); err != nil {
			return err
		}
	}
	return nil
}

func (p CreateReflexiveTrigger) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if len(p.Trigger.Content.Modes) == 0 {
		return errors.New("reflexive trigger requires content")
	}
	if p.Trigger.Content.IsModal() {
		return errors.New("reflexive trigger content must not be modal")
	}
	return nil
}

func (p CreateReplacement) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.Replacement == nil {
		return errors.New("create replacement requires a replacement")
	}
	if p.Replacement.MatchEvent == EventUnknown {
		return errors.New("create replacement requires an event")
	}
	if p.Object.Kind() != ObjectReferenceNone {
		return validateObjectReference(p.Object, targets, checkTargets)
	}
	return nil
}

func (p PreventDamage) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if err := validateQuantity(p.Amount, targets, checkTargets); err != nil {
		return err
	}
	if p.RedirectPreventedToSourceController {
		if !p.OneShot || !p.All {
			return errors.New("redirect-to-source-controller prevent damage requires OneShot and All")
		}
		if p.Player.Kind() != PlayerReferenceController {
			return errors.New("redirect-to-source-controller prevent damage requires the controller as the prevented player")
		}
	}
	hasObject := p.Object.Kind() != ObjectReferenceNone
	hasPlayer := p.Player.Kind() != PlayerReferenceNone
	_, hasAnyTarget := p.AnyTarget.AnyTargetObjectReference()
	if p.Global {
		if hasObject || hasPlayer || hasAnyTarget || p.BySource {
			return errors.New("global prevent damage must not set Object, Player, AnyTarget, or BySource")
		}
		return nil
	}
	if hasAnyTarget {
		if hasObject || hasPlayer || p.BySource {
			return errors.New("any-target prevent damage must not set Object, Player, or BySource")
		}
		object, _ := p.AnyTarget.AnyTargetObjectReference()
		if err := validateObjectReference(object, targets, checkTargets); err != nil {
			return err
		}
		player, _ := p.AnyTarget.AnyTargetPlayerReference()
		return validatePlayerReference(player, targets, checkTargets)
	}
	if hasObject == hasPlayer {
		return errors.New("prevent damage requires exactly one of Object, Player, or AnyTarget")
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
	siblingLinked map[LinkedKey]int,
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
			siblingLinked,
		); err != nil {
			return fmt.Errorf("mode %d: %w", i, err)
		}
	}
	return nil
}
