package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

type provokeChoiceAgent struct {
	target  id.ID
	accept  bool
	request []game.ChoiceRequest
}

func (a *provokeChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *provokeChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.request = append(a.request, request)
	switch request.Kind {
	case game.ChoiceTarget:
		for _, option := range request.Options {
			for _, target := range option.Targets {
				if target.PermanentID == a.target {
					return []int{option.Index}
				}
			}
		}
	case game.ChoiceMay:
		if a.accept {
			return []int{1}
		}
		return []int{0}
	}
	return defaultChoiceAgent{}.ChooseChoice(PlayerObservation{}, request)
}

func castProvokeCard(t *testing.T, g *game.Game, engine *Engine, name string, colored mana.Color, coloredCount, generic int) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, registryCardDef(t, name))
	if generic > 0 {
		g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	}
	g.Players[game.Player1].ManaPool.Add(colored, coloredCount)
	manaBefore := g.Players[game.Player1].ManaPool.Total()
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("could not cast %s for its printed mana cost", name)
	}
	if paid := manaBefore - g.Players[game.Player1].ManaPool.Total(); paid != coloredCount+generic {
		t.Fatalf("%s paid %d mana, want %d", name, paid, coloredCount+generic)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, cardID)
	if permanent == nil {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
	return permanent
}

func attackWithProvoke(t *testing.T, g *game.Game, engine *Engine, attacker, target *game.Permanent, clearSummoningSickness, accept bool) *provokeChoiceAgent {
	t.Helper()
	if clearSummoningSickness {
		attacker.SummoningSick = false
	}
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(attacker, game.AttackTarget{Player: game.Player2}),
	}}
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatalf("%v could not legally attack", attacker.CardInstanceID)
	}
	agent := &provokeChoiceAgent{target: target.ObjectID, accept: accept}
	agents := simPassPolicies()
	for player := range agents {
		agents[player] = agent
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("attacking did not put Provoke on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Provoke ability was not placed on the stack")
	}
	if len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("Provoke stack targets = %+v, want the selected defending creature %v", trigger.Targets, target.ObjectID)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(agent.request) == 0 || agent.request[len(agent.request)-1].Kind != game.ChoiceMay {
		t.Fatalf("Provoke choices = %+v, want an optional choice after target selection", agent.request)
	}
	var targetRequest *game.ChoiceRequest
	for _, request := range agent.request {
		if request.Kind == game.ChoiceTarget {
			targetRequest = &request
		}
	}
	var defendingCreatures []id.ID
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == game.Player2 && permanentHasType(g, permanent, types.Creature) {
			defendingCreatures = append(defendingCreatures, permanent.ObjectID)
		}
	}
	if len(defendingCreatures) > 1 && targetRequest == nil {
		t.Fatalf("Provoke omitted a target choice despite %d legal defending creatures", len(defendingCreatures))
	}
	if targetRequest != nil {
		if targetRequest.Player != game.Player1 {
			t.Fatalf("target chooser = %v, want attacking player", targetRequest.Player)
		}
		var eligible []id.ID
		for _, option := range targetRequest.Options {
			for _, chosen := range option.Targets {
				eligible = append(eligible, chosen.PermanentID)
			}
		}
		slices.Sort(eligible)
		slices.Sort(defendingCreatures)
		if !slices.Equal(eligible, defendingCreatures) {
			t.Fatalf("Provoke target options = %v, want defending player's creatures %v", eligible, defendingCreatures)
		}
	}
	if agent.request[len(agent.request)-1].Player != game.Player1 {
		t.Fatalf("optional-choice player = %v, want attacking player", agent.request[len(agent.request)-1].Player)
	}
	g.Turn.Step = game.StepDeclareBlockers
	return agent
}

