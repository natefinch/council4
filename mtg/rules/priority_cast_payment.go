package rules

import (
	"fmt"
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// The procedure owns the source transition. Only an invalid proposal or an
// opaque payment cancels it; normal payment choices always retain a witness.
type costedSpellSource struct {
	player   game.PlayerID
	card     id.ID
	zone     zone.Type
	place    castSourcePlace
	counters castCounterTransition
}

func (s *costedSpellSource) move(g *game.Game) bool {
	card, ok := g.GetCardInstance(s.card)
	if !ok {
		return false
	}
	s.counters = prepareCastCounterTransition(g, card, s.zone)
	s.place, ok = takeCastSourceCard(castSourcePlayer(g, g.Players[s.player], s.card, s.zone), s.card, s.zone)
	if ok {
		s.counters.apply(g)
	}
	return ok
}

func (s *costedSpellSource) cancel(g *game.Game) {
	restoreCastSourceCard(s.place)
	s.counters.undo(g)
}

// Availability reads the same CR 601.2a state as execution, without a Clone,
// speculative mana activation, ID allocation or mutation of the live graph.
func priorityCastView(g *game.Game, object *game.StackObject) *game.Game {
	view := g.ResourceView(game.ResourceProjection{})
	if !object.Copy {
		holder := castSourcePlayer(g, g.Players[object.Controller], object.SourceID, object.SourceZone)
		player := *holder
		source := castSourceZone(&player, object.SourceZone)
		if source == nil {
			return view
		}
		*source = source.Clone()
		source.Remove(object.SourceID)
		view.Players[holder.ID] = &player
		view.CardCounters = maps.Clone(g.CardCounters)
		if card, ok := g.GetCardInstance(object.SourceID); ok {
			prepareCastCounterTransition(g, card, object.SourceZone).apply(view)
		}
	}
	var stack game.Stack
	for _, previous := range g.Stack.Objects() {
		stack.Push(previous)
	}
	proposed := *object
	proposed.ID = id.ID(g.IDGen.Current() + 1)
	proposed.Casting = true
	stack.Push(&proposed)
	view.Stack = stack
	return view
}

func (e *Engine) priorityCastCostsAvailable(g *game.Game, object *game.StackObject, requests []payment.SpellRequest) bool {
	view := priorityCastView(g, object)
	for _, request := range requests {
		if !castRouteAllowsX(request.Card, request.XValue, request) {
			continue
		}
		if !object.Mutate && !spellTargetsValidForCast(view, request.PlayerID, priorityAnnouncementDef(request.Card, object.Overloaded),
			request.ChosenModes, request.Targets, request.XValue, effectiveKickerCount(request.KickerPaid, request.KickerCount),
			game.CastBranch{XValue: opt.Val(request.XValue), KickerCount: opt.Val(effectiveKickerCount(request.KickerPaid, request.KickerCount)),
				Kicked: request.KickerPaid, Bargained: request.Bargained, Bestowed: request.Bestowed, Offspring: request.Offspring, GiftPromised: object.GiftPromised}) {
			continue
		}
		if slices.ContainsFunc(request.Targets, func(target game.Target) bool { return target.Kind == game.TargetDeferred }) {
			cast := action.CastSpellAction{XValue: request.XValue, ChosenModes: request.ChosenModes, KickerPaid: request.KickerPaid,
				KickerCount: request.KickerCount, Bargained: request.Bargained, Bestowed: request.Bestowed, Offspring: request.Offspring, GiftPromised: object.GiftPromised}
			if _, ok := e.pendingSpellPayment(view, priorityAnnouncementDef(request.Card, object.Overloaded), cast)(request, payment.Partial{}); ok {
				return true
			}
			continue
		}
		if _, ok, _ := e.PaymentPlanner(view).Complete(payment.ForSpell(request), payment.Partial{}); ok {
			return true
		}
	}
	return false
}

func priorityAnnouncementDef(def *game.CardDef, overloaded bool) *game.CardDef {
	if overloaded {
		return overloadSpellDef(def)
	}
	return def
}

func costedSpellPresent(g *game.Game, object *game.StackObject) bool {
	return slices.Contains(g.Stack.Objects(), object)
}

type spellPaymentCompletion func(payment.SpellRequest, payment.Partial) (payment.Plan, bool)

// Search lazily until a complete payment with exact targets is found. A
// policy's target sample is never a feasibility proof (CR 601.2b/c/f/h).
func (e *Engine) pendingSpellPayment(g *game.Game, def *game.CardDef, cast action.CastSpellAction) spellPaymentCompletion {
	specs := prioritySpellTargetSpecs(def, cast.ChosenModes, castBranchForCast(cast), effectiveKickerCount(cast.KickerPaid, cast.KickerCount))
	return func(request payment.SpellRequest, partial payment.Partial) (payment.Plan, bool) {
		var plan payment.Plan
		found := false
		visitTargetChoicesForSpec(g, request.PlayerID, def, 0, game.Event{}, specs, 0, nil, nil,
			&[][]game.Target{}, &[][]int{}, targetChoiceOptions{xValue: opt.Val(cast.XValue), external: true, visit: func(targets []game.Target, _ []int) bool {
				request.Targets = targets
				plan, found, _ = e.PaymentPlanner(g).Complete(payment.ForSpell(request), partial)
				return !found
			}})
		return plan, found
	}
}

func prioritySpellTargetSpecs(def *game.CardDef, modes []int, branch game.CastBranch, kicker int) []game.TargetSpec {
	specs := spellTargetSpecs(def, modes, branch)
	for index := range specs {
		if specs[index].CountEqualsKickerPlusOne {
			specs[index].MinTargets, specs[index].MaxTargets = kicker+1, kicker+1
		}
	}
	return specs
}

// Routes differ in permission, alternative costs and spending riders, not in
// their procedure. Announce an option and each Phyrexian symbol only when the
// pending target choices have a joint completion honoring that announcement.
func (e *Engine) choosePrioritySpellPayment(g *game.Game, routes []castRoute, requests []payment.SpellRequest, complete spellPaymentCompletion, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (int, payment.Partial, bool) {
	type option struct {
		route   int
		summary payment.SpellOptionSummary
		partial payment.Partial
	}
	var choices []option
	for route, request := range requests {
		if !castRouteAllowsX(request.Card, request.XValue, request) {
			continue
		}
		for _, summary := range e.PaymentPlanner(g).SpellCostOptions(request) {
			partial := payment.Partial{Option: opt.Val(summary.Index)}
			if _, ok := complete(request, partial); ok {
				choices = append(choices, option{route: route, summary: summary, partial: partial})
			}
		}
	}
	if len(choices) == 0 {
		return 0, payment.Partial{}, false
	}
	selected := 0
	if len(choices) > 1 {
		permissions := make([]game.RuleEffect, len(routes))
		for index, route := range routes {
			permissions[index] = castRoutePermission(route)
		}
		sources := permissionSources(g, permissions)
		summaries := make([]payment.SpellOptionSummary, len(choices))
		for index, choice := range choices {
			summaries[index] = choice.summary
		}
		labels := distinctSpellCostLabels(summaries)
		labelCounts := make(map[string]int)
		for _, label := range labels {
			labelCounts[label]++
		}
		options := make([]game.ChoiceOption, len(choices))
		for index, choice := range choices {
			label := labels[index]
			if len(routes) > 1 && labelCounts[label] > 1 {
				label += " (" + sources[choice.route].label + ")"
			}
			options[index] = spellCostChoiceOption(index, label, choice.summary)
			if len(routes) > 1 {
				options[index].SourcePermanent = sources[choice.route].permanent
			}
		}
		selected = e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: requests[0].PlayerID, Prompt: "Choose spell cost",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
		}, log)[0]
	}
	chosen := choices[selected]
	partial := chosen.partial
	if chosen.summary.ManaCost != nil {
		phyrexian := 0
		for _, symbol := range *chosen.summary.ManaCost {
			if !isPhyrexianSymbol(symbol) {
				continue
			}
			step := payment.Step{Kind: payment.StepPhyrexian, Index: phyrexian}
			var payments []payment.Choice
			var options []game.ChoiceOption
			for _, life := range []bool{false, true} {
				choice := payment.Choice{PayLife: life}
				if _, ok := complete(requests[chosen.route], partial.With(step, choice)); !ok {
					continue
				}
				label := "Pay with mana"
				if life {
					label = "Pay 2 life"
				}
				options = append(options, game.ChoiceOption{Index: len(payments), Label: label})
				payments = append(payments, choice)
			}
			if len(payments) == 0 {
				return 0, partial, false
			}
			selected := e.chooseChoice(g, agents, game.ChoiceRequest{
				Kind: game.ChoicePayment, Player: requests[0].PlayerID, Prompt: fmt.Sprintf("Pay %s", (cost.Mana{symbol}).String()),
				Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
			}, log)[0]
			partial = partial.With(step, payments[selected])
			phyrexian++
		}
	}
	return chosen.route, partial, true
}

