package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Exquisite Huntmaster (oracle 393ef069-ea6b-4969-b878-
// d942281cc918, printing 444b75d9-f7c7-4a13-9318-02d7bbd80c9c, CMR 122), {3}{B}
// Creature — Elf Warrior 4/2: "When this creature dies, create a 1/1 green Elf
// Warrior creature token.\nEncore {4}{B} ({4}{B}, Exile this card from your
// graveyard: For each opponent, create a token copy that attacks that opponent
// this turn if able. They gain haste. Sacrifice them at the beginning of the next
// end step. Activate only as a sorcery.)" Rulings 2020-11-10 and 2023-07-28 from
// https://api.scryfall.com/cards/444b75d9-f7c7-4a13-9318-02d7bbd80c9c/rulings:
// each token must attack the appropriate player if able; opponents who left the
// game aren't counted; the tokens copy only the printed card; exiling the card is
// a cost; a token that can't attack (tapped, or attacking has a cost) needn't
// attack; a token another player controls when the delayed trigger resolves
// can't be sacrificed; if an effect stops a token attacking its specific player
// it may attack anyone or not attack. CR effective 2025-11-14 (official text
// from media.wizards.com, "MagicCompRules 20251114.txt"): 702.141a (encore),
// 508.1d (attack requirements), 707.2/707.9a (a separate "they gain haste"
// sentence is not a copiable value), 603.7/603.7c (delayed triggers), 702.10
// and 302.6 (haste and summoning sickness), 111.1/111.4 (tokens). The card is
// not admitted here; exquisiteHuntmasterShape plays its printed text through
// the shared Encore body to prove the per-token attack requirement.

func exquisiteHuntmasterShape() *game.CardDef {
	elfWarrior := &game.CardDef{CardFace: game.CardFace{
		Name:      "Elf Warrior",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Warrior},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Exquisite Huntmaster",
			ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.B}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: "When this creature dies, create a 1/1 green Elf Warrior creature token.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:            game.EventPermanentDied,
						Source:           game.TriggerSourceSelf,
						SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(elfWarrior)},
				}}}.Ability(),
			}},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EncoreActivatedBody(cost.Mana{cost.O(4), cost.B}),
			},
		},
	}
}

