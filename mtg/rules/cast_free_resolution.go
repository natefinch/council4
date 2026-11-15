package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

type freeCastOption struct {
	cardID id.ID
	cast   action.CastSpellAction
}

func (e *Engine) freeCastOptionsForCard(g *game.Game, playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int]) []freeCastOption {
	return e.freeCastSpellOptions(g, playerID, card, fromZone, selection, maxManaValue, freeCastSpell{})
}

// Only faces are selected before CR 601.2a. Modes, additional costs and targets
// are announced on the stack, using completion-preserving shared choices.
func (e *Engine) freeCastSpellOptions(g *game.Game, playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], spell freeCastSpell) []freeCastOption {
	if card == nil || card.Def == nil || castFromZoneProhibited(g, playerID, fromZone) {
		return nil
	}
	var options []freeCastOption
	for _, face := range freeCastSpellFaces(playerID, card, fromZone, spell) {
		def := cardFaceOrDefault(card, face)
		if !spell.faceAvailable(g, playerID, card, face, def, selection, maxManaValue) {
			continue
		}
		cast := action.CastSpellAction{CardID: card.ID, SourceZone: fromZone, Face: face}
		object := &game.StackObject{Kind: game.StackSpell, SourceID: card.ID, Controller: playerID,
			SourceZone: fromZone, Face: face, Copy: spell.copyCard, Casting: true}
		if spell.copyCard {
			object.SourceTokenDef = copyCardDef(card.Def)
		}
		holder, _ := playerHoldingCastSource(g, card.ID, fromZone)
		source := playerID
		if holder != nil {
			source = holder.ID
		}
		view := priorityCastView(g, object, source)
		if len(e.resolutionSpellChoices(view, cast, freeCastCostRequest(playerID, def, spell.costCast(cast)), true)) > 0 {
			options = append(options, freeCastOption{cardID: card.ID, cast: cast})
		}
	}
	return options
}

// freeCastFaces lists the faces a resolving effect may let playerID cast
// without paying its mana cost. A fuse card cast from its caster's own hand may
// be cast fused even without paying its mana cost (CR 702.102a; Wear // Tear
// ruling 2013-04-15), and the fused spell's combined mana value is what any
// mana value limit sees (CR 709.4d). From any other zone only the halves are
// offered.
func freeCastFaces(playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type) []game.FaceIndex {
	faces := card.Def.LegalCastFacesFromZone(fromZone)
	if canFuseFromZone(playerID, card, fromZone) {
		faces = append(faces, game.FaceFused)
	}
	return faces
}

// freeCastSpellFaces lists the faces spell may put on the stack. A card cast
// "transformed" is put on the stack with its back face up (CR 712.11a), and only
// that face is evaluated to determine whether it can be cast (CR 712.11c). Only
// a transforming double-faced card has a back face to cast that way; any other
// card, including a single-faced card that is a copy of one, can't be cast
// transformed and offers no face.
func freeCastSpellFaces(playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, spell freeCastSpell) []game.FaceIndex {
	if !spell.transformed {
		return freeCastFaces(playerID, card, fromZone)
	}
	if card.Def.IsTransformingDoubleFaced() && card.Def.Back.Exists {
		return []game.FaceIndex{game.FaceBack}
	}
	return nil
}

// freeCastCardFaceAvailable is freeCastFaceAvailable for the card itself cast
// with face up. A transforming double-faced card cast with its back face up has
// only its back face's characteristics, but its mana value is calculated using
// its front face's mana cost (CR 712.8c), so a mana value limit or selection
// sees the front mana value.
func freeCastCardFaceAvailable(g *game.Game, player game.PlayerID, card *game.CardInstance, face game.FaceIndex, def *game.CardDef, selection game.Selection, maxManaValue opt.V[int]) bool {
	if face == game.FaceBack && card.Def.IsTransformingDoubleFaced() {
		view := copyCardDef(def)
		view.ManaCost = card.Def.ManaCost
		return freeCastFaceMatchesSelection(g, card, view, selection, player) &&
			(!maxManaValue.Exists || stackManaValue(view, 0) <= maxManaValue.Val) &&
			freeCastFaceCastable(g, player, card.ID, def)
	}
	return freeCastFaceAvailable(g, player, card, card.ID, def, selection, maxManaValue)
}

