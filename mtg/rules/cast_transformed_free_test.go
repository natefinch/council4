package rules

import (
	"slices"
	"strings"
	"testing"

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

// Source review: issue #3647 (sub-issue of #3596), casting a transforming
// double-faced card transformed through the resolution-time free-cast route.
// Official CR effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// verified local copy), read 202.3b, 601.3, 707.10, 707.10g, 712.8a, 712.8c,
// 712.8e, 712.11, 712.11a, 712.11c, 712.11d. 712.11a: a double-faced card or a
// copy of one cast "transformed" is put on the stack with its back face up;
// 712.11c: only the face that will be up is evaluated to determine whether it
// can be cast; 712.8c: such a spell has only its back face's characteristics
// but its mana value uses its front face's mana cost; 202.3b: a copy of the
// back face of a nonmodal double-faced object has mana value 0; 712.8a: off the
// stack and battlefield the card has only its front face's characteristics.
// Cards (pinned corpus): Invasion of Segovia // Caetus, Sea Tyrant of Segovia
// (oracle 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, {2}{U} Battle — Siege; back
// Legendary Creature — Serpent 3/3, blue color indicator, "Noncreature spells
// you cast have convoke. At the beginning of your end step, untap up to four
// target creatures."); Invasion of Kylem // Valor's Reach Tag Team (oracle
// fee0aa18-5103-4814-94e3-87ad3eaf5a6f, {2}{R}{W} Battle — Siege; back Sorcery,
// red and white color indicator, "Create two 3/2 red and white Warrior creature
// tokens ..."); Counterspell (cc187110-1148-4090-bbb8-e205694a39f5, "Counter
// target spell."); Twincast (8f878efc-850f-43d2-a6fe-5ea8d1dd5afb, "Copy target
// instant or sorcery spell. You may choose new targets for the copy."); Beast
// Whisperer (5da7eea8-bb9e-47ce-a554-8a1ee058bd7a, "Whenever you cast a creature
// spell, draw a card."); Grizzly Bears (14c8f55d-d177-4c25-a931-ebeb9e6062a0).
// Scryfall rulings 2023-04-14 (Invasion of Kylem, Invasion of Segovia): a
// Siege's intrinsic trigger is "exile it, then you may cast it transformed
// without paying its mana cost"; a card that isn't represented by a
// transforming double-faced card can't be cast that way and remains in exile.
// The Siege trigger, battle state-based actions and counter-removal events are
// later sub-issues (#3648, #3649).

type transformedCastScene struct {
	g      *game.Game
	engine *Engine
	cardID id.ID
}

func newTransformedCastScene(t *testing.T, name string) transformedCastScene {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return transformedCastScene{
		g:      g,
		engine: choicesProgramEngine(t),
		cardID: addCardToExile(g, game.Player1, registryCard(t, name)),
	}
}

func (s transformedCastScene) castWith(agents [game.NumPlayers]PlayerAgent, selection game.Selection, maxManaValue opt.V[int], spell freeCastSpell) bool {
	spell.transformed = true
	return s.engine.chooseAndCastFreeSpell(s.g, s.g.Players[game.Player1], game.Player1, s.cardID, zone.Exile,
		selection, maxManaValue, spell, true, agents, &TurnLog{})
}

func (s transformedCastScene) cast(t *testing.T) *game.StackObject {
	t.Helper()
	if !s.castWith([game.NumPlayers]PlayerAgent{}, game.Selection{}, opt.V[int]{}, freeCastSpell{}) {
		t.Fatal("casting the card transformed without paying its mana cost = false, want true")
	}
	obj, ok := s.g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell {
		t.Fatal("no spell on the stack after casting transformed")
	}
	return obj
}

func (s transformedCastScene) resolveAll() {
	for !s.g.Stack.IsEmpty() {
		s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
}

func assertTransformedSpell(t *testing.T, g *game.Game, obj *game.StackObject, name string, cardTypes []types.Card, manaValue int, colors []color.Color) {
	t.Helper()
	if obj.Face != game.FaceBack {
		t.Fatalf("stack face = %v, want back face up", obj.Face)
	}
	def, ok := stackSpellCharacteristicsDef(g, obj)
	if !ok || def.Name != name || !slices.Equal(def.Types, cardTypes) {
		t.Fatalf("stack spell = %v %v, want %s %v", def.Name, def.Types, name, cardTypes)
	}
	if got, ok := stackObjectManaValue(g, obj); !ok || got != manaValue {
		t.Fatalf("stack mana value = %d, %v; want front-face %d (CR 712.8c)", got, ok, manaValue)
	}
	if got, _ := stackObjectColors(g, obj); !slices.Equal(got, colors) {
		t.Fatalf("stack colors = %v, want color indicator %v", got, colors)
	}
	if obj.ManaSpentToCast != 0 {
		t.Fatalf("mana spent = %d, want 0", obj.ManaSpentToCast)
	}
}

func warriorTokens(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == controller && permanent.TokenDef != nil && permanent.TokenDef.Name == "Warrior" {
			count++
		}
	}
	return count
}

// TestCastTransformedCaetusResolvesBackFaceUp is acceptance scenario 1.
func TestCastTransformedCaetusResolvesBackFaceUp(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Segovia")
	obj := s.cast(t)
	assertTransformedSpell(t, s.g, obj, "Caetus, Sea Tyrant of Segovia", []types.Card{types.Creature}, 3, []color.Color{color.Blue})
	if def, _ := stackSpellCharacteristicsDef(s.g, obj); !slices.Contains(def.Supertypes, types.Legendary) {
		t.Fatalf("stack supertypes = %v, want legendary", def.Supertypes)
	}
	if s.g.Players[game.Player1].Exile.Contains(s.cardID) {
		t.Fatal("the cast card is still in exile")
	}
	event := lastSpellCastEvent(t, s.g)
	if event.Face != game.FaceBack || !slices.Contains(event.CardTypes, types.Creature) ||
		event.ManaValue != opt.Val(3) || !slices.Equal(event.Colors, []color.Color{color.Blue}) ||
		event.FromZone != zone.Exile {
		t.Fatalf("spell cast event = face %v types %v mana value %+v colors %v from %v; want back-face creature, 3, blue, exile",
			event.Face, event.CardTypes, event.ManaValue, event.Colors, event.FromZone)
	}
	if got := spellsCastThisTurn(s.g, game.Player1); got != 1 {
		t.Fatalf("spells cast this turn (storm count) = %d, want 1", got)
	}
	views := NewObservation(s.g, game.Player2).Stack()
	if top := views[len(views)-1]; top.Name != "Caetus, Sea Tyrant of Segovia" || top.ManaValue != 3 || !slices.Equal(top.Types, []types.Card{types.Creature}) {
		t.Fatalf("opponent stack view = %+v, want Caetus creature with mana value 3", top)
	}

	s.resolveAll()
	caetus := permanentForCard(s.g, s.cardID)
	if caetus == nil || caetus.Face != game.FaceBack || !caetus.Transformed || !caetus.EnteredFromCast {
		t.Fatalf("Caetus permanent = %+v, want cast and transformed back face up", caetus)
	}
	if got, ok := effectivePermanentManaValue(s.g, caetus); !ok || got != 3 {
		t.Fatalf("Caetus permanent mana value = %d, %v; want 3", got, ok)
	}
	if len(s.g.Battlefield) != 1 {
		t.Fatalf("battlefield has %d permanents, want only Caetus (no Kraken from the front face)", len(s.g.Battlefield))
	}

	// Caetus's static grants convoke to noncreature spells only.
	bears := addCombatPermanent(s.g, game.Player1, registryCard(t, "Grizzly Bears"))
	creatureID := addCardToHand(s.g, game.Player1, plainSpell("Creature Spell", types.Creature, cost.Mana{cost.O(1), cost.U}))
	if s.engine.applyAction(s.g, game.Player1, action.CastSpell(creatureID, nil, 0, nil)) {
		t.Fatal("creature spell cast by convoke, want no convoke for creature spells")
	}
	sorceryID := addCardToHand(s.g, game.Player1, plainSpell("Noncreature Spell", types.Sorcery, cost.Mana{cost.O(1), cost.U}))
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(sorceryID, nil, 0, nil)) {
		t.Fatal("noncreature spell with Caetus's convoke grant could not be cast by tapping creatures")
	}
	if !caetus.Tapped || !bears.Tapped {
		t.Fatalf("convoke tapped Caetus=%v Bears=%v, want both", caetus.Tapped, bears.Tapped)
	}
}

