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

// costedAction describes a proposal, not a second payment implementation. A
// spell descriptor can supply its move-to-stack and target-dependent request;
// an ability supplies an object, and a special action supplies neither.
type costedAction struct {
	cost          activationCost
	object        *game.StackObject
	usesStack     bool
	disclose      func(*game.Game)
	announcements func(*game.Game, [game.NumPlayers]PlayerAgent, *TurnLog) bool
	targets       func(*game.Game, [game.NumPlayers]PlayerAgent, *TurnLog) bool
	lock          func(*game.Game) payment.Request
	beforePayment func(*game.Game, payment.Plan)
	commit        func(*game.Game, payment.ApplyResult, [game.NumPlayers]PlayerAgent, *TurnLog, int)
}

func (e *Engine) costedAction(g *game.Game, player game.PlayerID, descriptor costedAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	if !e.costedActionAvailable(g, player, descriptor.cost) {
		return actionFailed
	}
	defer g.OpenLibraryTopWindow()()
	ids, eventStart := g.IDGen.Current(), len(g.Events)
	eventsNil := g.Events == nil
	truncateAnnouncements := func() {
		g.Events = g.Events[:eventStart]
		if eventsNil {
			g.Events = nil
		}
	}
	var savedLog TurnLog
	if log != nil {
		savedLog = *log
		savedLog.Entries = slices.Clone(log.Entries)
	}
	cancelAnnouncement := func(restoreIDs bool) {
		if descriptor.usesStack {
			g.Stack.RemoveByID(descriptor.object.ID)
		}
		if restoreIDs {
			g.IDGen.Restore(ids)
		}
	}
	if descriptor.disclose != nil {
		descriptor.disclose(g)
	}
	if descriptor.object != nil {
		descriptor.object.ID = g.IDGen.Next()
		if descriptor.usesStack {
			g.Stack.Push(descriptor.object)
		}
	}
	failProposal := func() actionOutcome {
		cancelAnnouncement(true)
		truncateAnnouncements()
		if log != nil {
			*log = savedLog
		}
		return actionFailed
	}
	if descriptor.announcements != nil && !descriptor.announcements(g, agents, log) {
		return failProposal()
	}
	request := descriptor.cost.request
	partial := payment.Partial{}
	planner := e.PaymentPlanner(g, log)
	choice, ok := e.choosePaymentStep(g, player, planner, request, partial, payment.Step{Kind: payment.StepCostOption}, "Choose ability cost", agents, log)
	if !ok {
		return failProposal()
	}
	partial = partial.With(payment.Step{Kind: payment.StepCostOption}, choice)
	costs := choice.Plan.Costs()
	if costs.ManaCost != nil {
		phyrexian := 0
		for _, symbol := range *costs.ManaCost {
			if !isPhyrexianSymbol(symbol) {
				continue
			}
			step := payment.Step{Kind: payment.StepPhyrexian, Index: phyrexian}
			choice, ok := e.choosePaymentStep(g, player, planner, request, partial, step, "Choose Phyrexian payment", agents, log)
			if !ok {
				return failProposal()
			}
			partial = partial.With(step, choice)
			phyrexian++
		}
	}
	if descriptor.targets != nil && !descriptor.targets(g, agents, log) {
		return failProposal()
	}
	if descriptor.lock != nil {
		request = descriptor.lock(g)
	}
	plan, ok, _ := planner.Complete(request, partial)
	if !ok {
		return failProposal()
	}
	// Only the total is locked. Resource choices remain open (CR 601.2f–h).
	request, costs = plan.Request(), plan.Costs()
	for index, additional := range costs.AdditionalCosts {
		if additional.ChoiceGroup != 0 {
			group := int(additional.ChoiceGroup)
			if _, selected := partial.Groups[additional.ChoiceGroup]; !selected {
				step := payment.Step{Kind: payment.StepCostGroup, Index: group}
				choice, ok := e.choosePaymentStep(g, player, planner, request, partial, step, "Choose cost", agents, log)
				if !ok {
					return failProposal()
				}
				partial = partial.With(step, choice)
			}
			if partial.Groups[additional.ChoiceGroup] != index {
				continue
			}
		}
		step := payment.Step{Kind: payment.StepComponent, Index: index}
		for !partial.Components[index].Finished {
			prompt := payment.AdditionalCostText(additional)
			if additional.Kind == cost.AdditionalReturnUnblockedAttacker {
				prompt = "Choose an unblocked attacker to return"
			}
			choice, ok := e.choosePaymentStep(g, player, planner, request, partial, step, prompt, agents, log)
			if !ok {
				return failProposal()
			}
			partial = partial.With(step, choice)
		}
	}
	announcementEnd := len(g.Events)
	inheritedOpaque := g.OpaquePaymentActive()
	opaqueSelected := false
	var checkpoint *opaquePaymentCheckpoint
	var randomFinish func(bool)
	var keptLog TurnLog
	keepingLogs := false
	endRandom := func(restore bool) {
		if randomFinish != nil {
			randomFinish(restore)
			randomFinish = nil
		}
	}
	defer func() { endRandom(false) }()
	reverse := func() actionOutcome {
		if checkpoint == nil {
			if inheritedOpaque && opaqueSelected {
				if log != nil {
					*log = keptLog
				}
				return actionReversed
			}
			panic("complete costed action lost its witness without an opaque activation")
		}
		reversible := checkpoint.reversible(g)
		if reversible {
			checkpoint.restore(g)
			endRandom(true)
			cancelAnnouncement(true)
			truncateAnnouncements()
			if log != nil {
				*log = savedLog
			}
		} else {
			// CR 734.1 permits keeping legal mana abilities. Keep all of them
			// rather than replaying an irreversible effect or inventing RNG.
			endRandom(false)
			cancelAnnouncement(false)
			g.Events = append(g.Events[:eventStart], g.Events[announcementEnd:]...)
			g.TriggerEventCursor = min(g.TriggerEventCursor, len(g.Events))
			g.CommanderEventCursor = min(g.CommanderEventCursor, len(g.Events))
			if log != nil {
				*log = keptLog
			}
		}
		return actionReversed
	}
	announcedLife := make(map[int]bool)
	if costs.ManaCost != nil {
		phyrexian := 0
		for index, symbol := range *costs.ManaCost {
			if isPhyrexianSymbol(symbol) {
				announcedLife[index] = partial.Phyrexian[phyrexian]
				phyrexian++
			}
		}
	}
	if manaCostIncludesManaPayment(costs.ManaCost, announcedLife) {
		for {
			planner = e.PaymentPlanner(g, log)
			plan, ok, _ = planner.Complete(request, partial)
			if !ok || !descriptor.cost.present(g) {
				return reverse()
			}
			steps, choices, opaque := e.costedManaOptions(g, player, request, partial)
			if len(choices) == 0 {
				break
			}
			if len(choices) == 1 && steps[0].Source == 0 {
				break
			}
			options := make([]game.ChoiceOption, len(choices))
			for index, choice := range choices {
				options[index] = game.ChoiceOption{Index: index, Label: choice.Label}
			}
			fallback := 0
			if steps[0].Source != 0 {
				for _, activation := range plan.Mana().Activations {
					key := payment.ManaAbilityKey{Source: activation.Source, AbilityIndex: activation.AbilityIndex}
					if _, fixed := partial.ManaAbilities[key]; fixed {
						continue
					}
					for index, step := range steps {
						if step.Source == key.Source && step.Index == key.AbilityIndex {
							fallback = index
							break
						}
					}
					break
				}
			}
			selected := e.chooseUnloggedChoice(g, agents, game.ChoiceRequest{
				Kind: game.ChoicePayment, Player: player, Prompt: "Activate a mana ability while paying?",
				Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{fallback}, Private: true,
			}, log)[0]
			if steps[selected].Source == 0 {
				for index, step := range steps {
					if step.Source != 0 && opaque[index].Kind != action.ActionActivateAbility {
						partial = partial.With(step, payment.Choice{UseManaAbility: false})
					}
				}
				break
			}
			if opaque[selected].Kind == action.ActionActivateAbility {
				opaqueSelected = true
				if log != nil && !keepingLogs {
					keptLog = cloneOpaqueLog(savedLog)
					keepingLogs = true
				}
				if checkpoint == nil && !inheritedOpaque {
					checkpoint = newOpaquePaymentCheckpoint(g)
					randomFinish = g.BookmarkRandom()
				}
				activation, _ := opaque[selected].ActivateAbilityPayload()
				start, entry := len(g.Events), -1
				var beforeLog TurnLog
				if log != nil {
					beforeLog = *log
					recorded := &ActionLog{Player: player, Action: opaque[selected]}
					recordActionSource(g, player, recorded, opaque[selected])
					entry = log.addPaymentManaActivation(recorded)
				}
				applied := e.applyActivateAbilityWithChoices(g, player, activation, agents, log)
				if log != nil {
					appendOpaqueLog(&keptLog, beforeLog, *log, !applied)
				}
				if !applied {
					return reverse()
				}
				recordActionManaTaps(g, log, entry, start)
				continue
			}
			partial = partial.With(steps[selected], choices[selected])
			partial, ok = e.choosePlannedManaPayment(g, player, request, partial, steps[selected], agents, log)
			if !ok {
				if checkpoint != nil || inheritedOpaque && opaqueSelected {
					return reverse()
				}
				panic("selected mana source lost its completion")
			}
		}
	}
	planner = e.PaymentPlanner(g, log)
	plan, ok, _ = planner.Complete(request, partial)
	if !ok || !descriptor.cost.present(g) {
		return reverse()
	}
	if descriptor.beforePayment != nil {
		descriptor.beforePayment(g, plan)
	}
	paymentStart := len(g.Events)
	paid, applied, reason := planner.Apply(plan)
	if !applied {
		if checkpoint != nil || inheritedOpaque && opaqueSelected {
			return reverse()
		}
		panic("prevalidated costed action failed: " + reason)
	}
	endRandom(false)
	if descriptor.commit != nil {
		descriptor.commit(g, paid, agents, log, paymentStart)
	}
	return actionApplied
}

