package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func optionalInt(value int, ok bool) opt.V[int] {
	if !ok {
		return opt.V[int]{}
	}
	return opt.Val(value)
}

func snapshotPermanent(g *game.Game, permanent *game.Permanent, zoneType zone.Type) game.ObjectSnapshot {
	values := effectivePermanentValues(g, permanent)
	snapshot := game.ObjectSnapshot{
		ObjectID:       permanent.ObjectID,
		CardID:         permanent.CardInstanceID,
		TokenName:      permanentTokenName(permanent),
		TokenDef:       permanent.TokenDef,
		Face:           permanent.Face,
		Name:           values.name,
		Owner:          permanent.Owner,
		Controller:     values.controller,
		FromZone:       zoneType,
		Colors:         append([]color.Color(nil), values.colors...),
		Supertypes:     append([]types.Super(nil), values.supertypes...),
		Types:          append([]types.Card(nil), values.types...),
		Subtypes:       append([]types.Sub(nil), values.subtypes...),
		Power:          optionalInt(values.power, values.powerOK),
		Toughness:      optionalInt(values.toughness, values.toughnessOK),
		Keywords:       effectiveKeywords(values),
		MarkedDamage:   permanent.MarkedDamage,
		Attachments:    append([]id.ID(nil), permanent.Attachments...),
		AttachedTo:     permanent.AttachedTo,
		ZoneOrderIndex: -1,
	}
	snapshot.Counters = cloneCounters(permanent.Counters)
	return snapshot
}

func effectiveKeywords(values permanentEffectiveValues) []game.Keyword {
	keywords := make([]game.Keyword, 0, len(values.keywords))
	for keyword, present := range values.keywords {
		if present {
			keywords = append(keywords, keyword)
		}
	}
	slices.Sort(keywords)
	return keywords
}

func cloneCounters(counters counter.Set) counter.Set {
	var cloned counter.Set
	for kind, amount := range counters.All() {
		cloned.Add(kind, amount)
	}
	return cloned
}

func rememberLastKnown(g *game.Game, snapshot *game.ObjectSnapshot) {
	if snapshot.ObjectID == 0 {
		return
	}
	if g.LastKnownInformation == nil {
		g.LastKnownInformation = make(map[id.ID]game.ObjectSnapshot)
	}
	g.LastKnownInformation[snapshot.ObjectID] = *snapshot
}

func lastKnownObject(g *game.Game, objectID id.ID) (game.ObjectSnapshot, bool) {
	if objectID == 0 {
		return game.ObjectSnapshot{}, false
	}
	snapshot, ok := g.LastKnownInformation[objectID]
	return snapshot, ok
}

func linkedObjectSourceKey(g *game.Game, obj *game.StackObject, linkID string) game.LinkedObjectKey {
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	if sourceID == 0 {
		sourceID = sourceObjectID
	}
	return game.LinkedObjectKey{SourceID: sourceID, LinkID: linkID}
}

func rememberLinkedObject(g *game.Game, key game.LinkedObjectKey, ref game.LinkedObjectRef) {
	if key.SourceID == 0 || key.LinkID == "" || (ref.ObjectID == 0 && ref.CardID == 0) {
		return
	}
	if g.LinkedObjects == nil {
		g.LinkedObjects = make(map[game.LinkedObjectKey][]game.LinkedObjectRef)
	}
	g.LinkedObjects[key] = append(g.LinkedObjects[key], ref)
}

func linkedObjects(g *game.Game, key game.LinkedObjectKey) []game.LinkedObjectRef {
	if key.SourceID == 0 || key.LinkID == "" {
		return nil
	}
	return append([]game.LinkedObjectRef(nil), g.LinkedObjects[key]...)
}

func clearLinkedObjects(g *game.Game, key game.LinkedObjectKey) {
	if key.SourceID == 0 || key.LinkID == "" {
		return
	}
	delete(g.LinkedObjects, key)
}
