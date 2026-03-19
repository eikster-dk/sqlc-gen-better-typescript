import { describe, beforeAll, afterAll, expect, it } from "@effect/vitest"
import { Effect, Layer, Option } from "effect"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import { SqlError } from "effect/unstable/sql"
import { startPostgres, stopPostgres, makeTestLayer } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  CustomersRepository,
  customersRepositoryLive,
} from "../../src/repositories/CustomersRepository.js"

describe("CustomersRepository", () => {
  let container: StartedPostgreSqlContainer
  let testLayer: Layer.Layer<CustomersRepository, SqlError.SqlError>

  beforeAll(async () => {
    container = await startPostgres()
    const sqlLayer = makeTestLayer(container)
    testLayer = customersRepositoryLive.pipe(Layer.provide(sqlLayer))

    // Seed the database
    await Effect.runPromise(
      seedDatabase.pipe(Effect.provide(sqlLayer))
    )
  }, 120000)

  afterAll(async () => {
    await stopPostgres()
  })

  describe("getCustomer", () => {
    it.effect("returns customer when found", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomer({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        expect(Option.getOrNull(result)).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when customer not found", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomer({ id: 9999 })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getCustomerByEmail", () => {
    it.effect("returns customer when found by email", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomerByEmail({ email: "alice@example.com" })

        expect(Option.isSome(result)).toBe(true)
        expect(Option.getOrNull(result)).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when email not found", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomerByEmail({ email: "nonexistent@example.com" })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listCustomers", () => {
    it.effect("returns all customers ordered by created_at DESC", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.listCustomers()

        expect(result.length).toBe(10)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listCustomersPaginated", () => {
    it.effect("returns paginated customers - first page", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.listCustomersPaginated({ limit: 3, offset: 0 })

        expect(result.length).toBe(3)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns paginated customers - second page", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.listCustomersPaginated({ limit: 3, offset: 3 })

        expect(result.length).toBe(3)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array when offset exceeds data", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.listCustomersPaginated({ limit: 10, offset: 100 })

        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("searchCustomersByName", () => {
    it.effect("finds customers matching name pattern", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.searchCustomersByName({ arg1: "son" })

        // Should find Alice Johnson, Jack Anderson
        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array when no match", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.searchCustomersByName({ arg1: "xyz123" })

        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("createCustomer", () => {
    it.effect("creates a new customer with all fields", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.createCustomer({
          email: "newcustomer@example.com",
          name: "New Customer",
          phone: "+1-555-9999",
        })

        expect(Option.isSome(result)).toBe(true)
        const customer = Option.getOrNull(result)!
        expect(customer.email).toBe("newcustomer@example.com")
        expect(customer.name).toBe("New Customer")
        expect(Option.getOrNull(customer.phone)).toBe("+1-555-9999")
        expect(customer.id).toBeGreaterThan(10)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("creates a new customer without phone", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.createCustomer({
          email: "nophone@example.com",
          name: "No Phone Customer",
        })

        expect(Option.isSome(result)).toBe(true)
        const customer = Option.getOrNull(result)!
        expect(customer.email).toBe("nophone@example.com")
        expect(Option.isNone(customer.phone)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateCustomer", () => {
    it.effect("updates an existing customer", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.updateCustomer({
          id: 1,
          email: "alice.updated@example.com",
          name: "Alice Johnson Updated",
          phone: "+1-555-0001",
        })

        expect(Option.isSome(result)).toBe(true)
        const customer = Option.getOrNull(result)!
        expect(customer.email).toBe("alice.updated@example.com")
        expect(customer.name).toBe("Alice Johnson Updated")
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when updating non-existent customer", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.updateCustomer({
          id: 9999,
          email: "nonexistent@example.com",
          name: "Nonexistent",
        })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateCustomerEmail", () => {
    it.effect("updates customer email only", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository

        // First update the email
        yield* repo.updateCustomerEmail({
          id: 2,
          email: "bob.newemail@example.com",
        })

        // Then verify the change
        const customer = yield* repo.getCustomer({ id: 2 })
        expect(Option.isSome(customer)).toBe(true)
        expect(Option.getOrNull(customer)!.email).toBe("bob.newemail@example.com")
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("deleteCustomer", () => {
    it.effect("deletes an existing customer", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository

        // Create a customer to delete
        const created = yield* repo.createCustomer({
          email: "todelete@example.com",
          name: "To Delete",
        })
        const customerId = Option.getOrNull(created)!.id

        // Delete it
        yield* repo.deleteCustomer({ id: customerId })

        // Verify it's gone
        const result = yield* repo.getCustomer({ id: customerId })
        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("countCustomers", () => {
    it.effect("returns total customer count", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.countCustomers()

        expect(Option.isSome(result)).toBe(true)
        const count = Option.getOrNull(result)!
        // Count may vary due to createCustomer tests, but should be >= 10
        expect(count.total).toBeGreaterThanOrEqual(10n)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getCustomersByIds", () => {
    it.effect("returns customers for given IDs", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomersByIds({ ids: [1, 3, 5] })

        expect(result.length).toBe(3)
        // Validate structure without volatile timestamps
        const ids = result.map(c => c.id).sort((a, b) => a - b)
        expect(ids).toEqual([1, 3, 5])
        expect(result.every(c => c.email && c.name)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array for non-existent IDs", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomersByIds({ ids: [9998, 9999] })

        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns partial results for mixed IDs", () =>
      Effect.gen(function* () {
        const repo = yield* CustomersRepository
        const result = yield* repo.getCustomersByIds({ ids: [1, 9999, 2] })

        expect(result.length).toBe(2)
      }).pipe(Effect.provide(testLayer))
    )
  })
})
