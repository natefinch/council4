package rules

import (
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) applyPlayLand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return e.applyPlayLandFace(g, playerID, cardID, game.FaceFront)
}

func (e *Engine) applyPlayLandFace(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex) bool {
	return e.applyPlayLandFaceWithChoices(g, playerID, cardID, face, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyPlayLandFaceWithChoices(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !canPlayAnyLand(g, playerID) {
		return false
	}

	player := g.Players[playerID]
	card, ok := landCardInstanceFace(g, player, cardID, face)
	if !ok {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}

	if _, ok := createCardPermanentFaceWithChoices(e, g, card, playerID, zone.Hand, face, agents, log); !ok {
		return false
	}
	g.Turn.LandsPlayedThisTurn++
	return true
}

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	return e.applyCastSpellWithChoices(g, playerID, cast, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCastSpellWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if cast.Mutate {
		return e.applyMutateCastWithChoices(g, playerID, cast, agents, log)
	}
	sourceZone := normalizedCastSourceZone(cast)
	if sourceZone == zone.Battlefield {
		return e.applyPreparedCopyWithChoices(g, playerID, cast, agents, log)
	}

	if !e.canCastSpellFaceFromZoneWithKicker(g, playerID, cast.CardID, sourceZone, cast.Face, cast.Targets, cast.XValue, cast.ChosenModes, cast.KickerPaid) {
		return false
	}

	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, cast.Face)
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, spellDef, cast.ChosenModes, cast.Targets, agents, log)
	if !ok || !e.canCastSpellFaceFromZoneWithKicker(g, playerID, cast.CardID, sourceZone, cast.Face, completedTargets, cast.XValue, cast.ChosenModes, cast.KickerPaid) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, cast.ChosenModes, cast.Targets)
	if !ok {
		panic("validated spell targets could not be segmented")
	}
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, spellDef, cast.XValue, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{PlayerID: playerID, CardID: card.ID, SourceZone: sourceZone, Card: spellDef, XValue: cast.XValue, KickerPaid: cast.KickerPaid, Prefs: prefs})
	if !ok {
		return false
	}
	if !removeCastSourceCard(g, player, cast.CardID, sourceZone) {
		panic("cast spell disappeared from source zone after validation")
	}
	if sourceZone == zone.Command && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Face:                cast.Face,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		KickerPaid:          cast.KickerPaid,
		Flashback:           sourceZone == zone.Graveyard && spellDef.HasKeyword(game.Flashback),
		AdditionalCostsPaid: additionalCostsPaid,
		SourceZone:          sourceZone,
	}
	stormCopies := stormCopyCount(g, spellDef)
	pushSpellToStack(g, obj, game.Event{
		SourceID:       cast.CardID,
		StackObjectID:  obj.ID,
		Controller:     playerID,
		CardID:         cast.CardID,
		Face:           cast.Face,
		CardTypes:      cardTypes(spellDef),
		CardSupertypes: cardSupertypes(spellDef),
		CardSubtypes:   cardSubtypes(spellDef),
		Colors:         spellColors(spellDef),
		ManaValue:      opt.Val(stackManaValue(spellDef, cast.XValue)),
		KickerPaid:     cast.KickerPaid,
		FromZone:       sourceZone,
		ToZone:         zone.Stack,
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}

func (e *Engine) applyMutateCastWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	sourceZone := normalizedCastSourceZone(cast)
	if !canCastMutateSpell(g, playerID, cast.CardID, sourceZone, cast.MutateTargetID) {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	mutateCost, ok := spellDef.MutateCost()
	if !ok {
		return false
	}
	alternative := mutateAlternativeCost(mutateCost)
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, spellDef, 0, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:    playerID,
		CardID:      card.ID,
		SourceZone:  sourceZone,
		Card:        spellDef,
		Alternative: opt.Val(alternative),
		Prefs:       prefs,
	})
	if !ok {
		return false
	}
	if !removeCastSourceCard(g, player, cast.CardID, sourceZone) {
		panic("mutate spell disappeared from source zone after validation")
	}
	if sourceZone == zone.Command && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Face:                game.FaceFront,
		Controller:          playerID,
		Targets:             []game.Target{game.PermanentTarget(cast.MutateTargetID)},
		TargetCounts:        []int{1},
		Mutate:              true,
		MutateTargetID:      cast.MutateTargetID,
		AdditionalCostsPaid: additionalCostsPaid,
		SourceZone:          sourceZone,
	}
	pushSpellToStack(g, obj, game.Event{
		SourceID:       cast.CardID,
		StackObjectID:  obj.ID,
		Controller:     playerID,
		CardID:         cast.CardID,
		Face:           game.FaceFront,
		CardTypes:      cardTypes(spellDef),
		CardSupertypes: cardSupertypes(spellDef),
		CardSubtypes:   cardSubtypes(spellDef),
		Colors:         spellColors(spellDef),
		ManaValue:      opt.Val(stackManaValue(spellDef, 0)),
		FromZone:       sourceZone,
		ToZone:         zone.Stack,
	})
	return true
}

func canCastMutateSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, targetID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(player, cardID, sourceZone) {
		return false
	}
	switch sourceZone {
	case zone.Hand:
	case zone.Command:
		if player.CommanderInstanceID != cardID {
			return false
		}
	case zone.Exile:
		if !g.AdventureCards[cardID] {
			return false
		}
	default:
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	mutateCost, ok := spellDef.MutateCost()
	if !ok ||
		!spellDef.HasType(types.Creature) ||
		!isSupportedSpell(spellDef) ||
		!canCastAtCurrentTiming(g, playerID, spellDef) ||
		!mutateTargetLegal(g, playerID, card.Owner, spellDef, targetID) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:    playerID,
		CardID:      cardID,
		SourceZone:  sourceZone,
		Card:        spellDef,
		Alternative: opt.Val(mutateAlternativeCost(mutateCost)),
	})
}

func legalMutateTargets(g *game.Game, playerID, owner game.PlayerID, spellDef *game.CardDef) []*game.Permanent {
	var targets []*game.Permanent
	for _, permanent := range g.Battlefield {
		if mutateTargetLegal(g, playerID, owner, spellDef, permanent.ObjectID) {
			targets = append(targets, permanent)
		}
	}
	return targets
}

func mutateTargetLegal(g *game.Game, playerID, owner game.PlayerID, spellDef *game.CardDef, targetID id.ID) bool {
	target, ok := permanentByObjectID(g, targetID)
	return ok &&
		!target.PhasedOut &&
		target.Owner == owner &&
		permanentHasType(g, target, types.Creature) &&
		!permanentHasSubtype(g, target, types.Human) &&
		!targetProtectedFromSource(g, playerID, spellDef, 0, game.PermanentTarget(targetID))
}

func mutateAlternativeCost(manaCost cost.Mana) cost.Alternative {
	return cost.Alternative{
		Label:    "Mutate",
		ManaCost: opt.Val(append(cost.Mana(nil), manaCost...)),
	}
}

func canCastPreparedCopy(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, targets []game.Target, xValue int, chosenModes []int) bool {
	if !canAct(g, playerID) ||
		playerID != g.Turn.PriorityPlayer ||
		permanent == nil ||
		!permanent.Prepared ||
		permanent.PhasedOut ||
		effectiveController(g, permanent) != playerID ||
		xValue < 0 {
		return false
	}
	sourceID, sourceDef, ok := preparedSpellSource(g, permanent)
	if !ok {
		return false
	}
	spellDef, ok := sourceDef.FaceDef(game.FaceAlternate)
	if !ok {
		return false
	}
	if xValue != 0 && !costHasVariableMana(manaCostPtr(spellDef.ManaCost)) {
		return false
	}
	if !modesValidForSpell(spellDef, chosenModes) ||
		!isSupportedSpell(spellDef) ||
		(!spellDef.HasType(types.Instant) && !spellDef.HasType(types.Sorcery)) ||
		!targetsValidForSpell(g, playerID, spellDef, chosenModes, targets) ||
		!canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     xValue,
	})
}

func (e *Engine) applyPreparedCopyWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if cast.Face != game.FaceAlternate || cast.KickerPaid {
		return false
	}
	permanent := preparedSourcePermanent(g, cast.CardID)
	if !canCastPreparedCopy(g, playerID, permanent, cast.Targets, cast.XValue, cast.ChosenModes) {
		return false
	}
	sourceID, sourceDef, ok := preparedSpellSource(g, permanent)
	if !ok {
		return false
	}
	spellDef, ok := sourceDef.FaceDef(game.FaceAlternate)
	if !ok {
		return false
	}
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, spellDef, cast.ChosenModes, cast.Targets, agents, log)
	if !ok || !canCastPreparedCopy(g, playerID, permanent, completedTargets, cast.XValue, cast.ChosenModes) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, cast.ChosenModes, cast.Targets)
	if !ok {
		panic("validated prepared spell targets could not be segmented")
	}
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, sourceID, zone.Battlefield, spellDef, cast.XValue, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     cast.XValue,
		Prefs:      prefs,
	})
	if !ok {
		return false
	}
	permanent.Prepared = false
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            sourceID,
		Face:                game.FaceAlternate,
		SourceCardID:        permanent.CardInstanceID,
		SourceTokenDef:      permanent.TokenDef,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		Copy:                true,
		AdditionalCostsPaid: additionalCostsPaid,
		SourceZone:          zone.Battlefield,
	}
	stormCopies := stormCopyCount(g, spellDef)
	g.Stack.Push(obj)
	emitTargetEvents(g, obj)
	emitEvent(g, game.Event{
		Kind:           game.EventSpellCast,
		SourceID:       sourceID,
		StackObjectID:  obj.ID,
		Controller:     playerID,
		CardID:         permanent.CardInstanceID,
		Face:           game.FaceAlternate,
		PermanentID:    permanent.ObjectID,
		TokenDef:       permanent.TokenDef,
		CardTypes:      cardTypes(spellDef),
		CardSupertypes: cardSupertypes(spellDef),
		CardSubtypes:   cardSubtypes(spellDef),
		Colors:         spellColors(spellDef),
		ManaValue:      opt.Val(stackManaValue(spellDef, cast.XValue)),
		FromZone:       zone.Battlefield,
		ToZone:         zone.Stack,
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}

func preparedSourcePermanent(g *game.Game, sourceID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent != nil &&
			(permanent.CardInstanceID == sourceID || permanent.Token && permanent.ObjectID == sourceID) {
			return permanent
		}
	}
	return nil
}

