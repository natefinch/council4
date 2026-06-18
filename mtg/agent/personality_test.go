package agent

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

// TestAggressionAttacksIntoBlocks checks that aggression makes the agent value an
// attack it would otherwise decline (the defender can block it profitably).
func TestAggressionAttacksIntoBlocks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addObservedPermanent(g, game.Player1, creatureCardDef("Bear", 2, 2))
	addObservedPermanent(g, game.Player2, creatureCardDef("Blocker", 3, 3)) // blocks and survives
	obs := rules.NewObservation(g, game.Player1)
	attack := attackAction(game.Player2, attacker.ObjectID)

	neutral := GenericStrategy{}.ScoreAction(obs, attack)
	aggressive := GenericStrategy{Personality: Personality{Aggression: 3}}.ScoreAction(obs, attack)

	if neutral >= scorePass {
		t.Errorf("neutral attack into a profitable block (%v) should score below passing (%v)", neutral, scorePass)
	}
	if aggressive <= scorePass {
		t.Errorf("aggressive attack (%v) should beat passing (%v)", aggressive, scorePass)
	}
	if aggressive <= neutral {
		t.Errorf("aggression should raise the attack score: aggressive=%v neutral=%v", aggressive, neutral)
	}
}

// TestRiskToleranceReducesHoldUp checks that a risk-tolerant agent taps out for a
// threat the neutral agent would hold mana back from.
func TestRiskToleranceReducesHoldUp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, instantDef("Counterspell", 2, color.Blue))
	creatureID := addObservedHandCard(g, game.Player1, creatureWithCost("Big Beast", 4, 4, 4))
	for range 5 {
		addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	}
	obs := rules.NewObservation(g, game.Player1)
	cast := action.CastSpell(creatureID, nil, 0, nil)

	neutral := GenericStrategy{}.ScoreAction(obs, cast)
	risky := GenericStrategy{Personality: Personality{RiskTolerance: 2}}.ScoreAction(obs, cast)

	if risky <= neutral {
		t.Errorf("risk tolerance should reduce the hold-up penalty: risky=%v neutral=%v", risky, neutral)
	}
}

// TestPoliticsWeightFocusesThreats checks that a higher politics weight increases
// the value of aiming a spell at a dangerous opponent.
func TestPoliticsWeightFocusesThreats(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	burnID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Lava Spike",
		Types: []types.Card{types.Instant},
	}})
	addObservedPermanent(g, game.Player2, creatureCardDef("Threat", 5, 5)) // gives Player2 board threat
	obs := rules.NewObservation(g, game.Player1)
	burn := action.CastSpell(burnID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)

	neutral := GenericStrategy{}.ScoreAction(obs, burn)
	political := GenericStrategy{Personality: Personality{PoliticsWeight: 2}}.ScoreAction(obs, burn)

	if political <= neutral {
		t.Errorf("politics weight should raise the value of hitting a threatening player: political=%v neutral=%v", political, neutral)
	}
}

// TestNoiseIsDeterministicPerSeed checks that two strategies with identically
// seeded noise sources produce identical scores, and that noise actually
// perturbs the neutral score.
func TestNoiseIsDeterministicPerSeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureID := addObservedHandCard(g, game.Player1, creatureCardDef("Bear", 2, 2))
	obs := rules.NewObservation(g, game.Player1)
	actions := []action.Action{
		action.Pass(),
		action.CastSpell(creatureID, nil, 0, nil),
		action.PlayLandFace(g.IDGen.Next(), game.FaceFront),
	}

	personality := Personality{NoiseMagnitude: 5}
	first := GenericStrategy{Personality: personality.WithNoiseSource(seededRNG())}
	second := GenericStrategy{Personality: personality.WithNoiseSource(seededRNG())}
	neutral := GenericStrategy{}

	perturbed := false
	for _, act := range actions {
		a := first.ScoreAction(obs, act)
		b := second.ScoreAction(obs, act)
		if a != b {
			t.Fatalf("noise not deterministic for a fixed seed: %v vs %v", a, b)
		}
		if a != neutral.ScoreAction(obs, act) {
			perturbed = true
		}
	}
	if !perturbed {
		t.Error("noise did not perturb any score; expected jitter around the neutral scores")
	}
}

// TestZeroPersonalityIsNeutral checks that the zero Personality reproduces the
// plain generic strategy exactly.
func TestZeroPersonalityIsNeutral(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureID := addObservedHandCard(g, game.Player1, creatureCardDef("Bear", 2, 2))
	obs := rules.NewObservation(g, game.Player1)
	cast := action.CastSpell(creatureID, nil, 0, nil)

	plain := GenericStrategy{}.ScoreAction(obs, cast)
	zero := GenericStrategy{Personality: Personality{}}.ScoreAction(obs, cast)
	if plain != zero {
		t.Errorf("zero personality (%v) should equal the plain strategy (%v)", zero, plain)
	}
}

func seededRNG() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}
