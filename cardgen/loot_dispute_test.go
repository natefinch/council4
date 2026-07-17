package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

const lootDisputeOracle = "When this enchantment enters, you take the initiative and create a Treasure token.\n" +
	"Whenever you attack the player who has the initiative, create a Treasure token.\n" +
	"Loud Ruckus — Whenever you complete a dungeon, create a 5/5 red Dragon creature token with flying."

func TestLootDisputeLowersReusableMechanics(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Loot Dispute",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Enchantment",
		OracleText: lootDisputeOracle,
	})
	if len(face.TriggeredAbilities) != 3 {
		t.Fatalf("triggered abilities = %d, want 3", len(face.TriggeredAbilities))
	}

	etb := face.TriggeredAbilities[0]
	if len(etb.Content.Modes) != 1 || len(etb.Content.Modes[0].Sequence) != 2 {
		t.Fatalf("ETB content = %#v, want initiative then Treasure", etb.Content)
	}
	if _, ok := etb.Content.Modes[0].Sequence[0].Primitive.(game.TakeInitiative); !ok {
		t.Fatalf("ETB first primitive = %T, want game.TakeInitiative", etb.Content.Modes[0].Sequence[0].Primitive)
	}
	assertLootDisputeToken(t, etb.Content.Modes[0].Sequence[1].Primitive, "Treasure", 0, false)

	attack := face.TriggeredAbilities[1]
	if got := attack.Trigger.Pattern; got.Event != game.EventAttackerDeclared ||
		got.Controller != game.TriggerControllerYou ||
		got.Player != game.TriggerPlayerInitiative ||
		got.AttackRecipient != game.AttackRecipientPlayer ||
		!got.OneOrMore || !got.OneOrMorePerAttackTarget {
		t.Fatalf("attack pattern = %#v", got)
	}
	assertLootDisputeToken(t, attack.Content.Modes[0].Sequence[0].Primitive, "Treasure", 0, false)

	completion := face.TriggeredAbilities[2]
	if got := completion.Trigger.Pattern; got.Event != game.EventCompletedDungeon ||
		got.Player != game.TriggerPlayerYou {
		t.Fatalf("completion pattern = %#v", got)
	}
	assertLootDisputeToken(t, completion.Content.Modes[0].Sequence[0].Primitive, "Dragon", 5, true)
}

func TestLootDisputeGeneratesExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Loot Dispute",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Enchantment",
		OracleText: lootDisputeOracle,
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.TakeInitiative{",
		"game.EventAttackerDeclared",
		"game.TriggerPlayerInitiative",
		"OneOrMore:                true",
		"OneOrMorePerAttackTarget: true",
		"game.EventCompletedDungeon",
		`Name:      "Dragon"`,
		"Colors:    []color.Color{color.Red}",
		"Subtypes:  []types.Sub{types.Dragon}",
		"Power:     opt.Val(game.PT{Value: 5})",
		"game.FlyingStaticBody",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func assertLootDisputeToken(t *testing.T, primitive game.Primitive, name string, size int, flying bool) {
	t.Helper()
	create, ok := primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok || def.Name != name {
		t.Fatalf("token = %#v, want %s definition", def, name)
	}
	if name != "Dragon" {
		return
	}
	if !def.Power.Exists || !def.Toughness.Exists ||
		def.Power.Val.Value != size || def.Toughness.Val.Value != size ||
		len(def.Colors) != 1 || def.Colors[0] != color.Red ||
		len(def.Types) != 1 || def.Types[0] != types.Creature ||
		len(def.Subtypes) != 1 || def.Subtypes[0] != types.Dragon ||
		len(def.StaticAbilities) != 1 ||
		!game.BodyHasKeyword(&def.StaticAbilities[0], game.Flying) {
		t.Fatalf("Dragon definition = %#v", def.CardFace)
	}
	if create.Recipient.Exists {
		t.Fatalf("Dragon recipient = %#v, want default ability controller", create.Recipient)
	}
}
