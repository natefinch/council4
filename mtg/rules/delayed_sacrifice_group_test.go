package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDelayedCapturedGroupSacrificesOnlyPublishedPermanents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Orthion",
		Types: []types.Card{types.Creature},
	}})
	copyA := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	copyB := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	unrelated := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}
	const link = "delayed-sacrifice-1"
	key := linkedObjectSourceKey(g, obj, link)
	rememberLinkedObject(g, key, permanentLinkedObjectRef(copyA))
	rememberLinkedObject(g, key, permanentLinkedObjectRef(copyB))
	if !scheduleDelayedTrigger(g, obj, &game.DelayedTriggerDef{
		Timing:              game.DelayedAtBeginningOfNextEndStep,
		CapturedObjectGroup: opt.Val(game.LinkedObjectReference(link)),
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Sacrifice{
			Group: game.CapturedObjectsGroup(),
		}}}}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	for name, permanent := range map[string]*game.Permanent{"copy A": copyA, "copy B": copyB} {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
			t.Errorf("%s was not sacrificed", name)
		}
	}
	if _, ok := permanentByObjectID(g, unrelated.ObjectID); !ok {
		t.Error("unrelated permanent was sacrificed")
	}
}
