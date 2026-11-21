package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

const (
	assistPlayerPrompt       = "Choose a player to assist"
	assistManaPrompt         = "Activate a mana ability for Assist?"
	assistCasterManaPrompt   = "Activate a mana ability while casting with Assist?"
	assistContributionPrompt = "How much generic mana will you pay for Assist?"
)

func (e *Engine) chooseAssistPlayer(g *game.Game, caster game.PlayerID, plan payment.Plan, agents [game.NumPlayers]PlayerAgent, log *TurnLog) opt.V[game.PlayerID] {
	options := []game.ChoiceOption{{Index: 0, Label: "Cast without assistance"}}
	helpers := []opt.V[game.PlayerID]{{}}
	preferred, fallback := plan.SuggestedAssistant(), 0
	for _, p := range g.Players {
		if p == nil || p.ID == caster || p.Eliminated {
			continue
		}
		if preferred.Exists && preferred.Val == p.ID {
			fallback = len(options)
		}
		options = append(options, game.ChoiceOption{Index: len(options), Label: fmt.Sprintf("Player %d", p.ID+1), Targets: []game.Target{game.PlayerTarget(p.ID)}})
		helpers = append(helpers, opt.Val(p.ID))
	}
	if len(options) == 1 {
		return helpers[0]
	}
	index := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoicePlayer, Player: caster, Prompt: assistPlayerPrompt,
		Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{fallback},
	}, log)[0]
	return helpers[index]
}

func (e *Engine) assistManaWindow(g *game.Game, player game.PlayerID, prompt string, suggest func() opt.V[payment.ManaActivation],
	activate func(game.PlayerID, action.Action, opt.V[payment.ManaActivation]) bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	for {
		actions := e.costedPaymentManaActions(g, player)
		if len(actions) == 0 {
			return true
		}
		fallback := 0
		witness := suggest()
		if witness.Exists {
			for i, act := range actions {
				ability, _ := act.ActivateAbilityPayload()
				if ability.SourceID == witness.Val.Source && ability.AbilityIndex == witness.Val.AbilityIndex {
					fallback = i + 1
					break
				}
			}
		}
		// Candidate actions can identify cards in the payer's hand.
		selected, usedFallback := e.chooseUnloggedChoiceWithFallback(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: player, Prompt: prompt,
			Options: paymentManaAbilityOptions(g, player, actions), MinChoices: 1, MaxChoices: 1,
			DefaultSelection: []int{fallback}, Private: true,
		}, log)
		index := selected[0]
		if index == 0 {
			return true
		}
		// A valid agent selection still gets its own activation cost and
		// output choices, even if it chose the suggested source.
		if !usedFallback || index != fallback {
			witness = opt.V[payment.ManaActivation]{}
		}
		if !activate(player, actions[index-1], witness) {
			return false
		}
	}
}

func (e *Engine) chooseAssistContribution(g *game.Game, plan payment.Plan, helper game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) int {
	planner := e.PaymentPlanner(g)
	limit := planner.AssistPoolLimit(plan, helper)
	if limit == 0 {
		return 0
	}
	fallback := 0
	if plan.SuggestedAssistant().Exists {
		fallback = limit
	}
	return e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoiceNumber, Player: helper, Prompt: assistContributionPrompt, MinChoices: 1, MaxChoices: 1,
		NumberRange: opt.Val(game.NumberRange{Maximum: opt.Val(limit)}), DefaultSelection: []int{fallback},
	}, log)[0]
}

func (e *Engine) activateAssistWitness(g *game.Game, player game.PlayerID, witness payment.ManaActivation, log, keptLog *TurnLog) bool {
	state := &rulesPaymentState{g: g, engine: e, log: log}
	if log == nil {
		return state.ActivatePlannedMana(player, witness)
	}
	before := *log
	applied := state.ActivatePlannedMana(player, witness)
	if applied {
		appendOpaqueLog(keptLog, before, *log, false)
	}
	return applied
}

func (e *Engine) activateLivePaymentMana(g *game.Game, player game.PlayerID, act action.Action, agents [game.NumPlayers]PlayerAgent, log, keptLog *TurnLog) bool {
	activation, _ := act.ActivateAbilityPayload()
	start, entry := len(g.Events), -1
	var before TurnLog
	if log != nil {
		before = *log
		recorded := &ActionLog{Player: player, DecisionMaker: g.DecisionMaker(player, game.DecisionRules), Action: act}
		recordActionSource(g, player, recorded, act)
		entry = log.addPaymentManaActivation(recorded)
	}
	applied := e.applyActivateAbilityWithChoices(g, player, activation, agents, log)
	if log != nil {
		appendOpaqueLog(keptLog, before, *log, !applied)
	}
	if applied {
		recordActionManaTaps(g, log, entry, start)
	}
	return applied
}
