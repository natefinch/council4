package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledConditionScopeTiming(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text            string
		initialHand, wantHand int
		activated             bool
	}{
		{"group once", "If you have no cards in hand, draw a card, then draw a card.", 0, 2, false},
		{"independent clauses", "If you have no cards in hand, draw a card. If you have no cards in hand, draw a card.", 0, 1, false},
		{"false group", "If you have no cards in hand, draw a card, then draw a card.", 1, 1, false},
		{"after mutation", "Discard a card. If you have no cards in hand, draw a card, then draw a card.", 1, 2, false},
		{"different primitives", "If you have no cards in hand, draw a card, then you gain 2 life.", 0, 1, false},
		{"two distinct groups", "If you have no cards in hand, draw a card, then draw a card. If you have no cards in hand, draw a card, then draw a card.", 0, 2, false},
		{"activated group", "If you have no cards in hand, draw a card, then draw a card.", 0, 2, true},
		{"activated independent", "If you have no cards in hand, draw a card. If you have no cards in hand, draw a card.", 0, 1, true},
		{"activated false", "If you have no cards in hand, draw a card, then draw a card.", 1, 1, true},
		{"activated after mutation", "Discard a card. If you have no cards in hand, draw a card, then draw a card.", 1, 2, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := cardgen.ScryfallCard{
				Name: "Condition Scope", Layout: "normal", TypeLine: "Instant", OracleText: tt.text,
			}
			if tt.activated {
				card.TypeLine = "Enchantment"
				card.OracleText = "{1}: " + tt.text
			}
			def := compileUnlessCard(t, card)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for range tt.initialHand {
				addCardToHand(g, game.Player1, evidenceCard("Initial", 1))
			}
			for range 3 {
				addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			}
			life := g.Players[game.Player1].Life
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			content := def.SpellAbility.Val
			if tt.activated {
				ability := def.ActivatedAbilities[0]
				if ability.ActivationCondition.Exists {
					t.Fatal("resolving body condition became an activation restriction")
				}
				content = ability.Content
				obj.Kind = game.StackActivatedAbility
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[game.Player1].Hand.Size(); got != tt.wantHand {
				t.Fatalf("hand=%d, want %d", got, tt.wantHand)
			}
			if tt.name == "different primitives" && g.Players[game.Player1].Life != life+2 {
				t.Fatal("group's life gain re-evaluated the now-false empty-hand condition")
			}
		})
	}
}

func TestCompiledConditionBranchTiming(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text            string
		initialHand, wantHand int
		wantLife              int
	}{
		{"single otherwise", "If you have no cards in hand, draw a card. Otherwise, you gain 2 life.", 0, 1, 40},
		{"group otherwise true", "If you have no cards in hand, draw a card, then draw a card. Otherwise, you gain 2 life.", 0, 2, 40},
		{"group otherwise false", "If you have no cards in hand, draw a card, then draw a card. Otherwise, you gain 2 life.", 1, 1, 42},
		{"instead base mutates", "Discard a card. If you have no cards in hand, draw two cards instead.", 1, 0, 40},
		{"instead alternative", "Discard a card. If you have no cards in hand, draw two cards instead.", 0, 2, 40},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Condition Branch", Layout: "normal", TypeLine: "Instant", OracleText: tt.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for range tt.initialHand {
				addCardToHand(g, game.Player1, evidenceCard("Initial", 1))
			}
			for range 3 {
				addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if hand, life := g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Life; hand != tt.wantHand || life != tt.wantLife {
				t.Fatalf("hand=%d life=%d, want hand=%d life=%d", hand, life, tt.wantHand, tt.wantLife)
			}
		})
	}
}

