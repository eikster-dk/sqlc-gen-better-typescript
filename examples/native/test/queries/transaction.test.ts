import { describe, beforeAll, afterAll, expect, it } from "vitest"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import pg from "pg"
import { startPostgres, stopPostgres, makePool } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  getOrder,
  listOrders,
  createOrder,
  createOrderLine,
  listOrderLines,
  updateOrderTotal,
} from "../../src/ordersQueries.js"
import {
  getProduct,
} from "../../src/productsQueries.js"

// Transaction helper - equivalent to Effect's sql.withTransaction
async function withTransaction<T>(
  pool: pg.Pool,
  fn: (client: pg.PoolClient) => Promise<T>
): Promise<T> {
  const client = await pool.connect()
  try {
    await client.query("BEGIN")
    const result = await fn(client)
    await client.query("COMMIT")
    return result
  } catch (e) {
    await client.query("ROLLBACK")
    throw e
  } finally {
    client.release()
  }
}

// Savepoint helper - equivalent to Effect's nested sql.withTransaction (which uses savepoints)
async function withSavepoint<T>(
  client: pg.PoolClient,
  name: string,
  fn: () => Promise<T>
): Promise<T> {
  await client.query(`SAVEPOINT ${name}`)
  try {
    const result = await fn()
    await client.query(`RELEASE SAVEPOINT ${name}`)
    return result
  } catch (e) {
    await client.query(`ROLLBACK TO SAVEPOINT ${name}`)
    throw e
  }
}

// Custom error types for transaction scenarios - mirrors TransactionRepository.test.ts
class OrderLineCreationError extends Error {
  readonly _tag = "OrderLineCreationError"
  constructor(
    readonly message: string,
    readonly orderId: number
  ) {
    super(message)
  }
}

class InsufficientStockError extends Error {
  readonly _tag = "InsufficientStockError"
  constructor(
    readonly productId: number,
    readonly requested: number,
    readonly available: number
  ) {
    super(`Insufficient stock for product ${productId}: requested ${requested}, available ${available}`)
  }
}

