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
	return e.applyPlayLandFaceFromZoneWithChoices(g, playerID, cardID, zone.Hand, face, agents, log)
}

func (e *Engine) applyPlayLandFaceFromZoneWithChoices(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !canPlayAnyLand(g, playerID) {
		return false
	}

	player := g.Players[playerID]
	sourcePlayer := castSourcePlayer(g, player, cardID, sourceZone)
	card, ok := landCardInstanceFaceFromZone(g, sourcePlayer, cardID, sourceZone, face)
	if !ok {
		return false
	}
	var landPermission game.RuleEffect
	if sourceZone != zone.Hand {
		permission, ok := matchingPlayLandFromZoneEffect(g, playerID, cardID, sourceZone)
		if !ok {
			return false
		}
		landPermission = permission
	}
	source, ok := playerCardsInZone(sourcePlayer, sourceZone)
	if !ok || !source.Remove(cardID) {
		return false
	}

	if _, ok := createCardPermanentFaceWithChoices(e, g, card, playerID, sourceZone, face, agents, log); !ok {
		return false
	}
	g.Turn.LandsPlayedThisTurn++
	if sourceZone != zone.Hand {
		recordExilePlayPermissionUse(g, landPermission)
	}
	emitCardPlayedFromExileEvent(g, playerID, cardID, sourceZone)
	emitLandPlayedEvent(g, playerID, cardID)
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

	branch := castBranchForCast(cast)
	if !e.canCastSpellFaceFromZoneWithOptions(g, playerID, cast.CardID, sourceZone, cast.Face, cast.Targets, cast.XValue, cast.ChosenModes, effectiveKickerCount(cast.KickerPaid, cast.KickerCount), cast.Overloaded, cast.GiftPromised, cast.Bargained, cast.Bestowed, cast.Offspring) {
		return false
	}

	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, cast.Face)
	announcementDef := spellDef
	if cast.Overloaded {
		announcementDef = overloadSpellDef(spellDef)
	}
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, announcementDef, cast.ChosenModes, cast.Targets, agents, log, branch)
	if !ok || !e.canCastSpellFaceFromZoneWithOptions(g, playerID, cast.CardID, sourceZone, cast.Face, completedTargets, cast.XValue, cast.ChosenModes, effectiveKickerCount(cast.KickerPaid, cast.KickerCount), cast.Overloaded, cast.GiftPromised, cast.Bargained, cast.Bestowed, cast.Offspring) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, announcementDef, cast.ChosenModes, cast.Targets, branch)
	if !ok {
		panic("validated spell targets could not be segmented")
	}
	// Splice onto Arcane (CR 702.47): as an Arcane spell is cast from hand for its
	// normal cost, the caster may reveal splice cards in hand, pay their mana
	// splice costs as additional costs, and append their spell effects to this
	// spell. Overloaded casts and non-Arcane spells are never spliceable. When no
	// card is spliced these are all nil, leaving the cast unchanged.
	var splice spliceCastResult
	if sourceZone == zone.Hand && !cast.Overloaded && spellIsArcane(spellDef) {
		splicePermissions := castPermissionsForZone(g, playerID, card.ID, sourceZone, cast.Face)
		splice = e.chooseSpliceOntoArcane(g, playerID, card.ID, spellDef, cast, splicePermissions, agents, log)
	}
	// payLifeFromTop reads the card's position in its source zone (the top of
	// the library), so it must be determined before the card is moved to the
	// stack below.
	payLifeFromTop := sourceZone == zone.Library && castFromZoneRequiresPayLife(g, playerID, card.ID, sourceZone, cast.Face)
	// plotted marks a plotted card being cast from exile (CR 718): it is cast
	// without paying its mana cost. It must be read before the card leaves exile.
	plotted := sourceZone == zone.Exile && cast.Face == game.FaceFront && cardIsPlottedInExile(g, card.ID)
	// foretold marks a foretold card being cast from exile (CR 702.144): it is
	// cast for its foretell cost rather than its mana cost. It must be read before
	// the card leaves exile.
	foretold := sourceZone == zone.Exile && cast.Face == game.FaceFront && cardIsForetoldInExile(g, card.ID)
	// freeLinkedExile marks a card cast for free from the pool of cards exiled
	// under this source ("cast a spell from among cards exiled with this
	// enchantment without paying its mana cost.", Court of Locthwain). The
	// permission is a one-shot, so it is read before the card leaves exile and
	// remembered so it can be consumed once the spell is cast.
	var freeLinkedExilePermissionID id.ID
	freeLinkedExile := false
	if !plotted && !foretold && sourceZone == zone.Exile && cast.Face == game.FaceFront {
		if permission, permOK := castLinkedExileForFreePermission(g, playerID, card.ID); permOK {
			freeLinkedExile = true
			freeLinkedExilePermissionID = permission.ID
		}
	}
	// exilePlayPermission captures the play/cast-from-exile permission authorizing
	// this cast, if any, so a once-per-turn permission's shared per-turn use can be
	// recorded once the spell is successfully put on the stack ("Once each turn,
	// you may play a card from exile ...", Evelyn, the Covetous). It is read before
	// the card leaves exile because the permission's exile-counter filter no longer
	// matches once the card moves.
	var exilePlayPermission game.RuleEffect
	if sourceZone == zone.Exile {
		if permission, permOK := matchingCastSpellsFromZoneEffect(g, playerID, card.ID, sourceZone, cast.Face); permOK {
			exilePlayPermission = permission
		}
	}
	// spendAnyMana marks a card cast from exile under a play permission that lets
	// mana of any type pay its cost ("mana of any type can be spent to cast it.",
	// Court of Locthwain). It replaces the mana cost with an all-generic cost of
	// the same size. It never combines with the free-cast paths above.
	spendAnyMana := !plotted && !foretold && !freeLinkedExile && sourceZone == zone.Exile &&
		castFromZoneAllowsAnyMana(g, playerID, card.ID, sourceZone, cast.Face)

	// freePlayFromZone marks a card cast from exile under a per-card
	// RuleEffectPlayFromZone that lets the controller play it without paying its
	// mana cost ("You may play it this turn without paying its mana cost.", Dauthi
	// Voidwalker). Like the other free-cast paths it replaces the mana cost with an
	// empty payment; it never combines with the paths above.
	freePlayFromZone := !plotted && !foretold && !freeLinkedExile && !spendAnyMana &&
		sourceZone == zone.Exile &&
		castFromZoneWithoutPayingManaCost(g, playerID, card.ID, sourceZone, cast.Face)

	// CR 601.2a: proposing a cast first moves the card from its source zone to
	// the stack as the topmost object. Doing this before the spell's costs are
	// determined and paid (CR 601.2f-h) is what makes it impossible for a cost
	// paid from the source zone — "discard a card", "exile a blue card from your
	// hand", and the like — to ever select the very card being cast.
	obj := &game.StackObject{
		ID:                            g.IDGen.Next(),
		Kind:                          game.StackSpell,
		SourceID:                      cast.CardID,
		Face:                          cast.Face,
		Controller:                    playerID,
		Targets:                       append([]game.Target(nil), cast.Targets...),
		TargetCounts:                  targetCounts,
		ChosenModes:                   append([]int(nil), cast.ChosenModes...),
		XValue:                        cast.XValue,
		KickerPaid:                    cast.KickerPaid,
		KickerCount:                   cast.KickerCount,
		GiftPromised:                  cast.GiftPromised,
		GiftRecipient:                 cast.GiftRecipient,
		Bargained:                     cast.Bargained,
		OffspringPaid:                 cast.Offspring,
		Overloaded:                    cast.Overloaded,
		SourceZone:                    sourceZone,
		SplicedContent:                splice.contents,
		SplicedTargets:                splice.targets,
		SplicedTargetCounts:           splice.targetCounts,
		CastDuringControllerMainPhase: playerID == g.Turn.ActivePlayer && g.Turn.IsMainPhase(),
	}
	if !removeCastSourceCard(g, castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone) {
		return false
	}
	g.Stack.Push(obj)

	var prefs *payment.Preferences
	switch {
	case cast.Overloaded:
		overloadCost := append(cost.Mana(nil), spellDef.Overload.Val.Cost...)
		if cast.KickerPaid {
			kicker, _ := spellKicker(spellDef)
			overloadCost = append(overloadCost, kicker.Cost...)
		}
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&overloadCost,
			spellDef.AdditionalCosts,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	case payLifeFromTop:
		emptyMana := cost.Mana{}
		additional := append([]cost.Additional(nil), spellDef.AdditionalCosts...)
		additional = append(additional, payLifeManaValueAlternativeCost(spellDef, cast.XValue).AdditionalCosts...)
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&emptyMana,
			additional,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	case cast.Bestowed:
		bestowCost := bestowAlternativeCost(spellDef).ManaCost.Val
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&bestowCost,
			spellDef.AdditionalCosts,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	case plotted:
		emptyMana := cost.Mana{}
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&emptyMana,
			spellDef.AdditionalCosts,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	case freeLinkedExile:
		emptyMana := cost.Mana{}
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&emptyMana,
			spellDef.AdditionalCosts,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	case freePlayFromZone:
		emptyMana := cost.Mana{}
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&emptyMana,
			spellDef.AdditionalCosts,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	case spendAnyMana:
		anyManaCost := anyManaSymbols(spellDef.ManaCost)
		prefs = e.paymentPreferencesForCostFromSource(
			g,
			playerID,
			&anyManaCost,
			spellDef.AdditionalCosts,
			cast.XValue,
			card.ID,
			sourceZone,
			agents,
			log,
		)
	default:
		prefs = e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, cast.Face, spellDef, cast.XValue, agents, log)
	}
	permissions := castPermissionsForZone(g, playerID, card.ID, sourceZone, cast.Face)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		XValue:          cast.XValue,
		KickerPaid:      cast.KickerPaid,
		KickerCount:     cast.KickerCount,
		Bargained:       cast.Bargained,
		Offspring:       cast.Offspring,
		ChosenModes:     cast.ChosenModes,
		CastPermissions: permissions,
		Targets:         cast.Targets,
		Prefs:           prefs,
		SpliceManaCosts: splice.manaCosts,
		Bestowed:        cast.Bestowed,
	}
	switch {
	case cast.Overloaded:
		request.Alternative = opt.Val(overloadAlternativeCost(spellDef.Overload.Val.Cost))
	case cast.Bestowed:
		request.Alternative = opt.Val(bestowAlternativeCost(spellDef))
	case payLifeFromTop:
		request.Alternative = opt.Val(payLifeManaValueAlternativeCost(spellDef, cast.XValue))
	case plotted:
		request.Alternative = opt.Val(freeCastAlternativeCost())
	case foretold:
		request.Alternative = opt.Val(foretellAlternativeCost(spellDef))
	case freeLinkedExile:
		request.Alternative = opt.Val(freeCastAlternativeCost())
	case freePlayFromZone:
		request.Alternative = opt.Val(freeCastAlternativeCost())
	case spendAnyMana:
		request.Alternative = opt.Val(anyManaAlternativeCost(spellDef))
	default:
		// No alternative cost; the spell is cast for its normal mana cost.
	}
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		// CR 728: the proposed cast is illegal because its costs can't be paid,
		// so the proposal is undone — take the spell back off the stack and
		// return the card to the zone it was cast from.
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone)
		return false
	}
	if sourceZone == zone.Command && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	if freeLinkedExile {
		consumeCastLinkedExileForFreePermission(g, freeLinkedExilePermissionID)
	}
	recordExilePlayPermissionUse(g, exilePlayPermission)
	obj.Evoked = !cast.Overloaded && evokeAlternativeChosen(spellDef, prefs.AlternativeIndex)
	obj.Converted = !cast.Overloaded && convertedAlternativeChosen(spellDef, prefs.AlternativeIndex)
	obj.Dashed = !cast.Overloaded && dashAlternativeChosen(spellDef, prefs.AlternativeIndex)
	obj.Bestowed = cast.Bestowed
	obj.Flashback = paymentResult.CastPermission == payment.SpellCastPermissionFlashback
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)

	// stormCopyCount must be read before the spell-cast event is emitted, since
	// that event increments the storm count for later spells this turn.
	stormCopies := stormCopyCount(g, spellDef)
	// CR 601.2i: the costs are paid, so the spell becomes cast. The card is
	// already on the stack; emit its zone-change and spell-cast events now so
	// "when you cast" triggers fire after payment.
	emitSpellCastEvents(g, obj, game.Event{
		SourceID:                     cast.CardID,
		StackObjectID:                obj.ID,
		Controller:                   playerID,
		CardID:                       cast.CardID,
		Face:                         cast.Face,
		CardTypes:                    stackObjectCardTypes(obj, spellDef),
		CardSupertypes:               cardSupertypes(spellDef),
		CardSubtypes:                 stackObjectCardSubtypes(obj, spellDef),
		Colors:                       spellColors(spellDef),
		ManaValue:                    opt.Val(stackManaValue(spellDef, cast.XValue)),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		KickerPaid:                   cast.KickerPaid,
		FromZone:                     sourceZone,
		ToZone:                       zone.Stack,
	})
	createStormCopies(g, obj, spellDef, stormCopies)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
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

	// CR 601.2a: move the card to the stack before its costs are paid, so a
	// from-zone cost can't select the card being cast (see applyCastSpellWithChoices).
	obj := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       cast.CardID,
		Face:           game.FaceFront,
		Controller:     playerID,
		Targets:        []game.Target{game.PermanentTarget(cast.MutateTargetID)},
		TargetCounts:   []int{1},
		Mutate:         true,
		MutateTargetID: cast.MutateTargetID,
		SourceZone:     sourceZone,
	}
	if !removeCastSourceCard(g, castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone) {
		return false
	}
	g.Stack.Push(obj)

	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, game.FaceFront, spellDef, 0, agents, log)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:    playerID,
		CardID:      card.ID,
		SourceZone:  sourceZone,
		Card:        spellDef,
		Alternative: opt.Val(alternative),
		Prefs:       prefs,
	})
	if !ok {
		// CR 728: undo the proposal when costs can't be paid.
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone)
		return false
	}
	if sourceZone == zone.Command && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	emitSpellCastEvents(g, obj, game.Event{
		SourceID:                     cast.CardID,
		StackObjectID:                obj.ID,
		Controller:                   playerID,
		CardID:                       cast.CardID,
		Face:                         game.FaceFront,
		CardTypes:                    stackObjectCardTypes(obj, spellDef),
		CardSupertypes:               cardSupertypes(spellDef),
		CardSubtypes:                 stackObjectCardSubtypes(obj, spellDef),
		Colors:                       spellColors(spellDef),
		ManaValue:                    opt.Val(stackManaValue(spellDef, 0)),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		FromZone:                     sourceZone,
		ToZone:                       zone.Stack,
	})
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	return true
}

func canCastMutateSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, targetID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(castSourcePlayer(g, player, cardID, sourceZone), cardID, sourceZone) {
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
	if xValue != 0 &&
		!costHasVariableMana(manaCostPtr(spellDef.ManaCost)) &&
		!additionalCostsUseX(spellDef.AdditionalCosts) {
		return false
	}
	if !modesValidForSpellAt(g, playerID, spellDef, chosenModes) ||
		!isSupportedSpell(spellDef) ||
		(!spellDef.HasType(types.Instant) && !spellDef.HasType(types.Sorcery)) ||
		!targetsValidForSpell(g, playerID, spellDef, chosenModes, targets, game.CastBranch{}) ||
		!spellTargetCountsMatchX(g, playerID, spellDef, chosenModes, targets, xValue, game.CastBranch{}) ||
		!spellTargetCountsMatchKicker(g, playerID, spellDef, chosenModes, targets, 0, game.CastBranch{}) ||
		!spellTargetsSatisfyManaValueX(g, playerID, spellDef, chosenModes, targets, xValue, game.CastBranch{}) ||
		!canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     xValue,
		Targets:    targets,
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
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, spellDef, cast.ChosenModes, cast.Targets, agents, log, game.CastBranch{})
	if !ok || !canCastPreparedCopy(g, playerID, permanent, completedTargets, cast.XValue, cast.ChosenModes) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, cast.ChosenModes, cast.Targets, game.CastBranch{})
	if !ok {
		panic("validated prepared spell targets could not be segmented")
	}
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, sourceID, zone.Battlefield, game.FaceAlternate, spellDef, cast.XValue, agents, log)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     cast.XValue,
		Targets:    cast.Targets,
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
		AdditionalCostsPaid: paymentResult.AdditionalCostsPaid,
		SacrificedAsCostIDs: paymentResult.SacrificedIDs,
		SourceZone:          zone.Battlefield,
	}
	stormCopies := stormCopyCount(g, spellDef)
	g.Stack.Push(obj)
	emitTargetEvents(g, obj)
	emitEvent(g, game.Event{
		Kind:                         game.EventSpellCast,
		SourceID:                     sourceID,
		StackObjectID:                obj.ID,
		Controller:                   playerID,
		CardID:                       permanent.CardInstanceID,
		Face:                         game.FaceAlternate,
		PermanentID:                  permanent.ObjectID,
		TokenDef:                     permanent.TokenDef,
		CardTypes:                    stackObjectCardTypes(obj, spellDef),
		CardSupertypes:               cardSupertypes(spellDef),
		CardSubtypes:                 stackObjectCardSubtypes(obj, spellDef),
		Colors:                       spellColors(spellDef),
		ManaValue:                    opt.Val(stackManaValue(spellDef, cast.XValue)),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		FromZone:                     zone.Battlefield,
		ToZone:                       zone.Stack,

		PlayerEventOrdinalThisTurn: nextSpellCastOrdinalThisTurn(g, playerID),
	})
	createStormCopies(g, obj, spellDef, stormCopies)
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
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

