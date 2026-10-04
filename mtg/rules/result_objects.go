package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (r *effectResolver) publishInstructionResult(instruction *game.Instruction, result effectResolved) {
	if instruction.PublishResult == "" {
		return
	}
	recordResultKey(r.obj, instruction.PublishResult, result)
	if r.obj == nil || len(r.resultObjects) == 0 {
		return
	}
	if r.obj.ResolutionResultObjects == nil {
		r.obj.ResolutionResultObjects = make(map[string][]game.ObjectSnapshot)
	}
	r.obj.ResolutionResultObjects[string(instruction.PublishResult)] = r.resultObjects
}

func (r *effectResolver) rememberResultObject(ref game.LinkedObjectRef) {
	if r.currentInstruction == nil || r.currentInstruction.PublishResult == "" {
		return
	}
	if ref.ObjectID != 0 {
		if snapshot, ok := lastKnownObject(r.game, ref.ObjectID); ok {
			r.resultObjects = append(r.resultObjects, resultCharacteristicSnapshot(snapshot))
		}
		return
	}
	card, ok := r.game.GetCardInstance(ref.CardID)
	if !ok || card.Def == nil {
		return
	}
	face := card.Def.DefaultFace()
	r.resultObjects = append(r.resultObjects, resultCharacteristicSnapshot(game.ObjectSnapshot{
		CardID: card.ID, Name: face.Name, Owner: card.Owner, Controller: card.Owner,
		ZoneCards: []game.ZoneCardSnapshot{{CardID: card.ID, ZoneVersion: card.ZoneVersion}},
		Types:     face.Types, Supertypes: face.Supertypes, Subtypes: face.Subtypes, Colors: face.Colors,
	}))
}

func resultCharacteristicSnapshot(snapshot game.ObjectSnapshot) game.ObjectSnapshot {
	return game.ObjectSnapshot{
		ObjectID: snapshot.ObjectID, CardID: snapshot.CardID,
		ZoneCards: slices.Clone(snapshot.ZoneCards), Name: snapshot.Name,
		Owner: snapshot.Owner, Controller: snapshot.Controller, FromZone: snapshot.FromZone,
		Types: slices.Clone(snapshot.Types), Supertypes: slices.Clone(snapshot.Supertypes),
		Subtypes: slices.Clone(snapshot.Subtypes), Colors: slices.Clone(snapshot.Colors),
	}
}

func (r *effectResolver) rememberResultCard(cardID id.ID) {
	r.rememberResultObject(game.LinkedObjectRef{CardID: cardID})
}

func (r *effectResolver) rememberMovedResultCard(cardID id.ID, destination zone.Type) {
	if actual, ok := cardZone(r.game, cardID); ok && actual == destination {
		r.rememberResultCard(cardID)
	}
}

func (r *effectResolver) rememberDiscardResultEvents(start int) {
	if r.currentInstruction == nil || r.currentInstruction.PublishResult == "" {
		return
	}
	for _, event := range r.game.Events[start:] {
		if event.Kind == game.EventCardDiscarded {
			r.rememberResultCard(event.CardID)
		}
	}
}

func (r *effectResolver) rememberDepartedResultPermanents(permanents []*game.Permanent) {
	if r.currentInstruction == nil || r.currentInstruction.PublishResult == "" {
		return
	}
	for _, permanent := range permanents {
		if _, live := permanentByObjectID(r.game, permanent.ObjectID); !live {
			r.rememberResultObject(permanentObjectBindingRef(permanent))
		}
	}
}

func (r *effectResolver) moveResultPermanents(permanents []*game.Permanent, destination zone.Type) bool {
	moved := false
	for _, result := range movePermanentsToZoneSimultaneouslyWithResults(r.game, permanents, destination) {
		if !result.moved {
			continue
		}
		moved = true
		if result.destination == destination {
			r.rememberResultObject(permanentObjectBindingRef(result.permanent))
		}
	}
	return moved
}

func resultObjectsMatchSelection(g *game.Game, obj *game.StackObject, snapshots []game.ObjectSnapshot, selection game.Selection) bool {
	return resultObjectsMatchFilter(g, obj, snapshots, selection, false)
}

func resultObjectsMatchFilter(g *game.Game, obj *game.StackObject, snapshots []game.ObjectSnapshot, selection game.Selection, cardOnly bool) bool {
	for i := range snapshots {
		snapshot := &snapshots[i]
		if cardOnly && (snapshot.ObjectID != 0 || snapshot.CardID == 0) {
			continue
		}
		subject := selectionSubject{
			kind: subjectEventPermanent, g: g, snapshot: snapshot,
			controller: snapshot.Controller, viewer: obj.Controller,
			sourceObjectID: obj.SourceID, resolutionChoices: obj.ResolutionChoices, obj: obj,
		}
		if matchSelection(&subject, &selection) {
			return true
		}
	}
	return false
}
