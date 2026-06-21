package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// selectionSubjectKind identifies which characteristic source a selectionSubject
// reads from. It lets one Selection matcher serve targets, controller-controls
// conditions, trigger event subjects, cast spells, and mass effects without
// interface dispatch.
type selectionSubjectKind int

const (
	// subjectPermanent reads a live permanent's effective or base values.
	subjectPermanent selectionSubjectKind = iota
	// subjectEventPermanent reads a triggering event's permanent, including
	// last-known information, the cast card, or a token definition.
	subjectEventPermanent
	// subjectCastSpell reads a cast spell's card types from the event.
	subjectCastSpell
	// subjectCard reads printed characteristics from a card in a non-battlefield zone.
	subjectCard
)

// selectionSubject is the context a Selection is matched against. It captures
// the genuine per-context differences (effective vs base values, controller
// relativity, source exclusion, event last-known information) while the field
// semantics live once in matchSelection.
type selectionSubject struct {
	kind selectionSubjectKind
	g    *game.Game

	// permanent and values back subjectPermanent matching. values is the
	// already-computed effective or base value set so hot paths stay
	// allocation-free.
	permanent *game.Permanent
	values    *permanentEffectiveValues

	// controller is the subject's resolved controller for ControllerRelation
	// checks; viewer is the player that "you" is relative to.
	controller game.PlayerID
	viewer     game.PlayerID

	// sourceObjectID is the predicate source object excluded by ExcludeSource.
	sourceObjectID id.ID

	// clampPower selects the target-style power read (clamped to >= 0, always
	// applicable) over the strict controller-controls read (requires printed
	// power). useBase forfeits power and toughness, preserving the base
	// characteristic condition behavior.
	clampPower bool
	useBase    bool

	// event and cardTypes back the event-permanent and cast-spell subjects.
	event     game.Event
	cardTypes []types.Card
	card      *game.CardInstance

	// resolutionChoices holds the resolving stack object's published choices so a
	// SubtypeChoiceResolution predicate can read the creature type chosen earlier in
	// the same resolution. It is set only for group-membership matching, where the
	// resolving object is available; other callers leave it nil and a
	// SubtypeChoiceResolution predicate matches nothing.
	resolutionChoices map[string]game.ResolutionChoiceResult
}

