package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

var faceDownCastCost = cost.Mana{cost.O(3)}

func faceDownDisguiseWardBody() game.StaticAbility {
	return game.StaticAbility{
		Text:             "Ward {2}",
		KeywordAbilities: []game.KeywordAbility{game.WardKeyword{Cost: cost.Mana{cost.O(2)}}},
	}
}

func faceDownCostForCard(card *game.CardDef, kind game.FaceDownKind) (cost.Mana, bool) {
	costs := faceDownCostsForCard(card, kind)
	if len(costs) == 0 {
		return nil, false
	}
	return costs[0], true
}

func faceDownCostsForCard(card *game.CardDef, kind game.FaceDownKind) []cost.Mana {
	var costs []cost.Mana
	if kind == game.FaceDownManifest || kind == game.FaceDownCloak {
		if card.HasType(types.Creature) && card.ManaCost.Exists {
			costs = append(costs, card.ManaCost.Val)
		}
	}
	for i := range card.ActivatedAbilities {
		ability := &card.ActivatedAbilities[i]
		if kind == game.FaceDownMorph || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.ActivatedBodyMorphCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
		if kind == game.FaceDownDisguise || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.ActivatedBodyDisguiseCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
	}
	for i := range card.StaticAbilities {
		ability := &card.StaticAbilities[i]
		if kind == game.FaceDownMorph || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.StaticBodyMorphCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
		if kind == game.FaceDownDisguise || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.StaticBodyDisguiseCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
	}
	return costs
}

func faceDownKindsForCard(card *game.CardDef) []game.FaceDownKind {
	var kinds []game.FaceDownKind
	if _, ok := faceDownCostForCard(card, game.FaceDownMorph); ok {
		kinds = append(kinds, game.FaceDownMorph)
	}
	if _, ok := faceDownCostForCard(card, game.FaceDownDisguise); ok {
		kinds = append(kinds, game.FaceDownDisguise)
	}
	return kinds
}

func (e *Engine) legalFaceDownCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || splitSecondOnStack(g) {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, face := range legalFaceDownFaces(card.Def) {
			spellDef := cardFaceOrDefault(card, face)
			for _, kind := range faceDownKindsForCard(spellDef) {
				if e.canCastFaceDown(g, playerID, cardID, face, kind) {
					actions = append(actions, actionBuild.castFaceDown(cardID, face, kind))
				}
			}
		}
	}
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, face := range legalFaceDownFaces(card.Def) {
			for _, kind := range faceDownKindsForCard(cardFaceOrDefault(card, face)) {
				for _, permission := range permanentTypePlayPermissions(g, playerID, zone.Graveyard, faceDownSpellDef()) {
					if e.canCastGraveyardFaceDown(g, playerID, cardID, face, kind, permission) {
						actions = append(actions, action.CastFaceDownFromZone(cardID, zone.Graveyard, face, kind).WithPlayPermission(permission))
					}
				}
			}
		}
	}
	return actions
}

func faceDownSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}}
}

// faceDownStackSpellDef returns the characteristics of a face-down spell on the
// stack: a colorless 2/2 creature with no text, name, subtypes, supertypes or
// mana cost, plus ward {2} for disguise (CR 702.37c, 702.168b, 708.2, 708.4).
// The hidden card's own characteristics are not part of the spell.
func faceDownStackSpellDef(obj *game.StackObject) *game.CardDef {
	def := faceDownCopyDef(&game.Permanent{})
	if obj.FaceDownKind == game.FaceDownDisguise || obj.FaceDownKind == game.FaceDownCloak {
		def.StaticAbilities = []game.StaticAbility{faceDownDisguiseWardBody()}
	}
	return def
}

// faceDownSpellRequest determines a face-down cast's total cost from the {3}
// alternative cost and the face-down spell's characteristics (CR 702.37c,
// 702.168b, 601.2f): increases, reductions and additional costs that apply to
// a colorless 2/2 creature spell with no name, subtypes or mana cost apply,
// and those keyed to the hidden card's own characteristics do not.
func faceDownSpellRequest(player game.PlayerID, cardID id.ID, sourceZone zone.Type) payment.SpellRequest {
	permission := payment.SpellCastPermissionDefault
	if sourceZone == zone.Graveyard {
		permission = payment.SpellCastPermissionRuleEffect
	}
	return payment.SpellRequest{
		PlayerID: player, CardID: cardID, SourceZone: sourceZone, Card: faceDownSpellDef(),
		Alternative:     opt.Val(cost.Alternative{Label: "Face down", ManaCost: opt.Val(faceDownCastCost)}),
		CastPermissions: []payment.SpellCastPermission{permission},
	}
}

