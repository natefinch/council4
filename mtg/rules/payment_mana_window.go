package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// spellCostProposal is the cost a caster proposes for a spell before choosing
// its targets (CR 601.2b): one of the spell's cost options and how they
// announced paying each Phyrexian symbol in it.
type spellCostProposal struct {
	optionIndex int
	phyrexian   []phyrexianAnnouncement
}

// phyrexianAnnouncement records whether the caster announced paying a
// Phyrexian symbol with 2 life or with mana (CR 601.2b).
type phyrexianAnnouncement struct {
	symbol  cost.Symbol
	payLife bool
}

// proposeSpellCost asks the caster to choose among request's feasible cost
// options, defaulting to the first payable now, then to announce how each
// Phyrexian symbol in the chosen option will be paid (CR 601.2b). When
// targetsPending, request has no targets yet; if no option is feasible without
// them, every option is offered, since a target-dependent reduction may make
// one payable, and lockSpellCostProposal rechecks the total. Only proposals
// open leaves open are offered (see castAttempts); it reports false when there
// are none.
func (e *Engine) proposeSpellCost(g *game.Game, request payment.SpellRequest, targetsPending bool, open castProposals, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (spellCostProposal, bool) {
	planner := paymentOrch.planner(g)
	costs := planner.PrepareSpellCosts(request)
	manual := e.paymentManaWindow(g, request.PlayerID)
	summaries := planner.SpellCostOptions(request)
	closed := func(prepared payment.PreparedSpellCost) bool {
		index := slices.IndexFunc(summaries, func(option payment.SpellOptionSummary) bool { return option.Index == prepared.Summary.Index })
		return index >= 0 && !open.optionOpen(g, request.PlayerID, 0, summaries[index])
	}
	payable := make([]payment.PreparedSpellCost, 0, len(costs))
	fallback := 0
	foundPayable := false
	for _, prepared := range costs {
		if closed(prepared) {
			continue
		}
		canPay, feasible := preparedSpellCostFeasibility(g, request, prepared, manual, nil)
		if feasible {
			if canPay && !foundPayable {
				fallback, foundPayable = len(payable), true
			}
			payable = append(payable, prepared)
		}
	}
	if len(payable) == 0 && targetsPending {
		payable = slices.DeleteFunc(slices.Clone(costs), closed)
	}
	if len(payable) == 0 {
		return spellCostProposal{}, false
	}
	options := make([]game.ChoiceOption, len(payable))
	for index, prepared := range payable {
		options[index] = game.ChoiceOption{Index: index, Label: prepared.Summary.Label}
	}
	selected := []int{fallback}
	if len(options) > 1 {
		selected = e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: request.PlayerID, Prompt: "Choose spell cost before activating mana abilities.",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{fallback},
		}, log)
	}
	proposal := spellCostProposal{optionIndex: payable[selected[0]].Summary.Index}
	for _, option := range summaries {
		if option.Index == proposal.optionIndex {
			proposal.phyrexian = e.announcePhyrexianPayments(g, request.PlayerID, option.ManaCost,
				e.announcedSpellCostFeasibility(g, request, proposal.optionIndex, true),
				open.announcementsWithheld(0, proposal.optionIndex), agents, log)
			break
		}
	}
	return proposal, true
}

// announcedSpellCostFeasibility returns a check of whether request's cost
// option optionIndex can still be paid once the given Phyrexian payments are
// announced; any other Phyrexian symbol may still be paid either way. With
// manaWindow, paying after activating payment-time mana abilities that can
// actually make it payable counts too (CR 601.2g). The check reports true when
// request doesn't offer that option, so it never steers an announcement.
func (e *Engine) announcedSpellCostFeasibility(g *game.Game, request payment.SpellRequest, optionIndex int, manaWindow bool) func([]phyrexianAnnouncement) bool {
	var manual manaWindowCheck
	if manaWindow {
		manual = e.paymentManaWindow(g, request.PlayerID)
	}
	for _, prepared := range paymentOrch.planner(g).PrepareSpellCosts(request) {
		if prepared.Summary.Index != optionIndex {
			continue
		}
		return func(announcements []phyrexianAnnouncement) bool {
			announced := matchPhyrexianAnnouncements(prepared.Summary.ManaCost, announcements)
			_, feasible := preparedSpellCostFeasibility(g, request, prepared, manual, announced)
			return feasible
		}
	}
	return func([]phyrexianAnnouncement) bool { return true }
}