// matchSelection reports whether the subject satisfies every active predicate in
// sel. It is the single implementation of Selection field semantics; callers
// supply context through the subject.
func matchSelection(s *selectionSubject, sel *game.Selection) bool {
	if s.kind == subjectPermanent && !activeBattlefieldPermanent(s.permanent) {
		return false
	}
	if len(sel.AnyOf) > 0 && !slices.ContainsFunc(sel.AnyOf, func(alternative game.Selection) bool {
		return matchSelection(s, &alternative)
	}) {
		return false
	}
	for _, cardType := range sel.RequiredTypes {
		if !s.hasType(cardType) {
			return false
		}
	}
	if len(sel.RequiredTypesAny) > 0 && !s.hasAnyType(sel.RequiredTypesAny) {
		return false
	}
	if slices.ContainsFunc(sel.ExcludedTypes, s.hasType) {
		return false
	}
	for _, supertype := range sel.Supertypes {
		if !s.hasSupertype(supertype) {
			return false
		}
	}
	if sel.ExcludedSupertype != "" && s.hasSupertype(sel.ExcludedSupertype) {
		return false
	}
	if len(sel.SubtypesAny) > 0 && !s.hasAnySubtype(sel.SubtypesAny) {
		return false
	}
	if sel.ExcludedSubtype != "" && s.hasAnySubtype([]types.Sub{sel.ExcludedSubtype}) {
		return false
	}
	if sel.SubtypeChoice == game.SubtypeChoiceSourceEntry {
		subtype, ok := s.sourceEntryChoiceSubtype(game.EntryTypeChoiceKey)
		if !ok || !s.hasAnySubtype([]types.Sub{subtype}) {
			return false
		}
	}
	if sel.SubtypeChoice == game.SubtypeChoiceResolution {
		subtype, ok := s.resolutionChoiceSubtype(game.SpellChosenTypeChoiceKey)
		if !ok || !s.hasAnySubtype([]types.Sub{subtype}) {
			return false
		}
	}
	if sel.SubtypeChoice == game.SubtypeChoiceResolutionExcluded {
		subtype, ok := s.resolutionChoiceSubtype(game.SpellChosenTypeChoiceKey)
		if !ok || s.hasAnySubtype([]types.Sub{subtype}) {
			return false
		}
	}
	if len(sel.ColorsAny) > 0 && !s.hasAnyColor(sel.ColorsAny) {
		return false
	}
	if slices.ContainsFunc(sel.ExcludedColors, s.hasColor) {
		return false
	}
	if sel.Colorless && s.colorCount() != 0 {
		return false
	}
	if sel.Multicolored && s.colorCount() < 2 {
		return false
	}
	if !s.controllerMatches(sel.Controller) {
		return false
	}
	if sel.Tapped == game.TriTrue && !s.tapped() {
		return false
	}
	if sel.Tapped == game.TriFalse && s.tapped() {
		return false
	}
	if !s.combatStateMatches(sel.CombatState) {
		return false
	}
	if sel.Keyword != game.KeywordNone && !s.hasKeyword(sel.Keyword) {
		return false
	}
	if sel.ExcludedKeyword != game.KeywordNone && s.hasKeyword(sel.ExcludedKeyword) {
		return false
	}
	if sel.ManaValue.Exists {
		manaValue, ok := s.manaValue()
		if !ok || !sel.ManaValue.Val.Matches(manaValue) {
			return false
		}
	}
	if sel.Power.Exists {
		power, ok := s.power()
		if !ok || !sel.Power.Val.Matches(power) {
			return false
		}
	}
	if sel.Toughness.Exists {
		toughness, ok := s.toughness()
		if !ok || !sel.Toughness.Val.Matches(toughness) {
			return false
		}
	}
	if sel.MatchCounter && !s.hasCounter(sel.RequiredCounter) {
		return false
	}
	if sel.MatchAnyCounter && !s.hasAnyCounter() {
		return false
	}
	if sel.EnteredThisTurn && !s.enteredThisTurn() {
		return false
	}
	if sel.MatchModified && !s.modified() {
		return false
	}
	if sel.ExcludeSource && s.isSource() {
		return false
	}
	if sel.NonToken && s.isToken() {
		return false
	}
	if sel.TokenOnly && !s.isToken() {
		return false
	}
	return true
}

func (s *selectionSubject) hasType(cardType types.Card) bool {
	switch s.kind {
	case subjectPermanent:
		return slices.Contains(s.values.types, cardType)
	case subjectEventPermanent:
		return eventPermanentHasType(s.g, s.event, cardType)
	case subjectCastSpell:
		return slices.Contains(s.cardTypes, cardType)
	case subjectCard:
		if s.card == nil || s.card.Def == nil {
			return false
		}
		return slices.Contains(s.card.Def.DefaultFace().Types, cardType)
	default:
		return false
	}
}

func (s *selectionSubject) hasAnyType(cardTypes []types.Card) bool {
	return slices.ContainsFunc(cardTypes, s.hasType)
}

func (s *selectionSubject) hasSupertype(supertype types.Super) bool {
	if s.kind == subjectPermanent {
		return slices.Contains(s.values.supertypes, supertype)
	}
	if s.kind == subjectEventPermanent {
		if values, ok := s.eventPermanentValues(); ok {
			return slices.Contains(values.supertypes, supertype)
		}
		if def, ok := s.eventPermanentCardDef(); ok {
			return slices.Contains(def.DefaultFace().Supertypes, supertype)
		}
	}
	if s.kind == subjectCastSpell {
		return slices.Contains(s.event.CardSupertypes, supertype)
	}
	if s.kind == subjectCard && s.card != nil && s.card.Def != nil {
		return slices.Contains(s.card.Def.DefaultFace().Supertypes, supertype)
	}
	return false
}

