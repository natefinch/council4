package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func vanillaCreature(name string, power, toughness int, keywords ...game.Keyword) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      name,
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: toughness}),
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: game.SimpleKeywords(keywords...),
			}},
		},
	}
}

func addHandCard(g *game.Game, owner game.PlayerID, def *game.CardDef) game.PlayerID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	g.Players[owner].Hand.Add(cardID)
	return owner
}

func TestObservationBattlefieldEffectivePowerAndKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bear := addCombatPermanent(g, game.Player1, vanillaCreature("Grizzly Bears", 2, 2, game.Vigilance))
	// A +1/+1 counter raises effective P/T above the printed 2/2.
	bear.Counters.Add(counter.PlusOnePlusOne, 1)

	obs := observe(g, game.Player1)
	battlefield := obs.Battlefield()
	if len(battlefield) != 1 {
		t.Fatalf("Battlefield() = %d permanents, want 1", len(battlefield))
	}
	view := battlefield[0]
	if view.Name != "Grizzly Bears" {
		t.Errorf("Name = %q, want Grizzly Bears", view.Name)
	}
	if view.Power != 3 || view.Toughness != 3 {
		t.Errorf("effective P/T = %d/%d, want 3/3", view.Power, view.Toughness)
	}
	if !view.HasKeyword(game.Vigilance) {
		t.Error("HasKeyword(Vigilance) = false, want true")
	}
	if view.HasKeyword(game.Flying) {
		t.Error("HasKeyword(Flying) = true, want false")
	}
	if view.Controller != game.Player1 {
		t.Errorf("Controller = %v, want Player1", view.Controller)
	}
}

func TestObservationOwnHandVisibleOpponentHandCountOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addHandCard(g, game.Player1, vanillaCreature("My Secret", 1, 1))
	addHandCard(g, game.Player2, vanillaCreature("Their Secret A", 1, 1))
	addHandCard(g, game.Player2, vanillaCreature("Their Secret B", 1, 1))

	obs := observe(g, game.Player1)

	hand := obs.Hand()
	if len(hand) != 1 || hand[0].Name != "My Secret" {
		t.Fatalf("Hand() = %+v, want one visible card named My Secret", hand)
	}

	// The opponent's hand is reported only as a size, never its contents.
	if got := obs.PlayerState(game.Player2).HandSize; got != 2 {
		t.Errorf("Player2 HandSize = %d, want 2", got)
	}
}

func TestObservationLifeAndCommanderState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 33
	g.Players[game.Player1].CommanderCastCount = 2
	g.Players[game.Player3].CommanderDamage = map[id.ID]int{42: 7}

	obs := observe(g, game.Player1)

	if got := obs.Life(game.Player1); got != 33 {
		t.Errorf("Life(Player1) = %d, want 33", got)
	}
	if got := obs.PlayerState(game.Player1).CommanderCastCount; got != 2 {
		t.Errorf("CommanderCastCount = %d, want 2", got)
	}
	damage := obs.PlayerState(game.Player3).CommanderDamage
	if damage[42] != 7 {
		t.Errorf("CommanderDamage[42] = %d, want 7", damage[42])
	}
	// The returned map is a copy: mutating it must not affect game state.
	damage[42] = 999
	if g.Players[game.Player3].CommanderDamage[42] != 7 {
		t.Error("mutating the observation's CommanderDamage changed game state")
	}
}

func TestObservationStackBottomToTop(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	firstID := g.IDGen.Next()
	g.CardInstances[firstID] = &game.CardInstance{ID: firstID, Def: vanillaCreature("First Spell", 1, 1), Owner: game.Player1}
	secondID := g.IDGen.Next()
	g.CardInstances[secondID] = &game.CardInstance{ID: secondID, Def: vanillaCreature("Second Spell", 1, 1), Owner: game.Player2}

	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: firstID, Controller: game.Player1})
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: secondID, Controller: game.Player2})

	obs := observe(g, game.Player1)
	stack := obs.Stack()
	if len(stack) != 2 {
		t.Fatalf("Stack() = %d objects, want 2", len(stack))
	}
	if stack[0].Name != "First Spell" || stack[1].Name != "Second Spell" {
		t.Errorf("stack order = [%q, %q], want bottom-to-top [First Spell, Second Spell]", stack[0].Name, stack[1].Name)
	}
	if stack[1].Controller != game.Player2 {
		t.Errorf("top controller = %v, want Player2", stack[1].Controller)
	}
}

func TestObservationFaceDownHidesIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A morph face-down permanent is a public 2/2 with no name and no keywords.
	morph := addFaceDownPermanent(g, game.Player2, vanillaCreature("Hidden Bomb", 7, 7, game.Flying), game.FaceDownMorph)

	obs := observe(g, game.Player1)
	view := findPermanentView(t, obs, morph.ObjectID)
	if view.Name != "" {
		t.Errorf("face-down Name = %q, want empty (identity hidden)", view.Name)
	}
	if view.Power != 2 || view.Toughness != 2 {
		t.Errorf("face-down P/T = %d/%d, want 2/2", view.Power, view.Toughness)
	}
	if view.HasKeyword(game.Flying) {
		t.Error("face-down morph leaked the hidden Flying keyword")
	}
	if !view.FaceDown {
		t.Error("FaceDown = false, want true")
	}
}

func TestObservationDisguiseFaceDownShowsPublicWard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A disguise (or cloak) face-down permanent hides its identity but publicly
	// has Ward {2}.
	disguised := addFaceDownPermanent(g, game.Player2, disguiseCreature(cost.Mana{cost.W}), game.FaceDownDisguise)

	obs := observe(g, game.Player1)
	view := findPermanentView(t, obs, disguised.ObjectID)
	if view.Name != "" {
		t.Errorf("disguised Name = %q, want empty (identity hidden)", view.Name)
	}
	if !view.HasKeyword(game.Ward) {
		t.Error("disguised face-down HasKeyword(Ward) = false, want true (Ward is public)")
	}
}

func TestObservationStackFaceDownSpellHidesIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A face-down (morph) spell keeps its real source card on the stack, but its
	// printed identity must stay hidden from every observer.
	hiddenID := g.IDGen.Next()
	g.CardInstances[hiddenID] = &game.CardInstance{ID: hiddenID, Def: vanillaCreature("Hidden Bomb", 7, 7, game.Flying), Owner: game.Player2}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   hiddenID,
		Controller: game.Player2,
		FaceDown:   true,
	})

	obs := observe(g, game.Player1)
	stack := obs.Stack()
	if len(stack) != 1 {
		t.Fatalf("Stack() = %d objects, want 1", len(stack))
	}
	if stack[0].Name != "" {
		t.Errorf("face-down stack Name = %q, want empty (identity hidden)", stack[0].Name)
	}
	if !stack[0].FaceDown {
		t.Error("FaceDown = false, want true")
	}
}

func findPermanentView(t *testing.T, obs PlayerObservation, objectID id.ID) PermanentView {
	t.Helper()
	for _, view := range obs.Battlefield() {
		if view.ObjectID == objectID {
			return view
		}
	}
	t.Fatalf("permanent %v not found in observation battlefield", objectID)
	return PermanentView{}
}
