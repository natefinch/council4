package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/eval"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// runPriorityLoop runs the priority sequence for the current step or phase
// (CR 117). Each iteration is a point at which a player would receive priority,
// so it first applies state-based actions and puts triggered abilities on the
// stack, repeating until neither does anything (CR 117.5), before a player acts.
//
// Priority then moves per CR 117.3: the player with priority acts and keeps
// priority after casting a spell, activating an ability, or taking a special
// action (CR 117.3c); passing hands priority to the next player (CR 117.3d).
// When every active player passes in succession (CR 117.4), the top of the
// stack resolves and the active player receives priority again (CR 117.3b), or,
// if the stack is empty, the step or phase ends and the loop returns.
func (e *Engine) runPriorityLoop(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	consecutivePasses := 0

	for {
		// CR 117.5 / CR 603.3b: before a player gets priority, perform all
		// state-based actions (repeating until none apply), then put any waiting
		// triggered abilities on the stack; if either happened, restart the
		// check. Once the stack is stable the appropriate player gets priority.
		e.applyStateBasedActionsWithLog(g, log)
		if g.IsGameOver() {
			return
		}
		if e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
			// CR 603.3b: after triggered abilities are put on the stack, the
			// player who would have received priority gets it. That player is
			// already identified by g.Turn.PriorityPlayer — the active player at
			// the start of a step or after a resolution (CR 117.3a/b), or a
			// player who kept priority after acting (CR 117.3c) — so it is left
			// unchanged here. A new object on the stack restarts the all-passed
			// count.
			consecutivePasses = 0
			continue
		}

		activePlayers := g.TurnOrder.ActivePlayerCount()
		if activePlayers <= 0 {
			return
		}

		playerID := g.Turn.PriorityPlayer
		if !canAct(g, playerID) {
			g.Turn.PriorityPlayer = g.TurnOrder.NextPriority(playerID)
			continue
		}

		legal := e.legalActions(g, playerID)
		if len(legal) == 0 {
			legal = []action.Action{actionBuild.pass()}
		}

		chosen := actionBuild.pass()
		if agent := agentFor(agents, playerID); agent != nil {
			chosen = e.decideAction(g, agent, playerID, legal)
		}
		if !containsAction(legal, chosen) {
			chosen = actionBuild.pass()
		}

		actionLog := &ActionLog{
			Player: playerID,
			Action: chosen,
		}
		recordActionSource(g, playerID, actionLog, chosen)
		log.addAction(actionLog)
		entryIndex := lastEntryIndex(log)
		eventsBefore := len(g.Events)

		if !e.applyActionWithChoices(g, playerID, chosen, agents, log) {
			panic("applyAction failed for validated action")
		}
		recordActionManaTaps(g, log, entryIndex, eventsBefore)
		recordLandEnteredTapped(g, log, entryIndex, chosen)
		if chosen.Kind != action.ActionPass {
			e.notifyActionObservers(g, agents, playerID, chosen)
		}
		if chosen.Kind == action.ActionPass {
			consecutivePasses++
			// CR 117.4: when every active player has passed in succession, the
			// top of the stack resolves (or the step/phase ends if the stack is
			// empty). CR 117.3b: the active player receives priority after a
			// spell or ability resolves.
			if consecutivePasses >= activePlayers {
				if g.Stack.IsEmpty() {
					return
				}
				e.resolveTopOfStackWithChoices(g, agents, log)
				consecutivePasses = 0
				g.Turn.PriorityPlayer = g.Turn.ActivePlayer
				continue
			}
			// CR 117.3d: a passing player hands priority to the next player.
			g.Turn.PriorityPlayer = g.TurnOrder.NextPriority(playerID)
			continue
		}

		// CR 117.3c: a player who took an action (not a pass) keeps priority.
		consecutivePasses = 0
		g.Turn.PriorityPlayer = playerID
	}
}

func observe(g *game.Game, playerID game.PlayerID) PlayerObservation {
	return NewObservation(g, playerID)
}