// TestEncoreExquisiteHuntmasterShapeBindsEachTokenToItsOpponentAndSacrificesCohort
// activates Encore from the graveyard in a four-player main phase: it creates one
// untapped hasty copy per opponent, each required to attack only its matching
// opponent, and the next end step sacrifices the remaining cohort together.
func TestEncoreExquisiteHuntmasterShapeBindsEachTokenToItsOpponentAndSacrificesCohort(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	setMainPhasePriority(g, game.Player1)
	cardID := addCardToGraveyard(g, game.Player1, exquisiteHuntmasterShape())
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)

	encore := action.ActivateAbility(cardID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), encore) {
		t.Fatal("Encore was not offered from the graveyard during the main phase")
	}
	if !engine.applyActionWithChoices(g, game.Player1, encore, agents, &TurnLog{}) {
		t.Fatal("Encore activation failed")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Encore did not exile the card from the graveyard as its cost")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left after paying {4}{B} = %d, want 0", got)
	}
	resolveStackWithTriggers(engine, g, agents)

	tokenIDs := playerTokensNamed(g, game.Player1, "Exquisite Huntmaster")
	if len(tokenIDs) != 3 {
		t.Fatalf("created %d Exquisite Huntmaster copies, want one per opponent (3)", len(tokenIDs))
	}
	boundTo := map[game.PlayerID]id.ID{}
	for _, tokenID := range tokenIDs {
		token, _ := permanentByObjectID(g, tokenID)
		if token.Tapped {
			t.Fatalf("token %d entered tapped, want untapped", tokenID)
		}
		if toughness, _ := effectiveToughness(g, token); effectivePower(g, token) != 4 || toughness != 2 {
			t.Fatalf("token %d is %d/%d, want a 4/2 copy of the printed card", tokenID, effectivePower(g, token), toughness)
		}
		if !hasKeyword(g, token, game.Haste) {
			t.Fatalf("token %d did not gain haste", tokenID)
		}
		copiable, ok := permanentCopyDef(g, token)
		if !ok || copiable.HasKeyword(game.Haste) {
			t.Fatalf("token %d copiable values include haste, want the granted haste to stay noncopiable", tokenID)
		}
		required, directed := directedAttackRequiredPlayers(g, token)
		if !directed || len(required) != 1 {
			t.Fatalf("token %d attack requirements = %v, want exactly one opponent", tokenID, required)
		}
		if previous, dup := boundTo[required[0]]; dup {
			t.Fatalf("tokens %d and %d are both bound to Player%d", previous, tokenID, required[0]+1)
		}
		boundTo[required[0]] = tokenID
	}
	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if _, ok := boundTo[opponent]; !ok {
			t.Fatalf("no token is required to attack Player%d; bindings %v", opponent+1, boundTo)
		}
	}

	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	matching := []game.AttackDeclaration{
		{Attacker: boundTo[game.Player2], Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: boundTo[game.Player3], Target: game.AttackTarget{Player: game.Player3}},
		{Attacker: boundTo[game.Player4], Target: game.AttackTarget{Player: game.Player4}},
	}
	if !attackRequirementsSatisfied(g, game.Player1, matching) {
		t.Fatal("each token attacking its own opponent did not satisfy the attack requirements")
	}
	swapped := slices.Clone(matching)
	swapped[0].Target.Player, swapped[1].Target.Player = game.Player3, game.Player2
	if attackRequirementsSatisfied(g, game.Player1, swapped) {
		t.Fatal("tokens attacking each other's opponents satisfied the per-token requirements")
	}
	if attackRequirementsSatisfied(g, game.Player1, matching[:2]) {
		t.Fatal("leaving an able token out of the attack satisfied its requirement")
	}
	offered := combatEngine{engine}.legalAttackers(g, game.Player1)
	if len(offered) == 0 {
		t.Fatal("no attack declaration was offered")
	}
	for _, act := range offered {
		payload, ok := act.DeclareAttackersPayload()
		if !ok || len(payload.Attackers) != 3 {
			t.Fatalf("offered attack %+v, want all three tokens attacking", payload.Attackers)
		}
		for _, declaration := range payload.Attackers {
			if boundTo[declaration.Target.Player] != declaration.Attacker {
				t.Fatalf("offered token %d attacking Player%d, want its bound opponent", declaration.Attacker, declaration.Target.Player+1)
			}
		}
	}

	g.Combat = nil
	g.Turn.Phase = game.PhasePostcombatMain
	g.Turn.Step = game.StepNone
	setMainPhasePriority(g, game.Player1)
	if _, destroyed := destroyPermanent(g, boundTo[game.Player4]); !destroyed {
		t.Fatal("could not destroy the token bound to Player4")
	}
	resolveStackWithTriggers(engine, g, agents)
	if elves := playerTokensNamed(g, game.Player1, "Elf Warrior"); len(elves) != 1 {
		t.Fatalf("destroyed copy created %d Elf Warriors, want its copied dies trigger to create 1", len(elves))
	}

	engine.runEndingPhase(g, agents)

	if remaining := playerTokensNamed(g, game.Player1, "Exquisite Huntmaster"); len(remaining) != 0 {
		t.Fatalf("%d Encore tokens survived the next end step, want all sacrificed", len(remaining))
	}
	batches, _, _ := eventBatch(g, game.EventPermanentSacrificed)
	if len(batches) != 2 || !singleBatch(batches) {
		t.Fatalf("sacrifice batches = %v, want the two remaining tokens sacrificed together", batches)
	}
	resolveStackWithTriggers(engine, g, agents)
	if elves := playerTokensNamed(g, game.Player1, "Elf Warrior"); len(elves) != 3 {
		t.Fatalf("Elf Warriors after the end step = %d, want one per Encore copy that died (3)", len(elves))
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("the exiled Encore card left exile")
	}
}
