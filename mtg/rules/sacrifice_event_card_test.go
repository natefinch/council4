package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledSacrificeCardReturn(t *testing.T, control string) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Sacrifice Watcher", Layout: "normal", TypeLine: "Creature",
		Power: new("1"), Toughness: new("1"),
		OracleText: "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield" + control + ".",
	})
}

func sacrificedCardEvent(t *testing.T, g *game.Game, permanent *game.Permanent) game.Event {
	t.Helper()
	var found []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentSacrificed && event.PermanentID == permanent.ObjectID {
			found = append(found, event)
		}
	}
	if len(found) != 1 {
		t.Fatalf("sacrifice events=%d, want exactly one for object %d", len(found), permanent.ObjectID)
	}
	return found[0]
}

func TestSacrificeEventSnapshotsCompletedPhysicalMove(t *testing.T) {
	for _, destination := range []zone.Type{zone.Graveyard, zone.Exile, zone.Hand} {
		t.Run(destination.String(), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			victim := addCombatPermanent(g, game.Player3, vanillaCreature("Victim", 2, 3))
			victim.Controller = game.Player1
			if destination != zone.Graveyard {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: destination,
				})
			}
			if !sacrificePermanent(g, victim) {
				t.Fatal("actual sacrifice failed")
			}
			event := sacrificedCardEvent(t, g, victim)
			card, _ := g.GetCardInstance(victim.CardInstanceID)
			if event.CardID != card.ID || event.CardZoneVersion == 0 || event.CardZoneVersion != card.ZoneVersion ||
				event.FromZone != zone.Battlefield || event.ToZone != destination ||
				event.Controller != game.Player1 || event.Player != game.Player1 {
				t.Fatalf("event=%+v, actual card=%+v destination=%v", event, card, destination)
			}
			obj := &game.StackObject{Controller: game.Player2, HasTriggerEvent: true, TriggerEvent: event}
			if cardID, reached, ok := resolveCardReference(g, obj, game.CardReference{Kind: game.CardReferenceEvent}); !ok || cardID != card.ID || reached != destination {
				t.Fatal("completed emitted snapshot did not resolve exact physical card/destination")
			}
			if !moveCardBetweenZones(g, card.Owner, card.ID, destination, zone.Library) {
				t.Fatal("fixture departure failed")
			}
			if _, _, ok := resolveCardReference(g, obj, game.CardReference{Kind: game.CardReferenceEvent}); ok {
				t.Fatal("immutable sacrifice event followed a later incarnation")
			}
			if obj.TriggerEvent.CardID != event.CardID || obj.TriggerEvent.CardZoneVersion != event.CardZoneVersion ||
				obj.TriggerEvent.ToZone != event.ToZone {
				t.Fatal("resolution rewrote emitted event")
			}
		})
	}
}

func TestCompiledSacrificeEventCardReturnNaturalGameplay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := compiledSacrificeCardReturn(t, " under your control")
	source := addCombatPermanent(g, game.Player2, def)
	source.Owner = game.Player4
	g.CardInstances[source.CardInstanceID].Owner = game.Player4
	victimDef := vanillaCreature("Same Physical Name", 2, 3)
	victim := addCombatPermanent(g, game.Player3, victimDef)
	victim.Controller = game.Player1
	decoy := addCardToGraveyard(g, game.Player3, victimDef)
	if !sacrificePermanent(g, victim) {
		t.Fatal("ordinary sacrifice failed")
	}
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("ordinary compiled opponent sacrifice did not queue exactly once")
	}
	obj, _ := g.Stack.Peek()
	if obj.Controller != game.Player2 || obj.TriggerEvent.CardID != victim.CardInstanceID ||
		obj.TriggerEvent.PermanentID != victim.ObjectID || obj.TriggerEvent.CardZoneVersion == 0 {
		t.Fatal("queued ability lost actual sacrifice controller/card/object/version")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	returned, ok := findPermanentByCardID(g, victim.CardInstanceID)
	if !ok || returned.Owner != game.Player3 || returned.Controller != game.Player2 ||
		returned.ObjectID == victim.ObjectID || !g.Players[game.Player3].Graveyard.Contains(decoy) {
		t.Fatal("printed return lost exact new incarnation, owner/controller separation, or graveyard decoy")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("same sacrifice event dispatched twice")
	}
}

