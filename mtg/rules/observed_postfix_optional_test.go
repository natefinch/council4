package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledObservedMatter(t *testing.T) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Matter Reshaper", Layout: "normal", TypeLine: "Creature — Eldrazi", ManaCost: "{2}{C}",
		Power: new("3"), Toughness: new("2"),
		OracleText: "({C} represents colorless mana.)\nWhen this creature dies, reveal the top card of your library. You may put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand.",
	})
}

func TestCompiledObservedPostfixActualDeathSubject(t *testing.T) {
	matter := compiledObservedMatter(t)
	for _, tc := range []struct {
		name, cost, typeLine, text string
		accept                     bool
		wantHand                   bool
		prompts                    int
	}{
		{"expensive permanent", "{4}", "Creature — Horse", "", true, true, 0},
		{"Divination", "{2}{U}", "Sorcery", "Draw two cards.", true, true, 0},
		{"known zero accepted", "{0}", "Creature — Horse", "", true, false, 1},
		{"qualifying accepted", "{3}", "Creature — Horse", "", true, false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := cardgen.ScryfallCard{
				Name: tc.name, Layout: "normal", TypeLine: tc.typeLine, ManaCost: tc.cost,
				Power: new("2"), Toughness: new("2"), OracleText: tc.text,
			}
			if tc.typeLine == "Sorcery" {
				fixture.Power, fixture.Toughness = nil, nil
			}
			def := compileUnlessCard(t, fixture)
			assertObservedMatterDeath(t, matter, def, tc.accept, tc.wantHand, tc.prompts)
		})
	}
}

func TestCompiledObservedPostfixDeclineReachesHand(t *testing.T) {
	matter := compiledObservedMatter(t)
	for _, cost := range []string{"{0}", "{3}"} {
		t.Run(cost, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Qualifying Creature", Layout: "normal", TypeLine: "Creature — Horse", ManaCost: cost,
				Power: new("2"), Toughness: new("2"),
			})
			assertObservedMatterDeath(t, matter, def, false, true, 1)
		})
	}
}

func assertObservedMatterDeath(t *testing.T, matter, observed *game.CardDef, accept, wantHand bool, prompts int) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToLibrary(g, game.Player1, observed)
	decoy := addCardToLibrary(g, game.Player3, observed)
	source := addCombatPermanent(g, game.Player1, matter)
	engine := NewEngine(nil)
	if _, ok := destroyPermanent(g, source.ObjectID); !ok || !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("normal compiled self-death trigger did not fire")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || !obj.HasTriggerEvent ||
		obj.TriggerEvent.PermanentID != source.ObjectID {
		t.Fatal("trigger lost the actual departed source/event identity")
	}
	agent := &libraryPaymentAgent{accept: accept}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	if len(agent.prompts) != prompts {
		t.Errorf("optional choices=%d want %d", len(agent.prompts), prompts)
	}
	if g.Players[game.Player1].Hand.Contains(card) != wantHand ||
		(permanentForCard(g, card) != nil) == wantHand || g.Players[game.Player1].Library.Contains(card) {
		t.Errorf("actual observed card destination: hand=%v battlefield=%v library=%v; want hand=%v",
			g.Players[game.Player1].Hand.Contains(card), permanentForCard(g, card) != nil,
			g.Players[game.Player1].Library.Contains(card), wantHand)
	}
	if g.CardInstances[card].Owner != game.Player1 || !g.Players[game.Player3].Library.Contains(decoy) {
		t.Error("observation used another owner/library")
	}
	t.Logf("actual death: observed=%s mana=%d accepted=%v choices=%d hand=%v battlefield=%v library=%v",
		observed.Name, observed.ManaValue(), accept, len(agent.prompts), g.Players[game.Player1].Hand.Contains(card),
		permanentForCard(g, card) != nil, g.Players[game.Player1].Library.Contains(card))
}

