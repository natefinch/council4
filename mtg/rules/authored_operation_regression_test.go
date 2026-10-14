package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestAuthoredCopiesKeepOwnerAndIsolateManaReceipts(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	addImplementationSpellToStack(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Copy subject", Types: []types.Card{types.Instant},
	}}, nil)
	original, _ := g.Stack.Peek()
	original.ManaSpentByColorToCast = map[color.Color]int{color.Blue: 2}
	obj := &game.StackObject{
		Controller: game.Player1, HasTriggerEvent: true, TriggerEvent: game.Event{StackObjectID: original.ID},
	}
	ctx := &programContext{engine: engine, game: g, obj: obj, active: true, log: &TurnLog{}}
	outcome, err := ctx.Copy(game.CopyStackObject{Object: game.EventStackObjectReference()}, game.Player3)
	if err != nil || len(outcome.StackIDs) != 1 {
		t.Fatal("copy operation did not report the actually created stack object")
	}
	copied, _ := stackObjectByID(g, outcome.StackIDs[0])
	copied.Controller = game.Player4
	owner, known := (&resolvedObjectReference{stack: copied}).owner(g)
	if !known || owner != game.Player3 {
		t.Fatal("copy owner followed a later control change or borrowed the physical card's owner")
	}
	clone := g.Clone()
	copied.ManaSpentByColorToCast[color.Blue] = 99
	independent, _ := stackObjectByID(clone, copied.ID)
	if original.ManaSpentByColorToCast[color.Blue] != 2 || independent.ManaSpentByColorToCast[color.Blue] != 2 {
		t.Fatal("spell copies or game clones share a mutable mana payment receipt")
	}
}

func TestAuthoredExchangeFailureIsAtomic(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"source missing", "target missing", "same object", "same controller", "phased target", "departed recipient"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addCombatCreaturePermanent(g, game.Player1)
			second := addCombatCreaturePermanent(g, game.Player2)
			first.SummoningSick, second.SummoningSick = false, false
			obj := &game.StackObject{Controller: game.Player1, SourceID: first.ObjectID,
				Targets: []game.Target{game.PermanentTarget(second.ObjectID)}}
			switch scenario {
			case "source missing":
				movePermanentToZone(g, first, zone.Exile)
			case "target missing":
				movePermanentToZone(g, second, zone.Exile)
			case "same object":
				obj.Targets[0] = game.PermanentTarget(first.ObjectID)
			case "same controller":
				second.Controller = game.Player1
			case "phased target":
				second.PhasedOut = true
			case "departed recipient":
				g.Players[game.Player2].Eliminated = true
			default:
				t.Fatal("unknown scenario")
			}
			before := g.Clone()
			ctx := &programContext{engine: choicesProgramEngine(t), game: g, obj: obj, active: true, log: &TurnLog{},
				specs: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}}}
			outcome, err := ctx.Exchange(game.SourcePermanentReference(), game.TargetPermanentReference(0))
			if err != nil || outcome.Exchanged || len(outcome.After) != 0 ||
				!reflect.DeepEqual(g.ContinuousEffects, before.ContinuousEffects) ||
				!reflect.DeepEqual(g.Battlefield, before.Battlefield) || !reflect.DeepEqual(g.Events, before.Events) ||
				g.IDGen != before.IDGen {
				t.Fatal("failed exchange changed part of the board, created effects or consumed mutation IDs")
			}
		})
	}
}

func TestAuthoredGildedDrakeCannotSacrificeReincarnatedSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	def := authored.GildedDrake.New()
	drake := addCombatPermanent(g, game.Player1, def)
	target := addCombatCreaturePermanent(g, game.Player2)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: drake.ObjectID,
		SourceCardID: drake.CardInstanceID, Controller: game.Player1, InlineTrigger: &def.TriggeredAbilities[0],
		Targets: []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
	})
	movePermanentToZone(g, drake, zone.Exile)
	g.Players[game.Player1].Exile.Remove(drake.CardInstanceID)
	card, _ := g.GetCardInstance(drake.CardInstanceID)
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || returned.ObjectID == drake.ObjectID {
		t.Fatal("blink fixture did not create a new incarnation")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, card.ID) != returned || effectiveController(g, target) != game.Player2 {
		t.Fatal("exchange or sacrifice reused the new source incarnation")
	}
}