func (e *Engine) choosePlannedManaPayment(g *game.Game, player game.PlayerID, request payment.Request, partial payment.Partial, selected payment.Step, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.Partial, bool) {
	defer g.OpenLibraryTopWindow()()
	planner := e.PaymentPlanner(g)
	ability := payment.ManaAbilityKey{Source: selected.Source, AbilityIndex: selected.Index}
	for occurrence := 0; ; occurrence++ {
		plan, ok, _ := planner.Complete(request, partial)
		if !ok {
			return partial, false
		}
		key := payment.ManaActivationKey{ManaAbilityKey: ability, Occurrence: occurrence}
		var activation payment.ManaActivation
		seen, found := 0, false
		for _, candidate := range plan.Mana().Activations {
			if candidate.Source != key.Source || candidate.AbilityIndex != key.AbilityIndex {
				continue
			}
			if seen == occurrence {
				activation, found = candidate, true
				break
			}
			seen++
		}
		if !found || activation.CostPlan == nil {
			return partial, true
		}
		if activation.SourceZone == zone.Hand {
			emitEvent(g, game.Event{Kind: game.EventCardRevealed, Announcement: true, Player: player, Controller: player, CardID: key.Source, FromZone: zone.Hand})
		}
		for index, additional := range activation.CostPlan.Costs().AdditionalCosts {
			step := payment.Step{Kind: payment.StepComponent, Index: index, ManaActivation: opt.Val(key)}
			for !partial.ManaPayments[key].Components[index].Finished {
				choice, valid := e.choosePaymentStep(g, player, planner, request, partial, step, payment.AdditionalCostText(additional), agents, log)
				if !valid {
					return partial, false
				}
				partial = partial.With(step, choice)
			}
		}
		output := payment.Step{Kind: payment.StepManaOutput, ManaActivation: opt.Val(key)}
		if activation.Choices != nil {
			prompt := "Choose a color"
			for _, candidate := range (&rulesPaymentState{g: g, engine: e}).ManaActivationCandidates(player) {
				if candidate.Source != key.Source || candidate.AbilityIndex != key.AbilityIndex {
					continue
				}
				for _, instruction := range candidate.Body.Content.Modes[0].Sequence {
					if choose, ok := instruction.Primitive.(game.Choose); ok && choose.Choice.Prompt != "" {
						prompt = choose.Choice.Prompt
					}
				}
				break
			}
			choice, valid := e.choosePaymentStep(g, player, planner, request, partial, output, prompt, agents, log)
			if !valid {
				return partial, false
			}
			partial = partial.With(output, choice)
		}
	}
}