// freeCastFaceAvailable reports whether player may cast def, a face of card,
// without paying its mana cost. spellCard is card.ID when the card itself is
// cast and 0 when a copy of it is (CR 707.12).
func freeCastFaceAvailable(g *game.Game, player game.PlayerID, card *game.CardInstance, spellCard id.ID, def *game.CardDef, selection game.Selection, maxManaValue opt.V[int]) bool {
	return freeCastFaceMatchesSelection(g, card, def, selection, player) &&
		(!maxManaValue.Exists || stackManaValue(def, 0) <= maxManaValue.Val) &&
		freeCastFaceCastable(g, player, spellCard, def)
}

func freeCastFaceCastable(g *game.Game, player game.PlayerID, spellCard id.ID, def *game.CardDef) bool {
	castFace := castSelectionFace(g, player, spellCard, def, false)
	return cardCastRestrictionsSatisfied(g, player, def) && !spellCastProhibited(g, player, castFace) &&
		!spellCastLimitReached(g, player, castFace) && isSupportedSpell(def)
}

func freeCastFaceMatchesSelection(g *game.Game, card *game.CardInstance, spellDef *game.CardDef, selection game.Selection, viewer game.PlayerID) bool {
	faceCard := *card
	faceCard.Def = spellDef
	return handCardMatchesSelection(g, &faceCard, selection, viewer)
}

func freeCastBranchOptions(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type, face game.FaceIndex, spellDef *game.CardDef) []action.CastSpellAction {
	base := []action.CastSpellAction{{
		CardID:     cardID,
		SourceZone: fromZone,
		Face:       face,
	}}
	if spellHasKicker(spellDef) {
		base = append(base, action.CastSpellAction{
			CardID:      cardID,
			SourceZone:  fromZone,
			Face:        face,
			KickerPaid:  true,
			KickerCount: 1,
		})
	}
	base = expandFreeCastBooleanBranch(base, spellHasBargain(spellDef), func(cast *action.CastSpellAction) {
		cast.Bargained = true
	})
	base = expandFreeCastBooleanBranch(base, spellHasOffspring(spellDef), func(cast *action.CastSpellAction) {
		cast.Offspring = true
	})
	if !spellHasGift(spellDef) {
		return base
	}
	withGift := append([]action.CastSpellAction(nil), base...)
	for _, cast := range base {
		for _, opponent := range aliveOpponents(g, playerID) {
			gift := cast
			gift.GiftPromised = true
			gift.GiftRecipient = opponent
			withGift = append(withGift, gift)
		}
	}
	return withGift
}

func expandFreeCastBooleanBranch(base []action.CastSpellAction, enabled bool, apply func(*action.CastSpellAction)) []action.CastSpellAction {
	if !enabled {
		return base
	}
	expanded := append([]action.CastSpellAction(nil), base...)
	for _, cast := range base {
		branch := cast
		apply(&branch)
		expanded = append(expanded, branch)
	}
	return expanded
}

func freeCastCostRequest(playerID game.PlayerID, spellDef *game.CardDef, cast action.CastSpellAction) payment.SpellRequest {
	request := priorityCastRequest(playerID, spellDef, cast.SourceZone, cast,
		[]payment.SpellCastPermission{payment.SpellCastPermissionDefault}, nil)
	request.Alternative = opt.Val(freeCastAlternativeCost())
	return request
}

func (e *Engine) chooseAndCastFreeSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.chooseAndCastFreeSpellFromSourceMay(g, sourcePlayer, controllerID, cardID, fromZone, selection, maxManaValue, exileOnResolution, false, agents, log)
}

func (e *Engine) chooseAndCastFreeSpellFromSourceMay(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], exileOnResolution, mayDecline bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.chooseAndCastFreeSpell(g, sourcePlayer, controllerID, cardID, fromZone, selection, maxManaValue, freeCastSpell{exileOnResolution: exileOnResolution}, mayDecline, agents, log)
}

// freeCastSpell records how a free-cast route marks the spell it puts on the
// stack.
type freeCastSpell struct {
	exileOnResolution bool
	// copyCard casts a copy of the card rather than the card itself.
	copyCard bool
	// suspend marks a spell cast by its suspend ability; a creature spell cast
	// this way gains haste (CR 702.62a).
	suspend bool
	// transformed casts a transforming double-faced card with its back face up
	// (CR 712.11a). A card without such a back face can't be cast this way.
	transformed bool
}

// faceAvailable reports whether def, a face of card, may be cast by this
// route. A cast copy of the card (CR 707.12) is not the card itself, so it is
// evaluated without the card's identity.
func (s freeCastSpell) faceAvailable(g *game.Game, player game.PlayerID, card *game.CardInstance, face game.FaceIndex, def *game.CardDef, selection game.Selection, maxManaValue opt.V[int]) bool {
	if s.copyCard {
		return freeCastFaceAvailable(g, player, card, 0, def, selection, maxManaValue)
	}
	return freeCastCardFaceAvailable(g, player, card, face, def, selection, maxManaValue)
}

