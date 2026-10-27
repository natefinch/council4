package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledHandLibraryChosenX(t *testing.T) {
	t.Parallel()
	for _, x := range []int{0, 1, 3, 7} {
		for _, bottom := range []bool{false, true} {
			t.Run(fmt.Sprintf("X=%d/bottom=%v", x, bottom), func(t *testing.T) {
				t.Parallel()
				position := "on top of"
				if bottom {
					position = "on the bottom of"
				}
				def := legacyCard(t, carddata.ScryfallCard{
					Name: "Hand Library X", Layout: "normal", TypeLine: "Instant", ManaCost: "{X}{U}",
					OracleText: "Draw two cards. You gain 2 life. Put X cards from your hand " + position + " your library.",
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				old := addCardToHand(g, game.Player2, evidenceCard("Old", 1))
				other := addCardToHand(g, game.Player1, evidenceCard("Other player's card", 1))
				addCardToLibrary(g, game.Player2, evidenceCard("Anchor", 1))
				a := addCardToLibrary(g, game.Player2, evidenceCard("Drawn A", 1))
				b := addCardToLibrary(g, game.Player2, evidenceCard("Drawn B", 1))
				obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player2, XValue: x}
				order := []string{"Drawn A", "Old", "Drawn B"}[:min(x, 3)]
				agent := &orderedHandChoiceAgent{order: order}
				var agents [game.NumPlayers]PlayerAgent
				agents[game.Player2] = agent
				log := &TurnLog{}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, agents, log)
				count := min(x, 3)
				if g.Players[game.Player2].Hand.Size() != 3-count || g.Players[game.Player2].Life != 42 ||
					!g.Players[game.Player1].Hand.Contains(other) {
					t.Fatalf("hand=%d life=%d other hand=%v", g.Players[game.Player2].Hand.Size(), g.Players[game.Player2].Life, g.Players[game.Player1].Hand.All())
				}
				if x == 0 {
					if len(agent.requests) != 0 || !slices.Equal(g.Players[game.Player2].Hand.All(), []id.ID{a, b, old}) {
						t.Fatalf("X=0 requests=%d hand=%v", len(agent.requests), g.Players[game.Player2].Hand.All())
					}
				} else {
					if len(agent.requests) != 1 || agent.requests[0].Player != game.Player2 ||
						agent.requests[0].MinChoices != count || agent.requests[0].MaxChoices != count ||
						len(agent.requests[0].Options) != 3 || len(log.Choices) != 1 || log.Choices[0].UsedFallback {
						t.Fatalf("requests=%+v choices=%+v", agent.requests, log.Choices)
					}
					if bottom {
						assertLibraryBottomNames(t, g, game.Player2, order)
					} else {
						assertLibraryTopNames(t, g, game.Player2, order)
					}
				}
				assertBrainstormZoneEvents(t, g.Events, count)
			})
		}
	}
}

func TestCompiledHandLibraryStandaloneAndOptional(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text        string
		x, want, requests int
		decline           bool
	}{
		{"fixed insufficient", "Put five cards from your hand on top of your library.", 0, 2, 1, false},
		{"activated zero", "{X}: Put X cards from your hand on top of your library.", 0, 0, 0, false},
		{"activated one", "{X}: Put X cards from your hand on top of your library.", 1, 1, 1, false},
		{"optional decline", "You may put a card from your hand on top of your library.", 0, 0, 1, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			typeLine := "Instant"
			if tt.name == "activated zero" || tt.name == "activated one" {
				typeLine = "Artifact"
			}
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Hand Library Standalone", Layout: "normal", TypeLine: typeLine, OracleText: tt.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			a := addCardToHand(g, game.Player1, evidenceCard("A", 1))
			b := addCardToHand(g, game.Player1, evidenceCard("B", 1))
			cardA, _ := g.GetCardInstance(a)
			cardB, _ := g.GetCardInstance(b)
			versionA, versionB := cardA.ZoneVersion, cardB.ZoneVersion
			content := def.SpellAbility.Val
			if len(def.ActivatedAbilities) != 0 {
				content = def.ActivatedAbilities[0].Content
			}
			agent := &orderedHandChoiceAgent{order: []string{"B", "A"}[:tt.want]}
			if tt.decline {
				agent.answer = []int{0}
			}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = agent
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1, XValue: tt.x}
			if len(def.ActivatedAbilities) != 0 {
				obj.Kind = game.StackActivatedAbility
				obj.SourceID = addCombatPermanent(g, game.Player1, def).ObjectID
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Size() != 2-tt.want || len(agent.requests) != tt.requests {
				t.Fatalf("hand=%v requests=%+v", g.Players[game.Player1].Hand.All(), agent.requests)
			}
			if tt.want == 0 && !slices.Equal(g.Players[game.Player1].Hand.All(), []id.ID{b, a}) {
				t.Fatalf("zero/decline changed hand: %v", g.Players[game.Player1].Hand.All())
			}
			if tt.want == 0 && (cardA.ZoneVersion != versionA || cardB.ZoneVersion != versionB) {
				t.Fatal("zero/decline changed card incarnations")
			}
			assertBrainstormZoneEvents(t, g.Events, tt.want)
		})
	}
}

