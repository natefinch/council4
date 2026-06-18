package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// recordingAgent passes priority but records every action it is notified about
// via the optional ActionObserver hook.
type recordingAgent struct {
	seat     game.PlayerID
	observed []observedAction
}

type observedAction struct {
	actor game.PlayerID
	kind  action.Kind
	act   action.Action
}

func (*recordingAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *recordingAgent) ObserveAction(actor game.PlayerID, act action.Action, obs PlayerObservation) {
	if obs.Player != a.seat {
		panic("observer received an observation for the wrong seat")
	}
	a.observed = append(a.observed, observedAction{actor: actor, kind: act.Kind, act: act})
}

// plainAgent implements only PlayerAgent, not ActionObserver, so it must never
// be notified.
type plainAgent struct{}

func (plainAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return actionBuild.pass()
}

func TestNotifyActionObserversNotifiesOtherSeatsOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	observer1 := &recordingAgent{seat: game.Player1}
	observer3 := &recordingAgent{seat: game.Player3}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: observer1,
		game.Player2: plainAgent{},
		game.Player3: observer3,
		game.Player4: plainAgent{},
	}

	playLand := actionBuild.playLand(addCardToHand(g, game.Player2, basicLand()), game.FaceFront)
	engine.notifyActionObservers(g, agents, game.Player2, playLand)

	// The actor (Player2) is a plainAgent anyway, but every other observing seat
	// must be notified of Player2's action.
	if len(observer1.observed) != 1 || observer1.observed[0].actor != game.Player2 {
		t.Errorf("observer1.observed = %+v, want one action by Player2", observer1.observed)
	}
	if observer1.observed[0].kind != action.ActionPlayLand {
		t.Errorf("observed kind = %v, want ActionPlayLand", observer1.observed[0].kind)
	}
	if len(observer3.observed) != 1 || observer3.observed[0].actor != game.Player2 {
		t.Errorf("observer3.observed = %+v, want one action by Player2", observer3.observed)
	}
}

func TestNotifyActionObserversSkipsActorAndNonObservers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	actor := &recordingAgent{seat: game.Player1}
	other := &recordingAgent{seat: game.Player2}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: actor,
		game.Player2: other,
		game.Player3: plainAgent{},
		game.Player4: nil,
	}

	engine.notifyActionObservers(g, agents, game.Player1, actionBuild.pass())

	if len(actor.observed) != 0 {
		t.Errorf("actor was notified of its own action: %+v", actor.observed)
	}
	if len(other.observed) != 1 {
		t.Errorf("other.observed = %+v, want one notification", other.observed)
	}
}

func TestPriorityLoopNotifiesObserversOfRealActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Give Player1 a land to play so a non-pass action occurs during its turn.
	addCardToHand(g, game.Player1, basicLand())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	observer2 := &recordingAgent{seat: game.Player2}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: landPlayingAgent{},
		game.Player2: observer2,
		game.Player3: plainAgent{},
		game.Player4: plainAgent{},
	}

	log := &TurnLog{}
	engine.runPriorityLoop(g, agents, log)

	sawPlayLand := false
	for _, observed := range observer2.observed {
		if observed.kind == action.ActionPass {
			t.Errorf("observer was notified of a pass action: %+v", observer2.observed)
		}
		if observed.actor == game.Player1 && observed.kind == action.ActionPlayLand {
			sawPlayLand = true
		}
	}
	if !sawPlayLand {
		t.Errorf("observer2 did not see Player1 play a land: %+v", observer2.observed)
	}
}

func TestNotifyActionObserversRedactsFaceDownCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// A face-down cast action carries the real card id; observers must not learn
	// the hidden card's identity.
	hiddenID := addCardToHand(g, game.Player2, vanillaCreature("Hidden Bomb", 7, 7, game.Flying))
	faceDownCast := actionBuild.castFaceDown(hiddenID, game.FaceFront, game.FaceDownMorph)

	observer := &recordingAgent{seat: game.Player1}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: observer}

	engine.notifyActionObservers(g, agents, game.Player2, faceDownCast)

	if len(observer.observed) != 1 {
		t.Fatalf("observer.observed = %+v, want one notification", observer.observed)
	}
	got := observer.observed[0]
	payload, ok := got.act.CastFaceDownPayload()
	if !ok {
		t.Fatal("observed action is not a face-down cast")
	}
	if payload.CardID == hiddenID {
		t.Error("face-down cast leaked the hidden card id to an observer")
	}
	if payload.CardID != 0 {
		t.Errorf("redacted CardID = %v, want 0", payload.CardID)
	}
	if payload.FaceDownKind != game.FaceDownMorph {
		t.Errorf("FaceDownKind = %v, want FaceDownMorph (public)", payload.FaceDownKind)
	}
}

// landPlayingAgent plays the first available land action, then passes.
type landPlayingAgent struct{}

func (landPlayingAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if act.Kind == action.ActionPlayLand {
			return act
		}
	}
	return actionBuild.pass()
}
