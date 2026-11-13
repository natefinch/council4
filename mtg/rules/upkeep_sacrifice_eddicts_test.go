package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Defiler of Souls Oracle ID 34894b92-7ebd-49be-9a67-7dc6428aaab9, printing
// 16e8b418-04af-4b99-b8fd-deb1fe21a61c, revision 16e8b418-04af-4b99-b8fd-deb1fe21a61c:
// "Flying / At the beginning of each player's upkeep, that player sacrifices a monocolored creature of their choice." The current
// Scryfall rulings URI (https://api.scryfall.com/cards/16e8b418-04af-4b99-b8fd-deb1fe21a61c/rulings) returned no rulings on 2026-11-13.
// Read the official WotC Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt): 105.1-2a (monocolored means exactly one of five colors),
// 503.1/1a and 603.2b/3a (upkeep triggers and controller), 608.2c/d (ordered effect and resolution-time choice), 609.3 and 701.21a
// (impossible actions and who can sacrifice). These require the active upkeep player—not Defiler's controller—to choose one eligible creature.
func TestDefilerOfSoulsEachPlayerUpkeepSacrificesMonocoloredCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	defiler := castUpkeepEdictCard(t, g, engine, "34894b92-7ebd-49be-9a67-7dc6428aaab9", 3, mana.B, mana.B, mana.R)
	if !hasKeyword(g, defiler, game.Flying) {
		t.Fatal("Defiler of Souls does not have flying")
	}

	p1Mono := addEdictCreature(g, game.Player1, "Player 1 mono", color.Red)
	p1Multi := addEdictCreature(g, game.Player1, "Player 1 multicolor", color.Blue, color.Green)
	p2MonoA := addEdictCreature(g, game.Player2, "Player 2 mono A", color.Blue)
	p2MonoB := addEdictCreature(g, game.Player2, "Player 2 mono B", color.Black)
	p2Multi := addEdictCreature(g, game.Player2, "Player 2 multicolor", color.Red, color.Green)
	p2Colorless := addEdictCreature(g, game.Player2, "Player 2 colorless")

	agents := simPassPolicies()
	agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{1}}}
	resolveUpkeepSacrificeTrigger(t, g, engine, game.Player2, agents)
	if containsEdictPermanent(g, p2MonoB) || !containsEdictPermanent(g, p2MonoA) ||
		!containsEdictPermanent(g, p2Multi) || !containsEdictPermanent(g, p2Colorless) {
		t.Fatal("Player 2 did not choose exactly one of their monocolored creatures")
	}

	resolveUpkeepSacrificeTrigger(t, g, engine, game.Player1, simPassPolicies())
	if containsEdictPermanent(g, p1Mono) || !containsEdictPermanent(g, p1Multi) {
		t.Fatal("Player 1's upkeep did not sacrifice only their monocolored creature")
	}
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Woebringer Demon Oracle ID 21ea9501-9a0f-43a3-a2cc-e9648325c45e, printing
// 88b86280-f7a5-4d4d-bc00-71f21590835e, revision 88b86280-f7a5-4d4d-bc00-71f21590835e:
// "Flying / At the beginning of each player's upkeep, that player sacrifices a creature of their choice. If the player can't, sacrifice this creature."
// The current Scryfall rulings URI (https://api.scryfall.com/cards/88b86280-f7a5-4d4d-bc00-71f21590835e/rulings) returned no rulings on 2026-11-13.
// Read the official WotC Comprehensive Rules TXT, effective 2026-11-06
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt): 113.7a and 603.3a (trigger independence and controller), 503.1/1a and
// 603.2b (upkeep timing), 608.2c/d and 609.3 (ordered choices and impossible actions), and 701.21a (the player sacrificing must control the permanent).
// If the upkeep player has no creature to sacrifice, the second instruction sacrifices the exact source creature.
func TestWoebringerDemonUpkeepSacrificeAndSelfSacrificeFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	woebringer := castUpkeepEdictCard(t, g, engine, "21ea9501-9a0f-43a3-a2cc-e9648325c45e", 3, mana.B, mana.B)
	if !hasKeyword(g, woebringer, game.Flying) {
		t.Fatal("Woebringer Demon does not have flying")
	}

	p2CreatureA := addEdictCreature(g, game.Player2, "Player 2 creature A")
	p2CreatureB := addEdictCreature(g, game.Player2, "Player 2 creature B")
	p2Artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Player 2 artifact", Types: []types.Card{types.Artifact},
	}})
	p3Artifact := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Player 3 artifact", Types: []types.Card{types.Artifact},
	}})

	agents := simPassPolicies()
	agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{1}}}
	resolveUpkeepSacrificeTrigger(t, g, engine, game.Player2, agents)
	if containsEdictPermanent(g, p2CreatureB) || !containsEdictPermanent(g, p2CreatureA) ||
		!containsEdictPermanent(g, p2Artifact) || !containsEdictPermanent(g, woebringer) {
		t.Fatal("Player 2 did not choose exactly one creature to sacrifice")
	}

	resolveUpkeepSacrificeTrigger(t, g, engine, game.Player3, simPassPolicies())
	if containsEdictPermanent(g, woebringer) {
		t.Fatal("Woebringer Demon was not sacrificed when Player 3 had no creature")
	}
	if !containsEdictPermanent(g, p3Artifact) {
		t.Fatal("Woebringer Demon fallback sacrificed a noncreature permanent")
	}
}

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Wasitora, Nekoru Queen Oracle ID 13fc4168-fcc2-4011-804f-211d5d86b7dd,
// printing 07790bd5-0d27-490e-b1ae-3b866882461e, revision 07790bd5-0d27-490e-b1ae-3b866882461e:
// "Flying, trample / Whenever Wasitora deals combat damage to a player, that player sacrifices a creature of their choice. If the player can't, you create a 3/3 black, red, and green Cat Dragon creature token with flying."
// Its current Scryfall rulings URI (https://api.scryfall.com/cards/07790bd5-0d27-490e-b1ae-3b866882461e/rulings) returned one WotC ruling (2017-08-25):
// a creature dealt lethal combat damage simultaneously is destroyed before the damaged player chooses a creature. Read official WotC Comprehensive Rules TXT,
// effective 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt): 111.2-.3 (token owner/controller and characteristics), 117.5/704.3/704.5g
// (state-based actions before triggers), 510.2 (simultaneous combat damage), 603.2/3a (trigger event and controller), 608.2c/d, 609.3, and 701.21a.
// These require a combat-damage-only trigger, the damaged player to choose, and the source controller to create the full Cat Dragon token only if that sacrifice cannot happen.
func TestWasitoraCombatDamageSacrificeAndCatDragonFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	wasitora := castUpkeepEdictCard(t, g, engine, "13fc4168-fcc2-4011-804f-211d5d86b7dd", 2, mana.B, mana.R, mana.G)
	if !hasKeyword(g, wasitora, game.Flying) || !hasKeyword(g, wasitora, game.Trample) {
		t.Fatal("Wasitora does not have flying and trample")
	}

	p2CreatureA := addEdictCreature(g, game.Player2, "Player 2 creature A")
	p2CreatureB := addEdictCreature(g, game.Player2, "Player 2 creature B")
	p3Artifact := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Player 3 artifact", Types: []types.Card{types.Artifact},
	}})

	if got := dealPlayerDamage(g, wasitora.CardInstanceID, wasitora.ObjectID, game.Player1, game.Player3, 1, false); got != 1 {
		t.Fatalf("noncombat damage dealt = %d, want 1", got)
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Wasitora triggered from noncombat damage")
	}

	if got := dealPlayerDamage(g, wasitora.CardInstanceID, wasitora.ObjectID, game.Player1, game.Player2, 5, true); got != 5 {
		t.Fatalf("combat damage dealt = %d, want 5", got)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Wasitora did not trigger from combat damage to Player 2")
	}
	agents := simPassPolicies()
	agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{1}}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if containsEdictPermanent(g, p2CreatureB) || !containsEdictPermanent(g, p2CreatureA) {
		t.Fatal("Player 2 did not choose exactly one creature to sacrifice")
	}
	if got := countTokenNamed(g, "Cat Dragon", game.Player1); got != 0 {
		t.Fatalf("Cat Dragon tokens after successful sacrifice = %d, want 0", got)
	}

	if got := dealPlayerDamage(g, wasitora.CardInstanceID, wasitora.ObjectID, game.Player1, game.Player3, 5, true); got != 5 {
		t.Fatalf("combat damage dealt to Player 3 = %d, want 5", got)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Wasitora did not trigger from combat damage to Player 3")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if !containsEdictPermanent(g, p3Artifact) {
		t.Fatal("Wasitora's fallback affected a noncreature permanent")
	}
	if got := countTokenNamed(g, "Cat Dragon", game.Player1); got != 1 {
		t.Fatalf("Cat Dragon tokens after failed sacrifice = %d, want 1", got)
	}
	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentEffectiveName(g, permanent) == "Cat Dragon" {
			token = permanent
		}
	}
	if token == nil || token.Owner != game.Player1 || token.Controller != game.Player1 ||
		!permanentHasType(g, token, types.Creature) || !permanentHasSubtype(g, token, types.Cat) ||
		!permanentHasSubtype(g, token, types.Dragon) || !hasKeyword(g, token, game.Flying) ||
		effectivePower(g, token) != 3 {
		t.Fatal("Wasitora did not create its 3/3 black, red, and green flying Cat Dragon token under its controller")
	}
	if toughness, ok := effectiveToughness(g, token); !ok || toughness != 3 {
		t.Fatalf("Cat Dragon toughness = %d (available %t), want 3", toughness, ok)
	}
	colors := permanentEffectiveColors(g, token)
	if len(colors) != 3 || !slices.Contains(colors, color.Black) || !slices.Contains(colors, color.Red) || !slices.Contains(colors, color.Green) {
		t.Fatalf("Cat Dragon colors = %v, want black, red, and green", colors)
	}
}

func castUpkeepEdictCard(t *testing.T, g *game.Game, engine *Engine, oracleID string, generic int, colors ...mana.Color) *game.Permanent {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default-registry lookup for %s failed: %v", oracleID, err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	if generic > 0 {
		g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	}
	for _, c := range colors {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", def.Name)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("casting %s left %d mana, want exact payment", def.Name, got)
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	t.Fatalf("%s did not enter the battlefield", def.Name)
	return nil
}

func addEdictCreature(g *game.Game, controller game.PlayerID, name string, colors ...color.Color) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature}, Colors: colors,
	}})
}

func resolveUpkeepSacrificeTrigger(t *testing.T, g *game.Game, engine *Engine, activePlayer game.PlayerID, agents [game.NumPlayers]PlayerAgent) {
	t.Helper()
	g.Turn.ActivePlayer = activePlayer
	g.Turn.Step = game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatalf("upkeep trigger did not go on the stack for player %v", activePlayer)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

func containsEdictPermanent(g *game.Game, want *game.Permanent) bool {
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == want.ObjectID {
			return true
		}
	}
	return false
}
