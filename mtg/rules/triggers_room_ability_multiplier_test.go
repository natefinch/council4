package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// roomAbilityStaticBody is the ability Dungeon Delver grants and Hama Pashar
// prints: "Room abilities of dungeons you own trigger an additional time."
func roomAbilityStaticBody() game.StaticAbility {
	return game.StaticAbility{
		Text:        "Room abilities of dungeons you own trigger an additional time.",
		RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectAdditionalTriggerForRoomAbility}},
	}
}

// roomAbilityDungeonDef is a Dungeon whose room ability triggers on an event, so
// the multiplier machinery has an ordinary triggered ability of a Dungeon-typed
// source to double.
func roomAbilityDungeonDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Dungeon},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:       game.EventZoneChanged,
				Source:      game.TriggerSourceSelf,
				MatchToZone: true,
				ToZone:      zone.Battlefield,
			}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			}}.Ability(),
		}},
	}}
}

// printedRoomAbilityDoublerDef prints the room-ability doubler directly, as Hama
// Pashar, Ruin Seeker does.
func printedRoomAbilityDoublerDef(name string) *game.CardDef {
	body := roomAbilityStaticBody()
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Supertypes:      []types.Super{types.Legendary},
		Types:           []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{body},
	}}
}

// roomAbilityCreatureDef is an ordinary creature carrying no room-ability doubler.
func roomAbilityCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
}

func TestRoomAbilityPrintedDoublerMultipliesRoomTrigger(t *testing.T) {
	for name, tc := range map[string]struct {
		doublers  int
		wantStack int
	}{
		"no doubler leaves a single trigger": {doublers: 0, wantStack: 1},
		"one doubler adds one occurrence":    {doublers: 1, wantStack: 2},
		"two doublers stack additively":      {doublers: 2, wantStack: 3},
		"three doublers stack additively":    {doublers: 3, wantStack: 4},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for i := 0; i < tc.doublers; i++ {
				addCombatPermanent(g, game.Player1, printedRoomAbilityDoublerDef("Ruin Seeker"))
			}
			dungeon := addCombatPermanent(g, game.Player1, roomAbilityDungeonDef("Undercity"))

			emitSelfEnter(g, dungeon)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("room ability was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.wantStack {
				t.Fatalf("stack size = %d, want %d", got, tc.wantStack)
			}
		})
	}
}

func TestRoomAbilityDoublerIgnoresNonDungeonTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, printedRoomAbilityDoublerDef("Ruin Seeker"))
	// A creature (not a Dungeon) whose triggered ability fires the same way; its
	// trigger is not a room ability, so the doubler must not multiply it.
	source := addCombatPermanent(g, game.Player1, selfEntersTypedTriggerSourceDef(
		"Creature Source", []types.Super{types.Legendary}, []types.Card{types.Creature}, nil))

	emitSelfEnter(g, source)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (non-room trigger is not doubled)", got)
	}
}

func TestRoomAbilityDoublerRespectsDungeonOwner(t *testing.T) {
	for name, tc := range map[string]struct {
		dungeonOwner game.PlayerID
		wantStack    int
	}{
		"dungeon owned by the doubler's controller is doubled": {dungeonOwner: game.Player1, wantStack: 2},
		"dungeon owned by another player is not doubled":       {dungeonOwner: game.Player2, wantStack: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, printedRoomAbilityDoublerDef("Ruin Seeker"))
			dungeon := addCombatPermanent(g, tc.dungeonOwner, roomAbilityDungeonDef("Undercity"))

			emitSelfEnter(g, dungeon)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("room ability was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.wantStack {
				t.Fatalf("stack size = %d, want %d", got, tc.wantStack)
			}
		})
	}
}

// grantRoomAbilityToCommanders appends a continuous effect modeling Dungeon
// Delver's grant: "Commander creatures you own have \"...\"", lowered to a
// battlefield-wide grant filtered to commander creatures the source's controller
// OWNS (Owner: OwnerYou), regardless of who currently controls them.
func grantRoomAbilityToCommanders(g *game.Game, source *game.Permanent) {
	body := roomAbilityStaticBody()
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             g.IDGen.Next(),
		Controller:     source.Controller,
		SourceObjectID: source.ObjectID,
		Layer:          game.LayerAbility,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes:  []types.Card{types.Creature},
			MatchCommander: true,
			Owner:          game.OwnerYou,
		}),
		AddAbilities: []game.Ability{&body},
	})
}