func (e *Engine) canCastSpellFaceFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, effectiveKickerCount(kickerPaid, 0), false, false, false, false, false)
}

// canCastBestowSpellFaceFromZone validates a cast whose Bestow keyword is used
// (CR 702.103): the spell is cast for its Bestow alternative cost as an Aura
// spell, so its targets are validated on the bestowed branch (which requires the
// enchant-creature target) rather than the default branch.
func (e *Engine) canCastBestowSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, false, false, true, false)
}

// canCastGiftSpellFaceFromZone validates a cast whose Gift keyword action
// promises a gift (CR 702.171). Promising a gift activates the spell's
// gift-promised target specs, so its targets are validated on the promised
// branch rather than the default branch.
func (e *Engine) canCastGiftSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, true, false, false, false)
}

// canCastSpellFaceFromZoneWithMultikick validates a Multikicker cast whose
// kicker cost is paid kickerCount times (CR 702.32).
func (e *Engine) canCastSpellFaceFromZoneWithMultikick(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerCount int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, kickerCount, false, false, false, false, false)
}

// canCastBargainedSpellFaceFromZone validates a cast whose Bargain additional
// cost is paid (CR 702.166b). Bargaining activates the spell's bargained target
// specs, so its targets are validated on the bargained branch, and the payment
// planner requires the caster to be able to sacrifice an artifact, enchantment,
// or token.
func (e *Engine) canCastBargainedSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, false, true, false, false)
}