func declareProvokeBlock(t *testing.T, g *game.Game, engine *Engine, attacker, blocker *game.Permanent) {
	t.Helper()
	declare := action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: attacker.ObjectID,
	}}}
	if !engine.applyDeclareBlockers(g, game.Player2, declare) {
		t.Fatalf("%v could not block %v", blocker.CardInstanceID, attacker.CardInstanceID)
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Goblin Grappler (Oracle 48c1c84c-f690-42ce-9c5a-dd09b1e4197c, printing
// 5c948872-295c-41b9-8094-db7db7578b0d): {R}, 1/1 Goblin, "Provoke (Whenever
// this creature attacks, you may have target creature defending player
// controls untap and block it if able.)". Its Scryfall rulings URI
// https://api.scryfall.com/cards/5c948872-295c-41b9-8094-db7db7578b0d/rulings
// returned HTTP 200 with no rulings. Verified official CR effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// read 603.1-3, 508.1d, 509.1b-c, 510.1, 701.21, and 702.38. These require
// the attack trigger to target a creature controlled by the defending player,
// offer the optional effect, untap that target, and require it to block the
// triggering attacker this combat if able.
func TestGoblinGrapplerProvokesDefendingCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	outsideDefender := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	forced.Tapped = true
	free := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	attacker := castProvokeCard(t, g, engine, "Goblin Grappler", mana.R, 1, 0)

	attackWithProvoke(t, g, engine, attacker, forced, true, true)
	if forced.Tapped {
		t.Fatal("Provoke did not untap the targeted defending creature")
	}
	if !blocksLegal(g, []game.BlockDeclaration{{Blocker: forced.ObjectID, Blocking: attacker.ObjectID}}) {
		t.Fatal("targeted creature could not block Goblin Grappler")
	}
	if blocksLegal(g, nil) || blocksLegal(g, []game.BlockDeclaration{{Blocker: free.ObjectID, Blocking: attacker.ObjectID}}) {
		t.Fatal("the defender could ignore the targeted block requirement")
	}
	if blocksLegal(g, []game.BlockDeclaration{{Blocker: outsideDefender.ObjectID, Blocking: attacker.ObjectID}}) {
		t.Fatal("a creature controlled by a nondefending player was offered as a legal blocker")
	}

	declareProvokeBlock(t, g, engine, attacker, forced)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActionsWithLog(g, &TurnLog{})
	if _, alive := permanentByObjectID(g, attacker.ObjectID); alive {
		t.Fatal("Goblin Grappler survived combat with the 2/2 forced blocker")
	}
}

func TestGoblinGrapplerMayDeclineProvoke(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	attacker := castProvokeCard(t, g, engine, "Goblin Grappler", mana.R, 1, 0)

	attackWithProvoke(t, g, engine, attacker, target, true, false)
	if target.Tapped {
		t.Fatal("declining Provoke untapped the target")
	}
	if !blocksLegal(g, nil) {
		t.Fatal("declining Provoke left a must-block requirement")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Crested Craghorn (Oracle 597c390e-deeb-4452-aeda-7cf9312a8b01, printing
// aadb40c8-3d54-4705-82dc-54e8d6e315d5): {4}{R}, 4/1 Goat Beast, "Haste /
// Provoke (Whenever this creature attacks, you may have target creature
// defending player controls untap and block it if able.)". Its Scryfall
// rulings URI
// https://api.scryfall.com/cards/aadb40c8-3d54-4705-82dc-54e8d6e315d5/rulings
// returned HTTP 200 with no rulings. Verified official CR effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// read 603.1-3, 508.1d, 509.1b-c, 510.1, 701.21, 702.10, and 702.38.
// Haste permits the newly controlled creature to attack; Provoke then untaps
// and requires the targeted defending creature to block it this combat if able.
func TestCrestedCraghornHasteAndProvoke(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	forced.Tapped = true
	attacker := castProvokeCard(t, g, engine, "Crested Craghorn", mana.R, 1, 4)

	attackWithProvoke(t, g, engine, attacker, forced, false, true)
	if forced.Tapped {
		t.Fatal("Provoke did not untap Crested Craghorn's target")
	}
	declareProvokeBlock(t, g, engine, attacker, forced)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActionsWithLog(g, &TurnLog{})
	if _, alive := permanentByObjectID(g, forced.ObjectID); alive {
		t.Fatal("forced 2/2 blocker survived Crested Craghorn's combat damage")
	}
	if _, alive := permanentByObjectID(g, attacker.ObjectID); alive {
		t.Fatal("Crested Craghorn survived the forced blocker, despite its 1 toughness")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Brontotherium (Oracle 6e3301a5-a99b-4b17-aa40-34ccc5904977, printing
// a171f5e2-ed3d-4675-a4fc-953ebb907aa0): {4}{G}{G}, 5/3 Beast, "Trample /
// Provoke (Whenever this creature attacks, you may have target creature
// defending player controls untap and block it if able.)". Its Scryfall
// rulings URI
// https://api.scryfall.com/cards/a171f5e2-ed3d-4675-a4fc-953ebb907aa0/rulings
// returned HTTP 200 with no rulings. Verified official CR effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// read 603.1-3, 508.1d, 509.1b-c, 510.1, 701.21, 702.19, and 702.38.
// Provoke forces the selected target to block this combat if able; trample
// assigns damage beyond lethal damage to the defending player.
func TestBrontotheriumTrampleAndProvoke(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	forced.Tapped = true
	attacker := castProvokeCard(t, g, engine, "Brontotherium", mana.G, 2, 4)
	life := g.Players[game.Player2].Life

	attackWithProvoke(t, g, engine, attacker, forced, true, true)
	declareProvokeBlock(t, g, engine, attacker, forced)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActionsWithLog(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != life-3 {
		t.Fatalf("defender life = %d, want %d after 5 trample damage assigns 2 to the blocker", got, life-3)
	}
	if _, alive := permanentByObjectID(g, forced.ObjectID); alive {
		t.Fatal("Brontotherium did not deal lethal damage to its forced blocker")
	}
	if _, alive := permanentByObjectID(g, attacker.ObjectID); !alive {
		t.Fatal("Brontotherium did not survive the 2 damage assigned by its blocker")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Deftblade Elite (Oracle 9a3e3a8f-7c19-4b52-a1b4-08eb09f6127a, printing
// d576b8f2-9b5b-47c0-8f9f-faa209d59595): {W}, 1/1 Human Soldier, "Provoke
// (Whenever this creature attacks, you may have target creature defending
// player controls untap and block it if able.) / {1}{W}: Prevent all combat
// damage that would be dealt to and dealt by this creature this turn." Its
// Scryfall rulings URI
// https://api.scryfall.com/cards/d576b8f2-9b5b-47c0-8f9f-faa209d59595/rulings
// returned HTTP 200 with no rulings. Verified official CR effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// read 603.1-3, 508.1d, 509.1b-c, 510.1, 615.1, and 702.38. The paid ability
// prevents combat damage both to and from Deftblade Elite for the turn.
func TestDeftbladeElitePreventsCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	forced.Tapped = true
	attacker := castProvokeCard(t, g, engine, "Deftblade Elite", mana.W, 1, 0)

	attackWithProvoke(t, g, engine, attacker, forced, true, true)
	declareProvokeBlock(t, g, engine, attacker, forced)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player1
	manaBefore := g.Players[game.Player1].ManaPool.Total()
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(attacker.ObjectID, 0, nil, 0)) {
		t.Fatal("could not activate Deftblade Elite's {1}{W} damage-prevention ability")
	}
	if paid := manaBefore - g.Players[game.Player1].ManaPool.Total(); paid != 2 {
		t.Fatalf("Deftblade Elite activation paid %d mana, want 2", paid)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActionsWithLog(g, &TurnLog{})
	if _, alive := permanentByObjectID(g, attacker.ObjectID); !alive {
		t.Fatal("Deftblade Elite took combat damage despite its prevention ability")
	}
	if _, alive := permanentByObjectID(g, forced.ObjectID); !alive {
		t.Fatal("Deftblade Elite dealt combat damage despite its prevention ability")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Krosan Vorine (Oracle ddb95c3a-b383-4987-ba24-72e95f8e33bf, printing
// 31bd1b2f-6bea-4bd1-8a6e-50ab2d34bf82): {3}{G}, 3/2 Cat Beast, "Provoke
// (Whenever this creature attacks, you may have target creature defending
// player controls untap and block it if able.) / This creature can't be
// blocked by more than one creature." Its Scryfall rulings URI
// https://api.scryfall.com/cards/31bd1b2f-6bea-4bd1-8a6e-50ab2d34bf82/rulings
// returned HTTP 200 with no rulings. Verified official CR effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// read 603.1-3, 508.1d, 509.1b-c, 510.1, and 702.38. Its blocking restriction
// limits the attacker to one creature; the Provoke requirement still makes
// the targeted creature's block mandatory if able.
func TestKrosanVorineCannotBeBlockedByMoreThanOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	forced.Tapped = true
	free := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	attacker := castProvokeCard(t, g, engine, "Krosan Vorine", mana.G, 1, 3)

	attackWithProvoke(t, g, engine, attacker, forced, true, true)
	twoBlockers := []game.BlockDeclaration{
		{Blocker: forced.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: free.ObjectID, Blocking: attacker.ObjectID},
	}
	if blocksLegal(g, twoBlockers) {
		t.Fatal("Krosan Vorine was legally blocked by two creatures")
	}
	if blocksLegal(g, []game.BlockDeclaration{{Blocker: free.ObjectID, Blocking: attacker.ObjectID}}) {
		t.Fatal("the defender could substitute a different blocker for Provoke's target")
	}
	declareProvokeBlock(t, g, engine, attacker, forced)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActionsWithLog(g, &TurnLog{})
	if _, alive := permanentByObjectID(g, attacker.ObjectID); !alive {
		t.Fatal("Krosan Vorine did not survive the single 1/1 blocker")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Lowland Tracker (Oracle de06b33f-0793-419a-bf65-1961634af8e7, printing
// f8eaaded-b18a-4614-b5b5-b4bb49a2e1b1): {4}{W}, 2/2 Human Soldier, "First
// strike / Provoke (Whenever this creature attacks, you may have target
// creature defending player controls untap and block it if able.)". Its
// Scryfall rulings URI
// https://api.scryfall.com/cards/f8eaaded-b18a-4614-b5b5-b4bb49a2e1b1/rulings
// returned HTTP 200 with no rulings. Verified official CR effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// read 603.1-3, 508.1d, 509.1b-c, 510.1, 701.21, 702.7, and 702.38. First
// strike deals damage before a creature without first strike; Provoke requires
// the selected creature to block this combat if able.
func TestLowlandTrackerFirstStrikeAndProvoke(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	forced.Tapped = true
	attacker := castProvokeCard(t, g, engine, "Lowland Tracker", mana.W, 1, 4)

	attackWithProvoke(t, g, engine, attacker, forced, true, true)
	declareProvokeBlock(t, g, engine, attacker, forced)
	g.Turn.Step = game.StepCombatDamage
	log := &TurnLog{}
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, log)
	engine.applyStateBasedActionsWithLog(g, log)
	if _, alive := permanentByObjectID(g, forced.ObjectID); alive {
		t.Fatal("Lowland Tracker's first-strike damage did not destroy the forced 2/2 blocker")
	}
	engine.resolveCombatDamage(g, log)
	engine.applyStateBasedActionsWithLog(g, log)
	if _, alive := permanentByObjectID(g, attacker.ObjectID); !alive {
		t.Fatal("Lowland Tracker did not survive after first strike removed its blocker")
	}
}

var _ PlayerAgent = (*provokeChoiceAgent)(nil)