// proposableSpellOptions returns the cost options of request that are payable
// with some choice of targets in targetChoices, or with request's own targets
// when there are none. A cost proposed before the spell's targets are chosen
// (CR 601.2b) can be one whose total is payable only once they are, such as
// one reduced for targeting Elderwood Scion (CR 601.2f). With manual, which is
// nil when the caster has no payment-time mana abilities, an option whose total
// those abilities can make payable is proposable too (CR 601.2g).
func proposableSpellOptions(g *game.Game, request payment.SpellRequest, targetChoices [][]game.Target, manual manaWindowCheck) []payment.SpellOptionSummary {
	planner := paymentOrch.planner(g)
	if len(targetChoices) == 0 && manual == nil {
		return planner.PayableSpellOptions(request)
	}
	payable := map[int]bool{}
	for _, variant := range targetCostVariants(request, targetChoices) {
		for _, option := range planner.PayableSpellOptions(variant) {
			payable[option.Index] = true
		}
		if manual == nil {
			continue
		}
		for _, prepared := range planner.PrepareSpellCosts(variant) {
			if !payable[prepared.Summary.Index] {
				_, payable[prepared.Summary.Index] = preparedSpellCostFeasibility(g, variant, prepared, manual, nil)
			}
		}
	}
	var options []payment.SpellOptionSummary
	for _, option := range planner.SpellCostOptions(request) {
		if payable[option.Index] {
			options = append(options, option)
		}
	}
	return options
}

// proposedSpellCostFeasibility is announcedSpellCostFeasibility for a cost
// proposed before the spell's targets are chosen: the announcements are
// feasible when the option can be paid with them for some choice of targets
// in targetChoices, now or after activating payment-time mana abilities
// (CR 601.2g).
func (e *Engine) proposedSpellCostFeasibility(g *game.Game, request payment.SpellRequest, targetChoices [][]game.Target, optionIndex int) func([]phyrexianAnnouncement) bool {
	variants := targetCostVariants(request, targetChoices)
	checks := make([]func([]phyrexianAnnouncement) bool, len(variants))
	for i, variant := range variants {
		checks[i] = e.announcedSpellCostFeasibility(g, variant, optionIndex, true)
	}
	return func(announcements []phyrexianAnnouncement) bool {
		return slices.ContainsFunc(checks, func(check func([]phyrexianAnnouncement) bool) bool {
			return check(announcements)
		})
	}
}

// payableTargetChoices returns the choices of targets in open's targetChoices with
// which the cost proposed for request can still be paid with the Phyrexian
// payments announced with it, now or after activating payment-time mana
// abilities (CR 601.2b, 601.2f, 601.2g): a legal target such as an
// opponent's Elderwood Scion can raise the total cost beyond what the caster
// can pay, and choosing it would only make the cast illegal (CR 733.1). It
// returns nil, leaving the choice unrestricted, when open's targetChoices are
// a sample rather than every legal choice, or when no choice is payable and
// the cast will be reversed whatever targets are chosen.
func (e *Engine) payableTargetChoices(g *game.Game, request payment.SpellRequest, open castProposals, proposal spellCostProposal) [][]game.Target {
	if open.sampled {
		return nil
	}
	var payable [][]game.Target
	for _, variant := range targetCostVariants(request, open.targetChoices) {
		if e.announcedSpellCostFeasibility(g, variant, proposal.optionIndex, true)(proposal.phyrexian) {
			payable = append(payable, variant.Targets)
		}
	}
	return payable
}

// targetCostVariants returns request once with each choice of targets in
// targetChoices, or request itself when there are none.
func targetCostVariants(request payment.SpellRequest, targetChoices [][]game.Target) []payment.SpellRequest {
	if len(targetChoices) == 0 {
		return []payment.SpellRequest{request}
	}
	variants := make([]payment.SpellRequest, len(targetChoices))
	for i, targets := range targetChoices {
		variants[i] = request
		variants[i].Targets = targets
	}
	return variants
}

