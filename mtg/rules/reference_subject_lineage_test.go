package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func referenceSubjectActivated(t *testing.T, text string) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Reference Capability", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: " + text,
	})
}

func TestCompiledReferenceSubjectIndependentCardPredicate(t *testing.T) {
	for _, group := range []bool{false, true} {
		body := "Reveal the top card of your library. If it's a land card, put it onto the battlefield. If it's a snow card, you gain 2 life."
		if group {
			body = "Reveal the top card of your library. If it's a land card, put it onto the battlefield. If it's a snow card, draw a card and you gain 2 life."
		}
		def := referenceSubjectActivated(t, body)
		for _, tc := range []struct {
			name, typeLine string
			enter, snow    bool
		}{
			{"entered snow land", "Snow Land", true, true},
			{"skipped snow artifact", "Snow Artifact", false, true},
			{"entered nonsnow land", "Land", true, false},
			{"skipped nonsnow artifact", "Artifact", false, false},
			{"unavailable input", "", false, false},
		} {
			t.Run(tc.name+map[bool]string{false: "/clause", true: "/group"}[group], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player2, def)
				source.Owner = game.Player1
				g.CardInstances[source.CardInstanceID].Owner = game.Player1
				var observed game.CardInstance
				if tc.typeLine != "" {
					card := compileUnlessCard(t, cardgen.ScryfallCard{
						Name: "Observed Capability", Layout: "normal", TypeLine: tc.typeLine,
					})
					cardID := addCardToLibrary(g, game.Player2, card)
					observed = *g.CardInstances[cardID]
				}
				g.Players[game.Player2].ManaPool.Add(mana.C, 1)
				g.Turn.PriorityPlayer = game.Player2
				engine := NewEngine(nil)
				activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
				engine.resolveTopOfStack(g, &TurnLog{})
				want := 40
				if tc.snow {
					want = 42
				}
				if g.Players[game.Player2].Life != want {
					t.Errorf("independent card predicate life=%d want %d", g.Players[game.Player2].Life, want)
				}
				if tc.typeLine != "" && (permanentForCard(g, observed.ID) != nil) != tc.enter {
					t.Error("placement condition or exact card lineage changed")
				}
				if group && tc.snow && !tc.enter && !g.Players[game.Player2].Hand.Contains(observed.ID) {
					t.Error("group did not freeze its card predicate before drawing that card")
				}
				if g.Players[game.Player1].Life != 40 {
					t.Error("predicate consequence used source owner instead of controller")
				}
				t.Logf("card=%d capturedVersion=%d enter=%v snow=%v life=%d",
					observed.ID, observed.ZoneVersion, tc.enter, tc.snow, g.Players[game.Player2].Life)
			})
		}
	}
}

func TestCompiledReferenceSubjectReachedCardHandMovement(t *testing.T) {
	for _, optional := range []bool{false, true} {
		body := "Look at the top card of your library. Put it onto the battlefield. Put it into your hand."
		if optional {
			body = "Look at the top card of your library. You may put it onto the battlefield. Put it into your hand."
		}
		def := referenceSubjectActivated(t, body)
		sequence := def.ActivatedAbilities[0].Content.Modes[0].Sequence
		if len(sequence) != 3 {
			t.Fatalf("complete two-move CardDef has %d instructions", len(sequence))
		}
		move, ok := sequence[2].Primitive.(game.MovePermanent)
		if !ok || move.Object != game.LinkedObjectReference("sequence-effect-1-product") ||
			move.Destination != zone.Hand {
			t.Fatalf("reached card movement=%T %v", sequence[2].Primitive, move)
		}
		for _, state := range []string{"enter", "empty library", "diverted entry", "decline"} {
			if state == "decline" && !optional {
				continue
			}
			t.Run(map[bool]string{false: "mandatory/", true: "optional/"}[optional]+state, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player2, def)
				source.Owner = game.Player1
				g.CardInstances[source.CardInstanceID].Owner = game.Player1
				decoy := addCombatPermanent(g, game.Player3, referenceSubjectCreature(t))
				var observed game.CardInstance
				if state != "empty library" {
					card := referenceSubjectCreature(t)
					cardID := addCardToLibrary(g, game.Player2, card)
					observed = *g.CardInstances[cardID]
				}
				g.Players[game.Player2].ManaPool.Add(mana.C, 1)
				g.Turn.PriorityPlayer = game.Player2
				engine := NewEngine(nil)
				if state == "diverted entry" {
					resolveInstruction(engine, g, &game.StackObject{Controller: game.Player2}, game.CreateReplacement{
						Replacement: &game.ReplacementEffect{
							Description: "entry diverted to exile",
							MatchEvent:  game.EventZoneChanged,
							MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
						},
					}, nil)
				}
				activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
				agent := &libraryPaymentAgent{accept: state != "decline"}
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
				wantPrompts := 0
				if optional {
					wantPrompts = 1
				}
				if len(agent.prompts) != wantPrompts {
					t.Errorf("optional prompts=%d want %d", len(agent.prompts), wantPrompts)
				}
				if g.Players[game.Player2].Hand.Contains(observed.ID) != (state == "enter") {
					t.Error("hand movement did not consume only the actual entered card")
				}

				if permanentForCard(g, decoy.CardInstanceID) == nil || permanentForCard(g, source.CardInstanceID) == nil ||
					g.Players[game.Player1].Hand.Contains(observed.ID) {
					t.Error("movement used another object or the source owner's hand")
				}
				if state == "enter" && (permanentForCard(g, observed.ID) != nil ||
					g.CardInstances[observed.ID].ZoneVersion != observed.ZoneVersion+2) {
					t.Error("actual battlefield-to-hand transition did not occur")
				}
				if state == "diverted entry" && !g.Players[game.Player2].Exile.Contains(observed.ID) {
					t.Fatal("replacement fixture did not actually divert the attempted entry")
				}
			})
		}
	}
}

