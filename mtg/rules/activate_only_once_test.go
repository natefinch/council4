package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine fixture for #3956, modeled on Surge Engine
// (864d9db7-34ca-4440-8d6d-096f9eddedd4), not a card admission. Oracle:
// "Defender / {U}: This creature loses defender and gains "This creature can't
// be blocked." / {2}{U}: This creature becomes blue and has base power and
// toughness 5/4. Activate only if this creature doesn't have defender. /
// {4}{U}{U}: Draw three cards. Activate only if this creature is blue and only
// once." The fixture keeps only the color change and the draw ability.
// CR 20261106: 400.7 (a permanent that changes zones is a new object), 602.2
// (activation completes once costs are paid), 602.5b (a use restriction keeps
// applying to the object after a control change), 707.10 (a copy of an
// activated ability isn't activated).

const (
	onlyOnceBlueIndex = 0
	onlyOnceDrawIndex = 1
)

func activateOnlyOnceEngineDef() *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Only Once Engine",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{
			{
				ManaCost: opt.Val(cost.Mana{cost.U}),
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
					Object: opt.Val(game.SourcePermanentReference()),
					ContinuousEffects: []game.ContinuousEffect{{
						Duration:  game.DurationPermanent,
						Layer:     game.LayerColor,
						SetColors: []color.Color{color.Blue},
					}},
					Duration: game.DurationPermanent,
				}}}}.Ability(),
			},
			{
				ManaCost: opt.Val(cost.Mana{cost.O(4), cost.U, cost.U}),
				ActivationCondition: opt.Val(game.Condition{
					Object:        opt.Val(game.SourcePermanentReference()),
					ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Blue}}),
				}),
				ActivateOnlyOnce: true,
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
					Amount: game.Fixed(3),
					Player: game.ControllerReference(),
				}}}}.Ability(),
			},
		},
	}}
}

type activateOnlyOnceFixture struct {
	engine   *Engine
	registry *cards.Registry
	game     *game.Game
	source   *game.Permanent
}

func newActivateOnlyOnceFixture(t *testing.T) activateOnlyOnceFixture {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	for range 12 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
	}
	return activateOnlyOnceFixture{
		engine:   NewEngineWithPrograms(nil, registry.Programs()),
		registry: registry,
		game:     g,
		source:   addCombatPermanent(g, game.Player1, activateOnlyOnceEngineDef()),
	}
}

func (f activateOnlyOnceFixture) drawAction(source *game.Permanent) action.Action {
	return action.ActivateAbility(source.ObjectID, onlyOnceDrawIndex, nil, 0)
}

func (f activateOnlyOnceFixture) drawOffered(player game.PlayerID, source *game.Permanent) bool {
	return containsAction(f.engine.legalActions(f.game, player), f.drawAction(source))
}

// makeBlue activates and resolves the source's color-changing ability.
func (f activateOnlyOnceFixture) makeBlue(t *testing.T, player game.PlayerID, source *game.Permanent) {
	t.Helper()
	f.game.Players[player].ManaPool.Add(mana.U, 1)
	if !f.engine.applyAction(f.game, player, action.ActivateAbility(source.ObjectID, onlyOnceBlueIndex, nil, 0)) {
		t.Fatal("could not activate the color-changing ability")
	}
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
}

// activateDraw pays {4}{U}{U} and activates the draw ability, leaving it on
// the stack.
func (f activateOnlyOnceFixture) activateDraw(t *testing.T, player game.PlayerID, source *game.Permanent) {
	t.Helper()
	f.game.Players[player].ManaPool.Add(mana.U, 6)
	if !f.engine.applyAction(f.game, player, f.drawAction(source)) {
		t.Fatal("could not activate the only-once draw ability")
	}
}

func (f activateOnlyOnceFixture) resetMainPhase(player game.PlayerID) {
	f.game.Turn.ActivePlayer = player
	setMainPhasePriority(f.game, player)
}

