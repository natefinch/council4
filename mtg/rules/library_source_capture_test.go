package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledLibrarySource(t *testing.T, name, typeLine string) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: name, Layout: "normal", TypeLine: typeLine, Power: new("1"), Toughness: new("1"),
		OracleText: fmt.Sprintf("When %s is put into your graveyard from your library, you may exile it. If you do, %s deals 3 damage to each opponent and you gain 3 life.", name, name),
	})
}

func TestLibrarySourceCaptureRequiresExactEvent(t *testing.T) {
	for _, corruption := range []string{"valid", "missing card", "wrong ID", "wrong owner", "missing version", "stale", "wrong player", "wrong event", "wrong origin", "wrong destination", "permanent event", "wrong location"} {
		t.Run(corruption, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := compiledLibrarySource(t, "Capture Source", "Sorcery")
			card := addCardToLibrary(g, game.Player2, def)
			millCards(g, game.Player2, 1)
			event := g.Events[len(g.Events)-1]
			switch corruption {
			case "valid":
			case "missing card":
				delete(g.CardInstances, card)
			case "wrong ID":
				g.CardInstances[card].ID = g.IDGen.Next()
			case "wrong owner":
				g.CardInstances[card].Owner = game.Player3
			case "missing version":
				event.CardZoneVersion = 0
			case "stale":
				moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Exile, false)
				moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Exile, zone.Graveyard, false)
			case "wrong player":
				event.Player = game.Player1
			case "wrong event":
				event.Kind = game.EventCardDrawn
			case "wrong origin":
				event.FromZone = zone.Hand
			case "wrong destination":
				event.ToZone = zone.Exile
			case "permanent event":
				event.PermanentID = g.IDGen.Next()
			case "wrong location":
				g.Players[game.Player2].Graveyard.Remove(card)
				g.Players[game.Player2].Hand.Add(card)
			default:
				t.Fatal("unknown corruption")
			}
			captured := captureLibraryGraveyardSourceTriggers(g, event)
			if corruption != "valid" {
				if len(captured) != 0 {
					t.Fatal("unproved card/event incarnation acquired a source capture")
				}
				return
			}
			if len(captured) != 1 || captured[0].SourceCardID != card ||
				captured[0].SourceID != card || captured[0].Controller != game.Player2 ||
				captured[0].SourceZone != zone.Graveyard || captured[0].SourceZoneVersion != event.CardZoneVersion ||
				captured[0].Ability != &def.TriggeredAbilities[0] {
				t.Fatal("capture did not retain exact declared source/body/owner/post-event incarnation")
			}
		})
	}
}

func TestLibrarySourceCaptureRequiresDeclaredSelfPattern(t *testing.T) {
	for _, corruption := range []string{"watcher", "excluded self", "plural", "union", "excluded origin", "excluded destination", "wrong player", "explicit battlefield", "step type"} {
		t.Run(corruption, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := compiledLibrarySource(t, "Declared Source", "Sorcery")
			ability := &def.TriggeredAbilities[0]
			switch corruption {
			case "watcher":
				ability.Trigger.Pattern.Source = game.TriggerSourceAny
			case "excluded self":
				ability.Trigger.Pattern.ExcludeSelf = true
			case "plural":
				ability.Trigger.Pattern.OneOrMore = true
			case "union":
				ability.Trigger.Pattern.FromZones = []zone.Type{zone.Library, zone.Hand}
			case "excluded origin":
				ability.Trigger.Pattern.ExcludeFromZone = true
			case "excluded destination":
				ability.Trigger.Pattern.ExcludeToZone = true
			case "wrong player":
				ability.Trigger.Pattern.Player = game.TriggerPlayerOpponent
			case "explicit battlefield":
				ability.ZoneOfFunction = zone.Battlefield
			case "step type":
				ability.Trigger.Type = game.TriggerAt
			default:
				t.Fatal("unknown corruption")
			}
			addCardToLibrary(g, game.Player2, def)
			millCards(g, game.Player2, 1)
			for _, event := range g.Events {
				if len(event.TriggeredAbilities) != 0 {
					t.Fatal("ambiguous/foreign/unavailable declared predicate acquired an event-card capture")
				}
			}
			if NewEngine(nil).putTriggeredAbilitiesOnStack(g) {
				t.Fatal("a nonbattlefield watcher or unproved source body acquired a trigger")
			}
		})
	}
}

func TestLibrarySourceNaturalMillUsesOwnerNotMiller(t *testing.T) {
	def := compiledLibrarySource(t, "Creeping Chill", "Sorcery")
	miller := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Milling Spell", Layout: "normal",
		TypeLine: "Instant", OracleText: "Target player mills one card."})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToLibrary(g, game.Player2, def)
	foreign := addCardToLibrary(g, game.Player3, def)
	decoy := addCardToGraveyard(g, game.Player2, def)
	spell := addCardToHand(g, game.Player1, miller)
	engine := NewEngine(nil)
	engine.resolveAbilityContentWithChoices(g, &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
		SourceID: spell, SourceCardID: spell, Targets: []game.Target{game.PlayerTarget(game.Player2)},
	}, miller.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("actual opponent-library mill did not produce exactly one self-card trigger")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Controller != game.Player2 || obj.SourceCardID != card ||
		obj.SourceZone != zone.Graveyard || obj.SourceZoneVersion != g.CardInstances[card].ZoneVersion {
		t.Fatal("milling player replaced the original source owner/incarnation")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &repeatLandChoiceAgent{mayAnswers: []bool{true}}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	for playerID, player := range g.Players {
		want := 37
		if game.PlayerID(playerID) == game.Player2 {
			want = 43
		}
		if player.Life != want {
			t.Fatalf("player%d life=%d want%d", playerID, player.Life, want)
		}
	}
	if !g.Players[game.Player2].Exile.Contains(card) ||
		!g.Players[game.Player2].Graveyard.Contains(decoy) || !g.Players[game.Player3].Library.Contains(foreign) ||
		engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("capture moved a same-name/foreign card or dispatched its event twice")
	}
}