func (s *selectionSubject) hasAnySubtype(subtypes []types.Sub) bool {
	if s.kind == subjectPermanent {
		for _, subtype := range subtypes {
			if slices.Contains(s.values.subtypes, subtype) {
				return true
			}
		}
	}
	if s.kind == subjectEventPermanent {
		if values, ok := s.eventPermanentValues(); ok {
			return slices.ContainsFunc(subtypes, func(subtype types.Sub) bool {
				return slices.Contains(values.subtypes, subtype)
			})
		}
		if def, ok := s.eventPermanentCardDef(); ok {
			return slices.ContainsFunc(subtypes, func(subtype types.Sub) bool {
				return def.HasSubtype(subtype)
			})
		}
	}
	if s.kind == subjectCastSpell {
		for _, subtype := range subtypes {
			if slices.Contains(s.event.CardSubtypes, subtype) {
				return true
			}
		}
	}
	if s.kind == subjectCard && s.card != nil && s.card.Def != nil {
		for _, subtype := range subtypes {
			if slices.Contains(s.card.Def.DefaultFace().Subtypes, subtype) {
				return true
			}
		}
	}
	return false
}

// sourceEntryChoiceSubtype resolves the creature subtype the predicate's source
// permanent recorded under key as it entered the battlefield. It reports false
// when the source permanent, the choice, or its subtype is absent.
func (s *selectionSubject) sourceEntryChoiceSubtype(key game.ChoiceKey) (types.Sub, bool) {
	source, ok := permanentByObjectID(s.g, s.sourceObjectID)
	if !ok {
		return "", false
	}
	choice, ok := source.EntryChoices[key]
	if !ok || choice.Kind != game.ResolutionChoiceSubtype || choice.Subtype == "" {
		return "", false
	}
	return choice.Subtype, true
}

// resolutionChoiceSubtype resolves the creature subtype published under key by an
// earlier Choose instruction in the same resolution (the "of that type"
// back-reference). It reports false when the choice is absent, is not a subtype
// choice, or carries no subtype.
func (s *selectionSubject) resolutionChoiceSubtype(key game.ChoiceKey) (types.Sub, bool) {
	if s.resolutionChoices == nil {
		return "", false
	}
	choice, ok := s.resolutionChoices[string(key)]
	if !ok || choice.Kind != game.ResolutionChoiceSubtype || choice.Subtype == "" {
		return "", false
	}
	return choice.Subtype, true
}

func (s *selectionSubject) hasColor(c color.Color) bool {
	if s.kind == subjectPermanent {
		return slices.Contains(s.values.colors, c)
	}
	if s.kind == subjectEventPermanent {
		if values, ok := s.eventPermanentValues(); ok {
			return slices.Contains(values.colors, c)
		}
		if def, ok := s.eventPermanentCardDef(); ok {
			return slices.Contains(def.DefaultFace().Colors, c)
		}
	}
	if s.kind == subjectCastSpell {
		return slices.Contains(s.event.Colors, c)
	}
	if s.kind == subjectCard && s.card != nil && s.card.Def != nil {
		return slices.Contains(s.card.Def.DefaultFace().Colors, c)
	}
	return false
}

func (s *selectionSubject) hasAnyColor(colors []color.Color) bool {
	return slices.ContainsFunc(colors, s.hasColor)
}

func (s *selectionSubject) colorCount() int {
	seen := map[color.Color]bool{}
	var colors []color.Color
	switch s.kind {
	case subjectPermanent:
		colors = s.values.colors
	case subjectEventPermanent:
		if values, ok := s.eventPermanentValues(); ok {
			colors = values.colors
		} else if def, ok := s.eventPermanentCardDef(); ok {
			colors = def.DefaultFace().Colors
		}
	case subjectCastSpell:
		colors = s.event.Colors
	case subjectCard:
		if s.card != nil && s.card.Def != nil {
			colors = s.card.Def.DefaultFace().Colors
		}
	default:
		return 0
	}
	for _, c := range colors {
		seen[c] = true
	}
	return len(seen)
}

func (s *selectionSubject) hasKeyword(keyword game.Keyword) bool {
	if s.kind == subjectPermanent {
		return s.values.keywords[keyword]
	}
	if s.kind == subjectEventPermanent {
		if values, ok := s.eventPermanentValues(); ok {
			return values.keywords[keyword]
		}
		if def, ok := s.eventPermanentCardDef(); ok {
			return def.HasKeyword(keyword)
		}
	}
	if s.kind == subjectCard && s.card != nil && s.card.Def != nil {
		return s.card.Def.HasKeyword(keyword)
	}
	return false
}

