import { describe, beforeAll, afterAll, expect, it } from "@effect/vitest"
import { Effect, Layer, Option } from "effect"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import { SqlError } from "effect/unstable/sql"
import { startPostgres, stopPostgres, makeTestLayer } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  EmbedRepository,
  embedRepositoryLive,
} from "../../src/repositories/EmbedRepository.js"

describe("EmbedRepository", () => {
  let container: StartedPostgreSqlContainer
  let testLayer: Layer.Layer<EmbedRepository, SqlError.SqlError>

  beforeAll(async () => {
    container = await startPostgres()
    const sqlLayer = makeTestLayer(container)
    testLayer = embedRepositoryLive.pipe(Layer.provide(sqlLayer))

    // Seed the database
    await Effect.runPromise(
      seedDatabase.pipe(Effect.provide(sqlLayer))
    )
  }, 120000)

  afterAll(async () => {
    await stopPostgres()
  })

  describe("getOrderWithCustomerEmbed", () => {
    it.effect("returns nested order and customer when found", () =>
      Effect.gen(function* () {
        const repo = yield* EmbedRepository
        const result = yield* repo.getOrderWithCustomerEmbed({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        const row = Option.getOrNull(result)!

        expect(row.order.id).toBe(1)
        expect(row.order.customer_id).toBe(1)
        expect(row.order.status).toBe("delivered")
        expect(row.order.total_cents).toBe(19998)
        expect(Option.getOrNull(row.order.shipping_address)).toBe("123 Main St, New York, NY 10001")

        expect(row.customer.id).toBe(1)
        expect(row.customer.email).toBe("alice@example.com")
        expect(row.customer.name).toBe("Alice Johnson")
        expect(Option.getOrNull(row.customer.phone)).toBe("+1-555-0101")

        expect((row as Record<string, unknown>)["orders_id"]).toBeUndefined()
        expect((row as Record<string, unknown>)["customers_id"]).toBeUndefined()
        expect(row).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when order not found", () =>
      Effect.gen(function* () {
        const repo = yield* EmbedRepository
        const result = yield* repo.getOrderWithCustomerEmbed({ id: 9999 })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrdersWithCustomerEmbed", () => {
    it.effect("returns all embedded rows with nested mapping and ordering", () =>
      Effect.gen(function* () {
        const repo = yield* EmbedRepository
        const result = yield* repo.listOrdersWithCustomerEmbed()

        expect(result.length).toBe(15)
        expect(result[0]!.order.id).toBe(15)
        expect(result[0]!.customer.id).toBe(10)
        expect(result.every((row) => row.order.customer_id === row.customer.id)).toBe(true)

        const orderOne = result.find((row) => row.order.id === 1)
        expect(orderOne).toBeDefined()
        expect(orderOne!.customer.email).toBe("alice@example.com")
        expect(orderOne!.customer.name).toBe("Alice Johnson")

        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })
})
