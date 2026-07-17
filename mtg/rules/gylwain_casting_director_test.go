package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	carde "github.com/natefinch/council4/mtg/cards/e"
	cardg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func roleTokenFromInstruction(t *testing.T, instruction game.Instruction) *game.CardDef {
	t.Helper()
	create, ok := instruction.Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("instruction primitive = %T, want game.CreateToken", instruction.Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatalf("token source = %#v, want concrete definition", create.Source)
	}
	return def
}

func gylwainRoleDef(t *testing.T, mode int) *game.CardDef {
	t.Helper()
	def := cardg.GylwainCastingDirector()
	return roleTokenFromInstruction(t, def.TriggeredAbilities[0].Content.Modes[mode].Sequence[0])
}

func cursedRoleDef(t *testing.T) *game.CardDef {
	t.Helper()
	def := cardc.CursedCourtier()
	return roleTokenFromInstruction(t, def.TriggeredAbilities[0].Content.Modes[0].Sequence[0])
}

func wickedRoleDef(t *testing.T) *game.CardDef {
	t.Helper()
	def := cardc.CharmingScoundrel()
	return roleTokenFromInstruction(t, def.TriggeredAbilities[0].Content.Modes[2].Sequence[0])
}

func youngHeroRoleDef(t *testing.T) *game.CardDef {
	t.Helper()
	def := carde.EmberethVeteran()
	return roleTokenFromInstruction(t, def.ActivatedAbilities[0].Content.Modes[0].Sequence[0])
}

func roleAttachedTo(g *game.Game, target *game.Permanent) *game.Permanent {
	for _, attachmentID := range target.Attachments {
		attachment, ok := permanentByObjectID(g, attachmentID)
		if ok && permanentHasSubtype(g, attachment, types.Role) {
			return attachment
		}
	}
	return nil
}

func emitGylwainEntry(g *game.Game, permanent *game.Permanent, controller game.PlayerID, simultaneousID game.ObjectID) {
	emitEvent(g, game.Event{
		Kind:           game.EventPermanentEnteredBattlefield,
		Controller:     controller,
		PermanentID:    permanent.ObjectID,
		CardID:         permanent.CardInstanceID,
		SimultaneousID: simultaneousID,
	})
}

func TestGylwainCreatesEachRoleForEventPermanentWithTriggerController(t *testing.T) {
	for mode, wantName := range []string{"Royal Role", "Sorcerer Role", "Monster Role"} {
		t.Run(wantName, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			gylwain := addCombatPermanent(g, game.Player1, cardg.GylwainCastingDirector())
			entering := addRoleTestCreature(g, game.Player1, "Entering Creature")
			emitGylwainEntry(g, entering, game.Player1, 0)
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{{mode}}},
			}
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
				t.Fatal("Gylwain trigger was not put on the stack")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Controller != game.Player1 ||
				trigger.TriggerEvent.PermanentID != entering.ObjectID {
				t.Fatalf("trigger = %#v", trigger)
			}

			gylwain.Controller = game.Player3
			entering.Controller = game.Player2
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			role := roleAttachedTo(g, entering)
			if role == nil || role.TokenDef == nil || role.TokenDef.Name != wantName {
				t.Fatalf("attached Role = %#v, want %s", role, wantName)
			}
			if role.Controller != game.Player1 || role.Owner != game.Player1 {
				t.Fatalf("Role controller/owner = %v/%v, want Player1 snapshot", role.Controller, role.Owner)
			}
		})
	}
}

func TestGylwainEntryQualifiers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gylwain := addCombatPermanent(g, game.Player1, cardg.GylwainCastingDirector())
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {1}}},
	}

	emitGylwainEntry(g, gylwain, game.Player1, 0)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gylwain did not trigger for itself")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if role := roleAttachedTo(g, gylwain); role == nil || role.TokenDef.Name != "Royal Role" {
		t.Fatalf("self-attached Role = %#v", role)
	}

	controlled := addRoleTestCreature(g, game.Player1, "Controlled Nontoken")
	emitGylwainEntry(g, controlled, game.Player1, 0)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gylwain did not trigger for controlled nontoken creature")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if role := roleAttachedTo(g, controlled); role == nil || role.TokenDef.Name != "Sorcerer Role" {
		t.Fatalf("controlled creature Role = %#v", role)
	}

	token, ok := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Creature Token",
		Types: []types.Card{types.Creature},
	}})
	if !ok {
		t.Fatal("failed to create creature token")
	}
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: token.ObjectID,
		TokenName:   token.TokenDef.Name,
		TokenDef:    token.TokenDef,
	})
	opponent := addRoleTestCreature(g, game.Player2, "Opponent Nontoken")
	emitGylwainEntry(g, opponent, game.Player2, 0)
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gylwain triggered for a token or opponent-controlled creature")
	}
}