// canCastOffspringSpellFaceFromZone validates a cast whose Offspring additional
// cost is paid (CR 702.171). Paying the offspring cost activates the spell's
// offspring branch, so its targets are validated on the offspring branch, and
// the payment planner adds the offspring mana cost to the spell's total cost.
func (e *Engine) canCastOffspringSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, false, false, false, true)
}

// effectiveKickerCount resolves the number of times the kicker cost is paid from
// the binary kicker flag and the explicit Multikicker count: an explicit count
// wins, otherwise a paid ordinary kicker counts once.
func effectiveKickerCount(kickerPaid bool, kickerCount int) int {
	if kickerCount > 0 {
		return kickerCount
	}
	if kickerPaid {
		return 1
	}
	return 0
}

func (e *Engine) canCastOverloadedSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, chosenModes []int) bool {
	return e.canCastOverloadedSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, 0, chosenModes, false)
}

func (e *Engine) canCastOverloadedSpellFaceFromZoneWithOptions(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, nil, xValue, chosenModes, effectiveKickerCount(kickerPaid, 0), true, false, false, false, false)
}

func (*Engine) canCastSpellFaceFromZoneWithOptions(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerCount int, overloaded bool, giftPromised bool, bargained bool, bestowed bool, offspring bool) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	kickerPaid := kickerCount > 0
	branch := game.CastBranch{GiftPromised: giftPromised, Kicked: kickerPaid, Bargained: bargained, Bestowed: bestowed, Offspring: offspring}
	if xValue < 0 {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(castSourcePlayer(g, player, cardID, sourceZone), cardID, sourceZone) {
		return false
	}
	spellDef, ok := cardFaceDef(card, face)
	if !ok || !card.Def.CanChooseCastFace(face) {
		return false
	}
	if castFromZoneProhibited(g, playerID, sourceZone) {
		return false
	}
	plotted := sourceZone == zone.Exile && face == game.FaceFront && cardIsPlottedInExile(g, cardID)
	foretold := sourceZone == zone.Exile && face == game.FaceFront && cardIsForetoldInExile(g, cardID)
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
		hasRulePermission := canCastFromZoneByRuleEffect(g, playerID, cardID, sourceZone, face) ||
			canCastSpellsFromZoneByRuleEffect(g, playerID, cardID, sourceZone, face)
		freeLinkedExile := face == game.FaceFront && castLinkedExileForFree(g, playerID, cardID)
		if !g.AdventureCards[cardID] && !hasRulePermission && !plotted && !foretold && !freeLinkedExile {
			return false
		}
		if g.AdventureCards[cardID] && !hasRulePermission && !plotted && !foretold && !freeLinkedExile && face != game.FaceFront {
			return false
		}
	case zone.Library:
		if !canCastSpellsFromZoneByRuleEffect(g, playerID, cardID, sourceZone, face) {
			return false
		}
	default:
		return false
	}
	announcementDef := spellDef
	announcedManaCost := manaCostPtr(spellDef.ManaCost)
	if overloaded {
		if !spellDef.Overload.Exists {
			return false
		}
		announcementDef = overloadSpellDef(spellDef)
		overloadCost := spellDef.Overload.Val.Cost
		announcedManaCost = &overloadCost
	}
	if bestowed && !spellHasBestow(spellDef) {
		return false
	}
	if xValue != 0 &&
		!costHasVariableMana(announcedManaCost) &&
		!additionalCostsUseX(spellDef.AdditionalCosts) {
		return false
	}
	if !modesValidForSpellAtBranch(g, playerID, announcementDef, chosenModes, branch) || !isSupportedSpell(spellDef) || !targetsValidForSpell(g, playerID, announcementDef, chosenModes, targets, branch) {
		return false
	}
	if !spellTargetCountsMatchX(g, playerID, announcementDef, chosenModes, targets, xValue, branch) {
		return false
	}
	if !spellTargetCountsMatchKicker(g, playerID, announcementDef, chosenModes, targets, kickerCount, branch) {
		return false
	}
	if !spellTargetsSatisfyManaValueX(g, playerID, announcementDef, chosenModes, targets, xValue, branch) {
		return false
	}
	if !canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	if plotted && !isSorcerySpeed(g, playerID) {
		return false
	}
	if spellCastProhibited(g, playerID, spellDef) {
		return false
	}
	if spellCastLimitReached(g, playerID, spellDef) {
		return false
	}
	if kickerPaid && !spellHasKicker(spellDef) {
		return false
	}
	if giftPromised && !spellHasGift(spellDef) {
		return false
	}
	if bargained && !spellHasBargain(spellDef) {
		return false
	}
	if offspring && !spellHasOffspring(spellDef) {
		return false
	}
	if kickerCount > 1 && !spellHasMultikicker(spellDef) {
		return false
	}
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		XValue:          xValue,
		KickerPaid:      kickerPaid,
		KickerCount:     kickerCount,
		Bargained:       bargained,
		Offspring:       offspring,
		ChosenModes:     chosenModes,
		CastPermissions: castPermissionsForZone(g, playerID, card.ID, sourceZone, face),
		Targets:         targets,
		Bestowed:        bestowed,
	}
	switch {
	case overloaded:
		request.Alternative = opt.Val(overloadAlternativeCost(spellDef.Overload.Val.Cost))
	case bestowed:
		request.Alternative = opt.Val(bestowAlternativeCost(spellDef))
	case sourceZone == zone.Library && castFromZoneRequiresPayLife(g, playerID, card.ID, sourceZone, face):
		request.Alternative = opt.Val(payLifeManaValueAlternativeCost(spellDef, xValue))
	case plotted:
		request.Alternative = opt.Val(freeCastAlternativeCost())
	case foretold:
		request.Alternative = opt.Val(foretellAlternativeCost(spellDef))
	case sourceZone == zone.Exile && face == game.FaceFront && castLinkedExileForFree(g, playerID, card.ID):
		request.Alternative = opt.Val(freeCastAlternativeCost())
	case sourceZone == zone.Exile && castFromZoneWithoutPayingManaCost(g, playerID, card.ID, sourceZone, face):
		request.Alternative = opt.Val(freeCastAlternativeCost())
	case sourceZone == zone.Exile && castFromZoneAllowsAnyMana(g, playerID, card.ID, sourceZone, face):
		request.Alternative = opt.Val(anyManaAlternativeCost(spellDef))
	default:
		// No alternative cost; the spell is cast for its normal mana cost.
	}
	return paymentOrch.canPaySpellCosts(g, request)
}