func TestCompiledReferenceSubjectOptionalIndependentCardPredicate(t *testing.T) {
	def := referenceSubjectActivated(t,
		"Reveal the top card of your library. If it's a land card, you may put it onto the battlefield. If it's a snow card, draw a card and you gain 2 life.")
	for _, state := range []string{"enter", "decline", "skipped artifact", "empty", "diverted", "stale input", "copied resolution"} {
		t.Run(state, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player2, def)
			source.Owner = game.Player1
			g.CardInstances[source.CardInstanceID].Owner = game.Player1
			var observed game.CardInstance
			if state != "empty" {
				typeLine := "Snow Land"
				if state == "skipped artifact" {
					typeLine = "Snow Artifact"
				}
				card := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Optional Observed", Layout: "normal", TypeLine: typeLine,
				})
				observed = *g.CardInstances[addCardToLibrary(g, game.Player2, card)]
			}
			g.Players[game.Player2].ManaPool.Add(mana.C, 1)
			g.Turn.PriorityPlayer = game.Player2
			engine := NewEngine(nil)
			if state == "diverted" {
				resolveInstruction(engine, g, &game.StackObject{Controller: game.Player2}, game.CreateReplacement{
					Replacement: &game.ReplacementEffect{
						MatchEvent: game.EventZoneChanged, MatchToZone: true,
						ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
					},
				}, nil)
			}
			activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
			if state == "copied resolution" {
				g = g.Clone()
			}
			agent := &libraryPaymentAgent{accept: state != "decline"}
			if state == "stale input" {
				agent.afterMay = func() {
					if !moveCardBetweenZones(g, game.Player2, observed.ID, zone.Library, zone.Hand) ||
						!moveCardBetweenZones(g, game.Player2, observed.ID, zone.Hand, zone.Library) {
						t.Fatal("actual input incarnation change failed")
					}
				}
			}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
			wantPredicate := state != "empty" && state != "diverted" && state != "stale input"
			wantLife := 40
			if wantPredicate {
				wantLife = 42
			}
			if g.Players[game.Player2].Life != wantLife || g.Players[game.Player1].Life != 40 {
				t.Error("predicate used stale input, false action success, or source owner")
			}
			wantPrompts := 1
			if state == "empty" || state == "skipped artifact" {
				wantPrompts = 0
			}
			if len(agent.prompts) != wantPrompts {
				t.Fatalf("placement asked %d times, want %d", len(agent.prompts), wantPrompts)
			}
			wantEntry := state == "enter" || state == "copied resolution"
			if (permanentForCard(g, observed.ID) != nil) != wantEntry {
				t.Error("acceptance or stale input invented an entered permanent")
			}
			if (state == "decline" || state == "skipped artifact") &&
				!g.Players[game.Player2].Hand.Contains(observed.ID) {
				t.Error("input branch did not freeze its group predicate before drawing")
			}
			if state == "diverted" && !g.Players[game.Player2].Exile.Contains(observed.ID) {
				t.Fatal("actual replacement did not divert entry")
			}
		})
	}
}