func TestCompiledExpandedClauseConditionTiming(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Expanded Condition", Layout: "normal", TypeLine: "Instant",
		OracleText: "Tap target artifact. Unless you control a creature with power 2 or greater, put a +1/+1 counter on each of up to two target creatures you control.",
	})
	for _, tt := range []struct {
		name         string
		firstMissing bool
		largeOwn     bool
		largeOther   bool
		wantCounters int
	}{
		{name: "mutation after first target", wantCounters: 1},
		{name: "first target removed", firstMissing: true, wantCounters: 1},
		{name: "false condition", largeOwn: true},
		{name: "other controller isolated", largeOther: true, wantCounters: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Unrelated target", Types: []types.Card{types.Artifact},
			}})
			creature := &game.CardDef{CardFace: game.CardFace{
				Name: "Small creature", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			}}
			first := addCombatPermanent(g, game.Player1, creature)
			second := addCombatPermanent(g, game.Player1, creature)
			if tt.largeOwn || tt.largeOther {
				controller := game.Player1
				if tt.largeOther {
					controller = game.Player2
				}
				addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
					Name: "Large creature", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
				}})
			}
			if tt.firstMissing {
				removePermanentFromBattlefield(g, first.ObjectID)
			}
			obj := &game.StackObject{
				Kind: game.StackSpell, Controller: game.Player1,
				Targets: []game.Target{
					game.PermanentTarget(artifact.ObjectID),
					game.PermanentTarget(first.ObjectID),
					game.PermanentTarget(second.ObjectID),
				},
				TargetCounts: []int{1, 1, 1},
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !artifact.Tapped || artifact.Counters.Get(counter.PlusOnePlusOne) != 0 {
				t.Fatal("unrelated slot-zero target was not isolated")
			}
			if got := second.Counters.Get(counter.PlusOnePlusOne); got != tt.wantCounters {
				t.Fatalf("second target counters=%d, want %d", got, tt.wantCounters)
			}
			if !tt.firstMissing && first.Counters.Get(counter.PlusOnePlusOne) != tt.wantCounters {
				t.Fatal("first target did not receive the same clause decision")
			}
		})
	}
}

func TestCompiledGroupConditionDoesNotRequireFirstEffectSuccess(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Ineffective Group", Layout: "normal", TypeLine: "Instant",
		OracleText: "If you control a creature, destroy target creature, then draw a card.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Condition subject", Types: []types.Card{types.Creature},
	}})
	protected := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Indestructible target", Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{game.IndestructibleStaticBody},
	}})
	addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
		Targets: []game.Target{game.PermanentTarget(protected.ObjectID)},
	}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	_, survived := permanentByObjectID(g, protected.ObjectID)
	if g.Players[game.Player1].Hand.Size() != 1 || !survived {
		t.Fatal("a prevented destruction must not prevent the group's draw")
	}
}

func TestConditionEvaluationFalseUnavailableAndIsolation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name           string
		publish        bool
		falseCondition bool
		negateGate     bool
		wantLife       int
	}{
		{"true publication even if action skipped", true, false, false, 42},
		{"false publication", true, true, false, 40},
		{"unavailable publication", false, false, false, 40},
		{"false complemented publication", true, true, true, 42},
		{"true complemented publication", true, false, true, 40},
		{"missing complemented publication", false, false, true, 40},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			publisher := game.Instruction{
				Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					ControllerHandEmpty: true, Negate: tt.falseCondition,
				})}),
				PublishCondition: "scope",
				ResultGate:       opt.Val(game.InstructionResultGate{Key: "unavailable-result", Succeeded: game.TriTrue}),
			}
			consumer := game.Instruction{
				Primitive:           game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
				ConditionGate:       "scope",
				ConditionGateNegate: tt.negateGate,
			}
			sequence := []game.Instruction{consumer}
			if tt.publish {
				sequence = []game.Instruction{publisher, consumer}
			}
			engine := NewEngine(nil)
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if g.Players[game.Player1].Life != tt.wantLife {
				t.Fatalf("life=%d, want %d", g.Players[game.Player1].Life, tt.wantLife)
			}
			engine.resolveInstructionSequence(g, obj, []game.Instruction{consumer}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if g.Players[game.Player1].Life != tt.wantLife {
				t.Fatal("evaluation leaked across separate sequence resolutions")
			}
		})
	}
}
