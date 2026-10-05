package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCompiledNumericCurrentAndDepartedSubjects(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text                  string
		power, toughness, manaValue int
		want                        bool
	}{
		{"current power below", "Put a +1/+1 counter on target creature. If its power is 4 or greater, you gain 2 life.", 2, 6, 8, false},
		{"current power boundary", "Put a +1/+1 counter on target creature. If its power is 4 or greater, you gain 2 life.", 3, 6, 8, true},
		{"current toughness below", "Put a +1/+1 counter on target creature. Then if that creature has toughness 6 or greater, you gain 2 life.", 8, 4, 8, false},
		{"current toughness boundary", "Put a +1/+1 counter on target creature. Then if that creature has toughness 6 or greater, you gain 2 life.", 8, 5, 8, true},
		{"departed power snapshot", "Put a +1/+1 counter on target creature. Destroy that creature. You gain 1 life. If its power was 4 or greater, you gain 2 life.", 3, 6, 8, true},
		{"departed power near miss", "Put a +1/+1 counter on target creature. Destroy that creature. You gain 1 life. If its power was 4 or greater, you gain 2 life.", 2, 6, 8, false},
		{"departed mana boundary", "Return target creature to its owner's hand. If its mana value was 3 or less, you gain 2 life.", 8, 8, 3, true},
		{"departed mana above", "Destroy target creature. If it had mana value 3 or less, you gain 2 life.", 8, 8, 4, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Numeric Resolution", Layout: "normal",
				TypeLine: "Instant", OracleText: tt.text})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCreatureWithPowerToughness(g, game.Player2, tt.power, tt.toughness)
			g.CardInstances[target.CardInstanceID].Def.ManaCost = opt.Val(cost.Mana{cost.O(tt.manaValue)})
			target.Owner = game.Player3
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := before
			if tt.name == "departed power snapshot" || tt.name == "departed power near miss" {
				want++
			}
			if tt.want {
				want += 2
			}
			if got := g.Players[game.Player1].Life; got != want {
				t.Fatalf("life=%d, want %d", got, want)
			}
			if g.Players[game.Player2].Life != 40 || g.Players[game.Player3].Life != 40 {
				t.Fatal("numeric payoff leaked to subject owner/controller")
			}
		})
	}
}

func TestCompiledNumericTargetOccurrenceAndSourceIsolation(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Numeric Isolation", Layout: "normal", TypeLine: "Creature",
		Power: new("9"), Toughness: new("9"),
		OracleText: "Whenever another creature enters, tap target creature. Put a +1/+1 counter on another target creature. You gain 1 life. If its toughness is 6 or greater, you gain 2 life."})
	for _, toughness := range []int{4, 5, 6} {
		t.Run(fmt.Sprint(toughness), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			event := addCreatureWithPowerToughness(g, game.Player1, 9, 9)
			decoy := addCreatureWithPowerToughness(g, game.Player2, 9, 9)
			target := addCreatureWithPowerToughness(g, game.Player2, 1, toughness)
			obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID, HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: event.ObjectID},
				Targets:      []game.Target{game.PermanentTarget(decoy.ObjectID), game.PermanentTarget(target.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 41
			if toughness >= 5 {
				want += 2
			}
			if g.Players[game.Player1].Life != want {
				t.Fatalf("life=%d, want %d; qualifying event/source/decoy must not replace slot 1", g.Players[game.Player1].Life, want)
			}
		})
	}
}

func TestCompiledNumericGroupAndIndependentTiming(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, tail string
		want       int
	}{
		{"group", "If its power is 2 or less, put a +1/+1 counter on that creature, then you gain 2 life.", 42},
		{"independent", "If its power is 2 or less, put a +1/+1 counter on that creature. If its power is 2 or less, you gain 2 life.", 40},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Numeric Timing", Layout: "normal", TypeLine: "Instant",
				OracleText: "Tap target creature. " + tt.tail})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCreatureWithPowerToughness(g, game.Player2, 2, 5)
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if g.Players[game.Player1].Life != tt.want || target.Counters.Get(counter.PlusOnePlusOne) != 1 {
				t.Fatalf("life=%d counters=%d, want life=%d counters=1", g.Players[game.Player1].Life, target.Counters.Get(counter.PlusOnePlusOne), tt.want)
			}
		})
	}
}

