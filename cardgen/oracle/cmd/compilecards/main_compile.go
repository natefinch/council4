package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func compileCorpus(input io.Reader, workers int) ([]result, error) {
	decoder := json.NewDecoder(input)
	tkn, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("reading bulk-data array: %w", err)
	}
	if delimiter, ok := tkn.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("bulk data from Scryfall must be a top-level JSON array")
	}

	jobs := make(chan job)
	results := make(chan result)
	var workersDone sync.WaitGroup
	workersDone.Add(workers)
	for range workers {
		go func() {
			defer workersDone.Done()
			for item := range jobs {
				results <- compileCard(item)
			}
		}()
	}
	go func() {
		workersDone.Wait()
		close(results)
	}()

	decodeError := make(chan error, 1)
	go func() {
		defer close(jobs)
		sent := 0
		for decoder.More() {
			var card cardgen.ScryfallCard
			if err := decoder.Decode(&card); err != nil {
				decodeError <- fmt.Errorf("decoding card %d: %w", sent, err)
				return
			}
			// Disowned cards are omitted entirely: never generated and never
			// listed as supported, unsupported, or excluded.
			if cardgen.DisownedCard(card) {
				continue
			}
			jobs <- job{index: sent, card: card}
			sent++
		}
		if _, err := decoder.Token(); err != nil {
			decodeError <- fmt.Errorf("closing bulk-data array: %w", err)
			return
		}
		decodeError <- nil
	}()

	var all []result
	for compiled := range results {
		all = append(all, compiled)
	}
	if err := <-decodeError; err != nil {
		return nil, err
	}
	disambiguateCollisions(all)
	rejectPathCollisions(all)
	rejectIdentifierCollisions(all)
	slices.SortFunc(all, func(a, b result) int {
		return cmp.Compare(a.index, b.index)
	})
	return all, nil
}

func compileCard(item job) result {
	defer func() {
		if recovered := recover(); recovered != nil {
			panic(compileCardPanicContext(item, recovered))
		}
	}()
	card := item.card
	compiled := result{index: item.index, card: card}
	if reason, excluded := (cardgen.CorpusPolicy{}).Exclusion(card); excluded {
		compiled.exclusion = reason
		return compiled
	}
	identity, err := cardgen.GeneratedIdentity(&card, false)
	if err != nil {
		compiled.diagnostics = []shared.Diagnostic{{
			Severity: shared.SeverityWarning,
			Summary:  "invalid generated identity",
			Detail:   err.Error(),
		}}
		return compiled
	}
	letter := identity.PackageName
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		compiled.diagnostics = []shared.Diagnostic{{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported package letter",
			Detail:   fmt.Sprintf("card name %q does not map to an ASCII a-z package", card.Name),
		}}
		return compiled
	}
	compiled.relative = identity.RelativePath
	compiled.superseded = identity.SupersededPath
	compiled.source, compiled.diagnostics, compiled.err =
		(cardgen.ExecutableGenerator{IdentifierSuffix: identity.IdentifierSuffix}).
			GenerateCardSource(&card, identity.PackageName)
	return compiled
}

func compileCardPanicContext(item job, recovered any) string {
	name := item.card.Name
	if name == "" {
		name = "<unnamed>"
	}
	oracleID := item.card.OracleID
	if oracleID == "" {
		oracleID = "<missing>"
	}
	return fmt.Sprintf("compiling card %d %q (oracle_id %s): %v", item.index, name, oracleID, recovered)
}

