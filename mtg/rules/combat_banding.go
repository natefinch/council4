package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// defendingPlayerAssignsAttackerDamage reports whether CR 702.22j transfers the
// division of a blocked attacker's combat damage to the defending player: an
// attacking creature being blocked by a creature with banding has its damage
// divided by the defending player, among any creatures blocking it, rather than
// by the active player (an exception to CR 510.1c).
func defendingPlayerAssignsAttackerDamage(g *game.Game, blockers []*game.Permanent) bool {
	return slices.ContainsFunc(blockers, func(blocker *game.Permanent) bool {
		return blocker != nil && hasKeyword(g, blocker, game.Banding)
	})
}

// activePlayerAssignsBlockerDamage reports whether CR 702.22k transfers the
// division of a blocking creature's combat damage to the active player: a
// blocking creature blocking a creature with banding has its damage divided by
// the active player among any creatures it's blocking, rather than by the
// defending player (an exception to CR 510.1d).
func activePlayerAssignsBlockerDamage(g *game.Game, attackers []*game.Permanent) bool {
	return slices.ContainsFunc(attackers, func(attacker *game.Permanent) bool {
		return attacker != nil && hasKeyword(g, attacker, game.Banding)
	})
}

// combatDamageChooser routes the combat damage-assignment decisions banding
// transfers between players (CR 702.22j-k) to the deciding player's agent. A
// chooser without an engine, or a player without a ChoiceAgent, gets each
// decision's deterministic default.
type combatDamageChooser struct {
	engine *Engine
	agents [game.NumPlayers]PlayerAgent
	log    *TurnLog
}

func (c combatDamageChooser) choose(g *game.Game, request game.ChoiceRequest) []int {
	if c.engine == nil {
		return fallbackChoice(request)
	}
	return c.engine.chooseChoice(g, c.agents, request, c.log)
}

// trampleExcessToAttackTarget asks the defending player, who chooses how a
// trampling attacker blocked by a creature with banding assigns its damage (CR
// 702.22j), how much of its damage beyond lethal to each blocker, counting the
// step's other assignments, is assigned to the attack target (CR 702.19b). The
// default assigns none, keeping all of the
// damage on the blockers.
func (c combatDamageChooser) trampleExcessToAttackTarget(g *game.Game, player game.PlayerID, attacker *game.Permanent, target game.AttackTarget, damage, maxExcess int) int {
	request := game.ChoiceRequest{
		Kind:             game.ChoiceNumber,
		Player:           player,
		Prompt:           fmt.Sprintf("How much of %s's %d combat damage is assigned to %s? Any amount above 0 requires lethal damage to each creature blocking it", permanentChoiceLabel(g, attacker), damage, attackTargetChoiceLabel(g, target)),
		NumberRange:      opt.Val(game.NumberRange{Minimum: 0, Maximum: opt.Val(maxExcess)}),
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
		Subject:          permanentChoiceInfo(g, attacker),
		CombatAttacker:   attacker.ObjectID,
	}
	selected := c.choose(g, request)
	if len(selected) != 1 {
		return 0
	}
	return min(max(selected[0], 0), maxExcess)
}

// divideCombatDamage has player divide source's combat damage among recipients
// in any amounts, including none, that total damage (CR 702.22j-k). With a
// single recipient there is only one division, so no decision is asked. The
// default answer is protectiveCombatDamageDivision. sourceAttacking reports
// whether source is the attacker (blocked by the recipients) or a blocker.
func (c combatDamageChooser) divideCombatDamage(g *game.Game, player game.PlayerID, source *game.Permanent, recipients []*game.Permanent, damage int, sourceAttacking bool) []creatureDamageAssignment {
	recipients = slices.DeleteFunc(slices.Clone(recipients), func(recipient *game.Permanent) bool {
		return recipient == nil
	})
	if damage <= 0 || len(recipients) == 0 {
		return nil
	}
	relation := "it's blocking"
	if sourceAttacking {
		relation = "blocking it"
	}
	prompt := fmt.Sprintf("Divide %s's %d combat damage among the creatures %s", permanentChoiceLabel(g, source), damage, relation)
	return c.chooseCombatDamageDivision(g, player, source, recipients, damage, sourceAttacking, prompt, protectiveCombatDamageDivision(g, source, recipients, damage))
}

