package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Scryfall printing TDC 189, released 2025-04-11,
// printing ID 2002d779-9ea2-432b-8c15-d61995b8adda, Oracle ID
// b48a149c-aead-43b2-a99e-c964d6d28aa0. Its rulings URI
// https://api.scryfall.com/cards/2002d779-9ea2-432b-8c15-d61995b8adda/rulings
// returned two WotC rulings dated 2014-11-07: black and Zombie are added
// rather than replacing other characteristics, and a colorless creature
// becomes black. Read verified local CR copy
// /Users/natefinch/.copilot/session-state/2d21805f-c9bb-4910-a872-eb7b8d5d9459/files/high-review-rules.txt
// (effective 2026-09-25; SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 105.3 requires additive color to preserve existing colors; 400.6-7 require
// replacement-aware zone movement and new object identities; 608.2c-f require
// written resolution order, a legal resolving choice, and simultaneous actions;
// 611.2a/e make the untimed color/type effect indefinite and apply it as the
// creature enters; 613.1d/e apply subtype and color changes in their layers;
// 701.8 defines destruction and its regeneration replacement.

type necromanticSelectionChooser struct {
	name    string
	asked   bool
	options []string
}

func (a *necromanticSelectionChooser) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *necromanticSelectionChooser) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceResolution {
		return request.DefaultSelection
	}
	a.asked = true
	a.options = nil
	for _, option := range request.Options {
		a.options = append(a.options, option.Label)
	}
	for _, option := range request.Options {
		if option.Label == a.name {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

func checkedNecromanticSelection(t *testing.T) (*Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(authored.NecromanticSelection.OracleID)
	if err != nil || def == nil || def.Name != "Necromantic Selection" {
		t.Fatalf("checked default-registry lookup = %v, %v", def, err)
	}
	evidence := registry.SupportFor(def.Name)
	if len(evidence) != 1 || evidence[0].OracleID != authored.NecromanticSelection.OracleID ||
		!evidence[0].Loadable || !evidence[0].BehaviorVerified {
		t.Fatalf("exact authored identity was not checked into the default registry: %+v", evidence)
	}
	return NewEngineWithPrograms(nil, registry.Programs()), def
}

func castNecromanticSelection(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef) game.ObjectID {
	t.Helper()
	spell := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player1].ManaPool.Add(mana.B, 3)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("paid Necromantic Selection did not cast as a nontargeted sorcery")
	}
	return spell
}

func assertNecromanticDestructionBatch(t *testing.T, g *game.Game, want int) {
	t.Helper()
	var batch game.ObjectID
	count := 0
	for _, event := range g.Events {
		if event.Kind != game.EventZoneChanged || event.FromZone != zone.Battlefield ||
			event.ToZone != zone.Graveyard && event.ToZone != zone.Exile {
			continue
		}
		if event.SimultaneousID == 0 || batch != 0 && event.SimultaneousID != batch {
			t.Fatal("creature destruction events were not one simultaneous batch")
		}
		batch = event.SimultaneousID
		count++
	}
	if count != want {
		t.Fatalf("simultaneous creature destruction events = %d, want %d", count, want)
	}
}

func TestNecromanticSelectionCheckedAdmissionAndChoices(t *testing.T) {
	t.Parallel()
	engine, def := checkedNecromanticSelection(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	elfDef := vanillaCreature("Green Elf", 2, 3)
	elfDef.Colors, elfDef.Subtypes = []color.Color{color.Green}, []types.Sub{types.Elf}
	elf := addCombatPermanent(g, game.Player2, elfDef)
	colorless := addCombatPermanent(g, game.Player3, vanillaCreature("Colorless Creature", 2, 2))
	protectedDef := vanillaCreature("Indestructible Creature", 3, 3)
	protectedDef.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
	protected := addCombatPermanent(g, game.Player4, protectedDef)
	redirected := addCombatPermanent(g, game.Player4, vanillaCreature("Redirected Creature", 1, 1))
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Artifact", Types: []types.Card{types.Artifact},
	}})
	prior := addCardToGraveyard(g, game.Player2, vanillaCreature("Prior Graveyard Creature", 2, 2))
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Battlefield,
		MatchToZone: true, ToZone: zone.Graveyard, AffectedObjectID: redirected.ObjectID,
		ReplaceToZone: zone.Exile, Duration: game.DurationPermanent,
	})

	spell := castNecromanticSelection(t, g, engine, def)
	branch := g.Clone()
	originalChooser := &necromanticSelectionChooser{name: "Green Elf"}
	branchChooser := &necromanticSelectionChooser{name: "Colorless Creature"}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: originalChooser}, &TurnLog{})
	engine.resolveTopOfStackWithChoices(branch, [game.NumPlayers]PlayerAgent{game.Player1: branchChooser}, &TurnLog{})

	assertNecromanticDestructionBatch(t, g, 3)
	assertNecromanticDestructionBatch(t, branch, 3)
	if !originalChooser.asked || !branchChooser.asked {
		t.Fatal("controller was not offered a choice among actual creature-card arrivals")
	}
	if permanentByCardID(g, elf.CardInstanceID) == nil || permanentByCardID(g, colorless.CardInstanceID) != nil ||
		permanentByCardID(branch, colorless.CardInstanceID) == nil || permanentByCardID(branch, elf.CardInstanceID) != nil {
		t.Fatal("independent game clones did not return the controller's selected creature")
	}
	if !g.Players[game.Player3].Graveyard.Contains(colorless.CardInstanceID) ||
		!branch.Players[game.Player2].Graveyard.Contains(elf.CardInstanceID) ||
		!g.Players[game.Player2].Graveyard.Contains(prior) ||
		!branch.Players[game.Player2].Graveyard.Contains(prior) {
		t.Fatal("effect returned an unselected or preexisting graveyard card")
	}
	returnedElf := permanentByCardID(g, elf.CardInstanceID)
	if returnedElf == nil || effectiveController(g, returnedElf) != game.Player1 ||
		!permanentHasSubtype(g, returnedElf, types.Elf) || !permanentHasSubtype(g, returnedElf, types.Zombie) ||
		!slices.Contains(permanentEffectiveColors(g, returnedElf), color.Green) ||
		!slices.Contains(permanentEffectiveColors(g, returnedElf), color.Black) {
		t.Fatal("returned Elf did not retain its original type/color, gain Zombie/black, and follow the spell controller")
	}
	returnedColorless := permanentByCardID(branch, colorless.CardInstanceID)
	if returnedColorless == nil || effectiveController(branch, returnedColorless) != game.Player1 ||
		!permanentHasSubtype(branch, returnedColorless, types.Zombie) ||
		!slices.Equal(permanentEffectiveColors(branch, returnedColorless), []color.Color{color.Black}) {
		t.Fatal("returned colorless creature did not become black, gain Zombie, and follow the spell controller")
	}
	for _, state := range []*game.Game{g, branch} {
		if _, exists := permanentByObjectID(state, protected.ObjectID); !exists ||
			!state.Players[game.Player4].Exile.Contains(redirected.CardInstanceID) ||
			permanentByCardID(state, artifact.CardInstanceID) == nil ||
			!state.Players[game.Player1].Exile.Contains(spell) {
			t.Fatal("destroy protection, actual replacement destination, noncreature, or spell exile was incorrect")
		}
	}
}