func (e *Engine) choosePaymentStep(g *game.Game, player game.PlayerID, planner payment.Planner, request payment.Request, partial payment.Partial, step payment.Step, prompt string, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.Choice, bool) {
	choices := planner.Options(request, partial, step)
	if len(choices) == 0 {
		return payment.Choice{}, false
	}
	options := make([]game.ChoiceOption, len(choices))
	private := false
	for index, choice := range choices {
		label := choice.Label
		if label == "" {
			label = fmt.Sprint(choice.Value)
		}
		if step.Kind == payment.StepPhyrexian {
			label = "Pay with mana"
			if choice.PayLife {
				label = "Pay 2 life"
			}
		}
		options[index] = game.ChoiceOption{Index: index, Label: label}
		if objects := choice.Component.Objects; len(objects) > 0 {
			object := objects[len(objects)-1]
			if permanent, ok := permanentByObjectID(g, object); ok {
				options[index].Label, options[index].Card = permanentChoiceLabel(g, permanent), permanentChoiceInfo(g, permanent)
				options[index].Targets = []game.Target{game.PermanentTarget(object)}
			} else if card, ok := g.GetCardInstance(object); ok {
				owner := g.Players[card.Owner]
				private = private || owner.Hand.Contains(card.ID) || owner.Library.Contains(card.ID) || owner.Exile.IsFaceDown(card.ID)
				if card.Def != nil {
					options[index].Label = card.Def.Name
				}
				options[index].Card = cardChoiceInfo(g, card.ID)
			}
		}
		if counters := choice.Component.Counters; len(counters) > 0 {
			removal := counters[len(counters)-1]
			permanent, _ := permanentByObjectID(g, removal.Permanent)
			options[index].Label = counterChoiceLabel(permanentChoiceLabel(g, permanent), removal.Kind)
		}
	}
	selected := 0
	if len(choices) > 1 {
		selected = e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: player, Prompt: prompt, Options: options,
			MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0}, Private: private,
		}, log)[0]
	}
	return choices[selected], true
}

