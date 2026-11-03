package rules

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func reversibleManaCheckpoint(g *game.Game, log *TurnLog) func() {
	before := g.Clone()
	players, cards := g.Players, maps.Clone(g.CardInstances)
	permanents := make(map[id.ID]*game.Permanent, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		permanents[permanent.ObjectID] = permanent
	}
	objects := make(map[id.ID]*game.StackObject)
	for _, object := range g.Stack.Objects() {
		objects[object.ID] = object
	}
	entries := 0
	if log != nil {
		entries = len(log.Entries)
	}
	return func() {
		// Keep synchronous Engine references alive, including a search's Player
		// and a proposed spell's StackObject. Clone's RNG is not a continuation;
		// reversible activations did not consume the live random stream.
		for player, saved := range before.Players {
			if saved != nil && players[player] != nil {
				*players[player] = *saved
				before.Players[player] = players[player]
			}
		}
		for cardID, saved := range before.CardInstances {
			if live := cards[cardID]; live != nil {
				*live = *saved
				before.CardInstances[cardID] = live
			}
		}
		for index, saved := range before.Battlefield {
			if live := permanents[saved.ObjectID]; live != nil {
				*live = *saved
				before.Battlefield[index] = live
			}
		}
		var stack game.Stack
		for _, saved := range before.Stack.Objects() {
			if live := objects[saved.ID]; live != nil {
				*live = *saved
				stack.Push(live)
			} else {
				stack.Push(saved)
			}
		}
		before.Stack = stack
		g.RestoreReversibleManaPayment(before)
		if log != nil {
			// Choices remain recorded so failed proposals replay identically.
			log.Entries = log.Entries[:entries]
		}
	}
}

func manaActivationIsReversible(g *game.Game, beforeLibraries [game.NumPlayers]zone.Zone, eventStart int) bool {
	for player, library := range beforeLibraries {
		current := &g.Players[player].Library
		if library.ReorderVersion() != current.ReorderVersion() || !slices.Equal(library.All(), current.All()) {
			return false
		}
	}
	for _, event := range g.Events[eventStart:] {
		if event.Kind == game.EventCardRevealed && event.FromZone == zone.Library ||
			event.Kind == game.EventZoneChanged && (event.FromZone == zone.Library || event.ToZone == zone.Library) {
			return false
		}
	}
	return true
}

// manaPaymentStep records one mana ability activated while paying a cost
// (CR 601.2g) so that an unpaid cost can reverse it (CR 733.1).
type manaPaymentStep struct {
	// restore returns the game to just before the activation.
	restore    func()
	activation action.ActivateAbilityAction
	// reversible is false for an activation that moved a card to or from a
	// library, shuffled one or revealed a card from one (CR 733.1).
	reversible bool
	// spentMana reports whether the activation spent mana from a pool, which
	// may have been mana from an earlier activation.
	spentMana bool
	choices   []game.ChoiceDecision
	entries   []TurnLogEntry
	events    []replayableManaEvent
}

// undoManualPaymentMana reverses the mana abilities in steps, activated in
// that order while paying a cost that then went unpaid (CR 733.1). Each
// reversible activation is reversed unless mana it made may have been spent on
// an activation that can't be: one spending mana pins every earlier one. The
// irreversible activations keep the cards they moved, shuffled or revealed, and
// what they produced, so a reversible activation that came before one is
// reversed by returning to before it and repeating only the irreversible
// activations with the choices made for them. When a repeat can't reproduce an
// activation exactly, the reversible activations before the last irreversible
// one are kept instead, which CR 733.1 also permits.
func (e *Engine) undoManualPaymentMana(g *game.Game, player game.PlayerID, steps []manaPaymentStep, log *TurnLog) {
	pinned, last := -1, -1
	for i, step := range steps {
		if !step.reversible {
			last = i
			if step.spentMana {
				pinned = i
			}
		}
	}
	first := -1
	for i := pinned + 1; i < len(steps); i++ {
		if steps[i].reversible {
			first = i
			break
		}
	}
	if first < 0 {
		return
	}
	if last < first {
		steps[first].restore()
		return
	}
	if last+1 < len(steps) {
		steps[last+1].restore()
	}
	entries := slices.Clone(log.Entries)
	keep := reversibleManaCheckpoint(g, log)
	var libraries [game.NumPlayers][]id.ID
	for owner := range libraries {
		libraries[owner] = g.Players[owner].Library.All()
	}
	steps[first].restore()
	kept := steps[first : last+1]
	if e.repeatIrreversibleManaSteps(g, player, kept, libraries) {
		for _, step := range kept {
			if !step.reversible {
				log.Entries = append(log.Entries, step.entries...)
			}
		}
		return
	}
	keep()
	log.Entries = entries
}

