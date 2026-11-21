package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AssistSelection is invocation-local. A missing selection forecasts a
// possible payment; that forecast never authorizes spending another pool.
type AssistSelection struct {
	chosen bool
	helper opt.V[game.PlayerID]
	amount opt.V[int]
}

func (p Partial) WithAssistPlayer(helper opt.V[game.PlayerID]) Partial {
	p.Assist = AssistSelection{chosen: true, helper: helper}
	return p
}

func (p Partial) ForecastAssistPlayer(helper game.PlayerID) Partial {
	p.Assist = AssistSelection{helper: opt.Val(helper)}
	return p
}

func (p Partial) WithAssistContribution(helper game.PlayerID, amount int) Partial {
	p.Assist = AssistSelection{chosen: true, helper: opt.Val(helper), amount: opt.Val(amount)}
	return p
}

// AfterAssistManaWindow retains the announced spell and nonmana decisions.
// Both players must now pay from their pools or use casting cost mechanics.
func (p Partial) AfterAssistManaWindow() Partial {
	p.poolOnly = true
	p.ManaAbilities, p.ManaOccurrences, p.ManaPayments, p.ManaChoices = nil, nil, nil, nil
	return p
}

type assistPlan struct {
	player    game.PlayerID
	amount    int
	spent     map[mana.Unit]int
	first     opt.V[ManaActivation]
	committed bool
}

// AssistReceipt keeps spend riders attributable to the player who paid.
type AssistReceipt struct {
	Player game.PlayerID
	Before map[mana.Unit]int
	Spent  map[mana.Unit]int
}

func (r ApplyResult) TotalSpellPoolSpend() map[mana.Unit]int {
	if !r.Assist.Exists {
		return r.Spell.PoolSpend
	}
	total := clonePoolSpend(r.Spell.PoolSpend)
	if total == nil {
		total = make(map[mana.Unit]int)
	}
	for unit, n := range r.Assist.Val.Spent {
		total[unit] += n
	}
	return total
}

// AssistLimit reads the locked generic component, after taxes, reductions and
// X. Color-spending permissions and non-generic symbols do not increase it.
func (p Plan) AssistLimit() int {
	if p.request.kind != spellPayment || p.request.spell.PreparedCost == nil ||
		p.request.spell.FaceDown || !p.request.spell.Card.HasKeyword(game.Assist) {
		return 0
	}
	return genericCostAmount(p.request.spell.PreparedCost.option.manaCost)
}

func (p Plan) SuggestedAssistant() opt.V[game.PlayerID] {
	if p.assist == nil {
		return opt.V[game.PlayerID]{}
	}
	return opt.Val(p.assist.player)
}

// SuggestedAssistActivation is the first activation of the public witness.
// It is a fallback, not a commitment by the chosen player.
func (p Plan) SuggestedAssistActivation() opt.V[ManaActivation] {
	if p.assist == nil || !p.assist.first.Exists {
		return opt.V[ManaActivation]{}
	}
	return opt.Val(cloneAllocation(ManaAllocation{Activations: []ManaActivation{p.assist.first.Val}}).Activations[0])
}

// AssistPoolLimit counts only the selected player's eligible floating mana.
func (p Planner) AssistPoolLimit(plan Plan, player game.PlayerID) int {
	limit := plan.AssistLimit()
	helper, ok := p.s.Player(player)
	if limit <= 0 || !ok || helper.Eliminated || player == plan.request.spell.PlayerID {
		return 0
	}
	var available int
	for _, n := range snapshotPool(p.s, helper, assistSpendContext(plan.request.spell.PreparedCost.option)) {
		available += min(n, limit-available)
	}
	return available
}

func assistSpendContext(option spellCostOption) spendContext {
	ctx := option.spendContext()
	ctx.spendAs = game.ManaSpendNormal
	return ctx
}

func reducedForAssist(total *cost.Mana, contribution int) *cost.Mana {
	rest := copyManaCost(total)
	if rest == nil {
		return nil
	}
	for i, s := range *rest {
		if s.Kind == cost.GenericSymbol {
			n := min(contribution, s.Generic)
			(*rest)[i] = cost.O(s.Generic - n)
			contribution -= n
		}
	}
	return rest
}

