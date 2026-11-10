package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Builder's Talent Oracle ID 0b19082c-8f80-448b-9e1b-0f324a159409,
// printing 15fa581a-724e-4196-a9a3-ff84c54bdb7d; Scryfall's six rulings retrieved
// 2026-11-09
// (https://api.scryfall.com/cards/15fa581a-724e-4196-a9a3-ff84c54bdb7d/rulings)
// were reviewed. CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 111.1-2, 717.2-3, 603.2c/603.4, and 608.2b require the white 0/4 Wall,
// gated Class abilities, a single trigger for one-or-more entries, and a legal
// target when the level-up ability resolves.
func TestBuildersTalentOwnText(t *testing.T) {
	g, engine, class := castClassForTest(t, "0b19082c-8f80-448b-9e1b-0f324a159409", mana.W, 1)
	if class.ClassLevel != 1 {
		t.Fatalf("Builder's Talent entered at level %d, want 1", class.ClassLevel)
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Builder's Talent's enters trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	var wall *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanentEffectiveName(g, permanent) == "Wall" {
			wall = permanent
			break
		}
	}
	if wall == nil {
		t.Fatal("Builder's Talent did not create its Wall token")
	}
	if effectivePower(g, wall) != 0 {
		t.Fatalf("Wall power = %d, want 0", effectivePower(g, wall))
	}
	if toughness, ok := effectiveToughness(g, wall); !ok || toughness != 4 {
		t.Fatalf("Wall toughness = %d (defined=%v), want 4/true", toughness, ok)
	}
	if !hasKeyword(g, wall, game.Defender) {
		t.Fatal("Wall token lacks defender")
	}

	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	activateClassLevelForTest(t, g, engine, class, 0)
	enterClassTestPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Builder's Relic",
		Types: []types.Card{types.Artifact},
	}})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Builder's Talent's level-2 entry trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := wall.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Wall +1/+1 counters = %d, want 1", got)
	}

	graveyardCard := addClassTestCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Restored Relic",
		Types: []types.Card{types.Artifact},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.W, 5)
	activateClassLevelForTest(t, g, engine, class, 1)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Builder's Talent's level-3 trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	returned := classTestPermanentForCard(g, graveyardCard)
	if returned == nil || returned.Owner != game.Player1 || returned.Controller != game.Player1 {
		t.Fatalf("returned permanent = %+v, want the targeted artifact under its owner's control", returned)
	}
}

// Source review: pinned Leader's Talent Oracle ID 134430b7-d566-485a-bd5c-7db8a285f4ca,
// printing 4cbcb622-1aff-460f-b8fa-4502d991e0ad; all seven Scryfall rulings retrieved
// 2026-11-09
// (https://api.scryfall.com/cards/4cbcb622-1aff-460f-b8fa-4502d991e0ad/rulings)
// were reviewed, including that its level-3 trigger resolves before the spell.
// CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 508.3d, 603.2c/603.4, and 717.2-3 govern the one-or-more attack trigger,
// the intervening counter condition, and level gating; the cited ruling confirms
// the cast trigger resolves above the spell.
func TestLeadersTalentOwnText(t *testing.T) {
	g, engine, class := castClassForTest(t, "134430b7-d566-485a-bd5c-7db8a285f4ca", mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 3)
	activateClassLevelForTest(t, g, engine, class, 0)

	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	declareClassAttackers(t, g, engine, first, second)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Leader's Talent's attack trigger did not trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("attack-trigger stack size = %d, want one trigger for both attackers", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	counterBearer, plainCreature := second, first
	if first.Counters.Get(counter.PlusOnePlusOne) == 1 {
		counterBearer, plainCreature = first, second
	}
	if counterBearer.Counters.Get(counter.PlusOnePlusOne) != 1 || plainCreature.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatalf("attack trigger counters = %d/%d, want exactly one counter on one attacker",
			first.Counters.Get(counter.PlusOnePlusOne), second.Counters.Get(counter.PlusOnePlusOne))
	}

	startingLife := g.Players[game.Player1].Life
	if !movePermanentToZone(g, counterBearer, zone.Graveyard) {
		t.Fatal("moving the counter-bearing creature from the battlefield failed")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Leader's Talent did not trigger when its counter-bearing creature left")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player1].Life; got != startingLife+2 {
		t.Fatalf("life after counter-bearing creature left = %d, want %d", got, startingLife+2)
	}
	if !movePermanentToZone(g, plainCreature, zone.Graveyard) {
		t.Fatal("moving the creature without counters from the battlefield failed")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Leader's Talent triggered for a creature that left without a counter")
	}

	g.Combat = nil
	g.Players[game.Player1].ManaPool.Add(mana.W, 4)
	activateClassLevelForTest(t, g, engine, class, 1)
	first = addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second = addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, classTestInstant())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("casting a spell with Leader's Talent at level 3 failed")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Leader's Talent's spell trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	for _, creature := range []*game.Permanent{first, second} {
		if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
			t.Fatalf("creature's counters after spell trigger = %d, want 1", got)
		}
	}
	if got := g.Players[game.Player1].Life; got != startingLife+2 {
		t.Fatalf("spell resolved before Leader's Talent's trigger: life = %d, want %d", got, startingLife+2)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player1].Life; got != startingLife+3 {
		t.Fatalf("life after resolving the original spell = %d, want %d", got, startingLife+3)
	}
}

