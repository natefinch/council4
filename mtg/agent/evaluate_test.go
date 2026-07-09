package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

func evalOf(g *game.Game, player game.PlayerID) float64 {
	return Evaluate(rules.NewObservation(g, player))
}

func TestEvaluateSymmetricStartIsNeutral(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if got := evalOf(g, game.Player1); got != 0 {
		t.Fatalf("symmetric empty start scored %v, want 0", got)
	}
}

func TestEvaluateRewardsOwnBoardAndMana(t *testing.T) {
	base := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	baseline := evalOf(base, game.Player1)

	withCreature := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(withCreature, game.Player1, creatureCardDef("Bear", 3, 3))
	if got := evalOf(withCreature, game.Player1); got <= baseline {
		t.Fatalf("adding a 3/3 scored %v, want above baseline %v", got, baseline)
	}

	withRock := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(withRock, game.Player1, manaRockDef("Signet", 2))
	if got := evalOf(withRock, game.Player1); got <= baseline {
		t.Fatalf("adding a mana rock scored %v, want above baseline %v (mana development)", got, baseline)
	}
}

func TestEvaluatePenalizesOpponentBoard(t *testing.T) {
	base := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	baseline := evalOf(base, game.Player1)

	oppThreat := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(oppThreat, game.Player2, creatureCardDef("Wurm", 8, 8))
	if got := evalOf(oppThreat, game.Player1); got >= baseline {
		t.Fatalf("an opponent's 8/8 scored %v for me, want below baseline %v", got, baseline)
	}
}

func TestEvaluateValuesRemovingTheLeadersThreat(t *testing.T) {
	// Removing the strongest opponent's threat should raise my evaluation, so a
	// search that leads to that state prefers it.
	before := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(before, game.Player2, creatureCardDef("Wurm", 8, 8))
	after := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	if evalOf(after, game.Player1) <= evalOf(before, game.Player1) {
		t.Fatalf("removing the leader's 8/8 did not improve the evaluation: before=%v after=%v",
			evalOf(before, game.Player1), evalOf(after, game.Player1))
	}
}

func TestEvaluateRewardsCardAdvantage(t *testing.T) {
	base := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	baseline := evalOf(base, game.Player1)

	withCards := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(withCards, game.Player1, creatureCardDef("A", 1, 1))
	addObservedHandCard(withCards, game.Player1, creatureCardDef("B", 1, 1))
	if got := evalOf(withCards, game.Player1); got <= baseline {
		t.Fatalf("two extra cards scored %v, want above baseline %v", got, baseline)
	}
}

func TestEvaluateRewardsLife(t *testing.T) {
	low := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	low.Players[game.Player1].Life = 10
	high := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	high.Players[game.Player1].Life = 40

	if evalOf(high, game.Player1) <= evalOf(low, game.Player1) {
		t.Fatalf("more life did not raise the evaluation: life10=%v life40=%v",
			evalOf(low, game.Player1), evalOf(high, game.Player1))
	}
}

func TestEvaluateDeployingANoncreatureIsNotALoss(t *testing.T) {
	// A noncreature permanent — a ramp aura, an anthem, an artifact engine — is
	// the realized form of the card that made it, so casting it must not LOWER the
	// searcher's own evaluation. When a noncreature on the board was valued far
	// below the card it came from, one-ply search saw deploying any enchantment or
	// artifact as a net loss and refused to develop them, only creatures — measured
	// to make the search agent decline ramp auras and engines it should be casting.
	// So a noncreature on the battlefield must be worth at least the same card held.
	def := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:  "Wild Growth",
			Types: []types.Card{types.Enchantment},
		},
	}

	inHand := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(inHand, game.Player1, def)

	onBoard := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(onBoard, game.Player1, def)

	if evalOf(onBoard, game.Player1) < evalOf(inHand, game.Player1) {
		t.Fatalf("deploying a noncreature enchantment lowered the evaluation: "+
			"onBoard=%v inHand=%v; casting a noncreature must not look like a loss",
			evalOf(onBoard, game.Player1), evalOf(inHand, game.Player1))
	}
}

func TestEvaluateLifeIsConcave(t *testing.T) {
	// A point of life is worth more when a player is low than when they are high:
	// gaining life from 5 to 6 must raise the evaluation more than gaining it from
	// 39 to 40. A strong player spends life freely while healthy and fights for
	// every point when low, so the curve must be concave, not linear.
	eval := func(life int) float64 {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player1].Life = life
		return evalOf(g, game.Player1)
	}

	marginalWhenLow := eval(6) - eval(5)
	marginalWhenHigh := eval(40) - eval(39)
	if marginalWhenLow <= marginalWhenHigh {
		t.Fatalf("life is not concave: a point at low life (%v) should be worth more "+
			"than a point at high life (%v)", marginalWhenLow, marginalWhenHigh)
	}
}

func TestEvaluateRewardsRacingAKillableOpponent(t *testing.T) {
	// A position where an opponent is near death should score higher for me than
	// one where every opponent is healthy, so search commits to finishing a kill
	// instead of durdling to a draw. The wounded opponent is NOT the strongest
	// (Player3 has the biggest board), so this isolates the closing reward from
	// the "my power minus the strongest opponent" core: only being close to a kill
	// can raise the score here.
	healthy := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(healthy, game.Player3, creatureCardDef("Wurm", 8, 8))

	racing := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(racing, game.Player3, creatureCardDef("Wurm", 8, 8))
	racing.Players[game.Player2].Life = 3

	if evalOf(racing, game.Player1) <= evalOf(healthy, game.Player1) {
		t.Fatalf("an opponent at 3 life did not improve my evaluation: "+
			"racing=%v healthy=%v; the agent should value closing on a killable player",
			evalOf(racing, game.Player1), evalOf(healthy, game.Player1))
	}
}