// Choose one next target at a time. Every candidate extends the fixed prefix
// to a full set whose exact target-dependent total can be paid; two individually
// affordable targets can never combine into an unpayable announced set.
func (e *Engine) chooseCostedSpellTargets(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, request payment.SpellRequest, partial payment.Partial, agents [game.NumPlayers]PlayerAgent, log *TurnLog, completion ...spellPaymentCompletion) ([]game.Target, bool) {
	complete := func(request payment.SpellRequest, partial payment.Partial) (payment.Plan, bool) {
		plan, ok, _ := e.PaymentPlanner(g).Complete(payment.ForSpell(request), partial)
		return plan, ok
	}
	if len(completion) > 0 {
		complete = completion[0]
	}
	specs := prioritySpellTargetSpecs(def, cast.ChosenModes, castBranchForCast(cast), effectiveKickerCount(cast.KickerPaid, cast.KickerCount))
	var targets []game.Target
	var counts []int
	preferredCounts, _ := targetCountsForSpecs(g, player, def, 0, game.Event{}, specs, cast.Targets)
	offset := 0
	for index, raw := range specs {
		spec := normalizeTargetSpec(&raw)
		var preferred []game.Target
		if index < len(preferredCounts) {
			preferred = cast.Targets[offset : offset+preferredCounts[index]]
			offset += preferredCounts[index]
		}
		candidates := targetCandidatesForSpec(g, player, def, 0, game.Event{}, &spec)
		candidates = filterAnnouncedManaValueTargets(g, spec, candidates, opt.Val(cast.XValue))
		if spec.DistinctFromPriorTargets {
			candidates = filterTargetsDistinctFrom(candidates, targets)
		}
		extends := func(segment []game.Target, count int) bool {
			found := false
			visitTargetChoicesForSpec(g, player, def, 0, game.Event{}, specs, index, targets, counts,
				&[][]game.Target{}, &[][]int{}, targetChoiceOptions{xValue: opt.Val(cast.XValue), external: true,
					requiredSpec: index, required: segment, requiredCount: opt.Val(count), visit: func(all []game.Target, _ []int) bool {
						request.Targets = all
						_, found = complete(request, partial)
						return !found
					}})
			return found
		}
		if targetSpecUsesExternalChooser(&spec) {
			var opponents []game.PlayerID
			feasible := make(map[game.PlayerID][]game.Target)
			for _, opponent := range choosingOpponentsForTargetSpec(g, player, def, 0, &spec) {
				candidates := targetCandidatesForSpecChosenBy(g, player, opponent, def, 0, game.Event{}, &spec)
				candidates = filterAnnouncedManaValueTargets(g, spec, candidates, opt.Val(cast.XValue))
				if spec.DistinctFromPriorTargets {
					candidates = filterTargetsDistinctFrom(candidates, targets)
				}
				for _, candidate := range candidates {
					if extends([]game.Target{candidate}, 1) {
						feasible[opponent] = append(feasible[opponent], candidate)
					}
				}
				if len(feasible[opponent]) > 0 {
					opponents = append(opponents, opponent)
				}
			}
			if len(opponents) == 0 {
				return nil, false
			}
			opponent, ok := e.chooseTargetingOpponent(g, player, &spec, opponents, agents, log)
			if !ok {
				return nil, false
			}
			target, ok := e.chooseTargetFromCandidates(g, opponent, &spec, feasible[opponent], agents, log)
			if !ok {
				return nil, false
			}
			targets, counts = append(targets, target), append(counts, 1)
			continue
		}
		var possibleCounts []int
		for count := spec.MinTargets; count <= min(spec.MaxTargets, len(candidates)); count++ {
			if extends(nil, count) {
				possibleCounts = append(possibleCounts, count)
			}
		}
		if len(possibleCounts) == 0 {
			return nil, false
		}
		count := possibleCounts[0]
		if len(possibleCounts) > 1 {
			options := make([]game.ChoiceOption, len(possibleCounts))
			fallback := 0
			for position, count := range possibleCounts {
				options[position] = game.ChoiceOption{Index: position, Label: fmt.Sprint(count)}
				if count == len(preferred) {
					fallback = position
				}
			}
			selected := e.chooseChoice(g, agents, game.ChoiceRequest{
				Kind: game.ChoiceResolution, Player: player, Prompt: "Choose the number of targets.",
				Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{fallback},
			}, log)[0]
			count = possibleCounts[selected]
		}
		var chosen []game.Target
		for len(chosen) < count {
			var options []game.ChoiceOption
			var next []game.Target
			for _, candidate := range candidates {
				if targetsContainAll(chosen, []game.Target{candidate}) || !extends(append(slices.Clone(chosen), candidate), count) {
					continue
				}
				options = append(options, game.ChoiceOption{Index: len(next), Targets: []game.Target{candidate}})
				next = append(next, candidate)
			}
			if len(options) == 0 {
				return nil, false
			}
			fallback := 0
			provisional := false
			if len(chosen) < len(preferred) {
				for position, target := range next {
					if sameTargetObject(target, preferred[len(chosen)]) {
						fallback, provisional = position, true
						break
					}
				}
			}
			selected := e.chooseChoice(g, agents, game.ChoiceRequest{
				Kind: game.ChoiceTarget, Player: player, Prompt: "Choose spell targets.",
				Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{fallback}, Provisional: provisional,
			}, log)[0]
			chosen = append(chosen, next[selected])
		}
		if !extends(chosen, count) {
			return nil, false
		}
		targets = append(targets, chosen...)
		counts = append(counts, len(chosen))
	}
	return bindCardTargetZoneVersions(g, targets)
}