// costedManaOptions preserves each strategic source choice with a witness.
// Opaque activations alone are intentionally outside the planner's contract.
func (e *Engine) costedManaOptions(g *game.Game, player game.PlayerID, request payment.Request, partial payment.Partial) ([]payment.Step, []payment.Choice, []action.Action) {
	planner := e.PaymentPlanner(g)
	state := &rulesPaymentState{g: g, engine: e}
	var steps []payment.Step
	var choices []payment.Choice
	var opaque []action.Action
	finish := partial
	candidates := state.ManaActivationCandidates(player)
	seen := make(map[payment.ManaAbilityKey]bool)
	for _, candidate := range candidates {
		if candidate.Automatic {
			continue
		}
		key := payment.ManaAbilityKey{Source: candidate.Source, AbilityIndex: candidate.AbilityIndex}
		if _, fixed := partial.ManaAbilities[key]; !fixed {
			step := payment.Step{Kind: payment.StepManaAbility, Source: key.Source, Index: key.AbilityIndex}
			finish = finish.With(step, payment.Choice{UseManaAbility: false})
		}
	}
	if plan, ok, _ := planner.Complete(request, finish); ok {
		steps = append(steps, payment.Step{})
		choices = append(choices, payment.Choice{Label: "Finish paying", Plan: plan})
		opaque = append(opaque, action.Action{})
	}
	for _, candidate := range candidates {
		if candidate.Automatic {
			continue
		}
		key := payment.ManaAbilityKey{Source: candidate.Source, AbilityIndex: candidate.AbilityIndex}
		if seen[key] {
			continue
		}
		seen[key] = true
		if _, fixed := partial.ManaAbilities[key]; fixed {
			continue
		}
		step := payment.Step{Kind: payment.StepManaAbility, Source: key.Source, Index: key.AbilityIndex}
		for _, choice := range planner.Options(request, partial, step) {
			if choice.UseManaAbility {
				steps, choices, opaque = append(steps, step), append(choices, choice), append(opaque, action.Action{})
			}
		}
	}
	for _, act := range e.costedPaymentManaActions(g, player) {
		activation, _ := act.ActivateAbilityPayload()
		if !e.opaqueManaActivation(g, player, activation, candidates) {
			continue
		}
		label := paymentManaAbilityOptions(g, player, []action.Action{act})[1].Label
		steps = append(steps, payment.Step{Source: activation.SourceID, Index: activation.AbilityIndex})
		choices = append(choices, payment.Choice{Label: label + " (opaque)"})
		opaque = append(opaque, act)
	}
	return steps, choices, opaque
}

