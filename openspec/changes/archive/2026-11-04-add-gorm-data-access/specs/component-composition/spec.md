## ADDED Requirements

### Requirement: The database schema is the single source of truth
A data-access component SHALL read and write the tables defined by the schema
that the selected database component provides (trait `sql-schema`), and SHALL
NOT create, alter, or migrate tables itself. It SHALL use the connection the
database slot opens instead of opening its own.

#### Scenario: GORM uses the database schema
- **WHEN** a project is generated with `--db sqlite --data gorm` and the application starts on an empty database
- **THEN** the `items` table is created only by the database component's `schema.sql`
- **AND** an item created through the store gets its `id` and `created_at` from the database, as defined in that schema

#### Scenario: GORM wraps the opened connection
- **WHEN** the entry point passes the connection returned by the database slot to the store
- **THEN** the store issues its queries over that connection and opens no other database connection

### Requirement: Dialect-specific code is chosen from traits
A component that supports several values of a trait SHALL choose its
value-specific code (for example, a SQL dialect's driver) from the trait value
the selection provides, not from which other component is selected, so a newly
added component providing a supported value works without changes to it.

#### Scenario: SQLite dialect
- **WHEN** a project is generated with `--data gorm` and a database providing `sql-dialect: sqlite`
- **THEN** the store uses a pure-Go SQLite dialect for GORM and contains no PostgreSQL code

#### Scenario: PostgreSQL dialect
- **WHEN** a project is generated with `--data gorm` and a database providing `sql-dialect: postgresql` and `conn: database/sql`
- **THEN** the store uses GORM's PostgreSQL dialect over the same `*sql.DB`, and the project builds and passes `go vet`
