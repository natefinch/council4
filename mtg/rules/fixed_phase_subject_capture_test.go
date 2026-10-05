package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func compiledCaptureSequence(t *testing.T, text string) []game.Instruction {
	t.Helper()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Capture Capability", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: %v; diagnostics: %v; definitions: %d", err, diagnostics, len(defs))
	}
	return defs[0].SpellAbility.Val.Modes[0].Sequence
}

func TestFixedPhaseTargetCaptureDoesNotFollowReentry(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "Tap target creature. Untap target creature. You gain 1 life. Exile it at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	decoy := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	subject := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{
		{Kind: game.TargetPermanent, PermanentID: decoy.ObjectID},
		{Kind: game.TargetPermanent, PermanentID: subject.ObjectID},
	}, TargetCounts: []int{1, 1}, SourceID: subject.ObjectID, SourceCardID: subject.CardInstanceID}
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedObjectID != subject.ObjectID {
		t.Fatal("nonzero target was not frozen independently of the decoy")
	}
	r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
	handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile})
	handlePutOnBattlefield(r, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource})})
	returned, ok := findPermanentByCardID(g, subject.CardInstanceID)
	if !ok || returned.ObjectID == subject.ObjectID {
		t.Fatal("subject did not reenter as a new incarnation")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if _, ok := permanentByObjectID(g, returned.ObjectID); !ok {
		t.Fatal("delayed disposal followed the card into its new incarnation")
	}
	if _, ok := permanentByObjectID(g, decoy.ObjectID); !ok {
		t.Fatal("delayed disposal acted on the qualifying decoy")
	}
}

func TestFixedPhaseActualProductsStayIndependent(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t,
		"Create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step. "+
			"Create a 1/1 green Insect creature token. You gain 1 life. Exile it at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	unrelated := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	for range 2 {
		engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	if len(g.DelayedTriggers) != 4 {
		t.Fatalf("scheduled triggers = %d, want four actual products", len(g.DelayedTriggers))
	}
	seen := make(map[game.ObjectID]bool)
	for _, trigger := range g.DelayedTriggers {
		if len(trigger.CapturedObjectIDs) != 1 || seen[trigger.CapturedObjectIDs[0]] {
			t.Fatal("independent producers or same-source resolutions aliased")
		}
		seen[trigger.CapturedObjectIDs[0]] = true
	}
	cloned := g.Clone()
	engine.runEndingPhase(cloned, [game.NumPlayers]PlayerAgent{})
	for objectID := range seen {
		if _, live := permanentByObjectID(cloned, objectID); live {
			t.Fatal("a captured actual product escaped its delayed action")
		}
		if _, live := permanentByObjectID(g, objectID); !live {
			t.Fatal("resolving a clone changed the original game's products")
		}
	}
	if _, live := permanentByObjectID(cloned, unrelated.ObjectID); !live {
		t.Fatal("delayed cleanup removed an unrelated creature")
	}
}

func TestFixedPhaseSkippedProductDoesNotCaptureStaleGroup(t *testing.T) {
	t.Parallel()
	for _, reason := range []string{"condition", "declined", "empty"} {
		t.Run(reason, func(t *testing.T) {
			t.Parallel()
			sequence := compiledCaptureSequence(t, "Create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			first := g.DelayedTriggers[0].CapturedObjectIDs[0]
			switch reason {
			case "condition":
				sequence[0].Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true, Negate: true})})
			case "declined":
				sequence[0].Optional = true
			case "empty":
				create := sequence[0].Primitive.(game.CreateToken)
				create.Amount = game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
				sequence[0].Primitive = create
			}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{}}, &TurnLog{})
			if len(g.DelayedTriggers) != 2 || len(g.DelayedTriggers[1].CapturedObjectIDs) != 0 {
				t.Fatal("unavailable actual product reused an earlier publication")
			}
			if g.DelayedTriggers[0].CapturedObjectIDs[0] != first {
				t.Fatal("a later skipped publisher changed the earlier frozen group")
			}
		})
	}
}

func TestFixedPhaseEventCardFreezesReachedIncarnation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	subject := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Subject", Types: []types.Card{types.Creature}}})
	obj := &game.StackObject{Controller: game.Player1, SourceID: subject.ObjectID, SourceCardID: subject.CardInstanceID}
	r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
	handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Graveyard})
	card, _ := g.GetCardInstance(subject.CardInstanceID)
	obj.HasTriggerEvent = true
	obj.TriggerEvent = game.Event{Kind: game.EventPermanentDied, CardID: card.ID, CardZoneVersion: card.ZoneVersion, PermanentID: subject.ObjectID}
	def := &game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep, CapturedCard: opt.Val(game.EventPermanentReference()),
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
			Card: game.CapturedCardReference(), FromZone: zone.Graveyard, Destination: zone.Hand,
		}}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, obj, def) || g.DelayedTriggers[0].CapturedCardID != card.ID {
		t.Fatal("event card was not captured in the reached graveyard incarnation")
	}
	handlePutOnBattlefield(r, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource})})
	returned, _ := findPermanentByCardID(g, card.ID)
	obj.SourceID = returned.ObjectID
	handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Graveyard})
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if g.Players[game.Player2].Hand.Contains(card.ID) || !g.Players[game.Player2].Graveyard.Contains(card.ID) {
		t.Fatal("delayed return followed a card that left and reentered its graveyard")
	}
}

func TestFixedPhaseConditionEvaluatesAtPrintedTime(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text   string
		future bool
	}{
		{"Tap target creature. If you have 20 or more life, destroy it at the beginning of the next end step.", false},
		{"Tap target creature. At the beginning of the next end step, destroy it if you have 20 or more life.", true},
		{"Tap target creature. At the beginning of the next end step, if you have 20 or more life, destroy it.", true},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			sequence := compiledCaptureSequence(t, test.text)
			for _, initiallyTrue := range []bool{false, true} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				subject := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
				g.Players[game.Player1].Life = 19
				if initiallyTrue {
					g.Players[game.Player1].Life = 20
				}
				obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}, TargetCounts: []int{1}}
				engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				wantScheduled := 0
				if test.future || initiallyTrue {
					wantScheduled = 1
				}
				if len(g.DelayedTriggers) != wantScheduled {
					t.Fatalf("scheduled=%d, want %d: current and future conditions were conflated", len(g.DelayedTriggers), wantScheduled)
				}
				g.Players[game.Player1].Life = 20
				if initiallyTrue {
					g.Players[game.Player1].Life = 19
				}
				engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
				_, survived := permanentByObjectID(g, subject.ObjectID)
				wantDestroyed := initiallyTrue
				if test.future {
					wantDestroyed = !initiallyTrue
				}
				if survived == wantDestroyed {
					t.Fatal("condition was not evaluated at its printed resolving time")
				}
			}
		})
	}
}

func TestFixedPhaseUnavailableConditionIsNotInverted(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, negate := range []bool{false, true} {
		condition := opt.Val(game.Condition{
			Negate: negate, Object: opt.Val(game.CapturedObjectReference()),
			ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		})
		if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: &game.StackObject{}}, condition) {
			t.Fatal("unavailable captured subject became true, possibly by negation")
		}
	}
}
