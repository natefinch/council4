package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// castSpellsFromLibraryTopPermanent gives playerID a battlefield permanent whose
// static ability lets that player cast spells of the given types (empty means
// any) from the top of their library (Future Sight, Bolas's Citadel).
func castSpellsFromLibraryTopPermanent(g *game.Game, playerID game.PlayerID, spellTypes []types.Card) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Seer",
		StaticAbilities: []game.StaticAbility{{
			Text: "You may cast spells from the top of your library.",
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastSpellsFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Library,
				SpellTypes:     spellTypes,
				TopCardOnly:    true,
			}},
		}},
	}})
}

func TestCanCastSpellFromLibraryTopRequiresStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Bolt", Types: []types.Card{types.Instant}}})

	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Library, game.FaceFront) {
		t.Fatal("library spell is castable without the static permission")
	}

	castSpellsFromLibraryTopPermanent(g, game.Player1, nil)

	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Library, game.FaceFront) {
		t.Fatal("top library spell is not castable despite the static permission")
	}
	if canCastSpellsFromZoneByRuleEffect(g, game.Player2, spellID, zone.Library, game.FaceFront) {
		t.Fatal("opponent may cast from a library they do not control the static for")
	}
}

func TestCanCastSpellFromLibraryTopOnlyTopCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
	buriedID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Buried Bolt", Types: []types.Card{types.Instant}}})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Bolt", Types: []types.Card{types.Instant}}})

	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, topID, zone.Library, game.FaceFront) {
		t.Fatal("top library spell is not castable despite the static permission")
	}
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, buriedID, zone.Library, game.FaceFront) {
		t.Fatal("a non-top library spell must not be castable")
	}
}

func TestCanCastSpellFromLibraryTopRespectsTypeFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpellsFromLibraryTopPermanent(g, game.Player1, []types.Card{types.Creature})
	instantID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Bolt", Types: []types.Card{types.Instant}}})

	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, instantID, zone.Library, game.FaceFront) {
		t.Fatal("a non-creature spell must not be castable under a creature-only filter")
	}

	creatureID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Bear", Types: []types.Card{types.Creature}}})
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, creatureID, zone.Library, game.FaceFront) {
		t.Fatal("a creature spell is not castable despite the creature-only filter")
	}
}

// castSpellsFromLibraryTopColorlessPermanent gives playerID a permanent whose
// static lets that player cast spells matching the given card types or any
// colorless spell from the top of their library ("You may cast artifact spells
// and colorless spells from the top of your library.", Mystic Forge).
func castSpellsFromLibraryTopColorlessPermanent(g *game.Game, playerID game.PlayerID, spellTypes []types.Card) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Forge",
		StaticAbilities: []game.StaticAbility{{
			Text: "You may cast artifact spells and colorless spells from the top of your library.",
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastSpellsFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Library,
				SpellTypes:     spellTypes,
				SpellColorless: true,
				TopCardOnly:    true,
			}},
		}},
	}})
}

func TestCanCastSpellFromLibraryTopRespectsColorlessFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpellsFromLibraryTopColorlessPermanent(g, game.Player1, []types.Card{types.Artifact})

	coloredID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Blue Sorcery", Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Blue}}})
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, coloredID, zone.Library, game.FaceFront) {
		t.Fatal("a colored non-artifact spell must not be castable under an artifact-or-colorless filter")
	}

	colorlessID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Eldrazi", Types: []types.Card{types.Creature}}})
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, colorlessID, zone.Library, game.FaceFront) {
		t.Fatal("a colorless spell is not castable despite the colorless permission")
	}

	artifactID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Colored Artifact", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Blue}}})
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, artifactID, zone.Library, game.FaceFront) {
		t.Fatal("a colored artifact spell is not castable despite the artifact permission")
	}
}

func TestLegalCastActionsIncludesLibraryTopSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
	spellID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Free Instant", Types: []types.Card{types.Instant}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	found := false
	for _, act := range engine.legalCastActions(g, game.Player1) {
		payload, ok := act.CastSpellPayload()
		if ok && payload.CardID == spellID && payload.SourceZone == zone.Library {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no legal cast action offered for the top library spell despite the static permission")
	}
}

// castSpellsFromLibraryTopChosenTypePermanent gives playerID a battlefield
// permanent whose static lets that player cast creature spells sharing the
// creature subtype the permanent chose as it entered ("You may cast creature
// spells of the chosen type from the top of your library.", Realmwalker).
func castSpellsFromLibraryTopChosenTypePermanent(g *game.Game, playerID game.PlayerID, chosen types.Sub) *game.Permanent {
	permanent := addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Realmwalker",
		StaticAbilities: []game.StaticAbility{{
			Text: "You may cast creature spells of the chosen type from the top of your library.",
			RuleEffects: []game.RuleEffect{{
				Kind:                   game.RuleEffectCastSpellsFromZone,
				AffectedPlayer:         game.PlayerYou,
				CastFromZone:           zone.Library,
				SpellTypes:             []types.Card{types.Creature},
				TopCardOnly:            true,
				SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,
			}},
		}},
	}})
	permanent.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: chosen},
	}
	return permanent
}

func TestCanCastSpellFromLibraryTopRespectsChosenType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpellsFromLibraryTopChosenTypePermanent(g, game.Player1, types.Elf)

	matchID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf}}})
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, matchID, zone.Library, game.FaceFront) {
		t.Fatal("a creature spell of the chosen type is not castable despite the chosen-type permission")
	}

	mismatchID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Goblin", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Goblin}}})
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, mismatchID, zone.Library, game.FaceFront) {
		t.Fatal("a creature spell of a different subtype must not be castable under the chosen-type permission")
	}

	noncreatureID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Bolt", Types: []types.Card{types.Instant}}})
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, noncreatureID, zone.Library, game.FaceFront) {
		t.Fatal("a noncreature spell must not be castable under the chosen-type permission")
	}
}

func TestCanCastSpellFromLibraryTopChosenTypeRequiresChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := castSpellsFromLibraryTopChosenTypePermanent(g, game.Player1, types.Elf)
	permanent.EntryChoices = nil

	matchID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf}}})
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, matchID, zone.Library, game.FaceFront) {
		t.Fatal("no spell may be cast under the chosen-type permission before a type is chosen")
	}
}
