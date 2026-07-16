package game

import (
	"maps"
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

// Clone returns a deep copy of the entire game state. Mutating the clone never
// affects the original, and vice versa, which makes Clone suitable for snapshot
// replay, scenario fixtures, and search techniques such as MCTS.
//
// Sharing rules:
//   - *CardDef values are immutable shared definitions and are deliberately kept
//     shared (never deep-copied). The same holds for ability values such as
//     TriggeredAbility/ActivatedAbility/LoyaltyAbility, which are immutable rules
//     data derived from card definitions.
//   - Every mutable container (slices, maps) is reallocated so no backing array
//     or map is shared between the original and the clone.
//
// RNG: rand/v2 does not expose its source, so the stream cannot be cloned
// bit-for-bit. The clone receives a fresh, independent, deterministic *rand.Rand
// seeded from the IDGen counter. This is intentional: two clones must not draw
// from a shared stream, and an independent reproducible stream suits MCTS
// determinization. The clone's random stream is NOT a continuation of the
// original's.
func (g *Game) Clone() *Game {
	clone := &Game{
		Battlefield:                        cloneSliceFunc(g.Battlefield, clonePermanent),
		CardInstances:                      clonePtrMap(g.CardInstances, cloneCardInstance),
		CommanderIDs:                       cloneComparableMap(g.CommanderIDs),
		Stack:                              cloneStack(g.Stack),
		ContinuousEffects:                  cloneSlicePtr(g.ContinuousEffects, fixupContinuousEffect),
		DelayedTriggers:                    cloneDelayedTriggers(g.DelayedTriggers),
		PendingReflexiveTriggers:           cloneReflexiveTriggers(g.PendingReflexiveTriggers),
		PendingRoomAbilities:               cloneSlice(g.PendingRoomAbilities),
		PendingInitiativeVentures:          cloneSlice(g.PendingInitiativeVentures),
		PreventionShields:                  cloneSlice(g.PreventionShields),
		ReplacementDecisions:               cloneSliceFunc(g.ReplacementDecisions, cloneReplacementDecision),
		ReplacementEffects:                 cloneSlicePtr(g.ReplacementEffects, fixupReplacementEffect),
		SkippedSteps:                       cloneSkippedSteps(g.SkippedSteps),
		CostModifiers:                      cloneSlice(g.CostModifiers),
		RuleEffects:                        cloneSlicePtr(g.RuleEffects, fixupRuleEffect),
		SuspendedCards:                     cloneComparableMap(g.SuspendedCards),
		ReboundCards:                       cloneComparableMap(g.ReboundCards),
		AdventureCards:                     cloneComparableMap(g.AdventureCards),
		PlottedCards:                       cloneComparableMap(g.PlottedCards),
		ForetoldCards:                      cloneComparableMap(g.ForetoldCards),
		ExileCounters:                      cloneMapFunc(g.ExileCounters, cloneCounterSet),
		ExileCounterExiledBy:               cloneComparableMap(g.ExileCounterExiledBy),
		LastKnownInformation:               cloneMapFunc(g.LastKnownInformation, cloneObjectSnapshot),
		LinkedObjects:                      cloneLinkedObjects(g.LinkedObjects),
		Turn:                               cloneTurnState(g.Turn),
		TurnOrder:                          cloneTurnOrder(g.TurnOrder),
		FailedDraws:                        cloneComparableMap(g.FailedDraws),
		MarkedToLoseGame:                   cloneComparableMap(g.MarkedToLoseGame),
		Combat:                             cloneCombatState(g.Combat),
		Emblems:                            cloneSliceFunc(g.Emblems, cloneEmblem),
		DayNight:                           cloneDayNight(g.DayNight),
		Events:                             cloneEvents(g.Events),
		EventTurnStarts:                    cloneSlice(g.EventTurnStarts),
		TriggerEventCursor:                 g.TriggerEventCursor,
		StateTriggerLatches:                cloneComparableMap(g.StateTriggerLatches),
		FiredManaSpendRiders:               cloneSlice(g.FiredManaSpendRiders),
		ActivatedAbilitiesThisTurn:         cloneComparableMap(g.ActivatedAbilitiesThisTurn),
		AbilityActivationsThisTurn:         cloneComparableMap(g.AbilityActivationsThisTurn),
		ExilePlayPermissionUsedThisTurn:    cloneComparableMap(g.ExilePlayPermissionUsedThisTurn),
		TriggeredAbilitiesThisTurn:         cloneComparableMap(g.TriggeredAbilitiesThisTurn),
		ResolvedTriggeredAbilitiesThisTurn: cloneComparableMap(g.ResolvedTriggeredAbilitiesThisTurn),
		ChosenModesThisTurn:                cloneComparableMap(g.ChosenModesThisTurn),
	}

	clone.IDGen.Restore(g.IDGen.Current())

	for i := range g.Players {
		clone.Players[i] = clonePlayer(g.Players[i])
	}

	seed := g.IDGen.Current()
	clone.RNG = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))

	return clone
}

