package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// DelayedTriggerView is a read-only view of one pending delayed triggered
// ability (CR 603.7). Delayed triggers are created by the resolution of public
// spells and abilities, so their existence, controller, source name, text,
// timing and bound objects are public. A source that is now a face-down
// permanent the viewer may not look at reports no name; the trigger's text,
// captured from the public creating spell or ability, remains visible.
type DelayedTriggerView struct {
	ID         id.ID
	Controller game.PlayerID
	// SourceName names the delayed trigger's source (CR 603.7d-g).
	SourceName string
	// SourceObjectID is the source permanent while it remains on the
	// battlefield, else zero.
	SourceObjectID id.ID
	// Text is the delayed ability's public rules text, or empty when its
	// definition records none.
	Text string
	// Timing is the fixed step or phase the trigger waits for; it is zero for an
	// event-based delayed trigger, which instead has EventBased set and is
	// bounded by Window. OneShot reports that an event-based trigger fires once.
	Timing      game.DelayedTriggerTiming
	EventBased  bool
	OneShot     bool
	Window      game.DelayedTriggerWindow
	CreatedTurn int
	// BoundObjectIDs are the permanents the delayed trigger captured when it was
	// created (CR 603.7c): the objects it will act on or whose events it waits
	// for.
	BoundObjectIDs []id.ID
	// BoundCardID is the captured card while it remains in the same zone and is
	// visible to the viewer; otherwise zero.
	BoundCardID id.ID
	// BoundPlayer is the player whose events an event-based trigger waits for.
	BoundPlayer opt.V[game.PlayerID]
}

// DelayedTriggers returns every pending delayed triggered ability, in creation
// order.
func (o PlayerObservation) DelayedTriggers() []DelayedTriggerView {
	views := make([]DelayedTriggerView, 0, len(o.g.DelayedTriggers))
	for i := range o.g.DelayedTriggers {
		views = append(views, o.delayedTriggerView(&o.g.DelayedTriggers[i]))
	}
	return views
}

func (o PlayerObservation) permanentDelayedTriggers(objectID id.ID) []DelayedTriggerView {
	var views []DelayedTriggerView
	for i := range o.g.DelayedTriggers {
		view := o.delayedTriggerView(&o.g.DelayedTriggers[i])
		if view.SourceObjectID == objectID || slices.Contains(view.BoundObjectIDs, objectID) {
			views = append(views, view)
		}
	}
	return views
}

func (o PlayerObservation) delayedTriggerView(trigger *game.DelayedTrigger) DelayedTriggerView {
	view := DelayedTriggerView{
		ID:             trigger.ID,
		Controller:     trigger.Controller,
		SourceObjectID: o.battlefieldObjectID(trigger.SourceObjectID),
		Timing:         trigger.Timing,
		EventBased:     trigger.EventPattern.Exists,
		OneShot:        trigger.OneShot,
		Window:         trigger.Window,
		CreatedTurn:    trigger.CreatedTurn,
		BoundObjectIDs: boundObjectIDs(
			trigger.CapturedObjectID,
			trigger.CapturedObjectIDs,
			trigger.BoundDamageSourceObjectID,
			trigger.BoundAttackerObjectID,
			trigger.BoundDyingObjectID,
		),
		BoundCardID: o.visibleBoundCardID(trigger.CapturedCardID, trigger.CapturedCardZoneVersion),
		BoundPlayer: trigger.BoundEventPlayer,
		// The text was captured from the public spell or ability that created
		// the trigger, so it stays public even when the source is now face down.
		Text: trigger.Ability.Text,
	}
	if o.hiddenFaceDownPermanent(trigger.SourceObjectID) {
		return view
	}
	switch {
	case trigger.SourceTokenDef != nil:
		view.SourceName = trigger.SourceTokenDef.Name
	default:
		if card, ok := o.g.GetCardInstance(trigger.SourceID); ok {
			view.SourceName = card.Def.Name
		}
	}
	return view
}

// fillAbilityStackObjectView adds an ability's public text, source permanent
// and captured objects to its stack view.
func (o PlayerObservation) fillAbilityStackObjectView(view *StackObjectView, obj *game.StackObject) {
	view.SourceObjectID = o.battlefieldObjectID(obj.SourceID)
	view.BoundObjectIDs = boundObjectIDs(obj.CapturedObjectID, obj.CapturedObjectIDs, obj.BoundEventObjectIDs...)
	view.BoundCardID = o.visibleBoundCardID(obj.CapturedCardID, obj.CapturedCardZoneVersion)
	view.Text = o.stackAbilityText(obj)
}

// stackAbilityText returns the text of the ability a stack object represents.
// Inline abilities (delayed, granted, keyword and activated abilities) carry
// their own text. A face-down permanent has no printed abilities (CR 708.2), so
// for a hidden face-down source only inline text is reported, never text read
// from the hidden card's definition.
func (o PlayerObservation) stackAbilityText(obj *game.StackObject) string {
	inline := inlineAbilityBody(obj)
	if text := game.BodyText(inline); text != "" {
		return text
	}
	if o.hiddenFaceDownPermanent(obj.SourceID) {
		return ""
	}
	return printedAbilityText(o.g, obj, inline)
}