func overloadAlternativeCost(manaCost cost.Mana) cost.Alternative {
	return cost.Alternative{
		Label:    "Overload",
		ManaCost: opt.Val(append(cost.Mana(nil), manaCost...)),
	}
}

// spellHasBestow reports whether a spell face carries the Bestow keyword
// (CR 702.103) so it may be cast for its Bestow alternative cost as an Aura.
func spellHasBestow(spellDef *game.CardDef) bool {
	_, ok := game.CardDefBestow(spellDef)
	return ok
}

// bestowAlternativeCost is the alternative cost of casting a card with Bestow for
// its Bestow cost rather than its mana cost (CR 702.103b). It replaces the mana
// cost with the Bestow keyword's cost; a card reaching this path always has a
// Bestow keyword, so a missing keyword falls back to an empty (free) cost.
func bestowAlternativeCost(spellDef *game.CardDef) cost.Alternative {
	bestow, _ := game.CardDefBestow(spellDef)
	return cost.Alternative{
		Label:    "Bestow",
		ManaCost: opt.Val(append(cost.Mana(nil), bestow.Cost...)),
	}
}

// freeCastAlternativeCost is the alternative cost of a cast that pays no mana
// cost ("without paying its mana cost"): the mana cost is emptied and no
// additional cost replaces it. It backs the plotted cast from exile (CR 718).
func freeCastAlternativeCost() cost.Alternative {
	return cost.Alternative{
		Label:    "Without paying its mana cost",
		ManaCost: opt.Val(cost.Mana{}),
	}
}

