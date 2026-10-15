package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestUnqualifiedSelfSacrificeManaRecognition(t *testing.T) {
	t.Parallel()
	for _, qualified := range []bool{false, true} {
		t.Run(map[bool]string{false: "self-reference", true: "legacy artifact-qualified"}[qualified], func(t *testing.T) {
			t.Parallel()
			body := game.TapSacrificeAnyColorManaAbility("{T}, Sacrifice this token: Add one mana of any color.")
			if qualified {
				body = game.TapSacrificeAnyOneColorManaAbility("{T}, Sacrifice this artifact: Add one mana of any color.", 1)
			}
			output, ok := paymentManaAbilityOutput(&body)
			if !ok || !output.sacrifice || output.amount != 1 || len(output.colors) != 5 {
				t.Fatal("valid tap/self-sacrifice output not recognized")
			}
		})
	}
	for _, scenario := range []string{"two sacrifices", "extra life", "subtype constraint", "untap", "optional output", "extra effect"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			body := game.TapSacrificeAnyColorManaAbility("Unqualified source")
			switch scenario {
			case "two sacrifices":
				body.AdditionalCosts[1].Amount = 2
			case "extra life":
				body.AdditionalCosts = append(body.AdditionalCosts, cost.Additional{Kind: cost.AdditionalPayLife, Amount: 1})
			case "subtype constraint":
				body.AdditionalCosts[1].SubtypesAny = cost.SubtypeSet{types.Treasure}
			case "untap":
				body.AdditionalCosts[0].Kind = cost.AdditionalUntap
			case "optional output":
				body.Content.Modes[0].Sequence[1].Optional = true
			case "extra effect":
				body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence, game.Instruction{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Fixed(1),
				}})
			default:
				t.Fatal("unknown negative")
			}
			if _, ok := paymentManaAbilityOutput(&body); ok {
				t.Fatal("unmodeled compound/conditional mana source accepted")
			}
		})
	}
}
