package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// keepOnePerTypePrimitive extracts the single game.KeepOnePerType primitive from
// a lowered ability's content, failing the test if the content is not exactly one
// mode with one keep-one-per-type instruction.
func keepOnePerTypePrimitive(t *testing.T, content game.AbilityContent) game.KeepOnePerType {
	t.Helper()
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one mode with one instruction", content)
	}
	primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.KeepOnePerType)
	if !ok {
		t.Fatalf("primitive = %#v, want game.KeepOnePerType", content.Modes[0].Sequence[0].Primitive)
	}
	return primitive
}

// TestLowerLilianaDreadhordeGeneral proves the whole card lowers: its passive
// "creature you control dies, draw a card" trigger, its +1 token and −4 mass
// sacrifice loyalty abilities, and — the point of this feature — its −9
// "Each opponent chooses a permanent they control of each permanent type and
// sacrifices the rest." lowers to the generic KeepOnePerType primitive scoped to
// opponents over all six permanent types. lowerSingleFace fails on any
// diagnostic, so this asserts no ability is left unsupported.
func TestLowerLilianaDreadhordeGeneral(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Liliana, Dreadhorde General",
		Layout:   "normal",
		TypeLine: "Legendary Planeswalker — Liliana",
		ManaCost: "{4}{B}{B}",
		Loyalty:  new("6"),
		OracleText: "Whenever a creature you control dies, draw a card.\n" +
			"+1: Create a 2/2 black Zombie creature token.\n" +
			"−4: Each player sacrifices two creatures of their choice.\n" +
			"−9: Each opponent chooses a permanent they control of each permanent type and sacrifices the rest.",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1 (the passive draw)", len(face.TriggeredAbilities))
	}
	if len(face.LoyaltyAbilities) != 3 {
		t.Fatalf("loyalty abilities = %d, want 3", len(face.LoyaltyAbilities))
	}
	var ultimate *game.LoyaltyAbility
	for i := range face.LoyaltyAbilities {
		if face.LoyaltyAbilities[i].LoyaltyCost == -9 {
			ultimate = &face.LoyaltyAbilities[i]
		}
	}
	if ultimate == nil {
		t.Fatalf("no −9 loyalty ability found among %#v", face.LoyaltyAbilities)
	}
	primitive := keepOnePerTypePrimitive(t, ultimate.Content)
	if primitive.Players.Kind != game.PlayerGroupReferenceOpponents {
		t.Errorf("players = %v, want opponents", primitive.Players.Kind)
	}
	wantTypes := []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}
	if !equalCardTypes(primitive.Types, wantTypes) {
		t.Errorf("types = %v, want %v", primitive.Types, wantTypes)
	}
	if len(primitive.AffectedSelection.ExcludedTypes) != 0 || primitive.ControllerChoosesForAll {
		t.Errorf("affected selection = %#v, controllerChoosesForAll = %v, want whole board / per-player", primitive.AffectedSelection, primitive.ControllerChoosesForAll)
	}
}

// TestLowerCataclysm proves the listed-type, all-permanents variant lowers to
// KeepOnePerType over all players keeping one artifact, creature, enchantment,
// and land.
func TestLowerCataclysm(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Cataclysm",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}{W}",
		OracleText: "Each player chooses from among the permanents they control an artifact, a creature, an enchantment, and a land, then sacrifices the rest.",
	}
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	primitive := keepOnePerTypePrimitive(t, face.SpellAbility.Val)
	if primitive.Players.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Errorf("players = %v, want all players", primitive.Players.Kind)
	}
	wantTypes := []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land}
	if !equalCardTypes(primitive.Types, wantTypes) {
		t.Errorf("types = %v, want %v", primitive.Types, wantTypes)
	}
	if len(primitive.AffectedSelection.ExcludedTypes) != 0 {
		t.Errorf("affected selection = %#v, want whole board", primitive.AffectedSelection)
	}
}

// TestLowerCataclysmicGearhulk proves the nonland variant lowers to
// KeepOnePerType over all players keeping one artifact, creature, enchantment,
// and planeswalker, with the affected pool restricted to nonland permanents.
func TestLowerCataclysmicGearhulk(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:      "Cataclysmic Gearhulk",
		Layout:    "normal",
		TypeLine:  "Artifact Creature — Construct",
		ManaCost:  "{3}{W}{W}",
		Power:     new("4"),
		Toughness: new("5"),
		OracleText: "Vigilance\n" +
			"When this creature enters, each player chooses an artifact, a creature, an enchantment, and a planeswalker from among the nonland permanents they control, then sacrifices the rest.",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1 (the ETB)", len(face.TriggeredAbilities))
	}
	primitive := keepOnePerTypePrimitive(t, face.TriggeredAbilities[0].Content)
	if primitive.Players.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Errorf("players = %v, want all players", primitive.Players.Kind)
	}
	wantTypes := []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Planeswalker}
	if !equalCardTypes(primitive.Types, wantTypes) {
		t.Errorf("types = %v, want %v", primitive.Types, wantTypes)
	}
	if !equalCardTypes(primitive.AffectedSelection.ExcludedTypes, []types.Card{types.Land}) {
		t.Errorf("affected selection excluded types = %v, want [land]", primitive.AffectedSelection.ExcludedTypes)
	}
}

// TestLowerTragicArrogance proves the two-sentence controller-chooses form lowers
// to KeepOnePerType over all players keeping one artifact, creature, enchantment,
// and planeswalker, with the effect's controller choosing for every player and
// the affected pool restricted to nonland permanents.
func TestLowerTragicArrogance(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Tragic Arrogance",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{3}{W}{W}",
		OracleText: "For each player, you choose from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker. " +
			"Then each player sacrifices all other nonland permanents they control.",
	}
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	primitive := keepOnePerTypePrimitive(t, face.SpellAbility.Val)
	if primitive.Players.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Errorf("players = %v, want all players", primitive.Players.Kind)
	}
	if !primitive.ControllerChoosesForAll {
		t.Error("controllerChoosesForAll = false, want true (you choose for each player)")
	}
	wantTypes := []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Planeswalker}
	if !equalCardTypes(primitive.Types, wantTypes) {
		t.Errorf("types = %v, want %v", primitive.Types, wantTypes)
	}
	if !equalCardTypes(primitive.AffectedSelection.ExcludedTypes, []types.Card{types.Land}) {
		t.Errorf("affected selection excluded types = %v, want [land]", primitive.AffectedSelection.ExcludedTypes)
	}
}

// TestLowerKeepOnePerTypeFailsClosed proves a sacrifice sentence whose wording
// deviates from the recognized family stays unsupported rather than lowering an
// approximation.
func TestLowerKeepOnePerTypeFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Not Cataclysm",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}{W}",
		OracleText: "Each player chooses from among the permanents they control an artifact, a creature, an enchantment and a land, then sacrifices the rest.",
	}
	face := lowerSingleFaceExpectingUnsupported(t, card)
	if face.SpellAbility.Exists {
		t.Fatalf("spell ability lowered %#v, want unsupported", face.SpellAbility.Val)
	}
}
