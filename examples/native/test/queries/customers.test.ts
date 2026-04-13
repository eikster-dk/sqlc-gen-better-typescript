import { describe, beforeAll, afterAll, expect, it } from "vitest"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import pg from "pg"
import { startPostgres, stopPostgres, makePool } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  getCustomer,
  getCustomerByEmail,
  listCustomers,
  listCustomersPaginated,
  searchCustomersByName,
  createCustomer,
  updateCustomer,
  updateCustomerEmail,
  deleteCustomer,
  countCustomers,
  getCustomersByIds,
} from "../../src/customersQueries.js"

describe("CustomersQueries", () => {
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

  describe("getCustomer", () => {
    it("returns customer when found", async () => {
      const result = await getCustomer(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })

    it("returns null when customer not found", async () => {
      const result = await getCustomer(pool, { id: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("getCustomerByEmail", () => {
    it("returns customer when found by email", async () => {
      const result = await getCustomerByEmail(pool, { email: "alice@example.com" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })

    it("returns null when email not found", async () => {
      const result = await getCustomerByEmail(pool, { email: "nonexistent@example.com" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("listCustomers", () => {
    it("returns all customers ordered by created_at DESC", async () => {
      const result = await listCustomers(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(10)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listCustomersPaginated", () => {
    it("returns paginated customers - first page", async () => {
      const result = await listCustomersPaginated(pool, { limit: 3, offset: 0 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(3)
      expect(result.data).toMatchSnapshot()
    })

    it("returns paginated customers - second page", async () => {
      const result = await listCustomersPaginated(pool, { limit: 3, offset: 3 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(3)
      expect(result.data).toMatchSnapshot()
    })

    it("returns empty array when offset exceeds data", async () => {
      const result = await listCustomersPaginated(pool, { limit: 10, offset: 100 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })

  describe("searchCustomersByName", () => {
    it("finds customers matching name pattern", async () => {
      const result = await searchCustomersByName(pool, { arg1: "son" })

      // Should find Alice Johnson, Jack Anderson
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })

    it("returns empty array when no match", async () => {
      const result = await searchCustomersByName(pool, { arg1: "xyz123" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })

  describe("createCustomer", () => {
    it("creates a new customer with all fields", async () => {
      const result = await createCustomer(pool, {
        email: "newcustomer@example.com",
        name: "New Customer",
        phone: "+1-555-9999",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.email).toBe("newcustomer@example.com")
      expect(result.data!.name).toBe("New Customer")
      expect(result.data!.phone).toBe("+1-555-9999")
      expect(result.data!.id).toBeGreaterThan(10)
    })

    it("creates a new customer without phone", async () => {
      const result = await createCustomer(pool, {
        email: "nophone@example.com",
        name: "No Phone Customer",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.email).toBe("nophone@example.com")
      expect(result.data!.phone).toBeNull()
    })
  })

  describe("updateCustomer", () => {
    it("updates an existing customer", async () => {
      const result = await updateCustomer(pool, {
        id: 1,
        email: "alice.updated@example.com",
        name: "Alice Johnson Updated",
        phone: "+1-555-0001",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.email).toBe("alice.updated@example.com")
      expect(result.data!.name).toBe("Alice Johnson Updated")
    })

    it("returns null when updating non-existent customer", async () => {
      const result = await updateCustomer(pool, {
        id: 9999,
        email: "nonexistent@example.com",
        name: "Nonexistent",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("updateCustomerEmail", () => {
    it("updates customer email only", async () => {
      // First update the email
      await updateCustomerEmail(pool, {
        id: 2,
        email: "bob.newemail@example.com",
      })

      // Then verify the change
      const customer = await getCustomer(pool, { id: 2 })
      expect(customer.success).toBe(true)
      if (!customer.success) throw new Error("expected success")
      expect(customer.data!.email).toBe("bob.newemail@example.com")
    })
  })

  describe("deleteCustomer", () => {
    it("deletes an existing customer", async () => {
      // Create a customer to delete
      const created = await createCustomer(pool, {
        email: "todelete@example.com",
        name: "To Delete",
      })
      expect(created.success).toBe(true)
      if (!created.success) throw new Error("expected success")
      const customerId = created.data!.id

      // Delete it
      await deleteCustomer(pool, { id: customerId })

      // Verify it's gone
      const result = await getCustomer(pool, { id: customerId })
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")
      expect(result.data).toBeNull()
    })
  })

  describe("countCustomers", () => {
    it("returns total customer count", async () => {
      const result = await countCustomers(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      // Count may vary due to createCustomer tests, but should be >= 10
      expect(Number(result.data!.total)).toBeGreaterThanOrEqual(10)
    })
  })

  describe("getCustomersByIds", () => {
    it("returns customers for given IDs", async () => {
      const result = await getCustomersByIds(pool, { ids: [1, 3, 5] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(3)
      // Validate structure without volatile timestamps
      const ids = result.data.map(c => c.id).sort((a, b) => a - b)
      expect(ids).toEqual([1, 3, 5])
      expect(result.data.every(c => c.email && c.name)).toBe(true)
    })

    it("returns empty array for non-existent IDs", async () => {
      const result = await getCustomersByIds(pool, { ids: [9998, 9999] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })

    it("returns partial results for mixed IDs", async () => {
      const result = await getCustomersByIds(pool, { ids: [1, 9999, 2] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(2)
    })
  })
})