func cloneDelayedTriggers(triggers []DelayedTrigger) []DelayedTrigger {
	clone := slices.Clone(triggers)
	for i := range clone {
		clone[i].CapturedTargetControllerLKI = cloneComparableMap(clone[i].CapturedTargetControllerLKI)
		clone[i].CapturedTargetManaValueLKI = cloneComparableMap(clone[i].CapturedTargetManaValueLKI)
	}
	return clone
}

func cloneReflexiveTriggers(triggers []ReflexiveTrigger) []ReflexiveTrigger {
	clone := slices.Clone(triggers)
	for i := range clone {
		clone[i].CapturedTargetControllerLKI = cloneComparableMap(clone[i].CapturedTargetControllerLKI)
		clone[i].CapturedTargetManaValueLKI = cloneComparableMap(clone[i].CapturedTargetManaValueLKI)
		clone[i].TriggerEvent = cloneEvent(clone[i].TriggerEvent)
	}
	return clone
}

// --- Generic container helpers ---

// cloneSlice returns a new slice with the same elements. It is a full deep copy
// for slices of value types, and a fresh backing array sharing immutable element
// values (such as Ability interfaces or *CardDef pointers) otherwise.
func cloneSlice[T any](s []T) []T {
	if s == nil {
		return nil
	}
	return slices.Clone(s)
}

// cloneSliceFunc returns a new slice whose elements are produced by clone.
func cloneSliceFunc[T any](s []T, clone func(T) T) []T {
	if s == nil {
		return nil
	}
	out := make([]T, len(s))
	for i := range s {
		out[i] = clone(s[i])
	}
	return out
}

// cloneSlicePtr returns a new slice whose elements are deep-copied in place by
// fix. It avoids passing large element structs by value.
func cloneSlicePtr[T any](s []T, fix func(*T)) []T {
	if s == nil {
		return nil
	}
	out := slices.Clone(s)
	for i := range out {
		fix(&out[i])
	}
	return out
}

// cloneComparableMap returns a shallow copy of a map with value (non-container)
// values. The values must not themselves contain shared mutable containers.
func cloneComparableMap[K comparable, V any](m map[K]V) map[K]V {
	return maps.Clone(m)
}

// cloneMapFunc returns a new map whose values are produced by clone.
func cloneMapFunc[K comparable, V any](m map[K]V, clone func(V) V) map[K]V {
	if m == nil {
		return nil
	}
	out := make(map[K]V, len(m))
	for k, v := range m {
		out[k] = clone(v)
	}
	return out
}

// cloneCounterSet deep-copies a counter set value so the cloned game shares no
// counter map state with the original.
func cloneCounterSet(s counter.Set) counter.Set {
	return s.Clone()
}

// clonePtrMap returns a new map whose pointer values are deep-copied by clone.
func clonePtrMap[K comparable, V any](m map[K]*V, clone func(*V) *V) map[K]*V {
	if m == nil {
		return nil
	}
	out := make(map[K]*V, len(m))
	for k, v := range m {
		out[k] = clone(v)
	}
	return out
}

// --- Per-type clone functions ---

func clonePlayer(p *Player) *Player {
	clone := *p
	clone.CommanderDamage = cloneComparableMap(p.CommanderDamage)
	clone.ManaPool = p.ManaPool.Clone()
	clone.ManaRiders = cloneSlice(p.ManaRiders)
	clone.Library = p.Library.Clone()
	clone.Hand = p.Hand.Clone()
	clone.Graveyard = p.Graveyard.Clone()
	clone.Exile = p.Exile.Clone()
	clone.CommandZone = p.CommandZone.Clone()
	return &clone
}

// cloneCardInstance copies the instance but shares the immutable *CardDef.
func cloneCardInstance(ci *CardInstance) *CardInstance {
	clone := *ci
	return &clone
}

