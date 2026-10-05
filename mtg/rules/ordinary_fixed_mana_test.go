package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const ordinaryOrdinalManaBody = "{2}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}{R}{R}{R}{R}{R}{R}."

type fixedManaChoiceAgent struct {
	accept bool
	may    int
}

func (*fixedManaChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *fixedManaChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		a.may++
		if a.accept {
			return []int{1}
		}
	}
	return []int{0}
}

func TestOrdinaryFixedManaPaidResolutions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, text  string
		accept      bool
		wantMana    int
		wantChoices int
		costColor   mana.Color
		costAmount  int
	}{
		{"optional accepted", ordinaryOrdinalManaBody, true, 8, 1, mana.C, 2},
		{"optional declined", ordinaryOrdinalManaBody, false, 0, 1, mana.C, 2},
		{"mandatory", "{R}: Target creature you control gains trample until end of turn. If this is the third time this ability has resolved this turn, add {R}{R}{R}{R}.", false, 4, 0, mana.R, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g, engine, source := ordinalFixture(t, tc.text)
			target := addCombatPermanent(g, game.Player1, creatureDef("Recipient"))
			agent := &fixedManaChoiceAgent{accept: tc.accept}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			for i := range 4 {
				g.Players[game.Player1].ManaPool.Empty()
				g.Players[game.Player1].ManaPool.Add(tc.costColor, tc.costAmount)
				obj := activateOrdinal(t, g, engine, source.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, nil)
				if !g.Players[game.Player1].ManaPool.IsEmpty() || obj.ResolutionOrdinalThisTurn != 0 {
					t.Fatal("announcement must pay the cost, not grant mana or increment the resolution ordinal")
				}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				want := 0
				if i == 2 {
					want = tc.wantMana
				}
				if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != want ||
					g.Players[game.Player1].ManaPool.Total() != want || !hasKeyword(g, target, game.Trample) {
					t.Fatalf("resolution %d: red=%d total=%d trample=%v, want mana=%d and unconditional trample",
						i+1, got, g.Players[game.Player1].ManaPool.Total(), hasKeyword(g, target, game.Trample), want)
				}
				if g.ResolvedActivatedAbilitiesThisTurn[obj.ActivatedResolutionUse.Val] != i+1 ||
					g.Players[game.Player2].ManaPool.Amount(mana.R) != 0 {
					t.Fatal("incorrect ordinal or mana leaked to another player")
				}
			}
			if agent.may != tc.wantChoices {
				t.Fatalf("may choices=%d, want %d", agent.may, tc.wantChoices)
			}
		})
	}
}

func TestOrdinaryFixedManaCopiesAndCapturedController(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, ordinaryOrdinalManaBody)
	target := addCombatPermanent(g, game.Player2, creatureDef("Recipient"))
	chosen := []game.Target{game.PermanentTarget(target.ObjectID)}
	first := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	second := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	copyObj := game.NewStackObjectCopy(first, g.IDGen.Next())
	copyObj.Controller = game.Player2
	g.Stack.Push(copyObj)
	agent1, agent2 := &fixedManaChoiceAgent{accept: true}, &fixedManaChoiceAgent{accept: true}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent1, game.Player2: agent2}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	source.Controller = game.Player2
	if _, ok := g.Stack.RemoveByID(first.ID); !ok {
		t.Fatal("first announcement disappeared")
	}
	g.Stack.Push(first)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if g.ResolvedActivatedAbilitiesThisTurn[second.ActivatedResolutionUse.Val] != 3 ||
		g.Players[game.Player1].ManaPool.Amount(mana.R) != 8 || g.Players[game.Player2].ManaPool.Amount(mana.R) != 0 ||
		agent1.may != 1 || agent2.may != 0 {
		t.Fatal("resolution order, copy identity, captured recipient, or optional choice count is wrong")
	}
}

func TestOrdinaryFixedManaCounteredFizzledAndIllegalTargets(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, ordinaryOrdinalManaBody)
	target := addCombatPermanent(g, game.Player2, creatureDef("Recipient"))
	chosen := []game.Target{game.PermanentTarget(target.ObjectID)}
	agent := &fixedManaChoiceAgent{accept: true}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	for range 2 {
		activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	countered := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	if !counterStackObject(g, countered.ID) {
		t.Fatal("counter failed")
	}
	fizzled := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("target did not leave")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if fizzled.ResolutionOrdinalThisTurn != 0 || g.Players[game.Player1].ManaPool.Amount(mana.R) != 0 || agent.may != 0 {
		t.Fatal("fizzle advanced the ordinal or offered mana")
	}
	if engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, chosen, 0)) {
		t.Fatal("an already illegal target was accepted")
	}
	target = addCombatPermanent(g, game.Player2, creatureDef("New Recipient"))
	chosen[0] = game.PermanentTarget(target.ObjectID)
	obj := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if g.ResolvedActivatedAbilitiesThisTurn[obj.ActivatedResolutionUse.Val] != 3 ||
		g.Players[game.Player1].ManaPool.Amount(mana.R) != 8 || agent.may != 1 {
		t.Fatal("third actual legal resolution did not produce exactly one optional 8R")
	}
}