func (e *Engine) canCastGraveyardFaceDown(g *game.Game, player game.PlayerID, cardID id.ID, face game.FaceIndex, kind game.FaceDownKind, permission game.PlayPermission) bool {
	spell := castSelectionFace(g, player, faceDownSpellDef(), false)
	if !canAct(g, player) || player != g.Turn.PriorityPlayer || !canCastAtCurrentTiming(g, player, spell) || splitSecondOnStack(g) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !g.Players[player].Graveyard.Contains(cardID) || (kind != game.FaceDownMorph && kind != game.FaceDownDisguise) {
		return false
	}
	def, ok := cardFaceDef(card, face)
	if !ok {
		return false
	}
	if _, ok := faceDownCostForCard(def, kind); !ok {
		return false
	}
	if !slices.Contains(permanentTypePlayPermissions(g, player, zone.Graveyard, spell), permission) ||
		castFromZoneProhibited(g, player, zone.Graveyard) ||
		spellCastProhibited(g, player, spell) || spellCastLimitReached(g, player, spell) {
		return false
	}
	return e.faceDownCostProposable(g, player, action.CastFaceDownAction{CardID: cardID, Face: face, FaceDownKind: kind}, zone.Graveyard)
}

// faceDownCostProposable reports whether a face-down cast may be proposed:
// some total cost for it is payable now, or becomes payable once the caster
// activates the payment-time mana abilities available to them (CR 601.2g),
// which manualManaMayPay checks by trying them. Those abilities are tried in
// the state they'd be activated in, with the card already on the stack face
// down (CR 601.2a), so its old zone no longer contributes to them.
func (e *Engine) faceDownCostProposable(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, sourceZone zone.Type) bool {
	request := faceDownSpellRequest(player, cast.CardID, sourceZone)
	if paymentOrch.canPaySpellCosts(g, request) {
		return true
	}
	proposed := g.Clone()
	proposed.Execution = game.ExecutionDiagnostics{}
	if _, _, ok := stackFaceDownCast(proposed, player, cast, sourceZone); !ok ||
		!e.hasManualPaymentManaAbility(proposed, player) {
		return false
	}
	for _, prepared := range paymentOrch.planner(proposed).PrepareSpellCosts(request) {
		if _, feasible := preparedSpellCostFeasibility(proposed, request, prepared, e.manualManaMayPay, nil); feasible {
			return true
		}
	}
	return false
}

func legalFaceDownFaces(card *game.CardDef) []game.FaceIndex {
	var faces []game.FaceIndex
	for _, face := range card.FaceIndexes() {
		if def, ok := card.FaceDef(face); ok && len(faceDownKindsForCard(def)) > 0 {
			faces = append(faces, face)
		}
	}
	return faces
}

// canCastFaceDown judges timing and casting prohibitions against the face-down
// spell's characteristics, not the hidden card's (CR 702.37c, 708.4): an
// as-though-flash permission covering a colorless 2/2 creature spell applies,
// while the face-up card's own flash does not.
func (e *Engine) canCastFaceDown(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex, kind game.FaceDownKind) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || splitSecondOnStack(g) || kind == game.FaceDownNone {
		return false
	}
	spell := castSelectionFace(g, playerID, faceDownSpellDef(), false)
	if !canCastAtCurrentTiming(g, playerID, spell) || castFromZoneProhibited(g, playerID, zone.Hand) ||
		spellCastProhibited(g, playerID, spell) || spellCastLimitReached(g, playerID, spell) {
		return false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef, ok := cardFaceDef(card, face)
	if !ok {
		return false
	}
	if _, ok := faceDownCostForCard(spellDef, kind); !ok {
		return false
	}
	return e.faceDownCostProposable(g, playerID, action.CastFaceDownAction{CardID: cardID, Face: face, FaceDownKind: kind}, zone.Hand)
}

// attemptCastFaceDownWithChoices casts a face-down spell for playerID. A cast
// whose locked total cost isn't paid is reversed (CR 733.1) and recorded in
// attempts, which may be nil.
func (e *Engine) attemptCastFaceDownWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastFaceDownAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog, attempts *castAttempts) actionOutcome {
	if cast.SourceZone == zone.Graveyard {
		return e.attemptGraveyardFaceDown(g, playerID, cast, agents, log, attempts)
	}
	if cast.SourceZone != zone.None && cast.SourceZone != zone.Hand || cast.Permission != (game.PlayPermission{}) {
		return actionFailed
	}
	if !e.canCastFaceDown(g, playerID, cast.CardID, cast.Face, cast.FaceDownKind) {
		return actionFailed
	}
	return e.castFaceDownSpell(g, playerID, cast, zone.Hand, agents, log, attempts)
}