func TestCompiledObservedPostfixDeathIsolation(t *testing.T) {
	matter := compiledObservedMatter(t)
	creature := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Qualifying Creature", Layout: "normal", TypeLine: "Creature — Horse", ManaCost: "{3}",
		Power: new("2"), Toughness: new("2"),
	})
	for _, outcome := range []string{
		"clone", "different source owner", "reentered source", "empty",
		"accepted stale", "declined stale", "accepted diverted", "declined diverted",
	} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			controller := game.Player2
			observed := addCardToLibrary(g, controller, creature)
			decoy := addCardToLibrary(g, game.Player1, creature)
			source := addCombatPermanent(g, controller, matter)
			if outcome == "different source owner" {
				source.Owner = game.Player1
				g.CardInstances[source.CardInstanceID].Owner = game.Player1
			}
			if outcome == "empty" {
				g.Players[controller].Library.Remove(observed)
			}
			engine := NewEngine(nil)
			if _, ok := destroyPermanent(g, source.ObjectID); !ok || !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("actual death did not create a trigger")
			}
			if outcome == "clone" {
				g = g.Clone()
			}
			obj, ok := g.Stack.Peek()
			if !ok {
				t.Fatal("death trigger missing")
			}
			if outcome == "reentered source" {
				r := effectResolver{game: g, engine: engine, obj: obj}
				reentered, ok := r.putResolvedCardOnBattlefieldValue(g.CardInstances[source.CardInstanceID],
					zone.Graveyard, controller, nil, permanentCreationOptions{})
				if !ok || reentered.ObjectID == source.ObjectID {
					t.Fatal("actual source reentry did not create a new incarnation")
				}
			}
			if outcome == "accepted diverted" || outcome == "declined diverted" {
				to := zone.Battlefield
				if outcome == "declined diverted" {
					to = zone.Hand
				}
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true,
					FromZone: zone.Library, MatchToZone: true, ToZone: to, ReplaceToZone: zone.Exile,
				})
			}
			agent := &libraryPaymentAgent{accept: outcome != "declined stale" && outcome != "declined diverted"}
			if outcome == "accepted stale" || outcome == "declined stale" {
				agent.afterMay = func() {
					if !moveCardBetweenZones(g, controller, observed, zone.Library, zone.Exile) ||
						!moveCardBetweenZones(g, controller, observed, zone.Exile, zone.Library) {
						t.Fatal("real leave/reentry mutation failed")
					}
				}
			}
			obj.ResolutionResults = map[string]game.InstructionResolutionResult{"if-you-do": {Accepted: true, Succeeded: true, Amount: 41}}
			obj.ResolutionResultObjects = map[string][]game.ObjectSnapshot{"if-you-do": {{}}}
			obj.ResolvedAmounts = map[string]int{"if-you-do": 42}
			obj.ResolvedExcessDamage = map[string]int{"if-you-do": 43}
			var agents [game.NumPlayers]PlayerAgent
			agents[controller] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			wantPrompts := 1
			if outcome == "empty" {
				wantPrompts = 0
			}
			wantBattlefield := outcome == "clone" || outcome == "different source owner" || outcome == "reentered source"
			wantLibrary := outcome == "accepted stale" || outcome == "declined stale"
			wantExile := outcome == "accepted diverted" || outcome == "declined diverted"
			if len(agent.prompts) != wantPrompts || (permanentForCard(g, observed) != nil) != wantBattlefield ||
				g.Players[controller].Library.Contains(observed) != wantLibrary ||
				g.Players[controller].Exile.Contains(observed) != wantExile || g.Players[controller].Hand.Contains(observed) {
				t.Errorf("outcome: prompts=%d BF=%v library=%v exile=%v hand=%v",
					len(agent.prompts), permanentForCard(g, observed) != nil, g.Players[controller].Library.Contains(observed),
					g.Players[controller].Exile.Contains(observed), g.Players[controller].Hand.Contains(observed))
			}
			if !g.Players[game.Player1].Library.Contains(decoy) ||
				obj.ResolutionResults["if-you-do"].Amount != 41 || len(obj.ResolutionResults) != 1 ||
				len(obj.ResolutionResultObjects["if-you-do"]) != 1 || obj.ResolvedAmounts["if-you-do"] != 42 ||
				obj.ResolvedExcessDamage["if-you-do"] != 43 || len(obj.LocalLinkedProducts) != 0 || len(g.LinkedObjects) != 0 {
				t.Error("actual optional receipt/product leaked, aliased an enclosing value, or used a decoy")
			}
			if permanent := permanentForCard(g, observed); permanent != nil &&
				(permanent.Owner != controller || permanent.Controller != controller) {
				t.Error("placement confused source owner with observed-card owner/controller")
			}
			t.Logf("actual death isolation=%s choices=%d BF=%v library=%v exile=%v zoneVersion=%d",
				outcome, len(agent.prompts), wantBattlefield, wantLibrary, wantExile, g.CardInstances[observed].ZoneVersion)
		})
	}
}

func TestCompiledObservedPostfixAndPredicateLeadingChoices(t *testing.T) {
	for _, leading := range []bool{false, true} {
		conditional := "You may put that card onto the battlefield if it's a permanent card with mana value 3 or less."
		if leading {
			conditional = "If it's a permanent card with mana value 3 or less, you may put that card onto the battlefield."
		}
		def := compileUnlessCard(t, cardgen.ScryfallCard{
			Name: "Optional Condition Timing", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "Reveal the top card of your library. " + conditional + " Otherwise, put that card into your hand. You gain 1 life.",
		})
		for _, accept := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 40
			card := addCardToLibrary(g, game.Player1, compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Forest", Layout: "normal", TypeLine: "Basic Land — Forest", OracleText: "{T}: Add {G}.",
			}))
			agent := &libraryPaymentAgent{accept: accept}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if len(agent.prompts) != 1 || g.Players[game.Player1].Life != 41 ||
				(permanentForCard(g, card) != nil) != accept ||
				g.Players[game.Player1].Hand.Contains(card) != (!leading && !accept) ||
				g.Players[game.Player1].Library.Contains(card) != (leading && !accept) {
				t.Errorf("predicate-leading=%v accept=%v: choices=%d BF=%v hand=%v library=%v life=%d",
					leading, accept, len(agent.prompts), permanentForCard(g, card) != nil,
					g.Players[game.Player1].Hand.Contains(card), g.Players[game.Player1].Library.Contains(card), g.Players[game.Player1].Life)
			}
		}
	}
}
