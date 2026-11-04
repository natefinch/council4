package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3648 (sub-issue of #3596), a Siege's intrinsic defeat
// trigger when damage removes its last defense counter. Official CR effective
// 2026-09-25 (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// verified local copy), read 120.3h, 310.6, 310.7, 310.12b, 603.2, 603.3,
// 704.5v, 704.5w, 707.10, 712.11a. 120.3h/310.6: damage dealt to a battle
// removes that many defense counters from it; 310.12b: each Siege has "When the
// last defense counter is removed from this permanent, exile it, then you may
// cast it transformed without paying its mana cost."; 704.5v: a Siege with
// defense 0 that isn't the source of an ability that has triggered but not yet
// left the stack is put into its owner's graveyard; 707.10: a copy of an ability
// on the stack copies it entirely; 712.11a: a transforming double-faced card
// cast transformed is put on the stack back face up.
// Cards (pinned corpus): Invasion of Segovia // Caetus, Sea Tyrant of Segovia
// (oracle 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, {2}{U} Battle — Siege, defense
// 4; back Legendary Creature — Serpent 3/3); Invasion of Kylem // Valor's Reach
// Tag Team (oracle fee0aa18-5103-4814-94e3-87ad3eaf5a6f, {2}{R}{W} Battle —
// Siege, defense 5; back Sorcery "Create two 3/2 red and white Warrior creature
// tokens ..."); Grizzly Bears (14c8f55d-d177-4c25-a931-ebeb9e6062a0, 2/2);
// Lightning Bolt (4457ed35-7c10-48c8-9776-456485fdf070, "Lightning Bolt deals 3
// damage to any target."); Stifle (b3b00911-ece7-4484-bc36-f211ce72b6cc,
// "Counter target activated or triggered ability.").
// Scryfall rulings 2023-04-14 (Invasion of Segovia, Invasion of Kylem): the
// ability triggers when the last defense counter is removed, so a Siege put into
// a graveyard by the zero-defense state-based action doesn't trigger it; if the
// ability is countered, the battle is put into its owner's graveyard; a Siege
// that isn't a transforming double-faced card is exiled and can't be cast,
// remaining in exile. Removal by effects, costs or moves is #3649.

func siegeDefeatAbilityOnTop(t *testing.T, g *game.Game, siege *game.Permanent) *game.StackObject {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || !obj.SiegeDefeat ||
		obj.SourceID != siege.ObjectID || obj.Controller != siege.Controller {
		t.Fatalf("stack top = %+v, want the Siege's defeat ability controlled by %v", obj, siege.Controller)
	}
	return obj
}

func siegeDefeatAbilities(g *game.Game) int {
	count := 0
	for _, obj := range g.Stack.Objects() {
		if obj.SiegeDefeat {
			count++
		}
	}
	return count
}

func lastCountersRemovedEvent(t *testing.T, g *game.Game) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventCountersRemoved {
			return g.Events[i]
		}
	}
	t.Fatal("no counters-removed event")
	return game.Event{}
}

// combatDefeatedSegovia has Player1's Grizzly Bears deal combat damage to
// Player1's Invasion of Segovia, protected by Player2, removing its last two
// defense counters. It returns before state-based actions.
func combatDefeatedSegovia(t *testing.T, defense int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g, engine, _, siege := siegeUnderAttack(t, game.Player1, game.Player2)
	siege.Counters.Remove(counter.Defense, 4-defense)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	return g, engine, siege
}