func clonePermanent(p *Permanent) *Permanent {
	clone := *p
	clone.MergedCards = cloneSlice(p.MergedCards)
	clone.Counters = p.Counters.Clone()
	clone.Attachments = cloneSlice(p.Attachments)
	clone.Goaded = cloneComparableMap(p.Goaded)
	clone.EntryChoices = cloneComparableMap(p.EntryChoices)
	// TokenDef is an immutable shared definition and is intentionally shared.
	return &clone
}

func cloneStack(s Stack) Stack {
	return Stack{objects: cloneSliceFunc(s.objects, cloneStackObject)}
}

// NewStackObjectCopy returns a copy of o as a new stack object on the stack
// (CR 707) with a fresh ID and Copy set. The copy owns its own mutable target
// and resolution state while sharing immutable rules data (the inline ability
// bodies and source token definition), so re-choosing the copy's targets never
// disturbs the original.
func NewStackObjectCopy(o *StackObject, newID id.ID) *StackObject {
	clone := cloneStackObject(o)
	clone.ID = newID
	clone.Copy = true
	return clone
}

func cloneStackObject(o *StackObject) *StackObject {
	clone := *o
	clone.TriggerEvent = cloneEvent(o.TriggerEvent)
	clone.Targets = cloneSlice(o.Targets)
	clone.TargetCounts = cloneSlice(o.TargetCounts)
	clone.SplicedContent = cloneSliceFunc(o.SplicedContent, cloneAbilityContent)
	clone.SplicedTargets = cloneSliceFunc(o.SplicedTargets, cloneSlice[Target])
	clone.SplicedTargetCounts = cloneSliceFunc(o.SplicedTargetCounts, cloneSlice[int])
	clone.ChosenModes = cloneSlice(o.ChosenModes)
	clone.RuleEffects = cloneSliceFunc(o.RuleEffects, func(effect RuleEffect) RuleEffect {
		fixupRuleEffect(&effect)
		return effect
	})
	clone.AdditionalCostsPaid = cloneSlice(o.AdditionalCostsPaid)
	clone.SacrificedAsCostIDs = cloneSlice(o.SacrificedAsCostIDs)
	clone.ExiledAsCostIDs = cloneSlice(o.ExiledAsCostIDs)
	clone.GainsKeywordsUntilEndOfTurn = cloneSlice(o.GainsKeywordsUntilEndOfTurn)
	clone.ResolvedAmounts = cloneComparableMap(o.ResolvedAmounts)
	clone.ResolvedExcessDamage = cloneComparableMap(o.ResolvedExcessDamage)
	clone.ResolutionResults = cloneComparableMap(o.ResolutionResults)
	clone.ResolutionChoices = cloneComparableMap(o.ResolutionChoices)
	clone.TargetControllerLKI = cloneComparableMap(o.TargetControllerLKI)
	clone.TargetManaValueLKI = cloneComparableMap(o.TargetManaValueLKI)
	clone.TargetNameLKI = cloneComparableMap(o.TargetNameLKI)
	clone.CapturedTargetControllerLKI = cloneComparableMap(o.CapturedTargetControllerLKI)
	clone.CapturedTargetManaValueLKI = cloneComparableMap(o.CapturedTargetManaValueLKI)
	// InlineTrigger/InlineActivated/InlineLoyalty and SourceTokenDef are
	// immutable rules data and are intentionally shared.
	return &clone
}

func fixupContinuousEffect(e *ContinuousEffect) {
	e.DependsOn = cloneSlice(e.DependsOn)
	e.SetSupertypes = cloneSlice(e.SetSupertypes)
	e.AddSupertypes = cloneSlice(e.AddSupertypes)
	e.RemoveSupertypes = cloneSlice(e.RemoveSupertypes)
	e.SetTypes = cloneSlice(e.SetTypes)
	e.AddTypes = cloneSlice(e.AddTypes)
	e.RemoveTypes = cloneSlice(e.RemoveTypes)
	e.SetSubtypes = cloneSlice(e.SetSubtypes)
	e.AddSubtypes = cloneSlice(e.AddSubtypes)
	e.RemoveSubtypes = cloneSlice(e.RemoveSubtypes)
	e.SetColors = cloneSlice(e.SetColors)
	e.AddColors = cloneSlice(e.AddColors)
	e.RemoveColors = cloneSlice(e.RemoveColors)
	e.AddKeywords = cloneSlice(e.AddKeywords)
	e.RemoveKeywords = cloneSlice(e.RemoveKeywords)
	// AddAbilities holds immutable Ability values; a fresh backing array is
	// enough. Group, CopyValues, and the dynamic deltas are immutable rules
	// configuration and are copied by value with the surrounding struct.
	e.AddAbilities = cloneSlice(e.AddAbilities)
}