describe("Transaction Tests", () => {
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

  describe("Happy Path - Successful Transaction", () => {
    it("creates an order with two order lines within a transaction", async () => {
      const result = await withTransaction(pool, async (client) => {
        // Create the order
        const orderResult = await createOrder(client, {
          customerId: 1,
          status: "pending",
          shippingAddress: "123 Transaction St, Test City, TC 12345",
          billingAddress: "123 Transaction St, Test City, TC 12345",
          notes: "Transaction test order",
        })
        if (!orderResult.success || !orderResult.data) throw new Error("order creation failed")
        const order = orderResult.data

        // Create first order line
        const line1Result = await createOrderLine(client, {
          orderId: order.id,
          productId: 1, // Wireless Headphones
          quantity: 2,
          unitPriceCents: 14999,
          discountCents: 0,
        })
        if (!line1Result.success || !line1Result.data) throw new Error("line 1 creation failed")
        const line1 = line1Result.data

        // Create second order line
        const line2Result = await createOrderLine(client, {
          orderId: order.id,
          productId: 2, // USB-C Hub
          quantity: 1,
          unitPriceCents: 4999,
          discountCents: 500,
        })
        if (!line2Result.success || !line2Result.data) throw new Error("line 2 creation failed")
        const line2 = line2Result.data

        // Update order total
        await updateOrderTotal(client, { orderId: order.id })

        // Fetch the updated order
        const updatedResult = await getOrder(client, { id: order.id })
        if (!updatedResult.success || !updatedResult.data) throw new Error("order fetch failed")
        const updatedOrder = updatedResult.data

        return { order: updatedOrder, lines: [line1, line2] }
      })

      // Verify the transaction committed successfully
      expect(result.order.id).toBeGreaterThan(0)
      expect(result.order.customer_id).toBe(1)
      expect(result.order.status).toBe("pending")
      expect(result.order.shipping_address).toBe("123 Transaction St, Test City, TC 12345")

      // Verify order lines were created
      expect(result.lines).toHaveLength(2)
      expect(result.lines[0]!.product_id).toBe(1)
      expect(result.lines[0]!.quantity).toBe(2)
      expect(result.lines[1]!.product_id).toBe(2)
      expect(result.lines[1]!.quantity).toBe(1)

      // Verify order total was calculated correctly
      // Line 1: 2 * 14999 - 0 = 29998
      // Line 2: 1 * 4999 - 500 = 4499
      // Total: 34497
      expect(result.order.total_cents).toBe(34497)

      // Verify order and lines persist after transaction
      const verifyOrder = await getOrder(pool, { id: result.order.id })
      expect(verifyOrder.success).toBe(true)
      if (!verifyOrder.success) throw new Error("expected success")
      expect(verifyOrder.data).not.toBeNull()

      const verifyLines = await listOrderLines(pool, { orderId: result.order.id })
      expect(verifyLines.success).toBe(true)
      if (!verifyLines.success) throw new Error("expected success")
      expect(verifyLines.data).toHaveLength(2)
    })
  })

  describe("Rollback Scenario - Transaction Failure", () => {
    it("rolls back the entire transaction when an error occurs after creating the order", async () => {
      // Get initial order count to verify rollback
      const initialOrdersResult = await listOrders(pool)
      if (!initialOrdersResult.success) throw new Error("expected success")
      const initialOrderCount = initialOrdersResult.data.length

      // Track the order ID that was created (for verification)
      let createdOrderId: number | undefined
      let rolledBack = false
      let caughtError: OrderLineCreationError | undefined

      try {
        await withTransaction(pool, async (client) => {
          // Create the order - this succeeds
          const orderResult = await createOrder(client, {
            customerId: 2,
            status: "pending",
            shippingAddress: "456 Rollback Ave, Test City, TC 12345",
            notes: "This order should be rolled back",
          })
          if (!orderResult.success || !orderResult.data) throw new Error("order creation failed")
          createdOrderId = orderResult.data.id

          // Create first order line - this succeeds
          await createOrderLine(client, {
            orderId: orderResult.data.id,
            productId: 1,
            quantity: 1,
            unitPriceCents: 14999,
            discountCents: 0,
          })

          // Simulate a business logic failure before completing
          throw new OrderLineCreationError(
            "Simulated failure during order line creation",
            orderResult.data.id
          )
        })
      } catch (e) {
        if (e instanceof OrderLineCreationError) {
          rolledBack = true
          caughtError = e
        } else {
          throw e
        }
      }

      // Verify the transaction was rolled back
      expect(rolledBack).toBe(true)
      expect(caughtError!._tag).toBe("OrderLineCreationError")

      // Verify the order does NOT exist (was rolled back)
      if (createdOrderId !== undefined) {
        const orderCheck = await getOrder(pool, { id: createdOrderId })
        expect(orderCheck.success).toBe(true)
        if (!orderCheck.success) throw new Error("expected success")
        expect(orderCheck.data).toBeNull()
      }

      // Verify order count is unchanged
      const finalOrdersResult = await listOrders(pool)
      if (!finalOrdersResult.success) throw new Error("expected success")
      expect(finalOrdersResult.data.length).toBe(initialOrderCount)
    })

    it("rolls back when a database constraint is violated", async () => {
      // Get initial order count
      const initialOrdersResult = await listOrders(pool)
      if (!initialOrdersResult.success) throw new Error("expected success")
      const initialOrderCount = initialOrdersResult.data.length

      let createdOrderId: number | undefined
      let transactionRolledBack = false

      try {
        await withTransaction(pool, async (client) => {
          // Create a valid order
          const orderResult = await createOrder(client, {
            customerId: 1,
            status: "pending",
          })
          if (!orderResult.success || !orderResult.data) throw new Error("order creation failed")
          createdOrderId = orderResult.data.id

          // Create a valid order line
          await createOrderLine(client, {
            orderId: orderResult.data.id,
            productId: 1,
            quantity: 1,
            unitPriceCents: 14999,
            discountCents: 0,
          })

          // Try to create a duplicate order line (same order + product)
          // This violates the unique constraint: order_lines_order_product_idx
          await createOrderLine(client, {
            orderId: orderResult.data.id,
            productId: 1, // Same product - will violate unique constraint
            quantity: 2,
            unitPriceCents: 14999,
            discountCents: 0,
          })
        })
      } catch (e) {
        // Expected - constraint violation caused rollback
        transactionRolledBack = true
      }

      // Verify the transaction was rolled back due to constraint violation
      expect(transactionRolledBack).toBe(true)

      // Verify the order does NOT exist (was rolled back)
      if (createdOrderId !== undefined) {
        const orderCheck = await getOrder(pool, { id: createdOrderId })
        expect(orderCheck.success).toBe(true)
        if (!orderCheck.success) throw new Error("expected success")
        expect(orderCheck.data).toBeNull()
      }

      // Verify order count is unchanged
      const finalOrdersResult = await listOrders(pool)
      if (!finalOrdersResult.success) throw new Error("expected success")
      expect(finalOrdersResult.data.length).toBe(initialOrderCount)
    })
  })

  describe("Failure Scenario - Error Handling Patterns", () => {
    it("demonstrates typed error handling", async () => {
      // A more realistic scenario: check stock before creating order lines
      const createOrderWithStockCheck = async (requestedQuantity: number) => {
        return withTransaction(pool, async (client) => {
          // Check product stock first
          const productResult = await getProduct(client, { id: 1 })
          if (!productResult.success || !productResult.data) throw new Error("product not found")
          const product = productResult.data

          if (product.stock_quantity < requestedQuantity) {
            throw new InsufficientStockError(
              product.id,
              requestedQuantity,
              product.stock_quantity
            )
          }

          // Create the order
          const orderResult = await createOrder(client, {
            customerId: 3,
            status: "pending",
            notes: "Order with stock check",
          })
          if (!orderResult.success || !orderResult.data) throw new Error("order creation failed")
          const order = orderResult.data

          // Create the order line
          const lineResult = await createOrderLine(client, {
            orderId: order.id,
            productId: 1,
            quantity: requestedQuantity,
            unitPriceCents: product.price_cents,
            discountCents: 0,
          })
          if (!lineResult.success || !lineResult.data) throw new Error("line creation failed")

          return { order, line: lineResult.data }
        })
      }

      // Test case 1: Request more stock than available (should fail with InsufficientStockError)
      let failurePassed = false
      let requestedQty = 0
      let availableQty = 0

      try {
        await createOrderWithStockCheck(99999)
      } catch (e) {
        if (e instanceof InsufficientStockError) {
          failurePassed = true
          requestedQty = e.requested
          availableQty = e.available
        }
      }

      expect(failurePassed).toBe(true)
      expect(requestedQty).toBe(99999)
      expect(availableQty).toBeLessThan(99999)

      // Test case 2: Request valid quantity (should succeed)
      const successResult = await createOrderWithStockCheck(1)
      expect(successResult.order.id).toBeGreaterThan(0)

      // Verify the order was created
      const orderCheck = await getOrder(pool, { id: successResult.order.id })
      expect(orderCheck.success).toBe(true)
      if (!orderCheck.success) throw new Error("expected success")
      expect(orderCheck.data).not.toBeNull()
    })

    it("demonstrates recovery pattern with separate transactions", async () => {
      // IMPORTANT: PostgreSQL aborts the entire transaction when an error occurs.
      // Recovery must happen in a NEW transaction, not within the failed one.
      // This test demonstrates the proper pattern for fallback logic.

      let result: { _tag: "Primary" | "Fallback"; order: { customer_id: number; status: string; notes: string | null } }

      try {
        // Attempt to create an order for a non-existent customer (will fail FK constraint)
        const order = await withTransaction(pool, async (client) => {
          const orderResult = await createOrder(client, {
            customerId: 99999, // Non-existent customer - will fail
            status: "confirmed",
          })
          if (!orderResult.success || !orderResult.data) throw new Error("order creation failed")
          return orderResult.data
        })
        result = { _tag: "Primary", order }
      } catch {
        // Fallback: Create order for existing customer in a SEPARATE transaction
        const fallbackOrder = await withTransaction(pool, async (client) => {
          const orderResult = await createOrder(client, {
            customerId: 1, // Valid customer
            status: "pending",
            notes: "Created via fallback due to original order failure",
          })
          if (!orderResult.success || !orderResult.data) throw new Error("fallback order creation failed")
          return orderResult.data
        })
        result = { _tag: "Fallback", order: fallbackOrder }
      }

      // Verify the fallback was used
      expect(result._tag).toBe("Fallback")
      expect(result.order.customer_id).toBe(1)
      expect(result.order.status).toBe("pending")
      expect(result.order.notes).toContain("fallback")
    })

    it("demonstrates nested transaction behavior (savepoints)", async () => {
      // Note: PostgreSQL supports savepoints for partial rollbacks within transactions.
      // This test demonstrates using savepoints so that a failed nested operation
      // does not roll back the entire outer transaction.

      const result = await withTransaction(pool, async (client) => {
        // Create the main order (outer transaction)
        const orderResult = await createOrder(client, {
          customerId: 1,
          status: "pending",
          notes: "Main order",
        })
        if (!orderResult.success || !orderResult.data) throw new Error("order creation failed")
        const order = orderResult.data

        // Create first order line (succeeds)
        await createOrderLine(client, {
          orderId: order.id,
          productId: 1,
          quantity: 1,
          unitPriceCents: 14999,
          discountCents: 0,
        })

        // Try a nested savepoint that fails, but catch and continue
        await withSavepoint(client, "sp_duplicate", async () => {
          // Try to add a duplicate product line (will fail)
          await createOrderLine(client, {
            orderId: order.id,
            productId: 1, // Duplicate - violates unique constraint
            quantity: 2,
            unitPriceCents: 14999,
            discountCents: 0,
          })
        }).catch(() => {
          // Savepoint rolled back, but outer transaction continues
        })

        // Add a different product (should succeed)
        await createOrderLine(client, {
          orderId: order.id,
          productId: 2, // Different product
          quantity: 1,
          unitPriceCents: 4999,
          discountCents: 0,
        })

        // Update total
        await updateOrderTotal(client, { orderId: order.id })

        return order
      })

      // Verify the main transaction completed
      const verifyOrder = await getOrder(pool, { id: result.id })
      expect(verifyOrder.success).toBe(true)
      if (!verifyOrder.success) throw new Error("expected success")
      expect(verifyOrder.data).not.toBeNull()

      // Verify order lines: should have product 1 and product 2 (duplicate was rolled back to savepoint)
      const lines = await listOrderLines(pool, { orderId: result.id })
      expect(lines.success).toBe(true)
      if (!lines.success) throw new Error("expected success")
      expect(lines.data).toHaveLength(2)
      expect(lines.data.map(l => l.product_id).sort()).toEqual([1, 2])
    })
  })
})