// foretellAlternativeCost is the alternative cost of casting a foretold card from
// exile for its foretell cost rather than its mana cost (CR 702.144). It replaces
// the mana cost with the card's Foretell cost; a card reaching this path always
// has a Foretell keyword, so a missing cost falls back to an empty (free) cost.
func foretellAlternativeCost(card *game.CardDef) cost.Alternative {
	foretellCost, _ := card.ForetellCost()
	return cost.Alternative{
		Label:    "Foretell",
		ManaCost: opt.Val(append(cost.Mana(nil), foretellCost...)),
	}
}

// anyManaSymbols rewrites a mana cost so mana of any type may pay it ("mana of
// any type can be spent to cast it.", Court of Locthwain): each colored,
// colorless, hybrid, or Twobrid symbol becomes one generic mana, and Phyrexian
// symbols keep their pay-2-life option as a generic Phyrexian symbol. Generic,
// variable, snow, and already-generic Phyrexian symbols are unchanged, so the
// cost keeps its size but loses every color restriction.
func anyManaSymbols(manaCost opt.V[cost.Mana]) cost.Mana {
	if !manaCost.Exists {
		return cost.Mana{}
	}
	rewritten := make(cost.Mana, 0, len(manaCost.Val))
	for _, symbol := range manaCost.Val {
		switch symbol.Kind {
		case cost.ColoredSymbol, cost.ColorlessSymbol, cost.HybridSymbol, cost.TwobridSymbol:
			rewritten = append(rewritten, cost.O(1))
		case cost.PhyrexianSymbol:
			rewritten = append(rewritten, cost.PhyrexianGeneric(1))
		default:
			rewritten = append(rewritten, symbol)
		}
	}
	return rewritten
}

