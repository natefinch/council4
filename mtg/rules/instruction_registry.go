package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

// primitiveHandler is a type-erased handler function for a Primitive kind.
type primitiveHandler func(r *effectResolver, prim game.Primitive) effectResolved

// primitiveRegistry is an immutable handler table keyed by PrimitiveKind.
type primitiveRegistry struct {
	handlers [game.PrimitiveKindCount]primitiveHandler
}

// registerPrimitiveHandler registers a typed handler for primitive kind T.
// The one type assertion is localized here; it is safe because we only register
// this handler for the PrimitiveKind returned by T's zero value.
func registerPrimitiveHandler[T game.Primitive](reg *primitiveRegistry, handler func(*effectResolver, T) effectResolved) {
	var zero T
	kind := zero.Kind()
	if int(kind) < 0 || int(kind) >= len(reg.handlers) {
		panic(fmt.Sprintf("rules: unregistered primitive kind %d", kind))
	}
	if reg.handlers[kind] != nil {
		panic(fmt.Sprintf("rules: duplicate primitive handler %d", kind))
	}
	reg.handlers[kind] = func(r *effectResolver, prim game.Primitive) effectResolved {
		typed, ok := prim.(T)
		if !ok {
			panic(fmt.Sprintf("rules: primitive handler %d received %T", kind, prim))
		}
		return handler(r, typed)
	}
}

func (reg *primitiveRegistry) dispatch(kind game.PrimitiveKind) primitiveHandler {
	if int(kind) >= len(reg.handlers) || int(kind) < 0 {
		panic(UnsupportedError{
			Kind:   kind,
			Reason: fmt.Sprintf("primitive kind %d is out of range", kind),
		})
	}
	h := reg.handlers[kind]
	if h == nil {
		panic(UnsupportedError{
			Kind:   kind,
			Reason: fmt.Sprintf("primitive kind %d has no registered handler", kind),
		})
	}
	return h
}

// newPrimitiveRegistry builds and returns the global handler table.
func newPrimitiveRegistry() *primitiveRegistry {
	reg := &primitiveRegistry{}
	registerPrimitiveHandler(reg, handleDamage)
	registerPrimitiveHandler(reg, handleDraw)
	registerPrimitiveHandler(reg, handleDiscard)
	registerPrimitiveHandler(reg, handleDestroy)
	registerPrimitiveHandler(reg, handleAddMana)
	registerPrimitiveHandler(reg, handleAddCounter)
	registerPrimitiveHandler(reg, handleAddPlayerCounter)
	registerPrimitiveHandler(reg, handleMoveCounters)
	registerPrimitiveHandler(reg, handleApplyContinuous)
	registerPrimitiveHandler(reg, handleApplyRule)
	registerPrimitiveHandler(reg, handleModifyPT)
	registerPrimitiveHandler(reg, handleFight)
	registerPrimitiveHandler(reg, handleTap)
	registerPrimitiveHandler(reg, handleSearch)
	registerPrimitiveHandler(reg, handleReveal)
	registerPrimitiveHandler(reg, handlePutOnBattlefield)
	registerPrimitiveHandler(reg, handleCreateToken)
	registerPrimitiveHandler(reg, handleShufflePermanentIntoLibrary)
	registerPrimitiveHandler(reg, handlePutPermanentOnLibrary)
	registerPrimitiveHandler(reg, handleStartEngines)
	registerPrimitiveHandler(reg, handleSetClassLevel)
	registerPrimitiveHandler(reg, handleMonstrosity)
	registerPrimitiveHandler(reg, handleDiscoverCards)
	registerPrimitiveHandler(reg, handlePay)
	registerPrimitiveHandler(reg, handleChoose)
	registerPrimitiveHandler(reg, handleGainLife)
	registerPrimitiveHandler(reg, handleLoseLife)
	registerPrimitiveHandler(reg, handlePlayerLosesGame)
	registerPrimitiveHandler(reg, handleExile)
	registerPrimitiveHandler(reg, handleExileFromHand)
	registerPrimitiveHandler(reg, handlePutFromHand)
	registerPrimitiveHandler(reg, handleCastForFree)
	registerPrimitiveHandler(reg, handleReturnFromGraveyard)
	registerPrimitiveHandler(reg, handleMassReturnFromGraveyard)
	registerPrimitiveHandler(reg, handleBounce)
	registerPrimitiveHandler(reg, handleSacrifice)
	registerPrimitiveHandler(reg, handleSacrificePermanents)
	registerPrimitiveHandler(reg, handleUntap)
	registerPrimitiveHandler(reg, handleSkipNextUntap)
	registerPrimitiveHandler(reg, handleCounterObject)
	registerPrimitiveHandler(reg, handleChooseNewTargets)
	registerPrimitiveHandler(reg, handleMill)
	registerPrimitiveHandler(reg, handleScry)
	registerPrimitiveHandler(reg, handleSurveil)
	registerPrimitiveHandler(reg, handleDig)
	registerPrimitiveHandler(reg, handleImpulseExile)
	registerPrimitiveHandler(reg, handleInvestigate)
	registerPrimitiveHandler(reg, handleProliferate)
	registerPrimitiveHandler(reg, handleExplore)
	registerPrimitiveHandler(reg, handleManifest)
	registerPrimitiveHandler(reg, handleGoad)
	registerPrimitiveHandler(reg, handleRemoveCounter)
	registerPrimitiveHandler(reg, handleTransform)
	registerPrimitiveHandler(reg, handlePhaseOut)
	registerPrimitiveHandler(reg, handleRegenerate)
	registerPrimitiveHandler(reg, handleSkipStep)
	registerPrimitiveHandler(reg, handleCreateEmblem)
	registerPrimitiveHandler(reg, handleCreateDelayedTrigger)
	registerPrimitiveHandler(reg, handleCreateReplacement)
	registerPrimitiveHandler(reg, handlePreventDamage)
	registerPrimitiveHandler(reg, handleMoveCard)
	registerPrimitiveHandler(reg, handleMoveCommander)
	registerPrimitiveHandler(reg, handleGrantCastPermission)
	registerPrimitiveHandler(reg, handleAttach)
	registerPrimitiveHandler(reg, handleReorderLibraryTop)
	registerPrimitiveHandler(reg, handleShuffleLibrary)
	registerPrimitiveHandler(reg, handleLookAtLibraryTop)
	registerPrimitiveHandler(reg, handleGroupSourceDamage)
	return reg
}

// globalPrimitiveRegistry is the singleton handler table.
var globalPrimitiveRegistry = newPrimitiveRegistry()
