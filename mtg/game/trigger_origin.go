package game

import "github.com/natefinch/council4/mtg/game/id"

// TriggerOrigin identifies a declared runtime occurrence, never an effective
// ability-list slot. It does not certify same-object grant loss/reacquisition.
type TriggerOrigin struct {
	RecipientID id.ID
	SourceID    id.ID
	OriginID    id.ID
	Face        FaceIndex
	StaticIndex int
	EffectIndex int
	BodyIndex   int
	RuntimeID   id.ID
	Granted     bool
}

// SetTriggerOrigin captures the declared incarnation beside the event-time body.
func (e *EventTriggeredAbility) SetTriggerOrigin(origin TriggerOrigin) {
	e.triggerOrigin = origin
}

// TriggerOrigin returns the immutable captured declaration identity.
func (e *EventTriggeredAbility) TriggerOrigin() TriggerOrigin { return e.triggerOrigin }

// SetTriggerOrigin transfers captured provenance into a queued stack object.
func (o *StackObject) SetTriggerOrigin(origin TriggerOrigin) { o.triggerOrigin = origin }

// TriggerOrigin returns the queued declaration identity, including on copies.
func (o *StackObject) TriggerOrigin() TriggerOrigin { return o.triggerOrigin }

// ConsumeTriggerUse counts only committed ordinary triggers. Turn advancement
// resets the ledger; copies of queued abilities do not call this method.
func (g *Game) ConsumeTriggerUse(origin TriggerOrigin, maximum int) bool {
	if maximum <= 0 || origin.RecipientID == 0 || origin.OriginID == 0 ||
		origin.Granted && origin.SourceID == 0 && origin.RuntimeID == 0 {
		panic("game: limited trigger lacks a certified declared origin")
	}
	if g.triggerOriginTurn != g.Turn.TurnNumber {
		g.triggerOriginUses = nil
		g.triggerOriginTurn = g.Turn.TurnNumber
	}
	if g.triggerOriginUses == nil {
		g.triggerOriginUses = make(map[TriggerOrigin]int)
	}
	if g.triggerOriginUses[origin] >= maximum {
		return false
	}
	g.triggerOriginUses[origin]++
	return true
}