// anyManaAlternativeCost is the alternative cost of a card cast from exile under
// a play permission whose SpendAnyMana flag lets mana of any type pay its cost
// (Court of Locthwain). The printed mana cost is replaced by its all-generic
// rewrite; the spell's own additional costs are appended by the payment planner.
func anyManaAlternativeCost(spellDef *game.CardDef) cost.Alternative {
	return cost.Alternative{
		Label:    "Spend mana of any type",
		ManaCost: opt.Val(anyManaSymbols(spellDef.ManaCost)),
	}
}

// payLifeManaValueAlternativeCost is the alternative cost imposed when a spell is
// cast from the top of the library under a permission that replaces its mana cost
// with paying life equal to its mana value ("If you cast a spell this way, pay
// life equal to its mana value rather than pay its mana cost.", Bolas's Citadel,
// Gwenom, Remorseless). The mana cost is emptied and a pay-life additional cost
// equal to the cast spell's mana value (counting the announced X) takes its place;
// the spell's own additional costs are appended by the payment planner.
func payLifeManaValueAlternativeCost(spellDef *game.CardDef, xValue int) cost.Alternative {
	alt := cost.Alternative{
		Label:    "Pay life equal to mana value",
		ManaCost: opt.Val(cost.Mana{}),
	}
	if manaValue := stackManaValue(spellDef, xValue); manaValue > 0 {
		alt.AdditionalCosts = []cost.Additional{{
			Kind:   cost.AdditionalPayLife,
			Text:   "pay life equal to its mana value",
			Amount: manaValue,
		}}
	}
	return alt
}

func overloadSpellDef(card *game.CardDef) *game.CardDef {
	overloaded := *card
	overloaded.SpellAbility = opt.Val(card.Overload.Val.SpellAbility)
	return &overloaded
}

