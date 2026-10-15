// Command installcompiled adds a pinned compiler output without overwriting
// existing card implementations. It does not compile or interpret Oracle text.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/cardlist"
	"github.com/natefinch/council4/internal/preservation"
)

type installation struct {
	CorpusSHA256 string          `json:"corpus_sha256"`
	Cards        []installedCard `json:"cards"`
}

type installedCard struct {
	OracleID    string `json:"oracle_id"`
	Name        string `json:"name"`
	Layout      string `json:"layout"`
	Path        string `json:"path"`
	Constructor string `json:"constructor"`
	SHA256      string `json:"sha256"`
}

func main() {
	source := flag.String("source", ".cardwork/compiled-import", "fresh compilecards output")
	inventory := flag.String("inventory", ".cardwork/compiled-import-inventory.json", "matching pinned compiler inventory")
	destination := flag.String("destination", "mtg/cards", "card package root")
	manifest := flag.String("manifest", "testdata/card-preservation/installed-compiler.json", "installed source coverage")
	flag.Parse()
	if err := run(*source, *inventory, *destination, *manifest); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source, inventoryPath, destination, manifestPath string) error {
	inventory, err := preservation.Read[preservation.Manifest](inventoryPath)
	if err != nil {
		return err
	}
	pin, err := preservation.Read[preservation.Pin]("testdata/card-preservation/corpus.json")
	if err != nil {
		return err
	}
	if inventory.Input != pin {
		return errors.New("compiler inventory does not match the committed corpus pin")
	}
	if err := inventory.Validate(); err != nil {
		return err
	}
	accepted := make(map[string]preservation.Record)
	for _, record := range inventory.Compiler {
		if len(record.Refusals) == 0 {
			if _, duplicate := accepted[record.DefinitionSHA256]; duplicate {
				return fmt.Errorf("ambiguous generated source for %s", record.Name)
			}
			accepted[record.DefinitionSHA256] = record
		}
	}
	installed := installation{CorpusSHA256: pin.JSONSHA256}
	sourceRoot, err := os.OpenRoot(source)
	if err != nil {
		return err
	}
	defer sourceRoot.Close()
	if err := os.MkdirAll(destination, 0o750); err != nil {
		return err
	}
	destinationRoot, err := os.OpenRoot(destination)
	if err != nil {
		return err
	}
	defer destinationRoot.Close()
	directories := make(map[string]bool)
	symbols := make(map[string]map[string]string)
	added := 0
	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || entry.Name() == "cards.go" || entry.Name() == "doc.go" {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		directory := filepath.Dir(relative)
		if !cardDirectory(directory) {
			return fmt.Errorf("unexpected generated card directory %s", directory)
		}
		data, err := sourceRoot.ReadFile(relative)
		if err != nil {
			return err
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(data))
		record, exists := accepted[digest]
		if !exists {
			return fmt.Errorf("source %s has no accepted compiler identity", relative)
		}
		delete(accepted, digest)
		constructor, err := exportedConstructor(path)
		if err != nil {
			return err
		}
		dir := filepath.Join(destination, directory)
		if err := destinationRoot.MkdirAll(directory, 0o750); err != nil {
			return err
		}
		// Older generated filenames may transliterate accents differently.
		// Preserve their implementation by constructor, not filename alone.
		if symbols[directory] == nil {
			symbols[directory], err = existingConstructors(dir)
			if err != nil {
				return err
			}
		}
		target, exists := symbols[directory][constructor]
		if !exists {
			target = filepath.Join(destination, relative)
			if _, err := os.Stat(target); err == nil {
				return fmt.Errorf("generated path %s collides with a different existing constructor", target)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		installedPath, err := filepath.Rel(destination, target)
		if err != nil {
			return err
		}
		current, err := destinationRoot.ReadFile(installedPath)
		if errors.Is(err, os.ErrNotExist) {
			if err := destinationRoot.WriteFile(installedPath, data, 0o600); err != nil {
				return err
			}
			current = data
			added++
			symbols[directory][constructor] = target
		} else if err != nil {
			return err
		}
		installed.Cards = append(installed.Cards, installedCard{
			OracleID: record.Key, Name: record.Name, Layout: record.Layout,
			Path: filepath.ToSlash(installedPath), Constructor: constructor,
			SHA256: fmt.Sprintf("%x", sha256.Sum256(current)),
		})
		directories[directory] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(accepted) != 0 {
		return fmt.Errorf("%d compiler-accepted sources are missing", len(accepted))
	}
	for directory := range directories {
		dir := filepath.Join(destination, directory)
		source, err := cardlist.Generate(dir, filepath.Base(dir), "gencardlist")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "cards.go"), source, 0o600); err != nil {
			return err
		}
	}
	slices.SortFunc(installed.Cards, func(a, b installedCard) int {
		return strings.Compare(a.OracleID, b.OracleID)
	})
	data, err := json.MarshalIndent(installed, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(manifestPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	fmt.Printf("Installed %d new sources; retained %d existing implementations; %d accepted identities accounted for\n",
		added, len(installed.Cards)-added, len(installed.Cards))
	return nil
}

func existingConstructors(directory string) (map[string]string, error) {
	matches, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		return nil, err
	}
	symbols := make(map[string]string)
	for _, path := range matches {
		if filepath.Base(path) == "cards.go" || filepath.Base(path) == "doc.go" || strings.HasSuffix(path, "_test.go") {
			continue
		}
		symbol, err := exportedConstructor(path)
		if err != nil {
			return nil, err
		}
		if _, duplicate := symbols[symbol]; duplicate {
			return nil, fmt.Errorf("duplicate existing constructor %s", symbol)
		}
		symbols[symbol] = path
	}
	return symbols, nil
}

func cardDirectory(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) == 2 && parts[0] == "tokens" {
		parts = parts[1:]
	}
	return len(parts) == 1 && len(parts[0]) == 1 && parts[0][0] >= 'a' && parts[0][0] <= 'z'
}

func exportedConstructor(path string) (string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return "", err
	}
	var symbol string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || !ast.IsExported(value.Names[0].Name) {
				continue
			}
			if symbol != "" {
				return "", fmt.Errorf("%s has multiple exported constructors", path)
			}
			symbol = value.Names[0].Name
		}
	}
	if symbol == "" {
		return "", fmt.Errorf("%s has no exported constructor", path)
	}
	return symbol, nil
}