func preparedSpellSource(g *game.Game, permanent *game.Permanent) (id.ID, *game.CardDef, bool) {
	if permanent == nil {
		return 0, nil, false
	}
	if permanent.CardInstanceID != 0 {
		card, ok := g.GetCardInstance(permanent.CardInstanceID)
		if !ok || card.Def.Layout != game.LayoutPrepare || !card.Def.Alternate.Exists {
			return 0, nil, false
		}
		return permanent.CardInstanceID, card.Def, true
	}
	if !permanent.Token || permanent.TokenDef == nil ||
		permanent.TokenDef.Layout != game.LayoutPrepare ||
		!permanent.TokenDef.Alternate.Exists {
		return 0, nil, false
	}
	return permanent.ObjectID, permanent.TokenDef, true
}

func (e *Engine) canCastSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellWithKicker(g, playerID, cardID, targets, xValue, chosenModes, false)
}

func (e *Engine) canCastSpellWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, zone.Hand, game.FaceFront, targets, xValue, chosenModes, kickerPaid)
}

func (e *Engine) canCastSpellFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, game.FaceFront, targets, xValue, chosenModes, kickerPaid)
}

func (*Engine) canCastSpellFaceFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if xValue < 0 {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(player, cardID, sourceZone) {
		return false
	}
	spellDef, ok := cardFaceDef(card, face)
	if !ok || !card.Def.CanChooseCastFace(face) {
		return false
	}
	switch sourceZone {
	case zone.Command:
		if player.CommanderInstanceID != cardID {
			return false
		}
	case zone.Hand:
	case zone.Graveyard:
		if !canCastFromZoneByRuleEffect(g, playerID, cardID, sourceZone, face) {
			return false
		}
	case zone.Exile:
		if !g.AdventureCards[cardID] {
			return false
		}
		if face != game.FaceFront {
			return false
		}
	default:
		return false
	}
	if xValue != 0 && !costHasVariableMana(manaCostPtr(spellDef.ManaCost)) {
		return false
	}
	if !modesValidForSpell(spellDef, chosenModes) || !isSupportedSpell(spellDef) || !targetsValidForSpell(g, playerID, spellDef, chosenModes, targets) {
		return false
	}
	if !canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	if kickerPaid && !spellHasKicker(spellDef) {
		return false
	}
	if !paymentOrch.canPaySpellCosts(g, payment.SpellRequest{PlayerID: playerID, CardID: card.ID, SourceZone: sourceZone, Card: spellDef, XValue: xValue, KickerPaid: kickerPaid}) {
		return false
	}
	return true
}

func legalCastFacesForZone(g *game.Game, playerID game.PlayerID, card *game.CardInstance, sourceZone zone.Type) []game.FaceIndex {
	if sourceZone == zone.Graveyard {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if canCastFromZoneByRuleEffect(g, playerID, card.ID, sourceZone, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	return card.Def.LegalCastFaces()
}

func castSourceContains(player *game.Player, cardID id.ID, sourceZone zone.Type) bool {
	switch sourceZone {
	case zone.Hand:
		return player.Hand.Contains(cardID)
	case zone.Command:
		return player.CommandZone.Contains(cardID)
	case zone.Graveyard:
		return player.Graveyard.Contains(cardID)
	case zone.Exile:
		return player.Exile.Contains(cardID)
	default:
		return false
	}
}

func castSourceZoneCards(player *game.Player, sourceZone zone.Type) []id.ID {
	switch sourceZone {
	case zone.Hand:
		return player.Hand.All()
	case zone.Command:
		return player.CommandZone.All()
	case zone.Graveyard:
		return player.Graveyard.All()
	case zone.Exile:
		return player.Exile.All()
	default:
		return nil
	}
}

func removeCastSourceCard(g *game.Game, player *game.Player, cardID id.ID, sourceZone zone.Type) bool {
	switch sourceZone {
	case zone.Hand:
		return player.Hand.Remove(cardID)
	case zone.Command:
		return player.CommandZone.Remove(cardID)
	case zone.Graveyard:
		return player.Graveyard.Remove(cardID)
	case zone.Exile:
		return player.Exile.Remove(cardID)
	default:
		return false
	}
}