// recordActionSource snapshots the permanent that is the source of an action so
// the turn log can attribute the action to a named card, and flags activations
// of mana abilities. Only activated abilities need the source snapshot: their
// source is a battlefield permanent identified by object ID, whereas a land
// played or spell cast from hand is already identified by its card instance ID
// in the action payload. The snapshot is a no-op when the source is not a
// battlefield permanent (for example an ability activated from a card in hand).
func recordActionSource(g *game.Game, playerID game.PlayerID, actionLog *ActionLog, a action.Action) {
	payload, ok := a.ActivateAbilityPayload()
	if !ok {
		return
	}
	actionLog.addPermanentSnapshot(g, payload.SourceID)
	actionLog.ManaAbility = isManaAbilityActivation(g, playerID, payload)
	actionLog.AbilityText = activatedAbilityText(g, playerID, payload)
	actionLog.AbilityEffectSummary = activatedAbilityEffectSummary(g, playerID, payload)
}

// recordLandEnteredTapped flags a play-land action whose land entered the
// battlefield tapped, on that action's log entry. It is called just after the
// land enters, before the controller taps it for mana, so the permanent's tapped
// state reflects entry (for example a tapland or a conditional "enters tapped
// unless ..."). It writes to the logged entry by index because addAction stored
// a copy of the action log before the land entered. It is a no-op for any other
// action.
func recordLandEnteredTapped(g *game.Game, log *TurnLog, entryIndex int, a action.Action) {
	if log == nil || entryIndex < 0 || entryIndex >= len(log.Entries) {
		return
	}
	payload, ok := a.PlayLandPayload()
	if !ok {
		return
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == payload.CardID {
			if permanent.Tapped {
				log.Entries[entryIndex].Action.LandEnteredTapped = true
			}
			return
		}
	}
}

// activatedAbilityEffectSummary returns a short value-oriented gloss of what the
// activated ability costs and does, derived from the scorable-effect IR, or an
// empty string when the ability cannot be resolved or the IR models none of its
// effects.
func activatedAbilityEffectSummary(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) string {
	if _, body, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex); ok {
		return eval.Describe(eval.ScorableAbilityOfModes(body, activate.ChosenModes))
	}
	if _, body, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex); ok {
		return eval.Describe(eval.ScorableAbilityOfModes(&body, activate.ChosenModes))
	}
	return ""
}

// activatedAbilityText returns the rules text describing the ability the action
// activates, whether it is printed on a battlefield permanent or on a card
// activated from hand (for example cycling). Generated card defs omit per-ability
// text, so it falls back to the source's full oracle text. It returns "" when no
// text is available.
func activatedAbilityText(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) string {
	if _, body, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex); ok {
		if text := strings.TrimSpace(game.BodyText(body)); text != "" {
			return text
		}
		if permanent, ok := permanentByObjectID(g, activate.SourceID); ok {
			if def, ok := permanentFaceDef(g, permanent); ok {
				return strings.TrimSpace(def.OracleText)
			}
		}
		return ""
	}
	if card, body, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex); ok {
		if text := strings.TrimSpace(body.Text); text != "" {
			return text
		}
		if card != nil && card.Def != nil {
			return strings.TrimSpace(card.Def.OracleText)
		}
	}
	return ""
}

// lastEntryIndex returns the index of the most recently appended turn-log entry,
// or -1 when there is none (for example when log is nil and addAction was a
// no-op).
func lastEntryIndex(log *TurnLog) int {
	if log == nil {
		return -1
	}
	return len(log.Entries) - 1
}

// recordActionManaTaps attributes the permanents tapped for mana while applying
// an action to that action's log entry, so a report can show how a spell or
// ability was paid for, including lands tapped during cost payment. It scans the
// events emitted since eventsBefore for tapped-for-mana taps.
func recordActionManaTaps(g *game.Game, log *TurnLog, entryIndex, eventsBefore int) {
	if log == nil || entryIndex < 0 || entryIndex >= len(log.Entries) {
		return
	}
	var taps []ManaTap
	for i := eventsBefore; i < len(g.Events); i++ {
		event := g.Events[i]
		if event.Kind != game.EventPermanentTapped || !event.TappedForMana {
			continue
		}
		taps = append(taps, ManaTap{
			Source: tappedManaSourceName(g, event.PermanentID),
			Colors: manaColorCodes(event.ProducedManaColors),
		})
	}
	if len(taps) > 0 {
		log.Entries[entryIndex].Action.ManaTaps = taps
	}
}

