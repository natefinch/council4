package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestGenerateBloodchiefAscension proves Bloodchief Ascension lowers both of its
// abilities through reusable infrastructure. Ability 1 is an each-end-step
// (any player's end step) intervening-if gated on any single opponent having
// lost 2 or more life this turn via the AggregateAnyOpponentLifeLostThisTurn
// aggregate, whose body optionally adds a quest counter. Ability 2 is a
// card-into-an-opponent's-graveyard-from-anywhere trigger; because a token is
// not a card, its subject carries a non-token requirement. Its intervening-if
// requires three or more quest counters on the source, then optionally drains
// that opponent for 2 life and gains 2 life only if the drain happened.
func TestGenerateBloodchiefAscension(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Bloodchief Ascension",
		Layout:   "normal",
		TypeLine: "Enchantment",
		ManaCost: "{B}",
		OracleText: "At the beginning of each end step, if an opponent lost 2 or more life this turn, you may put a quest counter on Bloodchief Ascension. (Damage causes loss of life.)\n" +
			"Whenever a card is put into an opponent's graveyard from anywhere, if Bloodchief Ascension has three or more quest counters on it, you may have that player lose 2 life. If you do, you gain 2 life.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}

	// Ability 1: each end step, opponent-lost-2+-life intervening-if.
	quest := face.TriggeredAbilities[0]
	if quest.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		quest.Trigger.Pattern.Step != game.StepEnd {
		t.Fatalf("ability 1 pattern = %#v, want beginning of end step", quest.Trigger.Pattern)
	}
	if quest.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("ability 1 controller = %v, want TriggerControllerAny (each end step)", quest.Trigger.Pattern.Controller)
	}
	if !quest.Trigger.InterveningCondition.Exists {
		t.Fatal("ability 1 has no intervening condition")
	}
	if got := quest.Trigger.InterveningCondition.Val.Aggregates; len(got) != 1 ||
		got[0].Aggregate != game.AggregateAnyOpponentLifeLostThisTurn ||
		got[0].Op != compare.GreaterOrEqual ||
		got[0].Value != 2 {
		t.Fatalf("ability 1 aggregates = %#v, want any-opponent-life-lost-this-turn >= 2", got)
	}
	questSeq := quest.Content.Modes[0].Sequence
	if len(questSeq) != 1 {
		t.Fatalf("ability 1 sequence len = %d, want 1", len(questSeq))
	}
	addCounter, ok := questSeq[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("ability 1 sequence[0] = %T, want game.AddCounter", questSeq[0].Primitive)
	}
	if addCounter.CounterKind != counter.Quest {
		t.Fatalf("ability 1 counter kind = %v, want Quest", addCounter.CounterKind)
	}
	if !questSeq[0].Optional {
		t.Fatal("ability 1 add-counter is not optional (\"you may put\")")
	}

	// Ability 2: card into an opponent's graveyard from anywhere, non-token.
	drain := face.TriggeredAbilities[1]
	if drain.Trigger.Pattern.Event != game.EventZoneChanged ||
		drain.Trigger.Pattern.Player != game.TriggerPlayerOpponent ||
		!drain.Trigger.Pattern.MatchToZone ||
		drain.Trigger.Pattern.ToZone != zone.Graveyard {
		t.Fatalf("ability 2 pattern = %#v, want card into opponent's graveyard", drain.Trigger.Pattern)
	}
	if drain.Trigger.Pattern.MatchFromZone {
		t.Fatalf("ability 2 restricts the origin zone, want from anywhere: %#v", drain.Trigger.Pattern)
	}
	if !drain.Trigger.Pattern.SubjectSelection.NonToken {
		t.Fatalf("ability 2 subject = %#v, want a non-token requirement (a token is not a card)", drain.Trigger.Pattern.SubjectSelection)
	}
	if !drain.Trigger.InterveningCondition.Exists {
		t.Fatal("ability 2 has no intervening condition")
	}
	matches := drain.Trigger.InterveningCondition.Val.ObjectMatches
	if !matches.Exists ||
		matches.Val.RequiredCounter != counter.Quest ||
		!matches.Val.RequiredCounterCount.Exists ||
		matches.Val.RequiredCounterCount.Val.Op != compare.GreaterOrEqual ||
		matches.Val.RequiredCounterCount.Val.Value != 3 {
		t.Fatalf("ability 2 intervening = %#v, want source with >= 3 quest counters", drain.Trigger.InterveningCondition.Val)
	}
	drainSeq := drain.Content.Modes[0].Sequence
	if len(drainSeq) != 2 {
		t.Fatalf("ability 2 sequence len = %d, want 2", len(drainSeq))
	}
	loseLife, ok := drainSeq[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("ability 2 sequence[0] = %T, want game.LoseLife", drainSeq[0].Primitive)
	}
	if loseLife.Player != game.EventPlayerReference() {
		t.Fatalf("ability 2 lose-life recipient = %#v, want EventPlayerReference (that player)", loseLife.Player)
	}
	if !drainSeq[0].Optional {
		t.Fatal("ability 2 lose-life is not optional (\"you may have\")")
	}
	if drainSeq[0].PublishResult == "" {
		t.Fatal("ability 2 lose-life does not publish its result for the \"if you do\" gate")
	}
	gainLife, ok := drainSeq[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("ability 2 sequence[1] = %T, want game.GainLife", drainSeq[1].Primitive)
	}
	if gainLife.Player != game.ControllerReference() {
		t.Fatalf("ability 2 gain-life recipient = %#v, want ControllerReference (you)", gainLife.Player)
	}
	gate := drainSeq[1].ResultGate
	if !gate.Exists ||
		gate.Val.Key != drainSeq[0].PublishResult ||
		gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("ability 2 gain-life gate = %#v, want gated on the drain result succeeding", gate)
	}
}