func TestOrdinaryFixedManaSourceCloneAndTurnIsolation(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, ordinaryOrdinalManaBody)
	def, _ := permanentCardDef(g, source)
	other := addCombatPermanent(g, game.Player1, def)
	target := addCombatPermanent(g, game.Player2, creatureDef("Recipient"))
	chosen := []game.Target{game.PermanentTarget(target.ObjectID)}
	agent := &fixedManaChoiceAgent{accept: true}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	for range 2 {
		activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	independent := activateOrdinal(t, g, engine, other.ObjectID, 0, chosen, nil)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	old := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	cloned := g.Clone()
	engine.resolveTopOfStackWithChoices(cloned, agents, &TurnLog{})
	if cloned.Players[game.Player1].ManaPool.Amount(mana.R) != 8 ||
		g.Players[game.Player1].ManaPool.Amount(mana.R) != 0 ||
		g.ResolvedActivatedAbilitiesThisTurn[independent.ActivatedResolutionUse.Val] != 1 {
		t.Fatal("cloned tally/mana state or independent source identity leaked")
	}
	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("source did not leave")
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok {
		t.Fatal("source did not return")
	}
	fresh := activateOrdinal(t, g, engine, returned.ObjectID, 0, chosen, nil)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if g.ResolvedActivatedAbilitiesThisTurn[fresh.ActivatedResolutionUse.Val] != 1 ||
		g.ResolvedActivatedAbilitiesThisTurn[old.ActivatedResolutionUse.Val] != 3 ||
		fresh.ActivatedResolutionUse == old.ActivatedResolutionUse || g.Players[game.Player1].ManaPool.Amount(mana.R) != 8 {
		t.Fatal("returned incarnation shared the old tally or pending old-source body stopped resolving")
	}
	emptyManaPools(g)
	engine.advanceToNextTurn(g)
	if g.Players[game.Player1].ManaPool.Amount(mana.R) != 0 {
		t.Fatal("fixed mana leaked across turns")
	}
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	afterTurn := activateOrdinal(t, g, engine, returned.ObjectID, 0, chosen, nil)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if g.ResolvedActivatedAbilitiesThisTurn[afterTurn.ActivatedResolutionUse.Val] != 1 ||
		g.Players[game.Player1].ManaPool.Amount(mana.R) != 0 {
		t.Fatal("new turn reused prior resolutions")
	}
}

func TestOrdinaryFixedManaConditionTiming(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, text          string
		initialHand, red    int
		accept              bool
		may, wantHand, life int
	}{
		{"group snapshot", "If you have no cards in hand, draw a card, then add {R}{R}, then you gain 2 life.", 0, 2, false, 0, 1, 42},
		{"independent evaluation", "If you have no cards in hand, draw a card. If you have no cards in hand, add {R}{R}. You gain 2 life.", 0, 0, false, 0, 1, 42},
		{"false group with rider", "If you have no cards in hand, draw a card, then add {R}{R}. You gain 2 life.", 1, 0, false, 0, 1, 42},
		{"optional accept result", "You may add {R}{R}. If you do, draw a card.", 0, 2, true, 1, 1, 40},
		{"optional decline result", "You may add {R}{R}. If you do, draw a card.", 0, 0, false, 1, 0, 40},
		{"false optional gate", "You gain 2 life. If you have no cards in hand, you may add {R}{R}.", 1, 0, true, 0, 1, 42},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Mana Timing", Layout: "normal", TypeLine: "Instant", OracleText: tc.text})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for range tc.initialHand {
				addCardToHand(g, game.Player1, evidenceCard("Initial", 1))
			}
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			agent := &fixedManaChoiceAgent{accept: tc.accept}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, agents, &TurnLog{})
			if g.Players[game.Player1].ManaPool.Amount(mana.R) != tc.red || agent.may != tc.may ||
				g.Players[game.Player1].Hand.Size() != tc.wantHand || g.Players[game.Player1].Life != tc.life {
				t.Fatalf("red=%d may=%d hand=%d life=%d", g.Players[game.Player1].ManaPool.Amount(mana.R),
					agent.may, g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Life)
			}
		})
	}
}