func TestNecromanticSelectionNoEligibleArrival(t *testing.T) {
	t.Parallel()
	engine, def := checkedNecromanticSelection(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	protectedDef := vanillaCreature("Only Indestructible Creature", 2, 2)
	protectedDef.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
	protected := addCombatPermanent(g, game.Player2, protectedDef)
	redirected := addCombatPermanent(g, game.Player3, vanillaCreature("Exiled Instead", 2, 2))
	prior := addCardToGraveyard(g, game.Player2, vanillaCreature("Not Put There This Way", 2, 2))
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Battlefield,
		MatchToZone: true, ToZone: zone.Graveyard, AffectedObjectID: redirected.ObjectID,
		ReplaceToZone: zone.Exile, Duration: game.DurationPermanent,
	})

	spell := castNecromanticSelection(t, g, engine, def)
	chooser := &necromanticSelectionChooser{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: chooser}, &TurnLog{})

	assertNecromanticDestructionBatch(t, g, 1)
	if chooser.asked {
		t.Fatal("no choice should be offered when no creature card was put into a graveyard this way")
	}

	if _, exists := permanentByObjectID(g, protected.ObjectID); !exists ||
		!g.Players[game.Player3].Exile.Contains(redirected.CardInstanceID) ||
		!g.Players[game.Player2].Graveyard.Contains(prior) ||
		!g.Players[game.Player1].Exile.Contains(spell) {
		t.Fatal("empty eligible arrival set caused an invalid return or failed to finish the spell")
	}
}

func TestNecromanticSelectionExcludesRegeneratedCreatureAndKeepsEntryEffects(t *testing.T) {
	engine, def := checkedNecromanticSelection(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	regenerated := addCombatPermanent(g, game.Player2, vanillaCreature("Regenerated creature", 3, 3))
	regenerated.RegenerationShields = 1
	creatureDef := vanillaCreature("Actual graveyard arrival", 2, 3)
	creatureDef.Colors, creatureDef.Subtypes = []color.Color{color.Green}, []types.Sub{types.Elf}
	destroyed := addCombatPermanent(g, game.Player3, creatureDef)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventPermanentEnteredBattlefield,
		EntersTapped: true, EntersTappedOthers: true,
		EntersTappedSelection: &game.Selection{
			SubtypesAny: []types.Sub{types.Zombie}, ColorsAny: []color.Color{color.Black},
		},
	})
	spell := castNecromanticSelection(t, g, engine, def)
	chooser := &necromanticSelectionChooser{name: creatureDef.Name}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: chooser,
	}, &TurnLog{})

	if !chooser.asked || !slices.Equal(chooser.options, []string{creatureDef.Name}) {
		t.Fatalf("return choices = %v, want only the actual destroyed creature card", chooser.options)
	}
	if permanentByCardID(g, regenerated.CardInstanceID) != regenerated || !regenerated.Tapped ||
		regenerated.RegenerationShields != 0 || g.Players[game.Player2].Graveyard.Contains(regenerated.CardInstanceID) {
		t.Fatal("regeneration did not replace destruction without offering the protected creature")
	}
	returned := permanentByCardID(g, destroyed.CardInstanceID)
	if returned == nil || !returned.Tapped {
		t.Fatal("entry replacement did not see the entering creature as a black Zombie")
	}
	engine.runCleanupStep(g, simPassPolicies(), &TurnLog{})
	if !permanentHasSubtype(g, returned, types.Zombie) || !permanentHasSubtype(g, returned, types.Elf) ||
		!slices.Contains(permanentEffectiveColors(g, returned), color.Black) ||
		!slices.Contains(permanentEffectiveColors(g, returned), color.Green) ||
		!g.Players[game.Player1].Exile.Contains(spell) {
		t.Fatal("permanent additive entry effects or spell exile were lost at cleanup")
	}
}
