package game

import "github.com/natefinch/council4/mtg/game/id"

// AttackTarget specifies what an attacking creature is attacking.
// In Commander, a creature can attack a player, a planeswalker,
// or a battle.
type AttackTarget struct {
	// NoTarget records that the attacked permanent left combat after attackers
	// were declared. The attacker remains attacking but attacks nothing; Player
	// retains the original defending player for blocker declaration.
	NoTarget bool

	// Player is the PlayerID being attacked or originally defending an attacker
	// whose permanent target left combat. Even when attacking a planeswalker or
	// battle, the creature is attacking "in the direction of" a player.
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
	return !at.NoTarget && at.PlaneswalkerID == 0 && at.BattleID == 0
}

// AttackDeclaration records that a creature has been declared as an attacker.
type AttackDeclaration struct {
	// Attacker is the ObjectID of the attacking creature.
	Attacker id.ID

	// Target specifies what is being attacked.
	Target AttackTarget
}

// AttackingBand records creatures the active player announced as one attacking
// band while declaring attackers (CR 508.1e, 702.22c). Members are the
// attacking creatures' ObjectIDs.
type AttackingBand struct {
	Members []id.ID
}

// CloneAttackingBands returns a deep copy of bands.
func CloneAttackingBands(bands []AttackingBand) []AttackingBand {
	if bands == nil {
		return nil
	}
	out := make([]AttackingBand, len(bands))
	for i, band := range bands {
		out[i] = AttackingBand{Members: append([]id.ID(nil), band.Members...)}
	}
	return out
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
	// Completion is separate from positive attack history (CR 508.8), including
	// an empty declaration. These windows do not reopen during priority.
	AttackDeclarationComplete bool
	BlockDeclarationsComplete [NumPlayers]bool
	// FirstStrikeDamageStarted distinguishes a two-step combat from one with
	// only ordinary damage. Eligibility is captured before the first assignment.
	FirstStrikeDamageStarted bool
	// FirstStrikeEligible records the original ObjectIDs that had first or
	// double strike as the first damage step began (CR 510.4).
	FirstStrikeEligible map[id.ID]bool
	// Attackers lists the current attackers and their targets. A creature
	// removed from combat, including by leaving the battlefield, is removed from
	// this list (CR 506.4); AttackersDeclared and recorded events keep history.
	Attackers []AttackDeclaration

	// Bands lists the attacking bands announced with the attack declaration
	// (CR 702.22c-d). A band lasts for the rest of combat even if its members
	// later lose banding (CR 702.22e). A member removed from combat leaves its
	// band, and a band left with one member is no longer a band (CR 702.22f).
	// Creatures put onto the battlefield attacking never join a band.
	Bands []AttackingBand

	// AttackersDeclared records whether any creature was declared as an attacker
	// or put onto the battlefield attacking this combat (CR 508.8). It stays true
	// even if those creatures later leave combat, so the declare blockers and
	// combat damage steps are not skipped once a creature has attacked. Attackers
	// itself can shrink as attackers leave combat, so it can't answer this.
	AttackersDeclared bool
	// PlayersAttacked records defending players attacked this combat, including
	// attacks on permanents they control and creatures entering attacking.
	PlayersAttacked map[PlayerID]bool

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

	// BlockDeclarationBatchID groups blocker events from the one simultaneous
	// declare-blockers turn-based action, including declarations made by
	// different defending players.
	BlockDeclarationBatchID id.ID
}
