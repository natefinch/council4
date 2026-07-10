package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// addSubtypedCreaturePermanent puts a creature permanent controlled by
// controller on the battlefield carrying the given creature subtypes, so the
// runtime filters that read a permanent's effective subtypes (Goldbug's
// "attacking Humans you control" prevention and "Goldbug and at least one Human
// attack" trigger) can match it.
func addSubtypedCreaturePermanent(g *game.Game, controller game.PlayerID, subtypes ...types.Sub) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:     "Subtyped Creature",
			Types:    []types.Card{types.Creature},
			Subtypes: subtypes,
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestGoldbugPreventsCombatDamageToAttackingHumansYouControl covers Goldbug,
// Humanity's Ally's continuous prevention "Prevent all combat damage that would
// be dealt to attacking Humans you control." Combat damage to an attacking Human
// its controller owns is fully prevented, while every other recipient (a
// non-attacking Human, an attacking non-Human, an opponent's attacking Human)
// and noncombat damage are unaffected.
func TestGoldbugPreventsCombatDamageToAttackingHumansYouControl(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	attackingHuman := addSubtypedCreaturePermanent(g, game.Player1, types.Human)
	idleHuman := addSubtypedCreaturePermanent(g, game.Player1, types.Human)
	attackingSoldier := addSubtypedCreaturePermanent(g, game.Player1, types.Soldier)
	opponentHuman := addSubtypedCreaturePermanent(g, game.Player2, types.Human)

	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: attackingHuman.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: attackingSoldier.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: opponentHuman.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}}

	// Register Goldbug's continuous prevention with Player1 as the controller so
	// its "you control" recipient filter resolves to Player1.
	prevention := game.CombatDamagePreventionToGroupReplacement(
		"Prevent all combat damage that would be dealt to attacking Humans you control.",
		game.Selection{SubtypesAny: []types.Sub{types.Human}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking},
	).Replacement
	prevention.ID = g.IDGen.Next()
	prevention.Controller = game.Player1
	g.ReplacementEffects = append(g.ReplacementEffects, prevention)

	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, attackingHuman, 4, true); dealt != 0 {
		t.Fatalf("combat damage to attacking Human you control = %d, want 0", dealt)
	}
	// The shield persists for further combat damage to the same creature.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, attackingHuman, 5, true); dealt != 0 {
		t.Fatalf("second combat damage to attacking Human = %d, want 0", dealt)
	}
	// Noncombat damage to that same creature is unaffected (combat-only shield).
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, attackingHuman, 3, false); dealt != 3 {
		t.Fatalf("noncombat damage to attacking Human = %d, want 3", dealt)
	}
	// A Human you control that is not attacking is unaffected.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, idleHuman, 4, true); dealt != 4 {
		t.Fatalf("combat damage to non-attacking Human = %d, want 4", dealt)
	}
	// An attacking non-Human you control is unaffected.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, attackingSoldier, 4, true); dealt != 4 {
		t.Fatalf("combat damage to attacking non-Human = %d, want 4", dealt)
	}
	// An attacking Human an opponent controls is unaffected.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, opponentHuman, 4, true); dealt != 4 {
		t.Fatalf("combat damage to opponent's attacking Human = %d, want 4", dealt)
	}
}

// TestGoldbugHumanSpellsCantBeCountered covers Goldbug, Scrappy Scout's static
// "Human spells you control can't be countered." A Human spell its controller
// casts is uncounterable, while a non-Human spell it casts and an opponent's
// Human spell remain counterable.
func TestGoldbugHumanSpellsCantBeCountered(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectCantBeCountered,
		Controller:         game.Player1,
		AffectedController: game.ControllerYou,
		SpellSubtypes:      []types.Sub{types.Human},
		Duration:           game.DurationPermanent,
		CreatedTurn:        g.Turn.TurnNumber,
	})

	humanDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Human Soldier",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Human},
	}}
	elfDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Elf Warrior",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}}

	yourHuman := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: humanDef}
	if stackSpellCanBeCountered(g, yourHuman) {
		t.Fatal("a Human spell you control must be uncounterable")
	}

	yourElf := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: elfDef}
	if !stackSpellCanBeCountered(g, yourElf) {
		t.Fatal("a non-Human spell you control must remain counterable")
	}

	opponentHuman := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2, SourceTokenDef: humanDef}
	if !stackSpellCanBeCountered(g, opponentHuman) {
		t.Fatal("an opponent's Human spell must remain counterable")
	}
}

// TestGoldbugAndHumanAttackTriggerRelation covers Goldbug, Scrappy Scout's
// trigger relation "Whenever Goldbug and at least one Human attack": it holds
// only when the source and at least one other attacker that is a Human both
// attack. Goldbug attacking alone, alongside only a non-Human, or with a Human
// while Goldbug itself is not attacking, all fail.
func TestGoldbugAndHumanAttackTriggerRelation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addSubtypedCreaturePermanent(g, game.Player1)
	human := addSubtypedCreaturePermanent(g, game.Player1, types.Human)
	soldier := addSubtypedCreaturePermanent(g, game.Player1, types.Soldier)

	pattern := &game.TriggerPattern{
		Event:                     game.EventAttackerDeclared,
		Source:                    game.TriggerSourceSelf,
		AttacksAlongsideCount:     1,
		AttacksAlongsideSelection: game.Selection{SubtypesAny: []types.Sub{types.Human}},
	}
	event := game.Event{Kind: game.EventAttackerDeclared}

	// Goldbug and a Human both attack: the relation holds.
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: human.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	if !attacksAlongsideSelection(g, game.Player1, source, pattern, event) {
		t.Fatal("want true: Goldbug and a Human attack")
	}

	// Goldbug attacks alongside only a non-Human: the relation does not hold.
	g.Combat.Attackers[1].Attacker = soldier.ObjectID
	if attacksAlongsideSelection(g, game.Player1, source, pattern, event) {
		t.Fatal("want false: Goldbug attacks alongside only a non-Human")
	}

	// Goldbug attacks alone: the relation does not hold.
	g.Combat.Attackers = g.Combat.Attackers[:1]
	if attacksAlongsideSelection(g, game.Player1, source, pattern, event) {
		t.Fatal("want false: Goldbug attacks alone")
	}

	// A Human attacks but Goldbug does not: the relation does not hold.
	g.Combat.Attackers = []game.AttackDeclaration{
		{Attacker: human.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}
	if attacksAlongsideSelection(g, game.Player1, source, pattern, event) {
		t.Fatal("want false: Goldbug is not attacking")
	}
}