// TestSiegeCombatDefeatCastsCaetusTransformed is acceptance scenario 1.
func TestSiegeCombatDefeatCastsCaetusTransformed(t *testing.T) {
	g, engine, siege := combatDefeatedSegovia(t, 2)
	cardID := siege.CardInstanceID
	if defense := siege.Counters.Get(counter.Defense); defense != 0 {
		t.Fatalf("Siege defense = %d, want 0", defense)
	}
	removed := lastCountersRemovedEvent(t, g)
	if removed.PermanentID != siege.ObjectID || removed.CounterKind != counter.Defense ||
		removed.PreviousCounterAmount != 2 || removed.Amount != 2 || removed.Controller != game.Player1 {
		t.Fatalf("counters removed event = %+v, want 2 of 2 defense counters removed by Player1's Bears", removed)
	}

	// The ability has triggered but isn't on the stack yet (CR 704.5v).
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
		t.Fatal("Siege was put into the graveyard while its defeat ability was waiting to be put on the stack")
	}
	agents := [game.NumPlayers]PlayerAgent{}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	obj := siegeDefeatAbilityOnTop(t, g, siege)
	if siegeDefeatAbilities(g) != 1 {
		t.Fatalf("stack = %+v, want exactly one defeat ability", g.Stack.Objects())
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
		t.Fatal("Siege was put into the graveyard while its defeat ability was on the stack")
	}
	views := NewObservation(g, game.Player2).Stack()
	if top := views[len(views)-1]; top.ID != obj.ID || !strings.HasPrefix(top.Text, "When the last defense counter is removed") {
		t.Fatalf("opponent stack view = %+v, want the Siege ability's text", top)
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if _, ok := permanentByObjectID(g, siege.ObjectID); ok {
		t.Fatal("Siege is still on the battlefield after its defeat ability resolved")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID || spell.Face != game.FaceBack || spell.Controller != game.Player1 {
		t.Fatalf("stack top = %+v, want Caetus cast transformed by Player1", spell)
	}
	if spell.ManaSpentToCast != 0 {
		t.Fatalf("mana spent = %d, want a free cast", spell.ManaSpentToCast)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	caetus := permanentForCard(g, cardID)
	if caetus == nil || caetus.Face != game.FaceBack || !caetus.Transformed || caetus.Controller != game.Player1 || !caetus.EnteredFromCast {
		t.Fatalf("Caetus = %+v, want it cast and transformed under Player1's control", caetus)
	}
	if !g.Stack.IsEmpty() {
		t.Fatalf("stack = %+v, want empty", g.Stack.Objects())
	}
}

// kylemFacingBolt puts Player1's Invasion of Kylem, protected by Player2, on the
// battlefield with the given defense counters and has Player1 cast Lightning
// Bolt at it. It returns after the Bolt resolves, state-based actions are
// checked and triggered abilities are put on the stack.
func kylemFacingBolt(t *testing.T, defense int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	return boltedSiege(t, registryCardDef(t, "Invasion of Kylem"), defense)
}

func boltedSiege(t *testing.T, def *game.CardDef, defense int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	engine := choicesProgramEngine(t)
	siege := addCombatPermanent(g, game.Player1, def)
	siege.Counters.Add(counter.Defense, defense)
	siege.BattleProtector = opt.Val(game.Player2)
	boltID := addCardToHand(g, game.Player1, registryCard(t, "Lightning Bolt"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(boltID, []game.Target{game.PermanentTarget(siege.ObjectID)}, 0, nil)) {
		t.Fatalf("Lightning Bolt could not target %s", def.Name)
	}
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(boltID) {
		t.Fatal("Lightning Bolt did not resolve")
	}
	engine.applyStateBasedActions(g)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	return g, engine, siege
}

func declineFreeCast(offered *[]string) *controlScript {
	return &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if !strings.HasPrefix(request.Prompt, "Choose how to cast") {
			return fallbackChoice(request)
		}
		for _, option := range request.Options {
			*offered = append(*offered, option.Label)
		}
		return []int{len(request.Options) - 1}
	}}
}

// TestSiegeBoltDefeatDeclinedLeavesCardInExile is acceptance scenario 2,
// declining the cast.
func TestSiegeBoltDefeatDeclinedLeavesCardInExile(t *testing.T) {
	g, engine, kylem := kylemFacingBolt(t, 3)
	cardID := kylem.CardInstanceID
	siegeDefeatAbilityOnTop(t, g, kylem)
	var offered []string
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: declineFreeCast(&offered)}, &TurnLog{})
	if len(offered) != 2 || !strings.HasPrefix(offered[0], "Valor's Reach Tag Team") || !strings.HasPrefix(offered[1], "Don't cast") {
		t.Fatalf("offered = %q, want the back face and declining", offered)
	}
	if !g.Stack.IsEmpty() || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatalf("stack = %+v, exile = %v; want the declined card left in exile", g.Stack.Objects(), g.Players[game.Player1].Exile.Contains(cardID))
	}
	if permanentForCard(g, cardID) != nil || warriorTokens(g, game.Player1) != 0 {
		t.Fatal("declining still left Kylem on the battlefield or created Warriors")
	}
}