func TestEvaluateRacingCountsPoisonAndCommanderDamage(t *testing.T) {
	// Life is not the only clock: an opponent one hit from a poison or
	// commander-damage loss is just as killable, and a strong player finishes them
	// on whichever clock is furthest along. Both must register as closing progress.
	baseline := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(baseline, game.Player3, creatureCardDef("Wurm", 8, 8))
	base := evalOf(baseline, game.Player1)

	poisoned := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(poisoned, game.Player3, creatureCardDef("Wurm", 8, 8))
	poisoned.Players[game.Player2].PoisonCounters = 9
	if evalOf(poisoned, game.Player1) <= base {
		t.Fatalf("an opponent at 9 poison did not improve my evaluation: poisoned=%v base=%v",
			evalOf(poisoned, game.Player1), base)
	}

	commanderDamaged := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(commanderDamaged, game.Player3, creatureCardDef("Wurm", 8, 8))
	commanderDamaged.Players[game.Player2].CommanderDamage = map[id.ID]int{commanderDamaged.IDGen.Next(): 20}
	if evalOf(commanderDamaged, game.Player1) <= base {
		t.Fatalf("an opponent at 20 commander damage did not improve my evaluation: cmdr=%v base=%v",
			evalOf(commanderDamaged, game.Player1), base)
	}
}

func TestEvaluateDeployingACreatureBeatsHoardingIt(t *testing.T) {
	// A creature is worth more on the battlefield than in hand: it can attack,
	// block, tap, and pressure the table, where a held card only threatens to. If a
	// card in hand were valued as high as the creature it becomes, one-ply search
	// would hoard its hand and never develop a board — measured to make the search
	// agent lose every game to GenericStrategy. So deploying must raise the score.
	inHand := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(inHand, game.Player1, creatureCardDef("Bear", 2, 2))

	onBoard := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(onBoard, game.Player1, creatureCardDef("Bear", 2, 2))

	if evalOf(onBoard, game.Player1) <= evalOf(inHand, game.Player1) {
		t.Fatalf("a 2/2 on the battlefield (%v) did not beat the same 2/2 in hand (%v); "+
			"a card in hand must be worth less than the creature it becomes",
			evalOf(onBoard, game.Player1), evalOf(inHand, game.Player1))
	}
}

func TestEvaluateWinWhenOpponentsEliminated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, seat := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		g.Players[seat].Eliminated = true
	}
	if got := evalOf(g, game.Player1); got != evalWin {
		t.Fatalf("last player standing scored %v, want evalWin %v", got, evalWin)
	}
}

func TestEvaluateLossWhenEliminated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Eliminated = true
	if got := evalOf(g, game.Player1); got != evalLoss {
		t.Fatalf("eliminated player scored %v, want evalLoss %v", got, evalLoss)
	}
}

func TestEvaluateRewardsHeldUpInteraction(t *testing.T) {
	// Ending a turn able to answer a threat — an instant in hand with untapped
	// mana to cast it — is worth more than tapping out, so the searcher plays
	// around interaction instead of dumping its hand. The value is realized on an
	// opponent's turn, which one-ply search never simulates, so the eval rewards
	// the held-up option directly.
	instant := &game.CardDef{CardFace: game.CardFace{
		Name:  "Removal",
		Types: []types.Card{types.Instant},
	}}

	tappedOut := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(tappedOut, game.Player1, instant)
	spent := addObservedPermanent(tappedOut, game.Player1, manaRockDef("Signet", 2))
	spent.Tapped = true

	heldUp := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(heldUp, game.Player1, instant)
	addObservedPermanent(heldUp, game.Player1, manaRockDef("Signet", 2)) // untapped

	if evalOf(heldUp, game.Player1) <= evalOf(tappedOut, game.Player1) {
		t.Fatalf("holding up an instant with untapped mana (%v) should beat being tapped out (%v)",
			evalOf(heldUp, game.Player1), evalOf(tappedOut, game.Player1))
	}
}

func TestEvaluateHoldUpNeedsAnAnswerInHand(t *testing.T) {
	// Untapped mana alone is not reactive potential: without an instant to cast,
	// keeping mana up answers nothing, so it earns no hold-up bonus (that would
	// just make the agent decline to develop for no reason).
	withInstant := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(withInstant, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Removal", Types: []types.Card{types.Instant},
	}})
	addObservedPermanent(withInstant, game.Player1, manaRockDef("Signet", 2))

	withSorcery := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(withSorcery, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Ritual", Types: []types.Card{types.Sorcery},
	}})
	addObservedPermanent(withSorcery, game.Player1, manaRockDef("Signet", 2))

	// Same board and hand size; only the instant grants reactive potential.
	if evalOf(withInstant, game.Player1) <= evalOf(withSorcery, game.Player1) {
		t.Fatalf("an instant (%v) should give reactive potential a sorcery (%v) does not",
			evalOf(withInstant, game.Player1), evalOf(withSorcery, game.Player1))
	}
}
