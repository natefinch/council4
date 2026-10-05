package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCounterQuantityActualRemoval(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                  string
		have, requested, want int
		choose                bool
	}{
		{"fewer than requested", 2, 5, 2, false},
		{"zero observed", 0, 5, 0, false},
		{"zero requested", 3, 0, 0, false},
		{"choose kind unavailable", 0, 1, 0, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, evidenceCard("Counter Source", 1))
			source.Counters.Add(counter.Charge, tt.have)
			obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1}
			sequence := []game.Instruction{
				{Primitive: game.RemoveCounter{
					Object: game.SourcePermanentReference(), CounterKind: counter.Charge,
					Amount: game.Fixed(tt.requested), ChooseKind: tt.choose,
				}, PublishResult: "actual"},
				{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
						Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "actual", Multiplier: 1,
					}),
				}, ResultGate: opt.Val(game.InstructionResultGate{Key: "actual", AmountAvailable: true})},
			}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			got, available := obj.ResolvedAmounts["actual"]
			if !available || got != tt.want || source.Counters.Get(counter.Charge) != tt.have-tt.want ||
				g.Players[game.Player1].Life != 40+tt.want ||
				obj.ResolutionResults["actual"].Succeeded != (tt.want > 0) {
				t.Fatalf("amount=%d available=%v counters=%d life=%d", got, available, source.Counters.Get(counter.Charge), g.Players[game.Player1].Life)
			}
		})
	}
}

func TestCompiledCounterRemovalSuccessIsNotScalarAvailability(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Removal Success", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Remove a charge counter from target artifact. If you do, you gain 2 life.",
	})
	for _, have := range []int{0, 1} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatPermanent(g, game.Player1, def)
		target := addCombatPermanent(g, game.Player2, evidenceCard("Target", 1))
		target.Counters.Add(counter.Charge, have)
		obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
		NewEngine(nil).resolveInstructionSequence(g, obj, def.ActivatedAbilities[0].Content.Modes[0].Sequence,
			[game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if g.Players[game.Player1].Life != 40+2*have {
			t.Fatalf("have=%d life=%d, impossible removal must not satisfy If you do", have, g.Players[game.Player1].Life)
		}
	}
}

func TestCounterQuantityGroupAggregatesActual(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Aggregate", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Remove two +1/+1 counters from each creature you control. You gain life equal to the number of counters removed this way.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	for i, count := range []int{1, 3, 0} {
		target := addCombatPermanent(g, game.Player1, creatureDef("Counted"))
		target.Counters.Add(counter.PlusOnePlusOne, count)
		target.Counters.Add(counter.Charge, i+1)
	}
	decoy := addCombatPermanent(g, game.Player2, creatureDef("Other"))
	decoy.Counters.Add(counter.PlusOnePlusOne, 9)
	obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1}
	NewEngine(nil).resolveInstructionSequence(g, obj, def.ActivatedAbilities[0].Content.Modes[0].Sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if obj.ResolvedAmounts["removed-counter-quantity-1"] != 3 || g.Players[game.Player1].Life != 43 ||
		decoy.Counters.Get(counter.PlusOnePlusOne) != 9 {
		t.Fatalf("amounts=%v life=%d decoy=%d", obj.ResolvedAmounts, g.Players[game.Player1].Life, decoy.Counters.Get(counter.PlusOnePlusOne))
	}
}

