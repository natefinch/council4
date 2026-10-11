package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledEventEntryCounterAcrossVerbs(t *testing.T) {
	for _, verb := range []string{"put that card onto", "return that card to"} {
		t.Run(verb, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Entry Watcher", Layout: "normal", TypeLine: "Creature",
				Power: new("1"), Toughness: new("1"),
				OracleText: "Whenever an opponent sacrifices a nontoken permanent, " +
					verb + " the battlefield tapped under your control with a +1/+1 counter on it.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player2, def)
			victim := addCombatPermanent(g, game.Player3, vanillaCreature("Recipient", 2, 3))
			victim.Controller = game.Player1
			if !sacrificePermanent(g, victim) {
				t.Fatal("ordinary sacrifice failed")
			}
			card, _ := g.GetCardInstance(victim.CardInstanceID)
			before := card.ZoneVersion
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("compiled entry did not queue exactly once")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			returned, ok := findPermanentByCardID(g, card.ID)
			if !ok || returned.ObjectID == victim.ObjectID || returned.Owner != game.Player3 ||
				returned.Controller != game.Player2 || !returned.Tapped ||
				returned.Counters.Get(counter.PlusOnePlusOne) != 1 || card.ZoneVersion != before+1 {
				t.Fatalf("exact entry/counter/controller/version mismatch: returned=%+v version=%d", returned, card.ZoneVersion)
			}
		})
	}
}

func TestCompiledPaidEventEntryRequiresActualPayment(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Tenacious Dead", Layout: "normal", TypeLine: "Creature — Skeleton Warrior",
		ManaCost: "{B}", Power: new("1"), Toughness: new("1"),
		OracleText: "When this creature dies, you may pay {1}{B}. If you do, return it to the battlefield tapped under its owner's control.",
	})
	for _, scenario := range []string{"paid", "declined", "unpayable", "stale", "diverted"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			victim := addCombatPermanent(g, game.Player3, def)
			victim.Controller = game.Player2
			decoy := addCardToGraveyard(g, game.Player3, def)
			if scenario != "unpayable" {
				g.Players[game.Player2].ManaPool.Add(mana.B, 1)
				g.Players[game.Player2].ManaPool.Add(mana.C, 1)
			}
			g.Players[game.Player3].ManaPool.Add(mana.B, 2)
			if !sacrificePermanent(g, victim) {
				t.Fatal("ordinary death failed")
			}
			card, _ := g.GetCardInstance(victim.CardInstanceID)
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("paid death body did not queue once")
			}
			if scenario == "stale" {
				if !moveCardBetweenZones(g, game.Player3, card.ID, zone.Graveyard, zone.Exile) ||
					!moveCardBetweenZones(g, game.Player3, card.ID, zone.Exile, zone.Graveyard) {
					t.Fatal("stale-card control failed")
				}
			}

			before := card.ZoneVersion
			if scenario == "diverted" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Graveyard,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
				})
			}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player2: optionalMayAgent{accept: scenario != "declined"},
			}, &TurnLog{})
			returned, ok := findPermanentByCardID(g, card.ID)
			if ok != (scenario == "paid") ||
				ok && (returned.Owner != game.Player3 || returned.Controller != game.Player3 ||
					!returned.Tapped || returned.ObjectID == victim.ObjectID || card.ZoneVersion != before+1) ||
				!g.Players[game.Player3].Graveyard.Contains(decoy) {
				t.Fatalf("payment/owner/incarnation/decoy mismatch: returned=%+v", returned)
			}
			if scenario == "declined" && g.Players[game.Player2].ManaPool.Total() != 2 ||
				scenario != "declined" && g.Players[game.Player2].ManaPool.Total() != 0 ||
				g.Players[game.Player3].ManaPool.Total() != 2 {
				t.Fatal("payment used wrong player, paid on decline, or failed to consume actual cost")
			}
			if (scenario == "declined" || scenario == "unpayable" || scenario == "stale") &&
				card.ZoneVersion != before {
				t.Fatal("failed/stale entry changed card incarnation")
			}
		})
	}
}

func TestCompiledFrontFaceReturnPreservesWholeCard(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Vincent Valentine // Galian Beast", Layout: "transform",
		TypeLine: "Legendary Creature — Assassin // Legendary Creature — Werewolf Beast",
		CardFaces: []cardgen.ScryfallCardFace{{
			Name: "Vincent Valentine", ManaCost: "{2}{B}{B}",
			TypeLine: "Legendary Creature — Assassin", Power: new("2"), Toughness: new("2"),
			OracleText: "Whenever a creature an opponent controls dies, put a number of +1/+1 counters on Vincent Valentine equal to that creature's power.\nWhenever Vincent Valentine attacks, you may transform it.",
		}, {
			Name: "Galian Beast", TypeLine: "Legendary Creature — Werewolf Beast",
			Power: new("3"), Toughness: new("2"),
			OracleText: "Trample, lifelink\nWhen Galian Beast dies, return it to the battlefield tapped (front face up).",
		}},
	})
	for _, stale := range []bool{false, true} {
		t.Run(map[bool]string{false: "exact back-face death", true: "stale physical card"}[stale], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			victim := addCombatPermanent(g, game.Player3, def)
			victim.Controller = game.Player2
			victim.Transformed = true
			victim.Face = game.FaceBack
			if !sacrificePermanent(g, victim) {
				t.Fatal("back-face death failed")
			}
			card, _ := g.GetCardInstance(victim.CardInstanceID)
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("whole-card back-face death did not queue one return")
			}
			if stale {
				if !moveCardBetweenZones(g, game.Player3, card.ID, zone.Graveyard, zone.Exile) ||
					!moveCardBetweenZones(g, game.Player3, card.ID, zone.Exile, zone.Graveyard) {
					t.Fatal("stale incarnation control failed")
				}
			}
			before := card.ZoneVersion
			engine.resolveTopOfStack(g, &TurnLog{})
			returned, ok := findPermanentByCardID(g, card.ID)
			if ok == stale || ok && (returned.Transformed || returned.Face != game.FaceFront || !returned.Tapped ||
				returned.Owner != game.Player3 || returned.Controller != game.Player3 ||
				returned.ObjectID == victim.ObjectID || card.ZoneVersion != before+1) ||
				stale && card.ZoneVersion != before {
				t.Fatalf("explicit front-face return followed wrong face/owner/incarnation: %+v", returned)
			}
		})
	}
}
