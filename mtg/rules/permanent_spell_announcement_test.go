package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPermanentSpellCopyPreservesPaidXWithoutCast(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, x := range []int{0, 5} {
		for _, double := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := permanentCopyEngine(t)
			if double {
				addReplacementPermanent(t, g, game.Player1, registry.Lookup("Doubling Season"))
			}
			cardID := addCardToHand(g, game.Player1, registry.Lookup("Stonecoil Serpent"))
			g.Players[game.Player1].ManaPool.Add(mana.C, x)
			setMainPhasePriority(g, game.Player1)
			if !e.applyAction(g, game.Player1, action.CastSpell(cardID, nil, x, nil)) {
				t.Fatalf("Stonecoil Serpent paid X=%d cast failed", x)
			}
			original, _ := g.Stack.Peek()
			copied := copyPermanentSpell(t, e, g, original, game.Player1)
			if copied.XValue != x {
				t.Fatal("copy construction lost X")
			}
			if !counterStackObject(g, original.ID) {
				t.Fatal("could not remove the original spell")
			}
			start := len(g.Events)
			branch := e.Simulator().ResolvePriority(g, simPassPolicies())
			e.resolveTopOfStack(g, &TurnLog{})
			token := g.Battlefield[len(g.Battlefield)-1]
			want := x
			if double {
				want *= 2
			}
			if !token.Token || token.EnteredFromCast || !token.HasSpellAnnouncement || token.CastXValue != x ||
				token.Counters.Get(counter.PlusOnePlusOne) != want {
				t.Fatalf("copied X=%d with double=%v: %+v", x, double, token)
			}
			if x > 0 {
				simToken, found := permanentByObjectID(branch, token.ObjectID)
				if !found || simToken.CastXValue != x || simToken.Counters.Get(counter.PlusOnePlusOne) != want {
					t.Fatal("Simulator lost copied X or depended on the original")
				}
			}
			for _, event := range g.Events[start:] {
				if event.Kind == game.EventTokenCreated || event.Kind == game.EventSpellCast {
					t.Fatal("recovering copied X fabricated creation/cast")
				}
				if event.Kind == game.EventPermanentEnteredBattlefield &&
					(event.EnterWasCast || event.EnterHasCastController || event.EnterCastFromZone != zone.None ||
						!event.EnterHasSpellAnnouncement || event.EnterXValue != x) {
					t.Fatalf("entry event disagrees with the copied spell: %+v", event)
				}
			}
			if x > 0 {
				copiedDef, ok := permanentCopyDef(g, token)
				if !ok {
					t.Fatal("could not snapshot the resulting permanent")
				}
				ordinary, ok := createTokenPermanentWithChoices(e, g, game.Player1, copiedDef, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if !ok || ordinary.HasSpellAnnouncement || ordinary.CastXValue != 0 || ordinary.Counters.Get(counter.PlusOnePlusOne) != 0 {
					t.Fatal("ordinary permanent copy inherited historical X/counters")
				}
			}
		}
	}
}

func TestPermanentSpellCopyRetainsKickedChoiceNotManaSpend(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := permanentCopyEngine(t)
	def := kickedEntersWithCountersCreature()
	def.ReplacementAbilities = append(def.ReplacementAbilities,
		game.EntersWithCountersIfReplacement("Only if cast from your hand.",
			&game.Condition{EventPermanentWasCastFromControllerHand: true},
			game.CounterPlacement{Kind: counter.Charge, Amount: 1}),
		game.EntersWithCountersIfReplacement("Only if three green mana was spent.",
			&game.Condition{SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Green, Count: 3}},
			game.CounterPlacement{Kind: counter.Fade, Amount: 1}),
		game.EntersWithCountersReplacement("One per color spent.", game.CounterPlacement{
			Kind: counter.Time, Dynamic: opt.Val(&game.DynamicAmount{Kind: game.DynamicAmountColorsOfManaSpentToCast, Multiplier: 1}),
		}),
	)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	setMainPhasePriority(g, game.Player1)
	if !e.applyAction(g, game.Player1, action.CastKickedSpell(cardID, nil, 0, nil)) {
		t.Fatal("paid kicked cast failed")
	}
	original, _ := g.Stack.Peek()
	copied := copyPermanentSpell(t, e, g, original, game.Player1)
	// These remain on the copied stack receipt but must not be interpreted as
	// actual spending on the copy when entry evaluates Converge/Adamant.
	copied.ManaSpentByColorToCast[color.Green] = 99
	copied.ColorsOfManaSpentToCast = 5
	e.resolveTopOfStack(g, &TurnLog{})
	token := g.Battlefield[len(g.Battlefield)-1]
	if token.Counters.Get(counter.PlusOnePlusOne) != 2 || token.Counters.Get(counter.Charge) != 0 ||
		token.Counters.Get(counter.Fade) != 0 || token.Counters.Get(counter.Time) != 0 || token.EnteredFromCast {
		t.Fatalf("copied decisions and actual casting were conflated: %+v", token)
	}
	if original.ManaSpentByColorToCast[color.Green] == 99 {
		t.Fatal("copy shared mutable announcement receipts")
	}
	event := copyEntryEvent(t, g, token)
	if !event.KickerPaid || event.EnterWasCast {
		t.Fatal("kicked entry event lost the paid choice or fabricated a cast")
	}
}

