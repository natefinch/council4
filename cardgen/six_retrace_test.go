package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const sixOracle = "Reach\n" +
	"Whenever Six attacks, mill three cards. You may put a land card from among them into your hand.\n" +
	"During your turn, nonland permanent cards in your graveyard have retrace. (You may cast permanent cards from your graveyard by discarding a land card in addition to paying their other costs.)"

func sixCard() *ScryfallCard {
	power, toughness := "3", "3"
	return &ScryfallCard{
		Name:       "Six",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Legendary Creature — Treefolk",
		OracleText: sixOracle,
		Power:      &power,
		Toughness:  &toughness,
	}
}

func TestGenerateExecutableCardSourceSix(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(sixCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ReachStaticBody",
		"Kind:                           game.RuleEffectGrantGraveyardCardKeyword",
		"AffectedPlayer:                 game.PlayerYou",
		"GrantedKeyword:                 game.Retrace",
		"RestrictedDuringControllerTurn: true",
		"RequiredTypesAny: []types.Card{types.Creature, types.Artifact, types.Enchantment, types.Planeswalker, types.Battle}",
		"ExcludedTypes: []types.Card{types.Land}",
		"game.Mill{",
		"PublishLinked: game.LinkedKey(\"milled-cards\")",
		"game.ChooseFromZone{",
		"FromLinked: game.LinkedKey(\"milled-cards\")",
		"Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerSixStaticGrantsGraveyardRetrace(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, sixCard())
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d, want 2", len(face.StaticAbilities))
	}
	grant := face.StaticAbilities[1].Body
	if len(grant.RuleEffects) != 1 {
		t.Fatalf("rule effects = %d, want 1", len(grant.RuleEffects))
	}
	effect := grant.RuleEffects[0]
	if effect.Kind != game.RuleEffectGrantGraveyardCardKeyword {
		t.Fatalf("rule effect kind = %v", effect.Kind)
	}
	if effect.GrantedKeyword != game.Retrace {
		t.Fatalf("granted keyword = %v, want Retrace", effect.GrantedKeyword)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if !effect.RestrictedDuringControllerTurn {
		t.Fatal("expected RestrictedDuringControllerTurn for \"During your turn\"")
	}
	if !slices.Equal(effect.CardSelection.RequiredTypesAny, []types.Card{types.Creature, types.Artifact, types.Enchantment, types.Planeswalker, types.Battle}) {
		t.Fatalf("required types any = %#v", effect.CardSelection.RequiredTypesAny)
	}
	if !slices.Equal(effect.CardSelection.ExcludedTypes, []types.Card{types.Land}) {
		t.Fatalf("excluded types = %#v", effect.CardSelection.ExcludedTypes)
	}
}

func TestLowerSixMillThenOptionalLandReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, sixCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	mill, ok := sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("first primitive = %#v, want Mill", sequence[0].Primitive)
	}
	if mill.PublishLinked == "" {
		t.Fatal("mill must publish the milled cards under a linked key")
	}
	if sequence[0].Optional {
		t.Fatal("mill is mandatory")
	}
	ret, ok := sequence[1].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("second primitive = %#v, want ChooseFromZone", sequence[1].Primitive)
	}
	if !sequence[1].Optional {
		t.Fatal("land-to-hand return is optional")
	}
	if ret.Riders.FromLinked != mill.PublishLinked {
		t.Fatalf("return FromLinked = %q, mill PublishLinked = %q", ret.Riders.FromLinked, mill.PublishLinked)
	}
	if ret.Destination.Zone != zone.Hand {
		t.Fatalf("return destination = %v, want hand", ret.Destination.Zone)
	}
	if !slices.Equal(ret.Filter.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("return selection required types = %#v", ret.Filter.RequiredTypes)
	}
}

func TestGenerateExecutableCardSourceFlameJabRetrace(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Flame Jab",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: "Flame Jab deals 1 damage to any target.\nRetrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.RetraceStaticBody") {
		t.Fatalf("generated source missing Retrace keyword:\n%s", source)
	}
}

func TestSixStaticRetraceGrantFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		// Missing "your" before graveyard.
		"During your turn, nonland permanent cards in graveyard have retrace.",
		// Unsupported filter ("artifact").
		"During your turn, artifact and land cards in your graveyard have retrace.",
		// Battlefield zone instead of graveyard is a different (unsupported) shape.
		"During your turn, nonland permanent cards on the battlefield have retrace.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Near Miss",
				Layout:     "normal",
				TypeLine:   "Legendary Creature — Treefolk",
				OracleText: oracle,
				Power:      new("3"),
				Toughness:  new("3"),
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracle)
			}
			for _, face := range faces {
				for _, static := range face.StaticAbilities {
					for _, effect := range static.Body.RuleEffects {
						if effect.Kind == game.RuleEffectGrantGraveyardCardKeyword {
							t.Fatalf("unexpected graveyard keyword grant for %q", oracle)
						}
					}
				}
			}
		})
	}
}
