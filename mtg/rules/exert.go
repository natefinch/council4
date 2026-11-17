package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// exertPermanent is the shared exert operation (CR 701.43): playerID chooses
// to have p not untap during playerID's next untap step. Exerting is legal
// whether or not p is tapped and whether or not it was already exerted; a
// repeated exert by the same player expires in the same untap step, so it
// changes nothing (CR 701.43b). An object that isn't on the battlefield can't
// be exerted (CR 701.43c), and a phased-out permanent is treated as though it
// doesn't exist (CR 702.26b).
func exertPermanent(g *game.Game, playerID game.PlayerID, p *game.Permanent) bool {
	if !activeBattlefieldPermanent(p) || playerID < 0 || int(playerID) >= game.NumPlayers {
		return false
	}
	if current, ok := permanentByObjectID(g, p.ObjectID); !ok || current != p {
		return false
	}
	p.ExertedBy[playerID] = true
	return true
}

// consumeExertions ends every exert effect created by activePlayer as that
// player's untap step untaps permanents. Each one expires now whether or not
// the permanent was tapped (CR 701.43a-b), including on permanents another
// player controls, which the active player's untap step doesn't untap anyway.
// The returned permanents must not untap during this untap step.
func consumeExertions(g *game.Game, activePlayer game.PlayerID) map[*game.Permanent]bool {
	var exerted map[*game.Permanent]bool
	for _, permanent := range g.Battlefield {
		if permanent == nil || !permanent.ExertedBy[activePlayer] {
			continue
		}
		permanent.ExertedBy[activePlayer] = false
		if exerted == nil {
			exerted = make(map[*game.Permanent]bool)
		}
		exerted[permanent] = true
	}
	return exerted
}

type attackExertOption struct {
	attacker *game.Permanent
	cost     game.ExertAttackCost
	text     string
}

// chooseAttackExertions offers every "You may exert this creature as it
// attacks" optional cost of the declared attackers (CR 508.1g, 701.43d) in one
// choice. Each instance on each attacker is a separate option; a creature put
// onto the battlefield attacking was never declared, so it is never offered.
// Declining is the default. Each paid instance exerts the attacker for the
// attacking player and queues that instance's linked "When you do" trigger
// (CR 603.12, 607.2h), which is put on the stack with the other attack
// triggers. The trigger is queued even when it would have no legal target; it
// is then removed as it would be put on the stack (CR 603.3d).
func (ce combatEngine) chooseAttackExertions(g *game.Game, playerID game.PlayerID, attackers []game.AttackDeclaration, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	var offered []attackExertOption
	for _, declaration := range attackers {
		attacker, ok := permanentByObjectID(g, declaration.Attacker)
		if !ok || !activeBattlefieldPermanent(attacker) || effectiveController(g, attacker) != playerID {
			continue
		}
		for _, body := range permanentEffectiveAbilitiesView(g, attacker) {
			static, ok := body.(*game.StaticAbility)
			if !ok || !static.ExertAsItAttacks.Exists {
				continue
			}
			offered = append(offered, attackExertOption{attacker: attacker, cost: static.ExertAsItAttacks.Val, text: static.Text})
		}
	}
	if len(offered) == 0 {
		return
	}
	options := make([]game.ChoiceOption, len(offered))
	for i, option := range offered {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: "Exert " + permanentChoiceLabel(g, option.attacker),
			Card:  permanentChoiceInfo(g, option.attacker),
		}
	}
	selected := ce.e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:       game.ChoicePayment,
		Player:     playerID,
		Prompt:     "Choose attacking creatures to exert",
		Options:    options,
		MinChoices: 0,
		MaxChoices: len(options),
	}, log)
	seen := make(map[int]bool, len(selected))
	for _, index := range selected {
		if index < 0 || index >= len(offered) || seen[index] {
			continue
		}
		seen[index] = true
		option := offered[index]
		if !exertPermanent(g, playerID, option.attacker) || !option.cost.WhenYouDo.Exists {
			continue
		}
		g.PendingReflexiveTriggers = append(g.PendingReflexiveTriggers, game.ReflexiveTrigger{
			SourceID:       option.attacker.CardInstanceID,
			SourceObjectID: option.attacker.ObjectID,
			SourceTokenDef: option.attacker.TokenDef,
			Controller:     playerID,
			Ability: game.TriggeredAbility{
				Text:    option.text,
				Content: option.cost.WhenYouDo.Val,
			},
		})
	}
}