func (e *Engine) attemptGraveyardFaceDown(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog, attempts *castAttempts) actionOutcome {
	var ok bool
	cast.Permission, ok = e.selectPermanentTypePlay(g, player, cast.Permission, zone.Graveyard, faceDownSpellDef(), false, agents, log)
	if !ok {
		return actionFailed
	}
	if !e.canCastGraveyardFaceDown(g, player, cast.CardID, cast.Face, cast.FaceDownKind, cast.Permission) {
		return actionFailed
	}
	return e.castFaceDownSpell(g, player, cast, zone.Graveyard, agents, log, attempts)
}

// faceDownCastKey identifies the face-down cast an attempt redoes (see
// castAttempts).
func faceDownCastKey(cast action.CastFaceDownAction) string {
	return fmt.Sprintf("face down|%+v", cast)
}

// stackFaceDownCast moves the card cast to the stack as a face-down spell
// controlled by player (CR 601.2a, 702.37c), returning the spell and a function
// that undoes the move, returning the card to its place in sourceZone.
func stackFaceDownCast(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, sourceZone zone.Type) (*game.StackObject, func(), bool) {
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return nil, nil, false
	}
	owner := g.Players[player]
	counterTransition := prepareCastCounterTransition(g, card, sourceZone)
	place, ok := takeCastSourceCard(owner, cast.CardID, sourceZone)
	if !ok {
		return nil, nil, false
	}
	counterTransition.apply(g)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: cast.CardID, Controller: player,
		Face: cast.Face, FaceDown: true, FaceDownFace: cast.Face, FaceDownKind: cast.FaceDownKind,
		SourceZone: sourceZone,
	}
	g.Stack.Push(obj)
	return obj, func() {
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(place)
		counterTransition.undo(g)
	}, true
}

// castFaceDownSpell casts a validated face-down spell: the card moves to the
// stack face down first (CR 601.2a), then the {3} alternative cost is proposed
// and the total cost is determined from the face-down characteristics and
// locked (CR 702.37c, 601.2b, 601.2f). If that total includes a mana payment,
// the caster may then activate mana abilities (CR 601.2g) before paying it
// (CR 601.2h), and only then does the spell become cast (CR 601.2i). An unpaid
// proposal is undone along with its reversible mana activations (CR 733.1).
func (e *Engine) castFaceDownSpell(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, sourceZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog, attempts *castAttempts) actionOutcome {
	obj, rollback, ok := stackFaceDownCast(g, player, cast, sourceZone)
	if !ok {
		return actionFailed
	}
	request := faceDownSpellRequest(player, cast.CardID, sourceZone)
	proposals := newCastProposals(attempts, faceDownCastKey(cast), nil)
	proposal, ok := e.proposeSpellCost(g, request, false, proposals, agents, log)
	if !ok {
		rollback()
		return proposals.exhaust()
	}
	attempt := fmt.Sprintf("%s|%+v", faceDownCastKey(cast), proposal)
	prepared, ok := e.lockSpellCostProposal(g, request, proposal)
	if !ok {
		rollback()
		return proposals.reverseUnpayable(attempt, 0, proposal, nil, nil)
	}
	request.PreparedCost = &prepared
	prefs, undoMana := e.preparedSpellPaymentPreferences(g, request, prepared, proposal, agents, log)
	request.Prefs = prefs
	riderSnapshot, _ := manaSpendRiderSnapshot(g, player)
	result, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		undoMana()
		rollback()
		return proposals.reverse(attempt, 0, proposal, nil)
	}
	recordPermanentTypePlay(g, player, cast.Permission)
	obj.AdditionalCostsPaid = result.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = result.SacrificedIDs
	obj.PaidCostSubjects = result.PaidCostSubjects
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(result.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(result.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(result.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(result.PoolSpend)
	// The event carries only the face-down characteristics (CR 708.2), changed
	// by any type effects that apply to the controller's creature spells. The
	// card is turned face down before it moves (CR 708.4), so both the zone
	// change and the cast see that authoritative snapshot, never the hidden
	// card's printed characteristics.
	typeLine := stackSpellTypeLine(g, obj, faceDownStackSpellDef(obj))
	emitSpellCastEventsWithTargets(g, obj, game.Event{
		SourceID: cast.CardID, StackObjectID: obj.ID, CardID: cast.CardID, Controller: player,
		Face: cast.Face, FaceDown: true, CardCharacteristicsKnown: true,
		CardTypes: typeLine.types, CardSupertypes: typeLine.supertypes, CardSubtypes: typeLine.subtypes,
		Colors:                       nil, // Face-down spells are colorless (CR 708.2b).
		ManaValue:                    opt.Val(0),
		ManaSpentToCast:              opt.Val(obj.ManaSpentToCast),
		ManaFromCreaturesSpentToCast: opt.Val(obj.ManaFromCreaturesSpentToCast),
		FromZone:                     sourceZone, ToZone: zone.Stack,
	}, nil)
	resolveSpellCastManaSpendRiders(g, player, riderSnapshot, result.PoolSpend, request.Card, obj)
	return actionApplied
}

func (e *Engine) legalTurnFaceUpActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if e.canTurnFaceUp(g, playerID, permanent.ObjectID) {
			actions = append(actions, actionBuild.turnFaceUp(permanent.ObjectID))
		}
	}
	return actions
}

