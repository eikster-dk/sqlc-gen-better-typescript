import { describe, beforeAll, afterAll, expect, it } from "vitest"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import pg from "pg"
import { startPostgres, stopPostgres, makePool } from "../setup/testcontainers.js"
import { getCustomer } from "../../src/customersQueries.js"

const { Pool } = pg

describe("customersQueries", () => {
  let container: StartedPostgreSqlContainer
  let pool: pg.Pool

  beforeAll(async () => {
    container = await startPostgres()
    pool = makePool(container)

    // Seed a customer
    await pool.query(`
      INSERT INTO customers (id, email, name, phone, created_at, updated_at)
      VALUES (1, 'alice@example.com', 'Alice Johnson', '+1-555-0101',
              '2024-01-01 10:00:00+00', '2024-01-01 10:00:00+00')
    `)
    await pool.query(`SELECT setval('customers_id_seq', 1)`)
  }, 120000)

  afterAll(async () => {
    await pool.end()
    await stopPostgres()
  })

  describe("getCustomer", () => {
    it("returns the customer when found", async () => {
      const result = await getCustomer(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data?.id).toBe(1)
      expect(result.data?.email).toBe("alice@example.com")
      expect(result.data?.name).toBe("Alice Johnson")
      expect(result.data?.phone).toBe("+1-555-0101")
      expect(result.data?.created_at).toBeInstanceOf(Date)
      expect(result.data?.updated_at).toBeInstanceOf(Date)
    })

    it("returns null data when customer not found", async () => {
      const result = await getCustomer(pool, { id: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })

    it("returns an input error when params fail validation", async () => {
      // Pass invalid params (non-integer id)
      const result = await getCustomer(pool, { id: "not-a-number" as unknown as number })

      expect(result.success).toBe(false)
      if (result.success) throw new Error("expected failure")

      expect(result.phase).toBe("input")
      expect(result.error).toBeDefined()
    })
  })
})
