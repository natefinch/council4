package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// landGraveyardReanimationPattern mirrors the lowered Hedge Shredder trigger
// "Whenever one or more land cards are put into your graveyard from your
// library, ...".
func landGraveyardReanimationPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:            game.EventZoneChanged,
		Player:           game.TriggerPlayerYou,
		MatchFromZone:    true,
		FromZone:         zone.Library,
		MatchToZone:      true,
		ToZone:           zone.Graveyard,
		OneOrMore:        true,
		SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
	}
}

func landDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}}
}

// TestLandMilledIntoGraveyardBatchReanimatesTapped drives Hedge Shredder's
// "Whenever one or more land cards are put into your graveyard from your
// library, put them onto the battlefield tapped" ability end to end: milling
// two lands and a creature coalesces into one land trigger, and resolving it
// returns only the two land cards to the battlefield tapped while the creature
// and an unrelated graveyard land stay put.
func TestLandMilledIntoGraveyardBatchReanimatesTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, landGraveyardReanimationPattern(),
		[]game.Instruction{{Primitive: game.MassReturnFromGraveyard{
			Player:           game.ControllerReference(),
			Destination:      zone.Battlefield,
			EntryTapped:      true,
			FromTriggerBatch: true,
		}}}, nil)

	forestA := addCardToLibrary(g, game.Player1, landDef("Forest A"))
	forestB := addCardToLibrary(g, game.Player1, landDef("Forest B"))
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Grizzly Bears", Types: []types.Card{types.Creature},
	}})

	// A land already in the graveyard before the trigger must not be reanimated:
	// only the freshly milled batch is "them".
	stale := addCardToLibrary(g, game.Player1, landDef("Stale Wastes"))
	g.Players[game.Player1].Library.Remove(stale)
	g.Players[game.Player1].Graveyard.Add(stale)

	milled := millCards(g, game.Player1, 3)
	if len(milled) != 3 {
		t.Fatalf("millCards returned %d cards, want 3", len(milled))
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("land-into-graveyard trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced land trigger", got)
	}
	engine.resolveTopOfStack(g, nil)

	onBattlefield := map[game.PlayerID]int{}
	tappedForests := 0
	for _, permanent := range g.Battlefield {
		switch permanent.CardInstanceID {
		case forestA, forestB:
			onBattlefield[permanent.Controller]++
			if permanent.Tapped {
				tappedForests++
			}
		case bear, stale:
			t.Fatalf("card %v was reanimated but should remain in the graveyard", permanent.CardInstanceID)
		default:
		}
	}
	if onBattlefield[game.Player1] != 2 {
		t.Fatalf("Player1 controls %d reanimated lands, want 2", onBattlefield[game.Player1])
	}
	if tappedForests != 2 {
		t.Fatalf("%d reanimated lands entered tapped, want 2", tappedForests)
	}
	if !g.Players[game.Player1].Graveyard.Contains(bear) {
		t.Fatal("milled creature should remain in the graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(stale) {
		t.Fatal("pre-existing graveyard land should not be reanimated")
	}
	if g.Players[game.Player1].Graveyard.Contains(forestA) ||
		g.Players[game.Player1].Graveyard.Contains(forestB) {
		t.Fatal("reanimated lands should have left the graveyard")
	}
}
