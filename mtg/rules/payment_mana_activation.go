package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func (s *rulesPaymentState) ActivatePlannedMana(player game.PlayerID, activation payment.ManaActivation) bool {
	s.assertWritable()
	if s.engine == nil || activation.CostPlan == nil {
		return false
	}
	g := s.g
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	start := len(g.Events)
	var actionLog *ActionLog
	if s.log != nil {
		chosen := action.ActivateAbility(activation.Source, activation.AbilityIndex, nil, 0)
		actionLog = &ActionLog{Player: player, DecisionMaker: g.DecisionMaker(player, game.DecisionRules), Action: chosen}
		recordActionSource(g, player, actionLog, chosen)
	}
	var obj *game.StackObject
	var body *game.ManaAbility
	var permanent *game.Permanent
	var source manaProducedSource
	if activation.SourceZone == zone.Hand {
		card, ability, ok := handManaAbilitySource(g, player, activation.Source, activation.AbilityIndex)
		if !ok || !canActivateHandManaAbility(g, player, card.ID, &ability, activation.AbilityIndex) {
			return false
		}
		body = &ability
		obj = &game.StackObject{Kind: game.StackActivatedAbility, SourceID: card.ID, SourceCardID: card.ID, SourceZone: zone.Exile, AbilityIndex: activation.AbilityIndex, Controller: player}
	} else {
		p, ability, ok := activatedAbilitySource(g, player, activation.Source, activation.AbilityIndex)
		if !ok {
			return false
		}
		var manaOK bool
		body, manaOK = ability.(*game.ManaAbility)
		if !manaOK || !canActivateManaAbility(g, player, p, body, activation.AbilityIndex) {
			return false
		}
		permanent = p
		source = captureManaProducedSource(g, p)
		obj = &game.StackObject{Kind: game.StackActivatedAbility, SourceID: p.ObjectID, SourceCardID: p.CardInstanceID, SourceTokenDef: p.TokenDef, Face: p.Face, AbilityIndex: activation.AbilityIndex, Controller: player}
		seedEntryChoices(obj, p)
	}
	costPaid, ok, _ := s.engine.PaymentPlanner(g).Apply(*activation.CostPlan)
	if !ok {
		return false
	}
	entry := -1
	if s.log != nil {
		entry = s.log.addPaymentManaActivation(actionLog)
	}
	obj.ID = g.IDGen.Next()
	obj.PaidCostSubjects = costPaid.Ability.PaidCostSubjects
	s.engine.resolveManaAbility(g, obj, permanent, body, source, start, activation.Choices, [game.NumPlayers]PlayerAgent{}, s.log)
	recordActionManaTaps(g, s.log, entry, start)
	return true
}

// resolveManaAbility is the real immediate mana-ability operation used both by
// standalone actions and by a prevalidated payment plan (CR 605.3b).
func (e *Engine) resolveManaAbility(g *game.Game, obj *game.StackObject, permanent *game.Permanent, body *game.ManaAbility, source manaProducedSource, start int, choices map[game.ChoiceKey]game.ResolutionChoiceResult, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	before := manaPoolColorSnapshot(g, obj.Controller)
	content := body.Content
	if choices != nil {
		content.Modes = slices.Clone(content.Modes)
		sequence := make([]game.Instruction, 0, len(content.Modes[0].Sequence))
		for _, instruction := range content.Modes[0].Sequence {
			if choice, ok := instruction.Primitive.(game.Choose); ok {
				value, selected := choices[choice.PublishChoice]
				if !selected {
					panic("planned mana choice missing")
				}
				rememberResolutionChoice(obj, string(choice.PublishChoice), value)
				continue
			}
			sequence = append(sequence, instruction)
		}
		content.Modes[0].Sequence = sequence
	}
	e.resolveAbilityContentWithChoices(g, obj, content, agents, log)
	produced := producedManaColorsSince(g, obj.Controller, before)
	if permanent != nil {
		recordTappedForManaProduced(g, permanent.ObjectID, produced)
		emitManaProducedEvent(g, source, obj.Controller, produced, producedManaAmountSince(g, obj.Controller, before), manaAbilityTappedSourceSince(g, permanent.ObjectID, start))
	} else {
		emitManaProducedEvent(g, manaProducedSource{sourceID: obj.SourceCardID, controller: obj.Controller}, obj.Controller, produced, producedManaAmountSince(g, obj.Controller, before), false)
	}
	var permanentID id.ID
	if permanent != nil {
		permanentID = permanent.ObjectID
	}
	emitAbilityActivatedEvent(g, obj, permanentID, true)
	recordActivatedAbilityUse(g, obj.SourceID, obj.AbilityIndex, body.Timing)
	e.resolveImmediateManaTriggers(g, start, agents, log)
}