// TestSiegeBoltDefeatCastsTagTeam is acceptance scenario 2, accepting the cast.
func TestSiegeBoltDefeatCastsTagTeam(t *testing.T) {
	g, engine, kylem := kylemFacingBolt(t, 3)
	cardID := kylem.CardInstanceID
	siegeDefeatAbilityOnTop(t, g, kylem)
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID || spell.Face != game.FaceBack {
		t.Fatalf("stack top = %+v, want Valor's Reach Tag Team cast transformed", spell)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := warriorTokens(g, game.Player1); got != 2 {
		t.Fatalf("Warrior tokens = %d, want 2", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Invasion of Kylem is not in its owner's graveyard after Valor's Reach Tag Team resolved")
	}
}

// TestStifledSiegeDefeatPutsBattleIntoGraveyard is acceptance scenario 3.
func TestStifledSiegeDefeatPutsBattleIntoGraveyard(t *testing.T) {
	g, engine, kylem := kylemFacingBolt(t, 3)
	cardID := kylem.CardInstanceID
	obj := siegeDefeatAbilityOnTop(t, g, kylem)
	stifleID := addCardToHand(g, game.Player2, registryCard(t, "Stifle"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, action.CastSpell(stifleID, []game.Target{game.StackObjectTarget(obj.ID)}, 0, nil)) {
		t.Fatal("Stifle could not target the Siege's defeat ability")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Stack.IsEmpty() {
		t.Fatalf("stack = %+v, want the defeat ability countered", g.Stack.Objects())
	}
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("countered Siege was not put into its owner's graveyard by state-based actions")
	}
	if warriorTokens(g, game.Player1) != 0 || !g.Stack.IsEmpty() {
		t.Fatal("countered Siege ability still cast the back face")
	}
}

// TestSiegeDamageLeavingDefenseDoesNotTrigger is acceptance scenario 4.
func TestSiegeDamageLeavingDefenseDoesNotTrigger(t *testing.T) {
	g, engine, siege := combatDefeatedSegovia(t, 3)
	if defense := siege.Counters.Get(counter.Defense); defense != 1 {
		t.Fatalf("Siege defense = %d, want 1", defense)
	}
	removed := lastCountersRemovedEvent(t, g)
	if removed.PreviousCounterAmount != 3 || removed.Amount != 2 {
		t.Fatalf("counters removed event = %+v, want 2 of 3", removed)
	}
	engine.applyStateBasedActions(g)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if siegeDefeatAbilities(g) != 0 {
		t.Fatalf("stack = %+v, want no defeat ability", g.Stack.Objects())
	}
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
		t.Fatal("Siege with defense left the battlefield")
	}

	// Bolt for exactly the remaining defense.
	g2, _, kylem := kylemFacingBolt(t, 4)
	if defense := kylem.Counters.Get(counter.Defense); defense != 1 || siegeDefeatAbilities(g2) != 0 {
		t.Fatalf("Kylem defense = %d with %d defeat abilities, want 1 and none", defense, siegeDefeatAbilities(g2))
	}
}

// TestCopiedSiegeDefeatAbilityCastsOnce proves a copy of the ability (CR 707.10)
// keeps the Siege behavior: the first to resolve exiles and casts the card; the
// other finds its source gone and does nothing.
func TestCopiedSiegeDefeatAbilityCastsOnce(t *testing.T) {
	g, engine, kylem := kylemFacingBolt(t, 3)
	cardID := kylem.CardInstanceID
	obj := siegeDefeatAbilityOnTop(t, g, kylem)
	addEffectSpellToStack(g, game.Player2, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || !copied.SiegeDefeat || copied.ID == obj.ID || copied.Controller != game.Player2 {
		t.Fatalf("stack top = %+v, want Player2's copy of the defeat ability", copied)
	}

	// The clone keeps the defeat behavior through replay (fork safety).
	clone := g.Clone()
	engine.resolveTopOfStackWithChoices(clone, agents, &TurnLog{})
	if spell, ok := clone.Stack.Peek(); !ok || spell.Kind != game.StackSpell || spell.Controller != game.Player2 {
		t.Fatalf("clone stack top = %+v, want the card cast by the copy's controller", spell)
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID || spell.Controller != game.Player2 {
		t.Fatalf("stack top = %+v, want the copy's controller casting the card transformed", spell)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := warriorTokens(g, game.Player2); got != 2 {
		t.Fatalf("Player2 Warrior tokens = %d, want 2", got)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !g.Stack.IsEmpty() || warriorTokens(g, game.Player1) != 0 || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("stack = %+v; the original ability acted on a card that already left", g.Stack.Objects())
	}
}

// TestNonTransformingSiegeStaysInExile proves the ruling for a Siege that isn't
// a transforming double-faced card: it is exiled and can't be cast.
func TestNonTransformingSiegeStaysInExile(t *testing.T) {
	g, engine, siege := boltedSiege(t, &game.CardDef{CardFace: game.CardFace{
		Name: "Single-Faced Siege", Types: []types.Card{types.Battle}, Subtypes: []types.Sub{types.Siege}, Defense: opt.Val(3),
	}}, 3)
	siegeDefeatAbilityOnTop(t, g, siege)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Stack.IsEmpty() || !g.Players[game.Player1].Exile.Contains(siege.CardInstanceID) {
		t.Fatalf("stack = %+v; want the single-faced Siege exiled and not cast", g.Stack.Objects())
	}
}

// TestNonSiegeBattleHasNoDefeatAbility proves the intrinsic ability belongs to
// Sieges only (CR 310.12b, 704.5w).
func TestNonSiegeBattleHasNoDefeatAbility(t *testing.T) {
	g, _, battle := boltedSiege(t, &game.CardDef{CardFace: game.CardFace{
		Name: "Plain Battle", Types: []types.Card{types.Battle}, Defense: opt.Val(3),
	}}, 3)
	if siegeDefeatAbilities(g) != 0 || !g.Players[game.Player1].Graveyard.Contains(battle.CardInstanceID) {
		t.Fatalf("stack = %+v; want the battle put into the graveyard without a defeat ability", g.Stack.Objects())
	}
}