// TestCastTransformedSorceryResolvesToGraveyard is acceptance scenario 2.
func TestCastTransformedSorceryResolvesToGraveyard(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Kylem")
	obj := s.cast(t)
	assertTransformedSpell(t, s.g, obj, "Valor's Reach Tag Team", []types.Card{types.Sorcery}, 4, []color.Color{color.Red, color.White})
	s.resolveAll()
	if got := warriorTokens(s.g, game.Player1); got != 2 {
		t.Fatalf("Warrior tokens = %d, want 2", got)
	}
	if !s.g.Players[game.Player1].Graveyard.Contains(s.cardID) {
		t.Fatal("Invasion of Kylem is not in its owner's graveyard after resolving")
	}
	if card, _ := s.g.GetCardInstance(s.cardID); card.Def.Name != "Invasion of Kylem" {
		t.Fatalf("graveyard card name = %s, want front face", card.Def.Name)
	}
}

// TestCounterspellCountersTransformedSorcery is acceptance scenario 3.
func TestCounterspellCountersTransformedSorcery(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Kylem")
	obj := s.cast(t)
	counterID := addCardToHand(s.g, game.Player2, registryCard(t, "Counterspell"))
	fuseAddMana(s.g, game.Player2, mana.U, mana.U)
	s.g.Turn.PriorityPlayer = game.Player2
	if !s.engine.applyAction(s.g, game.Player2, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(obj.ID)}, 0, nil)) {
		t.Fatal("Counterspell targeting the transformed spell = false, want true")
	}
	s.resolveAll()
	if got := warriorTokens(s.g, game.Player1); got != 0 {
		t.Fatalf("Warrior tokens = %d, want 0 after the spell was countered", got)
	}
	if !s.g.Players[game.Player1].Graveyard.Contains(s.cardID) {
		t.Fatal("countered Invasion of Kylem is not in its owner's graveyard")
	}
}