func TestRoomAbilityGrantedToCommanderMultiplies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	delver := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dungeon Delver",
		Types: []types.Card{types.Enchantment},
	}})
	grantRoomAbilityToCommanders(g, delver)

	// A commander creature the Dungeon Delver controller owns gains the doubler.
	commander := addCombatPermanent(g, game.Player1, roomAbilityCreatureDef("Commander Creature"))
	g.CommanderIDs[commander.CardInstanceID] = true
	// A noncommander creature does not gain the doubler.
	addCombatPermanent(g, game.Player1, roomAbilityCreatureDef("Ordinary Creature"))

	dungeon := addCombatPermanent(g, game.Player1, roomAbilityDungeonDef("Undercity"))

	emitSelfEnter(g, dungeon)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("room ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (granted commander doubles the room ability once)", got)
	}
}

func TestRoomAbilityGrantMultipleDelversStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dungeon Delver A", Types: []types.Card{types.Enchantment}}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dungeon Delver B", Types: []types.Card{types.Enchantment}}})
	grantRoomAbilityToCommanders(g, first)
	grantRoomAbilityToCommanders(g, second)

	commander := addCombatPermanent(g, game.Player1, roomAbilityCreatureDef("Commander Creature"))
	g.CommanderIDs[commander.CardInstanceID] = true

	dungeon := addCombatPermanent(g, game.Player1, roomAbilityDungeonDef("Undercity"))

	emitSelfEnter(g, dungeon)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("room ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("stack size = %d, want 3 (two grants each add one occurrence)", got)
	}
}

func TestRoomAbilityGrantFollowsOwnershipNotControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	delver := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dungeon Delver", Types: []types.Card{types.Enchantment}}})
	grantRoomAbilityToCommanders(g, delver)

	// A commander Player1 owns is stolen by Player2. "Commander creatures you own"
	// keys off OWNERSHIP, so the grant still applies to the commander even though
	// Player2 now controls it: the doubler rides the commander into Player2's
	// control, where it multiplies room abilities of dungeons Player2 owns.
	commander := addCombatPermanent(g, game.Player1, roomAbilityCreatureDef("Commander Creature"))
	g.CommanderIDs[commander.CardInstanceID] = true
	commander.Controller = game.Player2

	// The dungeon Player2 owns and controls has its room ability doubled by the
	// stolen commander, because the commander's effective controller (Player2)
	// matches the dungeon's owner.
	p2Dungeon := addCombatPermanent(g, game.Player2, roomAbilityDungeonDef("Player2 Undercity"))
	emitSelfEnter(g, p2Dungeon)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("player2 room ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (owner-based grant follows the stolen commander)", got)
	}

	// Player1's own dungeon is not doubled: the stolen commander carrying the
	// doubler is controlled by Player2, so it does not multiply Player1's dungeon.
	g.Stack = game.Stack{}
	p1Dungeon := addCombatPermanent(g, game.Player1, roomAbilityDungeonDef("Player1 Undercity"))
	emitSelfEnter(g, p1Dungeon)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("player1 room ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (stolen commander doubles its controller's dungeons, not the owner's)", got)
	}
}

func TestRoomAbilityGrantSkipsControlledOpponentOwnedCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	delver := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dungeon Delver", Types: []types.Card{types.Enchantment}}})
	grantRoomAbilityToCommanders(g, delver)

	// A commander Player2 OWNS but Player1 currently controls. "Commander
	// creatures you own" excludes it: Player1 owns the Dungeon Delver but not this
	// commander, so the grant does not apply and no room-ability doubler is added.
	commander := addCombatPermanent(g, game.Player2, roomAbilityCreatureDef("Commander Creature"))
	g.CommanderIDs[commander.CardInstanceID] = true
	commander.Controller = game.Player1

	p1Dungeon := addCombatPermanent(g, game.Player1, roomAbilityDungeonDef("Player1 Undercity"))
	emitSelfEnter(g, p1Dungeon)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("player1 room ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (grant does not apply to a commander Player1 controls but does not own)", got)
	}
}
