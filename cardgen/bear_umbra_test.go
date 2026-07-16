package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

const bearUmbraOracle = "Enchant creature\n" +
	`Enchanted creature gets +2/+2 and has "Whenever this creature attacks, untap all lands you control."` + "\n" +
	"Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)"

func TestLowerBearUmbraComposesGrantAndUmbraArmor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bear Umbra",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: bearUmbraOracle,
	})

	var foundArmor bool
	var granted *game.TriggeredAbility
	for i := range face.StaticAbilities {
		body := &face.StaticAbilities[i].Body
		if game.BodyHasKeyword(body, game.UmbraArmor) {
			foundArmor = true
		}
		for _, effect := range body.ContinuousEffects {
			for _, ability := range effect.AddAbilities {
				if trigger, ok := ability.(*game.TriggeredAbility); ok {
					granted = trigger
				}
			}
		}
	}
	if !foundArmor {
		t.Fatal("lowered face has no Umbra armor static ability")
	}
	if granted == nil {
		t.Fatal("lowered face has no quoted granted triggered ability")
	}
	if granted.Trigger.Pattern.Event != game.EventAttackerDeclared ||
		granted.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("granted trigger = %#v, want self attack trigger", granted.Trigger)
	}
	if len(granted.Content.Modes) != 1 || len(granted.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("granted content = %#v, want one untap instruction", granted.Content)
	}
	untap, ok := granted.Content.Modes[0].Sequence[0].Primitive.(game.Untap)
	if !ok {
		t.Fatalf("granted primitive = %T, want game.Untap", granted.Content.Modes[0].Sequence[0].Primitive)
	}
	selection := untap.Group.Selection()
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Land}) ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("untap group = %#v, want lands controlled by ability controller", untap.Group)
	}
}

func TestGenerateBearUmbraUsesReusableMechanics(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Bear Umbra",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: bearUmbraOracle,
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.UmbraArmorStaticBody",
		"game.EventAttackerDeclared",
		"game.TriggerSourceSelf",
		"game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou})",
		"types.Land",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
