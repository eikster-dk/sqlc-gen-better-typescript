import { PostgreSqlContainer, type StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import { Layer, Redacted } from "effect"
import { PgClient } from "@effect/sql-pg"
import { SqlClient, SqlError } from "effect/unstable/sql"
import * as fs from "node:fs"
import * as path from "node:path"

let container: StartedPostgreSqlContainer | null = null

export const startPostgres = async (): Promise<StartedPostgreSqlContainer> => {
  container = await new PostgreSqlContainer("postgres:16-alpine")
    .withDatabase("testdb")
    .withUsername("test")
    .withPassword("test")
    .start()

  // Run schema migrations
  const schemaDir = path.join(import.meta.dirname, "../../schema")
  const schemaFiles = fs.readdirSync(schemaDir).sort()

  for (const file of schemaFiles) {
    if (file.endsWith(".sql")) {
      const sql = fs.readFileSync(path.join(schemaDir, file), "utf-8")
      await container.exec([
        "psql",
        "-U", "test",
        "-d", "testdb",
        "-c", sql,
      ])
    }
  }

  return container
}

export const stopPostgres = async (): Promise<void> => {
  if (container) {
    await container.stop()
    container = null
  }
}

export const makeTestLayer = (container: StartedPostgreSqlContainer): Layer.Layer<SqlClient.SqlClient | PgClient.PgClient, SqlError.SqlError> => {
  return PgClient.layer({
    url: Redacted.make(container.getConnectionUri()),
  })
}