func TestCounterQuantitySkippedDeclinedAndUnavailable(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"condition", "card condition", "result gate", "declined", "missing source", "changed incarnation"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, evidenceCard("Source", 1))
			source.Counters.Add(counter.Charge, 4)
			obj := &game.StackObject{
				SourceID: source.ObjectID, Controller: game.Player1,
				ResolvedAmounts:   map[string]int{"actual": 9, "unrelated": 17},
				ResolutionResults: map[string]game.InstructionResolutionResult{"actual": {Accepted: true, Succeeded: true, Amount: 9}},
			}
			publisher := game.Instruction{Primitive: game.RemoveCounter{
				Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(2),
			}, PublishResult: "actual"}
			switch mode {
			case "condition":
				publisher.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true})})
				addCardToHand(g, game.Player1, evidenceCard("Hand", 1))
			case "card condition":
				publisher.CardCondition = opt.Val(game.CardSelection{
					Card: game.CardReference{Kind: game.CardReferenceLinked, LinkID: "missing"},
				})
			case "result gate":
				publisher.ResultGate = opt.Val(game.InstructionResultGate{Key: "absent", Succeeded: game.TriTrue})
			case "declined":
				publisher.Optional = true
			case "missing source", "changed incarnation":
				if !movePermanentToZone(g, source, zone.Exile) {
					t.Fatal("source did not leave")
				}
				if mode == "changed incarnation" {
					card, _ := g.GetCardInstance(source.CardInstanceID)
					returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
					if !ok {
						t.Fatal("source did not return")
					}
					returned.Counters.Add(counter.Charge, 7)
				}
			}
			consumer := game.Instruction{
				Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				ResultGate: opt.Val(game.InstructionResultGate{Key: "actual", AmountAvailable: true}),
			}
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			before := g.Players[game.Player1].Hand.Size()
			engine.resolveInstructionSequence(g, obj, []game.Instruction{publisher, consumer},
				[game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{}}, &TurnLog{})
			if _, available := obj.ResolvedAmounts["actual"]; available ||
				g.Players[game.Player1].Hand.Size() != before || obj.ResolvedAmounts["unrelated"] != 17 {
				t.Fatalf("stale/unavailable quantity reused: amounts=%v hand=%d", obj.ResolvedAmounts, g.Players[game.Player1].Hand.Size())
			}
		})
	}
}

func TestCompiledCounterQuantityIndependentAndRepeated(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Independent", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Remove a charge counter from this artifact. You gain that much life. Remove two charge counters from this artifact. Draw that many cards.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Charge, 2)
	for range 5 {
		addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
	}
	obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1}
	sequence := def.ActivatedAbilities[0].Content.Modes[0].Sequence
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if obj.ResolvedAmounts["removed-counter-quantity-1"] != 1 ||
		obj.ResolvedAmounts["removed-counter-quantity-3"] != 1 ||
		g.Players[game.Player1].Life != 41 || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("independent amounts=%v", obj.ResolvedAmounts)
	}
	cloned := g.Clone()
	engine.resolveInstructionSequence(cloned, game.NewStackObjectCopy(obj, cloned.IDGen.Next()), sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if obj.ResolvedAmounts["removed-counter-quantity-1"] != 0 ||
		obj.ResolvedAmounts["removed-counter-quantity-3"] != 0 ||
		g.Players[game.Player1].Life != 41 || g.Players[game.Player1].Hand.Size() != 1 ||
		cloned.Players[game.Player1].Life != 41 || cloned.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("repeated or cloned resolution reused old quantities")
	}
}

func TestCompiledCounterQuantityNonzeroTargetSlot(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Target Quantity", Layout: "normal", TypeLine: "Instant",
		OracleText: "Tap target creature. Remove all +1/+1 counters from target creature. You gain that much life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	decoy := addCombatPermanent(g, game.Player1, creatureDef("Decoy"))
	target := addCombatPermanent(g, game.Player2, creatureDef("Actual"))
	decoy.Counters.Add(counter.PlusOnePlusOne, 9)
	target.Counters.Add(counter.PlusOnePlusOne, 2)
	target.Counters.Add(counter.Loyalty, 3)
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{
		game.PermanentTarget(decoy.ObjectID), game.PermanentTarget(target.ObjectID),
	}}
	NewEngine(nil).resolveInstructionSequence(g, obj, def.SpellAbility.Val.Modes[0].Sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != 42 || target.Counters.Get(counter.PlusOnePlusOne) != 0 ||
		target.Counters.Get(counter.Loyalty) != 3 || decoy.Counters.Get(counter.PlusOnePlusOne) != 9 {
		t.Fatal("removal quantity selected the wrong target or counter kind")
	}
}

func TestCounterQuantityZeroIsAvailable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Empty", Types: []types.Card{types.Artifact}}})
	obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1}
	addCardToLibrary(g, game.Player1, evidenceCard("Proof", 1))
	NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{
		{Primitive: game.RemoveCounter{Object: game.SourcePermanentReference(), Amount: game.Fixed(0), CounterKind: counter.Charge}, PublishResult: "actual"},
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}, ResultGate: opt.Val(game.InstructionResultGate{Key: "actual", AmountAvailable: true})},
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 || obj.ResolutionResults["actual"].Succeeded {
		t.Fatal("observed zero was confused with unavailable or effect success")
	}
}
