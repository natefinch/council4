package rules

import (
	"fmt"
	"sync"

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
	registerPrimitiveHandler(reg, handlePlayerMayPayGenericOrRule)
	registerPrimitiveHandler(reg, handleModifyPT)
	registerPrimitiveHandler(reg, handleFight)
	registerPrimitiveHandler(reg, handleTap)
	registerPrimitiveHandler(reg, handleTapOrUntap)
	registerPrimitiveHandler(reg, handleSearch)
	registerPrimitiveHandler(reg, handleReveal)
	registerPrimitiveHandler(reg, handlePutOnBattlefield)
	registerPrimitiveHandler(reg, handleCreateToken)
	registerPrimitiveHandler(reg, handleShufflePermanentIntoLibrary)
	registerPrimitiveHandler(reg, handleShuffleSpellIntoLibrary)
	registerPrimitiveHandler(reg, handlePutPermanentOnLibrary)
	registerPrimitiveHandler(reg, handlePutLinkedExiledCardsInLibrary)
	registerPrimitiveHandler(reg, handleConditionalDestinationPlace)
	registerPrimitiveHandler(reg, handleStartEngines)
	registerPrimitiveHandler(reg, handleBecomeMonarch)
	registerPrimitiveHandler(reg, handleCantBecomeMonarch)
	registerPrimitiveHandler(reg, handlePartitionExiledCostCards)
	registerPrimitiveHandler(reg, handleRingTempts)
	registerPrimitiveHandler(reg, handleVote)
	registerPrimitiveHandler(reg, handleSetClassLevel)
	registerPrimitiveHandler(reg, handleMonstrosity)
	registerPrimitiveHandler(reg, handleRenown)
	registerPrimitiveHandler(reg, handleAdapt)
	registerPrimitiveHandler(reg, handleConnive)
	registerPrimitiveHandler(reg, handleBecomeSaddled)
	registerPrimitiveHandler(reg, handleDiscoverCards)
	registerPrimitiveHandler(reg, handlePay)
	registerPrimitiveHandler(reg, handlePayRepeatedly)
	registerPrimitiveHandler(reg, handleChoose)
	registerPrimitiveHandler(reg, handleGainLife)
	registerPrimitiveHandler(reg, handleLoseLife)
	registerPrimitiveHandler(reg, handlePlayerLosesGame)
	registerPrimitiveHandler(reg, handlePlayerWinsGame)
	registerPrimitiveHandler(reg, handleExile)
	registerPrimitiveHandler(reg, handleExileEntireHand)
	registerPrimitiveHandler(reg, handleReturnExiledCardsToHand)
	registerPrimitiveHandler(reg, handleExileForEachPlayer)
	registerPrimitiveHandler(reg, handleChampionExile)
	registerPrimitiveHandler(reg, handleReturnLinkedExiledCardsToBattlefield)
	registerPrimitiveHandler(reg, handleDestroyForEachPlayer)
	registerPrimitiveHandler(reg, handleEachPlayerChooseDestroy)
	registerPrimitiveHandler(reg, handleCreateTokenForEachDestroyed)
	registerPrimitiveHandler(reg, handleExileForEachOpponent)
	registerPrimitiveHandler(reg, handleDrawForEachExiled)
	registerPrimitiveHandler(reg, handleRemoveTargetsForToken)
	registerPrimitiveHandler(reg, handleCastForFree)
	registerPrimitiveHandler(reg, handleChooseFromZone)
	registerPrimitiveHandler(reg, handleMassReturnFromGraveyard)
	registerPrimitiveHandler(reg, handleMassReanimationExchange)
	registerPrimitiveHandler(reg, handleBounce)
	registerPrimitiveHandler(reg, handleSacrifice)
	registerPrimitiveHandler(reg, handleSacrificePermanents)
	registerPrimitiveHandler(reg, handleUntap)
	registerPrimitiveHandler(reg, handleSkipNextUntap)
	registerPrimitiveHandler(reg, handleRemoveFromCombat)
	registerPrimitiveHandler(reg, handleCounterObject)
	registerPrimitiveHandler(reg, handleChooseNewTargets)
	registerPrimitiveHandler(reg, handleCopyStackObject)
	registerPrimitiveHandler(reg, handleMill)
	registerPrimitiveHandler(reg, handleExileTopOfLibrary)
	registerPrimitiveHandler(reg, handleRevealUntil)
	registerPrimitiveHandler(reg, handlePutHandOnLibraryThenDraw)
	registerPrimitiveHandler(reg, handleDiscardThenDraw)
	registerPrimitiveHandler(reg, handleDiscardUnlessType)
	registerPrimitiveHandler(reg, handleScry)
	registerPrimitiveHandler(reg, handleSurveil)
	registerPrimitiveHandler(reg, handleDig)
	registerPrimitiveHandler(reg, handlePileSplit)
	registerPrimitiveHandler(reg, handleRevealTopPartition)
	registerPrimitiveHandler(reg, handleImpulseExile)
	registerPrimitiveHandler(reg, handleExileLibraryUntilNonlandCast)
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
	registerPrimitiveHandler(reg, handleCreateReflexiveTrigger)
	registerPrimitiveHandler(reg, handleCreateReplacement)
	registerPrimitiveHandler(reg, handlePreventDamage)
	registerPrimitiveHandler(reg, handleMoveCard)
	registerPrimitiveHandler(reg, handleMoveCommander)
	registerPrimitiveHandler(reg, handleGrantCastPermission)
	registerPrimitiveHandler(reg, handleExileForPlay)
	registerPrimitiveHandler(reg, handleAttach)
	registerPrimitiveHandler(reg, handleReorderLibraryTop)
	registerPrimitiveHandler(reg, handleShuffleLibrary)
	registerPrimitiveHandler(reg, handleShuffleGraveyardIntoLibrary)
	registerPrimitiveHandler(reg, handleLookAtHand)
	registerPrimitiveHandler(reg, handleChooseDiscardFromHand)
	registerPrimitiveHandler(reg, handleLookAtLibraryTop)
	registerPrimitiveHandler(reg, handleGroupSourceDamage)
	registerPrimitiveHandler(reg, handleGroupSelfPowerDamage)
	registerPrimitiveHandler(reg, handlePunisherEachLoseLife)
	registerPrimitiveHandler(reg, handleRepeatProcess)
	registerPrimitiveHandler(reg, handleBecomeCopy)
	registerPrimitiveHandler(reg, handleAmass)
	registerPrimitiveHandler(reg, handleAddExtraPhases)
	registerPrimitiveHandler(reg, handleRollDie)
	registerPrimitiveHandler(reg, handleHideawayExile)
	registerPrimitiveHandler(reg, handlePlayHideawayCard)
	return reg
}

// globalPrimitiveRegistry returns the singleton handler table. It is built
// lazily on first use rather than in a var initializer so handlers may
// statically reference the resolution machinery (which reads this registry)
// without forming an initialization cycle.
func globalPrimitiveRegistry() *primitiveRegistry {
	globalPrimitiveRegistryOnce.Do(func() {
		globalPrimitiveRegistryInstance = newPrimitiveRegistry()
	})
	return globalPrimitiveRegistryInstance
}

var (
	globalPrimitiveRegistryOnce     sync.Once
	globalPrimitiveRegistryInstance *primitiveRegistry
)