// chooseCombatDamageDivision asks player to divide damage of source's combat
// damage among the non-nil recipients, offering defaultDivision (which must
// total damage among recipients). With a single recipient no decision is asked.
// Requests describe creatures only by their public characteristics, so a
// face-down creature's identity stays hidden (CR 708.2).
func (c combatDamageChooser) chooseCombatDamageDivision(g *game.Game, player game.PlayerID, source *game.Permanent, recipients []*game.Permanent, damage int, sourceAttacking bool, prompt string, defaultDivision []creatureDamageAssignment) []creatureDamageAssignment {
	if len(recipients) == 1 {
		return []creatureDamageAssignment{{permanent: recipients[0], damage: damage}}
	}
	options := make([]game.ChoiceOption, len(recipients))
	for i, recipient := range recipients {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, recipient), Card: permanentChoiceInfo(g, recipient)}
	}
	defaultSelection := make([]int, 0, damage)
	for _, assignment := range defaultDivision {
		index := slices.Index(recipients, assignment.permanent)
		for range assignment.damage {
			defaultSelection = append(defaultSelection, index)
		}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceCombatDamageDivision,
		Player:           player,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       damage,
		MaxChoices:       damage,
		DefaultSelection: defaultSelection,
		Subject:          permanentChoiceInfo(g, source),
	}
	if sourceAttacking {
		request.CombatAttacker = source.ObjectID
	} else {
		request.CombatBlocker = source.ObjectID
	}
	amounts := make([]int, len(recipients))
	for _, index := range c.choose(g, request) {
		if index >= 0 && index < len(recipients) {
			amounts[index]++
		}
	}
	var assignments []creatureDamageAssignment
	for i, recipient := range recipients {
		if amounts[i] > 0 {
			assignments = append(assignments, creatureDamageAssignment{permanent: recipient, damage: amounts[i]})
		}
	}
	return assignments
}

// attackTargetChoiceLabel labels what an attacker is attacking, by public
// characteristics only.
func attackTargetChoiceLabel(g *game.Game, target game.AttackTarget) string {
	if permanent, ok := attackTargetPermanent(g, target); ok {
		return permanentChoiceLabel(g, permanent)
	}
	return fmt.Sprintf("Player %d", target.Player+1)
}

// defenderAssignsAsThoughUnblocked asks the defending player, who chooses how
// an attacker blocked by a creature with banding assigns its damage (CR
// 702.22j), whether that attacker's "may assign its combat damage as though it
// weren't blocked" permission is used. The default keeps the damage on the
// blockers.
func (c combatDamageChooser) defenderAssignsAsThoughUnblocked(g *game.Game, player game.PlayerID, attacker *game.Permanent) bool {
	request := mayChoiceRequest(player, fmt.Sprintf("Have %s assign its combat damage as though it weren't blocked?", permanentChoiceLabel(g, attacker)))
	request.DefaultSelection = []int{0}
	request.Subject = permanentChoiceInfo(g, attacker)
	request.CombatAttacker = attacker.ObjectID
	selected := c.choose(g, request)
	return len(selected) == 1 && selected[0] == 1
}

// protectiveCombatDamageDivision is the default answer offered to a player who
// divides an opposing creature's combat damage among the creatures that player
// controls, as banding allows (CR 702.22j-k), and the division used when that
// player gives no valid answer. The damage is divided only among recipients, in
// any amounts that total the source's damage. The default protects the
// player's creatures: an indestructible recipient absorbs all of it; otherwise the
// damage is spread below lethal (heaviest share on the creatures that can take
// the most) when that absorbs it all, and when it can't, all of it goes to the
// single recipient with the lowest power so only one creature is dealt lethal
// damage. Lethal amounts account for deathtouch (CR 702.2c) and damage already
// marked.
func protectiveCombatDamageDivision(g *game.Game, source *game.Permanent, recipients []*game.Permanent, damage int) []creatureDamageAssignment {
	recipients = slices.DeleteFunc(slices.Clone(recipients), func(recipient *game.Permanent) bool {
		return recipient == nil
	})
	if damage <= 0 || len(recipients) == 0 {
		return nil
	}
	for _, recipient := range recipients {
		if hasKeyword(g, recipient, game.Indestructible) {
			return []creatureDamageAssignment{{permanent: recipient, damage: damage}}
		}
	}
	type sharedRecipient struct {
		permanent *game.Permanent
		capacity  int
	}
	shared := make([]sharedRecipient, 0, len(recipients))
	totalCapacity := 0
	for _, recipient := range recipients {
		capacity := max(0, lethalDamageRemainingFromSource(g, source, recipient)-1)
		shared = append(shared, sharedRecipient{permanent: recipient, capacity: capacity})
		totalCapacity += capacity
	}
	if totalCapacity >= damage {
		slices.SortStableFunc(shared, func(a, b sharedRecipient) int {
			return b.capacity - a.capacity
		})
		var assignments []creatureDamageAssignment
		remaining := damage
		for _, recipient := range shared {
			if remaining <= 0 {
				break
			}
			assigned := min(remaining, recipient.capacity)
			if assigned <= 0 {
				continue
			}
			assignments = append(assignments, creatureDamageAssignment{permanent: recipient.permanent, damage: assigned})
			remaining -= assigned
		}
		return assignments
	}
	sacrificed := recipients[0]
	for _, recipient := range recipients[1:] {
		if effectivePower(g, recipient) < effectivePower(g, sacrificed) {
			sacrificed = recipient
		}
	}
	return []creatureDamageAssignment{{permanent: sacrificed, damage: damage}}
}