func TestGylwainSimultaneousEntriesKeepEventAttachmentsDistinct(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardg.GylwainCastingDirector())
	first := addRoleTestCreature(g, game.Player1, "First")
	second := addRoleTestCreature(g, game.Player1, "Second")
	batch := g.IDGen.Next()
	emitGylwainEntry(g, first, game.Player1, batch)
	emitGylwainEntry(g, second, game.Player1, batch)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {2}}},
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) || g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want two Gylwain triggers", g.Stack.Size())
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	firstRole := roleAttachedTo(g, first)
	secondRole := roleAttachedTo(g, second)
	if firstRole == nil || secondRole == nil || firstRole.ObjectID == secondRole.ObjectID {
		t.Fatalf("batch Roles = %#v / %#v", firstRole, secondRole)
	}
	if !firstRole.AttachedTo.Exists || firstRole.AttachedTo.Val != first.ObjectID ||
		!secondRole.AttachedTo.Exists || secondRole.AttachedTo.Val != second.ObjectID {
		t.Fatalf("batch attachments = %v / %v", firstRole.AttachedTo, secondRole.AttachedTo)
	}
}

func TestGylwainDoesNotCreateRoleWhenEventPermanentLeft(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardg.GylwainCastingDirector())
	entering := addRoleTestCreature(g, game.Player1, "Departing Creature")
	emitGylwainEntry(g, entering, game.Player1, 0)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Gylwain trigger was not put on the stack")
	}
	if !movePermanentToZone(g, entering, zone.Graveyard) {
		t.Fatal("failed to remove entering creature")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := countTokenPermanentsNamed(g, "Royal Role"); got != 0 {
		t.Fatalf("Royal Role tokens = %d, want 0 after attachment failed", got)
	}
}

type seededGylwainAgent struct {
	rng *rand.Rand
}

func (*seededGylwainAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *seededGylwainAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceModal || len(request.Options) == 0 {
		return nil
	}
	return []int{request.Options[a.rng.IntN(len(request.Options))].Index}
}

func randomGylwainRoleSequence(t *testing.T, seed uint64) []string {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardg.GylwainCastingDirector())
	agent := &seededGylwainAgent{rng: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	var names []string
	for i := range 30 {
		entering := addRoleTestCreature(g, game.Player1, "Random Role Target")
		emitGylwainEntry(g, entering, game.Player1, 0)
		if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
			t.Fatalf("entry %d did not trigger Gylwain", i)
		}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		role := roleAttachedTo(g, entering)
		if role == nil || role.TokenDef == nil {
			t.Fatalf("entry %d created no attached Role", i)
		}
		names = append(names, role.TokenDef.Name)
	}
	return names
}

func TestGylwainSeededRandomChoiceIsReproducibleAndUsesEveryMode(t *testing.T) {
	first := randomGylwainRoleSequence(t, 42)
	second := randomGylwainRoleSequence(t, 42)
	if !slices.Equal(first, second) {
		t.Fatalf("same seed diverged:\n%v\n%v", first, second)
	}
	for _, want := range []string{"Royal Role", "Sorcerer Role", "Monster Role"} {
		if !slices.Contains(first, want) {
			t.Fatalf("seeded choices never created %s: %v", want, first)
		}
	}
}

