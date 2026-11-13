package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: prototype (CR 702.160a; 719.1-719.5) from the verified
// Comprehensive Rules effective 2026-11-06, with 105.2 (an object's colors are
// its mana cost's colored symbols), 118.9 (alternative costs) and 601.2b/601.2f
// (cast choices and cost totals). 719.3a judges castability only by the
// prototype values; 719.3b gives a prototyped spell and permanent the
// prototype mana cost, colors, power and toughness; 719.3c/719.3d make copies
// of a prototyped spell or permanent prototyped; 719.4 keeps the normal
// characteristics in every other zone; 719.5 leaves all other characteristics
// unchanged. Scryfall rulings for Phyrexian Fleshgorger
// (d3a5a830-cd14-49da-9412-c50049c74c92, 2022-10-14) and Blitz Automaton
// (19f344ee-031d-4346-9553-c5d32b02df43): prototype is not an alternative cost,
// so a "without paying its mana cost" cast may be prototyped and it may be cast
// prototyped from any zone it could be cast from; token copies and objects
// that become copies of a prototyped permanent are prototyped; a prototyped
// card that leaves the stack or battlefield is its normal self again (Blitz
// Automaton in a graveyard is colorless with mana value 7). These are Engine
// scenarios on a test-local Blitz Automaton-shaped card, not a registered card.
// Spell-route checklist: storm is not exercised because a storm copy copies the
// storm spell itself and no prototype card has storm; granted graveyard
// alternative costs (flashback, escape, jump-start) remain front-face routes
// because no prototype card prints them.

const prototypeTestName = "Prototype Test Automaton"

func prototypeTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      prototypeTestName,
		ManaCost:  opt.Val(cost.Mana{cost.O(7)}),
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{Value: 6}),
		Toughness: opt.Val(game.PT{Value: 4}),
		Prototype: opt.Val(game.PrototypeProfile{
			ManaCost:  cost.Mana{cost.O(2), cost.R},
			Power:     game.PT{Value: 3},
			Toughness: game.PT{Value: 2},
		}),
		StaticAbilities: []game.StaticAbility{{
			Text:             "Haste",
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Haste}},
		}},
		OracleText: "Prototype {2}{R} — 3/2\nHaste",
	}}
}

func prototypeScenario(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return g, NewEngineWithPrograms(nil, cards.NewDefaultRegistry().Programs())
}

func castPrototypeTestCard(t *testing.T, g *game.Game, engine *Engine, cardID id.ID, face game.FaceIndex) *game.StackObject {
	t.Helper()
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, face, nil, 0, nil)) {
		t.Fatalf("cast of face %d failed", face)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID {
		t.Fatal("the prototype test card is not on top of the stack")
	}
	return obj
}

func assertPrototypeSpell(t *testing.T, g *game.Game, obj *game.StackObject, wantMV int, wantColors []color.Color) {
	t.Helper()
	if mv, ok := stackObjectManaValue(g, obj); !ok || mv != wantMV {
		t.Fatalf("spell mana value = %d, %v; want %d", mv, ok, wantMV)
	}
	if colors, _ := stackObjectColors(g, obj); !slices.Equal(colors, wantColors) {
		t.Fatalf("spell colors = %v, want %v", colors, wantColors)
	}
}

func assertPrototypePermanent(t *testing.T, g *game.Game, permanent *game.Permanent, power, toughness, manaValue int, wantColors []color.Color) {
	t.Helper()
	if permanent == nil {
		t.Fatal("no permanent")
	}
	if got := effectivePower(g, permanent); got != power {
		t.Fatalf("power = %d, want %d", got, power)
	}
	if got, _ := effectiveToughness(g, permanent); got != toughness {
		t.Fatalf("toughness = %d, want %d", got, toughness)
	}
	if got, _ := effectivePermanentManaValue(g, permanent); got != manaValue {
		t.Fatalf("mana value = %d, want %d", got, manaValue)
	}
	if got := permanentEffectiveColors(g, permanent); !slices.Equal(got, wantColors) {
		t.Fatalf("colors = %v, want %v", got, wantColors)
	}
	def, ok := permanentFaceDef(g, permanent)
	if !ok || def.Name != prototypeTestName || !def.HasType(types.Artifact) || !def.HasType(types.Creature) ||
		!slices.Contains(def.Subtypes, types.Construct) {
		t.Fatalf("permanent lost its other characteristics: %+v", def)
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("permanent lost haste")
	}
}

