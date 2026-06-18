package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func costFor(manaValue int) opt.V[cost.Mana] {
	if manaValue <= 0 {
		return opt.V[cost.Mana]{}
	}
	return opt.Val(cost.Mana{cost.O(manaValue)})
}

func spell(name string, manaValue int, cardType types.Card, c color.Color, targeted bool, primitives ...game.Primitive) *game.CardDef {
	sequence := make([]game.Instruction, 0, len(primitives))
	for _, primitive := range primitives {
		sequence = append(sequence, game.Instruction{Primitive: primitive})
	}
	mode := game.Mode{Sequence: sequence}
	if targeted {
		mode.Targets = []game.TargetSpec{{}}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:         name,
		Types:        []types.Card{cardType},
		Colors:       []color.Color{c},
		ManaCost:     costFor(manaValue),
		SpellAbility: opt.Val(mode.Ability()),
	}}
}

func manaRock(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Artifact},
		ManaCost:      costFor(2),
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
	}}
}

func flyer(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.White},
		ManaCost:  costFor(manaValue),
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flying),
		}},
	}}
}

func bigThreat(name string, manaValue, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		ManaCost:  costFor(manaValue),
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}}
}

func repeat(def *game.CardDef, n int) []*game.CardDef {
	cards := make([]*game.CardDef, 0, n)
	for range n {
		cards = append(cards, def)
	}
	return cards
}

func manaDork(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Creature},
		Colors:        []color.Color{color.Green},
		ManaCost:      costFor(1),
		Power:         opt.Val(game.PT{Value: 1}),
		Toughness:     opt.Val(game.PT{Value: 1}),
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}}
}

func TestAnalyzeDeckTagsAndCurve(t *testing.T) {
	deck := []*game.CardDef{
		manaRock("Mind Stone"),
		manaDork("Llanowar Elves"),
		spell("Doom Blade", 2, types.Instant, color.Black, true, game.Destroy{}),
		spell("Wrath", 4, types.Sorcery, color.White, false, game.Destroy{}),
		spell("Counterspell", 2, types.Instant, color.Blue, true, game.CounterObject{}),
		spell("Divination", 3, types.Sorcery, color.Blue, false, game.Draw{}),
		spell("Demonic Tutor", 2, types.Sorcery, color.Black, false, game.Search{}),
		bigThreat("Hydra", 6, 7),
		&game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}},
	}
	profile := AnalyzeDeck(game.PlayerConfig{Deck: deck})

	wantTags := map[CardTag]int{
		TagManaRock:     1,
		TagManaDork:     1,
		TagRamp:         2,
		TagRemoval:      1,
		TagBoardWipe:    1,
		TagCounterspell: 1,
		TagDraw:         1,
		TagTutor:        1,
		TagInteraction:  2, // Doom Blade (spot instant) + Counterspell
		TagThreat:       1,
	}
	for tag, want := range wantTags {
		if got := profile.TagCounts[tag]; got != want {
			t.Errorf("TagCounts[%v] = %d, want %d", tag, got, want)
		}
	}
	if profile.Curve.NonlandCount != 8 {
		t.Errorf("NonlandCount = %d, want 8 (land excluded)", profile.Curve.NonlandCount)
	}
	if profile.Curve.Buckets[2] != 4 {
		t.Errorf("Buckets[2] = %d, want 4 (Mind Stone, Doom Blade, Counterspell, Demonic Tutor)", profile.Curve.Buckets[2])
	}
	if profile.Curve.Buckets[6] != 1 {
		t.Errorf("Buckets[6] = %d, want 1 (Hydra)", profile.Curve.Buckets[6])
	}
}

func TestAnalyzeDeckClassifiesControl(t *testing.T) {
	deck := []*game.CardDef{}
	deck = append(deck, repeat(spell("Removal", 2, types.Instant, color.Black, true, game.Destroy{}), 6)...)
	deck = append(deck, repeat(spell("Counter", 2, types.Instant, color.Blue, true, game.CounterObject{}), 4)...)
	deck = append(deck, repeat(spell("Cantrip", 1, types.Sorcery, color.Blue, false, game.Draw{}), 6)...)
	deck = append(deck, repeat(spell("Wrath", 4, types.Sorcery, color.White, false, game.Destroy{}), 2)...)
	profile := AnalyzeDeck(game.PlayerConfig{Deck: deck})

	if profile.Archetype != ArchetypeControl {
		t.Errorf("Archetype = %v, want ArchetypeControl", profile.Archetype)
	}
}