// lockSpellCostProposal locks the total cost of the proposed option once the
// spell's targets and other choices are final (CR 601.2f). It fails when that
// total can't be paid now or after activating payment-time mana abilities.
func (e *Engine) lockSpellCostProposal(g *game.Game, request payment.SpellRequest, proposal spellCostProposal) (payment.PreparedSpellCost, bool) {
	manual := e.paymentManaWindow(g, request.PlayerID)
	for _, prepared := range paymentOrch.planner(g).PrepareSpellCosts(request) {
		if prepared.Summary.Index == proposal.optionIndex {
			announced := matchPhyrexianAnnouncements(prepared.Summary.ManaCost, proposal.phyrexian)
			_, feasible := preparedSpellCostFeasibility(g, request, prepared, manual, announced)
			return prepared, feasible
		}
	}
	return payment.PreparedSpellCost{}, false
}

// spellCostFeasibility reports whether request's total cost can be paid from
// current resources (payableNow), and whether it can be paid now or after the
// caster activates payment-time mana abilities (feasible). manual checks
// whether those abilities can make a total payable, nil when the caster has
// none. Phyrexian payments are not announced yet, so each may be paid with mana
// or with 2 life.
func spellCostFeasibility(g *game.Game, request payment.SpellRequest, manual manaWindowCheck) (payableNow, feasible bool) {
	if paymentOrch.canPaySpellCosts(g, request) {
		return true, true
	}
	for _, prepared := range paymentOrch.planner(g).PrepareSpellCosts(request) {
		canPay, ok := preparedSpellCostFeasibility(g, request, prepared, manual, nil)
		if canPay {
			return true, true
		}
		feasible = feasible || ok
	}
	return false, feasible
}

// preparedSpellCostFeasibility classifies one locked total cost (CR 601.2f).
// announced maps positions of the total's mana cost to the payment announced
// for each Phyrexian symbol there, true for 2 life (CR 601.2b); the total is
// payable now when some payment honoring those announcements is, each other
// Phyrexian symbol paid with either mana or life. Mana abilities may be
// activated while paying only when the total, with those announcements,
// includes a mana payment (CR 601.2g), and they add only mana, so a total that
// is not payable now is feasible only if it includes mana, its non-mana costs,
// including the announced life, are payable now, and manual, which is nil when
// no such window is considered, finds the abilities can make it payable.
func preparedSpellCostFeasibility(g *game.Game, request payment.SpellRequest, prepared payment.PreparedSpellCost, manual manaWindowCheck, announced map[int]bool) (payableNow, feasible bool) {
	proposal := request
	proposal.PreparedCost = &prepared
	planner := paymentOrch.planner(g)
	for _, prefs := range phyrexianPaymentCandidates(g, request.PlayerID, prepared.Summary.ManaCost, announced) {
		prefs.AlternativeIndex = prepared.Summary.Index
		proposal.Prefs = prefs
		if planner.CanPaySpellCostsAsPreferred(proposal) {
			return true, true
		}
	}
	proposal.Prefs = nil
	return false, manual != nil && manaCostIncludesManaPayment(prepared.Summary.ManaCost, announced) &&
		announcedPhyrexianLifePayable(g, request.PlayerID, announced) &&
		planner.CanPayPreparedNonManaCosts(proposal) && manual(g, request, prepared, announced)
}

// manaWindowCheck reports whether activating payment-time mana abilities in g
// can make the locked total prepared for request payable with the Phyrexian
// payments announced (CR 601.2g-h).
type manaWindowCheck func(g *game.Game, request payment.SpellRequest, prepared payment.PreparedSpellCost, announced map[int]bool) bool

// paymentManaWindow returns the check of whether player's payment-time mana
// abilities can make a total payable (see manualManaMayPay), or nil when
// player has none to activate.
func (e *Engine) paymentManaWindow(g *game.Game, player game.PlayerID) manaWindowCheck {
	if !e.hasManualPaymentManaAbility(g, player) {
		return nil
	}
	return e.manualManaMayPay
}

// maxOpenPhyrexianCandidates bounds the Phyrexian symbols whose mana-or-life
// payments phyrexianPaymentCandidates enumerates; any further open symbol is
// paid with mana when it can be and otherwise with 2 life.
const maxOpenPhyrexianCandidates = 8

