package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// inheritedRemovalTargetObjectRef resolves the object reference for an inherited
// antecedent target at the given clause-local occurrence index. A permanent
// target — any selector permanentTargetSpec accepts, including bare subtype
// ("Destroy target Mountain.") and compound type ("Exile target artifact or
// land.") leads — yields a permanent reference; a spell on the stack yields a
// stack-object reference. It fails closed (ok=false) for any other antecedent
// shape, including player and opponent targets, which the callers resolve to a
// player reference directly. Delegating the permanent test to permanentTargetSpec
// keeps the recognized permanent leads in one place so the inherited-recipient
// helpers stay in step with the targets the rest of the backend accepts.
func inheritedRemovalTargetObjectRef(
	target compiler.CompiledTarget,
	occurrence int,
) (game.ObjectReference, bool) {
	if target.Selector.Kind == compiler.SelectorSpell {
		return game.TargetStackObjectReference(occurrence), true
	}
	if _, ok := permanentTargetSpec(target); ok {
		return game.TargetPermanentReference(occurrence), true
	}
	return game.ObjectReference{}, false
}
