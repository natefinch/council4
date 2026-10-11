package compiler

import "testing"

func compiledRoleContent(t *testing.T, text string, context pipelineContext, ability int) AbilityContent {
	t.Helper()
	compilation, diagnostics := compileSource(text, context)
	if len(diagnostics) != 0 || ability >= len(compilation.Abilities) {
		t.Fatalf("diagnostics=%v abilities=%d", diagnostics, len(compilation.Abilities))
	}
	return compilation.Abilities[ability].Content
}

func TestReferenceRoleCastCostIsPolicyNotObject(t *testing.T) {
	content := compiledRoleContent(t,
		"Electrodominance deals X damage to any target. You may cast a spell with mana value X or less from your hand without paying its mana cost.",
		pipelineContext{CardName: "Electrodominance", InstantOrSorcery: true}, 0)
	found := 0
	for _, reference := range content.References {
		if reference.Pronoun != ReferencePronounIts {
			continue
		}
		found++
		if reference.SubjectDomain() != ReferenceSubjectPolicy || !reference.SupportsUse(ReferenceUsePolicy) ||
			reference.SupportsUse(ReferenceUseMutation) || reference.SupportsUse(ReferenceUseCharacteristic) ||
			reference.SupportsUse(ReferenceUseAttribution) {
			t.Fatalf("selected spell's cost policy gained object authority: %+v", reference)
		}
	}
	if found != 1 {
		t.Fatalf("cost policy references = %d, want 1", found)
	}
}

func TestReferenceRoleTopExiledCardProducer(t *testing.T) {
	content := compiledRoleContent(t,
		"When this enchantment enters, you become the monarch.\nAt the beginning of your upkeep, exile the top card of target opponent's library. You may play that card for as long as it remains exiled, and mana of any type can be spent to cast it. If you're the monarch, until end of turn, you may cast a spell from among cards exiled with this enchantment without paying its mana cost.",
		pipelineContext{CardName: "Court of Locthwain"}, 1)
	if len(content.Effects) != 3 {
		t.Fatalf("effects = %d, want 3", len(content.Effects))
	}
	producer := content.Effects[0].ClauseID
	if len(content.Effects[1].References) != 3 {
		t.Fatal("permission lost an exiled-card reference")
	}
	for _, reference := range content.Effects[1].References {
		if !content.OwnsSubject(reference) || reference.Binding != ReferenceBindingPriorInstructionResult ||
			reference.PriorInstruction != 0 || reference.ProducerClauseID != producer ||
			reference.SubjectDomain() != ReferenceSubjectCard ||
			reference.SubjectLifetime() != ReferenceLifetimeActualProduct ||
			reference.SupportsUse(ReferenceUseMutation) {
			t.Fatalf("exiled-card reference does not name the actual exile product: %+v", reference)
		}
	}
	for _, reference := range content.Effects[2].References {
		if reference.ProducerClauseID == producer {
			t.Fatal("free-cast cost policy was captured by the top-card exile producer")
		}
	}
}

func TestReferenceRoleTopExiledCardProducerNearMisses(t *testing.T) {
	for _, text := range []string{
		"At the beginning of your upkeep, exile the top two cards of your library. You may play that card this turn.",
		"At the beginning of your upkeep, exile the top card of your library. Exile target creature. You may play that card this turn.",
	} {
		t.Run(text, func(t *testing.T) {
			compilation, _ := compileSource(text, pipelineContext{CardName: "Near Miss"})
			for _, ability := range compilation.Abilities {
				content := ability.Content
				if len(content.Effects) == 0 {
					continue
				}
				first := content.Effects[0].ClauseID
				for _, reference := range content.References {
					if reference.ProducerClauseID == first && reference.SubjectSupported() {
						t.Fatalf("non-singular or interrupted exile produced a card subject: %+v", reference)
					}
				}
			}
		})
	}
}