// tappedManaSourceName resolves the display name of a tapped mana source.
func tappedManaSourceName(g *game.Game, permanentID id.ID) string {
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok {
		return ""
	}
	if permanent.Token {
		return permanentTokenName(permanent)
	}
	return permanentEffectiveName(g, permanent)
}

// manaColorCodes converts produced mana colors to their string codes.
func manaColorCodes(colors []mana.Color) []string {
	if len(colors) == 0 {
		return nil
	}
	codes := make([]string, 0, len(colors))
	for _, c := range colors {
		codes = append(codes, string(c))
	}
	return codes
}

func agentFor(agents [game.NumPlayers]PlayerAgent, playerID game.PlayerID) PlayerAgent {
	if playerID < 0 || int(playerID) >= len(agents) {
		return nil
	}
	return agents[playerID]
}

func containsAction(actions []action.Action, want action.Action) bool {
	for _, act := range actions {
		if actionsEqual(act, want) {
			return true
		}
	}
	return false
}

func actionsEqual(a, b action.Action) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case action.ActionPass:
		return true
	case action.ActionPlayLand:
		aPayload, aOK := a.PlayLandPayload()
		bPayload, bOK := b.PlayLandPayload()
		return aOK && bOK && aPayload == bPayload
	case action.ActionCastSpell:
		aPayload, aOK := a.CastSpellPayload()
		bPayload, bOK := b.CastSpellPayload()
		return aOK && bOK &&
			aPayload.CardID == bPayload.CardID &&
			aPayload.SourceZone == bPayload.SourceZone &&
			aPayload.Face == bPayload.Face &&
			aPayload.XValue == bPayload.XValue &&
			aPayload.KickerPaid == bPayload.KickerPaid &&
			aPayload.Bargained == bPayload.Bargained &&
			aPayload.Offspring == bPayload.Offspring &&
			aPayload.Bestowed == bPayload.Bestowed &&
			aPayload.GiftPromised == bPayload.GiftPromised &&
			aPayload.GiftRecipient == bPayload.GiftRecipient &&
			aPayload.Overloaded == bPayload.Overloaded &&
			aPayload.Mutate == bPayload.Mutate &&
			aPayload.MutateTargetID == bPayload.MutateTargetID &&
			slices.Equal(aPayload.Targets, bPayload.Targets) &&
			slices.Equal(aPayload.ChosenModes, bPayload.ChosenModes)
	case action.ActionActivateAbility:
		aPayload, aOK := a.ActivateAbilityPayload()
		bPayload, bOK := b.ActivateAbilityPayload()
		return aOK && bOK &&
			aPayload.SourceID == bPayload.SourceID &&
			aPayload.AbilityIndex == bPayload.AbilityIndex &&
			aPayload.XValue == bPayload.XValue &&
			slices.Equal(aPayload.Targets, bPayload.Targets) &&
			slices.Equal(aPayload.TargetCounts, bPayload.TargetCounts) &&
			slices.Equal(aPayload.ChosenModes, bPayload.ChosenModes)
	case action.ActionSuspendCard:
		aPayload, aOK := a.SuspendCardPayload()
		bPayload, bOK := b.SuspendCardPayload()
		return aOK && bOK && aPayload == bPayload
	case action.ActionPlotCard:
		aPayload, aOK := a.PlotCardPayload()
		bPayload, bOK := b.PlotCardPayload()
		return aOK && bOK && aPayload == bPayload
	case action.ActionForetellCard:
		aPayload, aOK := a.ForetellCardPayload()
		bPayload, bOK := b.ForetellCardPayload()
		return aOK && bOK && aPayload == bPayload
	case action.ActionDeclareAttackers:
		aPayload, aOK := a.DeclareAttackersPayload()
		bPayload, bOK := b.DeclareAttackersPayload()
		return aOK && bOK && slices.Equal(aPayload.Attackers, bPayload.Attackers)
	case action.ActionDeclareBlockers:
		aPayload, aOK := a.DeclareBlockersPayload()
		bPayload, bOK := b.DeclareBlockersPayload()
		return aOK && bOK && slices.Equal(aPayload.Blockers, bPayload.Blockers)
	default:
		return false
	}
}
