package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerCavesOfChaosConditionalImpulse(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Caves of Chaos Adventurer",
		Layout:   "normal",
		TypeLine: "Creature — Human Barbarian",
		OracleText: "Trample\n" +
			"When this creature enters, you take the initiative.\n" +
			"Whenever this creature attacks, exile the top card of your library. " +
			"If you've completed a dungeon, you may play that card this turn without paying its mana cost. " +
			"Otherwise, you may play that card this turn.",
	})
	var attack *game.TriggeredAbility
	for i := range face.TriggeredAbilities {
		trigger := &face.TriggeredAbilities[i]
		if trigger.Trigger.Pattern.Event == game.EventAttackerDeclared {
			attack = trigger
			break
		}
	}
	if attack == nil {
		t.Fatal("attack trigger not lowered")
	}
	sequence := attack.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v", sequence)
	}
	for i := range sequence {
		impulse, ok := sequence[i].Primitive.(game.ImpulseExile)
		if !ok ||
			impulse.Player != game.ControllerReference() ||
			impulse.Amount.Value() != 1 ||
			impulse.Duration != game.DurationThisTurn ||
			impulse.WithoutPayingManaCost != (i == 0) {
			t.Fatalf("impulse instruction %d = %#v", i, sequence[i])
		}
		if !sequence[i].Condition.Exists || !sequence[i].Condition.Val.Condition.Exists {
			t.Fatalf("impulse instruction %d missing dungeon gate", i)
		}
		condition := sequence[i].Condition.Val.Condition.Val
		if !condition.ControllerCompletedADungeon || condition.Negate != (i == 1) {
			t.Fatalf("impulse instruction %d condition = %#v", i, condition)
		}
	}
}