func TestOrdinaryFixedManaGroupIndependentOfPrimitiveSuccess(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Mana Group", Layout: "normal", TypeLine: "Instant",
		OracleText: "If you control a creature, destroy target creature, then add {R}{G}.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, creatureDef("Condition Subject"))
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Protected", Types: []types.Card{types.Creature}, StaticAbilities: []game.StaticAbility{game.IndestructibleStaticBody},
	}})
	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if _, live := permanentByObjectID(g, target.ObjectID); !live ||
		g.Players[game.Player1].ManaPool.Amount(mana.R) != 1 || g.Players[game.Player1].ManaPool.Amount(mana.G) != 1 {
		t.Fatal("fixed output must use the captured group decision, not the failed destruction's success")
	}
}

func TestOrdinaryFixedManaModesAndTriggers(t *testing.T) {
	t.Parallel()
	t.Run("chosen modes count one whole resolution", func(t *testing.T) {
		t.Parallel()
		g, engine, source := ordinalFixture(t,
			"{1}: Choose one \u2014\n\u2022 Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}.\n\u2022 You gain 2 life.")
		target := addCombatPermanent(g, game.Player2, creatureDef("Recipient"))
		agent := &fixedManaChoiceAgent{accept: true}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
		for i, mode := range []int{1, 0, 0, 0} {
			g.Players[game.Player1].ManaPool.Empty()
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			var targets []game.Target
			if mode == 0 {
				targets = []game.Target{game.PermanentTarget(target.ObjectID)}
			}
			obj := activateOrdinal(t, g, engine, source.ObjectID, 0, targets, []int{mode})
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.ResolvedActivatedAbilitiesThisTurn[obj.ActivatedResolutionUse.Val] != i+1 {
				t.Fatal("chosen modes did not count the whole ability exactly once")
			}
			want := 0
			if i == 2 {
				want = 2
			}
			if g.Players[game.Player1].ManaPool.Amount(mana.R) != want {
				t.Fatalf("resolution %d: red=%d, want %d", i+1, g.Players[game.Player1].ManaPool.Amount(mana.R), want)
			}
		}
		if agent.may != 1 || g.Players[game.Player1].Life != 42 {
			t.Fatal("ungated mode failed to count, or targeted mode output/optional envelope was wrong")
		}
	})
	t.Run("trigger stack body", func(t *testing.T) {
		t.Parallel()
		def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Mana Trigger", Layout: "normal", TypeLine: "Artifact",
			OracleText: "Whenever a creature you control enters, target creature gains first strike until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}."})
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatPermanent(g, game.Player1, def)
		target := addCombatPermanent(g, game.Player2, creatureDef("Recipient"))
		agent := &fixedManaChoiceAgent{accept: true}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
		engine := NewEngine(nil)
		for range 4 {
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, Controller: game.Player1, AbilityIndex: 0,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)},
			})
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if !hasKeyword(g, target, game.FirstStrike) {
				t.Fatal("ordinary trigger's independent first-strike effect was gated")
			}
		}
		if g.Players[game.Player1].ManaPool.Amount(mana.R) != 2 || agent.may != 1 ||
			len(g.ResolvedActivatedAbilitiesThisTurn) != 0 {
			t.Fatal("triggered mana output/choices or separate resolution tally was wrong")
		}
	})
}

func TestOrdinaryFixedManaMissingAndFalseGroupDecisions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name            string
		publish, negate bool
	}{
		{"missing decision", false, false},
		{"false decision", true, true},
		{"true decision despite skipped producer", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			publisher := game.Instruction{
				Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true, Negate: tc.negate})}),
				ResultGate: opt.Val(game.InstructionResultGate{
					Key: "unavailable-result", Succeeded: game.TriTrue,
				}),
				PublishCondition: "mana-decision",
			}
			if !tc.publish {
				publisher.PublishCondition = ""
			}
			content := game.Mode{Sequence: []game.Instruction{publisher, {
				Primitive: game.AddMana{Amount: game.Fixed(8), ManaColor: mana.R}, ConditionGate: "mana-decision", Optional: true,
			}}}.Ability()
			agent := &fixedManaChoiceAgent{accept: true}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			wantMana, wantChoices := 0, 0
			if tc.publish && !tc.negate {
				wantMana, wantChoices = 8, 1
			}
			if g.Players[game.Player1].ManaPool.Amount(mana.R) != wantMana || agent.may != wantChoices {
				t.Fatal("absent/false gate leaked mana, or condition evaluation was confused with primitive success")
			}
		})
	}
}