func TestCompiledEventCardExplicitRecipientOverridesOwner(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "default owner", true: "explicit controller"}[explicit], func(t *testing.T) {
			control := ""
			if explicit {
				control = " under your control"
			}
			def := compiledSacrificeCardReturn(t, control)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player2, def)
			cardID := addCardToGraveyard(g, game.Player3, vanillaCreature("Event Card", 2, 3))
			card, _ := g.GetCardInstance(cardID)
			card.ZoneVersion = 1
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player2, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventPermanentSacrificed, Player: game.Player1,
					Controller: game.Player1, CardID: cardID, CardZoneVersion: card.ZoneVersion},
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			returned, ok := findPermanentByCardID(g, cardID)
			want := game.Player3
			if explicit {
				want = game.Player2
			}
			if !ok || returned.Owner != game.Player3 || returned.Controller != want {
				t.Fatalf("explicit=%v returned=%+v want controller=%d", explicit, returned, want)
			}
		})
	}
}
func TestCompiledSacrificeEventReturnCapturedLifecycle(t *testing.T) {
	for _, scenario := range []string{
		"ordinary", "source leaves before queue", "source returns before queue", "source control changes",
		"source body changes", "source leaves before resolution", "card leaves before queue",
		"card reenters before queue", "card leaves before resolution", "cloned pending", "cloned stack",
		"copied same controller", "copied other controller", "diverted sacrifice", "diverted return",
	} {
		t.Run(scenario, func(t *testing.T) {
			def := compiledSacrificeCardReturn(t, " under your control")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player2, def)
			victimDef := vanillaCreature("Physical Card", 2, 3)
			victim := addCombatPermanent(g, game.Player3, victimDef)
			victim.Controller = game.Player1
			decoy := addCardToGraveyard(g, game.Player3, victimDef)
			if scenario == "diverted sacrifice" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
				})
			}
			if !sacrificePermanent(g, victim) {
				t.Fatal("ordinary sacrifice failed")
			}
			event := sacrificedCardEvent(t, g, victim)
			if scenario == "source leaves before queue" || scenario == "source returns before queue" {
				movePermanentToZone(g, source, zone.Graveyard)
				if scenario == "source returns before queue" {
					newEffectResolver(NewEngine(nil), g, &game.StackObject{Controller: game.Player4,
						Targets: []game.Target{currentCardTarget(t, g, source.CardInstanceID)}},
						[game.NumPlayers]PlayerAgent{}, &TurnLog{}).resolveInstruction(&game.Instruction{
						Primitive: game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})},
					})
				}
			}
			if scenario == "source control changes" {
				source.Controller = game.Player4
			}
			if scenario == "source body changes" {
				g.CardInstances[source.CardInstanceID].Def = vanillaCreature("Different Body", 1, 1)
			}
			if scenario == "card leaves before queue" || scenario == "card reenters before queue" {
				moveCardBetweenZones(g, game.Player3, victim.CardInstanceID, zone.Graveyard, zone.Exile)
				if scenario == "card reenters before queue" {
					moveCardBetweenZones(g, game.Player3, victim.CardInstanceID, zone.Exile, zone.Graveyard)
				}
			}
			if scenario == "cloned pending" {
				g = g.Clone()
			}
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("event-time source departure/control/body lost ability or returned incarnation dispatched twice")
			}
			obj, _ := g.Stack.Peek()
			if obj.Controller != game.Player2 || obj.SourceID != source.ObjectID ||
				obj.TriggerEvent.CardZoneVersion != event.CardZoneVersion ||
				obj.TriggerEvent.CardID != victim.CardInstanceID {
				t.Fatal("queued snapshot was rebound from later source/card state")
			}
			if scenario == "source leaves before resolution" {
				movePermanentToZone(g, source, zone.Graveyard)
			}
			if scenario == "card leaves before resolution" {
				moveCardBetweenZones(g, game.Player3, victim.CardInstanceID, zone.Graveyard, zone.Exile)
			}
			if scenario == "cloned stack" {
				g = g.Clone()
			}
			controller := game.Player2
			if scenario == "copied same controller" || scenario == "copied other controller" {
				copied := game.NewStackObjectCopy(obj, g.IDGen.Next())
				if scenario == "copied other controller" {
					copied.Controller = game.Player4
					controller = game.Player4
				}
				g.Stack.Push(copied)
			}
			if scenario == "diverted return" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Graveyard,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
				})
			}
			for g.Stack.Size() > 0 {
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			returned, ok := findPermanentByCardID(g, victim.CardInstanceID)
			want := scenario != "card leaves before queue" && scenario != "card reenters before queue" &&
				scenario != "card leaves before resolution" && scenario != "diverted return"
			if ok != want || ok && (returned.Controller != controller || returned.Owner != game.Player3 ||
				returned.ObjectID == victim.ObjectID) || !g.Players[game.Player3].Graveyard.Contains(decoy) {
				t.Fatal("exact return rebound owner/controller/incarnation/decoy or accepted a stale event")
			}
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("old sacrifice event dispatched again")
			}
		})
	}
}