func (s *selectionSubject) tapped() bool {
	if s.kind == subjectEventPermanent {
		if s.event.PermanentID != 0 {
			if permanent, ok := permanentByObjectID(s.g, s.event.PermanentID); ok {
				return permanent.Tapped
			}
			if snapshot, ok := lastKnownObject(s.g, s.event.PermanentID); ok {
				return snapshot.Tapped
			}
		}
		return false
	}
	return s.kind == subjectPermanent && s.permanent != nil && s.permanent.Tapped
}

// hasCounter reports whether the subject permanent carries at least one counter
// of the given kind. Counters live only on battlefield permanents, so a card or
// cast-spell subject never matches; an event permanent reads its live or
// last-known counters.
func (s *selectionSubject) hasCounter(kind counter.Kind) bool {
	if s.kind == subjectPermanent {
		return s.permanent != nil && s.permanent.Counters.Has(kind)
	}
	if s.kind == subjectEventPermanent && s.event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(s.g, s.event.PermanentID); ok {
			return permanent.Counters.Has(kind)
		}
		if snapshot, ok := lastKnownObject(s.g, s.event.PermanentID); ok {
			return snapshot.Counters.Has(kind)
		}
	}
	return false
}

func (s *selectionSubject) hasAnyCounter() bool {
	if s.kind == subjectPermanent {
		return s.permanent != nil && !s.permanent.Counters.IsEmpty()
	}
	if s.kind == subjectEventPermanent && s.event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(s.g, s.event.PermanentID); ok {
			return !permanent.Counters.IsEmpty()
		}
		if snapshot, ok := lastKnownObject(s.g, s.event.PermanentID); ok {
			return !snapshot.Counters.IsEmpty()
		}
	}
	return false
}

// enteredThisTurn reports whether the subject permanent entered the battlefield
// this turn. Only a live or event permanent can match; a card or cast-spell
// subject never entered the battlefield and fails closed.
func (s *selectionSubject) enteredThisTurn() bool {
	if s.kind == subjectPermanent {
		return s.permanent != nil && permanentEnteredThisTurn(s.g, s.permanent.ObjectID)
	}
	if s.kind == subjectEventPermanent && s.event.PermanentID != 0 {
		return permanentEnteredThisTurn(s.g, s.event.PermanentID)
	}
	return false
}

// modified reports whether the permanent is modified: it carries one or more
// counters, or has one or more Auras or Equipment attached to it. Only
// battlefield permanents can be modified; other subjects never match.
func (s *selectionSubject) modified() bool {
	if s.kind == subjectPermanent {
		return s.permanent != nil &&
			(!s.permanent.Counters.IsEmpty() || len(s.permanent.Attachments) > 0)
	}
	if s.kind == subjectEventPermanent && s.event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(s.g, s.event.PermanentID); ok {
			return !permanent.Counters.IsEmpty() || len(permanent.Attachments) > 0
		}
		if snapshot, ok := lastKnownObject(s.g, s.event.PermanentID); ok {
			return !snapshot.Counters.IsEmpty() || len(snapshot.Attachments) > 0
		}
	}
	return false
}

func (s *selectionSubject) combatStateMatches(filter game.CombatStateFilter) bool {
	if filter == game.CombatStateAny {
		return true
	}
	if s.kind == subjectPermanent && s.permanent != nil {
		return combatStateMatches(s.g, s.permanent, filter)
	}
	if s.kind != subjectEventPermanent || s.event.PermanentID == 0 {
		return false
	}
	if permanent, ok := permanentByObjectID(s.g, s.event.PermanentID); ok {
		return combatStateMatches(s.g, permanent, filter)
	}
	snapshot, ok := lastKnownObject(s.g, s.event.PermanentID)
	if !ok {
		return false
	}
	switch filter {
	case game.CombatStateAttacking:
		return snapshot.Attacking
	case game.CombatStateBlocking:
		return snapshot.Blocking
	case game.CombatStateAttackingOrBlocking:
		return snapshot.Attacking || snapshot.Blocking
	default:
		return false
	}
}

