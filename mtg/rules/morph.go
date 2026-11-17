package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
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

// faceDownCostSource identifies the rule that offers a turn-face-up cost.
// Equal mana costs from different sources stay distinct choices because
// paying them has different consequences (CR 702.37b, 701.40c, 701.58c).
type faceDownCostSource uint8

const (
	// faceDownCostManaCost is a manifested or cloaked creature card's mana
	// cost (CR 701.40b, 701.58b).
	faceDownCostManaCost faceDownCostSource = iota + 1
	faceDownCostMorph
	faceDownCostMegamorph
	faceDownCostDisguise
)

func (source faceDownCostSource) String() string {
	switch source {
	case faceDownCostManaCost:
		return "mana cost"
	case faceDownCostMorph:
		return "morph"
	case faceDownCostMegamorph:
		return "megamorph"
	case faceDownCostDisguise:
		return "disguise"
	}
	return "unknown"
}

// faceDownCostOption is one turn-face-up cost together with the rule that
// offers it, so the turn-face-up commit knows which cost was paid.
type faceDownCostOption struct {
	cost   cost.Mana
	source faceDownCostSource
}

func (option faceDownCostOption) label() string {
	return fmt.Sprintf("Pay %s (%s)", option.cost, option.source)
}

func faceDownCostForCard(card *game.CardDef, kind game.FaceDownKind) (cost.Mana, bool) {
	costs := faceDownCostsForCard(card, kind)
	if len(costs) == 0 {
		return nil, false
	}
	return costs[0].cost, true
}