// costCast returns cast as its costs are evaluated. A cast copy has its own
// costs, not the identity of the card it copies, so payment must not reserve or
// consume that card as the casting source.
func (s freeCastSpell) costCast(cast action.CastSpellAction) action.CastSpellAction {
	if s.copyCard {
		cast.CardID = 0
	}
	return cast
}

func (e *Engine) chooseAndCastFreeSpell(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], spell freeCastSpell, mayDecline bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil || !castSourceContains(sourcePlayer, cardID, fromZone) || castFromZoneProhibited(g, controllerID, fromZone) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	options := e.freeCastSpellOptions(g, controllerID, card, fromZone, selection, maxManaValue, spell)
	if len(options) == 0 {
		return false
	}
	index := 0
	if len(options) > 1 || mayDecline {
		index, ok = e.chooseFreeCastOption(g, controllerID, card, fromZone, options, mayDecline, agents, log)
		if !ok {
			return false
		}
	}
	cast := options[index].cast
	return e.resolutionSpell(g, sourcePlayer, card, cast,
		freeCastCostRequest(controllerID, cardFaceOrDefault(card, cast.Face), spell.costCast(cast)), spell, true, agents, log) == actionApplied
}

func (e *Engine) chooseFreeCastOption(g *game.Game, controllerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, options []freeCastOption, mayDecline bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (int, bool) {
	choices := make([]game.ChoiceOption, len(options))
	for index, option := range options {
		choices[index] = game.ChoiceOption{Index: index, Label: freeCastOptionLabel(cardFaceOrDefault(card, option.cast.Face), option.cast), Card: cardChoiceInfo(g, card.ID)}
	}
	if mayDecline {
		choices = append(choices, game.ChoiceOption{Index: len(options), Label: "Don't cast " + cardChoiceLabel(g, card.ID), Card: cardChoiceInfo(g, card.ID)})
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoiceResolution, Player: controllerID, Prompt: "Choose how to cast " + cardChoiceLabel(g, card.ID) + " without paying its mana cost",
		Options: choices, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0}, Private: fromZone == zone.Hand || fromZone == zone.Library,
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(options) {
		return 0, false
	}
	return selected[0], true
}

func freeCastOptionLabel(spellDef *game.CardDef, cast action.CastSpellAction) string {
	label := spellDef.Name
	if cast.Face == game.FaceFused {
		label += " (fused)"
	} else if cast.Face == game.FacePrototype {
		label += " (prototype)"
	}
	if cast.Face != game.FaceFused && len(cast.ChosenModes) > 0 {
		label += fmt.Sprintf(" modes %v", cast.ChosenModes)
	}
	if cast.KickerCount > 1 {
		label += fmt.Sprintf(" with multikicker %d", cast.KickerCount)
	} else if cast.KickerPaid {
		label += " with kicker"
	}
	if cast.Bargained {
		label += " bargained"
	}
	if cast.Offspring {
		label += " with offspring"
	}
	if cast.GiftPromised {
		label += fmt.Sprintf(" promising gift to player %d", cast.GiftRecipient)
	}
	return label
}

func (e *Engine) castFreeOption(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpell(g, sourcePlayer, controllerID, card, cast, freeCastSpell{exileOnResolution: exileOnResolution}, agents, log)
}

func (e *Engine) castFreeOptionWithCopy(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, exileOnResolution, copyCard bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpell(g, sourcePlayer, controllerID, card, cast, freeCastSpell{exileOnResolution: exileOnResolution, copyCard: copyCard}, agents, log)
}

func (e *Engine) castFreeSpell(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, spell freeCastSpell, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if card == nil || castFromZoneProhibited(g, controllerID, cast.SourceZone) ||
		!freeCastFaceCastable(g, controllerID, spell.costCast(cast).CardID, cardFaceOrDefault(card, cast.Face)) ||
		!spell.copyCard && (sourcePlayer == nil || !castSourceContains(sourcePlayer, card.ID, cast.SourceZone)) {
		return false
	}
	return e.resolutionSpell(g, sourcePlayer, card, cast, freeCastCostRequest(controllerID, cardFaceOrDefault(card, cast.Face), spell.costCast(cast)), spell, false, agents, log) == actionApplied
}
