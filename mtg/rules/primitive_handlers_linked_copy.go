package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// handleCopyCard offers the resolving controller the chance to copy the card
// exiled by this source under the imprint link (CR 707.12). The enclosing
// instruction's Optional flag already gathered "you may" consent, so this
// consent step performs no observable action: it merely reports whether a card
// linked to this source under prim.LinkID still rests in exile, so the paired
// PlayLinkedExiledCard cast (gated on this result's success) is offered only
// when a copy is actually available. A copy that is never cast ceases to exist
// (CR 707.12a), so declining the following cast leaves the game unchanged.
func handleCopyCard(r *effectResolver, prim game.CopyCard) effectResolved {
	res := effectResolved{accepted: true}
	if _, _, ok := linkedExiledCardInExile(r.game, r.obj, prim.LinkID); ok {
		res.succeeded = true
	}
	return res
}

// handlePlayLinkedExiledCard casts the card exiled by this source under the
// imprint link (CR 707.12). With prim.Copy set it casts a copy of that card
// (Isochron Scepter, Spellbinder): the linked card stays in exile and a spell
// carrying its copiable values is put on the stack, ceasing to exist when it
// leaves the stack. The cast chooses the first legal targets and modes with any
// X treated as 0 and, with prim.WithoutPayingManaCost set, pays no mana cost. It
// is a legal no-op when the linked card no longer rests in exile (the source
// left the battlefield since it imprinted, or the card moved) or has no legal
// way to be cast.
func handlePlayLinkedExiledCard(r *effectResolver, prim game.PlayLinkedExiledCard) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	cardID, _, ok := linkedExiledCardInExile(r.game, r.obj, prim.LinkID)
	if !ok {
		return res
	}
	if prim.Copy {
		if r.engine.castFreeCopyOfCard(r.game, playerID, cardID, r.agents, r.log) {
			res.succeeded = true
		}
		return res
	}
	if r.engine.castFreeTargetedSpell(r.game, playerID, cardID, zone.Exile, false, r.agents, r.log) {
		res.succeeded = true
	}
	return res
}

// linkedExiledCardInExile returns the first card linked to obj's source under
// linkID (the imprinted card, CR 707.12) that still rests in a player's exile
// zone, together with the player who owns that exile bucket. The object-scoped
// link follows the specific source permanent's object identity, so a copied
// permanent (a fresh object ID) shares no imprint link, and a source that has
// left the battlefield resolves to no link. It reports false when the source
// imprinted nothing, declined the imprint, or the imprinted card has since left
// exile.
func linkedExiledCardInExile(g *game.Game, obj *game.StackObject, linkID string) (id.ID, game.PlayerID, bool) {
	for _, ref := range linkedObjects(g, linkedObjectByObjectKey(g, obj, linkID)) {
		cardID := ref.CardID
		if cardID == 0 {
			cardID = ref.ObjectID
		}
		if cardID == 0 {
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for i := range g.Players {
			if g.Players[i].Exile.Contains(card.ID) {
				return card.ID, g.Players[i].ID, true
			}
		}
	}
	return 0, 0, false
}

// castFreeCopyOfCard casts a copy of cardID (a card resting in exile) for
// controllerID without paying its mana cost, choosing the first legal
// modes/targets with X as 0 and putting a copy spell on the stack (CR 707.12).
// The copy carries the card's copiable values through an embedded token
// definition rather than the card instance, so the original card stays in exile
// and the copy ceases to exist when it leaves the stack. It returns false
// (casting nothing) when the card has no legal cast choice.
func (*Engine) castFreeCopyOfCard(g *game.Game, controllerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	modes, targets, ok := firstLegalSpellCastChoice(g, controllerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, controllerID, spellDef, modes, targets, game.CastBranch{})
	if !ok {
		panic("validated free-copy spell targets could not be segmented")
	}
	obj := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       cardID,
		SourceCardID:   card.ID,
		SourceTokenDef: copyCardDef(card.Def),
		Face:           game.FaceFront,
		Controller:     controllerID,
		Targets:        append([]game.Target(nil), targets...),
		TargetCounts:   targetCounts,
		ChosenModes:    append([]int(nil), modes...),
		Copy:           true,
		SourceZone:     zone.Exile,
	}
	g.Stack.Push(obj)
	emitTargetEvents(g, obj)
	emitEvent(g, game.Event{
		Kind:                         game.EventSpellCast,
		SourceID:                     cardID,
		StackObjectID:                obj.ID,
		Controller:                   controllerID,
		CardID:                       card.ID,
		Face:                         game.FaceFront,
		CardTypes:                    cardTypes(spellDef),
		CardSupertypes:               cardSupertypes(spellDef),
		CardSubtypes:                 cardSubtypes(spellDef),
		Colors:                       spellColors(spellDef),
		ManaValue:                    opt.Val(stackManaValue(spellDef, 0)),
		ManaSpentToCast:              opt.Val(0),
		ManaFromCreaturesSpentToCast: opt.Val(0),
		FromZone:                     zone.Exile,
		ToZone:                       zone.Stack,
		PlayerEventOrdinalThisTurn:   nextSpellCastOrdinalThisTurn(g, controllerID),
	})
	return true
}
