package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// sharedCreatureTypeAttackingCount builds the Shared Animosity dynamic +1/+0-per-
// other-attacking-creature-sharing-a-type amount ("it gets +1/+0 until end of
// turn for each other attacking creature that shares a creature type with it.").
func sharedCreatureTypeAttackingCount() game.DynamicAmount {
	return game.DynamicAmount{
		Kind:       game.DynamicAmountSharedCreatureTypeCountInGroup,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		}),
	}
}

// TestSharedAnimosityPumpScalesWithSharedTypeAttackers proves that the Shared
// Animosity trigger ("Whenever a creature you control attacks, it gets +1/+0
// until end of turn for each other attacking creature that shares a creature
// type with it.") pumps the triggering attacker by the number of OTHER ATTACKING
// creatures that share a creature type with it: non-attacking creatures of the
// same type, attackers of a different type, and the attacker itself are excluded.
func TestSharedAnimosityPumpScalesWithSharedTypeAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchantment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Shared Animosity",
		Types: []types.Card{types.Enchantment},
	}})

	attacker := goblinWithPT("Attacking Goblin", 2, 2)
	attackerPermanent := addCombatPermanent(g, game.Player1, attacker)
	goblinAlly := addCombatPermanent(g, game.Player1, goblinWithPT("Attacking Goblin Ally", 1, 1))
	goblinFoe := addCombatPermanent(g, game.Player2, goblinWithPT("Attacking Goblin Foe", 1, 1))

	// A Zombie attacker shares no creature type with the Goblins, so it is not
	// counted.
	zombie := creatureWithPT("Attacking Zombie", 2, 2)
	zombie.Subtypes = []types.Sub{types.Zombie}
	zombiePermanent := addCombatPermanent(g, game.Player1, zombie)

	// A bench Goblin shares a type but is not attacking, so it is not counted.
	addCombatPermanent(g, game.Player1, goblinWithPT("Bench Goblin", 1, 1))

	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attackerPermanent.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: goblinAlly.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: goblinFoe.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
			{Attacker: zombiePermanent.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	obj := &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        enchantment.ObjectID,
		SourceCardID:    enchantment.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: attackerPermanent.ObjectID},
	}
	count := sharedCreatureTypeAttackingCount()
	resolveInstruction(engine, g, obj, game.ModifyPT{
		Object:         game.EventPermanentReference(),
		PowerDelta:     game.Dynamic(count),
		ToughnessDelta: game.Fixed(0),
		Duration:       game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	// Two other attacking Goblins (the ally and the foe across controllers) share
	// a creature type; the Zombie attacker and the benched Goblin do not count.
	if got := effectivePower(g, attackerPermanent); got != 4 {
		t.Fatalf("effective power = %d, want 4 (2 base + 2 shared-type attackers)", got)
	}
	if got, _ := effectiveToughness(g, attackerPermanent); got != 2 {
		t.Fatalf("effective toughness = %d, want 2 (+0 toughness)", got)
	}

	expireCleanupDurations(g)
	if got := effectivePower(g, attackerPermanent); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want 2 (buff expired)", got)
	}
}
