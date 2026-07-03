package rules

import "github.com/natefinch/council4/mtg/game"

// CombatView is a read-only snapshot of the current combat. It is empty outside
// the combat phase (no attackers). It exposes every declared attacker with its
// effective characteristics, whom it is attacking, and the blockers assigned to
// it so far, so an agent can reason about incoming damage, lethal, and blocks.
type CombatView struct {
	// Attackers lists every creature currently attacking, in declaration order.
	Attackers []AttackerView
}

// AttackerView describes one attacking creature: its effective characteristics
// (via the shared PermanentView), whom it is attacking, and the blockers already
// assigned to it.
type AttackerView struct {
	// Attacker is the attacking creature's effective public characteristics.
	Attacker PermanentView
	// DefendingPlayer is the player this attacker is attacking, or in whose
	// direction it attacks when its target is a planeswalker or battle (CR 508.4).
	DefendingPlayer game.PlayerID
	// AttacksPlayerDirectly is true when the attacker is attacking a player
	// directly rather than a planeswalker or battle, so its damage lands on that
	// player's life total (and, for a commander, as commander damage).
	AttacksPlayerDirectly bool
	// Blocked reports whether the attacker has been blocked this combat (it stays
	// true even if its blockers later leave combat, CR 509.1).
	Blocked bool
	// Blockers lists the creatures currently blocking this attacker.
	Blockers []PermanentView
}

// Combat returns a snapshot of the current combat. Outside combat, or before any
// attackers are declared, the returned view has no attackers.
func (o PlayerObservation) Combat() CombatView {
	var view CombatView
	if o.g.Combat == nil {
		return view
	}
	combat := o.g.Combat
	for i := range combat.Attackers {
		declaration := combat.Attackers[i]
		attacker, ok := permanentByObjectID(o.g, declaration.Attacker)
		if !ok {
			continue
		}
		attackerView := AttackerView{
			Attacker:              o.permanentView(attacker),
			DefendingPlayer:       declaration.Target.Player,
			AttacksPlayerDirectly: declaration.Target.IsPlayerAttack(),
			Blocked:               combat.BlockedAttackers[declaration.Attacker],
		}
		for j := range combat.Blockers {
			if combat.Blockers[j].Blocking != declaration.Attacker {
				continue
			}
			if blocker, ok := permanentByObjectID(o.g, combat.Blockers[j].Blocker); ok {
				attackerView.Blockers = append(attackerView.Blockers, o.permanentView(blocker))
			}
		}
		view.Attackers = append(view.Attackers, attackerView)
	}
	return view
}

// AttackersAgainst returns the attackers currently attacking the given player
// directly — the ones whose combat damage would reduce that player's life. It is
// a convenience over Combat for computing incoming damage and lethal.
func (o PlayerObservation) AttackersAgainst(playerID game.PlayerID) []AttackerView {
	combat := o.Combat()
	var against []AttackerView
	for i := range combat.Attackers {
		if combat.Attackers[i].AttacksPlayerDirectly && combat.Attackers[i].DefendingPlayer == playerID {
			against = append(against, combat.Attackers[i])
		}
	}
	return against
}
