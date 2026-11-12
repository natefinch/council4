package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: exact Oracle identities, revisions, text, mana costs, and type
// lines below came from pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Burrenton Bombardier, oracle 13ec2390-e08a-4645-99e9-bac7e41917bb,
// revision fd872395-f351-4f7e-9290-c34d1c98a567 ({2}{W}, 2/2 Creature — Kithkin
// Soldier): "Flying / Reinforce 2—{2}{W} ({2}{W}, Discard this card: Put two
// +1/+1 counters on target creature.)" Scryfall rulings: none.
// Break Ties, oracle 1e44bb4e-aeea-4f43-9cbd-4c24a9b6b022, revision
// 723ec2ef-aa82-449d-88b0-405d06454fc2 ({2}{W} Instant): "Choose one — Destroy
// target artifact. Destroy target enchantment. Exile target card from a
// graveyard. / Reinforce 1—{W} ({W}, Discard this card: Put a +1/+1 counter on
// target creature.)" Scryfall rulings: none.
// Fowl Strike, oracle 390320f7-11b8-4b61-8316-36c396470a09, revision
// 83b6c7ab-f42f-40d2-9cb6-89291d57e27f ({1}{G} Instant): "Destroy target
// creature with flying. / Reinforce 2—{2}{G} ({2}{G}, Discard this card: Put
// two +1/+1 counters on target creature.)" Scryfall rulings: none.
// Swell of Courage, oracle 60033841-71c7-4370-a2bc-e570a5c100ea, revision
// 0fddb694-d863-4364-9bf9-650d7be3b471 ({3}{W}{W} Instant): "Creatures you
// control get +2/+2 until end of turn. / Reinforce X—{X}{W}{W} ({X}{W}{W},
// Discard this card: Put X +1/+1 counters on target creature.)" Scryfall rulings:
// none.
// Hunting Triad, oracle 66bf0c4b-0c36-4d63-a929-5a7427a07279, revision
// c24930f3-34e8-40e1-a798-3df5d2d12ef1 ({3}{G} Kindred Sorcery — Elf): "Create
// three 1/1 green Elf Warrior creature tokens. / Reinforce 3—{3}{G} ({3}{G},
// Discard this card: Put three +1/+1 counters on target creature.)"
// Scryfall's three 2024-06-07 rulings at
// https://api.scryfall.com/cards/c24930f3-34e8-40e1-a798-3df5d2d12ef1/rulings:
// Tribal was replaced by Kindred with no gameplay change; Kindred permits
// noncreature cards to have creature types (and is counted as a card type); and
// Kindred is a card type counted by effects that count card types in a zone.
// Mosquito Guard, oracle 6beaeec2-b4e9-4904-b817-c0118549d4a9, revision
// 626040ca-adb0-49c8-82cd-d1e27926cb79 ({W}, 1/1 Creature — Kithkin Soldier):
// "First strike / Reinforce 1—{1}{W} ({1}{W}, Discard this card: Put a +1/+1
// counter on target creature.)" Scryfall rulings: none.
// Earthbrawn, oracle 91edf44a-0f46-4936-909a-934e2c4fe176, revision
// 478208df-9776-4354-b8a9-8da16ae1f50a ({1}{G} Instant): "Target creature gets
// +3/+3 until end of turn. / Reinforce 1—{1}{G} ({1}{G}, Discard this card: Put
// a +1/+1 counter on target creature.)" Scryfall rulings: none.
// Wren's Run Hydra, oracle a2196bc3-2fe6-49d4-8c79-d6beb3d6c5d8, revision
// 5198ac65-118c-4616-8315-d71d41b883ad ({X}{G}, 0/0 Creature — Hydra): "Reach /
// This creature enters with X +1/+1 counters on it. / Reinforce X—{X}{G}{G}
// ({X}{G}{G}, Discard this card: Put X +1/+1 counters on target creature.)"
// Scryfall rulings: none.
// Rustic Clachan, oracle cde68428-0033-4ede-92f1-ab91de0a41fb, revision
// f50160e9-0b25-4e81-814d-cdcda2fb325d (Land; white color identity): "As this
// land enters, you may reveal a Kithkin card from your hand. If you don't, this
// land enters tapped. / {T}: Add {W}. / Reinforce 1—{1}{W} ({1}{W}, Discard this
// card: Put a +1/+1 counter on target creature.)" Scryfall rulings: none.
// Bannerhide Krushok, oracle e3191093-1aaa-4002-a603-7c8d132b6a31, revision
// 1271251b-7d79-4cb4-80bb-98574aa63249 ({3}{G}, 4/4 Creature — Beast):
// "Trample / Reinforce 2—{1}{G} ({1}{G}, Discard this card: Put two +1/+1
// counters on target creature.) / Scavenge {5}{G}{G} ({5}{G}{G}, Exile this
// card from your graveyard: Put a number of +1/+1 counters equal to this card's
// power on target creature. Scavenge only as a sorcery.)"
// Its two 2021-06-18 Scryfall rulings at
// https://api.scryfall.com/cards/1271251b-7d79-4cb4-80bb-98574aa63249/rulings
// state that the exiled card is part of Scavenge's activation cost and that
// opponents may respond to a Reinforce activation before a later Scavenge
// activation if the stack is not empty.
//
// Rules source reviewed before implementation: Wizards Comprehensive Rules,
// effective 2026-11-06, verified local copy .cardwork/MagicCompRules_20261106.txt.
// Sections 107.3a and 601.2b require X to be announced; 111.4 supplies the
// "Elf Warrior Token" name; 118.8 and 602.2b make Reinforce's discard part of
// its activation cost; 601.2b-c require the modal choice and its appropriate
// target; 601.2f-g require total-cost determination and allow mana abilities
// before payment; 602.2a and 701.20a govern hand activation and revealing;
// 608.2b governs resolution target legality; 614.1 governs the land's entry
// replacement; 702.7a, 702.9a, 702.17a, and 702.19a describe first strike,
// flying, reach, and trample; 702.77a defines hand-only Reinforce; 702.97a
// defines graveyard-only, sorcery-speed Scavenge with exile as a cost; and
// 122.1 defines counters. These require the actual zone-limited costs, X values,
// targets, tokens, keywords, reveal choice, and counter results exercised below.
const (
	burrentonBombardierTestOracleID = "13ec2390-e08a-4645-99e9-bac7e41917bb"
	breakTiesTestOracleID           = "1e44bb4e-aeea-4f43-9cbd-4c24a9b6b022"
	fowlStrikeTestOracleID          = "390320f7-11b8-4b61-8316-36c396470a09"
	swellOfCourageTestOracleID      = "60033841-71c7-4370-a2bc-e570a5c100ea"
	huntingTriadTestOracleID        = "66bf0c4b-0c36-4d63-a929-5a7427a07279"
	mosquitoGuardTestOracleID       = "6beaeec2-b4e9-4904-b817-c0118549d4a9"
	earthbrawnTestOracleID          = "91edf44a-0f46-4936-909a-934e2c4fe176"
	wrensRunHydraTestOracleID       = "a2196bc3-2fe6-49d4-8c79-d6beb3d6c5d8"
	rusticClachanTestOracleID       = "cde68428-0033-4ede-92f1-ab91de0a41fb"
	bannerhideKrushokTestOracleID   = "e3191093-1aaa-4002-a603-7c8d132b6a31"
)

