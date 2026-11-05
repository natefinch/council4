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
// Each later attempt offers a proposal only with the choices of targets it
// isn't withheld with (see proposalTargets), whether the legal choices are few
// enough to list or only a sample of them is, so each other target stays open
// with the payment revised for it. A selection that begins only withheld
// choices is replaced by one that begins an open choice, found by searching
// every legal choice when needed, so a search that stops early never withholds
// the choices it didn't reach. A proposal is withheld with every choice of
// targets only once none of them remains open with it, or once an attempt
// whose choices couldn't be filtered repeats withheld ones (see
// repeatWithheld).
//
// A cast none of whose proposals remain is exhausted and no longer offered.
// Each cast offers finitely many proposals and choices of targets, and each
// repeated attempt withholds one of them or exhausts its cast, so a
// deterministic caster always moves on. Activations reversed the same way are
// recorded too (see reverseActivationAttempt).
type castAttempts struct {
	reversed  map[string]bool
	withheld  map[string]bool
	exhausted map[string]bool
	// unpayable are the proposals reversed because they couldn't be paid
	// with their targets; each is withheld with every choice of targets it
	// can't be paid with.
	unpayable map[string]bool
	// closedPrefixes are the targets chosen for a proposal's first target
	// specs that no open choice of targets for it begins with.
	closedPrefixes map[string]bool
	// withheldProposals are the proposals withheld with every choice of
	// targets.
	withheldProposals map[string]bool
}

// withholding reports whether a has withheld anything.
func (a *castAttempts) withholding() bool {
	return a != nil && (len(a.withheld) > 0 || len(a.unpayable) > 0 ||
		len(a.closedPrefixes) > 0 || len(a.withheldProposals) > 0)
}

// castActionKey identifies the cast an attempt redoes. A cast's targets are
// chosen after its cost is proposed (CR 601.2b-c), so any targets the action
// names are only its default choice and not part of the cast.
func castActionKey(cast action.CastSpellAction) string {
	cast.Targets = nil
	return fmt.Sprintf("%+v", cast)
}

