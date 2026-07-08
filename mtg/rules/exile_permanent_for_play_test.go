package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestExilePermanentForPlayGrantsOwnerPlayPermission proves Prowl, Stoic
// Strategist's exile clause: exiling a target permanent another player owns
// removes it from the battlefield, moves the card to that owner's exile, and
// grants only the owner permission to play it while it remains exiled. The
// resolving controller (an opponent of the owner) gains no permission. The
// exiled card also joins the source-keyed linked pool for the plays-a-card
// trigger.
func TestExilePermanentForPlayGrantsOwnerPlayPermission(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	const prowlLink = game.LinkedKey("prowl-exile")
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Prowl Source",
		Types: []types.Card{types.Creature},
	}})
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Victim",
		Types: []types.Card{types.Creature},
	}})
	victimCardID := victim.CardInstanceID

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		InlineTrigger: &game.TriggeredAbility{
			Content: game.Mode{
				Targets: []game.TargetSpec{{MinTargets: 0, MaxTargets: 1, Constraint: "creature"}},
				Sequence: []game.Instruction{{
					Primitive: game.ExilePermanentForPlay{
						Object:    game.TargetPermanentReference(0),
						LinkedKey: prowlLink,
					},
				}},
			}.Ability(),
		},
		Targets: []game.Target{game.PermanentTarget(victim.ObjectID)},
	}

	engine.resolveStackObject(g, obj, &TurnLog{})

	if permanentForCard(g, victimCardID) != nil {
		t.Fatal("victim was not removed from the battlefield")
	}
	if !g.Players[game.Player2].Exile.Contains(victimCardID) {
		t.Fatal("victim card was not moved to its owner's exile")
	}
	if !canCastFromZoneByRuleEffect(g, game.Player2, victimCardID, zone.Exile, game.FaceFront) {
		t.Fatal("owner (Player2) was not granted permission to play the exiled card")
	}
	if canCastFromZoneByRuleEffect(g, game.Player1, victimCardID, zone.Exile, game.FaceFront) {
		t.Fatal("resolving controller (Player1) must not gain permission to play the owner's card")
	}
	key := game.LinkedObjectKey{SourceID: source.CardInstanceID, LinkID: string(prowlLink)}
	if !cardInLinkedObjectPool(g, key, victimCardID) {
		t.Fatal("exiled card was not published to the source-keyed linked pool")
	}
}