// repeatIrreversibleManaSteps activates again, in order, the irreversible
// activations among steps with the choices recorded for them. It reports
// whether each produced the same events, asking the same choices, and every
// library ended as libraries.
func (e *Engine) repeatIrreversibleManaSteps(g *game.Game, player game.PlayerID, steps []manaPaymentStep, libraries [game.NumPlayers][]id.ID) bool {
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	if previous, ok := g.ChoiceContext(); ok {
		defer g.SetChoiceContext(previous)
	} else {
		defer g.ClearChoiceContext()
	}
	for _, step := range steps {
		if step.reversible {
			continue
		}
		script := &manaStepScript{choices: step.choices}
		var agents [game.NumPlayers]PlayerAgent
		for i := range agents {
			agents[i] = script
		}
		scratch := &TurnLog{}
		e.setReplacementChoiceContext(g, agents, scratch)
		eventStart := len(g.Events)
		if !e.applyActivateAbilityWithChoices(g, player, step.activation, agents, scratch) || !script.complete() ||
			!slices.Equal(replayableManaEvents(g.Events[eventStart:]), step.events) {
			return false
		}
	}
	for owner, library := range libraries {
		if !slices.Equal(g.Players[owner].Library.All(), library) {
			return false
		}
	}
	return true
}

// manaStepScript answers the choices of a repeated mana ability activation
// with the selections recorded for it, noting any request that differs.
type manaStepScript struct {
	choices  []game.ChoiceDecision
	next     int
	mismatch bool
}

func (*manaStepScript) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (s *manaStepScript) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if s.next >= len(s.choices) {
		s.mismatch = true
		return fallbackChoice(request)
	}
	recorded := s.choices[s.next]
	s.next++
	if recorded.Request.Player != request.Player || recorded.Request.Kind != request.Kind ||
		len(recorded.Request.Options) != len(request.Options) || !choiceSelectionValid(request, recorded.Selected) {
		s.mismatch = true
		return fallbackChoice(request)
	}
	return slices.Clone(recorded.Selected)
}

func (s *manaStepScript) complete() bool {
	return !s.mismatch && s.next == len(s.choices)
}

// replayableManaEvent is the part of an event that a repeated activation must
// reproduce: what happened, to which objects and players, between which zones.
type replayableManaEvent struct {
	kind          game.EventKind
	source        id.ID
	sourceObject  id.ID
	card          id.ID
	permanent     id.ID
	player        game.PlayerID
	controller    game.PlayerID
	from, to      zone.Type
	amount        int
	abilityIndex  int
	manaAbility   bool
	tappedForMana bool
}

func replayableManaEvents(events []game.Event) []replayableManaEvent {
	replayable := make([]replayableManaEvent, len(events))
	for i, event := range events {
		replayable[i] = replayableManaEvent{
			kind: event.Kind, source: event.SourceID, sourceObject: event.SourceObjectID,
			card: event.CardID, permanent: event.PermanentID, player: event.Player, controller: event.Controller,
			from: event.FromZone, to: event.ToZone, amount: event.Amount, abilityIndex: event.AbilityIndex,
			manaAbility: event.ManaAbility, tappedForMana: event.TappedForMana,
		}
	}
	return replayable
}

// manaSpent is the total mana spent from every player's pool so far.
func manaSpent(g *game.Game) int {
	total := 0
	for _, player := range g.Players {
		if player != nil {
			total += player.ManaPool.Spent()
		}
	}
	return total
}
