package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Recover Oracle text in the pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 for
// Grim Harvest, Resize, Icefall, Sun's Bounty, Controvert and Krovikan Rot.
// Exact printings, Oracle IDs and empty Scryfall ruling responses are recorded
// in the retained Recover family probe. Wizards CR effective 2026-09-25,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (verified local SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 702.59a, 603.6c/603.10a and 700.4 require the source in the graveyard before
// the death; 400.3 and 113.8 use the owners, not the battlefield controllers;
// 118.12 and 608.2g govern optional payment; 113.7a and 400.7 keep the trigger
// independent of its source but prevent moving a later incarnation.
func recoverTriggerFixture(t *testing.T) *game.CardDef {
	t.Helper()
	source := game.CardReference{Kind: game.CardReferenceSource}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Trigger Fixture",
		Types: []types.Card{types.Instant},
		TriggeredAbilities: []game.TriggeredAbility{{
			ZoneOfFunction: zone.Graveyard,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:            game.EventPermanentDied,
					Player:           game.TriggerPlayerYou,
					SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{
				{
					Primitive: game.Pay{Payment: game.ResolutionPayment{
						Prompt:   "Pay {1}{B}?",
						ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					}},
					PublishResult: "paid",
				},
				{
					Primitive:  game.MoveCard{Card: source, FromZone: zone.Graveyard, Destination: zone.Hand},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriTrue}),
				},
				{
					Primitive:  game.MoveCard{Card: source, FromZone: zone.Graveyard, Destination: zone.Exile},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriFalse}),
				},
			}}.Ability(),
		}},
	}}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("Recover fixture not admitted: %+v", issues)
	}
	return def
}

func TestRecoverGraveyardTriggerPaymentAndIncarnation(t *testing.T) {
	def := recoverTriggerFixture(t)
	for _, tc := range []struct {
		name        string
		pay         bool
		affordable  bool
		incarnation string
		want        zone.Type
	}{
		{name: "pay", pay: true, affordable: true, want: zone.Hand},
		{name: "decline", affordable: true, want: zone.Exile},
		{name: "unaffordable colored component", pay: true, want: zone.Exile},
		{name: "return before priority and clone", pay: true, affordable: true, incarnation: "before priority", want: zone.Graveyard},
		{name: "return after stacking and decline copied trigger", affordable: true, incarnation: "copied trigger", want: zone.Graveyard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			const owner = game.Player2
			source := addCardToGraveyard(g, owner, def)
			version := g.CardInstances[source].ZoneVersion
			dying := addCombatCreaturePermanent(g, owner)
			// A creature you own can die into your graveyard under someone
			// else's control. Their identical source must not trigger.
			dying.Controller = game.Player1
			decoy := addCardToGraveyard(g, game.Player1, def)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a real death")
			}
			leaveAndReturn := func() {
				t.Helper()
				if !moveCardBetweenZonesWithPlacement(g, owner, source, zone.Graveyard, zone.Exile, false) ||
					!moveCardBetweenZonesWithPlacement(g, owner, source, zone.Exile, zone.Graveyard, false) {
					t.Fatal("failed to change source incarnation")
				}
			}
			if tc.incarnation == "before priority" {
				leaveAndReturn()
				g = g.Clone()
			}
			agent := &repeatLandChoiceAgent{mayAnswers: []bool{tc.pay}}
			agents := [game.NumPlayers]PlayerAgent{owner: agent}
			pool := &g.Players[owner].ManaPool
			if tc.affordable {
				pool.Add(mana.B, 1)
				pool.Add(mana.C, 1)
			} else {
				pool.Add(mana.C, 2)
			}
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) || g.Stack.Size() != 1 {
				t.Fatalf("expected exactly the owner's trigger, got %d", g.Stack.Size())
			}
			trigger, _ := g.Stack.Peek()
			if trigger.Kind != game.StackTriggeredAbility || trigger.Controller != owner ||
				trigger.SourceCardID != source || trigger.SourceZone != zone.Graveyard ||
				trigger.SourceZoneVersion != version || trigger.TriggerEvent.PermanentID != dying.ObjectID ||
				len(trigger.Targets) != 0 || agent.mayIndex != 0 {
				t.Fatalf("wrong trigger identity, target, or timing: %+v, prompts=%d", trigger, agent.mayIndex)
			}
			if tc.incarnation == "copied trigger" {
				leaveAndReturn()
				_, _ = g.Stack.Pop()
				g.Stack.Push(game.NewStackObjectCopy(trigger, g.IDGen.Next()))
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if got, _ := cardZone(g, source); got != tc.want || !g.Players[game.Player1].Graveyard.Contains(decoy) {
				t.Fatalf("source ended in %v, want %v; opponent's identical source must stay put", got, tc.want)
			}
			wantMana := 2
			if tc.pay && tc.affordable {
				wantMana = 0
			}
			if got := pool.Total(); got != wantMana {
				t.Fatalf("mana after resolution = %d, want %d", got, wantMana)
			}
		})
	}
}