func (e *Engine) costedPaymentManaActions(g *game.Game, player game.PlayerID) []action.Action {
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		for index, ability := range permanentEffectiveAbilitiesView(g, permanent) {
			body, ok := ability.(*game.ManaAbility)
			if ok && manaAbilityAnnouncementLegal(g, player, permanent, body, index) {
				activation := action.ActivateAbilityAction{SourceID: permanent.ObjectID, AbilityIndex: index}
				if descriptor, valid := e.describeActivation(g, player, activation); valid && e.costedActionAvailable(g, player, descriptor.cost) {
					actions = append(actions, action.ActivateAbility(permanent.ObjectID, index, nil, 0))
				}
			}
		}
	}
	return append(actions, e.legalHandManaAbilityActions(g, player)...)
}

func (e *Engine) opaqueManaActivation(g *game.Game, player game.PlayerID, activation action.ActivateAbilityAction, candidates []payment.ManaActivationCandidate) bool {
	var body *game.ManaAbility
	if _, ability, ok := activatedAbilitySource(g, player, activation.SourceID, activation.AbilityIndex); ok {
		body, _ = ability.(*game.ManaAbility)
	} else if _, ability, ok := handManaAbilitySource(g, player, activation.SourceID, activation.AbilityIndex); ok {
		body = &ability
	}
	if body == nil || !manaBodyHasAddManaEffect(body) {
		return false
	}
	if _, class := payment.ClassifyManaAbility(body); class == payment.OpaqueManaAbility {
		return true
	}
	key := payment.ManaAbilityKey{Source: activation.SourceID, AbilityIndex: activation.AbilityIndex}
	return !e.PaymentPlanner(g).CanForecastMana(player, key)
}