// TestTwincastCopiesTransformedSorcery is acceptance scenario 4. The copy has
// the back face's characteristics; as a copy of a back face its mana value is 0
// (CR 202.3b).
func TestTwincastCopiesTransformedSorcery(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Kylem")
	obj := s.cast(t)
	twincastID := addCardToHand(s.g, game.Player1, registryCard(t, "Twincast"))
	fuseAddMana(s.g, game.Player1, mana.U, mana.U)
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(twincastID, []game.Target{game.StackObjectTarget(obj.ID)}, 0, nil)) {
		t.Fatal("Twincast targeting the transformed spell = false, want true")
	}
	s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	copyObj, ok := s.g.Stack.Peek()
	if !ok || !copyObj.Copy || copyObj.ID == obj.ID {
		t.Fatalf("top of stack = %+v, want Twincast's copy", copyObj)
	}
	def, ok := stackSpellCharacteristicsDef(s.g, copyObj)
	if !ok || def.Name != "Valor's Reach Tag Team" || !slices.Equal(def.Types, []types.Card{types.Sorcery}) {
		t.Fatalf("copy = %v %v, want the Valor's Reach Tag Team sorcery", def.Name, def.Types)
	}
	if got, _ := stackObjectColors(s.g, copyObj); !slices.Equal(got, []color.Color{color.Red, color.White}) {
		t.Fatalf("copy colors = %v, want red and white", got)
	}
	if got, ok := stackObjectManaValue(s.g, copyObj); !ok || got != 0 {
		t.Fatalf("copy mana value = %d, %v; want 0 (CR 202.3b)", got, ok)
	}
	if got, _ := stackObjectManaValue(s.g, obj); got != 4 {
		t.Fatalf("original mana value = %d, want 4", got)
	}
	s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := warriorTokens(s.g, game.Player1); got != 2 {
		t.Fatalf("Warrior tokens after the copy resolved = %d, want 2", got)
	}
	s.resolveAll()
	if got := warriorTokens(s.g, game.Player1); got != 4 {
		t.Fatalf("Warrior tokens after both resolved = %d, want 4", got)
	}
	if !s.g.Players[game.Player1].Graveyard.Contains(s.cardID) {
		t.Fatal("Invasion of Kylem is not in its owner's graveyard")
	}
}

