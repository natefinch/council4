package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
)

// actionBuilderType is the single point in the rules package for constructing
// action.Action values. Every method validates the constructed action via
// Action.Validate() and panics on a programming error (invalid inputs that
// should have been checked by the caller before building the action).
//
// All legal action generation and action application code in the rules package
// must use actionBuild instead of calling action package constructors directly.
type actionBuilderType struct{}

// actionBuild is the package-level singleton for constructing actions.
var actionBuild actionBuilderType

func (actionBuilderType) mustBuild(act action.Action) action.Action {
	if err := act.Validate(); err != nil {
		panic(fmt.Sprintf("rules: actionBuilder produced invalid action: %v", err))
	}
	return act
}

func (b actionBuilderType) pass() action.Action {
	return b.mustBuild(action.Pass())
}

func (b actionBuilderType) playLand(cardID id.ID, face game.FaceIndex) action.Action {
	return b.playLandFromZone(cardID, zone.Hand, face)
}

func (b actionBuilderType) playLandFromZone(cardID id.ID, sourceZone zone.Type, face game.FaceIndex) action.Action {
	return b.mustBuild(action.PlayLandFaceFromZone(cardID, sourceZone, face))
}

// castSpell builds a normal (non-kicked) CastSpell action for the given card,
// source zone, face, targets, X value, and chosen modes.
func (b actionBuilderType) castSpell(cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, modes []int) action.Action {
	return b.mustBuild(action.CastSpellFaceFromZone(cardID, sourceZone, face, targets, xValue, modes))
}

// castKickedSpell builds a kicked CastSpell action.
func (b actionBuilderType) castKickedSpell(cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, modes []int) action.Action {
	return b.mustBuild(action.CastKickedSpellFaceFromZone(cardID, sourceZone, face, targets, xValue, modes))
}

// castMultikickedSpell builds a CastSpell action whose Multikicker cost is paid
// kickerCount times.
func (b actionBuilderType) castMultikickedSpell(cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, modes []int, kickerCount int) action.Action {
	return b.mustBuild(action.CastMultikickedSpellFaceFromZone(cardID, sourceZone, face, targets, xValue, modes, kickerCount))
}

func (b actionBuilderType) castOverloadedSpell(cardID id.ID, sourceZone zone.Type, face game.FaceIndex, xValue int, modes []int, kickerPaid bool) action.Action {
	return b.mustBuild(action.CastOverloadedSpellFaceFromZoneWithOptions(cardID, sourceZone, face, xValue, modes, kickerPaid))
}

func (b actionBuilderType) castMutateSpell(cardID id.ID, sourceZone zone.Type, targetID id.ID) action.Action {
	return b.mustBuild(action.CastMutateSpellFromZone(cardID, sourceZone, targetID))
}

func (b actionBuilderType) activateAbility(sourceID id.ID, abilityIndex int, targets []game.Target, xValue int) action.Action {
	return b.mustBuild(action.ActivateAbility(sourceID, abilityIndex, targets, xValue))
}

func (b actionBuilderType) activateAbilityWithModes(sourceID id.ID, abilityIndex int, targets []game.Target, targetCounts []int, xValue int, modes []int) action.Action {
	return b.mustBuild(action.ActivateAbilityWithModesAndTargetCounts(sourceID, abilityIndex, targets, targetCounts, xValue, modes))
}

func (b actionBuilderType) suspendCard(cardID id.ID) action.Action {
	return b.mustBuild(action.SuspendCard(cardID))
}

func (b actionBuilderType) plotCard(cardID id.ID) action.Action {
	return b.mustBuild(action.PlotCard(cardID))
}

func (b actionBuilderType) castFaceDown(cardID id.ID, face game.FaceIndex, kind game.FaceDownKind) action.Action {
	return b.mustBuild(action.CastFaceDown(cardID, face, kind))
}

func (b actionBuilderType) turnFaceUp(permanentID id.ID) action.Action {
	return b.mustBuild(action.TurnFaceUp(permanentID))
}

func (b actionBuilderType) declareAttackers(attackers []game.AttackDeclaration) action.Action {
	return b.mustBuild(action.DeclareAttackers(attackers))
}

func (b actionBuilderType) declareBlockers(blockers []game.BlockDeclaration) action.Action {
	return b.mustBuild(action.DeclareBlockers(blockers))
}
