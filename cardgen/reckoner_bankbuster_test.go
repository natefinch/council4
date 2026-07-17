package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
)

const reckonerBankbusterOracleText = "Reckoner Bankbuster enters the battlefield with three charge counters on it.\n" +
	"{2}, {T}, Remove a charge counter from Reckoner Bankbuster: Draw a card. Then if there are no charge counters on Reckoner Bankbuster, create a Treasure token and a 1/1 colorless Pilot creature token with \"This creature crews Vehicles as though its power were 2 greater.\"\n" +
	"Crew 3"

func bankbusterCreateToken(t *testing.T, instruction game.Instruction) game.CreateToken {
	t.Helper()
	create, ok := instruction.Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", instruction.Primitive)
	}
	return create
}

func TestLowerReckonerBankbuster(t *testing.T) {
	t.Parallel()
	power, toughness := "4", "4"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reckoner Bankbuster",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact — Vehicle",
		OracleText: reckonerBankbusterOracleText,
		Power:      &power,
		Toughness:  &toughness,
	})

	if len(face.ReplacementAbilities) != 1 ||
		!reflect.DeepEqual(face.ReplacementAbilities[0], game.EntersWithCountersReplacement(
			"Reckoner Bankbuster enters the battlefield with three charge counters on it.",
			game.CounterPlacement{Kind: counter.Charge, Amount: 3},
		)) {
		t.Fatalf("replacement abilities = %#v", face.ReplacementAbilities)
	}

	if len(face.ActivatedAbilities) != 2 {
		t.Fatalf("activated abilities = %d, want draw ability and Crew", len(face.ActivatedAbilities))
	}
	drawAbility := face.ActivatedAbilities[0]
	if !drawAbility.ManaCost.Exists ||
		!reflect.DeepEqual(drawAbility.ManaCost.Val, cost.Mana{cost.O(2)}) {
		t.Fatalf("mana cost = %#v, want {2}", drawAbility.ManaCost)
	}
	if len(drawAbility.AdditionalCosts) != 2 ||
		drawAbility.AdditionalCosts[0].Kind != cost.AdditionalTap ||
		drawAbility.AdditionalCosts[1].Kind != cost.AdditionalRemoveCounter ||
		drawAbility.AdditionalCosts[1].CounterKind != counter.Charge ||
		drawAbility.AdditionalCosts[1].Amount != 1 {
		t.Fatalf("additional costs = %#v", drawAbility.AdditionalCosts)
	}
	sequence := drawAbility.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want draw then two token creates", sequence)
	}
	if _, ok := sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("sequence[0] = %T, want game.Draw", sequence[0].Primitive)
	}
	for i := 1; i < 3; i++ {
		if !sequence[i].Condition.Exists {
			t.Fatalf("sequence[%d] has no zero-charge-counter gate", i)
		}
	}

	treasure := multiTokenDef(t, bankbusterCreateToken(t, sequence[1]))
	if treasure.Name != "Treasure" || len(treasure.StaticAbilities) != 0 {
		t.Fatalf("Treasure token = %#v", treasure)
	}
	pilot := multiTokenDef(t, bankbusterCreateToken(t, sequence[2]))
	if pilot.Name != "Pilot" ||
		len(pilot.StaticAbilities) != 1 ||
		pilot.StaticAbilities[0].CrewPowerBonus != 2 {
		t.Fatalf("Pilot token = %#v", pilot)
	}

	crew := face.ActivatedAbilities[1]
	if !reflect.DeepEqual(crew, game.CrewActivatedAbility(3)) {
		t.Fatalf("Crew ability = %#v, want Crew 3", crew)
	}
	if got := crew.AdditionalCosts[0].PowerContribution; got != cost.PowerContributionCrew {
		t.Fatalf("Crew contribution kind = %v, want crew", got)
	}
}

func TestLowerReckonerBankbusterCurrentOracle(t *testing.T) {
	t.Parallel()
	power, toughness := "4", "4"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Reckoner Bankbuster",
		Layout:   "normal",
		ManaCost: "{2}",
		TypeLine: "Artifact — Vehicle",
		OracleText: "This Vehicle enters with three charge counters on it.\n" +
			"{2}, {T}, Remove a charge counter from this Vehicle: Draw a card. Then if there are no charge counters on this Vehicle, create a Treasure token and a 1/1 colorless Pilot creature token with \"This token crews Vehicles as though its power were 2 greater.\"\n" +
			"Crew 3",
		Power:     &power,
		Toughness: &toughness,
	})
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	pilot := multiTokenDef(t, bankbusterCreateToken(t, sequence[2]))
	if len(pilot.StaticAbilities) != 1 || pilot.StaticAbilities[0].CrewPowerBonus != 2 {
		t.Fatalf("Pilot token = %#v", pilot)
	}
}