func TestRoleDefinitionsApplyStaticAndGrantedAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	cursedTarget := addCombatPermanent(g, game.Player1, vanillaCreature("Cursed Target", 4, 5))
	cursed, _ := createTokenPermanent(g, game.Player1, cursedRoleDef(t))
	attachPermanent(g, cursed, cursedTarget)
	if power, toughness := effectivePower(g, cursedTarget), roleEffectiveToughness(t, g, cursedTarget); power != 1 || toughness != 1 {
		t.Fatalf("Cursed target = %d/%d, want 1/1", power, toughness)
	}

	monsterTarget := addRoleTestCreature(g, game.Player1, "Monster Target")
	monster, _ := createTokenPermanent(g, game.Player1, gylwainRoleDef(t, 2))
	attachPermanent(g, monster, monsterTarget)
	if power, toughness := effectivePower(g, monsterTarget), roleEffectiveToughness(t, g, monsterTarget); power != 3 || toughness != 3 {
		t.Fatalf("Monster target = %d/%d, want 3/3", power, toughness)
	}
	if !hasKeyword(g, monsterTarget, game.Trample) {
		t.Fatal("Monster Role did not grant trample")
	}

	royalTarget := addRoleTestCreature(g, game.Player1, "Royal Target")
	royal, _ := createTokenPermanent(g, game.Player1, gylwainRoleDef(t, 0))
	attachPermanent(g, royal, royalTarget)
	if power, toughness := effectivePower(g, royalTarget), roleEffectiveToughness(t, g, royalTarget); power != 3 || toughness != 3 {
		t.Fatalf("Royal target = %d/%d, want 3/3", power, toughness)
	}
	if !hasKeyword(g, royalTarget, game.Ward) {
		t.Fatal("Royal Role did not grant ward")
	}
}

func roleEffectiveToughness(t *testing.T, g *game.Game, permanent *game.Permanent) int {
	t.Helper()
	value, ok := effectiveToughness(g, permanent)
	if !ok {
		t.Fatalf("%s has no effective toughness", permanentName(g, permanent))
	}
	return value
}

func TestSorcererAndYoungHeroRolesGrantAttackTriggers(t *testing.T) {
	t.Run("Sorcerer scries", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		attacker := addRoleTestCreature(g, game.Player1, "Sorcerer")
		role, _ := createTokenPermanent(g, game.Player1, gylwainRoleDef(t, 1))
		attachPermanent(g, role, attacker)
		if power, toughness := effectivePower(g, attacker), roleEffectiveToughness(t, g, attacker); power != 3 || toughness != 3 {
			t.Fatalf("Sorcerer target = %d/%d, want 3/3", power, toughness)
		}
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		emitEvent(g, game.Event{
			Kind:        game.EventAttackerDeclared,
			Controller:  game.Player1,
			PermanentID: attacker.ObjectID,
		})
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("Sorcerer Role attack trigger was not put on the stack")
		}
		agents := [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
		}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		assertEvent(t, g.Events, game.EventScry, func(event game.Event) bool {
			return event.Player == game.Player1 && event.Amount == 1
		})
	})

	t.Run("Young Hero checks toughness each time", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Young Hero", 2, 3))
		role, _ := createTokenPermanent(g, game.Player1, youngHeroRoleDef(t))
		attachPermanent(g, role, attacker)
		attack := func() bool {
			emitEvent(g, game.Event{
				Kind:        game.EventAttackerDeclared,
				Controller:  game.Player1,
				PermanentID: attacker.ObjectID,
			})
			return engine.putTriggeredAbilitiesOnStack(g)
		}
		if !attack() {
			t.Fatal("Young Hero did not trigger at toughness 3")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := attacker.Counters.Get(counter.PlusOnePlusOne); got != 1 {
			t.Fatalf("+1/+1 counters = %d, want 1", got)
		}
		if attack() {
			t.Fatal("Young Hero triggered at toughness 4")
		}
	})
}

func TestWickedRoleUsesControllerAtDeathAgainstAllOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addRoleTestCreature(g, game.Player1, "Wicked Target")
	role, _ := createTokenPermanent(g, game.Player1, wickedRoleDef(t))
	attachPermanent(g, role, target)
	if power, toughness := effectivePower(g, target), roleEffectiveToughness(t, g, target); power != 3 || toughness != 3 {
		t.Fatalf("Wicked target = %d/%d, want 3/3", power, toughness)
	}
	role.Controller = game.Player3
	startingLife := [game.NumPlayers]int{}
	for player := range game.NumPlayers {
		startingLife[player] = g.Players[player].Life
	}

	if !movePermanentToZone(g, role, zone.Graveyard) {
		t.Fatal("failed to put Wicked Role into graveyard")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Wicked Role graveyard trigger was not put on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Controller != game.Player3 {
		t.Fatalf("Wicked Role trigger = %#v, want Player3 controller", trigger)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for player := range game.NumPlayers {
		want := startingLife[player] - 1
		if game.PlayerID(player) == game.Player3 {
			want = startingLife[player]
		}
		if got := g.Players[player].Life; got != want {
			t.Fatalf("Player%d life = %d, want %d", player+1, got, want)
		}
	}
}