func TestReferenceRoleEnclosingSpellParagraphTarget(t *testing.T) {
	const addendum = "\nAddendum — If you cast this spell during your main phase, tap that creature and it doesn't untap during its controller's next untap step."
	compilation, diagnostics := compileSource(
		"Target creature gets -4/-0 until end of turn.\nDraw a card."+addendum,
		pipelineContext{CardName: "Code of Constraint", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	borrowed := 0
	for _, ability := range compilation.Abilities {
		content := ability.Content
		if len(content.Targets) != 0 {
			continue
		}
		for _, reference := range content.References {
			if reference.Binding != ReferenceBindingTarget {
				continue
			}
			if _, ok := reference.EnclosingSpellTargetOccurrence(); !ok || !content.OwnsSubject(reference) {
				t.Fatalf("addendum reference lost the spell's exact earlier target: %+v", reference)
			}
			borrowed++
		}
	}
	if borrowed == 0 {
		t.Fatal("no addendum reference bound to the enclosing spell target")
	}

	for _, first := range []string{
		"Up to two target creatures get -4/-0 until end of turn.",
		"Choose one —\n• Target creature gets -4/-0 until end of turn.\n• Draw a card.",
	} {
		t.Run(first, func(t *testing.T) {
			compilation, _ := compileSource(first+addendum,
				pipelineContext{CardName: "Near Miss", InstantOrSorcery: true})
			for _, ability := range compilation.Abilities {
				if len(ability.Content.Targets) != 0 || len(ability.Content.Modes) != 0 {
					continue
				}
				for _, reference := range ability.Content.References {
					if _, ok := reference.EnclosingSpellTargetOccurrence(); ok {
						t.Fatalf("plural or modal earlier target was borrowed: %+v", reference)
					}
				}
			}
		})
	}
}

func referencesInDomain(content AbilityContent, domain ReferenceSubjectDomain) []CompiledReference {
	var found []CompiledReference
	for _, reference := range content.References {
		if reference.SubjectDomain() == domain {
			found = append(found, reference)
		}
	}
	return found
}

func TestReferenceRoleDefendingPaymentDecisionPolicy(t *testing.T) {
	content := compiledRoleContent(t,
		"Whenever this creature attacks, defending player may pay {4}. If that player doesn't, this creature can't be blocked this turn.",
		pipelineContext{CardName: "Shrouded Serpent"}, 0)
	policies := referencesInDomain(content, ReferenceSubjectPolicy)
	if len(policies) != 1 || policies[0].Kind != ReferenceThatPlayer || !policies[0].SupportsUse(ReferenceUsePolicy) {
		t.Fatalf("payment decision subject = %+v, want one owned decision policy", policies)
	}
	for _, text := range []string{
		"At the beginning of your upkeep, target opponent may pay {4}. If that player doesn't, draw a card.",
		"Whenever this creature attacks, defending player loses 1 life. If that player has no cards in hand, draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			compilation, _ := compileSource(text, pipelineContext{CardName: "Near Miss"})
			for _, ability := range compilation.Abilities {
				if policies := referencesInDomain(ability.Content, ReferenceSubjectPolicy); len(policies) != 0 {
					t.Fatalf("unowned player reference became a payment decision: %+v", policies)
				}
			}
		})
	}
}

func TestReferenceRoleCorrelatedOpponentPolicy(t *testing.T) {
	content := compiledRoleContent(t,
		"Each opponent exiles a creature with the greatest power among creatures that player controls.\nSpell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Olórin's Searing Light deals damage to each opponent equal to the power of the creature they exiled.",
		pipelineContext{CardName: "Olórin's Searing Light", InstantOrSorcery: true}, 0)
	if policies := referencesInDomain(content, ReferenceSubjectPolicy); len(policies) == 0 {
		t.Fatal("correlated member references lost their policy proof")
	}
	compilation, _ := compileSource("Each opponent exiles a creature they control.",
		pipelineContext{CardName: "Near Miss", InstantOrSorcery: true})
	for _, ability := range compilation.Abilities {
		if policies := referencesInDomain(ability.Content, ReferenceSubjectPolicy); len(policies) != 0 {
			t.Fatalf("uncorrelated exile acquired a correlated policy: %+v", policies)
		}
	}
}