// TestBeastWhispererSeesTransformedCreatureSpell is acceptance scenario 5.
func TestBeastWhispererSeesTransformedCreatureSpell(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Segovia")
	whisperer := addCombatPermanent(s.g, game.Player1, registryCard(t, "Beast Whisperer"))
	addCardToLibrary(s.g, game.Player1, registryCard(t, "Grizzly Bears"))
	s.cast(t)
	if !s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) ||
		triggersFromSource(s.g, whisperer.ObjectID) != 1 {
		t.Fatal("Beast Whisperer did not trigger for the transformed creature spell")
	}
	handBefore := s.g.Players[game.Player1].Hand.Size()
	s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := s.g.Players[game.Player1].Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand size = %d, want %d after Beast Whisperer's draw", got, handBefore+1)
	}
}

// TestBeastWhispererIgnoresTransformedSorcery proves the cast trigger reads the
// back face: a battle card cast as its sorcery back face isn't a creature spell.
func TestBeastWhispererIgnoresTransformedSorcery(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Kylem")
	whisperer := addCombatPermanent(s.g, game.Player1, registryCard(t, "Beast Whisperer"))
	s.cast(t)
	s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := triggersFromSource(s.g, whisperer.ObjectID); got != 0 {
		t.Fatalf("Beast Whisperer triggers = %d, want 0 for a sorcery spell", got)
	}
}

// TestCascadeUsesTransformedSpellFrontManaValue proves cascade reads the cast
// spell's front-face mana value (CR 712.8c): a card with mana value 3 is passed
// over and one with mana value 2 is found.
func TestCascadeUsesTransformedSpellFrontManaValue(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Segovia")
	addCombatPermanent(s.g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cascade Granter", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{cascadeGrantStatic("Creature spells you cast have cascade.",
			game.Selection{RequiredTypes: []types.Card{types.Creature}}, zone.None)},
	}})
	hit := addCardToLibrary(s.g, game.Player1, simpleGainLifeInstantWithManaValue("Two", 2))
	tooBig := addCardToLibrary(s.g, game.Player1, simpleGainLifeInstantWithManaValue("Three", 3))
	spell := s.cast(t)
	if event := lastSpellCastEvent(t, s.g); event.CascadeInstances != opt.Val(1) {
		t.Fatalf("cast event cascade = %+v, want 1", event.CascadeInstances)
	}
	s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := cascadeTriggersFromSpell(s.g, spell); got != 1 {
		t.Fatalf("cascade triggers = %d, want 1", got)
	}
	s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !stackContainsSource(s.g, hit) {
		t.Fatal("cascade did not cast the mana value 2 card")
	}
	if stackContainsSource(s.g, tooBig) || !s.g.Players[game.Player1].Library.Contains(tooBig) {
		t.Fatal("cascade did not pass over the mana value 3 card to the library bottom")
	}
}

