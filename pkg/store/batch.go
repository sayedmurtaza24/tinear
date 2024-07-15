package store

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
)

type idGetter interface{ getID() string }

var spaceRemoveRegEx = regexp.MustCompile("( |\t){2,}")

func prettify(query string) string {
	return spaceRemoveRegEx.ReplaceAllString(query, " ")
}

func removeDuplicates[T idGetter](list []T) []T {
	m := make(map[string]T)
	for _, item := range list {
		m[item.getID()] = item
	}
	var res []T
	for _, v := range m {
		res = append(res, v)
	}
	return res
}

func batchInsert[T idGetter](
	db *sql.DB,
	tableName string,
	columns []string,
	rows []T,
	rowArgs func(v *T) []any,
	onconflict string,
) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("couldn't begin tx: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf("\nINSERT INTO %s (%s)\nVALUES ", tableName, strings.Join(columns, ", "))

	hasOrgID := slices.Contains(columns, "org_id")

	var placeholders string
	if hasOrgID {
		placeholders = strings.Join(strings.Split(strings.Repeat("?", len(columns)-1), ""), ", ")
		placeholders += ", (SELECT id FROM orgs WHERE orgs.active = TRUE)"
	} else {
		placeholders = strings.Join(strings.Split(strings.Repeat("?", len(columns)), ""), ", ")
	}

	rows = removeDuplicates(rows)

	if len(rows) == 0 {
		return nil
	}

	separator := func(i int) string {
		if i != len(rows)-1 {
			return ", "
		}
		return " "
	}

	var args []any
	for i, value := range rows {
		rowArgs := rowArgs(&value)
		args = append(args, rowArgs...)
		query += fmt.Sprintf("(%s)%s\n", placeholders, separator(i))
	}

	query += onconflict

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("couldn't prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		log.Println(prettify(query))
		return fmt.Errorf("couldn't execute statement: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit tx: %w", err)
	}

	return nil
}

func batchSelect[T idGetter](
	db *sql.DB,
	tableName string,
	columns []string,
	scan func(v *T) []any,
	where string,
	whereArgs ...any,
) ([]T, error) {
	query := fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(columns, ", "), tableName, where)

	rows, err := db.Query(query, whereArgs...)
	if err != nil {
		log.Println(query)
		return nil, fmt.Errorf("couldn't query rows: %w", err)
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		var row T
		err := rows.Scan(scan(&row)...)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan row: %w", err)
		}
		result = append(result, row)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows returned err: %w", rows.Err())
	}

	return result, nil
}

func batchDelete(db *sql.DB, tableName string, where string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s %s", tableName, where))
	if err != nil {
		return fmt.Errorf("couldn't execute delete query")
	}
	return nil
}
