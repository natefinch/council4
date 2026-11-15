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

// spellIsArcane reports whether a spell has the Arcane subtype (CR 205.3k), the
// only spells onto which "Splice onto Arcane" cards may be spliced (CR 702.47a).
func spellIsArcane(spellDef *game.CardDef) bool {
	return spellDef != nil && spellDef.HasSubtype(types.Arcane)
}

// spliceCandidate is one card in hand eligible to be spliced onto the Arcane
// spell being cast: it has a mana splice cost the caster can afford (on top of
// the host spell and any already-chosen splices), a spell ability, and at least
// one legal target combination for that ability.
type spliceCandidate struct {
	cardID     id.ID
	def        *game.CardDef
	content    game.AbilityContent
	spliceCost cost.Mana
}

// spliceCastResult is the outcome of the splice offer: the spliced spell
// contents, their announced targets, their per-spec target counts, and the mana
// splice costs paid as additional costs, all indexed in parallel in the order the
// caster chose. A zero value (all-nil) means nothing was spliced.
type spliceCastResult struct {
	cards        []spliceCandidate
	contents     []game.AbilityContent
	targets      [][]game.Target
	targetCounts [][]int
	manaCosts    []cost.Mana
}

// Spliced text has the host's characteristics, not the spliced card's colors
// or name (CR 702.47c). This definition is only for announcing its targets.
func spliceTargetDef(host, spliced *game.CardDef) *game.CardDef {
	def := *host
	def.SpellAbility = spliced.SpellAbility
	return &def
}

func (e *Engine) completeSplicedPayment(g *game.Game, player game.PlayerID, host *game.CardDef, cards []spliceCandidate, request payment.SpellRequest, partial payment.Partial) (payment.Plan, bool) {
	if len(cards) == 0 {
		plan, ok, _ := e.PaymentPlanner(g).Complete(payment.ForSpell(request), partial)
		return plan, ok
	}
	def := spliceTargetDef(host, cards[0].def)
	specs := spellTargetSpecs(def, nil, game.CastBranch{})
	var plan payment.Plan
	found := false
	visitTargetChoicesForSpec(g, player, def, 0, game.Event{}, specs, 0, nil, nil,
		&[][]game.Target{}, &[][]int{}, targetChoiceOptions{external: true, visit: func(targets []game.Target, _ []int) bool {
			next := request
			next.Targets = append(slices.Clone(request.Targets), targets...)
			plan, found = e.completeSplicedPayment(g, player, host, cards[1:], next, partial)
			return !found
		}})
	return plan, found
}