// TestDeclineTransformedCastLeavesCardInExile is acceptance scenario 6.
func TestDeclineTransformedCastLeavesCardInExile(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Segovia")
	var offered []string
	script := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if !strings.HasPrefix(request.Prompt, "Choose how to cast") {
			return fallbackChoice(request)
		}
		for _, option := range request.Options {
			offered = append(offered, option.Label)
		}
		return []int{len(request.Options) - 1}
	}}
	if s.castWith([game.NumPlayers]PlayerAgent{game.Player1: script}, game.Selection{}, opt.V[int]{}, freeCastSpell{}) {
		t.Fatal("declined transformed cast = true, want false")
	}
	if len(offered) != 2 || !strings.HasPrefix(offered[0], "Caetus, Sea Tyrant of Segovia") || !strings.HasPrefix(offered[1], "Don't cast") {
		t.Fatalf("offered = %q, want only the back face and declining", offered)
	}
	if !s.g.Stack.IsEmpty() || !s.g.Players[game.Player1].Exile.Contains(s.cardID) {
		t.Fatal("declining did not leave the card in exile")
	}
}

// TestCastTransformedEvaluatesOnlyBackFace proves only the back face is
// evaluated to determine whether it can be cast (CR 712.11c), using the front
// face's mana value for a mana value limit (CR 712.8c), and that a card with no
// transforming back face can't be cast transformed and stays in exile.
func TestCastTransformedEvaluatesOnlyBackFace(t *testing.T) {
	creatureOnly := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	for _, tc := range []struct {
		name      string
		card      string
		selection game.Selection
		max       opt.V[int]
		want      bool
	}{
		{"creature back face matches a creature selection", "Invasion of Segovia", creatureOnly, opt.V[int]{}, true},
		{"sorcery back face fails a creature selection", "Invasion of Kylem", creatureOnly, opt.V[int]{}, false},
		{"front mana value 3 within limit 3", "Invasion of Segovia", game.Selection{}, opt.Val(3), true},
		{"front mana value 3 exceeds limit 2", "Invasion of Segovia", game.Selection{}, opt.Val(2), false},
		{"selection reads front mana value", "Invasion of Kylem",
			game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}, opt.V[int]{}, true},
		{"single-faced card can't be cast transformed", "Grizzly Bears", game.Selection{}, opt.V[int]{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTransformedCastScene(t, tc.card)
			got := s.castWith([game.NumPlayers]PlayerAgent{}, tc.selection, tc.max, freeCastSpell{})
			if got != tc.want {
				t.Fatalf("cast = %v, want %v", got, tc.want)
			}
			if got {
				if obj, _ := s.g.Stack.Peek(); obj.Face != game.FaceBack {
					t.Fatalf("stack face = %v, want back", obj.Face)
				}
				return
			}
			if !s.g.Stack.IsEmpty() || !s.g.Players[game.Player1].Exile.Contains(s.cardID) {
				t.Fatal("an uncastable card did not stay in exile")
			}
		})
	}
}

// TestCastTransformedCopyOfCard covers casting a copy of a card transformed
// (CR 712.11a): the copy is cast with its back face up and the front mana
// value, while the card itself stays in exile.
func TestCastTransformedCopyOfCard(t *testing.T) {
	t.Run("sorcery", func(t *testing.T) {
		s := newTransformedCastScene(t, "Invasion of Kylem")
		if !s.castWith([game.NumPlayers]PlayerAgent{}, game.Selection{}, opt.V[int]{}, freeCastSpell{copyCard: true}) {
			t.Fatal("casting a copy transformed = false, want true")
		}
		obj, _ := s.g.Stack.Peek()
		if !obj.Copy || !obj.CastCopy {
			t.Fatalf("stack object = %+v, want a cast copy", obj)
		}
		assertTransformedSpell(t, s.g, obj, "Valor's Reach Tag Team", []types.Card{types.Sorcery}, 4, []color.Color{color.Red, color.White})
		if event := lastSpellCastEvent(t, s.g); event.ManaValue != opt.Val(4) || event.Face != game.FaceBack {
			t.Fatalf("cast event mana value %+v face %v, want 4 back", event.ManaValue, event.Face)
		}
		s.resolveAll()
		if got := warriorTokens(s.g, game.Player1); got != 2 {
			t.Fatalf("Warrior tokens = %d, want 2", got)
		}
		if !s.g.Players[game.Player1].Exile.Contains(s.cardID) {
			t.Fatal("the copied card left exile")
		}
	})
	t.Run("creature", func(t *testing.T) {
		s := newTransformedCastScene(t, "Invasion of Segovia")
		if !s.castWith([game.NumPlayers]PlayerAgent{}, game.Selection{}, opt.V[int]{}, freeCastSpell{copyCard: true}) {
			t.Fatal("casting a copy transformed = false, want true")
		}
		obj, _ := s.g.Stack.Peek()
		assertTransformedSpell(t, s.g, obj, "Caetus, Sea Tyrant of Segovia", []types.Card{types.Creature}, 3, []color.Color{color.Blue})
		s.resolveAll()
		if len(s.g.Battlefield) != 1 {
			t.Fatalf("battlefield has %d permanents, want only the Caetus token", len(s.g.Battlefield))
		}
		token := s.g.Battlefield[0]
		def, _ := permanentCopyDef(s.g, token)
		if !token.Token || token.Face != game.FaceBack || !token.Transformed || def.Name != "Caetus, Sea Tyrant of Segovia" {
			t.Fatalf("token = %+v (%s), want a transformed Caetus token", token, def.Name)
		}
		if !s.g.Players[game.Player1].Exile.Contains(s.cardID) {
			t.Fatal("the copied card left exile")
		}
	})
}