func faceDownCostsForCard(card *game.CardDef, kind game.FaceDownKind) []faceDownCostOption {
	var costs []faceDownCostOption
	if kind == game.FaceDownManifest || kind == game.FaceDownCloak {
		if card.HasType(types.Creature) && card.ManaCost.Exists {
			costs = append(costs, faceDownCostOption{cost: card.ManaCost.Val, source: faceDownCostManaCost})
		}
	}
	morphKind := kind == game.FaceDownMorph || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect
	disguiseKind := kind == game.FaceDownDisguise || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect
	add := func(body game.Ability) {
		if ka, ok := game.BodyKeywordAbility(body, game.Morph); ok && morphKind {
			if morph, ok := ka.(game.MorphKeyword); ok {
				source := faceDownCostMorph
				if morph.Megamorph {
					source = faceDownCostMegamorph
				}
				costs = append(costs, faceDownCostOption{cost: morph.Cost, source: source})
			}
		}
		if ka, ok := game.BodyKeywordAbility(body, game.Disguise); ok && disguiseKind {
			if disguise, ok := ka.(game.DisguiseKeyword); ok {
				costs = append(costs, faceDownCostOption{cost: disguise.Cost, source: faceDownCostDisguise})
			}
		}
	}
	for i := range card.ActivatedAbilities {
		add(&card.ActivatedAbilities[i])
	}
	for i := range card.StaticAbilities {
		add(&card.StaticAbilities[i])
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
				cast := action.CastFaceDownAction{CardID: cardID, SourceZone: zone.Graveyard, Face: face, FaceDownKind: kind}
				if e.canCastFaceDownFromZone(g, playerID, cast, zone.Graveyard) {
					actions = append(actions, action.CastFaceDownFromZone(cardID, zone.Graveyard, face, kind))
				}
				for _, permission := range permanentTypePlayPermissions(g, playerID, zone.Graveyard, faceDownSpellDef()) {
					if e.canCastGraveyardFaceDown(g, playerID, cardID, face, kind, permission) {
						actions = append(actions, action.CastFaceDownFromZone(cardID, zone.Graveyard, face, kind).WithPlayPermission(permission))
					}
				}
			}
		}
	}
	for _, from := range []zone.Type{zone.Command, zone.Exile, zone.Library} {
		source := castSourceZone(player, from)
		cards := source.All()
		if from == zone.Exile {
			cards = append(cards, foreignExileCastableCards(g, playerID)...)
		}
		for _, cardID := range cards {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			for _, face := range legalFaceDownFaces(card.Def) {
				for _, kind := range faceDownKindsForCard(cardFaceOrDefault(card, face)) {
					cast := action.CastFaceDownAction{CardID: cardID, SourceZone: from, Face: face, FaceDownKind: kind}
					if e.canCastFaceDownFromZone(g, playerID, cast, from) {
						actions = append(actions, action.CastFaceDownFromZone(cardID, from, face, kind))
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
		FaceDown:        true,
	}
}

func (e *Engine) canCastGraveyardFaceDown(g *game.Game, player game.PlayerID, cardID id.ID, face game.FaceIndex, kind game.FaceDownKind, permission game.PlayPermission) bool {
	spell := castSelectionFace(g, player, cardID, faceDownSpellDef(), false)
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

// Availability uses the face-down CR 601.2a payment view, never a clone probe.
func (e *Engine) faceDownCostProposable(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, sourceZone zone.Type) bool {
	request := faceDownSpellRequest(player, cast.CardID, sourceZone)
	routes := permanentTypePlayCastRoutes(faceDownPermissionRoutes(g, player, cast, sourceZone), cast.Permission)
	return e.priorityCastCostsAvailable(g, faceDownCastObject(player, cast, sourceZone), mutateCastRouteRequests(request, routes))
}

func faceDownPermissionRoutes(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, from zone.Type) []castRoute {
	if from == zone.Hand || from == zone.Command {
		return []castRoute{{}}
	}
	routes, _ := castFromZonePermissionRoutes(g, player, cast.CardID, from, cast.Face, false, false, faceDownSpellDef())
	if from == zone.Graveyard && len(permanentTypePlayPermissions(g, player, from, faceDownSpellDef())) > 0 ||
		from == zone.Exile && game.PhysicalFace(cast.Face) == game.FaceFront && hasAdventureExilePermission(g, player, cast.CardID) {
		routes = append(routes, castRoute{})
	}
	return orderedCastRoutes(routes)
}

func (e *Engine) canCastFaceDownFromZone(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, from zone.Type) bool {
	if from == zone.Hand {
		return cast.Permission == (game.PlayPermission{}) && e.canCastFaceDown(g, player, cast.CardID, cast.Face, cast.FaceDownKind)
	}
	if from == zone.Graveyard && cast.Permission != (game.PlayPermission{}) {
		return e.canCastGraveyardFaceDown(g, player, cast.CardID, cast.Face, cast.FaceDownKind, cast.Permission)
	}
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok || !castSourceContains(castSourcePlayer(g, g.Players[player], cast.CardID, from), cast.CardID, from) ||
		(cast.FaceDownKind != game.FaceDownMorph && cast.FaceDownKind != game.FaceDownDisguise) {
		return false
	}
	def, ok := cardFaceDef(card, cast.Face)
	if !ok {
		return false
	}
	if _, ok := faceDownCostForCard(def, cast.FaceDownKind); !ok {
		return false
	}
	if from == zone.Command {
		if !isCommanderOwnedBy(g, player, card.ID) {
			return false
		}
	} else if from != zone.Exile && from != zone.Library && from != zone.Graveyard {
		return false
	} else if len(faceDownPermissionRoutes(g, player, cast, from)) == 0 {
		return false
	}
	spell := castSelectionFace(g, player, card.ID, faceDownSpellDef(), false)
	return canAct(g, player) && player == g.Turn.PriorityPlayer && !splitSecondOnStack(g) &&
		canCastAtCurrentTiming(g, player, spell) && !castFromZoneProhibited(g, player, from) &&
		!spellCastProhibited(g, player, spell) && !spellCastLimitReached(g, player, spell) &&
		e.faceDownCostProposable(g, player, cast, from)
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
	spell := castSelectionFace(g, playerID, cardID, faceDownSpellDef(), false)
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

// attemptCastFaceDownWithChoices uses the shared procedure. Only an opaque
// payment can reverse a witnessed cast (CR 734.1).
func (e *Engine) attemptCastFaceDownWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastFaceDownAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	defer g.OpenLibraryTopWindow()()
	if cast.SourceZone == zone.Graveyard {
		return e.attemptGraveyardFaceDown(g, playerID, cast, agents, log)
	}
	from := cast.SourceZone
	if from == zone.None {
		from = zone.Hand
	}
	if cast.Permission != (game.PlayPermission{}) {
		return actionFailed
	}
	if !e.canCastFaceDownFromZone(g, playerID, cast, from) {
		return actionFailed
	}
	return e.castFaceDownSpell(g, playerID, cast, from, agents, log)
}

func (e *Engine) attemptGraveyardFaceDown(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	permission := cast.Permission
	if permission == (game.PlayPermission{}) {
		permissions := permanentTypePlayPermissions(g, player, zone.Graveyard, faceDownSpellDef())
		if len(permissions) > 0 {
			permission = permissions[0]
		}
	}
	if permission != (game.PlayPermission{}) && !e.canCastGraveyardFaceDown(g, player, cast.CardID, cast.Face, cast.FaceDownKind, permission) ||
		permission == (game.PlayPermission{}) && !e.canCastFaceDownFromZone(g, player, cast, zone.Graveyard) {
		return actionFailed
	}
	return e.castFaceDownSpell(g, player, cast, zone.Graveyard, agents, log)
}

func faceDownCastObject(player game.PlayerID, cast action.CastFaceDownAction, sourceZone zone.Type) *game.StackObject {
	return &game.StackObject{
		Kind: game.StackSpell, SourceID: cast.CardID, Controller: player, Casting: true,
		Face: cast.Face, FaceDown: true, FaceDownFace: cast.Face, FaceDownKind: cast.FaceDownKind,
		SourceZone: sourceZone,
	}
}

// castFaceDownSpell casts a validated face-down spell: the card moves to the
// stack face down first (CR 601.2a), then the {3} alternative cost is proposed
// and the total cost is determined from the face-down characteristics and
// locked (CR 702.37c, 601.2b, 601.2f). If that total includes a mana payment,
// the caster may then activate mana abilities (CR 601.2g) before paying it
// (CR 601.2h), and only then does the spell become cast (CR 601.2i). An unpaid
// opaque proposal is undone along with its reversible mana activations (CR 734.1).
func (e *Engine) castFaceDownSpell(g *game.Game, player game.PlayerID, cast action.CastFaceDownAction, sourceZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	obj := faceDownCastObject(player, cast, sourceZone)
	request := faceDownSpellRequest(player, cast.CardID, sourceZone)
	sourceRoutes := faceDownPermissionRoutes(g, player, cast, sourceZone)
	independentRoutes, _ := castFromZonePermissionRoutes(g, player, cast.CardID, sourceZone, cast.Face, false, false, faceDownSpellDef())
	routes := permanentTypePlayCastRoutes(sourceRoutes, cast.Permission)
	requests := mutateCastRouteRequests(request, routes)
	route := 0
	descriptor := costedAction{
		object: obj, usesStack: true,
		spellSource: &costedSpellSource{player: player, card: cast.CardID, zone: sourceZone},
		available:   func(g *game.Game) bool { return e.priorityCastCostsAvailable(g, obj, requests) },
		cost:        activationCost{request: payment.ForSpell(request), present: func(g *game.Game) bool { return costedSpellPresent(g, obj) }},
		proposal: func(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.Request, payment.Partial, bool) {
			if sourceZone == zone.Graveyard && (cast.Permission != (game.PlayPermission{}) ||
				len(permanentTypePlayPermissions(g, player, sourceZone, faceDownSpellDef())) > 0) {
				var ok bool
				cast.Permission, ok = e.selectPermanentTypePlay(g, player, cast.Permission, sourceZone, faceDownSpellDef(), len(independentRoutes) > 0, agents, log)
				if !ok {
					return payment.Request{}, payment.Partial{}, false
				}
				routes = permanentTypePlayCastRoutes(sourceRoutes, cast.Permission)
				requests = mutateCastRouteRequests(request, routes)
			}
			complete := func(request payment.SpellRequest, partial payment.Partial) (payment.Plan, bool) {
				plan, ok, _ := e.PaymentPlanner(g).Complete(payment.ForSpell(request), partial)
				return plan, ok
			}
			selected, partial, ok := e.choosePrioritySpellPayment(g, routes, requests, complete, agents, log)
			if !ok {
				return payment.Request{}, partial, false
			}
			route, request = selected, requests[selected]
			return payment.ForSpell(request), partial, true
		},
		commit: func(g *game.Game, paid payment.ApplyResult, _ [game.NumPlayers]PlayerAgent, _ *TurnLog, _ int) {
			result := paid.Spell
			obj.Casting = false
			recordPermanentTypePlay(g, player, cast.Permission)
			if sourceZone == zone.Command {
				recordCommanderCast(g, player, cast.CardID)
			}
			recordExilePlayPermissionUse(g, routes[route].permission)
			obj.CastPermissionEntryCounters = routes[route].castPermissionEntryCounters(result.CastPermission)
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
			resolveSpellCastManaSpendRiders(g, player, paid.ManaBefore, result.PoolSpend, request.Card, obj)
		},
	}
	return e.costedAction(g, player, descriptor, agents, log)
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

// canTurnFaceUp reports whether the player may turn the face-down permanent
// permanentID face up (CR 116.2b, 702.37e, 701.40b, 701.58b): they control it
// with priority and one of its turn-face-up costs is payable now or with
// payment-time mana abilities (CR 605.3a).
func (e *Engine) canTurnFaceUp(g *game.Game, playerID game.PlayerID, permanentID id.ID) bool {
	return len(e.feasibleTurnFaceUpCosts(g, playerID, permanentID)) > 0
}

// turnableFaceDown reports whether permanentID is a face-down permanent the
// player controls and may try to turn face up in the game checked.
func turnableFaceDown(g *game.Game, playerID game.PlayerID, permanentID id.ID) (*game.Permanent, bool) {
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok || permanent.PhasedOut || !permanent.FaceDown || effectiveController(g, permanent) != playerID || permanent.FaceDownKind == game.FaceDownNone {
		return nil, false
	}
	return permanent, true
}

// turnFaceUpCost is the activationCost of turning permanentID face up for
// option. Its payment requires the permanent to remain face down under the
// player's control after any payment-time mana abilities.
func turnFaceUpCost(g *game.Game, playerID game.PlayerID, permanentID id.ID, option faceDownCostOption) activationCost {
	permanent, _ := permanentByObjectID(g, permanentID)
	manaCost := option.cost
	return genericRequestCost(payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Source: permanent,
		TurnFaceUp: turnFaceUpPayment(permanent.FaceDownKind, option.source)}, func(g *game.Game) bool {
		_, ok := turnableFaceDown(g, playerID, permanentID)
		return ok
	})
}

// turnFaceUpPayment identifies the turn-face-up cost paid for source on a
// permanent made face down by kind. A mana cost is a manifested creature's
// (CR 701.40b) unless the permanent was cloaked (CR 701.58c).
func turnFaceUpPayment(kind game.FaceDownKind, source faceDownCostSource) payment.TurnFaceUpCost {
	switch source {
	case faceDownCostMorph, faceDownCostMegamorph:
		return payment.TurnFaceUpMorphCost
	case faceDownCostDisguise:
		return payment.TurnFaceUpDisguiseCost
	case faceDownCostManaCost:
		if kind == game.FaceDownManifest {
			return payment.TurnFaceUpManifestedManaCost
		}
		return payment.TurnFaceUpCloakedManaCost
	default:
		panic(fmt.Sprintf("unknown face-down cost source %d", source))
	}
}

// feasibleTurnFaceUpCosts lists the distinct turn-face-up cost options of
// permanentID that the player can pay now or with payment-time mana abilities.
// Options are distinct by both mana cost and source.
func (e *Engine) feasibleTurnFaceUpCosts(g *game.Game, playerID game.PlayerID, permanentID id.ID) []faceDownCostOption {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	permanent, ok := turnableFaceDown(g, playerID, permanentID)
	if !ok {
		return nil
	}
	var costs []faceDownCostOption
	for _, face := range faceDownTurnUpFaces(g, permanent) {
		for _, option := range faceDownCostsForCard(face, permanent.FaceDownKind) {
			duplicate := slices.ContainsFunc(costs, func(prior faceDownCostOption) bool {
				return prior.source == option.source && prior.cost.String() == option.cost.String()
			})
			if !duplicate && e.costedActionAvailable(g, playerID, turnFaceUpCost(g, playerID, permanentID, option)) {
				costs = append(costs, option)
			}
		}
	}
	return costs
}

// applyTurnFaceUpWithChoices turns a face-down permanent face up as a special
// action: the player chooses a turn-face-up cost and pays it, activating mana
// abilities while paying if they choose (CR 605.3a). If the paid cost is a
// megamorph cost, a +1/+1 counter is put on the permanent as it turns face up
// (CR 702.37b, 708.11), before any turned-face-up trigger event.
func (e *Engine) applyTurnFaceUpWithChoices(g *game.Game, playerID game.PlayerID, permanentID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	costs := e.feasibleTurnFaceUpCosts(g, playerID, permanentID)
	option, ok := e.chooseFaceDownCostOptions(g, playerID, costs, agents, log)
	if !ok {
		return actionFailed
	}
	manaCost := option.cost
	c := turnFaceUpCost(g, playerID, permanentID, option)
	return e.costedAction(g, playerID, costedAction{cost: c, disclose: func(g *game.Game) {
		permanent, _ := permanentByObjectID(g, permanentID)
		event := game.Event{Kind: game.EventActionCostAnnounced, Controller: playerID, Player: playerID,
			PermanentID: permanentID, AnnouncedManaCost: manaCost.String()}
		if permanent.FaceDownKind == game.FaceDownManifest || permanent.FaceDownKind == game.FaceDownCloak {
			// A morph/disguise payment on a manifested card shows only that
			// cost; the creature-card payment shows the underlying card.
			if option.source == faceDownCostManaCost {
				emitFaceDownRevealEvent(g, permanent)
				event.CardID, event.Face = permanent.CardInstanceID, permanent.FaceDownFace
			}
		}
		emitEvent(g, event)
	}, commit: func(g *game.Game, _ payment.ApplyResult, _ [game.NumPlayers]PlayerAgent, _ *TurnLog, _ int) {
		permanent, _ := permanentByObjectID(g, permanentID)
		permanent.Face = permanent.FaceDownFace
		permanent.FaceDown = false
		permanent.FaceDownKind = game.FaceDownNone
		permanent.FaceDownCharacteristics = opt.V[game.FaceDownCharacteristics]{}
		emitFaceDownRevealEvent(g, permanent)
		emitMergedTurnFaceUpRevealEvents(g, permanent)
		if option.source == faceDownCostMegamorph {
			addCountersToPermanent(g, permanent, counter.PlusOnePlusOne, 1)
		}
		emitEvent(g, game.Event{
			Kind:        game.EventPermanentTurnedFaceUp,
			Controller:  playerID,
			Player:      permanent.Owner,
			CardID:      permanent.CardInstanceID,
			Face:        permanent.Face,
			PermanentID: permanent.ObjectID,
		})
	}}, agents, log)
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

func (e *Engine) chooseFaceDownCostOptions(g *game.Game, playerID game.PlayerID, costs []faceDownCostOption, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (faceDownCostOption, bool) {
	if len(costs) == 0 {
		return faceDownCostOption{}, false
	}
	if len(costs) == 1 {
		return costs[0], true
	}
	options := make([]game.ChoiceOption, 0, len(costs))
	for i, option := range costs {
		options = append(options, game.ChoiceOption{
			Index: i,
			Label: option.label(),
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
