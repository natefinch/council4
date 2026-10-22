package rules

import (
	"sync"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

var (
	activatedUtilityRegistryOnce sync.Once
	activatedUtilityRegistry     *cards.Registry
)

func activatedUtilityCard(t *testing.T, oracleID string) *game.CardDef {
	t.Helper()
	activatedUtilityRegistryOnce.Do(func() {
		activatedUtilityRegistry = cards.NewDefaultRegistry()
	})
	def, err := activatedUtilityRegistry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("LookupOracleID(%q): %v", oracleID, err)
	}
	return def
}

func TestActivatedUtilityRegistryUsesExactOracleIdentities(t *testing.T) {
	tests := []struct {
		oracleID string
		name     string
		text     string
	}{
		{
			oracleID: "3fb4161d-a5b9-411d-942c-2fdb039e7bde",
			name:     "Mundungu",
			text:     "{T}: Counter target spell unless its controller pays {1} and 1 life.",
		},
		{
			oracleID: "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98",
			name:     "Sly Instigator",
			text:     "{U}, {T}: Until your next turn, target creature an opponent controls can't be blocked. Goad that creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)",
		},
		{
			oracleID: "00d11662-9d10-4915-b80c-c1ad54b4e71b",
			name:     "Half-Elf Monk",
			text:     "Vigilance\nStunning Strike — {1}{W}, {T}: Tap target creature.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := activatedUtilityCard(t, tt.oracleID)
			if def.Name != tt.name || def.OracleText != tt.text {
				t.Fatalf("registered card = %q %q, want %q %q", def.Name, def.OracleText, tt.name, tt.text)
			}
		})
	}
}

func TestMundunguTargetsOnlySpells(t *testing.T) {
	def := activatedUtilityCard(t, "3fb4161d-a5b9-411d-942c-2fdb039e7bde")

	t.Run("spell", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, def)
		spell := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
		setMainPhasePriority(g, game.Player1)

		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
			[]game.Target{game.StackObjectTarget(spell.ID)}, 0)) {
			t.Fatal("Mundungu could not target a spell")
		}
	})

	t.Run("activated ability", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, def)
		ability := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player2}
		g.Stack.Push(ability)
		setMainPhasePriority(g, game.Player1)

		if engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
			[]game.Target{game.StackObjectTarget(ability.ID)}, 0)) {
			t.Fatal("Mundungu illegally targeted an activated ability")
		}
		if source.Tapped || g.Stack.Size() != 1 {
			t.Fatal("illegal target paid the tap cost or added an ability")
		}
	})
}

func TestMundunguCurrentControllerPaysCombinedCostAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1,
		activatedUtilityCard(t, "3fb4161d-a5b9-411d-942c-2fdb039e7bde"))
	spell := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Instant})
	g.CardInstances[spell.SourceID].Owner = game.Player3
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.StackObjectTarget(spell.ID)}, 0)) {
		t.Fatal("Mundungu activation failed")
	}

	// The spell changes controllers while the ability waits on the stack.
	spell.Controller = game.Player4
	g.Players[game.Player4].ManaPool.Add(mana.C, 1)
	clone := g.Clone()
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player4] = optionalMayAgent{accept: true}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(clone, agents, &log)

	if got := clone.Players[game.Player4].ManaPool.Total(); got != 0 {
		t.Fatalf("current spell controller mana = %d, want 0 after paying {1}", got)
	}
	if got := clone.Players[game.Player4].Life; got != 39 {
		t.Fatalf("current spell controller life = %d, want 39 after paying 1 life", got)
	}
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player1} {
		if clone.Players[player].Life != 40 || clone.Players[player].ManaPool.Total() != 0 {
			t.Fatalf("player %v paid or changed life: life=%d mana=%d", player, clone.Players[player].Life, clone.Players[player].ManaPool.Total())
		}
	}
	if _, ok := stackObjectByID(clone, spell.ID); !ok {
		t.Fatal("Mundungu countered the spell after its current controller paid")
	}
	for _, choice := range log.Choices {
		if choice.Request.Player != game.Player4 {
			t.Fatalf("payment choice was offered to %v, want current controller Player4", choice.Request.Player)
		}
	}

	if got := g.Players[game.Player4].ManaPool.Total(); got != 1 || g.Players[game.Player4].Life != 40 {
		t.Fatalf("resolving the clone changed original controller state: mana=%d life=%d", got, g.Players[game.Player4].Life)
	}
	if g.Stack.Size() != 2 {
		t.Fatalf("original stack size = %d, want 2 after clone resolution", g.Stack.Size())
	}
}

