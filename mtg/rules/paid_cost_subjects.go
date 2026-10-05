package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (s *rulesPaymentState) PaySacrificeSubject(permanent *game.Permanent, key string) (game.PaidCostSubject, bool) {
	subject := game.PaidCostSubject{Key: key, Kind: game.PaidCostSacrifice}
	if permanent == nil {
		return subject, false
	}
	subject.Snapshot = snapshotPermanent(s.g, permanent, zone.Battlefield)
	_, subject.CharacteristicsKnown = permanentCardDef(s.g, permanent)
	value, known := paidPermanentManaValue(s.g, permanent)
	subject.Snapshot.ManaValue = optionalInt(value, known)
	if card, ok := s.g.GetCardInstance(permanent.CardInstanceID); ok {
		subject.CardZoneVersion = opt.Val(card.ZoneVersion)
	}
	return subject, sacrificePermanent(s.g, permanent)
}

func captureDiscardCostSubject(g *game.Game, playerID game.PlayerID, cardID id.ID, key string) game.PaidCostSubject {
	subject := game.PaidCostSubject{Key: key, Kind: game.PaidCostDiscard}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return subject
	}
	subject.CardZoneVersion = opt.Val(card.ZoneVersion)
	subject.Snapshot = game.ObjectSnapshot{
		CardID: cardID, Owner: card.Owner, Controller: playerID, FromZone: zone.Hand,
		Face: game.FaceFront,
	}
	if card.Def == nil {
		return subject
	}
	face := card.Def.DefaultFace()
	subject.CharacteristicsKnown = true
	subject.Snapshot.Name = face.Name
	subject.Snapshot.CopiableDef = card.Def
	subject.Snapshot.Types = slices.Clone(face.Types)
	subject.Snapshot.Supertypes = slices.Clone(face.Supertypes)
	subject.Snapshot.Subtypes = slices.Clone(face.Subtypes)
	subject.Snapshot.Colors = spellColors(card.Def)
	subject.Snapshot.ManaValue = opt.Val(card.Def.ManaValue())
	subject.Snapshot.Power = optionalInt(cardFacePT(card, func(face game.CardFace) opt.V[game.PT] { return face.Power }))
	subject.Snapshot.Toughness = optionalInt(cardFacePT(card, func(face game.CardFace) opt.V[game.PT] { return face.Toughness }))
	return subject
}

func (s *rulesPaymentState) PayDiscardSubject(playerID game.PlayerID, cardID id.ID, key string) (game.PaidCostSubject, bool) {
	subject := captureDiscardCostSubject(s.g, playerID, cardID, key)
	return subject, discardCardFromHand(s.g, playerID, cardID)
}

func (s *rulesPaymentState) PayRandomDiscardSubjects(playerID game.PlayerID, keys []string) ([]game.PaidCostSubject, bool) {
	player, ok := playerByID(s.g, playerID)
	if !ok || player.Hand.Size() < len(keys) {
		return nil, false
	}
	// Freeze the simultaneous batch before any of its moves. The existing
	// random-discard operation alone chooses and reports the actual paid cards.
	candidates := make(map[id.ID]game.PaidCostSubject, player.Hand.Size())
	for _, cardID := range player.Hand.All() {
		candidates[cardID] = captureDiscardCostSubject(s.g, playerID, cardID, "")
	}
	discarded := discardCardsAtRandomFromHand(s.g, playerID, len(keys))
	if len(discarded) != len(keys) {
		return nil, false
	}
	var subjects []game.PaidCostSubject
	for i, cardID := range discarded {
		if keys[i] == "" {
			continue
		}
		subject := candidates[cardID]
		subject.Key = keys[i]
		subjects = append(subjects, subject)
	}
	return subjects, true
}

func resolvePaidCostSubject(obj *game.StackObject, ref game.ObjectReference) (resolvedObjectReference, bool) {
	if obj == nil || ref.Kind() != game.ObjectReferencePaidCost || len(ref.Validate()) != 0 {
		return resolvedObjectReference{}, false
	}
	var found *game.PaidCostSubject
	for i := range obj.PaidCostSubjects {
		subject := &obj.PaidCostSubjects[i]
		if subject.Key != ref.CostKey() {
			continue
		}
		if found != nil || subject.Kind != ref.CostKind() {
			return resolvedObjectReference{}, false
		}
		found = subject
	}
	if found == nil || !found.CharacteristicsKnown ||
		found.Kind == game.PaidCostSacrifice && found.Snapshot.ObjectID == 0 ||
		found.Kind == game.PaidCostDiscard && (found.Snapshot.CardID == 0 || !found.CardZoneVersion.Exists) {
		return resolvedObjectReference{}, false
	}
	if ref.CostNoun() != "" && !slices.Contains(found.Snapshot.Types, ref.CostNoun()) {
		return resolvedObjectReference{}, false
	}
	return resolvedObjectReference{snapshot: found.Snapshot, frozen: true}, true
}

func paidCostConditionAvailable(obj *game.StackObject, condition *game.Condition) bool {
	if !condition.Object.Exists || condition.Object.Val.Kind() != game.ObjectReferencePaidCost {
		return true
	}
	resolved, ok := resolvePaidCostSubject(obj, condition.Object.Val)
	if !ok {
		return false
	}
	if condition.ObjectMatches.Exists {
		selection := condition.ObjectMatches.Val
		return game.PaidCostSelectionSupported(selection) &&
			paidCostNumericInformationAvailable(&resolved.snapshot, selection)
	}
	return true
}

func paidCostNumericInformationAvailable(snapshot *game.ObjectSnapshot, selection game.Selection) bool {
	for _, alternative := range selection.AnyOf {
		if !paidCostNumericInformationAvailable(snapshot, alternative) {
			return false
		}
	}
	return (!selection.Power.Exists || snapshot.Power.Exists) &&
		(!selection.Toughness.Exists || snapshot.Toughness.Exists) &&
		(!selection.ManaValue.Exists || snapshot.ManaValue.Exists)
}

func paidPermanentManaValue(g *game.Game, permanent *game.Permanent) (int, bool) {
	if permanent.Face == game.FaceBack && !permanent.FaceDown && !permanentHasCopyLayer(g, permanent) {
		physical, known := physicalPermanentDef(g, permanent)
		if !known || physical == nil {
			return 0, false
		}
		// CR 202.3b: an original transforming back face keeps its front-face
		// mana value; a copy reads its own copiable mana cost instead.
		if physical.IsTransformingDoubleFaced() {
			return physical.ManaValue(), true
		}
		if physical.Layout == game.LayoutMeld {
			return 0, false
		}
	}
	return effectivePermanentManaValue(g, permanent)
}
