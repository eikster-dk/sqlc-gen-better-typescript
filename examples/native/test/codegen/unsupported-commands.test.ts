import { mkdtemp, mkdir, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join, resolve } from "node:path"
import { execFile } from "node:child_process"
import { promisify } from "node:util"
import { describe, expect, it } from "vitest"

const execFileAsync = promisify(execFile)

async function writeProject(query: string) {
  const dir = await mkdtemp(join(tmpdir(), "sqlc-native-unsupported-"))
  await mkdir(join(dir, "schema"))
  await mkdir(join(dir, "queries"))
  await writeFile(
    join(dir, "schema", "schema.sql"),
    `CREATE TABLE products (
  id integer PRIMARY KEY,
  name text NOT NULL
);
`,
  )
  await writeFile(join(dir, "queries", "queries.sql"), query)
  await writeFile(
    join(dir, "sqlc.yaml"),
    `version: '2'
plugins:
- name: native
  wasm:
    url: file://${resolve("../../dist/sqlc-gen-native.wasm")}
sql:
- schema: schema/
  queries: queries/
  engine: postgresql
  codegen:
  - out: src
    plugin: native
    options:
      import_extension: ".js"
`,
  )
  return dir
}

async function expectUnsupportedCommand(query: string, command: string) {
  const dir = await writeProject(query)

  await expect(execFileAsync("sqlc", ["generate"], { cwd: dir })).rejects.toMatchObject({
    stderr: expect.stringContaining(`unsupported sqlc command ${command}`),
  })
}

describe("native unsupported command codegen errors", () => {
  it("fails for :copyfrom", async () => {
    await expectUnsupportedCommand(
      `-- name: CopyProducts :copyfrom
INSERT INTO products (id, name) VALUES ($1, $2);
`,
      ":copyfrom",
    )
  })

  it("fails for :batchexec", async () => {
    await expectUnsupportedCommand(
      `-- name: BatchRenameProducts :batchexec
UPDATE products SET name = $2 WHERE id = $1;
`,
      ":batchexec",
    )
  })

  it("fails for :batchone", async () => {
    await expectUnsupportedCommand(
      `-- name: BatchGetProduct :batchone
SELECT id, name FROM products WHERE id = $1;
`,
      ":batchone",
    )
  })

  it("fails for :batchmany", async () => {
    await expectUnsupportedCommand(
      `-- name: BatchListProducts :batchmany
SELECT id, name FROM products WHERE id > $1;
`,
      ":batchmany",
    )
  })
})