// TestActivateOnlyOnceRequiresConditionAndNeverReturns proves the draw ability
// combines its color condition with the only-once restriction: it isn't offered
// while the creature is colorless, works once when it's blue, and is neither
// offered nor applicable again, even on a later turn.
func TestActivateOnlyOnceRequiresConditionAndNeverReturns(t *testing.T) {
	f := newActivateOnlyOnceFixture(t)
	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player1, f.source) {
		t.Fatal("draw ability was offered while the creature isn't blue")
	}
	f.game.Players[game.Player1].ManaPool.Empty()
	f.makeBlue(t, game.Player1, f.source)

	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if !f.drawOffered(game.Player1, f.source) {
		t.Fatal("draw ability was not offered once the creature is blue")
	}
	if !f.engine.applyAction(f.game, game.Player1, f.drawAction(f.source)) {
		t.Fatal("could not activate the draw ability")
	}
	hand := f.game.Players[game.Player1].Hand.Size()
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
	if got := f.game.Players[game.Player1].Hand.Size() - hand; got != 3 {
		t.Fatalf("draw ability drew %d cards, want 3", got)
	}

	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player1, f.source) {
		t.Fatal("draw ability was offered again after its only activation")
	}
	if f.engine.applyAction(f.game, game.Player1, f.drawAction(f.source)) {
		t.Fatal("draw ability was activated a second time")
	}

	f.engine.advanceToNextTurn(f.game)
	f.resetMainPhase(game.Player1)
	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player1, f.source) {
		t.Fatal("draw ability was offered again on a later turn")
	}
	if !containsAction(f.engine.legalActions(f.game, game.Player1), action.ActivateAbility(f.source.ObjectID, onlyOnceBlueIndex, nil, 0)) {
		t.Fatal("the restriction blocked the source's other, unrestricted ability")
	}
}

// TestActivateOnlyOnceCounteredActivationStillCounts proves the use is spent
// when the ability is activated, so Stifle countering it doesn't restore it.
func TestActivateOnlyOnceCounteredActivationStillCounts(t *testing.T) {
	f := newActivateOnlyOnceFixture(t)
	f.makeBlue(t, game.Player1, f.source)
	f.activateDraw(t, game.Player1, f.source)
	ability, ok := f.game.Stack.Peek()
	if !ok || ability.Kind != game.StackActivatedAbility {
		t.Fatalf("stack top = %+v, want the draw ability", ability)
	}

	defs, err := f.registry.LookupAllChecked("Stifle")
	if err != nil || len(defs) != 1 {
		t.Fatalf("Stifle definition: %d, %v", len(defs), err)
	}
	stifleID := addCardToHand(f.game, game.Player2, defs[0])
	f.game.Players[game.Player2].ManaPool.Add(mana.U, 1)
	f.game.Turn.PriorityPlayer = game.Player2
	if !f.engine.applyAction(f.game, game.Player2, action.CastSpell(stifleID, []game.Target{game.StackObjectTarget(ability.ID)}, 0, nil)) {
		t.Fatal("Stifle could not target the draw ability")
	}
	hand := f.game.Players[game.Player1].Hand.Size()
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
	if f.game.Stack.Size() != 0 {
		t.Fatalf("Stifle did not counter the draw ability: stack=%+v", f.game.Stack.Objects())
	}
	if f.game.Players[game.Player1].Hand.Size() != hand {
		t.Fatal("countered draw ability still drew cards")
	}

	f.resetMainPhase(game.Player1)
	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player1, f.source) {
		t.Fatal("draw ability was offered again after its countered activation")
	}
}