func cloneOpaqueLog(log TurnLog) TurnLog {
	log.Entries = slices.Clone(log.Entries)
	log.Choices = slices.Clone(log.Choices)
	log.Draws = slices.Clone(log.Draws)
	log.Losses = slices.Clone(log.Losses)
	log.Actions = slices.Clone(log.Actions)
	log.Resolves = slices.Clone(log.Resolves)
	log.CombatDamage = slices.Clone(log.CombatDamage)
	log.CreatureDamage = slices.Clone(log.CreatureDamage)
	log.Deaths = slices.Clone(log.Deaths)
	log.Phases = slices.Clone(log.Phases)
	log.Steps = slices.Clone(log.Steps)
	log.Triggers = slices.Clone(log.Triggers)
	return log
}

func appendOpaqueLog(kept *TurnLog, before, after TurnLog, failed bool) {
	entries := after.Entries[len(before.Entries):]
	if failed && len(entries) > 0 {
		entries = entries[1:]
	}
	kept.Entries = append(kept.Entries, entries...)
	kept.Choices = append(kept.Choices, after.Choices[len(before.Choices):]...)
	kept.Draws = append(kept.Draws, after.Draws[len(before.Draws):]...)
	kept.Losses = append(kept.Losses, after.Losses[len(before.Losses):]...)
	kept.Actions = append(kept.Actions, after.Actions[len(before.Actions):]...)
	kept.Resolves = append(kept.Resolves, after.Resolves[len(before.Resolves):]...)
	kept.CombatDamage = append(kept.CombatDamage, after.CombatDamage[len(before.CombatDamage):]...)
	kept.CreatureDamage = append(kept.CreatureDamage, after.CreatureDamage[len(before.CreatureDamage):]...)
	kept.Deaths = append(kept.Deaths, after.Deaths[len(before.Deaths):]...)
	kept.Phases = append(kept.Phases, after.Phases[len(before.Phases):]...)
	kept.Steps = append(kept.Steps, after.Steps[len(before.Steps):]...)
	kept.Triggers = append(kept.Triggers, after.Triggers[len(before.Triggers):]...)
}

type opaquePaymentCheckpoint struct {
	world      *game.Game
	players    [game.NumPlayers]*game.Player
	cards      map[id.ID]*game.CardInstance
	permanents map[id.ID]*game.Permanent
	objects    map[id.ID]*game.StackObject
}

func newOpaquePaymentCheckpoint(g *game.Game) *opaquePaymentCheckpoint {
	checkpoint := &opaquePaymentCheckpoint{
		world: g.Clone(), players: g.Players, cards: maps.Clone(g.CardInstances),
		permanents: make(map[id.ID]*game.Permanent), objects: make(map[id.ID]*game.StackObject),
	}
	for _, permanent := range g.Battlefield {
		checkpoint.permanents[permanent.ObjectID] = permanent
	}
	for _, object := range g.Stack.Objects() {
		checkpoint.objects[object.ID] = object
	}
	return checkpoint
}

func (c *opaquePaymentCheckpoint) reversible(g *game.Game) bool {
	var libraries [game.NumPlayers]zone.Zone
	for player, saved := range c.world.Players {
		libraries[player] = saved.Library
	}
	return manaActivationIsReversible(g, libraries, len(c.world.Events))
}

func (c *opaquePaymentCheckpoint) restore(g *game.Game) {
	before := c.world
	for player, saved := range before.Players {
		*c.players[player] = *saved
		before.Players[player] = c.players[player]
	}
	for cardID, saved := range before.CardInstances {
		if live := c.cards[cardID]; live != nil {
			*live = *saved
			before.CardInstances[cardID] = live
		}
	}
	for index, saved := range before.Battlefield {
		if live := c.permanents[saved.ObjectID]; live != nil {
			*live = *saved
			before.Battlefield[index] = live
		}
	}
	var stack game.Stack
	for _, saved := range before.Stack.Objects() {
		if live := c.objects[saved.ID]; live != nil {
			*live = *saved
			stack.Push(live)
		} else {
			stack.Push(saved)
		}
	}
	before.Stack = stack
	g.RestoreOpaquePayment(before)
}