func assertNormalPrototypeCard(t *testing.T, g *game.Game, cardID id.ID, where string) {
	t.Helper()
	normal := game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 7}), Colorless: true}
	if !cardMatchesSelection(g, nil, g.CardInstances[cardID], normal) {
		t.Fatalf("card %s is not its normal colorless mana value 7 self", where)
	}
}

// TestPrototypeCastsOfferedOnlyWhenPayable proves the normal and prototyped
// casts are separate choices, each offered only when its own mana cost can be
// paid (CR 719.3, 719.3a).
func TestPrototypeCastsOfferedOnlyWhenPayable(t *testing.T) {
	tests := []struct {
		name          string
		pool          []mana.Color
		front, protod bool
	}{
		{"prototype only", []mana.Color{mana.C, mana.C, mana.R}, false, true},
		{"normal only", slices.Repeat([]mana.Color{mana.C}, 7), true, false},
		{"both", append(slices.Repeat([]mana.Color{mana.C}, 7), mana.R), true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, engine := prototypeScenario(t)
			cardID := addCardToHand(g, game.Player1, prototypeTestDef())
			fuseAddMana(g, game.Player1, tt.pool...)
			actions := engine.legalActions(g, game.Player1)
			if got := fuseCastFaceOffered(actions, cardID, zone.Hand, game.FaceFront); got != tt.front {
				t.Fatalf("normal cast offered = %v, want %v", got, tt.front)
			}
			if got := fuseCastFaceOffered(actions, cardID, zone.Hand, game.FacePrototype); got != tt.protod {
				t.Fatalf("prototype cast offered = %v, want %v", got, tt.protod)
			}
			if !tt.protod && engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FacePrototype, nil, 0, nil)) {
				t.Fatal("unpayable prototyped cast was applied")
			}
			if !tt.front && engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceFront, nil, 0, nil)) {
				t.Fatal("unpayable normal cast was applied")
			}
		})
	}
}

// TestPrototypeCastUsesAlternateCharacteristicsOnlyOnStackAndBattlefield
// casts the card prototyped: the spell and the permanent are red, mana value
// 3 and 3/2 with the same name, types, abilities and card identity, and a
// red-spell cast trigger (Dragon's Claw) sees it. In hand and after it dies the
// card is its normal colorless 6/4 mana value 7 self.
func TestPrototypeCastUsesAlternateCharacteristicsOnlyOnStackAndBattlefield(t *testing.T) {
	g, engine := prototypeScenario(t)
	addCombatPermanent(g, game.Player1, registryCard(t, "Dragon's Claw"))
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	assertNormalPrototypeCard(t, g, cardID, "in hand")
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.R)

	obj := castPrototypeTestCard(t, g, engine, cardID, game.FacePrototype)
	if obj.Face != game.FacePrototype {
		t.Fatalf("spell face = %d, want prototype", obj.Face)
	}
	assertPrototypeSpell(t, g, obj, 3, []color.Color{color.Red})
	cast := lastSpellCastEvent(t, g)
	if !slices.Equal(cast.Colors, []color.Color{color.Red}) || cast.ManaValue != opt.Val(3) {
		t.Fatalf("cast event colors %v mana value %v, want red 3", cast.Colors, cast.ManaValue)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Dragon's Claw didn't trigger on the red prototyped spell")
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())

	permanent := permanentForCard(g, cardID)
	if permanent == nil || permanent.Face != game.FacePrototype {
		t.Fatalf("permanent = %+v, want the prototyped card", permanent)
	}
	assertPrototypePermanent(t, g, permanent, 3, 2, 3, []color.Color{color.Red})

	movePermanentToZone(g, permanent, zone.Graveyard)
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("card didn't reach the graveyard")
	}
	assertNormalPrototypeCard(t, g, cardID, "in the graveyard")
}

