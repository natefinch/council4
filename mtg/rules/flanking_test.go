package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func flankingCreature(name string) *game.CardDef {
	def := vanillaCreature(name, 2, 2)
	def.TriggeredAbilities = []game.TriggeredAbility{game.FlankingTriggeredBody}
	return def
}

// TestFlankingTriggerMatchesBlockerWithoutFlanking covers the Flanking keyword
// (CR 702.25): the becomes-blocked trigger fires only when the blocking
// creature lacks flanking. A blocker that itself has flanking is excluded by
// the "without flanking" filter (Flanking does not stack against other
// flankers), and the related-subject filter resolves the BLOCKER rather than
// the attacker.
func TestFlankingTriggerMatchesBlockerWithoutFlanking(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, flankingCreature("Flanker"))
	plainBlocker := addCombatPermanent(g, game.Player2, vanillaCreature("Footsoldier", 2, 2))
	flankingBlocker := addCombatPermanent(g, game.Player2, flankingCreature("Cavalier"))
	pattern := &game.FlankingTriggeredBody.Trigger.Pattern

	if !triggerMatchesEvent(g, attacker, pattern, game.Event{
		Kind:               game.EventAttackerBecameBlocked,
		Controller:         game.Player1,
		PermanentID:        attacker.ObjectID,
		RelatedPermanentID: plainBlocker.ObjectID,
	}) {
		t.Fatal("flanking did not trigger against a blocker without flanking")
	}

	if triggerMatchesEvent(g, attacker, pattern, game.Event{
		Kind:               game.EventAttackerBecameBlocked,
		Controller:         game.Player1,
		PermanentID:        attacker.ObjectID,
		RelatedPermanentID: flankingBlocker.ObjectID,
	}) {
		t.Fatal("flanking wrongly triggered against a blocker that has flanking")
	}
}

// TestFlankingDebuffsTheBlocker confirms the canonical effect references the
// blocking creature (the event's related permanent), not the attacker, so the
// -1/-1 lands on the blocker.
func TestFlankingDebuffsTheBlocker(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, flankingCreature("Flanker"))
	blocker := addCombatPermanent(g, game.Player2, vanillaCreature("Footsoldier", 2, 2))

	modify, ok := game.FlankingTriggeredBody.Content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("flanking content is not a ModifyPT: %#v", game.FlankingTriggeredBody.Content.Modes[0].Sequence[0].Primitive)
	}
	stackObject := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:               game.EventAttackerBecameBlocked,
			Controller:         game.Player1,
			PermanentID:        attacker.ObjectID,
			RelatedPermanentID: blocker.ObjectID,
		},
	}
	resolver := newReferenceResolver(g, stackObject)
	resolved, ok := resolver.object(modify.Object)
	if !ok || resolved.permanent == nil {
		t.Fatal("flanking effect object reference did not resolve to a permanent")
	}
	if resolved.permanent.ObjectID != blocker.ObjectID {
		t.Fatalf("flanking debuff targeted %v, want blocker %v", resolved.permanent.ObjectID, blocker.ObjectID)
	}
}