func TestAnalyzeDeckClassifiesAggroAndPower(t *testing.T) {
	deck := repeat(flyer("Faerie", 1), 20)
	profile := AnalyzeDeck(game.PlayerConfig{Deck: deck})

	if profile.Archetype != ArchetypeAggro {
		t.Errorf("Archetype = %v, want ArchetypeAggro", profile.Archetype)
	}
	if profile.Curve.AverageMV > fastCurveMV {
		t.Errorf("AverageMV = %v, want a fast curve <= %v", profile.Curve.AverageMV, fastCurveMV)
	}
	// A fast, threat-dense deck should goldfish quickly and land in a high band.
	if profile.Bracket != BracketHigh && profile.Bracket != BracketCEDH {
		t.Errorf("Bracket = %v, want a fast (High/cEDH) bracket", profile.Bracket)
	}
}

func TestCommanderProfileWinconAndTrajectory(t *testing.T) {
	commander := bigThreat("Voltron Lord", 4, 6)
	commander.ColorIdentity = color.NewIdentity(color.Green)
	profile := AnalyzeDeck(game.PlayerConfig{Commander: commander})

	if profile.Commander.Name != "Voltron Lord" {
		t.Errorf("Commander.Name = %q, want Voltron Lord", profile.Commander.Name)
	}
	if profile.Commander.Role != RoleWincon {
		t.Errorf("Commander.Role = %v, want RoleWincon (power %d)", profile.Commander.Role, 6)
	}
	if profile.Commander.CastTrajectory != [4]int{4, 6, 8, 10} {
		t.Errorf("CastTrajectory = %v, want [4 6 8 10]", profile.Commander.CastTrajectory)
	}
}

func TestAnalyzeDeckIsDeterministic(t *testing.T) {
	deck := []*game.CardDef{
		manaRock("Sol Ring"),
		bigThreat("Beast", 5, 5),
		spell("Bolt", 1, types.Instant, color.Red, true, game.Damage{}),
	}
	config := game.PlayerConfig{Deck: deck}
	first := AnalyzeDeck(config)
	for range 10 {
		again := AnalyzeDeck(config)
		if again.Archetype != first.Archetype || again.GoldfishTurn != first.GoldfishTurn || again.Curve != first.Curve {
			t.Fatalf("AnalyzeDeck not deterministic: %+v vs %+v", again, first)
		}
	}
}

// TestModalRemovalWithSharedTargetsIsSpotRemoval checks that a modal removal
// spell whose targets live in the ability's SharedTargets (charms/commands) is
// classified as spot removal and interaction, not as a board wipe.
func TestModalRemovalWithSharedTargetsIsSpotRemoval(t *testing.T) {
	charm := &game.CardDef{CardFace: game.CardFace{
		Name:     "Charm",
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.White},
		ManaCost: costFor(3),
		SpellAbility: opt.Val(game.AbilityContent{
			SharedTargets: []game.TargetSpec{{}},
			Modes: []game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.Destroy{}}}},
				{Sequence: []game.Instruction{{Primitive: game.Draw{}}}},
			},
			MinModes: 1,
			MaxModes: 1,
		}),
	}}
	profile := AnalyzeDeck(game.PlayerConfig{Deck: []*game.CardDef{charm}})

	if profile.TagCounts[TagRemoval] != 1 {
		t.Errorf("TagRemoval = %d, want 1 (modal removal with shared targets)", profile.TagCounts[TagRemoval])
	}
	if profile.TagCounts[TagBoardWipe] != 0 {
		t.Errorf("TagBoardWipe = %d, want 0 (shared-target removal is not a wipe)", profile.TagCounts[TagBoardWipe])
	}
	if profile.TagCounts[TagInteraction] != 1 {
		t.Errorf("TagInteraction = %d, want 1 (instant-speed spot removal)", profile.TagCounts[TagInteraction])
	}
}

// TestGenericStrategyCarriesProfile checks a strategy can consult the deck
// analysis.
func TestGenericStrategyCarriesProfile(t *testing.T) {
	profile := AnalyzeDeck(game.PlayerConfig{Deck: repeat(flyer("Faerie", 1), 20)})
	strategy := GenericStrategy{Profile: &profile}
	if strategy.Profile == nil || strategy.Profile.Archetype != ArchetypeAggro {
		t.Errorf("strategy did not carry the expected deck profile: %+v", strategy.Profile)
	}
}
