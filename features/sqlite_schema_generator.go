package features

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type IDatabaseSchemaGenerator interface {
	GenerateSchemaInfo() (*SchemaInfo, error)
	GenerateSchemaSQL() (string, error)
}

type SchemaInfo struct {
	Tables      []string                    `json:"tables"`
	Columns     map[string][]ColumnInfo     `json:"columns"`
	Indexes     map[string][]IndexInfo      `json:"indexes"`
	ForeignKeys map[string][]ForeignKeyInfo `json:"foreignKeys"`
}

type ColumnInfo struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	NotNull      bool           `json:"notNull"`
	DefaultValue sql.NullString `json:"defaultValue"`
	PrimaryKey   bool           `json:"primaryKey"`
}

type IndexInfo struct {
	Name    string   `json:"name"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
}

type ForeignKeyInfo struct {
	ID              int    `json:"id"`
	Seq             int    `json:"seq"`
	ReferencedTable string `json:"referencedTable"`
	FromColumn      string `json:"fromColumn"`
	ToColumn        string `json:"toColumn"`
	OnUpdate        string `json:"onUpdate"`
	OnDelete        string `json:"onDelete"`
	Match           string `json:"match"`
}

func NewSqliteSchemaGenerator(db *sql.DB) IDatabaseSchemaGenerator {
	return &SqliteSchemaGenerator{
		db: db,
	}
}

type SqliteSchemaGenerator struct {
	db *sql.DB
}

func (g *SqliteSchemaGenerator) GenerateSchemaInfo() (*SchemaInfo, error) {

	// Initialize schema structure
	schema := &SchemaInfo{
		Tables:      []string{},
		Columns:     make(map[string][]ColumnInfo),
		Indexes:     make(map[string][]IndexInfo),
		ForeignKeys: make(map[string][]ForeignKeyInfo),
	}

	// Get all tables
	rows, err := g.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}
		schema.Tables = append(schema.Tables, tableName)
	}

	// For each table, get columns, indexes, and foreign keys
	for _, table := range schema.Tables {
		// Get columns
		columnRows, err := g.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
		if err != nil {
			return nil, fmt.Errorf("error querying columns for table %s: %w", table, err)
		}

		var columns []ColumnInfo
		for columnRows.Next() {
			var id int
			var name, dataType string
			var notNull int
			var defaultValue sql.NullString
			var primaryKey int

			if err := columnRows.Scan(&id, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
				columnRows.Close()
				return nil, fmt.Errorf("error scanning column info: %w", err)
			}

			columns = append(columns, ColumnInfo{
				ID:           id,
				Name:         name,
				Type:         dataType,
				NotNull:      notNull == 1,
				DefaultValue: defaultValue,
				PrimaryKey:   primaryKey == 1,
			})
		}
		columnRows.Close()
		schema.Columns[table] = columns

		// Get column names for reference in indexes
		columnNames := make(map[int]string)
		for _, col := range columns {
			columnNames[col.ID] = col.Name
		}

		// Get indexes
		indexRows, err := g.db.Query(fmt.Sprintf("PRAGMA index_list(%s)", table))
		if err != nil {
			return nil, fmt.Errorf("error querying indexes for table %s: %w", table, err)
		}

		var indexes []IndexInfo
		for indexRows.Next() {
			var seq int
			var name string
			var unique bool
			var origin, partial string

			// Different SQLite versions might have different number of columns
			if err := indexRows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
				// Try with fewer columns if the above fails
				indexRows.Close()

				// Reopen a new query and try with fewer columns
				indexRows, err = g.db.Query(fmt.Sprintf("PRAGMA index_list(%s)", table))
				if err != nil {
					return nil, fmt.Errorf("error requerying indexes: %w", err)
				}

				for indexRows.Next() {
					if err := indexRows.Scan(&seq, &name, &unique); err != nil {
						indexRows.Close()
						return nil, fmt.Errorf("error scanning index info with reduced columns: %w", err)
					}

					// Get columns in this index
					var indexColumns []string
					indexInfoRows, err := g.db.Query(fmt.Sprintf("PRAGMA index_info(%s)", name))
					if err != nil {
						return nil, fmt.Errorf("error querying index info: %w", err)
					}

					for indexInfoRows.Next() {
						var indexSeq, cid int
						var colName string
						if err := indexInfoRows.Scan(&indexSeq, &cid, &colName); err != nil {
							indexInfoRows.Close()
							return nil, fmt.Errorf("error scanning index column info: %w", err)
						}
						indexColumns = append(indexColumns, colName)
					}
					indexInfoRows.Close()

					indexes = append(indexes, IndexInfo{
						Name:    name,
						Unique:  unique,
						Columns: indexColumns,
					})
				}
				continue
			}

			// Process normally if the full scan worked
			// Get columns in this index
			var indexColumns []string
			indexInfoRows, err := g.db.Query(fmt.Sprintf("PRAGMA index_info(%s)", name))
			if err != nil {
				return nil, fmt.Errorf("error querying index info: %w", err)
			}

			for indexInfoRows.Next() {
				var indexSeq, cid int
				var colName string
				if err := indexInfoRows.Scan(&indexSeq, &cid, &colName); err != nil {
					indexInfoRows.Close()
					return nil, fmt.Errorf("error scanning index column info: %w", err)
				}
				indexColumns = append(indexColumns, colName)
			}
			indexInfoRows.Close()

			indexes = append(indexes, IndexInfo{
				Name:    name,
				Unique:  unique,
				Columns: indexColumns,
			})
		}
		indexRows.Close()
		schema.Indexes[table] = indexes

		// Get foreign keys
		fkRows, err := g.db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s)", table))
		if err != nil {
			return nil, fmt.Errorf("error querying foreign keys for table %s: %w", table, err)
		}

		var foreignKeys []ForeignKeyInfo
		for fkRows.Next() {
			var id, seq int
			var refTable, fromCol, toCol, onUpdate, onDelete, match string

			if err := fkRows.Scan(&id, &seq, &refTable, &fromCol, &toCol, &onUpdate, &onDelete, &match); err != nil {
				fkRows.Close()
				return nil, fmt.Errorf("error scanning foreign key info: %w", err)
			}

			foreignKeys = append(foreignKeys, ForeignKeyInfo{
				ID:              id,
				Seq:             seq,
				ReferencedTable: refTable,
				FromColumn:      fromCol,
				ToColumn:        toCol,
				OnUpdate:        onUpdate,
				OnDelete:        onDelete,
				Match:           match,
			})
		}
		fkRows.Close()
		schema.ForeignKeys[table] = foreignKeys
	}

	return schema, nil
}

// generateCreateTableSQL generates CREATE TABLE SQL statements for all tables
func (g *SqliteSchemaGenerator) generateCreateTablesSQL(schema *SchemaInfo) []string {
	var statements []string

	// Process tables in dependency order to handle foreign keys correctly
	processed := make(map[string]bool)
	tablesToProcess := schema.Tables

	// Keep track of tables that couldn't be processed due to dependencies
	var skippedTables []string

	// We may need multiple passes to handle circular dependencies
	for len(tablesToProcess) > 0 && len(skippedTables) < len(tablesToProcess) {
		skippedTables = []string{}

		for _, table := range tablesToProcess {
			if processed[table] {
				continue
			}

			// Check if all referenced tables are already processed
			canProcess := true
			for _, fk := range schema.ForeignKeys[table] {
				if !processed[fk.ReferencedTable] && fk.ReferencedTable != table { // Allow self-referencing tables
					canProcess = false
					break
				}
			}

			if !canProcess {
				skippedTables = append(skippedTables, table)
				continue
			}

			// Generate CREATE TABLE statement
			statement := g.generateCreateTableStatement(table, schema)
			statements = append(statements, statement)
			processed[table] = true
		}

		tablesToProcess = skippedTables
	}

	// If we have tables that couldn't be processed due to circular dependencies
	// Process them without foreign key constraints, then add ALTER TABLE statements
	if len(skippedTables) > 0 {
		for _, table := range skippedTables {
			// Generate CREATE TABLE without foreign keys
			statement := g.generateCreateTableStatementNoFK(table, schema)
			statements = append(statements, statement)
			processed[table] = true

			// Add ALTER TABLE statements for foreign keys
			for _, fk := range schema.ForeignKeys[table] {
				alterStatement := fmt.Sprintf(
					"ALTER TABLE %s ADD CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES %s(%s)",
					g.quoteIdentifier(table),
					table,
					fk.FromColumn,
					g.quoteIdentifier(fk.FromColumn),
					g.quoteIdentifier(fk.ReferencedTable),
					g.quoteIdentifier(fk.ToColumn),
				)

				if fk.OnDelete != "" && fk.OnDelete != "NO ACTION" {
					alterStatement += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
				}

				if fk.OnUpdate != "" && fk.OnUpdate != "NO ACTION" {
					alterStatement += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
				}

				alterStatement += ";"
				statements = append(statements, alterStatement)
			}
		}
	}

	// Generate CREATE INDEX statements
	for _, table := range schema.Tables {
		for _, idx := range schema.Indexes[table] {
			// Skip indexes that might be for PRIMARY KEY or UNIQUE constraints
			// This is a heuristic and might need adjustment
			if strings.HasPrefix(idx.Name, "sqlite_autoindex_") {
				continue
			}

			uniqueStr := ""
			if idx.Unique {
				uniqueStr = " UNIQUE"
			}

			columns := make([]string, len(idx.Columns))
			for i, col := range idx.Columns {
				columns[i] = g.quoteIdentifier(col)
			}

			indexStatement := fmt.Sprintf(
				"CREATE%s INDEX %s ON %s (%s);",
				uniqueStr,
				g.quoteIdentifier(idx.Name),
				g.quoteIdentifier(table),
				strings.Join(columns, ", "),
			)
			statements = append(statements, indexStatement)
		}
	}

	return statements
}

// Generate CREATE TABLE SQL with foreign key constraints
func (g *SqliteSchemaGenerator) generateCreateTableStatement(table string, schema *SchemaInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", g.quoteIdentifier(table)))

	// Process columns
	columns := schema.Columns[table]
	primaryKeys := []string{}

	for i, col := range columns {
		sb.WriteString("    " + g.quoteIdentifier(col.Name) + " " + col.Type)

		if col.NotNull {
			sb.WriteString(" NOT NULL")
		}

		if col.DefaultValue.Valid {
			// Handle string vs numeric default values
			if _, err := strconv.ParseFloat(col.DefaultValue.String, 64); err == nil {
				// It's a number
				sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String))
			} else if strings.HasPrefix(col.DefaultValue.String, "CURRENT_") {
				// It's a function like CURRENT_TIMESTAMP
				sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String))
			} else {
				// It's a string
				sb.WriteString(fmt.Sprintf(" DEFAULT '%s'", g.escapeSingleQuotes(col.DefaultValue.String)))
			}
		}

		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, col.Name)
		}

		if i < len(columns)-1 || len(primaryKeys) > 0 || len(schema.ForeignKeys[table]) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// Add PRIMARY KEY constraint if there are multiple primary keys
	if len(primaryKeys) > 0 {
		quotedKeys := make([]string, len(primaryKeys))
		for i, key := range primaryKeys {
			quotedKeys[i] = g.quoteIdentifier(key)
		}
		sb.WriteString(fmt.Sprintf("    PRIMARY KEY (%s)", strings.Join(quotedKeys, ", ")))

		if len(schema.ForeignKeys[table]) > 0 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString("\n")
		}
	}

	// Add FOREIGN KEY constraints
	for i, fk := range schema.ForeignKeys[table] {
		sb.WriteString(fmt.Sprintf("    FOREIGN KEY (%s) REFERENCES %s(%s)",
			g.quoteIdentifier(fk.FromColumn),
			g.quoteIdentifier(fk.ReferencedTable),
			g.quoteIdentifier(fk.ToColumn),
		))

		if fk.OnDelete != "" && fk.OnDelete != "NO ACTION" {
			sb.WriteString(fmt.Sprintf(" ON DELETE %s", fk.OnDelete))
		}

		if fk.OnUpdate != "" && fk.OnUpdate != "NO ACTION" {
			sb.WriteString(fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate))
		}

		if i < len(schema.ForeignKeys[table])-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(");")
	return sb.String()
}

// Generate CREATE TABLE SQL without foreign key constraints
func (g *SqliteSchemaGenerator) generateCreateTableStatementNoFK(table string, schema *SchemaInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", g.quoteIdentifier(table)))

	// Process columns
	columns := schema.Columns[table]
	primaryKeys := []string{}

	for i, col := range columns {
		sb.WriteString("    " + g.quoteIdentifier(col.Name) + " " + col.Type)

		if col.NotNull {
			sb.WriteString(" NOT NULL")
		}

		if col.DefaultValue.Valid {
			// Handle string vs numeric default values
			if _, err := strconv.ParseFloat(col.DefaultValue.String, 64); err == nil {
				// It's a number
				sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String))
			} else if strings.HasPrefix(col.DefaultValue.String, "CURRENT_") {
				// It's a function like CURRENT_TIMESTAMP
				sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String))
			} else {
				// It's a string
				sb.WriteString(fmt.Sprintf(" DEFAULT '%s'", g.escapeSingleQuotes(col.DefaultValue.String)))
			}
		}

		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, col.Name)
		}

		if i < len(columns)-1 || len(primaryKeys) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// Add PRIMARY KEY constraint if there are multiple primary keys
	if len(primaryKeys) > 0 {
		quotedKeys := make([]string, len(primaryKeys))
		for i, key := range primaryKeys {
			quotedKeys[i] = g.quoteIdentifier(key)
		}
		sb.WriteString(fmt.Sprintf("    PRIMARY KEY (%s)\n", strings.Join(quotedKeys, ", ")))
	}

	sb.WriteString(");")
	return sb.String()
}

// Quote identifier if needed
func (g *SqliteSchemaGenerator) quoteIdentifier(id string) string {
	// If the identifier contains special characters or is a reserved keyword, quote it
	if strings.ContainsAny(id, " ,-+*/()[]{}.") ||
		strings.ToLower(id) == "table" ||
		strings.ToLower(id) == "index" ||
		strings.ToLower(id) == "select" ||
		strings.ToLower(id) == "where" ||
		strings.ToLower(id) == "from" ||
		strings.ToLower(id) == "join" {
		return fmt.Sprintf("\"%s\"", id)
	}
	return id
}

// Escape single quotes in string values
func (g *SqliteSchemaGenerator) escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// PrintSchemaAsSQL prints the database schema as SQL statements
func (g *SqliteSchemaGenerator) GenerateSchemaSQL() (string, error) {

	schema, err := g.GenerateSchemaInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get database schema: %w", err)
	}

	schemaStr := ""
	schemaStr += fmt.Sprintln("-- SQLite Database Schema")
	schemaStr += fmt.Sprintln("-- Generated by SQLite Schema Extractor")
	schemaStr += fmt.Sprintln("PRAGMA foreign_keys = ON;")
	schemaStr += fmt.Sprintln("")

	statements := g.generateCreateTablesSQL(schema)
	for _, stmt := range statements {
		schemaStr += fmt.Sprintln(stmt)
		schemaStr += fmt.Sprintln("")
	}

	return schemaStr, nil
}
