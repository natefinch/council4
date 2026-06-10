package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
}

// matchSelection reports whether the subject satisfies every active predicate in
// sel. It is the single implementation of Selection field semantics; callers
// supply context through the subject.
func matchSelection(s *selectionSubject, sel *game.Selection) bool {
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
	if len(sel.SubtypesAny) > 0 && !s.hasAnySubtype(sel.SubtypesAny) {
		return false
	}
	if len(sel.ColorsAny) > 0 && !s.hasAnyColor(sel.ColorsAny) {
		return false
	}
	if slices.ContainsFunc(sel.ExcludedColors, s.hasColor) {
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
	return false
}

func (s *selectionSubject) hasAnySubtype(subtypes []types.Sub) bool {
	if s.kind != subjectPermanent {
		return false
	}
	for _, subtype := range subtypes {
		if slices.Contains(s.values.subtypes, subtype) {
			return true
		}
	}
	return false
}

func (s *selectionSubject) hasColor(c color.Color) bool {
	if s.kind == subjectPermanent {
		return slices.Contains(s.values.colors, c)
	}
	if s.kind == subjectCastSpell {
		return slices.Contains(s.event.Colors, c)
	}
	return false
}

func (s *selectionSubject) hasAnyColor(colors []color.Color) bool {
	return slices.ContainsFunc(colors, s.hasColor)
}

func (s *selectionSubject) hasKeyword(keyword game.Keyword) bool {
	if s.kind == subjectPermanent {
		return s.values.keywords[keyword]
	}
	return false
}

func (s *selectionSubject) tapped() bool {
	return s.kind == subjectPermanent && s.permanent != nil && s.permanent.Tapped
}

func (s *selectionSubject) combatStateMatches(filter game.CombatStateFilter) bool {
	if filter == game.CombatStateAny {
		return true
	}
	if s.kind != subjectPermanent || s.permanent == nil {
		return false
	}
	return combatStateMatches(s.g, s.permanent, filter)
}

func (s *selectionSubject) manaValue() (int, bool) {
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
	if s.kind != subjectPermanent || s.useBase {
		return 0, false
	}
	return s.values.toughness, s.values.toughnessOK
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