func TestCompiledHandLibraryConditionAndOptionalCompositions(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text                  string
		hand, moved, life, requests int
	}{
		{"life then put", "You gain 2 life. Put a card from your hand on top of your library.", 1, 1, 42, 1},
		{"leading condition true", "If you have no cards in hand, draw two cards, then put a card from your hand on top of your library.", 0, 1, 40, 1},
		{"leading condition false", "If you have no cards in hand, draw two cards, then put a card from your hand on top of your library.", 1, 0, 40, 0},
		{"post draw condition false", "Draw a card. If you have no cards in hand, put a card from your hand on top of your library.", 0, 0, 40, 0},
		{"post draw optional decline", "Draw a card. You may put a card from your hand on top of your library.", 0, 0, 40, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Hand Library Envelope", Layout: "normal", TypeLine: "Instant", OracleText: tt.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.hand != 0 {
				addCardToHand(g, game.Player1, evidenceCard("Old", 1))
			}
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn A", 1))
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn B", 1))
			agent := &orderedHandChoiceAgent{answer: []int{0}}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = agent
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
				def.SpellAbility.Val, agents, &TurnLog{})
			if len(agent.requests) != tt.requests || g.Players[game.Player1].Life != tt.life {
				t.Fatalf("requests=%+v life=%d", agent.requests, g.Players[game.Player1].Life)
			}
			assertBrainstormZoneEvents(t, g.Events, tt.moved)
		})
	}
}

func TestCompiledHandLibraryMultipleIndependentChoices(t *testing.T) {
	t.Parallel()
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Independent Hand Placements", Layout: "normal", TypeLine: "Instant",
		OracleText: "Put a card from your hand on top of your library. You gain 2 life. Put a card from your hand on the bottom of your library.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	a := addCardToHand(g, game.Player1, evidenceCard("A", 1))
	b := addCardToHand(g, game.Player1, evidenceCard("B", 1))
	addCardToLibrary(g, game.Player1, evidenceCard("Anchor", 1))
	agent := &orderedHandChoiceAgent{answer: []int{0}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
		def.SpellAbility.Val, agents, &TurnLog{})
	if len(agent.requests) != 2 || len(agent.requests[0].Options) != 2 || len(agent.requests[1].Options) != 1 ||
		agent.requests[0].Options[0].Card.Val.CardID != b || agent.requests[1].Options[0].Card.Val.CardID != a ||
		g.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player1].Life != 42 {
		t.Fatalf("choices=%+v hand=%v life=%d", agent.requests, g.Players[game.Player1].Hand.All(), g.Players[game.Player1].Life)
	}
	assertLibraryTopNames(t, g, game.Player1, []string{"B"})
	assertLibraryBottomNames(t, g, game.Player1, []string{"A"})
}

func TestCompiledHandLibraryReplacementKeepsActualDestination(t *testing.T) {
	t.Parallel()
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Hand Library Replacement", Layout: "normal", TypeLine: "Instant",
		OracleText: "Put two cards from your hand on the bottom of your library.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	redirected := selfLibraryReplacementCardDef()
	replacement := &redirected.ReplacementAbilities[0].Replacement
	replacement.ToZone = zone.Library
	replacement.ReplaceToZone = zone.Exile
	replacement.ShuffleIntoLibrary = false
	replacement.RevealSource = false
	cardID := addCardToHand(g, game.Player1, redirected)
	addCardToHand(g, game.Player1, evidenceCard("Normal", 1))
	agent := &orderedHandChoiceAgent{answer: []int{0, 1}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
		def.SpellAbility.Val, agents, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Library.Contains(cardID) ||
		g.Players[game.Player1].Hand.Size() != 0 || len(agent.requests) != 1 {
		t.Fatalf("replacement card: exile=%v library=%v hand=%v choices=%d",
			g.Players[game.Player1].Exile.All(), g.Players[game.Player1].Library.All(), g.Players[game.Player1].Hand.All(), len(agent.requests))
	}
	assertLibraryBottomNames(t, g, game.Player1, []string{"Normal"})
}
