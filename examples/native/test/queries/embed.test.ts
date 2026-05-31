import { describe, beforeAll, afterAll, expect, it } from "vitest"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import pg from "pg"
import { startPostgres, stopPostgres, makePool } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  getOrderWithCustomerEmbed,
  listOrdersWithCustomerEmbed,
} from "../../src/embedQueries.js"

describe("EmbedQueries", () => {
  let container: StartedPostgreSqlContainer
  let pool: pg.Pool

  beforeAll(async () => {
    container = await startPostgres()
    pool = makePool(container)
    await seedDatabase(pool)
  }, 120000)

  afterAll(async () => {
    await pool.end()
    await stopPostgres()
  })

  describe("getOrderWithCustomerEmbed", () => {
    it("returns nested order and customer when found", async () => {
      const result = await getOrderWithCustomerEmbed(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      const row = result.data!

      expect(row.order.id).toBe(1)
      expect(row.order.customer_id).toBe(1)
      expect(row.order.status).toBe("delivered")
      expect(row.order.total_cents).toBe(19998)
      expect(row.order.shipping_address).toBe("123 Main St, New York, NY 10001")

      expect(row.customer.id).toBe(1)
      expect(row.customer.email).toBe("alice@example.com")
      expect(row.customer.name).toBe("Alice Johnson")
      expect(row.customer.phone).toBe("+1-555-0101")

      expect((row as Record<string, unknown>)["orders_id"]).toBeUndefined()
      expect((row as Record<string, unknown>)["customers_id"]).toBeUndefined()
      expect(row).toMatchSnapshot()
    })

    it("returns null when order not found", async () => {
      const result = await getOrderWithCustomerEmbed(pool, { id: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")
      expect(result.data).toBeNull()
    })
  })

  describe("listOrdersWithCustomerEmbed", () => {
    it("returns all embedded rows with nested mapping and ordering", async () => {
      const result = await listOrdersWithCustomerEmbed(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(15)
      expect(result.data[0]!.order.id).toBe(15)
      expect(result.data[0]!.customer.id).toBe(10)
      expect(result.data.every((row) => row.order.customer_id === row.customer.id)).toBe(true)

      const orderOne = result.data.find((row) => row.order.id === 1)
      expect(orderOne).toBeDefined()
      expect(orderOne!.customer.email).toBe("alice@example.com")
      expect(orderOne!.customer.name).toBe("Alice Johnson")

      expect(result.data).toMatchSnapshot()
    })
  })
})