func TestMundunguFailedPaymentCounters(t *testing.T) {
	for _, tt := range []struct {
		name    string
		mana    int
		decline bool
	}{
		{name: "declined", mana: 1, decline: true},
		{name: "unaffordable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1,
				activatedUtilityCard(t, "3fb4161d-a5b9-411d-942c-2fdb039e7bde"))
			spell := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
			g.CardInstances[spell.SourceID].Owner = game.Player3
			g.Players[game.Player2].ManaPool.Add(mana.C, tt.mana)
			setMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
				[]game.Target{game.StackObjectTarget(spell.ID)}, 0)) {
				t.Fatal("Mundungu activation failed")
			}

			agents := [game.NumPlayers]PlayerAgent{}
			if tt.decline {
				agents[game.Player2] = optionalMayAgent{accept: false}
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			if _, ok := stackObjectByID(g, spell.ID); ok {
				t.Fatal("spell remained after its controller declined or could not pay")
			}
			if !g.Players[game.Player3].Graveyard.Contains(spell.SourceID) {
				t.Fatal("countered spell did not move to its owner's graveyard")
			}
			if g.Players[game.Player2].Life != 40 || g.Players[game.Player2].ManaPool.Total() != tt.mana {
				t.Fatalf("declined/unaffordable payment changed controller resources: life=%d mana=%d",
					g.Players[game.Player2].Life, g.Players[game.Player2].ManaPool.Total())
			}
			if tt.decline {
				if len(log.Choices) != 1 || log.Choices[0].Request.Player != game.Player2 {
					t.Fatalf("decline choices = %#v, want one offer to the target controller", log.Choices)
				}
			} else if len(log.Choices) != 0 {
				t.Fatalf("unaffordable payment prompted a choice: %#v", log.Choices)
			}
		})
	}
}

func TestMundunguUncounterableAndDepartedTargets(t *testing.T) {
	t.Run("uncounterable spell", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1,
			activatedUtilityCard(t, "3fb4161d-a5b9-411d-942c-2fdb039e7bde"))
		spell := addStackSpellWithFace(g, game.Player2, &game.CardFace{
			Name: "Uncounterable Spell", Types: []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{game.CantBeCounteredStaticBody},
		})
		setMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
			[]game.Target{game.StackObjectTarget(spell.ID)}, 0)) {
			t.Fatal("Mundungu could not target the uncounterable spell")
		}
		agents := [game.NumPlayers]PlayerAgent{game.Player2: optionalMayAgent{accept: false}}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		if _, ok := stackObjectByID(g, spell.ID); !ok {
			t.Fatal("uncounterable spell was countered after its controller declined")
		}
	})

	t.Run("spell leaves stack", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1,
			activatedUtilityCard(t, "3fb4161d-a5b9-411d-942c-2fdb039e7bde"))
		spell := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Instant})
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
			[]game.Target{game.StackObjectTarget(spell.ID)}, 0)) {
			t.Fatal("Mundungu activation failed")
		}
		addEffectSpellToStack(g, game.Player4,
			game.CounterObject{Object: game.TargetStackObjectReference(0)},
			[]game.Target{game.StackObjectTarget(spell.ID)})
		engine.resolveTopOfStack(g, &TurnLog{})

		log := TurnLog{}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
		if len(log.Choices) != 0 {
			t.Fatalf("Mundungu asked for payment after its target left the stack: %#v", log.Choices)
		}
		if g.Stack.Size() != 0 || !g.Players[game.Player2].Graveyard.Contains(spell.SourceID) {
			t.Fatal("departed target did not remain countered while Mundungu's ability resolved")
		}
	})
}

func TestSlyInstigatorTargetRevalidation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player3,
		activatedUtilityCard(t, "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98"))
	ownCreature := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Players[game.Player3].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player3)

	if engine.applyAction(g, game.Player3, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PermanentTarget(ownCreature.ObjectID)}, 0)) {
		t.Fatal("Sly Instigator illegally targeted its controller's creature")
	}
	if !engine.applyAction(g, game.Player3, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PermanentTarget(opponentCreature.ObjectID)}, 0)) {
		t.Fatal("Sly Instigator could not target an opponent's creature")
	}

	// A target that stops being controlled by an opponent before resolution is illegal.
	opponentCreature.Controller = game.Player3
	engine.resolveTopOfStack(g, &TurnLog{})
	if wasGoadedBy(opponentCreature, game.Player3) || !canBlockAttacker(g, ownCreature, opponentCreature) {
		t.Fatal("Sly Instigator's effect resolved after its target ceased to be an opponent's creature")
	}
}

func TestSlyInstigatorSourceDepartureAndCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player3,
		activatedUtilityCard(t, "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98"))
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Players[game.Player3].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player3)
	if !engine.applyAction(g, game.Player3, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0)) {
		t.Fatal("Sly Instigator activation failed")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.Kind != game.StackActivatedAbility {
		t.Fatal("Sly Instigator activation was not placed on the stack")
	}

	addEffectSpellToStack(g, game.Player3,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.Kind != game.StackActivatedAbility || copy.ID == original.ID {
		t.Fatalf("copied Sly ability = %+v, want an independent activated-ability copy", copy)
	}

	addEffectSpellToStack(g, game.Player1, game.Destroy{Object: game.TargetPermanentReference(0)},
		[]game.Target{game.PermanentTarget(source.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, present := permanentByObjectID(g, source.ObjectID); present {
		t.Fatal("Sly Instigator remained on the battlefield after destruction")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !wasGoadedBy(target, game.Player3) || canBlockAttacker(g, addCombatCreaturePermanent(g, game.Player2), target) {
		t.Fatal("copied Sly ability did not apply both effects after its source left")
	}
}

func TestSlyInstigatorDurationUsesAbilityController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player3,
		activatedUtilityCard(t, "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98"))
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Players[game.Player3].ManaPool.Add(mana.U, 1)

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attack, ok := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}).DeclareAttackersPayload()
	if !ok || !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("opponent creature could not attack before Sly Instigator resolved")
	}
	g.Turn.Step = game.StepDeclareBlockers
	if !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("blocker could not block before Sly Instigator resolved")
	}

	g.Turn.PriorityPlayer = game.Player3
	if !engine.applyAction(g, game.Player3, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PermanentTarget(attacker.ObjectID)}, 0)) {
		t.Fatal("Sly Instigator could not be activated during another player's combat")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !wasGoadedBy(attacker, game.Player3) || canBlockAttacker(g, blocker, attacker) {
		t.Fatal("Sly Instigator did not independently goad and prevent blocking")
	}
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("combat step = %v, want declare blockers", g.Turn.Step)
	}
	blocks, ok := action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: blocker.ObjectID, Blocking: attacker.ObjectID,
	}}).DeclareBlockersPayload()
	if !ok || engine.applyDeclareBlockers(g, game.Player2, blocks) {
		t.Fatal("Engine accepted a block against Sly Instigator's unblockable target")
	}

	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player2
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !wasGoadedBy(attacker, game.Player3) || canBlockAttacker(g, blocker, attacker) {
		t.Fatal("Sly Instigator's effects expired on another player's turn")
	}
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player3
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if wasGoadedBy(attacker, game.Player3) || !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("Sly Instigator's effects did not expire on its ability controller's next turn")
	}
}

func TestHalfElfMonkVigilance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	monk := addCombatPermanent(g, game.Player1,
		activatedUtilityCard(t, "00d11662-9d10-4915-b80c-c1ad54b4e71b"))
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	attack, ok := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: monk.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}).DeclareAttackersPayload()
	if !ok || !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("Half-Elf Monk could not attack")
	}
	if monk.Tapped {
		t.Fatal("Half-Elf Monk was tapped attacking despite vigilance")
	}
}

func TestHalfElfMonkActivationCostsAndTiming(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	monk := addCombatPermanent(g, game.Player1,
		activatedUtilityCard(t, "00d11662-9d10-4915-b80c-c1ad54b4e71b"))
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	monk.SummoningSick = true
	act := action.ActivateAbility(monk.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)

	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("summoning-sick Half-Elf Monk activated its tap ability")
	}
	if monk.Tapped || g.Players[game.Player1].ManaPool.Total() != 2 {
		t.Fatal("illegal activation paid costs")
	}
	monk.SummoningSick = false
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("Half-Elf Monk could not activate at instant speed during an opponent's upkeep")
	}
	if !monk.Tapped || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("activation costs not paid: tapped=%v mana=%d", monk.Tapped, g.Players[game.Player1].ManaPool.Total())
	}
	if target.Tapped {
		t.Fatal("target creature tapped before the ability resolved")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped {
		t.Fatal("Half-Elf Monk did not tap the target creature")
	}
}

func TestHalfElfMonkTargetLoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	monk := addCombatPermanent(g, game.Player1,
		activatedUtilityCard(t, "00d11662-9d10-4915-b80c-c1ad54b4e71b"))
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	unrelated := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(monk.ObjectID, 0,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0)) {
		t.Fatal("Half-Elf Monk activation failed")
	}
	addEffectSpellToStack(g, game.Player3, game.Destroy{Object: game.TargetPermanentReference(0)},
		[]game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	if _, present := permanentByObjectID(g, target.ObjectID); present {
		t.Fatal("destroyed target remained on the battlefield")
	}
	if unrelated.Tapped {
		t.Fatal("Half-Elf Monk's ability affected a different creature after its target left")
	}
	if !monk.Tapped || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("activation costs were not paid before the target left")
	}
}