func (s *selectionSubject) manaValue() (int, bool) {
	if s.kind == subjectCastSpell {
		return s.event.ManaValue.Val, s.event.ManaValue.Exists
	}
	if s.kind == subjectCard && s.card != nil && s.card.Def != nil {
		return s.card.Def.ManaValue(), true
	}
	if s.kind == subjectEventPermanent {
		def, ok := s.eventPermanentCardDef()
		if !ok {
			return 0, false
		}
		return def.ManaValue(), true
	}
	if s.kind != subjectPermanent {
		return 0, false
	}
	def, ok := permanentCardDef(s.g, s.permanent)
	if !ok {
		return 0, false
	}
	return def.ManaValue(), true
}

func (s *selectionSubject) power() (int, bool) {
	if s.kind == subjectEventPermanent {
		values, ok := s.eventPermanentValues()
		return values.power, ok && values.powerOK
	}
	if s.kind != subjectPermanent || s.useBase {
		return 0, false
	}
	if s.clampPower {
		if s.values.powerOK {
			return max(0, s.values.power), true
		}
		return 0, true
	}
	return s.values.power, s.values.powerOK
}

func (s *selectionSubject) toughness() (int, bool) {
	if s.kind == subjectEventPermanent {
		values, ok := s.eventPermanentValues()
		return values.toughness, ok && values.toughnessOK
	}
	if s.kind != subjectPermanent || s.useBase {
		return 0, false
	}
	return s.values.toughness, s.values.toughnessOK
}

func (s *selectionSubject) eventPermanentValues() (permanentEffectiveValues, bool) {
	if s.event.PermanentID == 0 {
		return permanentEffectiveValues{}, false
	}
	if permanent, ok := permanentByObjectID(s.g, s.event.PermanentID); ok {
		return effectivePermanentValues(s.g, permanent), true
	}
	snapshot, ok := lastKnownObject(s.g, s.event.PermanentID)
	if !ok {
		return permanentEffectiveValues{}, false
	}
	keywords := make(map[game.Keyword]bool, len(snapshot.Keywords))
	for _, keyword := range snapshot.Keywords {
		keywords[keyword] = true
	}
	return permanentEffectiveValues{
		name:        snapshot.Name,
		colors:      snapshot.Colors,
		supertypes:  snapshot.Supertypes,
		types:       snapshot.Types,
		subtypes:    snapshot.Subtypes,
		controller:  snapshot.Controller,
		power:       snapshot.Power.Val,
		powerOK:     snapshot.Power.Exists,
		toughness:   snapshot.Toughness.Val,
		toughnessOK: snapshot.Toughness.Exists,
		keywords:    keywords,
	}, true
}

func (s *selectionSubject) eventPermanentCardDef() (*game.CardDef, bool) {
	if s.event.TokenDef != nil {
		return s.event.TokenDef.FaceDef(s.event.Face)
	}
	if s.event.CardID == 0 {
		return nil, false
	}
	card, ok := s.g.GetCardInstance(s.event.CardID)
	if !ok || card.Def == nil {
		return nil, false
	}
	return card.Def.FaceDef(s.event.Face)
}

func (s *selectionSubject) controllerMatches(relation game.ControllerRelation) bool {
	switch relation {
	case game.ControllerYou:
		return s.controller == s.viewer
	case game.ControllerOpponent, game.ControllerNotYou:
		return s.controller != s.viewer && isPlayerAlive(s.g, s.controller)
	default:
		return true
	}
}

func (s *selectionSubject) isSource() bool {
	if s.sourceObjectID == 0 {
		return false
	}
	return s.kind == subjectPermanent && s.permanent != nil && s.permanent.ObjectID == s.sourceObjectID
}

func (s *selectionSubject) isToken() bool {
	switch s.kind {
	case subjectPermanent:
		return s.permanent != nil && s.permanent.Token
	case subjectEventPermanent:
		return eventPermanentIsToken(s.g, s.event)
	default:
		return false
	}
}

// selectionPlayerRelationMatches applies the player-relation portion of a
// Selection to a player target. Player targets do not flow through
// matchSelection because they carry no characteristic data.
func selectionPlayerRelationMatches(relation game.PlayerRelation, playerID, viewer game.PlayerID) bool {
	switch relation {
	case game.PlayerYou:
		return playerID == viewer
	case game.PlayerOpponent, game.PlayerNotYou:
		return playerID != viewer
	default:
		return true
	}
}
