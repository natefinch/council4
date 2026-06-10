package game

import "github.com/natefinch/council4/mtg/game/id"

// AttackTarget specifies what an attacking creature is attacking.
// In Commander, a creature can attack a player, a planeswalker,
// or a battle.
type AttackTarget struct {
	// Player is the PlayerID being attacked. Always set — even when
	// attacking a planeswalker or battle, the creature is attacking
	// "in the direction of" a player.
	Player PlayerID

	// PlaneswalkerID is the ObjectID of the planeswalker being attacked.
	// Zero if attacking a player directly.
	PlaneswalkerID id.ID

	// BattleID is the ObjectID of the battle being attacked.
	// Zero if not attacking a battle.
	BattleID id.ID
}

// IsPlayerAttack reports whether this attack targets a player directly
// (not a planeswalker or battle).
func (at AttackTarget) IsPlayerAttack() bool {
	return at.PlaneswalkerID == 0 && at.BattleID == 0
}

// AttackDeclaration records that a creature has been declared as an attacker.
type AttackDeclaration struct {
	// Attacker is the ObjectID of the attacking creature.
	Attacker id.ID

	// Target specifies what is being attacked.
	Target AttackTarget
}

// BlockDeclaration records that a creature has been declared as a blocker.
type BlockDeclaration struct {
	// Blocker is the ObjectID of the blocking creature.
	Blocker id.ID

	// Blocking is the ObjectID of the attacking creature being blocked.
	Blocking id.ID
}

// CombatState tracks the current combat — attackers, blockers, and
// damage assignment order. It exists only during the combat phase
// and is nil otherwise.
type CombatState struct {
	// Attackers lists all declared attackers and their targets.
	Attackers []AttackDeclaration

	// Blockers lists all declared blockers and which attacker they block.
	Blockers []BlockDeclaration

	// BlockedAttackers preserves which attackers became blocked even if all of
	// their blockers later leave combat.
	BlockedAttackers map[id.ID]bool

	// BlockerOrder maps each attacker's ObjectID to the ordered list of
	// its blockers for damage assignment (CR 510.1c). The attacking
	// player chooses this order.
	BlockerOrder map[id.ID][]id.ID

	// DamageAssignment maps each creature (attacker or blocker) to the
	// amount of damage it deals in combat. Populated during the combat
	// damage step.
	DamageAssignment map[id.ID]int
}