func cloneReplacementDecision(d ReplacementDecision) ReplacementDecision {
	d.Options = cloneSlice(d.Options)
	d.Selected = cloneSlice(d.Selected)
	return d
}

func fixupReplacementEffect(e *ReplacementEffect) {
	e.DamageSourceColors = cloneSlice(e.DamageSourceColors)
	e.EntersWithCounters = cloneSlice(e.EntersWithCounters)
	// Condition, Selection, EntersAsCopySelection, EntersWithCountersRecipient,
	// CounterRecipientSelection, and EntersTappedSelection are immutable rules
	// data and are copied by value or shared pointer.
}

func fixupRuleEffect(e *RuleEffect) {
	e.PermanentTypes = cloneSlice(e.PermanentTypes)
	e.SpellTypes = cloneSlice(e.SpellTypes)
	e.SpellSubtypes = cloneSlice(e.SpellSubtypes)
	e.CantCastFromZones = cloneSlice(e.CantCastFromZones)
	e.EnterFromZones = cloneSlice(e.EnterFromZones)
	e.Protection.FromColors = cloneSlice(e.Protection.FromColors)
	e.Protection.FromTypes = cloneSlice(e.Protection.FromTypes)
	e.Protection.FromSubtypes = cloneSlice(e.Protection.FromSubtypes)
	// CostModifier, CardSelection, and GrantedAbility are immutable rules data
	// copied by value.
}

func cloneEmblem(e Emblem) Emblem {
	// Abilities are immutable rules data; a fresh backing array is enough.
	e.Abilities = cloneSlice(e.Abilities)
	return e
}

func cloneObjectSnapshot(s ObjectSnapshot) ObjectSnapshot {
	s.MergedCards = cloneSlice(s.MergedCards)
	s.Colors = cloneSlice(s.Colors)
	s.Supertypes = cloneSlice(s.Supertypes)
	s.Types = cloneSlice(s.Types)
	s.Subtypes = cloneSlice(s.Subtypes)
	s.Keywords = cloneSlice(s.Keywords)
	s.Attachments = cloneSlice(s.Attachments)
	s.Counters = s.Counters.Clone()
	s.EntryChoices = cloneComparableMap(s.EntryChoices)
	s.RuleEffectKinds = cloneSlice(s.RuleEffectKinds)
	// TokenDef is an immutable shared definition.
	return s
}

func cloneCombatState(c *CombatState) *CombatState {
	if c == nil {
		return nil
	}
	clone := *c
	clone.Attackers = cloneSlice(c.Attackers)
	clone.PlayersAttacked = cloneComparableMap(c.PlayersAttacked)
	clone.Blockers = cloneSlice(c.Blockers)
	clone.BlockedAttackers = cloneComparableMap(c.BlockedAttackers)
	clone.DamageAssignment = cloneComparableMap(c.DamageAssignment)
	clone.BlockerOrder = cloneBlockerOrder(c.BlockerOrder)
	return &clone
}

func cloneBlockerOrder(m map[id.ID][]id.ID) map[id.ID][]id.ID {
	if m == nil {
		return nil
	}
	out := make(map[id.ID][]id.ID, len(m))
	for k, v := range m {
		out[k] = cloneSlice(v)
	}
	return out
}

func cloneTurnState(t TurnState) TurnState {
	t.ExtraTurns = cloneSlice(t.ExtraTurns)
	t.ExtraPhases = cloneSlice(t.ExtraPhases)
	return t
}

func cloneTurnOrder(t TurnOrder) TurnOrder {
	t.Eliminated = cloneComparableMap(t.Eliminated)
	return t
}

func cloneDayNight(d *DayNightState) *DayNightState {
	if d == nil {
		return nil
	}
	clone := *d
	return &clone
}

func cloneSkippedSteps(m map[PlayerID]map[Step]int) map[PlayerID]map[Step]int {
	if m == nil {
		return nil
	}
	out := make(map[PlayerID]map[Step]int, len(m))
	for k, v := range m {
		out[k] = cloneComparableMap(v)
	}
	return out
}

func cloneLinkedObjects(m map[LinkedObjectKey][]LinkedObjectRef) map[LinkedObjectKey][]LinkedObjectRef {
	if m == nil {
		return nil
	}
	out := make(map[LinkedObjectKey][]LinkedObjectRef, len(m))
	for k, v := range m {
		out[k] = cloneSlice(v)
	}
	return out
}