func reinforceCardDef(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default-registry lookup for %s (%s): %v", name, oracleID, err)
	}
	if def.Name != name {
		t.Fatalf("oracle %s resolved to %q, want %q", oracleID, def.Name, name)
	}
	return def
}

func reinforceTestCreature(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func reinforceScenario(t *testing.T, oracleID, name string, x, counterCount, manaValue int) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := reinforceCardDef(t, oracleID, name)
	cardID := addCardToHand(g, game.Player1, def)
	target := addCombatPermanent(g, game.Player1, reinforceTestCreature("Reinforce target"))
	for _, c := range []mana.Color{mana.C, mana.W, mana.G} {
		g.Players[game.Player1].ManaPool.Add(c, 10)
	}
	setMainPhasePriority(g, game.Player1)

	activation := action.ActivateAbility(cardID, def.ActivatedAbilityIndex(0), []game.Target{game.PermanentTarget(target.ObjectID)}, x)
	if !containsAction(engine.legalActions(g, game.Player1), activation) {
		t.Fatalf("%s reinforce activation is not legal with X=%d", name, x)
	}
	if !engine.applyAction(g, game.Player1, activation) {
		t.Fatalf("%s reinforce activation failed", name)
	}
	if g.Players[game.Player1].Hand.Contains(cardID) || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("%s was not discarded from hand to pay Reinforce", name)
	}
	if got, want := g.Players[game.Player1].ManaPool.Total(), 30-manaValue; got != want {
		t.Fatalf("%s mana remaining = %d, want %d after paying Reinforce", name, got, want)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != counterCount {
		t.Fatalf("%s put %d +1/+1 counters on target, want %d", name, got, counterCount)
	}
}

