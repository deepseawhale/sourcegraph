package definition

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/gitfs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func StitchDefinitions() (*Definitions, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	schemaName := "frontend"
	revs := []string{
		// "v3.36.0", // no directories
		// "v3.37.0", // no elevated permissions
		"v3.38.0",
		"v3.39.0",
		"v3.40.0",
		"HEAD",
	}

	definitionMap := map[int]Definition{}
	for _, rev := range revs {
		revDefinitions, err := readDefinitionsAtRevision(schemaName, wd, rev)
		if err != nil {
			return nil, err
		}

		for _, newDefinition := range revDefinitions {
			existingDefinition, ok := definitionMap[newDefinition.ID]
			if !ok {
				definitionMap[newDefinition.ID] = newDefinition
				continue
			}

			if !compareDefinitions(newDefinition, existingDefinition) && !strings.HasPrefix(newDefinition.Name, "squashed migrations") {
				// Ignore new squashed definitions that have re-used an old definition identifier
				return nil, errors.Newf("migration %d unexpectedly edited in release %s", newDefinition.ID, rev)
			}
		}
	}

	migrationDefinitions := make([]Definition, 0, len(definitionMap))
	for _, v := range definitionMap {
		migrationDefinitions = append(migrationDefinitions, v)
	}

	if err := reorderDefinitions(migrationDefinitions); err != nil {
		return nil, err
	}

	return newDefinitions(migrationDefinitions), nil
}

func readDefinitionsAtRevision(schemaName, wd, rev string) ([]Definition, error) {
	prefix := filepath.Join("migrations", schemaName)

	revDefinitions, err := readDefinitions(gitfs.New(rev, wd, prefix), prefix)
	if err != nil {
		return nil, errors.Wrap(err, "@"+rev)
	}

	return revDefinitions, nil
}

func compareDefinitions(x, y Definition) bool {
	return cmp.Diff(x, y, cmp.Comparer(func(x, y *sqlf.Query) bool {
		// Note: migrations do not have args to compare here, so we can compare only
		// the query text safely. If we ever need to add runtime arguments to the
		// migration runner, this assumption _might_ change.
		return x.Query(sqlf.PostgresBindVar) == y.Query(sqlf.PostgresBindVar)
	})) == ""
}