func (e *Engine) choosePrioritySplice(g *game.Game, player game.PlayerID, hostID id.ID, request payment.SpellRequest, partial payment.Partial, _ spellPaymentCompletion, agents [game.NumPlayers]PlayerAgent, log *TurnLog) spliceCastResult {
	var result spliceCastResult
	for {
		var candidates []spliceCandidate
		for _, cardID := range g.Players[player].Hand.All() {
			if cardID == hostID || slices.ContainsFunc(result.cards, func(c spliceCandidate) bool { return c.cardID == cardID }) {
				continue
			}
			card, _ := g.GetCardInstance(cardID)
			def := cardFaceOrDefault(card, game.FaceFront)
			spliceCost, ok := def.SpliceCost()
			if !ok {
				continue
			}
			content, ok := firstSpellAbility(def)
			if !ok {
				continue
			}
			candidate := spliceCandidate{cardID: cardID, def: def, content: *content, spliceCost: slices.Clone(spliceCost)}
			cards := append(slices.Clone(result.cards), candidate)
			next := request
			next.SpliceManaCosts = append(slices.Clone(result.manaCosts), candidate.spliceCost)
			cast := action.CastSpellAction{XValue: request.XValue, ChosenModes: request.ChosenModes, KickerPaid: request.KickerPaid, KickerCount: request.KickerCount, Bargained: request.Bargained, Bestowed: request.Bestowed, Offspring: request.Offspring}
			specs := prioritySpellTargetSpecs(request.Card, cast.ChosenModes, castBranchForCast(cast), effectiveKickerCount(cast.KickerPaid, cast.KickerCount))
			found := false
			visitTargetChoicesForSpec(g, player, request.Card, 0, game.Event{}, specs, 0, nil, nil,
				&[][]game.Target{}, &[][]int{}, targetChoiceOptions{xValue: opt.Val(request.XValue), external: true, visit: func(targets []game.Target, _ []int) bool {
					next.Targets = targets
					_, found = e.completeSplicedPayment(g, player, request.Card, cards, next, partial)
					return !found
				}})
			if found {
				candidates = append(candidates, candidate)
			}
		}
		if len(candidates) == 0 {
			break
		}
		cardID, ok := e.chooseSpliceCard(g, player, candidates, agents, log)
		if !ok {
			break
		}
		selected, _ := spliceCandidateByID(candidates, cardID)
		result.cards = append(result.cards, selected)
		result.contents = append(result.contents, selected.content)
		result.manaCosts = append(result.manaCosts, selected.spliceCost)
	}
	// No target decision or observation intervenes between these disclosures:
	// all selected cards are revealed together at CR 601.2b / 702.47b.
	for _, card := range result.cards {
		emitEvent(g, game.Event{Kind: game.EventCardRevealed, Announcement: true, Player: player, Controller: player, CardID: card.cardID, FromZone: zone.Hand})
	}
	return result
}

func (e *Engine) completePrioritySpliceTargets(g *game.Game, player game.PlayerID, host *game.CardDef, cast action.CastSpellAction, splice *spliceCastResult, request payment.SpellRequest, partial payment.Partial, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	fixed := slices.Clone(cast.Targets)
	for index, card := range splice.cards {
		def := spliceTargetDef(host, card.def)
		complete := func(request payment.SpellRequest, partial payment.Partial) (payment.Plan, bool) {
			request.Targets = append(slices.Clone(fixed), request.Targets...)
			return e.completeSplicedPayment(g, player, host, splice.cards[index+1:], request, partial)
		}
		targets, ok := e.chooseCostedSpellTargets(g, player, def, action.CastSpellAction{XValue: cast.XValue}, request, partial, agents, log, complete)
		if !ok {
			return false
		}
		counts, ok := spellTargetCounts(g, player, def, nil, targets, game.CastBranch{})
		if !ok {
			return false
		}
		splice.targets = append(splice.targets, targets)
		splice.targetCounts = append(splice.targetCounts, counts)
		fixed = append(fixed, targets...)
	}
	return true
}

// spliceCandidateByID returns the candidate with the given card ID.
func spliceCandidateByID(candidates []spliceCandidate, cardID id.ID) (spliceCandidate, bool) {
	for _, cand := range candidates {
		if cand.cardID == cardID {
			return cand, true
		}
	}
	return spliceCandidate{}, false
}

// chooseSpliceCard asks the caster which eligible card to splice next, or to
// decline. The request offers no default and permits an empty selection
// (MinChoices 0), so an agent that does not answer declines to splice — keeping
// existing simulations unchanged whenever no scripted decision is supplied.
func (e *Engine) chooseSpliceCard(
	g *game.Game,
	playerID game.PlayerID,
	candidates []spliceCandidate,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
) (id.ID, bool) {
	options := make([]game.ChoiceOption, len(candidates))
	for i, cand := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("Splice %s onto this Arcane spell", cand.def.Name),
			Card:  cardChoiceInfo(g, cand.cardID),
		}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:       game.ChoiceZoneSelection,
		Player:     playerID,
		Prompt:     "Choose a card to splice onto this Arcane spell, or decline.",
		Options:    options,
		MinChoices: 0,
		MaxChoices: 1,
	}, log)
	if len(selected) != 1 {
		return 0, false
	}
	index := selected[0]
	if index < 0 || index >= len(candidates) {
		return 0, false
	}
	return candidates[index].cardID, true
}