func castReinforceTestSpell(t *testing.T, g *game.Game, engine *Engine, oracleID, name string, targets []game.Target, modes []int, x int) id.ID {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, reinforceCardDef(t, oracleID, name))
	setMainPhasePriority(g, game.Player1)
	cast := action.CastSpell(cardID, targets, x, modes)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatalf("%s spell cast with targets %v and modes %v is not legal", name, targets, modes)
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatalf("casting %s failed", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	return cardID
}

// Source review: Burrenton Bombardier's exact text, zero-rulings result, and
// hand-only Reinforce cost are recorded above; CR 702.77a and 602.2a apply.
func TestBurrentonBombardierReinforceFromHand(t *testing.T) {
	reinforceScenario(t, burrentonBombardierTestOracleID, "Burrenton Bombardier", 0, 2, 3)
}

func TestBurrentonBombardierHasFlying(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, reinforceCardDef(t, burrentonBombardierTestOracleID, "Burrenton Bombardier"))
	if !hasKeyword(g, permanent, game.Flying) {
		t.Fatal("Burrenton Bombardier does not have flying")
	}
}

// Source review: Break Ties' exact three modes, Reinforce text, and zero rulings
// are recorded above; CR 601.2c requires selecting one mode and its target.
func TestBreakTiesReinforceFromHand(t *testing.T) {
	reinforceScenario(t, breakTiesTestOracleID, "Break Ties", 0, 1, 1)
}

func TestBreakTiesSpellModes(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode int
	}{
		{name: "destroy artifact", mode: 0},
		{name: "destroy enchantment", mode: 1},
		{name: "exile graveyard card", mode: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			var target game.Target
			var targetID id.ID
			switch tc.mode {
			case 0:
				permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Artifact", Types: []types.Card{types.Artifact}}})
				targetID, target = permanent.CardInstanceID, game.PermanentTarget(permanent.ObjectID)
			case 1:
				permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Oath", Types: []types.Card{types.Enchantment}}})
				targetID, target = permanent.CardInstanceID, game.PermanentTarget(permanent.ObjectID)
			case 2:
				targetID = addCardToHand(g, game.Player2, reinforceTestCreature("Graveyard card"))
				g.Players[game.Player2].Hand.Remove(targetID)
				g.Players[game.Player2].Graveyard.Add(targetID)
				target = currentCardTarget(t, g, targetID)
			}
			castReinforceTestSpell(t, g, engine, breakTiesTestOracleID, "Break Ties", []game.Target{target}, []int{tc.mode}, 0)
			if tc.mode == 2 {
				if !g.Players[game.Player2].Exile.Contains(targetID) {
					t.Fatal("selected graveyard card was not exiled")
				}
			} else if !g.Players[game.Player2].Graveyard.Contains(targetID) {
				t.Fatalf("selected %s was not destroyed", tc.name)
			}
		})
	}
}

// Source review: Fowl Strike's exact text and zero-rulings result are recorded
// above; CR 702.9a defines flying, which is the spell's target restriction.
func TestFowlStrikeReinforceFromHand(t *testing.T) {
	reinforceScenario(t, fowlStrikeTestOracleID, "Fowl Strike", 0, 2, 3)
}

func TestFowlStrikeDestroysFlyingCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	flyerDef := reinforceTestCreature("Flying creature")
	flyerDef.StaticAbilities = []game.StaticAbility{game.FlyingStaticBody}
	flyer := addCombatPermanent(g, game.Player2, flyerDef)
	nonflyer := addCombatPermanent(g, game.Player2, reinforceTestCreature("Ground creature"))
	cardID := addCardToHand(g, game.Player1, reinforceCardDef(t, fowlStrikeTestOracleID, "Fowl Strike"))
	setMainPhasePriority(g, game.Player1)
	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(cardID, []game.Target{game.PermanentTarget(nonflyer.ObjectID)}, 0, nil)) {
		t.Fatal("Fowl Strike offered a nonflying creature as a legal target")
	}
	cast := action.CastSpell(cardID, []game.Target{game.PermanentTarget(flyer.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) || !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Fowl Strike could not be cast targeting the flying creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(flyer.CardInstanceID) {
		t.Fatal("Fowl Strike did not destroy the flying creature")
	}
}

// Source review: Swell of Courage's exact spell and X-Reinforce text and zero
// rulings are recorded above; CR 107.3a governs the announced X value.
func TestSwellOfCourageReinforceFromHand(t *testing.T) {
	reinforceScenario(t, swellOfCourageTestOracleID, "Swell of Courage", 3, 3, 5)
}

func TestSwellOfCourageBuffsYourCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.W, 5)
	friendly := addCombatPermanent(g, game.Player1, reinforceTestCreature("Friendly creature"))
	opponent := addCombatPermanent(g, game.Player2, reinforceTestCreature("Opponent creature"))
	castReinforceTestSpell(t, g, engine, swellOfCourageTestOracleID, "Swell of Courage", nil, nil, 0)
	if got := effectivePower(g, friendly); got != 4 {
		t.Fatalf("friendly creature power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, friendly); !ok || got != 4 {
		t.Fatalf("friendly creature toughness = %d (defined=%v), want 4", got, ok)
	}
	if got := effectivePower(g, opponent); got != 2 {
		t.Fatalf("opponent creature power = %d, want unchanged 2", got)
	}
}