func TestPermanentSpellCopyEntryTriggersUseAnnouncedX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := permanentCopyEngine(t)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Ravenous draw"}})
	addImplementationSpellToStack(g, game.Player1, ravenousCreatureDef(), nil)
	original, _ := g.Stack.Peek()
	original.XValue = 5
	copyPermanentSpell(t, e, g, original, game.Player1)
	g.TriggerEventCursor = len(g.Events)
	e.resolveTopOfStack(g, &TurnLog{})
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("copied Ravenous X=5 did not trigger")
	}
	trigger, _ := g.Stack.Peek()
	if trigger.XValue != 5 || trigger.TriggerEvent.EnterWasCast {
		t.Fatal("trigger lost copied X or invented a cast")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("copied Ravenous trigger did not draw")
	}
}

func TestPermanentSpellCopyPreservesAdditionalAndAlternativeDecisions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := permanentCopyEngine(t)
	def := vanillaCreature("Copied paid decisions", 2, 2)
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersIfReplacement("Bargained entry.",
			&game.Condition{EventPermanentWasBargained: true}, game.CounterPlacement{Kind: counter.Charge, Amount: 1}),
	}
	addImplementationSpellToStack(g, game.Player1, def, nil)
	original, _ := g.Stack.Peek()
	original.Bargained, original.OffspringPaid, original.Evoked, original.Dashed = true, true, true, true
	copyPermanentSpell(t, e, g, original, game.Player1)
	e.resolveTopOfStack(g, &TurnLog{})
	token := g.Battlefield[len(g.Battlefield)-1]
	event := copyEntryEvent(t, g, token)
	if token.Counters.Get(counter.Charge) != 1 || !event.Bargained || !event.OffspringPaid ||
		!event.EnterEvoked || !event.EnterDashed || event.EnterWasCast {
		t.Fatal("copy dropped supported additional/alternative cost decisions")
	}
}

func TestPermanentSpellCopyPaidMechanicsExecute(t *testing.T) {
	for _, mechanic := range []string{"evoke", "offspring", "bestow"} {
		t.Run(mechanic, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := permanentCopyEngine(t)
			def := evokeTestCreature()
			if mechanic == "offspring" {
				def = cards.NewDefaultRegistry().Lookup("Coruscation Mage")
			} else if mechanic == "bestow" {
				def = bestowTestCreature()
			}
			host := addCombatCreaturePermanent(g, game.Player1)
			drawn := addCardToLibrary(g, game.Player1, vanillaCreature("Drawn", 2, 2))
			var targets []game.Target
			if mechanic == "bestow" {
				targets = []game.Target{game.PermanentTarget(host.ObjectID)}
			}
			addImplementationSpellToStack(g, game.Player1, def, targets)
			original, _ := g.Stack.Peek()
			original.Evoked, original.OffspringPaid, original.Bestowed = mechanic == "evoke", mechanic == "offspring", mechanic == "bestow"
			copyPermanentSpell(t, e, g, original, game.Player1)
			counterStackObject(g, original.ID)
			resolveStackWithTriggers(e, g, [game.NumPlayers]PlayerAgent{})
			tokens := 0
			for _, permanent := range g.Battlefield {
				if permanent.Token {
					tokens++
					if mechanic == "bestow" && (!permanent.Bestowed || !permanent.AttachedTo.Exists ||
						permanent.AttachedTo.Val != host.ObjectID) {
						t.Fatal("bestowed copy lost its Aura status or target")
					}
				}
			}
			want := map[string]int{"evoke": 0, "offspring": 2, "bestow": 1}[mechanic]
			if tokens != want || mechanic == "evoke" && !g.Players[game.Player1].Hand.Contains(drawn) {
				t.Fatalf("copied %s did not execute its paid mechanic without recursion: tokens %d", mechanic, tokens)
			}
		})
	}
}

func copyEntryEvent(t *testing.T, g *game.Game, permanent *game.Permanent) game.Event {
	t.Helper()
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
			return event
		}
	}
	t.Fatal("copy entry event missing")
	return game.Event{}
}