func disambiguateCollisions(results []result) {
	parent := make([]int, len(results))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(index int) int {
		if parent[index] != index {
			parent[index] = find(parent[index])
		}
		return parent[index]
	}
	union := func(indexes []int) {
		if len(indexes) < 2 {
			return
		}
		root := find(indexes[0])
		for _, index := range indexes[1:] {
			parent[find(index)] = root
		}
	}
	byPath := make(map[string][]int)
	byIdentifier := make(map[string][]int)
	for i := range results {
		if results[i].err != nil || len(results[i].diagnostics) > 0 {
			continue
		}
		byPath[results[i].relative] = append(byPath[results[i].relative], i)
		file, err := parser.ParseFile(token.NewFileSet(), results[i].relative, results[i].source, 0)
		if err != nil {
			results[i].err = fmt.Errorf("parsing generated source: %w", err)
			continue
		}
		for _, name := range cardDefNames(file) {
			key := filepath.Dir(results[i].relative) + "\x00" + name
			byIdentifier[key] = append(byIdentifier[key], i)
		}
	}
	for _, indexes := range byPath {
		union(indexes)
	}
	for _, indexes := range byIdentifier {
		union(indexes)
	}
	components := make(map[int][]int)
	for i := range results {
		if results[i].err == nil && len(results[i].diagnostics) == 0 {
			components[find(i)] = append(components[find(i)], i)
		}
	}
	colliding := make(map[int]bool)
	for _, indexes := range components {
		if len(indexes) < 2 {
			continue
		}
		slices.SortFunc(indexes, func(a, b int) int {
			aKey := cardIdentityKey(&results[a].card)
			bKey := cardIdentityKey(&results[b].card)
			if byIdentity := strings.Compare(aKey, bKey); byIdentity != 0 {
				return byIdentity
			}
			return cmp.Compare(results[a].index, results[b].index)
		})
		for _, index := range indexes[1:] {
			colliding[index] = true
		}
	}
	for index := range colliding {
		card := &results[index].card
		identity, err := cardgen.GeneratedIdentity(card, true)
		if err != nil {
			results[index].source = ""
			results[index].diagnostics = []shared.Diagnostic{{
				Severity: shared.SeverityWarning,
				Summary:  "generated identity collision",
				Detail:   err.Error(),
			}}
			continue
		}
		results[index].relative = identity.RelativePath
		results[index].superseded = identity.SupersededPath
		results[index].source, results[index].diagnostics, results[index].err =
			(cardgen.ExecutableGenerator{IdentifierSuffix: identity.IdentifierSuffix}).GenerateCardSource(
				card,
				identity.PackageName,
			)
	}
	finalPaths := make(map[string]bool)
	for i := range results {
		if results[i].err == nil && len(results[i].diagnostics) == 0 {
			finalPaths[results[i].relative] = true
		}
	}
	for i := range results {
		if finalPaths[results[i].superseded] {
			results[i].superseded = ""
		}
	}
}

func cardIdentityKey(card *cardgen.ScryfallCard) string {
	if card.OracleID != "" {
		return card.OracleID
	}
	return card.ID
}

func rejectPathCollisions(results []result) {
	byPath := make(map[string][]int)
	for i := range results {
		if results[i].exclusion == "" && results[i].err == nil && len(results[i].diagnostics) == 0 {
			byPath[results[i].relative] = append(byPath[results[i].relative], i)
		}
	}
	for path, indexes := range byPath {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			results[index].source = ""
			results[index].diagnostics = []shared.Diagnostic{{
				Severity: shared.SeverityWarning,
				Summary:  "generated path collision",
				Detail:   fmt.Sprintf("%d Oracle cards map to %s", len(indexes), path),
			}}
		}
	}
}

func rejectIdentifierCollisions(results []result) {
	byName := make(map[string][]int)
	for i := range results {
		if results[i].exclusion != "" || results[i].err != nil || len(results[i].diagnostics) > 0 {
			continue
		}
		file, err := parser.ParseFile(
			token.NewFileSet(),
			results[i].relative,
			results[i].source,
			0,
		)
		if err != nil {
			results[i].err = fmt.Errorf("parsing generated source: %w", err)
			continue
		}
		seen := make(map[string]bool)
		for _, name := range cardDefNames(file) {
			if seen[name] {
				results[i].source = ""
				results[i].diagnostics = []shared.Diagnostic{{
					Severity: shared.SeverityWarning,
					Summary:  "duplicate generated identifier",
					Detail:   fmt.Sprintf("generated source declares %s more than once", name),
				}}
				continue
			}
			seen[name] = true
			byName[name] = append(byName[name], i)
		}
	}
	for name, indexes := range byName {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			results[index].source = ""
			results[index].diagnostics = []shared.Diagnostic{{
				Severity: shared.SeverityWarning,
				Summary:  "generated identifier collision",
				Detail: fmt.Sprintf(
					"%d Oracle cards declare %s in package %s",
					len(indexes),
					name,
					filepath.Dir(results[index].relative),
				),
			}}
		}
	}
}
