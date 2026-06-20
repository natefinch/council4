package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerImpulseExileGeneralized(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		oracle   string
		amount   int
		duration game.EffectDuration
	}{
		{
			name:     "single card this turn",
			oracle:   "Exile the top card of your library. You may play that card this turn.",
			amount:   1,
			duration: game.DurationThisTurn,
		},
		{
			name:     "single card until end of turn",
			oracle:   "Exile the top card of your library. You may play it until end of turn.",
			amount:   1,
			duration: game.DurationUntilEndOfTurn,
		},
		{
			name:     "two cards until end of your next turn",
			oracle:   "Exile the top two cards of your library. You may play those cards until the end of your next turn.",
			amount:   2,
			duration: game.DurationUntilEndOfYourNextTurn,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Standalone Impulse",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				ManaCost:   "{2}{R}",
				OracleText: tc.oracle,
			})
			content := face.SpellAbility.Val
			if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
				t.Fatalf("content = %+v", content)
			}
			impulse, ok := content.Modes[0].Sequence[0].Primitive.(game.ImpulseExile)
			if !ok || impulse.Amount.Value() != tc.amount || impulse.Duration != tc.duration {
				t.Fatalf("primitive = %+v", content.Modes[0].Sequence[0].Primitive)
			}
		})
	}
}

func TestLowerImpulseExileActivatedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Treasure Breaker",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		ManaCost:   "{2}{R}",
		OracleText: "Sacrifice a Treasure: Exile the top card of your library. You may play that card this turn.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %+v", face.ActivatedAbilities)
	}
}