// phyrexianPaymentCandidates returns the Phyrexian payments to try for
// manaCost, one Preferences per combination, in the order the planner pays its
// effective Phyrexian symbols. A symbol announced as 2 life is paid with life
// and one announced as mana is paid only with mana (CR 601.2b), unless an
// effect such as K'rrik, Son of Yawgmoth lets the caster pay 2 life for that
// color while paying (CR 601.2h). Every other Phyrexian symbol may be paid
// either way, so each choice is tried rather than the planner's greedy mana
// first, which could spend mana another symbol needs.
func phyrexianPaymentCandidates(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, announced map[int]bool) []*payment.Preferences {
	if manaCost == nil {
		return []*payment.Preferences{{}}
	}
	state := &rulesPaymentState{g: g}
	var lifeChoices, manaOnly []bool
	var open []int
	for i, symbol := range payment.EffectiveManaCost(state, playerID, *manaCost) {
		if !isPhyrexianSymbol(symbol) {
			continue
		}
		payLife, ok := announced[i]
		if ok && !payLife && symbol.Kind == cost.PhyrexianSymbol && state.PayLifeForManaColor(playerID, symbol.Color) {
			ok = false
		}
		if !ok && len(open) < maxOpenPhyrexianCandidates {
			open = append(open, len(lifeChoices))
		}
		lifeChoices = append(lifeChoices, payLife)
		manaOnly = append(manaOnly, ok && !payLife)
	}
	candidates := make([]*payment.Preferences, 0, 1<<len(open))
	for mask := range 1 << len(open) {
		prefs := &payment.Preferences{
			PhyrexianLifeChoices: append([]bool(nil), lifeChoices...),
			PhyrexianManaOnly:    append([]bool(nil), manaOnly...),
		}
		for bit, position := range open {
			prefs.PhyrexianLifeChoices[position] = mask&(1<<bit) != 0
		}
		candidates = append(candidates, prefs)
	}
	return candidates
}

// announcedPhyrexianLifePayable reports whether playerID can pay the life
// announced for Phyrexian symbols; mana abilities can't help pay it.
func announcedPhyrexianLifePayable(g *game.Game, playerID game.PlayerID, announced map[int]bool) bool {
	life := 0
	for _, payLife := range announced {
		if payLife {
			life += 2
		}
	}
	return life == 0 || playerLife(g, playerID) >= life && (&rulesPaymentState{g: g}).CanPayLife(playerID)
}

// preparedSpellPaymentPreferences takes the caster's remaining payment steps
// for a locked total cost. Each Phyrexian symbol of the proposal keeps the
// payment announced for it (CR 601.2b); the payment-time mana ability window
// then opens only when the total, with those announcements, still includes a
// mana payment (CR 601.2g). The returned undo reverses the reversible mana
// abilities activated in that window (CR 733.1).
func (e *Engine) preparedSpellPaymentPreferences(g *game.Game, request payment.SpellRequest, prepared payment.PreparedSpellCost, proposal spellCostProposal, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*payment.Preferences, func()) {
	summary := prepared.Summary
	announced := matchPhyrexianAnnouncements(summary.ManaCost, proposal.phyrexian)
	undoMana := func() {}
	if manaCostIncludesManaPayment(summary.ManaCost, announced) {
		undoMana = e.activateManualPaymentManaAbilities(g, request.PlayerID, agents, log)
	}
	return e.lockedSpellPaymentPreferences(g, request, summary, announced, agents, log), undoMana
}

// lockedSpellPaymentPreferences takes the caster's remaining payment choices
// for the locked total cost summary. announced maps positions of its mana cost
// to the payment announced for each Phyrexian symbol there, true for 2 life
// (CR 601.2b); see phyrexianPaymentChoicesAfter.
func (e *Engine) lockedSpellPaymentPreferences(g *game.Game, request payment.SpellRequest, summary payment.SpellOptionSummary, announced map[int]bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	lifeChoices, manaOnly := e.phyrexianPaymentChoicesAfter(g, request.PlayerID, summary.ManaCost, announced, agents, log)
	prefs := e.additionalCostPaymentPreferences(g, request.PlayerID, summary.AdditionalCosts, request.XValue,
		request.CardID, request.SourceZone, agents, log)
	prefs.PhyrexianLifeChoices = lifeChoices
	prefs.PhyrexianManaOnly = manaOnly
	prefs.AlternativeIndex = summary.Index
	return prefs
}

