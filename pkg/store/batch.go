package store

import (
	"database/sql"
	"fmt"
	"strings"
)

func batchInsert[T any](
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

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(columns, ", "))

	placeholders := strings.Join(strings.Split(strings.Repeat("?", len(columns)), ""), ", ")

	separator := func(i int) string {
		if i != len(rows)-1 {
			return ", "
		}
		return " "
	}

	var args []any
	for i, value := range rows {
		rowArgs := rowArgs(&value)
		if len(rowArgs) != len(columns) {
			return fmt.Errorf("len args != len columns")
		}
		args = append(args, rowArgs...)
		query += fmt.Sprintf("(%s)%s", placeholders, separator(i))
	}

	query += onconflict

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("couldn't prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return fmt.Errorf("couldn't execute statement: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit tx: %w", err)
	}

	return nil
}

func batchSelect[T any](
	db *sql.DB,
	tableName string,
	columns []string,
	scan func(v *T) []any,
	where string,
) ([]T, error) {
	query := fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(columns, ", "), tableName, where)

	rows, err := db.Query(query)
	if err != nil {
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
		return nil, fmt.Errorf("rows returned err: %w", err)
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