// TestTransformedSpellSurvivesCloneReplay proves the back-face-up stack object
// survives cloning and resolves identically on the clone.
func TestTransformedSpellSurvivesCloneReplay(t *testing.T) {
	s := newTransformedCastScene(t, "Invasion of Segovia")
	obj := s.cast(t)
	replay := transformedCastScene{g: s.g.Clone(), engine: s.engine, cardID: s.cardID}
	cloned, ok := replay.g.Stack.Peek()
	if !ok || cloned == obj || cloned.ID != obj.ID {
		t.Fatal("clone did not deep-copy the stack object")
	}
	assertTransformedSpell(t, replay.g, cloned, "Caetus, Sea Tyrant of Segovia", []types.Card{types.Creature}, 3, []color.Color{color.Blue})
	s.resolveAll()
	replay.resolveAll()
	for _, g := range []*game.Game{s.g, replay.g} {
		if permanent := permanentForCard(g, s.cardID); permanent == nil || permanent.Face != game.FaceBack || !permanent.Transformed {
			t.Fatalf("resolved permanent = %+v, want transformed Caetus", permanent)
		}
	}
	if len(s.g.Events) != len(replay.g.Events) {
		t.Fatalf("clone replay diverged: %d events, want %d", len(replay.g.Events), len(s.g.Events))
	}
}

// TestTransformingCardFromHandOffersOnlyFrontFace is acceptance scenario 7: a
// transforming double-faced card is cast with its front face up by default
// (CR 712.11), both normally and without paying its mana cost.
func TestTransformingCardFromHandOffersOnlyFrontFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := choicesProgramEngine(t)
	cardID := addCardToHand(g, game.Player1, registryCard(t, "Invasion of Segovia"))
	fuseAddMana(g, game.Player1, mana.U, mana.U, mana.U)
	var faces []game.FaceIndex
	for _, act := range engine.legalActions(g, game.Player1) {
		if cast, ok := act.CastSpellPayload(); ok && cast.CardID == cardID {
			faces = append(faces, cast.Face)
		}
	}
	if len(faces) == 0 || slices.ContainsFunc(faces, func(face game.FaceIndex) bool { return face != game.FaceFront }) {
		t.Fatalf("hand cast faces = %v, want only the front face", faces)
	}
	card, _ := g.GetCardInstance(cardID)
	options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Hand, game.Selection{}, opt.V[int]{})
	if len(options) == 0 {
		t.Fatal("no free-cast option from hand")
	}
	for _, option := range options {
		if option.cast.Face != game.FaceFront {
			t.Fatalf("free-cast face = %v, want front", option.cast.Face)
		}
	}
}