func targetsContainAll(targets, required []game.Target) bool {
	for _, target := range required {
		if !slices.ContainsFunc(targets, func(candidate game.Target) bool { return sameTargetObject(candidate, target) }) {
			return false
		}
	}
	return true
}

// Numeric spells need one witnessed proposal, not a sample of target sets.
// Their full target domain is offered during CR 601.2c, after announcements.
func (e *Engine) prioritySpellTargets(g *game.Game, player game.PlayerID, cardID id.ID, source zone.Type, face game.FaceIndex, def *game.CardDef, modes []int, branch game.CastBranch, kickers ...int) targetChoiceResult {
	kicker := 0
	if branch.Kicked {
		kicker = 1
	}
	if len(kickers) > 0 {
		kicker = kickers[0]
	}
	specs := prioritySpellTargetSpecs(def, modes, branch, kicker)
	numeric := costHasVariableMana(manaCostPtr(def.ManaCost)) || additionalCostsUseX(def.AdditionalCosts) || spellHasMultikicker(def)
	combinatorial := len(specs) > 1 || slices.ContainsFunc(specs, func(spec game.TargetSpec) bool { return spec.MaxTargets > 1 })
	for _, raw := range specs {
		spec := normalizeTargetSpec(&raw)
		if !targetSpecValid(&spec) {
			return targetChoiceResult{kind: targetInvalidSpec}
		}
	}
	result := targetChoiceResult{kind: targetNoLegalChoices}
	object := &game.StackObject{
		Kind: game.StackSpell, SourceID: cardID, Controller: player, SourceZone: source, Face: face,
		XValue: branch.XValue.Val, ChosenModes: modes, Copy: source == zone.Battlefield, Bestowed: branch.Bestowed,
	}
	view := priorityCastView(g, object)
	visitTargetChoicesForSpec(view, player, def, 0, game.Event{}, specs, 0, nil, nil,
		&[][]game.Target{}, &[][]int{}, targetChoiceOptions{xValue: branch.XValue, visit: func(targets []game.Target, counts []int) bool {
			ok := false
			if source == zone.Battlefield {
				ok = e.canCastPreparedCopy(g, player, preparedSourcePermanent(g, cardID), targets, branch.XValue.Val, modes)
			} else {
				ok = e.canCastSpellFaceFromZoneWithOptions(g, player, cardID, source, face, targets, branch.XValue.Val, modes, kicker,
					false, branch.GiftPromised, branch.Bargained, branch.Bestowed, branch.Offspring)
			}
			if !ok {
				return true
			}
			result.kind = targetLegalChoicesFound
			result.choices = append(result.choices, slices.Clone(targets))
			result.targetCounts = append(result.targetCounts, slices.Clone(counts))
			return !numeric || !combinatorial
		}})
	return result
}
