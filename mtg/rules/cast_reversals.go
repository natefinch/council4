package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// castAttempts records the casts reversed while a player may still redo them:
// while one player holds priority, or while a resolving effect lets them cast a
// spell without paying its mana cost. That player may redo a reversed action in
// a legal way or take another (CR 733.2), so a reversal withdraws only the
// proposals it shows can't be completed, never the cast itself while any other
// proposal for it remains:
//
//   - A proposal whose announced cost can't be paid with its targets, even
//     after activating mana abilities, is withheld as soon as it is reversed,
//     together with every other choice of targets it can't be paid with.
//   - Any other reversed proposal failed because of a later choice, such as
//     how an additional cost was paid; it is withheld only once an attempt
//     repeats it with the same choices, since repeating those would only be
//     reversed again.
//
// A cast none of whose proposals remain is exhausted and no longer offered.
// Each cast offers finitely many proposals, and each repeated attempt withholds
// one or exhausts its cast, so a deterministic caster always moves on.
type castAttempts struct {
	reversed  map[string]bool
	withheld  map[string]bool
	exhausted map[string]bool
}

// castActionKey identifies the cast an attempt redoes. A cast's targets are
// chosen after its cost is proposed (CR 601.2b-c), so any targets the action
// names are only its default choice and not part of the cast.
func castActionKey(cast action.CastSpellAction) string {
	cast.Targets = nil
	return fmt.Sprintf("%+v", cast)
}

// castExhausted reports whether act is a cast none of whose proposals remain.
func (a *castAttempts) castExhausted(act action.Action) bool {
	if a == nil || len(a.exhausted) == 0 {
		return false
	}
	cast, ok := act.CastSpellPayload()
	return ok && a.exhausted[castActionKey(cast)]
}

// castProposals are the cost proposals (CR 601.2b) and choices of targets
// (CR 601.2c) still open for one attempt to cast a spell. The zero value
// withholds nothing.
type castProposals struct {
	attempts *castAttempts
	cast     string
	// targetChoices are the choices of targets chosen after the proposal, nil
	// when the spell has none or they were fixed before it.
	targetChoices [][]game.Target
}

// newCastProposals returns the proposals attempts leaves open for cast, whose
// targets are chosen among targetChoices after its cost is proposed.
func newCastProposals(attempts *castAttempts, cast string, targetChoices [][]game.Target) castProposals {
	return castProposals{attempts: attempts, cast: cast, targetChoices: targetChoices}
}

func (p castProposals) key(route int, proposal spellCostProposal, targets []game.Target) string {
	return fmt.Sprintf("%s|%d|%+v|%v", p.cast, route, proposal, targets)
}

func (p castProposals) withheld(route int, proposal spellCostProposal, targets []game.Target) bool {
	return p.attempts != nil && p.attempts.withheld[p.key(route, proposal, targets)]
}

func (p castProposals) withhold(route int, proposal spellCostProposal, targets []game.Target) {
	if p.attempts == nil {
		return
	}
	if p.attempts.withheld == nil {
		p.attempts.withheld = map[string]bool{}
	}
	p.attempts.withheld[p.key(route, proposal, targets)] = true
}

// sampled reports whether targetChoices may be a sample rather than every
// legal choice of targets (see payableTargetChoices).
func (p castProposals) sampled() bool {
	return len(p.targetChoices) >= maxSampledNumericTargetChoices
}

// proposalWithheld reports whether proposal through route is withheld with
// every choice of targets the spell may still be given.
func (p castProposals) proposalWithheld(route int, proposal spellCostProposal) bool {
	if p.attempts == nil || len(p.attempts.withheld) == 0 {
		return false
	}
	if len(p.targetChoices) == 0 {
		return p.withheld(route, proposal, nil)
	}
	if p.sampled() {
		return false
	}
	for _, targets := range p.targetChoices {
		if !p.withheld(route, proposal, targets) {
			return false
		}
	}
	return true
}

// announcementsWithheld returns a check of whether announcing the given
// Phyrexian payments for cost option optionIndex through route is withheld, or
// nil when nothing is withheld.
func (p castProposals) announcementsWithheld(route, optionIndex int) func([]phyrexianAnnouncement) bool {
	if p.attempts == nil || len(p.attempts.withheld) == 0 {
		return nil
	}
	return func(announcements []phyrexianAnnouncement) bool {
		return p.proposalWithheld(route, spellCostProposal{optionIndex: optionIndex, phyrexian: announcements})
	}
}

// optionOpen reports whether some way to announce the Phyrexian payments of
// option through route, as playerID may announce them, remains open.
func (p castProposals) optionOpen(g *game.Game, playerID game.PlayerID, route int, option payment.SpellOptionSummary) bool {
	withheld := p.announcementsWithheld(route, option.Index)
	return withheld == nil ||
		announcementsOpen(phyrexianSymbols(option.ManaCost), playerLife(g, playerID), nil, withheld)
}