func (*Engine) canTurnFaceUp(g *game.Game, playerID game.PlayerID, permanentID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok || permanent.PhasedOut || !permanent.FaceDown || effectiveController(g, permanent) != playerID || permanent.FaceDownKind == game.FaceDownNone {
		return false
	}
	return len(payableFaceDownCostsForPermanent(g, playerID, permanent)) > 0
}

func (e *Engine) applyTurnFaceUpWithChoices(g *game.Game, playerID game.PlayerID, permanentID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canTurnFaceUp(g, playerID, permanentID) {
		return false
	}
	permanent, _ := permanentByObjectID(g, permanentID)
	manaCost, ok := e.chooseFaceDownCostOptions(
		g,
		playerID,
		payableFaceDownCostsForPermanent(g, playerID, permanent),
		agents,
		log,
	)
	if !ok {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Prefs: prefs}) {
		return false
	}
	permanent.Face = permanent.FaceDownFace
	permanent.FaceDown = false
	permanent.FaceDownKind = game.FaceDownNone
	permanent.FaceDownCharacteristics = opt.V[game.FaceDownCharacteristics]{}
	emitFaceDownRevealEvent(g, permanent)
	emitMergedTurnFaceUpRevealEvents(g, permanent)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentTurnedFaceUp,
		Controller:  playerID,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		PermanentID: permanent.ObjectID,
	})
	return true
}

func payableFaceDownCosts(g *game.Game, playerID game.PlayerID, face *game.CardDef, kind game.FaceDownKind) []cost.Mana {
	var costs []cost.Mana
	for _, manaCost := range faceDownCostsForCard(face, kind) {
		if paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost}) {
			costs = append(costs, manaCost)
		}
	}
	return costs
}

func payableFaceDownCostsForPermanent(g *game.Game, playerID game.PlayerID, permanent *game.Permanent) []cost.Mana {
	var costs []cost.Mana
	for _, face := range faceDownTurnUpFaces(g, permanent) {
		for _, manaCost := range payableFaceDownCosts(g, playerID, face, permanent.FaceDownKind) {
			duplicate := false
			for _, prior := range costs {
				if prior.String() == manaCost.String() {
					duplicate = true
					break
				}
			}
			if !duplicate {
				costs = append(costs, manaCost)
			}
		}
	}
	return costs
}

func faceDownTurnUpFaces(g *game.Game, permanent *game.Permanent) []*game.CardDef {
	faceUp := *permanent
	faceUp.FaceDown = false
	faceUp.FaceDownKind = game.FaceDownNone
	faceUp.FaceDownCharacteristics = opt.V[game.FaceDownCharacteristics]{}
	def, ok := permanentCopyDef(g, &faceUp)
	if !ok {
		return nil
	}
	return []*game.CardDef{def}
}

func (e *Engine) chooseFaceDownCost(g *game.Game, playerID game.PlayerID, face *game.CardDef, kind game.FaceDownKind, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (cost.Mana, bool) {
	return e.chooseFaceDownCostOptions(g, playerID, payableFaceDownCosts(g, playerID, face, kind), agents, log)
}

func (e *Engine) chooseFaceDownCostOptions(g *game.Game, playerID game.PlayerID, costs []cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (cost.Mana, bool) {
	if len(costs) == 0 {
		return nil, false
	}
	if len(costs) == 1 {
		return costs[0], true
	}
	options := make([]game.ChoiceOption, 0, len(costs))
	for i, manaCost := range costs {
		options = append(options, game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("Pay %s", manaCost),
		})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose turn face-up cost",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(costs) {
		return costs[selected[0]], true
	}
	return costs[0], true
}

func emitMergedTurnFaceUpRevealEvents(g *game.Game, permanent *game.Permanent) {
	for _, component := range permanent.MergedCards {
		if component.FaceDown {
			continue
		}
		emitEvent(g, game.Event{
			Kind:       game.EventCardRevealed,
			Controller: effectiveController(g, permanent),
			Player:     component.Owner,
			CardID:     component.CardInstanceID,
			Face:       component.Face,
			TokenName:  permanentTokenDefName(component.TokenDef),
			TokenDef:   component.TokenDef,
		})
	}
}

func emitFaceDownRevealEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, game.Event{
		Kind:        game.EventCardRevealed,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.FaceDownFace,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
	})
}