func TestAuthoredExchangeResetsSicknessAndCombatMembership(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player2)
	first.SummoningSick, second.SummoningSick = false, false
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	obj := &game.StackObject{Controller: game.Player1, SourceID: first.ObjectID,
		Targets: []game.Target{game.PermanentTarget(second.ObjectID)}}
	ctx := &programContext{engine: choicesProgramEngine(t), game: g, obj: obj, active: true, log: &TurnLog{},
		specs: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}}}
	outcome, err := ctx.Exchange(game.SourcePermanentReference(), game.TargetPermanentReference(0))
	if err != nil || !outcome.Exchanged || !first.SummoningSick || !second.SummoningSick || len(g.Combat.Attackers) != 0 {
		t.Fatal("successful exchange omitted control-change bookkeeping")
	}
}

func TestAuthoredExchangeControlGrantEndsWhenRecipientLeaves(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: first.ObjectID,
		Targets: []game.Target{game.PermanentTarget(second.ObjectID)}}
	engine := choicesProgramEngine(t)
	ctx := &programContext{engine: engine, game: g, obj: obj, active: true, log: &TurnLog{},
		specs: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}}}
	outcome, err := ctx.Exchange(game.SourcePermanentReference(), game.TargetPermanentReference(0))
	if err != nil || !outcome.Exchanged {
		t.Fatal("fixture failed to exchange")
	}
	second.SummoningSick = false
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player3}}},
	}
	engine.eliminatePlayer(g, game.Player1)
	if effectiveController(g, second) != game.Player2 || !second.SummoningSick || len(g.Combat.Attackers) != 0 {
		t.Fatal("exchange effect continued giving an eliminated player control")
	}
}

func TestAuthoredFoundCardHandleDoesNotSurviveShuffle(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Revealed"}})
	ctx := &programContext{
		engine: choicesProgramEngine(t), game: g, obj: &game.StackObject{Controller: game.Player1},
		active: true, log: &TurnLog{},
	}

	found, err := ctx.Search(game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(1),
		Spec: game.SearchSpec{SourceZone: zone.Library, Reveal: true, RevealOnly: true},
	})
	if err != nil || len(found.Cards) != 1 {
		t.Fatal("revealed search fixture failed")
	}
	if _, err := ctx.Execute(game.ShuffleLibrary{Player: game.ControllerReference()}); err != nil {
		t.Fatal(err)
	}
	moved, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Hand})
	if err != nil || len(moved) != 0 || !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("shuffle preserved a correlation to a now-hidden new object")
	}
	clone := g.Clone()
	if clone.Players[game.Player1].Library.ReorderVersion() != g.Players[game.Player1].Library.ReorderVersion() {
		t.Fatal("clone lost library correlation invalidation state")
	}
}

func TestAuthoredRevealSearchRejectsUnimplementedRidersBeforeChoices(t *testing.T) {
	t.Parallel()
	for _, spec := range []game.SearchSpec{
		{AlsoGraveyard: true},
		{ConditionalShuffle: true},
		{ExileFaceDown: true},
		{AnyNumber: true},
		{SlotFilters: []game.Selection{{RequiredTypes: []types.Card{types.Creature}}}},
	} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addCardToLibrary(g, game.Player1, vanillaCreature("Search candidate", 2, 2))
		spec.SourceZone, spec.Reveal, spec.RevealOnly = zone.Library, true, true
		ctx := &programContext{
			engine: choicesProgramEngine(t), game: g, obj: &game.StackObject{Controller: game.Player1},
			active: true, log: &TurnLog{},
		}
		if _, err := ctx.Search(game.Search{Player: game.ControllerReference(), Amount: game.Fixed(4), Spec: spec}); err == nil {
			t.Fatal("a reveal-only search silently ignored a rider its shared operation cannot execute")
		}
		if len(ctx.log.Choices) != 0 || len(g.Events) != 0 || g.Players[game.Player1].Library.Size() != 1 {
			t.Fatal("invalid search changed the game or asked a success-shaped choice")
		}
	}
}
