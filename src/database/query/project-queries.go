package query

import (
	"context"
	"goto/src/database"
	"goto/src/model"

	"github.com/jackc/pgx/v5"
)

func CreateProject(ctx context.Context, p model.Project) error {
	tx, err := database.ConnPool.Begin(ctx)
	defer tx.Rollback(ctx)

	projectEntries := [][]any{
		{p.Url, p.Dir, p.Name, p.Language, p.Containerization, p.SrcDir, p.StubDir},
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"project"},
		[]string{"url", "dir", "name", "language", "containerization", "srcdir", "stubdir"},
		pgx.CopyFromRows(projectEntries),
	)
	if err != nil {
		return err
	}

	var projectId int
	err = tx.QueryRow(ctx, "SELECT id FROM project WHERE dir = $1", p.Dir).Scan(&projectId)
	if err != nil {
		return err
	}

	moduleEntries := make([][]any, len(p.Modules))
	for i, mod := range p.Modules {
		moduleEntries[i] = []any{projectId, mod}
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"project_module"},
		[]string{"project_id", "name"},
		pgx.CopyFromRows(moduleEntries),
	)
	if err != nil {
		return err
	}

	taskEntries := make([][]any, len(p.Tasks))
	for i, t := range p.Tasks {
		taskEntries[i] = []any{projectId, t.Name, t.Description, t.RunTarget}
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"task"},
		[]string{"project_id", "name", "description", "runtarget"},
		pgx.CopyFromRows(taskEntries),
	)
	if err != nil {
		return err
	}

	injectFilesByTaskName := make(map[string]map[string]string)
	for _, t := range p.Tasks {
		injectFilesByTaskName[t.Name] = t.InjectFiles
	}

	injectFileEntries := [][]any{}
	rows, err := tx.Query(ctx, "SELECT * FROM task WHERE project_id = $1", projectId)
	_, err = pgx.ForEachRow(rows, []any{}, func() error {
		injectFileEntries = append(injectFileEntries, []any{fileName, filePath})
		return nil
	})
	if err != nil {
		return err
	}
	rows.Close()

	err = tx.Commit(ctx)
	return err
}

// func CreateProject(ctx context.Context, project model.Project) error {
// 	rows, err := database.ConnPool.Query(ctx, `
//         INSERT INTO project(url, dir, name, language, containerization, srcdir, stubdir)
//         VALUES ($1, $2, $3, $4)
//         ;`)
// 	if err != nil {
// 		return err
// 	}
// 	rows.Close()
//     return  nil
// }