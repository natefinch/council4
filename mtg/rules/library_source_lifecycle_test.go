package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLibrarySourceCapturedBodySurvivesDeparture(t *testing.T) {
	for _, scenario := range []string{"matching", "declined", "departed before queue", "returned before queue", "returned to battlefield", "departed before resolution", "cloned pending", "cloned stack", "copied stack", "copied controller", "replaced exile", "owner changed after capture", "definition changed after capture"} {
		t.Run(scenario, func(t *testing.T) {
			typeLine := "Sorcery"
			name := "Creeping Chill"
			if scenario == "returned to battlefield" {
				typeLine = "Creature"
				name = "Lifecycle Source"
			}
			def := compiledLibrarySource(t, name, typeLine)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			card := addCardToLibrary(g, game.Player2, def)
			decoy := addCardToGraveyard(g, game.Player2, def)
			millCards(g, game.Player2, 1)
			version := g.CardInstances[card].ZoneVersion
			if scenario == "departed before queue" || scenario == "returned before queue" {
				if !moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Exile, false) {
					t.Fatal("source departure failed")
				}
				if scenario == "returned before queue" &&
					!moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Exile, zone.Graveyard, false) {
					t.Fatal("source reentry failed")
				}
			}
			if scenario == "returned to battlefield" {
				newEffectResolver(NewEngine(nil), g, &game.StackObject{
					Controller: game.Player2, Targets: []game.Target{currentCardTarget(t, g, card)},
				}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).resolveInstruction(&game.Instruction{
					Primitive: game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})},
				})
				if _, ok := findPermanentByCardID(g, card); !ok {
					t.Fatal("fixture did not return the physical source as a new battlefield incarnation")
				}
			}
			if scenario == "cloned pending" {
				g = g.Clone()
			}
			if scenario == "owner changed after capture" {
				g.CardInstances[card].Owner = game.Player3
			}
			if scenario == "definition changed after capture" {
				g.CardInstances[card].Def = vanillaCreature("Different Body", 1, 1)
			}
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("source departure lost its declared captured ability or a returned incarnation duplicated it")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.Controller != game.Player2 || obj.SourceCardID != card ||
				obj.SourceZone != zone.Graveyard || obj.SourceZoneVersion != version || obj.InlineTrigger != &def.TriggeredAbilities[0] {
				t.Fatal("queue rebound original body/owner/card/zone/version from current source state")
			}
			if scenario == "departed before resolution" {
				moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Exile, false)
			}
			if scenario == "cloned stack" {
				g = g.Clone()
			}
			if scenario == "copied stack" || scenario == "copied controller" {
				copied := game.NewStackObjectCopy(obj, g.IDGen.Next())
				if scenario == "copied controller" {
					copied.Controller = game.Player3
				}
				g.Stack.Push(copied)
			}
			if scenario == "replaced exile" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Graveyard, MatchToZone: true, ToZone: zone.Exile,
					ReplaceToZone: zone.Hand,
				})
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player2: &repeatLandChoiceAgent{mayAnswers: []bool{scenario != "declined", true}},
				game.Player3: &repeatLandChoiceAgent{mayAnswers: []bool{true}},
			}
			for g.Stack.Size() != 0 {
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			}
			success := scenario == "matching" || scenario == "cloned pending" || scenario == "cloned stack" ||
				scenario == "copied stack" || scenario == "copied controller" || scenario == "definition changed after capture"
			controller := game.Player2
			if scenario == "copied controller" {
				controller = game.Player3
			}
			for playerID, player := range g.Players {
				want := 40
				if success {
					want = 37
					if game.PlayerID(playerID) == controller {
						want = 43
					}
				}
				if player.Life != want {
					t.Fatalf("player%d life=%d want%d; stale or copied ability reused another incarnation/payment", playerID, player.Life, want)
				}
			}
			if !g.Players[game.Player2].Graveyard.Contains(decoy) || engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("source event rebound a decoy or was dispatched twice")
			}
		})
	}
}

func TestLibrarySourceReplacedMillDoesNotCapture(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := compiledLibrarySource(t, "Creeping Chill", "Sorcery")
	card := addCardToLibrary(g, game.Player2, def)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Library, MatchToZone: true, ToZone: zone.Graveyard,
		ReplaceToZone: zone.Exile,
	})
	if len(millCards(g, game.Player2, 1)) != 0 || !g.Players[game.Player2].Exile.Contains(card) {
		t.Fatal("mill was not actually replaced before graveyard entry")
	}
	engine := NewEngine(nil)
	if engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 {
		t.Fatal("replacement acquired a Library-to-Graveyard source trigger or paid its owned exile")
	}
	for _, player := range g.Players {
		if player.Life != 40 {
			t.Fatal("replacement rewarded an event that never occurred")
		}
	}
}

func TestLibrarySourceMulticardMillKeepsWatchers(t *testing.T) {
	def := compiledLibrarySource(t, "Creeping Chill", "Sorcery")
	watcher := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Library Watcher", Layout: "normal", TypeLine: "Creature", Power: new("1"), Toughness: new("1"),
		OracleText: "Whenever a card is put into a graveyard from a library, you gain 1 life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToLibrary(g, game.Player2, def)
	second := addCardToLibrary(g, game.Player2, def)
	decoy := addCardToGraveyard(g, game.Player2, def)
	live := addCombatPermanent(g, game.Player1, watcher)
	if cards := millCards(g, game.Player2, 2); len(cards) != 2 {
		t.Fatal("fixture did not perform one actual two-card mill batch")
	}
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 4 {
		t.Fatal("two physical source cards and ordinary watcher did not each dispatch once per matching event")
	}
	counts := map[game.ObjectID]int{}
	for _, obj := range g.Stack.Objects() {
		counts[obj.SourceID]++
	}
	if counts[first] != 1 || counts[second] != 1 || counts[live.ObjectID] != 2 {
		t.Fatal("captured card and battlefield watcher dispatch were merged, duplicated or rebound by name")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &repeatLandChoiceAgent{mayAnswers: []bool{true, true}}}
	for g.Stack.Size() != 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	if g.Players[game.Player1].Life != 36 || g.Players[game.Player2].Life != 46 ||
		g.Players[game.Player3].Life != 34 || g.Players[game.Player4].Life != 34 ||
		!g.Players[game.Player2].Exile.Contains(first) || !g.Players[game.Player2].Exile.Contains(second) ||
		!g.Players[game.Player2].Graveyard.Contains(decoy) || engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("natural batch rewards, exact-card exile or ordinary watcher semantics changed")
	}
}
