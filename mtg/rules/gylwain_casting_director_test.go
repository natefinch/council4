package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func testRoleDef(name, oracleText string, abilities ...game.StaticAbility) *game.CardDef {
	staticAbilities := []game.StaticAbility{game.EnchantStaticAbility(&game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "creature",
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature},
		}),
	})}
	staticAbilities = append(staticAbilities, abilities...)
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Enchantment},
		Subtypes:        []types.Sub{types.Aura, types.Role},
		StaticAbilities: staticAbilities,
		OracleText:      oracleText,
	}}
}

func testAttachedRoleAbility(effects ...game.ContinuousEffect) game.StaticAbility {
	for i := range effects {
		effects[i].Group = game.AttachedObjectGroup(game.SourcePermanentReference())
	}
	return game.StaticAbility{ContinuousEffects: effects}
}

func cursedRoleDef() *game.CardDef {
	return testRoleDef(
		"Cursed Role",
		"Enchant creature\nEnchanted creature has base power and toughness 1/1.",
		testAttachedRoleAbility(game.ContinuousEffect{
			Layer:        game.LayerPowerToughnessSet,
			SetPower:     opt.Val(game.PT{Value: 1}),
			SetToughness: opt.Val(game.PT{Value: 1}),
		}),
	)
}

func monsterRoleDef() *game.CardDef {
	return testRoleDef(
		"Monster Role",
		"Enchant creature\nEnchanted creature gets +1/+1 and has trample.",
		testAttachedRoleAbility(
			game.ContinuousEffect{
				Layer:       game.LayerAbility,
				AddKeywords: []game.Keyword{game.Trample},
			},
			game.ContinuousEffect{
				Layer:          game.LayerPowerToughnessModify,
				PowerDelta:     1,
				ToughnessDelta: 1,
			},
		),
	)
}

func royalRoleDef() *game.CardDef {
	ward := game.WardStaticAbility(cost.Mana{cost.O(1)})
	return testRoleDef(
		"Royal Role",
		"Enchant creature\nEnchanted creature gets +1/+1 and has ward {1}.\n(Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)",
		testAttachedRoleAbility(
			game.ContinuousEffect{
				Layer:        game.LayerAbility,
				AddAbilities: []game.Ability{&ward},
			},
			game.ContinuousEffect{
				Layer:          game.LayerPowerToughnessModify,
				PowerDelta:     1,
				ToughnessDelta: 1,
			},
		),
	)
}

func sorcererRoleDef() *game.CardDef {
	granted := &game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Scry{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			},
		}}}.Ability(),
	}
	return testRoleDef(
		"Sorcerer Role",
		"Enchant creature\nEnchanted creature gets +1/+1 and has \"Whenever this creature attacks, scry 1.\"",
		testAttachedRoleAbility(
			game.ContinuousEffect{
				Layer:        game.LayerAbility,
				AddAbilities: []game.Ability{granted},
			},
			game.ContinuousEffect{
				Layer:          game.LayerPowerToughnessModify,
				PowerDelta:     1,
				ToughnessDelta: 1,
			},
		),
	)
}

func wickedRoleDef() *game.CardDef {
	def := testRoleDef(
		"Wicked Role",
		"Enchant creature\nEnchanted creature gets +1/+1.\nWhen this Aura is put into a graveyard from the battlefield, each opponent loses 1 life.",
		testAttachedRoleAbility(game.ContinuousEffect{
			Layer:          game.LayerPowerToughnessModify,
			PowerDelta:     1,
			ToughnessDelta: 1,
		}),
	)
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				Source:        game.TriggerSourceSelf,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
				MatchToZone:   true,
				ToZone:        zone.Graveyard,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.LoseLife{
				Amount:      game.Fixed(1),
				PlayerGroup: game.OpponentsReference(),
			},
		}}}.Ability(),
	}}
	return def
}

func youngHeroRoleDef() *game.CardDef {
	granted := &game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
			InterveningIf: "if its toughness is 3 or less",
			InterveningCondition: opt.Val(game.Condition{
				Object: opt.Val(game.EventPermanentReference()),
				ObjectMatches: opt.Val(game.Selection{
					Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3}),
				}),
			}),
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      game.Fixed(1),
				Object:      game.EventPermanentReference(),
				CounterKind: counter.PlusOnePlusOne,
			},
		}}}.Ability(),
	}
	return testRoleDef(
		"Young Hero Role",
		"Enchant creature\nEnchanted creature has \"Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.\"",
		testAttachedRoleAbility(game.ContinuousEffect{
			Layer:        game.LayerAbility,
			AddAbilities: []game.Ability{granted},
		}),
	)
}

func gylwainTestDef() *game.CardDef {
	roles := []*game.CardDef{royalRoleDef(), sorcererRoleDef(), monsterRoleDef()}
	modeText := []string{
		"Create a Royal Role token attached to that creature.",
		"Create a Sorcerer Role token attached to that creature.",
		"Create a Monster Role token attached to that creature.",
	}
	modes := make([]game.Mode, len(roles))
	for i, role := range roles {
		modes[i] = game.Mode{
			Text: modeText[i],
			Sequence: []game.Instruction{{
				Primitive: game.CreateToken{
					Amount:          game.Fixed(1),
					Source:          game.TokenDef(role),
					EntryAttachedTo: opt.Val(game.EventPermanentReference()),
				},
			}},
		}
	}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:       "Gylwain, Casting Director",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G, cost.W}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Bard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:                  game.EventPermanentEnteredBattlefield,
						Controller:             game.TriggerControllerYou,
						SubjectSelectionOrSelf: true,
						SubjectSelection: game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							NonToken:      true,
						},
					},
				},
				Content: game.AbilityContent{
					Modes:    modes,
					MinModes: 1,
					MaxModes: 1,
				},
			}},
			OracleText: "Whenever Gylwain or another nontoken creature you control enters, choose one —\n" +
				"• Create a Royal Role token attached to that creature.\n" +
				"• Create a Sorcerer Role token attached to that creature.\n" +
				"• Create a Monster Role token attached to that creature.",
		},
	}
}

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
	def := gylwainTestDef()
	return roleTokenFromInstruction(t, def.TriggeredAbilities[0].Content.Modes[mode].Sequence[0])
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
			gylwain := addCombatPermanent(g, game.Player1, gylwainTestDef())
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
	gylwain := addCombatPermanent(g, game.Player1, gylwainTestDef())
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
	addCombatPermanent(g, game.Player1, gylwainTestDef())
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
	addCombatPermanent(g, game.Player1, gylwainTestDef())
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
	addCombatPermanent(g, game.Player1, gylwainTestDef())
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
	cursed, _ := createTokenPermanent(g, game.Player1, cursedRoleDef())
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
		role, _ := createTokenPermanent(g, game.Player1, youngHeroRoleDef())
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
	role, _ := createTokenPermanent(g, game.Player1, wickedRoleDef())
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