// spellCostsIncludePhyrexian reports whether any cost option of any request
// includes a Phyrexian symbol whose payment the caster announces (CR 601.2b).
func spellCostsIncludePhyrexian(g *game.Game, requests []payment.SpellRequest) bool {
	planner := paymentOrch.planner(g)
	for _, request := range requests {
		for _, option := range planner.SpellCostOptions(request) {
			if option.ManaCost != nil && slices.ContainsFunc(*option.ManaCost, isPhyrexianSymbol) {
				return true
			}
		}
	}
	return false
}

// lockCastPayment locks the total cost of proposal's option for request once
// the spell's targets and other choices are final (CR 601.2f). When that total,
// with the Phyrexian payments announced with the proposal (CR 601.2b), includes
// a mana payment, the caster may then activate mana abilities (CR 601.2g)
// before taking the remaining payment choices (CR 601.2h). It fails, activating
// nothing, when request no longer offers the option or its total can't be paid
// now or after activating payment-time mana abilities. The returned undo
// reverses the reversible mana abilities activated (CR 733.1).
func (e *Engine) lockCastPayment(g *game.Game, request payment.SpellRequest, proposal spellCostProposal, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.SpellRequest, func(), bool) {
	prepared, feasible := e.lockSpellCostProposal(g, request, proposal)
	if !feasible {
		return request, func() {}, false
	}
	request.PreparedCost = &prepared
	prefs, undoMana := e.preparedSpellPaymentPreferences(g, request, prepared, proposal, agents, log)
	request.Prefs = prefs
	return request, undoMana, true
}

// matchPhyrexianAnnouncements maps each announced Phyrexian symbol to its
// position in the locked total manaCost, true when it is paid with 2 life.
// Cost modifiers keep the relative order of the proposal's symbols, so each
// symbol takes the first unused equal announcement. Phyrexian symbols the
// proposal didn't include, such as a tax or a splice cost, stay unannounced
// and are chosen while paying (CR 601.2h).
func matchPhyrexianAnnouncements(manaCost *cost.Mana, announcements []phyrexianAnnouncement) map[int]bool {
	if manaCost == nil || len(announcements) == 0 {
		return nil
	}
	matched := map[int]bool{}
	used := make([]bool, len(announcements))
	for i, symbol := range *manaCost {
		for j, announcement := range announcements {
			if !used[j] && announcement.symbol == symbol {
				used[j] = true
				matched[i] = announcement.payLife
				break
			}
		}
	}
	return matched
}

// manaCostIncludesManaPayment reports whether manaCost includes a mana payment
// once each Phyrexian symbol announced as a life payment is paid with life. A
// {B} that an effect such as K'rrik, Son of Yawgmoth lets the caster pay with
// life remains a mana payment: that effect changes only how the cost is paid
// (CR 601.2h), not the total cost.
func manaCostIncludesManaPayment(manaCost *cost.Mana, announced map[int]bool) bool {
	if manaCost == nil {
		return false
	}
	for i, symbol := range *manaCost {
		if !announced[i] && (cost.Mana{symbol}).ManaValue() > 0 {
			return true
		}
	}
	return false
}

func (e *Engine) hasManualPaymentManaAbility(g *game.Game, player game.PlayerID) bool {
	return len(e.paymentManaAbilityActions(g, player)) != 0
}

// paymentManaAbilityActions returns the manual mana abilities player may
// activate while paying a cost (CR 601.2g).
func (e *Engine) paymentManaAbilityActions(g *game.Game, player game.PlayerID) []action.Action {
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	return e.legalManaAbilityActions(g, player)
}