// TestActivateOnlyOnceNewObjectAfterFlicker proves the use belongs to the
// permanent object: after the creature is exiled and returned it is a new
// object whose draw ability can be activated once (CR 400.7).
func TestActivateOnlyOnceNewObjectAfterFlicker(t *testing.T) {
	f := newActivateOnlyOnceFixture(t)
	f.makeBlue(t, game.Player1, f.source)
	f.activateDraw(t, game.Player1, f.source)
	f.engine.resolveTopOfStack(f.game, &TurnLog{})

	obj := &game.StackObject{
		ID:         f.game.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(f.game, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blink Spell"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(f.source.ObjectID)},
	}
	instrs := immediateBlinkInstructions(game.PutOnBattlefield{})
	for i := range instrs {
		f.engine.resolveInstructionWithChoices(f.game, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	returned, ok := findPermanentByCardID(f.game, f.source.CardInstanceID)
	if !ok || returned.ObjectID == f.source.ObjectID {
		t.Fatalf("returned permanent = %+v, want a new object for the same card", returned)
	}

	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player1, returned) {
		t.Fatal("returned creature isn't blue but its draw ability was offered")
	}
	f.game.Players[game.Player1].ManaPool.Empty()
	f.makeBlue(t, game.Player1, returned)
	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if !f.drawOffered(game.Player1, returned) {
		t.Fatal("the new object's draw ability was not offered")
	}
	if !f.engine.applyAction(f.game, game.Player1, f.drawAction(returned)) {
		t.Fatal("could not activate the new object's draw ability")
	}
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
	f.game.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player1, returned) {
		t.Fatal("the new object's draw ability was offered a second time")
	}
}

// TestActivateOnlyOnceFollowsControlChange proves a new controller can't
// activate an ability the object already used (CR 602.5b).
func TestActivateOnlyOnceFollowsControlChange(t *testing.T) {
	f := newActivateOnlyOnceFixture(t)
	f.makeBlue(t, game.Player1, f.source)
	f.activateDraw(t, game.Player1, f.source)
	f.engine.resolveTopOfStack(f.game, &TurnLog{})

	steal := &game.StackObject{
		ID:         f.game.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(f.game, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Steal Spell"}}),
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(f.source.ObjectID)},
	}
	f.engine.resolveInstructionWithChoices(f.game, steal, &game.Instruction{Primitive: game.ApplyContinuous{
		Object:            opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerControl, NewController: opt.Val(game.Player2)}},
		Duration:          game.DurationPermanent,
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := effectiveController(f.game, f.source); got != game.Player2 {
		t.Fatalf("controller = %v, want Player2", got)
	}

	f.engine.advanceToNextTurn(f.game)
	f.resetMainPhase(game.Player2)
	f.game.Players[game.Player2].ManaPool.Add(mana.U, 6)
	if f.drawOffered(game.Player2, f.source) {
		t.Fatal("new controller was offered the object's already used draw ability")
	}
}

// TestActivateOnlyOnceSurvivesGameClone proves the used state is game-owned
// and copied by Clone, and that the clone and original stay independent.
func TestActivateOnlyOnceSurvivesGameClone(t *testing.T) {
	f := newActivateOnlyOnceFixture(t)
	f.makeBlue(t, game.Player1, f.source)
	before := f.game.Clone()
	f.activateDraw(t, game.Player1, f.source)
	f.engine.resolveTopOfStack(f.game, &TurnLog{})
	after := f.game.Clone()

	after.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if containsAction(f.engine.legalActions(after, game.Player1), f.drawAction(f.source)) {
		t.Fatal("clone taken after the activation offered the used draw ability")
	}
	before.Players[game.Player1].ManaPool.Add(mana.U, 6)
	if !containsAction(f.engine.legalActions(before, game.Player1), f.drawAction(f.source)) {
		t.Fatal("clone taken before the activation lost the unused draw ability")
	}
	if !f.engine.applyAction(before, game.Player1, f.drawAction(f.source)) {
		t.Fatal("clone taken before the activation could not activate the draw ability")
	}
	if !f.game.ActivatedAbilitiesUsedOnce[game.ActivatedAbilityUse{SourceID: f.source.ObjectID, AbilityIndex: onlyOnceDrawIndex}] {
		t.Fatal("original game lost its recorded use")
	}
}

// TestActivateOnlyOnceFailsClosedOutsideBattlefield proves an only-once
// ability that would function from a card zone is not offered, since its use
// can't be tracked by a battlefield object.
func TestActivateOnlyOnceFailsClosedOutsideBattlefield(t *testing.T) {
	for _, onlyOnce := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		setMainPhasePriority(g, game.Player1)
		cardID := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Graveyard Ability",
			Types: []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{{
				ZoneOfFunction:   zone.Graveyard,
				ActivateOnlyOnce: onlyOnce,
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
					Amount: game.Fixed(1), Player: game.ControllerReference(),
				}}}}.Ability(),
			}},
		}})
		g.Players[game.Player1].Graveyard.Add(cardID)
		offered := containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(cardID, 0, nil, 0))
		if offered == onlyOnce {
			t.Fatalf("graveyard ability with ActivateOnlyOnce=%v offered=%v", onlyOnce, offered)
		}
	}
}