func TestReferenceRoleSearchedPermanentProduct(t *testing.T) {
	content := compiledRoleContent(t,
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
		pipelineContext{CardName: "Fabled Passage"}, 0)
	products := 0
	for _, reference := range content.References {
		if reference.Kind == ReferenceThatObject && reference.SubjectDomain() == ReferenceSubjectPermanent &&
			reference.SubjectLifetime() == ReferenceLifetimeActualProduct {
			products++
		}
	}
	if products != 1 {
		t.Fatal("untap subject does not name the actual entered searched land")
	}
	compilation, _ := compileSource(
		"{T}, Sacrifice this land: Search your library for a basic land card, put it into your hand, then shuffle. Then untap that land.",
		pipelineContext{CardName: "Near Miss"})
	for _, ability := range compilation.Abilities {
		for _, reference := range ability.Content.References {
			if reference.Kind == ReferenceThatObject && reference.SubjectDomain() == ReferenceSubjectPermanent {
				t.Fatalf("card put into hand became an entered permanent: %+v", reference)
			}
		}
	}
}

func castEntryReferences(t *testing.T, chapter string) []CompiledReference {
	t.Helper()
	compilation, _ := compileSource(
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\nI — Draw a card.\nII — "+chapter+"\nIII — Draw a card.",
		pipelineContext{CardName: "Cast Entry", Saga: true})
	var references []CompiledReference
	for _, ability := range compilation.Abilities {
		for _, effect := range ability.Content.Effects {
			if effect.DelayedTriggerAbility == nil {
				continue
			}
			inner, _ := effect.DelayedTriggerAbility.Inner()
			nested, _ := Compile(inner, Context{})
			for _, child := range nested.Abilities {
				references = append(references, child.Content.References...)
			}
		}
	}
	return references
}

func TestReferenceRoleCastEntryIsPolicyNotMutation(t *testing.T) {
	references := castEntryReferences(t,
		"When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.")
	supported := 0
	for _, reference := range references {
		if !reference.SubjectSupported() {
			continue
		}
		supported++
		if reference.SubjectDomain() != ReferenceSubjectStackObject || !reference.SupportsUse(ReferenceUsePolicy) ||
			reference.SupportsUse(ReferenceUseMutation) {
			t.Fatalf("cast creature's entry replacement gained live authority: %+v", reference)
		}
	}
	if supported == 0 {
		t.Fatal("complete cast-entry composition lost its stack subject")
	}
	for _, reference := range castEntryReferences(t,
		"When you next cast an artifact spell this turn, that creature enters with an additional +1/+1 counter on it.") {
		if reference.SubjectNoun == "ObjectNounCreature" && reference.SubjectSupported() {
			t.Fatalf("creature noun was accepted for a noncreature cast: %+v", reference)
		}
	}
}

func TestReferenceRoleCompoundPriorSubjectDamageAttribution(t *testing.T) {
	content := compiledRoleContent(t,
		"{U}{U}: This creature gets -1/-0 until end of turn and deals 1 damage to target attacking creature without flying.",
		pipelineContext{CardName: "Marjhan"}, 0)
	attributed := 0
	for _, reference := range content.References {
		if reference.DamageAttribution() == DamageAttributionOriginalObject {
			attributed++
		}
	}
	if attributed != 1 {
		t.Fatalf("compound subject damage source attributions = %d, want 1", attributed)
	}
	content = compiledRoleContent(t,
		"{U}{U}: This creature gets -1/-0 until end of turn. Target creature gets +1/+0 until end of turn.",
		pipelineContext{CardName: "Near Miss"}, 0)
	for _, reference := range content.References {
		if reference.DamageAttribution() != DamageAttributionUnsupported {
			t.Fatalf("non-damage subject gained damage attribution: %+v", reference)
		}
	}
}