func (p Planner) complete(request Request, partial Partial) (Plan, completion, bool) {
	assist := partial.Assist
	if request.kind != spellPayment || request.spell.Card == nil || request.spell.FaceDown ||
		!request.spell.Card.HasKeyword(game.Assist) || assist.chosen && !assist.helper.Exists {
		return p.completeSinglePayer(request, partial)
	}
	if !assist.amount.Exists && (!assist.helper.Exists || assist.chosen) {
		if plan, resolved, ok := p.completeSinglePayer(request, partial); ok {
			return plan, resolved, true
		}
	}
	if state, ok := p.s.(interface{ BeginPaymentRead() func() }); ok {
		defer state.BeginPaymentRead()()
	}
	req := request.spell
	if partial.X.Exists {
		req.XValue = partial.X.Val
	}
	var totals []PreparedSpellCost
	if request.spell.PreparedCost != nil {
		totals = []PreparedSpellCost{*request.spell.PreparedCost}
	} else {
		totals = p.PrepareSpellCosts(req)
	}
	for _, total := range totals {
		option := total.option
		if partial.Option.Exists && option.index != partial.Option.Val {
			continue
		}
		if request.spell.PreparedCost != nil {
			option = spellCostOptionsForRequest(p.s, request.spell)[0]
		}
		limit := genericCostAmount(option.manaCost)
		if limit <= 0 {
			continue
		}
		for seat := range game.NumPlayers {
			helper := game.PlayerID(seat)
			if helper == request.spell.PlayerID || assist.helper.Exists && assist.helper.Val != helper {
				continue
			}
			if plan, resolved, ok := p.completeAssistForPlayer(request, partial, option, helper, limit); ok {
				// Only the caster's share changed; future revalidation must
				// still start from the original locked total.
				plan.request.spell.PreparedCost = &total
				return plan, resolved, true
			}
		}
	}
	return Plan{}, completion{}, false
}

func (p Planner) completeAssistForPlayer(request Request, partial Partial, option spellCostOption, helper game.PlayerID, limit int) (Plan, completion, bool) {
	assist := partial.Assist
	player, ok := p.s.Player(helper)
	if !ok || player.Eliminated || assist.amount.Exists && (assist.amount.Val < 0 || assist.amount.Val > limit) {
		return Plan{}, completion{}, false
	}
	ctx := assistSpendContext(option)
	var plan Plan
	var resolved completion
	// The helper's projected mana window comes first. Try the caster's
	// remainder in that same resource view, not against the original board.
	trySplit := func(state State, activations []ManaActivation, _ game.ResourceProjection) (paymentPlan, bool) {
		helperState, _ := state.Player(helper)
		available := 0
		for _, n := range snapshotPool(state, helperState, ctx) {
			available += min(n, limit-available)
		}
		minimum := 1
		if assist.amount.Exists {
			minimum = assist.amount.Val
			available = min(available, minimum)
		}
		for n := available; n >= minimum; n-- {
			remainder := option
			remainder.manaCost = reducedForAssist(option.manaCost, n)
			split := request
			split.spell.PreparedCost = &PreparedSpellCost{Summary: remainder.summary(), option: remainder}
			caster, paid, ok := New(state).completeSinglePayer(split, partial)
			if !ok {
				continue
			}
			generic := cost.Mana{cost.O(n)}
			helperMana, ok := completeMana(state, helper, &generic, 0, manaSourceRestrictions{}, ctx, nil, 0, nil, 0, zone.None, Partial{poolOnly: true})
			if !ok {
				continue
			}
			allocation := &assistPlan{player: helper, amount: n, spent: helperMana.poolSpend, committed: assist.amount.Exists}
			if len(activations) > 0 {
				allocation.first = opt.Val(activations[0])
			}
			caster.assist, paid.assist, paid.option = allocation, allocation, option
			plan, resolved = caster, paid
			return helperMana, true
		}
		return paymentPlan{}, false
	}
	if assist.chosen || partial.poolOnly {
		_, ok = trySplit(p.s, nil, game.ResourceProjection{})
	} else if state, modeled := p.s.(ManaState); modeled {
		// Only public resources can make the cast appear in the caster's
		// options. The invited player can still choose private or opaque
		// abilities in their live window.
		restrictions := manaSourceRestrictions{allowResourceActivation: publicManaCandidate, resourceCompletion: trySplit}
		_, ok = completeManaActivations(state, helper, nil, 0, restrictions, ctx, nil, 0, nil, 0, zone.None, Partial{})
	} else {
		_, ok = trySplit(p.s, nil, game.ResourceProjection{})
	}
	return plan, resolved, ok
}

func publicManaCandidate(candidate ManaActivationCandidate) bool {
	return candidate.SourceZone == zone.Battlefield && !slices.ContainsFunc(candidate.Body.AdditionalCosts, func(c cost.Additional) bool {
		return c.Kind == cost.AdditionalDiscard || c.Kind == cost.AdditionalExile && cardCostZone(c) != zone.Graveyard
	})
}
