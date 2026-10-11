package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledLibraryGraveyardDamageAttribution(t *testing.T) {
	for _, typeLine := range []string{"Sorcery", "Creature"} {
		def := compileUnlessCard(t, cardgen.ScryfallCard{
			Name: "Event Source", Layout: "normal", TypeLine: typeLine, Power: new("1"), Toughness: new("1"),
			OracleText: "When Event Source is put into your graveyard from your library, you may exile it. If you do, Event Source deals 3 damage to each opponent and you gain 3 life.",
		})
		for _, scenario := range []string{"success", "copied controller", "declined", "stale", "missing version", "foreign event", "wrong event", "replaced exile", "older permanent"} {
			if scenario == "older permanent" && typeLine != "Creature" {
				continue
			}
			t.Run(typeLine+"/"+scenario, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				owner := game.Player2
				if scenario == "copied controller" {
					owner = game.Player1
				}
				var card, oldObject id.ID
				if scenario == "older permanent" {
					older := addCombatPermanent(g, owner, def)
					older.Controller = game.Player3
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), AffectedObjectID: older.ObjectID, Layer: game.LayerAbility,
						AddKeywords: []game.Keyword{game.Lifelink, game.Deathtouch},
					})
					card, oldObject = older.CardInstanceID, older.ObjectID
					if !movePermanentToZone(g, older, zone.Library) {
						t.Fatal("older incarnation did not leave the battlefield")
					}
					snapshot, ok := lastKnownObject(g, oldObject)
					if !ok || snapshot.Controller != game.Player3 || len(snapshot.Keywords) != 2 {
						t.Fatal("older incarnation LKI control was not retained")
					}
				} else {
					card = addCardToLibrary(g, owner, def)
				}
				if !moveCardBetweenZonesWithPlacement(g, owner, card, zone.Library, zone.Graveyard, false) {
					t.Fatal("actual library-to-graveyard event failed")
				}
				event := g.Events[len(g.Events)-1]
				if event.Kind != game.EventZoneChanged || event.CardID != card || event.CardZoneVersion == 0 ||
					event.FromZone != zone.Library || event.ToZone != zone.Graveyard {
					t.Fatal("fixture did not capture the real source-card occurrence")
				}
				decoy := addCardToGraveyard(g, owner, def)
				if scenario == "stale" {
					if !moveCardBetweenZonesWithPlacement(g, owner, card, zone.Graveyard, zone.Exile, false) ||
						!moveCardBetweenZonesWithPlacement(g, owner, card, zone.Exile, zone.Graveyard, false) {
						t.Fatal("fixture did not create a new card incarnation")
					}
				}
				if scenario == "missing version" {
					event.CardZoneVersion = 0
				}
				if scenario == "foreign event" {
					event.CardID = decoy
				}
				if scenario == "wrong event" {
					event.FromZone = zone.Hand
				}
				if scenario == "replaced exile" {
					g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
						ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
						MatchFromZone: true, FromZone: zone.Graveyard,
						MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Hand,
					})
				}
				trigger := def.TriggeredAbilities[0]
				source := &game.Permanent{ObjectID: card, CardInstanceID: card, Owner: owner, Controller: owner}
				matched := triggerMatchesEventForController(g, source, owner, &trigger.Trigger.Pattern, event)
				if matched == (scenario == "foreign event" || scenario == "wrong event") {
					t.Fatal("compiled self-event matcher confused the original source or zone")
				}
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2,
					SourceID: card, SourceCardID: card, SourceZone: zone.Graveyard, SourceZoneVersion: event.CardZoneVersion,
					TriggerEvent: event, HasTriggerEvent: true,
					ResolutionResults: map[string]game.InstructionResolutionResult{
						"if-you-do": {Succeeded: true, Amount: 777},
					},
				}
				agents := [game.NumPlayers]PlayerAgent{}
				agents[game.Player2] = &repeatLandChoiceAgent{mayAnswers: []bool{scenario != "declined", scenario != "declined"}}
				if matched {
					engine := NewEngine(nil)
					engine.resolveAbilityContentWithChoices(g, obj, trigger.Content, agents, &TurnLog{})
					engine.resolveAbilityContentWithChoices(g, obj, trigger.Content, agents, &TurnLog{})
				}
				success := scenario == "success" || scenario == "copied controller" || scenario == "older permanent"
				for playerID, player := range g.Players {
					want := 40
					if success {
						want = 37
						if game.PlayerID(playerID) == game.Player2 {
							want = 43
						}
					}
					if player.Life != want {
						t.Fatalf("player%d life=%d want%d; exile success or original-card attribution changed", playerID, player.Life, want)
					}
				}
				if g.Players[owner].Exile.Contains(card) != success || !g.Players[owner].Graveyard.Contains(decoy) {
					t.Fatal("exile moved a stale incarnation or same-name decoy")
				}
				damageEvents := 0
				for _, recorded := range g.Events {
					if recorded.Kind != game.EventDamageDealt {
						continue
					}
					damageEvents++
					if recorded.SourceID != card || recorded.SourceObjectID != card ||
						recorded.SourceObjectID == oldObject || recorded.Controller != game.Player2 {
						t.Fatal("damage rebound to older permanent LKI, another card, or original owner")
					}
				}
				if success && damageEvents != 3 || !success && damageEvents != 0 ||
					len(obj.ResolutionResults) != 1 || obj.ResolutionResults["if-you-do"].Amount != 777 {
					t.Fatal("actual-result scope reused a prior success or failed to restore parent state")
				}
			})
		}
	}
}

func TestCompiledLibraryWatcherKeepsPermanentAttribution(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Library Watcher", Layout: "normal", TypeLine: "Creature", Power: new("1"), Toughness: new("1"),
		OracleText: "Whenever a card is put into your graveyard from your library, Library Watcher deals 3 damage to each opponent. You gain 3 life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	source.Controller = game.Player3
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID, Layer: game.LayerAbility,
		AddKeywords: []game.Keyword{game.Lifelink},
	})
	subject := addCardToLibrary(g, game.Player3, vanillaCreature("Foreign Source", 1, 1))
	if !moveCardBetweenZonesWithPlacement(g, game.Player3, subject, zone.Library, zone.Graveyard, false) {
		t.Fatal("foreign card event failed")
	}
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		TriggerEvent: g.Events[len(g.Events)-1], HasTriggerEvent: true,
	}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for playerID, player := range g.Players {
		want := 37
		if game.PlayerID(playerID) == game.Player2 {
			want = 43
		}
		if game.PlayerID(playerID) == game.Player3 {
			want = 46
		}
		if player.Life != want {
			t.Fatalf("player%d life=%d want%d; source lifelink/controller confused with ability controller or event card", playerID, player.Life, want)
		}
	}
}
