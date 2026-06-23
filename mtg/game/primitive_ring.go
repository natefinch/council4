package game

// RingTempts resolves "The Ring tempts you." (CR 701.51). The referenced player
// gets the Ring emblem if they don't already have it, advances it to the next
// of its four levels, and chooses a creature they control to become (or remain)
// their Ring-bearer. The player is the resolving controller for every printed
// "the Ring tempts you" wording.
type RingTempts struct {
	Player PlayerReference
}

// Kind implements Primitive for RingTempts.
func (RingTempts) Kind() PrimitiveKind { return PrimitiveRingTempts }

func (RingTempts) isPrimitive() {}

func (RingTempts) instructionRefs() primitiveRefs { return primitiveRefs{} }

func (p RingTempts) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	return validatePlayerReference(p.Player, targets, checkTargets)
}

func (p RingTempts) validateCapturedTargetControllerReferences(targets []TargetSpec, checkTargets bool) error {
	return validateCapturedTargetControllerReference(p.Player, targets, checkTargets)
}