// Source review: Hunting Triad's exact text and its three Kindred rulings are
// recorded above; CR 111.4 determines the token name from Elf Warrior subtypes.
func TestHuntingTriadReinforceFromHand(t *testing.T) {
	reinforceScenario(t, huntingTriadTestOracleID, "Hunting Triad", 0, 3, 4)
}

func TestHuntingTriadCreatesElfWarriorTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	castReinforceTestSpell(t, g, engine, huntingTriadTestOracleID, "Hunting Triad", nil, nil, 0)
	if got := countTokenPermanentsNamed(g, "Elf Warrior Token"); got != 3 {
		t.Fatalf("Elf Warrior Token count = %d, want 3", got)
	}
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanentTokenName(permanent) != "Elf Warrior Token" {
			continue
		}
		if !permanentHasType(g, permanent, types.Creature) || !permanentHasSubtype(g, permanent, types.Elf) || !permanentHasSubtype(g, permanent, types.Warrior) {
			t.Fatalf("token %q lacks creature, Elf, or Warrior types", permanentTokenName(permanent))
		}
		if got := effectivePower(g, permanent); got != 1 {
			t.Fatalf("Elf Warrior token power = %d, want 1", got)
		}
		if got, ok := effectiveToughness(g, permanent); !ok || got != 1 {
			t.Fatalf("Elf Warrior token toughness = %d (defined=%v), want 1", got, ok)
		}
		if permanent.TokenDef == nil || len(permanent.TokenDef.Colors) != 1 || permanent.TokenDef.Colors[0] != color.Green {
			t.Fatal("Elf Warrior token is not green")
		}
	}
}

// Source review: Mosquito Guard's exact First strike and Reinforce text and
// zero rulings are recorded above; CR 702.7a defines its combat keyword.
func TestMosquitoGuardReinforceFromHand(t *testing.T) {
	reinforceScenario(t, mosquitoGuardTestOracleID, "Mosquito Guard", 0, 1, 2)
}

func TestMosquitoGuardHasFirstStrike(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, reinforceCardDef(t, mosquitoGuardTestOracleID, "Mosquito Guard"))
	if !hasKeyword(g, permanent, game.FirstStrike) {
		t.Fatal("Mosquito Guard does not have first strike")
	}
}

// Source review: Earthbrawn's exact text and zero-rulings result are recorded
// above; CR 601.2c and 608.2b govern its target and resolution.
func TestEarthbrawnReinforceFromHand(t *testing.T) {
	reinforceScenario(t, earthbrawnTestOracleID, "Earthbrawn", 0, 1, 2)
}

func TestEarthbrawnTargetsCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	target := addCombatPermanent(g, game.Player2, reinforceTestCreature("Earthbrawn target"))
	castReinforceTestSpell(t, g, engine, earthbrawnTestOracleID, "Earthbrawn", []game.Target{game.PermanentTarget(target.ObjectID)}, nil, 0)
	if got := effectivePower(g, target); got != 5 {
		t.Fatalf("target creature power = %d, want 5 after +3/+3", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 5 {
		t.Fatalf("target creature toughness = %d (defined=%v), want 5 after +3/+3", got, ok)
	}
}

// Source review: Wren's Run Hydra's exact text and zero-rulings result are
// recorded above; CR 107.3a defines cast X and 614.1 its counter-entry event.
func TestWrensRunHydraReinforceFromHand(t *testing.T) {
	reinforceScenario(t, wrensRunHydraTestOracleID, "Wren's Run Hydra", 3, 3, 5)
}

func TestWrensRunHydraEntersWithCastX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.G, 5)
	cardID := addCardToHand(g, game.Player1, reinforceCardDef(t, wrensRunHydraTestOracleID, "Wren's Run Hydra"))
	setMainPhasePriority(g, game.Player1)
	cast := action.CastSpell(cardID, nil, 4, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) || !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Wren's Run Hydra could not be cast with X=4")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, cardID)
	if permanent == nil {
		t.Fatal("Wren's Run Hydra did not enter the battlefield")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("Wren's Run Hydra entered with %d +1/+1 counters, want 4", got)
	}
	if !hasKeyword(g, permanent, game.Reach) {
		t.Fatal("Wren's Run Hydra does not have reach")
	}
}