func TestCompiledCounteredSpellNumericInformation(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Past Spell", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell. You gain 1 life. If that spell's mana value was 3 or less, you gain 2 life."})
	for _, value := range []int{0, 2, 3, 4} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(value)})
			addInstructionSpellToStackForController(g, game.Player1, def.SpellAbility.Val.Modes[0].Sequence,
				[]game.Target{game.StackObjectTarget(target.ID)})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			want := 41
			if value <= 3 {
				want += 2
			}
			if g.Players[game.Player1].Life != want {
				t.Fatalf("life=%d, want %d", g.Players[game.Player1].Life, want)
			}
			if _, live := stackObjectByID(g, target.ID); live {
				t.Fatal("target spell survived")
			}
		})
	}
}

func TestNumericConditionUnknownFailsClosedUnderNegation(t *testing.T) {
	t.Parallel()
	for _, ref := range []game.ObjectReference{game.TargetStackObjectReference(0), game.EventStackObjectReference(), game.TargetPermanentReference(0)} {
		for _, negate := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := &game.StackObject{Controller: game.Player1, HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventSpellCast, StackObjectID: 99},
				Targets:      []game.Target{game.StackObjectTarget(99)}}
			condition := opt.Val(game.Condition{Negate: negate, Object: opt.Val(ref),
				ObjectMatches: opt.Val(game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 0})})})
			if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
				t.Fatalf("unknown object %v matched numeric zero with negate=%t", ref.Kind(), negate)
			}
		}
	}
}

func TestCompiledSplitskinNumericUnlessIsolation(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Splitskin Doll", Layout: "normal", TypeLine: "Artifact Creature",
		OracleText: "When this creature enters, draw a card. Then discard a card unless you control another creature with power 2 or less.",
		Power:      new("2"), Toughness: new("1")})
	for _, tt := range []struct {
		name                            string
		power                           int
		controller                      game.PlayerID
		sameName, token, blink, mutated bool
		want                            int
	}{
		{name: "source only", power: -1, want: 0},
		{name: "below boundary", power: 1, want: 1},
		{name: "at boundary", power: 2, want: 1},
		{name: "above boundary", power: 3, want: 0},
		{name: "opponent decoy", power: 2, controller: game.Player2, want: 0},
		{name: "same name distinct object", power: 2, sameName: true, want: 1},
		{name: "token candidate", power: 2, token: true, want: 1},
		{name: "blink source is distinct", power: -1, blink: true, want: 1},
		{name: "current power disqualifies printed two", power: 2, mutated: true, want: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID, HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: source.ObjectID}}
			if tt.power >= 0 {
				controller := tt.controller
				if controller == 0 {
					controller = game.Player1
				}
				candidateDef := resolvingSourceCreature("Candidate", tt.power)
				if tt.sameName {
					candidateDef = def
				}
				candidate := addCombatPermanent(g, controller, candidateDef)
				if tt.token {
					candidate.Token = true
					candidate.TokenDef = candidateDef
				}
				if tt.mutated {
					candidate.Counters.Add(counter.PlusOnePlusOne, 1)
				}
			}
			if tt.blink {
				for _, instruction := range selfBlinkInstructions() {
					NewEngine(nil).resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				}
			}
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[game.Player1].Hand.Size(); got != tt.want {
				t.Fatalf("hand=%d, want %d", got, tt.want)
			}
			if g.Players[game.Player2].Hand.Size() != 0 {
				t.Fatal("discard leaked controllers")
			}
		})
	}
}

func TestExpandedTargetCardTypeSelectionZoneVersion(t *testing.T) {
	t.Parallel()
	for _, cardType := range []types.Card{types.Artifact, types.Creature, types.Instant} {
		for _, stale := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card", Types: []types.Card{cardType}}})
			target := currentCardTarget(t, g, cardID)
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{target}}
			resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
			resolver.resolveInstruction(&game.Instruction{PublishResult: "card-move", Primitive: game.MoveCard{Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile}})
			if stale {
				g.CardInstances[cardID].ZoneVersion++
			}
			selection := game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}
			condition := opt.Val(game.Condition{TargetCardResultKey: "card-move", Object: opt.Val(game.TargetCardReference(0)), ObjectMatches: opt.Val(selection)})
			want := !stale && cardType != types.Instant
			if got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition); got != want {
				t.Fatalf("type=%s stale=%t: got %t, want %t", cardType, stale, got, want)
			}
		}
	}
}