func TestRecoverGraveyardDeathEventBoundaries(t *testing.T) {
	def := recoverTriggerFixture(t)
	t.Run("one trigger per creature, not for simultaneous arrivals", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCardToGraveyard(g, game.Player1, def)
		// Its front face has Recover, but it is a face-down creature on
		// the battlefield until the same event that kills the other creature.
		arriving := addCombatPermanent(g, game.Player1, def)
		arriving.FaceDown = true
		creature := addCombatCreaturePermanent(g, game.Player1)
		if !sacrificePermanentsSimultaneously(g, []*game.Permanent{arriving, creature}) {
			t.Fatal("could not sacrifice both fixtures")
		}
		if !g.Players[game.Player1].Graveyard.Contains(arriving.CardInstanceID) {
			t.Fatal("missing face-down creature's card in graveyard")
		}
		if !NewEngine(nil).putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
			t.Fatalf("expected two triggers from the already-present source, got %d", g.Stack.Size())
		}
		seen := map[game.ObjectID]bool{}
		for _, trigger := range g.Stack.Objects() {
			if trigger.SourceCardID != source {
				t.Fatal("a card arriving in the same event triggered")
			}
			seen[trigger.TriggerEvent.PermanentID] = true
		}
		if !seen[arriving.ObjectID] || !seen[creature.ObjectID] {
			t.Fatal("triggers did not follow the two individual deaths")
		}
	})
	for _, scenario := range []string{"noncreature", "opponent's graveyard", "replaced death", "late source", "battlefield default"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCombatCreaturePermanent(g, game.Player1)
			switch scenario {
			case "noncreature":
				creature = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Noncreature Fixture", Types: []types.Card{types.Artifact},
				}})
			case "opponent's graveyard":
				creature = addCombatCreaturePermanent(g, game.Player2)
			case "replaced death":
				graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, true)
			}
			switch scenario {
			case "late source":
				card := addCardToHand(g, game.Player1, def)
				if !sacrificePermanent(g, creature) ||
					!moveCardBetweenZonesWithPlacement(g, game.Player1, card, zone.Hand, zone.Graveyard, false) {
					t.Fatal("failed to arrange death before source arrival")
				}
			case "battlefield default":
				defaultDef := recoverTriggerFixture(t)
				defaultDef.TriggeredAbilities[0].ZoneOfFunction = zone.None
				addCardToGraveyard(g, game.Player1, defaultDef)
				if !sacrificePermanent(g, creature) {
					t.Fatal("failed to sacrifice creature")
				}
			default:
				addCardToGraveyard(g, game.Player1, def)
				if !sacrificePermanent(g, creature) {
					t.Fatal("failed to sacrifice permanent")
				}
				if scenario == "replaced death" && !g.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
					t.Fatal("replacement did not exile the sacrificed creature")
				}
			}
			if NewEngine(nil).putTriggeredAbilitiesOnStack(g) {
				t.Fatal("queued a trigger for a nonmatching event or source")
			}
		})
	}
}