// TestPrototypeNormalCastKeepsNormalCharacteristics casts the same card for
// its normal cost: a colorless 6/4 with mana value 7 that no red-spell trigger
// sees.
func TestPrototypeNormalCastKeepsNormalCharacteristics(t *testing.T) {
	g, engine := prototypeScenario(t)
	addCombatPermanent(g, game.Player1, registryCard(t, "Dragon's Claw"))
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	fuseAddMana(g, game.Player1, slices.Repeat([]mana.Color{mana.C}, 7)...)

	obj := castPrototypeTestCard(t, g, engine, cardID, game.FaceFront)
	assertPrototypeSpell(t, g, obj, 7, nil)
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Dragon's Claw triggered on the colorless normal cast")
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	permanent := permanentForCard(g, cardID)
	if permanent == nil || permanent.Face != game.FaceFront {
		t.Fatalf("permanent = %+v, want the normal card", permanent)
	}
	assertPrototypePermanent(t, g, permanent, 6, 4, 7, nil)
}

// TestPrototypeChoiceSurvivesCloneAndReplay clones the game with the
// prototyped spell on the stack and replays the same cast on a pre-cast clone;
// every game resolves to the same prototyped permanent and event stream.
func TestPrototypeChoiceSurvivesCloneAndReplay(t *testing.T) {
	g, engine := prototypeScenario(t)
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.R)
	replay := g.Clone()

	castPrototypeTestCard(t, g, engine, cardID, game.FacePrototype)
	cloned := g.Clone()
	_, replayEngine := prototypeScenario(t)
	castPrototypeTestCard(t, replay, replayEngine, cardID, game.FacePrototype)

	_, clonedEngine := prototypeScenario(t)
	engine.resolveTopOfStack(g, &TurnLog{})
	clonedEngine.resolveTopOfStack(cloned, &TurnLog{})
	replayEngine.resolveTopOfStack(replay, &TurnLog{})
	for name, state := range map[string]*game.Game{"original": g, "clone": cloned, "replay": replay} {
		permanent := permanentForCard(state, cardID)
		if permanent == nil || permanent.Face != game.FacePrototype {
			t.Fatalf("%s permanent = %+v, want prototyped", name, permanent)
		}
		assertPrototypePermanent(t, state, permanent, 3, 2, 3, []color.Color{color.Red})
	}
	if !reflect.DeepEqual(g.Events, replay.Events) || !reflect.DeepEqual(g.Events, cloned.Events) {
		t.Fatal("event streams differ between original, clone and replay")
	}
}

// TestPrototypeSpellCopyIsPrototyped copies the prototyped spell with a "copy
// that spell" cast trigger: the copy is prototyped (CR 719.3c), resolves to a
// red 3/2 token, and copying isn't casting.
func TestPrototypeSpellCopyIsPrototyped(t *testing.T) {
	g, engine := prototypeScenario(t)
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.R)
	obj := castPrototypeTestCard(t, g, engine, cardID, game.FacePrototype)

	source := addCreaturePermanent(g, game.Player1)
	trigger := game.TriggeredAbility{Content: game.Mode{Sequence: []game.Instruction{{
		Primitive: game.CopyStackObject{Object: game.EventStackObjectReference()},
	}}}.Ability()}
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
		SourceCardID: source.CardInstanceID, Controller: game.Player1, InlineTrigger: &trigger,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{Kind: game.EventSpellCast, Controller: game.Player1, StackObjectID: obj.ID},
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, ok := g.Stack.Peek()
	if !ok || copyObj.ID == obj.ID || !copyObj.Copy || copyObj.Face != game.FacePrototype {
		t.Fatalf("top of stack = %+v, want a prototyped copy", copyObj)
	}
	assertPrototypeSpell(t, g, copyObj, 3, []color.Color{color.Red})
	engine.resolveTopOfStack(g, &TurnLog{})
	tokens := playerTokensNamed(g, game.Player1, prototypeTestName)
	if len(tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(tokens))
	}
	token, _ := g.PermanentByID(tokens[0])
	assertPrototypePermanent(t, g, token, 3, 2, 3, []color.Color{color.Red})
	if got := countEventsOfKind(g.Events, game.EventSpellCast); got != 1 {
		t.Fatalf("spell cast events = %d, want only the original cast", got)
	}
}