// openTargetChoices returns the choices of targets to offer with proposal
// through route: those of payable (see payableTargetChoices) that aren't
// withheld with it, or, when there are none, every choice the spell may still
// be given that isn't. It returns payable unchanged when nothing is withheld
// with proposal or targetChoices may be a sample.
func (p castProposals) openTargetChoices(route int, proposal spellCostProposal, payable [][]game.Target) [][]game.Target {
	if p.attempts == nil || len(p.attempts.withheld) == 0 || len(p.targetChoices) == 0 || p.sampled() {
		return payable
	}
	open := func(choices [][]game.Target) [][]game.Target {
		return slices.DeleteFunc(slices.Clone(choices), func(targets []game.Target) bool {
			return p.withheld(route, proposal, targets)
		})
	}
	if !slices.ContainsFunc(p.targetChoices, func(targets []game.Target) bool { return p.withheld(route, proposal, targets) }) {
		return payable
	}
	if choices := open(payable); len(choices) > 0 {
		return choices
	}
	if choices := open(p.targetChoices); len(choices) > 0 {
		return choices
	}
	return payable
}

// unpayableTargetChoices returns the choices of targets in targetChoices with
// which proposal can't be paid through request, given that it can't be paid
// with request's own targets (see payableTargetChoices).
func (e *Engine) unpayableTargetChoices(g *game.Game, request payment.SpellRequest, targetChoices [][]game.Target, proposal spellCostProposal) [][]game.Target {
	if len(targetChoices) == 0 {
		return nil
	}
	var unpayable [][]game.Target
	for _, variant := range targetCostVariants(request, targetChoices) {
		if !e.announcedSpellCostFeasibility(g, variant, proposal.optionIndex, false)(proposal.phyrexian) {
			unpayable = append(unpayable, variant.Targets)
		}
	}
	return unpayable
}

// reverse records an attempt reversed because a later choice, such as how
// an additional cost was paid, couldn't be completed (CR 733.1). attempt
// identifies the attempt's choices; it made proposal through route with
// targets. Once an attempt repeats, its proposal and targets are withheld; if
// they already were, the cast is exhausted.
func (p castProposals) reverse(attempt string, route int, proposal spellCostProposal, targets []game.Target) actionOutcome {
	a := p.attempts
	if a == nil {
		return actionReversed
	}
	if a.reversed[attempt] {
		if p.withheld(route, proposal, targets) {
			return p.exhaust()
		}
		p.withhold(route, proposal, targets)
	}
	if a.reversed == nil {
		a.reversed = map[string]bool{}
	}
	a.reversed[attempt] = true
	return actionReversed
}

// reverseUnpayable records an attempt reversed because proposal through route
// can't be paid with targets (CR 733.1). Its proposal is withheld at once with
// those targets and with each other choice of targets in unpayable, those it
// can't be paid with either.
func (p castProposals) reverseUnpayable(attempt string, route int, proposal spellCostProposal, targets []game.Target, unpayable [][]game.Target) actionOutcome {
	if p.attempts == nil {
		return actionReversed
	}
	if p.withheld(route, proposal, targets) {
		return p.exhaust()
	}
	p.withhold(route, proposal, targets)
	for _, choice := range unpayable {
		p.withhold(route, proposal, choice)
	}
	return p.reverse(attempt, route, proposal, targets)
}

// exhaust records that no proposal for the cast remains and reverses it.
func (p castProposals) exhaust() actionOutcome {
	if a := p.attempts; a != nil {
		if a.exhausted == nil {
			a.exhausted = map[string]bool{}
		}
		a.exhausted[p.cast] = true
	}
	return actionReversed
}

// exhausted reports whether no proposal for the cast remains.
func (p castProposals) exhausted() bool {
	return p.attempts != nil && p.attempts.exhausted[p.cast]
}

// phyrexianSymbols returns the Phyrexian symbols of manaCost in cost order.
func phyrexianSymbols(manaCost *cost.Mana) []cost.Symbol {
	if manaCost == nil {
		return nil
	}
	var symbols []cost.Symbol
	for _, symbol := range *manaCost {
		if isPhyrexianSymbol(symbol) {
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

// announcementsOpen reports whether some way to announce the rest of symbols
// after announced, paying 2 life for a symbol only while life remains
// available (see announcePhyrexianPayments), isn't withheld. With more symbols
// than phyrexianPaymentCandidates enumerates, every announcement stays open.
func announcementsOpen(symbols []cost.Symbol, availableLife int, announced []phyrexianAnnouncement, withheld func([]phyrexianAnnouncement) bool) bool {
	if withheld == nil || len(symbols) > maxOpenPhyrexianCandidates {
		return true
	}
	if len(announced) == len(symbols) {
		return !withheld(announced)
	}
	symbol := symbols[len(announced)]
	if announcementsOpen(symbols, availableLife, append(slices.Clone(announced), phyrexianAnnouncement{symbol: symbol}), withheld) {
		return true
	}
	return availableLife >= 2 && announcementsOpen(symbols, availableLife-2,
		append(slices.Clone(announced), phyrexianAnnouncement{symbol: symbol, payLife: true}), withheld)
}