// Source review: pinned Paladin Class Oracle ID ac9d52b2-78a5-4b17-8aaf-1c1cb936a508,
// printing 5bf81fb1-7992-4ae9-b1a8-80c31579a2bf; all nine Scryfall rulings retrieved
// 2026-11-09
// (https://api.scryfall.com/cards/5bf81fb1-7992-4ae9-b1a8-80c31579a2bf/rulings)
// were reviewed, including cumulative taxes and the single-attacker ruling.
// CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 601.2f, 613.4c, 508.3d, 717.2-3, and 702.4a-b govern the turn-limited spell
// tax, level-gated anthem, batched attack trigger, and double strike.
func TestPaladinClassOwnText(t *testing.T) {
	g, engine, class := castClassForTest(t, "ac9d52b2-78a5-4b17-8aaf-1c1cb936a508", mana.W, 0)

	opponentSpell := addCardToHand(g, game.Player2, classTestInstant())
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, action.CastSpell(opponentSpell, nil, 0, nil)) {
		t.Fatal("opponent's spell was not payable with Paladin Class's additional generic cost")
	}
	if got := g.Players[game.Player2].ManaPool.Total(); got != 0 {
		t.Fatalf("opponent left %d mana after casting, want zero after paying the additional {1}", got)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	ownSpell := addCardToHand(g, game.Player1, classTestInstant())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(ownSpell, nil, 0, nil)) {
		t.Fatal("Paladin Class incorrectly taxed a spell outside its controller's turn")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	g.Players[game.Player1].ManaPool.Add(mana.W, 8)
	activateClassLevelForTest(t, g, engine, class, 0)
	ally := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponent := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	if effectivePower(g, ally) != 3 || effectivePower(g, opponent) != 2 {
		t.Fatalf("level-2 anthem powers = %d/%d, want own 3 and opponent 2", effectivePower(g, ally), effectivePower(g, opponent))
	}
	activateClassLevelForTest(t, g, engine, class, 1)

	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	declareClassAttackers(t, g, engine, first, second)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Paladin Class's level-3 attack trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	buffed, other := first, second
	if effectivePower(g, first) != 4 {
		buffed, other = second, first
	}
	if effectivePower(g, buffed) != 4 || effectiveToughnessMust(t, g, buffed) != 4 {
		t.Fatalf("targeted attacker is %d/%d, want 4/4 after level-2 and level-3 bonuses",
			effectivePower(g, buffed), effectiveToughnessMust(t, g, buffed))
	}
	if !hasKeyword(g, buffed, game.DoubleStrike) {
		t.Fatal("targeted attacker did not gain double strike")
	}
	if effectivePower(g, other) != 3 || hasKeyword(g, other, game.DoubleStrike) {
		t.Fatal("non-targeted attacker received the targeted bonus")
	}

	singleGame, singleEngine, singleClass := castClassForTest(t, "ac9d52b2-78a5-4b17-8aaf-1c1cb936a508", mana.W, 0)
	singleGame.Players[game.Player1].ManaPool.Add(mana.W, 8)
	activateClassLevelForTest(t, singleGame, singleEngine, singleClass, 0)
	activateClassLevelForTest(t, singleGame, singleEngine, singleClass, 1)
	singleAttacker := addCombatCreaturePermanentWithPower(singleGame, game.Player1, 2)
	declareClassAttackers(t, singleGame, singleEngine, singleAttacker)
	if !singleEngine.putTriggeredAbilitiesOnStackWithChoices(singleGame, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Paladin Class did not trigger for a single attacker")
	}
	singleEngine.resolveTopOfStackWithChoices(singleGame, simPassPolicies(), &TurnLog{})
	if effectivePower(singleGame, singleAttacker) != 3 || effectiveToughnessMust(t, singleGame, singleAttacker) != 3 {
		t.Fatal("single attacker received a +1/+1 bonus despite having no other attackers")
	}
	if !hasKeyword(singleGame, singleAttacker, game.DoubleStrike) {
		t.Fatal("single attacker did not gain double strike")
	}
}