// Source review: Rustic Clachan's exact text and zero-rulings result are
// recorded above; CR 701.20a governs revealing a matching Kithkin, while 614.1
// governs the alternative tapped-entry event.
func TestRusticClachanReinforceFromHand(t *testing.T) {
	reinforceScenario(t, rusticClachanTestOracleID, "Rustic Clachan", 0, 1, 2)
}

func TestRusticClachanRevealOrEntersTappedAndProducesWhite(t *testing.T) {
	for _, revealKithkin := range []bool{true, false} {
		name := "without Kithkin enters tapped"
		if revealKithkin {
			name = "revealing Kithkin enters untapped and pays Reinforce"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := reinforceCardDef(t, rusticClachanTestOracleID, "Rustic Clachan")
			var kithkinID id.ID
			if revealKithkin {
				kithkinID = addCardToHand(g, game.Player1, reinforceCardDef(t, burrentonBombardierTestOracleID, "Burrenton Bombardier"))
			}
			landID := addCardToHand(g, game.Player1, def)
			setMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
				t.Fatal("Rustic Clachan could not be played")
			}
			land := permanentForCard(g, landID)
			if land == nil {
				t.Fatal("Rustic Clachan did not enter the battlefield")
			}
			if land.Tapped != !revealKithkin {
				t.Fatalf("Rustic Clachan tapped = %v, want %v", land.Tapped, !revealKithkin)
			}
			if revealKithkin {
				if !g.Players[game.Player1].Hand.Contains(kithkinID) || !eventRevealedCardFromZone(g, game.Player1, landID, kithkinID, zone.Hand) {
					t.Fatal("Rustic Clachan did not reveal the Kithkin while leaving it in hand")
				}
				creature := addCombatPermanent(g, game.Player1, reinforceTestCreature("Clachan target"))
				reinforceID := addCardToHand(g, game.Player1, reinforceCardDef(t, breakTiesTestOracleID, "Break Ties"))
				setMainPhasePriority(g, game.Player1)
				activation := action.ActivateAbility(reinforceID, reinforceCardDef(t, breakTiesTestOracleID, "Break Ties").ActivatedAbilityIndex(0), []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)
				if !containsAction(engine.legalActions(g, game.Player1), activation) || !engine.applyAction(g, game.Player1, activation) {
					t.Fatal("Rustic Clachan did not produce white mana to pay Reinforce")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if !land.Tapped || creature.Counters.Get(counter.PlusOnePlusOne) != 1 {
					t.Fatalf("Clachan tapped=%v, target counters=%d; want tapped and one counter", land.Tapped, creature.Counters.Get(counter.PlusOnePlusOne))
				}
			}
		})
	}
}

// Source review: Bannerhide Krushok's exact text and its two 2021-06-18 rulings
// are recorded above; CR 702.97a defines Scavenge's sorcery timing, target, and
// exile-as-cost behavior.
func TestBannerhideKrushokReinforceFromHand(t *testing.T) {
	reinforceScenario(t, bannerhideKrushokTestOracleID, "Bannerhide Krushok", 0, 2, 2)
}

func TestBannerhideKrushokScavengesFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, reinforceCardDef(t, bannerhideKrushokTestOracleID, "Bannerhide Krushok"))
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	target := addCombatPermanent(g, game.Player1, reinforceTestCreature("Scavenge target"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	setMainPhasePriority(g, game.Player1)
	activation := action.ActivateAbility(cardID, reinforceCardDef(t, bannerhideKrushokTestOracleID, "Bannerhide Krushok").ActivatedAbilityIndex(1), []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("Bannerhide Krushok Scavenge was not legal during the main phase")
	}
	if !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("Bannerhide Krushok Scavenge activation failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("Bannerhide Krushok was not exiled to pay Scavenge")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("Scavenge put %d counters on the target, want source power 4", got)
	}
}

func TestBannerhideKrushokHasTrample(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, reinforceCardDef(t, bannerhideKrushokTestOracleID, "Bannerhide Krushok"))
	if !hasKeyword(g, permanent, game.Trample) {
		t.Fatal("Bannerhide Krushok does not have trample")
	}
}