func legalCastFacesForZone(g *game.Game, playerID game.PlayerID, card *game.CardInstance, sourceZone zone.Type) []game.FaceIndex {
	if sourceZone == zone.Graveyard || sourceZone == zone.Exile {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if sourceZone == zone.Exile && g.AdventureCards[card.ID] && face == game.FaceFront {
				faces = append(faces, face)
				continue
			}
			if sourceZone == zone.Exile && face == game.FaceFront && cardIsPlottedInExile(g, card.ID) {
				faces = append(faces, face)
				continue
			}
			if sourceZone == zone.Exile && face == game.FaceFront && cardIsForetoldInExile(g, card.ID) {
				faces = append(faces, face)
				continue
			}
			if canCastFromZoneByRuleEffect(g, playerID, card.ID, sourceZone, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	if sourceZone == zone.Library {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if canCastSpellsFromZoneByRuleEffect(g, playerID, card.ID, sourceZone, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	return card.Def.LegalCastFaces()
}

// castSourcePlayer returns the player whose zone physically holds the card being
// played or cast from sourceZone. A player plays or casts from their own hand,
// graveyard, library, or command zone, but exile is a shared zone: a card exiled
// from another player's library — e.g. a card Court of Locthwain exiled from an
// opponent's library — rests in its owner's exile bucket even though a different
// player is permitted to play or cast it. Resolving the owner keeps the
// containment, removal, and restore steps pointed at the bucket that actually
// holds the card. For every other card the owner is the actor, so this changes
// nothing.
func castSourcePlayer(g *game.Game, caster *game.Player, cardID id.ID, sourceZone zone.Type) *game.Player {
	if sourceZone != zone.Exile {
		return caster
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return caster
	}
	if owner := g.Players[card.Owner]; owner != nil {
		return owner
	}
	return caster
}

// foreignExileCastableCards returns the cards resting in another player's exile
// bucket that playerID currently holds an active permission to play or cast. A
// card Court of Locthwain exiled from an opponent's library lives in that
// opponent's exile bucket even though the controller may play it (for as long as
// it remains exiled) or, while monarch, free-cast it from the accumulated pool.
// The action enumerators scan the acting player's own exile bucket directly, so
// this surfaces only the cross-player cards those scans would otherwise miss.
//
// It gathers the affected card of every active RuleEffectPlayFromZone /
// RuleEffectCastFromZone whose exile permission reaches playerID, plus the
// linked-exile pool of every active RuleEffectCastLinkedExileForFree the player
// controls (the monarch-gated free cast — the effect only exists while the gate
// is satisfied). Cards already in playerID's own exile are omitted, cards no
// longer in exile are skipped, and each card is returned at most once. The
// result is empty for any player who holds no such cross-player exile
// permission, so it adds nothing to enumeration for every other card.
func foreignExileCastableCards(g *game.Game, playerID game.PlayerID) []id.ID {
	own, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var cards []id.ID
	seen := make(map[id.ID]struct{})
	consider := func(cardID id.ID) {
		if cardID == 0 || own.Exile.Contains(cardID) {
			return
		}
		if _, dup := seen[cardID]; dup {
			return
		}
		if z, ok := cardZone(g, cardID); !ok || z != zone.Exile {
			return
		}
		seen[cardID] = struct{}{}
		cards = append(cards, cardID)
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectPlayFromZone, game.RuleEffectCastFromZone:
			if effect.CastFromZone == zone.Exile {
				consider(effect.AffectedCardID)
			}
		case game.RuleEffectCastLinkedExileForFree:
			if effect.ExiledLinkKey == "" {
				continue
			}
			key := game.LinkedObjectKey{SourceID: effect.SourceCardID, LinkID: string(effect.ExiledLinkKey)}
			for _, ref := range linkedObjects(g, key) {
				consider(ref.CardID)
			}
		default:
			continue
		}
	}
	return cards
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
	case zone.Library:
		top, ok := player.Library.Top()
		return ok && top == cardID
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
	case zone.Library:
		if topID, ok := player.Library.Top(); ok {
			return []id.ID{topID}
		}
		return nil
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
	case zone.Library:
		return player.Library.Remove(cardID)
	default:
		return false
	}
}

// restoreCastSourceCard returns a card to the zone it was cast from. It reverses
// removeCastSourceCard when a proposed cast is abandoned because its costs can't
// be paid (CR 601.2h / 728, handling an illegal action by undoing the
// proposal), so the card is never lost.
func restoreCastSourceCard(player *game.Player, cardID id.ID, sourceZone zone.Type) {
	switch sourceZone {
	case zone.Hand:
		player.Hand.Add(cardID)
	case zone.Command:
		player.CommandZone.Add(cardID)
	case zone.Graveyard:
		player.Graveyard.Add(cardID)
	case zone.Exile:
		player.Exile.Add(cardID)
	case zone.Library:
		player.Library.Add(cardID)
	default:
	}
}
