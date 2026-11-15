package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestResourceViewIsolatesProjectedRecordsAndTransientFrames(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	p := &Permanent{ObjectID: 1, Controller: Player1, Owner: Player1}
	p.Counters.Add(counter.Charge, 1)
	g.Battlefield = []*Permanent{p, {ObjectID: 2}}
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	g.choiceCtx = "live choice context"
	g.pendingEntryEvents = []Event{{Kind: EventPermanentEnteredBattlefield}}
	view := g.ResourceView(ResourceProjection{
		Permanents: map[id.ID]PermanentResources{1: {Tapped: true, Counters: map[counter.Kind]int{counter.Charge: 2}}, 2: {Removed: true}},
		Players:    map[PlayerID]PlayerResources{Player1: {Life: 3, Mana: map[mana.Unit]int{{Color: mana.U}: 1}}},
		Used:       map[ActivatedAbilityUse]bool{{SourceID: 1, AbilityIndex: 0}: true},
		Events:     []Event{{Kind: EventLifeLost, Player: Player1, Amount: 1}},
	})
	if view.staticFrame != nil || view.choiceCtx != nil || len(view.pendingEntryEvents) != 0 || g.staticFrame == nil {
		t.Fatal("view shared transient evaluation/choice frames")
	}
	if len(view.Battlefield) != 1 || !view.Battlefield[0].Tapped || view.Battlefield[0].Counters.Get(counter.Charge) != 2 ||
		view.Players[0].Life != 3 || view.Players[0].ManaPool.Amount(mana.U) != 1 {
		t.Fatal("projected resources were not exposed to readers")
	}
	view.Battlefield[0].Counters.Add(counter.Charge, 1)
	view.Players[0].ManaPool.Add(mana.U, 1)
	if p.Tapped || p.Counters.Get(counter.Charge) != 1 || g.Players[0].Life != 40 || g.Players[0].ManaPool.Total() != 0 ||
		len(g.Battlefield) != 2 || len(g.Events) != 0 || len(g.ActivatedAbilitiesThisTurn) != 0 {
		t.Fatal("projected resource storage leaked into the original game")
	}
}