// paymentManaAbilityOptions offers "Finish activating mana abilities" and then
// each of actions, identified by its source: a permanent by its object ID and
// current public characteristics (a face-down one only as face-down, CR 708.2),
// so distinct same-named permanents and tokens stay distinguishable, and a card
// in player's hand by its card instance. A label names the source before the
// ability's text, adding the source's ID when sources share a name.
func paymentManaAbilityOptions(g *game.Game, player game.PlayerID, actions []action.Action) []game.ChoiceOption {
	type source struct {
		name      string
		id        id.ID
		permanent id.ID
		card      opt.V[game.ChoiceCardInfo]
	}
	sources := make([]source, len(actions))
	namedSources := map[string]map[id.ID]bool{}
	for i, candidate := range actions {
		activation, _ := candidate.ActivateAbilityPayload()
		src := source{id: activation.SourceID}
		if permanent, ok := permanentByObjectID(g, activation.SourceID); ok {
			src.name, src.permanent, src.card = permanentChoiceLabel(g, permanent), permanent.ObjectID, permanentChoiceInfo(g, permanent)
		} else {
			src.name, src.card = cardChoiceLabel(g, activation.SourceID), cardChoiceInfo(g, activation.SourceID)
		}
		if namedSources[src.name] == nil {
			namedSources[src.name] = map[id.ID]bool{}
		}
		namedSources[src.name][src.id] = true
		sources[i] = src
	}
	options := []game.ChoiceOption{{Index: 0, Label: "Finish activating mana abilities"}}
	for i, candidate := range actions {
		activation, _ := candidate.ActivateAbilityPayload()
		src := sources[i]
		name := src.name
		if len(namedSources[name]) > 1 {
			name = fmt.Sprintf("%s [%d]", name, src.id)
		}
		label := name + ": " + activatedAbilityText(g, player, activation)
		options = append(options, game.ChoiceOption{Index: len(options), Label: label, Card: src.card, SourcePermanent: src.permanent})
	}
	return options
}

// activateManualPaymentManaAbilities lets player activate mana abilities while
// paying a cost (CR 601.2g). The returned undo reverses those activations
// when the cost then goes unpaid (CR 733.1); see undoManualPaymentMana.
func (e *Engine) activateManualPaymentManaAbilities(g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) func() {
	if log == nil {
		// The undo replays kept activations from the choices logged for them.
		log = &TurnLog{}
	}
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	var steps []manaPaymentStep
	finish := func() {
		e.undoManualPaymentMana(g, player, steps, log)
	}
	for {
		actions := e.legalManaAbilityActions(g, player)
		if len(actions) == 0 {
			return finish
		}
		entryStart := len(log.Entries)
		// Options can name cards in the player's hand (CR 400.2), and whether
		// a window opens and how many options it has can reveal one, so no
		// window is logged; each activation chosen in it is logged instead.
		selected := e.chooseUnloggedChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: player, Prompt: "Activate a mana ability while paying?",
			Options: paymentManaAbilityOptions(g, player, actions), MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
			Private: true,
		}, log)
		if selected[0] == 0 {
			return finish
		}
		chosen := actions[selected[0]-1]
		activation, _ := chosen.ActivateAbilityPayload()
		step := manaPaymentStep{restore: reversibleManaCheckpoint(g, log), activation: activation}
		var libraries [game.NumPlayers]zone.Zone
		for owner := range libraries {
			libraries[owner] = g.Players[owner].Library.Clone()
		}
		eventStart, choiceStart, spent := len(g.Events), len(log.Choices), manaSpent(g)
		actionLog := &ActionLog{Player: player, DecisionMaker: g.DecisionMaker(player, game.DecisionRules), Action: chosen}
		recordActionSource(g, player, actionLog, chosen)
		actionEntry := log.addPaymentManaActivation(actionLog)
		if !e.applyActivateAbilityWithChoices(g, player, activation, agents, log) {
			panic("validated payment-time mana ability failed")
		}
		recordActionManaTaps(g, log, actionEntry, eventStart)
		for i := entryStart; i < len(log.Entries); i++ {
			log.Entries[i].paymentMana = true
		}
		// CR 733.1 forbids reversing a draw, library move, shuffle or reveal.
		step.reversible = manaActivationIsReversible(g, libraries, eventStart)
		step.spentMana = manaSpent(g) != spent
		step.choices = slices.Clone(log.Choices[choiceStart:])
		step.entries = slices.Clone(log.Entries[entryStart:])
		step.events = replayableManaEvents(g.Events[eventStart:])
		steps = append(steps, step)
	}
}
