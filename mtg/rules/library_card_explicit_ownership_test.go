package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestCompiledLibraryCardExplicitObservationOwnership(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		owner      game.PlayerID
		firstLook  bool
		targeted   bool
		latest     bool
		linked     bool
		moveSecond bool
	}{
		{name: "look before reveal", text: "Look at the top card of your library. Reveal the top card of target opponent's library. Put the looked-at card into your hand.", owner: game.Player1, firstLook: true},
		{name: "reveal before look", text: "Reveal the top card of your library. Look at the top card of target opponent's library. Put the revealed card into your hand.", owner: game.Player1},
		{name: "looked card and its owner", text: "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand.", owner: game.Player2, firstLook: true, targeted: true},
		{name: "revealed card and its owner", text: "Reveal the top card of target player's library. Look at the top card of target opponent's library. Put the revealed card into that player's hand.", owner: game.Player2, targeted: true},
		{name: "generic card remains latest", text: "Look at the top card of your library. Reveal the top card of target opponent's library. Put that card into your hand.", owner: game.Player1, firstLook: true, latest: true},
		{name: "linked reveal preserves card", text: "Look at the top card of your library. Reveal it. Put the revealed card into your hand.", owner: game.Player1, firstLook: true, linked: true},
		{name: "linked reveal of earlier look", text: "Look at the top card of your library. Reveal the top card of target opponent's library. Reveal the looked-at card. Put the revealed card into your hand.", owner: game.Player1, firstLook: true, linked: true},
		{name: "independent products keep owners", text: "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand. Put the revealed card into that player's graveyard.", owner: game.Player2, firstLook: true, targeted: true, moveSecond: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Explicit Observation Runtime", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addCardToLibrary(g, tc.owner, vanillaCreature("Card A", 2, 3))
			second := addCardToLibrary(g, game.Player3, vanillaCreature("Card B", 4, 5))
			decoy := addCardToLibrary(g, game.Player4, vanillaCreature("Decoy", 6, 7))
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player3)}}
			if tc.targeted {
				obj.Targets = []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want, owner := first, tc.owner
			if tc.latest {
				want, owner = second, game.Player3
			}
			if !g.Players[owner].Hand.Contains(want) ||
				g.Players[tc.owner].Library.Contains(first) != tc.latest ||
				g.Players[game.Player3].Library.Contains(second) != (!tc.latest && !tc.moveSecond) ||
				g.Players[game.Player3].Graveyard.Contains(second) != tc.moveSecond ||
				!g.Players[game.Player4].Library.Contains(decoy) {
				t.Fatalf("wrong actual card/owner: first hand/library=%v/%v, second hand/library=%v/%v; want card %v owner %v",
					g.Players[tc.owner].Hand.Contains(first), g.Players[tc.owner].Library.Contains(first),
					g.Players[game.Player3].Hand.Contains(second), g.Players[game.Player3].Library.Contains(second), want, owner)
			}
			for _, event := range g.Events {
				if event.Kind != game.EventCardRevealed {
					continue
				}
				if event.CardID == first && event.Player != tc.owner || event.CardID == second && event.Player != game.Player3 {
					t.Fatalf("revealed card owner replaced by observer: %#v", event)
				}
			}
			if eventRevealedCard(g, first, obj.ID) != (!tc.firstLook || tc.linked) {
				t.Fatal("explicit producer identity changed the actual first-card reveal")
			}
			if len(g.LinkedObjects) != 0 || len(obj.LocalLinkedProducts) != 0 {
				t.Fatal("independent observation products escaped their sequence")
			}
			var untouched id.ID
			if tc.latest {
				untouched = first
			} else {
				untouched = second
			}
			for _, player := range g.Players {
				if player.Hand.Contains(untouched) {
					t.Fatal("the unrelated observation was also moved")
				}
			}
		})
	}
}
