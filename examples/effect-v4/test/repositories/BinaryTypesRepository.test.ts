import { afterAll, beforeAll, describe, expect, it } from "@effect/vitest"
import { Effect, Layer, Option } from "effect"
import { SqlClient, SqlError } from "effect/sql"
import { BinaryTypesRepository } from "../../src/repositories/BinaryTypesRepository.js"
import { makeTestLayer, startPostgres, stopPostgres } from "../setup/testcontainers.js"

describe("binary PostgreSQL types", () => {
  let testLayer: Layer.Layer<BinaryTypesRepository | SqlClient.SqlClient, SqlError.SqlError>

  beforeAll(async () => {
    const container = await startPostgres()
    testLayer = BinaryTypesRepository.layer.pipe(Layer.provideMerge(makeTestLayer(container)))
  }, 120000)

  afterAll(stopPostgres)

  it.effect("round-trips signed int64 limits and decodes bigint arrays without precision loss", () =>
    Effect.gen(function* () {
      const repo = yield* BinaryTypesRepository
      const min = -9223372036854775808n
      const max = 9223372036854775807n
      const values = [min, -9007199254740993n, 0n, 9007199254740993n, max]

      for (const value of values) {
        const result = yield* repo.echoBigInts({ value })
        expect(result).toEqual(Option.some({ value, values: [value] }))
      }
    }).pipe(Effect.provide(testLayer))
  )

  it.effect("reads search vectors cast to text with the default driver codecs", () =>
    Effect.gen(function* () {
      const sql = yield* SqlClient.SqlClient
      const rows = yield* sql<{ id: number; text: string | null }>`
        SELECT id, vector::text AS text
        FROM (VALUES
          (1, ''::tsvector),
          (2, $$'a':1A,2B,3C,4 '猫'$$::tsvector),
          (3, $$'a''b\\\\c':16383A$$::tsvector),
          (4, to_tsvector('english', 'Leave at door')),
          (5, NULL::tsvector)
        ) AS vectors(id, vector)
        ORDER BY id
      `

      expect(rows).toEqual([
        { id: 1, text: "" },
        { id: 2, text: "'a':1A,2B,3C,4 '猫'" },
        { id: 3, text: "'a''b\\\\c':16383A" },
        { id: 4, text: "'door':3 'leav':1" },
        { id: 5, text: null },
      ])
    }).pipe(Effect.provide(testLayer))
  )
})
