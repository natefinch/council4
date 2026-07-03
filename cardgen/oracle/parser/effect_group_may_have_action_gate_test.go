package parser

import "testing"

// TestParseGroupMayHaveActionGate covers the parser recognition of the
// multiplayer "may have" causative gate — every player, or every opponent, is
// offered a source-actor "deal N damage to them" ("Any player may have Browbeat
// deal 5 damage to them", "any opponent may have it deal 4 damage to them") and a
// resolving consequence branches on whether at least one player accepted. The
// recognizer encodes the chooser scope on the "have" grant's context
// (EffectContextEachPlayer or EffectContextEachOpponent) and appends the
// ConditionPredicatePriorInstruction{Accepted,NotAccepted} clause: "If no one
// does" / "If no player does" the negative branch, "If a player does" the
// affirmative branch.
func TestParseGroupMayHaveActionGate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		body          string
		cardName      string
		instantOrSorc bool
		haveContext   EffectContextKind
		wantPredicate ConditionPredicateKind
	}{
		{
			name:          "any player, if no one does",
			body:          "Any player may have Browbeat deal 5 damage to them. If no one does, target player draws three cards.",
			cardName:      "Browbeat",
			instantOrSorc: true,
			haveContext:   EffectContextEachPlayer,
			wantPredicate: ConditionPredicatePriorInstructionNotAccepted,
		},
		{
			name:          "any opponent, if a player does",
			body:          "When this creature enters, any opponent may have it deal 4 damage to them. If a player does, sacrifice this creature.",
			cardName:      "Vexing Devil",
			haveContext:   EffectContextEachOpponent,
			wantPredicate: ConditionPredicatePriorInstructionAccepted,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.body, Context{CardName: test.cardName, InstantOrSorcery: test.instantOrSorc})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			ability := document.Abilities[0]
			if len(ability.Sentences[0].Effects) != 2 {
				t.Fatalf("action sentence effects = %#v, want [have, action]", ability.Sentences[0].Effects)
			}
			have := ability.Sentences[0].Effects[0]
			if have.Kind != EffectGrantKeyword || !have.Optional || have.Context != test.haveContext {
				t.Fatalf("have effect = %#v, want optional grant with context %v", have, test.haveContext)
			}
			action := ability.Sentences[0].Effects[1]
			if action.Kind != EffectDealDamage || !action.Amount.Known {
				t.Fatalf("action effect = %#v, want known-amount deal damage", action)
			}
			clauses := ability.ConditionClauses
			if len(clauses) != 1 || clauses[0].Predicate != test.wantPredicate {
				t.Fatalf("condition clauses = %#v, want one %v", clauses, test.wantPredicate)
			}
		})
	}
}

// TestParseGroupMayHaveGateRegenerationRiderThirdSentence confirms the
// recognizer handles a multiplayer "may have" offer whose consequence carries a
// credited regeneration rider as a third sentence (Breaking Point's "Creatures
// destroyed this way can't be regenerated."): the rider folds onto the destroy
// consequence during sentence parsing, so the gate tolerates it, sets the group
// chooser context, and appends the negative-branch prior-instruction gate.
func TestParseGroupMayHaveGateRegenerationRiderThirdSentence(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Any player may have Breaking Point deal 6 damage to them. If no one does, destroy all creatures. Creatures destroyed this way can't be regenerated.",
		Context{CardName: "Breaking Point", InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	have := ability.Sentences[0].Effects[0]
	if have.Context != EffectContextEachPlayer {
		t.Fatalf("have context = %v, want EffectContextEachPlayer", have.Context)
	}
	destroy := ability.Sentences[1].Effects[0]
	if !destroy.PreventRegeneration {
		t.Fatal("destroy consequence PreventRegeneration = false, want true (credited rider)")
	}
	found := false
	for _, clause := range ability.ConditionClauses {
		if clause.Predicate == ConditionPredicatePriorInstructionNotAccepted {
			found = true
		}
	}
	if !found {
		t.Fatalf("condition clauses = %#v, want a not-accepted prior-instruction gate", ability.ConditionClauses)
	}
}

// TestParseGroupMayHaveGateFailsClosedOnThirdSentence confirms the recognizer
// still declines a multiplayer "may have" offer followed by a third semantic
// sentence that is not a credited rider ("Draw a card."): such a sentence is a
// distinct third effect the gate cannot model, so the grant context stays unset
// and no prior-instruction ConditionClause is appended.
func TestParseGroupMayHaveGateFailsClosedOnThirdSentence(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Any player may have Browbeat deal 5 damage to them. If no one does, target player draws three cards. Draw a card.",
		Context{CardName: "Browbeat", InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	have := ability.Sentences[0].Effects[0]
	if have.Context == EffectContextEachPlayer || have.Context == EffectContextEachOpponent {
		t.Fatalf("have context = %v, want unset (gate must decline third sentence)", have.Context)
	}
	for _, clause := range ability.ConditionClauses {
		if clause.Predicate == ConditionPredicatePriorInstructionNotAccepted ||
			clause.Predicate == ConditionPredicatePriorInstructionAccepted {
			t.Fatalf("condition clauses = %#v, want no prior-instruction gate", ability.ConditionClauses)
		}
	}
}