// TestPrototypePermanentTokenCopyIsPrototyped copies the prototyped permanent
// with Cackling Counterpart: the token copy has the prototype values (CR
// 719.3d; Blitz Automaton ruling).
func TestPrototypePermanentTokenCopyIsPrototyped(t *testing.T) {
	g, engine := prototypeScenario(t)
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.R)
	castPrototypeTestCard(t, g, engine, cardID, game.FacePrototype)
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, cardID)

	counterpart := addCardToHand(g, game.Player1, registryCard(t, "Cackling Counterpart"))
	fuseAddMana(g, game.Player1, mana.U, mana.U, mana.U)
	targets := []game.Target{game.PermanentTarget(permanent.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(counterpart, targets, 0, nil)) {
		t.Fatal("could not cast Cackling Counterpart")
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	tokens := playerTokensNamed(g, game.Player1, prototypeTestName)
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1", len(tokens))
	}
	token, _ := g.PermanentByID(tokens[0])
	assertPrototypePermanent(t, g, token, 3, 2, 3, []color.Color{color.Red})
}

type prototypeFreeCastAgent struct {
	labels *[]string
}

func (prototypeFreeCastAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return action.Pass()
}

func (a prototypeFreeCastAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	if strings.HasPrefix(request.Prompt, "Choose how to cast") {
		for _, option := range request.Options {
			*a.labels = append(*a.labels, option.Label)
		}
	}
	for _, option := range request.Options {
		if strings.Contains(option.Label, "(prototype)") {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

// TestPrototypeFreeCastUsesPrototypeManaValue lets Electrodominance (X=3) cast
// the card without paying its mana cost: prototype isn't an alternative cost,
// so the mana value 3 prototyped cast is offered and the mana value 7 normal
// cast is not (CR 719.3a; Fleshgorger ruling).
func TestPrototypeFreeCastUsesPrototypeManaValue(t *testing.T) {
	g, engine := prototypeScenario(t)
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	pushElectrodominance(g, game.PlayerTarget(game.Player2), 3)

	var labels []string
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: prototypeFreeCastAgent{labels: &labels},
	}, &TurnLog{})
	if !slices.Equal(labels, []string{prototypeTestName + " (prototype)"}) {
		t.Fatalf("free-cast face options = %q, want only the mana value 3 prototyped cast", labels)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FacePrototype {
		t.Fatalf("free-cast spell = %+v, want the prototyped card (options %q)", obj, labels)
	}
	assertPrototypeSpell(t, g, obj, 3, []color.Color{color.Red})
}

// TestPrototypeCostReductionAppliesToPrototypeCost casts the prototyped card
// under Etherium Sculptor: the reduction applies to the prototype cost, so
// {1}{R} pays it (CR 601.2f).
func TestPrototypeCostReductionAppliesToPrototypeCost(t *testing.T) {
	g, engine := prototypeScenario(t)
	addCombatPermanent(g, game.Player1, registryCard(t, "Etherium Sculptor"))
	cardID := addCardToHand(g, game.Player1, prototypeTestDef())
	fuseAddMana(g, game.Player1, mana.C, mana.R)
	if !fuseCastFaceOffered(engine.legalActions(g, game.Player1), cardID, zone.Hand, game.FacePrototype) {
		t.Fatal("reduced prototyped cast not offered")
	}
	obj := castPrototypeTestCard(t, g, engine, cardID, game.FacePrototype)
	assertPrototypeSpell(t, g, obj, 3, []color.Color{color.Red})
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatal("reduced prototype cost left mana unspent")
	}
}

// TestPrototypeCastFromGraveyardPermission casts the card prototyped from the
// graveyard with Muldrotha's artifact permission; prototype works from any zone
// the card could be cast from (Fleshgorger ruling).
func TestPrototypeCastFromGraveyardPermission(t *testing.T) {
	g, engine, source := muldrothaGame(t)
	cardID := addCardToGraveyard(g, game.Player1, prototypeTestDef())
	assertNormalPrototypeCard(t, g, cardID, "in the graveyard")
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.R)
	actions := engine.legalActions(g, game.Player1)
	if !fuseCastFaceOffered(actions, cardID, zone.Graveyard, game.FacePrototype) ||
		fuseCastFaceOffered(actions, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("graveyard casts should offer only the payable prototyped cast")
	}
	cast := action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FacePrototype, nil, 0, nil).
		WithPlayPermission(muldrothaPermission(source, types.Artifact))
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("prototyped cast from the graveyard failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FacePrototype {
		t.Fatalf("stack top = %+v, want the prototyped card", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	assertPrototypePermanent(t, g, permanentForCard(g, cardID), 3, 2, 3, []color.Color{color.Red})
}

// TestPrototypeFrontFacePermissionCoversPrototypedCast uses Emry, Lurker of
// the Loch's "You may cast that card this turn." permission, which names the
// card's front face. Prototype keeps the physical front face (CR 719.3,
// 719.5), so the permission authorizes both the normal and the prototyped
// cast, and the prototyped cast resolves to the prototyped permanent.
func TestPrototypeFrontFacePermissionCoversPrototypedCast(t *testing.T) {
	g, engine := prototypeScenario(t)
	emryDef := registryCard(t, "Emry, Lurker of the Loch")
	emry := addCombatPermanent(g, game.Player1, emryDef)
	emry.SummoningSick = false
	cardID := addCardToGraveyard(g, game.Player1, prototypeTestDef())
	activate := action.ActivateAbility(emry.ObjectID, emryDef.ActivatedAbilityIndex(0),
		[]game.Target{currentCardTarget(t, g, cardID)}, 0)
	if !engine.applyAction(g, game.Player1, activate) {
		t.Fatal("Emry activation failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !hasCastFromZoneRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("Emry's front-face cast permission was not granted")
	}
	fuseAddMana(g, game.Player1, append(slices.Repeat([]mana.Color{mana.C}, 7), mana.R)...)
	actions := engine.legalActions(g, game.Player1)
	if !fuseCastFaceOffered(actions, cardID, zone.Graveyard, game.FaceFront) ||
		!fuseCastFaceOffered(actions, cardID, zone.Graveyard, game.FacePrototype) {
		t.Fatal("Emry's permission should offer both the normal and the prototyped cast")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FacePrototype, nil, 0, nil)) {
		t.Fatal("prototyped cast through Emry's permission failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FacePrototype {
		t.Fatalf("stack top = %+v, want the prototyped card", obj)
	}
	assertPrototypeSpell(t, g, obj, 3, []color.Color{color.Red})
	engine.resolveTopOfStack(g, &TurnLog{})
	assertPrototypePermanent(t, g, permanentForCard(g, cardID), 3, 2, 3, []color.Color{color.Red})
}

// TestPrototypeInvalidProfileFailsClosed checks that a malformed prototype
// profile is never offered as a cast (not even for free from an empty cost),
// while the card's valid normal cast remains, and that catalog admission
// rejects the card even though it has no program references.
func TestPrototypeInvalidProfileFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*game.CardDef)
	}{
		{"empty cost", func(def *game.CardDef) { def.Prototype.Val.ManaCost = nil }},
		{"negative generic", func(def *game.CardDef) { def.Prototype.Val.ManaCost = cost.Mana{cost.O(-1)} }},
		{"unprintable symbol", func(def *game.CardDef) {
			def.Prototype.Val.ManaCost = cost.Mana{cost.PhyrexianGeneric(2), cost.R}
		}},
		{"star power", func(def *game.CardDef) { def.Prototype.Val.Power = game.PT{IsStar: true} }},
		{"noncreature", func(def *game.CardDef) { def.Types = []types.Card{types.Artifact} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := prototypeTestDef()
			tt.mutate(def)
			if err := cards.NewDefaultRegistry().Programs().ValidateCard(def); err == nil {
				t.Fatal("catalog admitted an invalid prototype profile")
			}
			g, engine := prototypeScenario(t)
			cardID := addCardToHand(g, game.Player1, def)
			fuseAddMana(g, game.Player1, append(slices.Repeat([]mana.Color{mana.C}, 7), mana.R)...)
			actions := engine.legalActions(g, game.Player1)
			if fuseCastFaceOffered(actions, cardID, zone.Hand, game.FacePrototype) {
				t.Fatal("invalid prototype profile offered as a cast")
			}
			if engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FacePrototype, nil, 0, nil)) {
				t.Fatal("invalid prototyped cast was applied")
			}
			if def.HasType(types.Creature) && !fuseCastFaceOffered(actions, cardID, zone.Hand, game.FaceFront) {
				t.Fatal("the payable normal cast disappeared")
			}
		})
	}
}