// inlineAbilityBody returns the ability a stack object carries inline, or nil.
func inlineAbilityBody(obj *game.StackObject) game.Ability {
	switch {
	case obj.InlineTrigger != nil:
		return obj.InlineTrigger
	case obj.InlineActivated != nil:
		return obj.InlineActivated
	case obj.InlineLoyalty != nil:
		return obj.InlineLoyalty
	}
	return nil
}

// printedAbilityText returns the rules text of the printed ability a stack
// object represents. Definitions that omit per-ability text fall back to the
// ability's own paragraph of the source face's Oracle text
// (printedAbilityParagraphs), never to the face's other abilities; an ability
// whose paragraph cannot be identified has no text. An untexted inline ability
// gets that fallback only when it is one of the source face's own printed
// abilities; a granted or keyword ability without text has none, since the
// source's Oracle text does not contain it. A copied ability already carries its
// original card's paragraph (fillCopiedAbilityText). Callers enforce face-down
// secrecy.
func printedAbilityText(g *game.Game, obj *game.StackObject, inline game.Ability) string {
	def, ok := abilitySourceFaceView(g, obj)
	if !ok || def == nil {
		return ""
	}
	index := obj.AbilityIndex
	if inline != nil {
		if index, ok = faceAbilityIndex(def, inline); !ok {
			return ""
		}
	} else if text := strings.TrimSpace(game.BodyText(def.BodyAt(index))); text != "" {
		return text
	}
	return printedAbilityParagraph(&def.CardFace, def.OracleText, index)
}

// abilitySourceFaceView returns the read-only printed face an ability on the
// stack came from. It shares the face's ability bodies, so an inline body can
// be recognized as one of them by identity.
func abilitySourceFaceView(g *game.Game, obj *game.StackObject) (*game.CardDef, bool) {
	switch {
	case obj.SourceCardID != 0:
		card, ok := g.GetCardInstance(obj.SourceCardID)
		if !ok {
			return nil, false
		}
		return cardFaceDef(card, obj.Face)
	case obj.SourceTokenDef != nil:
		return obj.SourceTokenDef.FaceDefView(obj.Face)
	}
	if card, ok := g.GetCardInstance(obj.SourceID); ok {
		return cardFaceDef(card, obj.Face)
	}
	if permanent, ok := permanentByObjectID(g, obj.SourceID); ok {
		if physical, physicalOK := physicalPermanentDef(g, permanent); physicalOK {
			return physical.FaceDefView(obj.Face)
		}
	}
	return nil, false
}

// faceAbilityIndex returns the canonical index of body among the face's own
// ability bodies, recognized by identity.
func faceAbilityIndex(def *game.CardDef, body game.Ability) (int, bool) {
	for i := range def.AbilityCount() {
		if def.BodyAt(i) == body {
			return i, true
		}
	}
	return 0, false
}

// oracleRulesText returns Oracle text as its rules-text lines, without the
// source indentation of multi-line card definitions.
func oracleRulesText(text string) string {
	var lines []string
	for line := range strings.Lines(text) {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// hiddenFaceDownPermanent reports whether objectID is a face-down battlefield
// permanent whose identity the viewer may not see (CR 708.5).
func (o PlayerObservation) hiddenFaceDownPermanent(objectID id.ID) bool {
	permanent, ok := permanentByObjectID(o.g, objectID)
	return ok && permanent.FaceDown && !o.g.CanSeePlayerInformation(o.viewer, effectiveController(o.g, permanent))
}

func (o PlayerObservation) battlefieldObjectID(objectID id.ID) id.ID {
	if _, ok := permanentByObjectID(o.g, objectID); ok {
		return objectID
	}
	return 0
}

// visibleBoundCardID returns a captured card's ID while the card remains in the
// zone incarnation it was captured in (CR 603.7c) and that position is visible
// to the viewer: a public zone, a permanent, or a hand the viewer may see. A
// library position or another player's hand is never exposed.
func (o PlayerObservation) visibleBoundCardID(cardID id.ID, zoneVersion uint64) id.ID {
	if cardID == 0 {
		return 0
	}
	card, ok := o.g.GetCardInstance(cardID)
	if !ok || card.ZoneVersion != zoneVersion {
		return 0
	}
	for _, player := range o.g.Players {
		switch {
		case player.Graveyard.Contains(cardID), player.Exile.Contains(cardID), player.CommandZone.Contains(cardID):
			return cardID
		case player.Hand.Contains(cardID):
			if o.g.CanSeePlayerInformation(o.viewer, player.ID) {
				return cardID
			}
			return 0
		case player.Library.Contains(cardID):
			return 0
		}
	}
	for _, permanent := range o.g.Battlefield {
		if permanent != nil && permanent.CardInstanceID == cardID {
			return cardID
		}
	}
	return 0
}

func boundObjectIDs(single id.ID, group []id.ID, more ...id.ID) []id.ID {
	var ids []id.ID
	add := func(objectID id.ID) {
		if objectID != 0 && !slices.Contains(ids, objectID) {
			ids = append(ids, objectID)
		}
	}
	add(single)
	for _, objectID := range group {
		add(objectID)
	}
	for _, objectID := range more {
		add(objectID)
	}
	return ids
}