// Source review: pinned Gossip's Talent Oracle ID 96e6d220-87db-4dab-abaa-e80c9577d89a,
// printing b299889a-03d6-4659-b0e1-f0830842e40f; all eleven Scryfall rulings retrieved
// 2026-11-09
// (https://api.scryfall.com/cards/b299889a-03d6-4659-b0e1-f0830842e40f/rulings)
// were reviewed, including resolution-time power checks, per-creature damage
// triggers, new-object returns, and token behavior. CR 2026-11-06, official TXT
// at https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 400.7, 508.3d, 510.2/510.3a, 717.2-3, and 701.25 govern these interactions.
func TestGossipsTalentOwnText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First surveilled card"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second surveilled card"}})
	class := castClassForTestInGame(t, g, engine, "96e6d220-87db-4dab-abaa-e80c9577d89a", mana.U, 0)

	choiceAgent := &classCardChoiceAgent{acceptOptional: true}
	agents := simPassPolicies()
	agents[game.Player1] = choiceAgent
	entered := enterClassTestPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "New Creature", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
	}})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gossip's Talent did not trigger for a creature entering")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(g.Players[game.Player1].Graveyard.All()) != 1 || len(g.Players[game.Player1].Library.All()) != 1 {
		t.Fatal("Gossip's Talent did not surveil one card when a creature entered")
	}
	if !entered.SummoningSick {
		t.Fatal("newly entered creature should not be eligible to attack yet")
	}

	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	activateClassLevelForTest(t, g, engine, class, 0)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	activateClassLevelForTest(t, g, engine, class, 1)

	declareClassAttackers(t, g, engine, attacker)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gossip's Talent's level-2 attack trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	g.Turn.Step = game.StepDeclareBlockers
	if engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{{
		Blocker: blocker.ObjectID, Blocking: attacker.ObjectID,
	}}}) {
		t.Fatal("the target of Gossip's Talent's level-2 trigger was blockable")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{}) {
		t.Fatal("empty blocker declaration was rejected")
	}
	g.Turn.Step = game.StepCombatDamage
	lifeBeforeDamage := g.Players[game.Player2].Life
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != lifeBeforeDamage-2 {
		t.Fatalf("unblocked attacker dealt life change to %d, want %d", got, lifeBeforeDamage-2)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gossip's Talent's level-3 combat-damage trigger did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("exiled-and-returned attacker retained its old permanent object")
	}
	returned := classTestPermanentForCard(g, g.CardInstances[attacker.CardInstanceID])
	if returned == nil || returned.ObjectID == attacker.ObjectID ||
		returned.Owner != game.Player1 || returned.Controller != game.Player1 {
		t.Fatalf("returned attacker = %+v, want a new object under its owner's control", returned)
	}
	for _, declaration := range g.Combat.Attackers {
		if declaration.Attacker == returned.ObjectID {
			t.Fatal("returned creature remained attacking")
		}
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the returned creature did not trigger Gossip's Talent's creature-entry ability")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(g.Players[game.Player1].Graveyard.All()) != 2 || len(g.Players[game.Player1].Library.All()) != 0 {
		t.Fatal("Gossip's Talent did not surveil again when its level-3 ability returned the creature")
	}
}

type classCardChoiceAgent struct {
	simPassPolicy
	acceptOptional bool
}

func (a *classCardChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		if a.acceptOptional {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceSurveil:
		return []int{1}
	default:
		return request.DefaultSelection
	}
}

func castClassForTest(t *testing.T, oracleID string, color mana.Color, generic int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	class := castClassForTestInGame(t, g, engine, oracleID, color, generic)
	return g, engine, class
}

func castClassForTestInGame(t *testing.T, g *game.Game, engine *Engine, oracleID string, color mana.Color, generic int) *game.Permanent {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s failed: %v", oracleID, err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(color, 1)
	if generic > 0 {
		g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", def.Name)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	permanent := classTestPermanentForCard(g, g.CardInstances[cardID])
	if permanent == nil {
		t.Fatalf("%s did not enter the battlefield", def.Name)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("casting %s left %d mana, want exact payment", def.Name, got)
	}
	return permanent
}

func activateClassLevelForTest(t *testing.T, g *game.Game, engine *Engine, class *game.Permanent, ability int) {
	t.Helper()
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(class.ObjectID, ability, nil, 0)) {
		t.Fatalf("activating %s level ability %d failed", permanentEffectiveName(g, class), ability)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
}

func classTestPermanentForCard(g *game.Game, card *game.CardInstance) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == card.ID {
			return permanent
		}
	}
	return nil
}

func enterClassTestPermanent(t *testing.T, g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, controller, def)
	permanent, ok := createCardPermanent(g, g.CardInstances[cardID], controller, zone.Hand)
	if !ok {
		t.Fatalf("creating %s on the battlefield failed", def.Name)
	}
	return permanent
}

func addClassTestCardToGraveyard(g *game.Game, owner game.PlayerID, def *game.CardDef) *game.CardInstance {
	cardID := g.IDGen.Next()
	card := &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	g.CardInstances[cardID] = card
	g.Players[owner].Graveyard.Add(cardID)
	return card
}

func classTestInstant() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Class test instant",
		ManaCost: opt.Val(cost.Mana{cost.U}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

func declareClassAttackers(t *testing.T, g *game.Game, engine *Engine, attackers ...*game.Permanent) {
	t.Helper()
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.PriorityPlayer = game.Player1
	g.Combat = &game.CombatState{}
	declarations := make([]game.AttackDeclaration, len(attackers))
	for i, attacker := range attackers {
		declarations[i] = game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}
	}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: declarations}) {
		t.Fatal("attacker declaration failed")
	}
}

func effectiveToughnessMust(t *testing.T, g *game.Game, permanent *game.Permanent) int {
	t.Helper()
	toughness, ok := effectiveToughness(g, permanent)
	if !ok {
		t.Fatalf("%s has no effective toughness", permanentEffectiveName(g, permanent))
	}
	return toughness
}