// actionExhausted reports whether act is a cast, including a face-down cast,
// none of whose proposals remain, or an activation or special action already
// reversed and repeated (see reverseActivationAttempt).
func (a *castAttempts) actionExhausted(act action.Action) bool {
	if a == nil || len(a.exhausted) == 0 {
		return false
	}
	if faceDown, ok := act.CastFaceDownPayload(); ok {
		return a.exhausted[faceDownCastKey(faceDown)]
	}
	if activate, ok := act.ActivateAbilityPayload(); ok {
		return a.exhausted[activationAttemptKey(activate)]
	}
	if isManaCostSpecialAction(act) {
		return a.exhausted[specialActionAttemptKey(act)]
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
	// sampled reports that targetChoices are a sample of the legal choices
	// rather than every one of them.
	sampled bool
	// payableWith reports whether a proposal through a route can be paid
	// with a choice of targets (see payableTargetChoices), or is nil when
	// the spell has no targets chosen after its proposal.
	payableWith func(route int, proposal spellCostProposal, targets []game.Target) bool
	// payable caches payableWith by key.
	payable map[string]bool
}

// newCastProposals returns the proposals attempts leaves open for cast, whose
// targets are chosen among targetChoices after its cost is proposed.
func newCastProposals(attempts *castAttempts, cast string, targetChoices [][]game.Target) castProposals {
	return castProposals{attempts: attempts, cast: cast, targetChoices: targetChoices, payable: map[string]bool{}}
}

// offerTargets records the choices of targets enumerated for the spell after
// its proposal (CR 601.2c).
func (p *castProposals) offerTargets(result targetChoiceResult) {
	p.targetChoices, p.sampled = result.choices, result.sampled
}

// checkPayment records how to check whether a proposal through a route can be
// paid with a choice of targets, given the request each route proposes with.
func (e *Engine) checkPayment(g *game.Game, p *castProposals, requests []payment.SpellRequest) {
	p.payableWith = func(route int, proposal spellCostProposal, targets []game.Target) bool {
		request := requests[route]
		request.Targets = targets
		return e.announcedSpellCostFeasibility(g, request, proposal.optionIndex, true)(proposal.phyrexian)
	}
}

func (p castProposals) proposalKey(route int, proposal spellCostProposal) string {
	return fmt.Sprintf("%s|%d|%+v", p.cast, route, proposal)
}

func (p castProposals) key(route int, proposal spellCostProposal, targets []game.Target) string {
	return fmt.Sprintf("%s|%v", p.proposalKey(route, proposal), targets)
}

// withheld reports whether proposal through route is withheld with targets.
func (p castProposals) withheld(route int, proposal spellCostProposal, targets []game.Target) bool {
	a := p.attempts
	if a == nil {
		return false
	}
	proposalKey := p.proposalKey(route, proposal)
	if a.withheldProposals[proposalKey] || a.withheld[p.key(route, proposal, targets)] {
		return true
	}
	return a.unpayable[proposalKey] && targets != nil && !p.payableWithTargets(route, proposal, targets)
}

// payableWithTargets reports whether proposal through route can be paid with
// targets, now or after activating payment-time mana abilities. It reports
// true when that can't be checked.
func (p castProposals) payableWithTargets(route int, proposal spellCostProposal, targets []game.Target) bool {
	if p.payableWith == nil {
		return true
	}
	key := p.key(route, proposal, targets)
	payable, ok := p.payable[key]
	if !ok {
		payable = p.payableWith(route, proposal, targets)
		if p.payable != nil {
			p.payable[key] = payable
		}
	}
	return payable
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

// proposalWithheld reports whether proposal through route is withheld with
// every choice of targets the spell may still be given.
func (p castProposals) proposalWithheld(route int, proposal spellCostProposal) bool {
	if !p.attempts.withholding() {
		return false
	}
	if p.attempts.withheldProposals[p.proposalKey(route, proposal)] {
		return true
	}
	if len(p.targetChoices) == 0 {
		return p.withheld(route, proposal, nil)
	}
	if p.sampled {
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
	if !p.attempts.withholding() {
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
	if !p.attempts.withholding() || len(p.targetChoices) == 0 || p.sampled {
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

// reverse records an attempt reversed because a later choice, such as how
// an additional cost was paid, couldn't be completed (CR 733.1). attempt
// identifies the attempt's choices; it made proposal through route with
// targets. Once an attempt repeats, its proposal and targets are withheld; if
// they already were, so is its proposal with every choice of targets (see
// repeatWithheld).
func (p castProposals) reverse(attempt string, route int, proposal spellCostProposal, targets []game.Target) actionOutcome {
	a := p.attempts
	if a == nil {
		return actionReversed
	}
	if a.reversed[attempt] {
		if p.withheld(route, proposal, targets) {
			return p.repeatWithheld(route, proposal)
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
// those targets and with every other choice of targets it can't be paid with
// either (see payableWithTargets). If they already were, so is its proposal
// with every choice of targets (see repeatWithheld).
func (p castProposals) reverseUnpayable(attempt string, route int, proposal spellCostProposal, targets []game.Target) actionOutcome {
	a := p.attempts
	if a == nil {
		return actionReversed
	}
	if p.withheld(route, proposal, targets) {
		return p.repeatWithheld(route, proposal)
	}
	p.withhold(route, proposal, targets)
	if targets != nil && p.payableWith != nil {
		if a.unpayable == nil {
			a.unpayable = map[string]bool{}
		}
		a.unpayable[p.proposalKey(route, proposal)] = true
	}
	return p.reverse(attempt, route, proposal, targets)
}

// repeatWithheld reverses an attempt that made proposal through route with
// targets it was already withheld with. Withheld choices of targets are left
// out of later attempts' prompts or replaced (see proposalTargets), so an
// attempt repeats them only when its choices weren't filtered, such as more
// Phyrexian symbols than its announcements filter. Only that proposal is
// withheld then, with every choice of targets; each other proposal stays open
// for the caster to redo the cast with (CR 733.2). A proposal already withheld
// that way exhausts the cast.
func (p castProposals) repeatWithheld(route int, proposal spellCostProposal) actionOutcome {
	a := p.attempts
	key := p.proposalKey(route, proposal)
	if a.withheldProposals[key] {
		return p.exhaust()
	}
	if a.withheldProposals == nil {
		a.withheldProposals = map[string]bool{}
	}
	a.withheldProposals[key] = true
	return actionReversed
}

// proposalTargets are the choices of targets still open with one proposal
// through one route (CR 601.2c, 733.2). A nil *proposalTargets withholds
// nothing.
type proposalTargets struct {
	proposals castProposals
	route     int
	proposal  spellCostProposal
}

// targetsFor returns the choices of targets still open with proposal through
// route, or nil when nothing is withheld.
func (p castProposals) targetsFor(route int, proposal spellCostProposal) *proposalTargets {
	if !p.attempts.withholding() {
		return nil
	}
	return &proposalTargets{proposals: p, route: route, proposal: proposal}
}

func (t *proposalTargets) prefixKey(prefix []game.Target, counts []int) string {
	return fmt.Sprintf("%s|%v|%v", t.proposals.proposalKey(t.route, t.proposal), prefix, counts)
}

// closed reports whether the complete choice targets, split among the target
// specs by counts, is withheld, including because no open choice begins with
// its targets for the first specs.
func (t *proposalTargets) closed(targets []game.Target, counts []int) bool {
	if t == nil {
		return false
	}
	offset := 0
	for spec, count := range counts {
		offset += count
		if offset <= len(targets) && t.prefixClosed(targets[:offset], counts[:spec+1]) {
			return true
		}
	}
	return t.proposals.withheld(t.route, t.proposal, targets)
}

// prefixClosed reports whether no open choice of targets begins with prefix,
// the targets chosen for the first len(counts) target specs.
func (t *proposalTargets) prefixClosed(prefix []game.Target, counts []int) bool {
	if t == nil {
		return false
	}
	a := t.proposals.attempts
	if len(counts) == 0 {
		return a.withheldProposals[t.proposals.proposalKey(t.route, t.proposal)]
	}
	return a.closedPrefixes[t.prefixKey(prefix, counts)]
}

// closePrefix records that no open choice of targets begins with prefix, the
// targets chosen for the first len(counts) target specs. With no targets
// chosen yet, the proposal is withheld with every choice of targets.
func (t *proposalTargets) closePrefix(prefix []game.Target, counts []int) {
	a := t.proposals.attempts
	if len(counts) == 0 {
		if a.withheldProposals == nil {
			a.withheldProposals = map[string]bool{}
		}
		a.withheldProposals[t.proposals.proposalKey(t.route, t.proposal)] = true
		return
	}
	if a.closedPrefixes == nil {
		a.closedPrefixes = map[string]bool{}
	}
	a.closedPrefixes[t.prefixKey(prefix, counts)] = true
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
