package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (r *effectResolver) clearPermanentResultPublication(instruction *game.Instruction) {
	if instruction.Primitive == nil {
		return
	}
	kind := instruction.Primitive.Kind()
	if (kind == game.PrimitiveMoveTopOfLibrary || kind == game.PrimitiveMovePermanent || kind == game.PrimitiveMoveCard) &&
		!instruction.ClearLinkedBeforeGate {
		return
	}
	if kind != game.PrimitivePutOnBattlefield && kind != game.PrimitiveCreateToken &&
		kind != game.PrimitiveMoveTopOfLibrary && kind != game.PrimitiveMovePermanent && kind != game.PrimitiveMoveCard {
		return
	}
	if key := game.PublishedLinkedKey(instruction.Primitive); key != "" {
		clearLinkedObjects(r.game, linkedObjectSourceKey(r.game, r.obj, string(key)))
	}
}

func (r *effectResolver) publishEnteredPermanents(key game.LinkedKey, permanents ...*game.Permanent) {
	if key == "" {
		return
	}
	for _, permanent := range permanents {
		ref := permanentObjectBindingRef(permanent)
		if card, ok := r.game.GetCardInstance(ref.CardID); ok {
			ref.CardZoneVersion = card.ZoneVersion
		}
		rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, string(key)), ref)
	}
}

func (r *effectResolver) returnLinkedNonBattlefieldPermanents(
	linkID string,
	returnZones []zone.Type,
	controllerOverride opt.V[game.PlayerID],
	options permanentCreationOptions,
) []*game.Permanent {
	key := linkedObjectSourceKey(r.game, r.obj, linkID)
	var returned []*game.Permanent
	for _, ref := range linkedObjects(r.game, key) {
		snapshot, ok := lastKnownObject(r.game, ref.ObjectID)
		if !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card, ok := linkedCardInstance(r.game, ref)
		if !ok {
			continue
		}
		current := slices.ContainsFunc(snapshot.ZoneCards, func(c game.ZoneCardSnapshot) bool {
			return c.CardID == ref.CardID && c.ZoneVersion == card.ZoneVersion
		})
		if !current {
			continue
		}
		fromZone, ok := cardZone(r.game, ref.CardID)
		if !ok || !slices.Contains(returnZones, fromZone) {
			continue
		}
		controller := card.Owner
		if controllerOverride.Exists {
			controller = controllerOverride.Val
		}
		permanent, entered := r.putResolvedCardOnBattlefieldValue(card, fromZone, controller, nil, options)
		if entered {
			returned = append(returned, permanent)
		}
	}
	clearLinkedObjects(r.game, key)
	return returned
}
