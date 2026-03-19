import { describe, beforeAll, afterAll, expect, it } from "@effect/vitest"
import { Effect, Layer, Option, Data, Match } from "effect"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import { SqlClient, SqlError } from "effect/unstable/sql"
import { startPostgres, stopPostgres, makeTestLayer } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  OrdersRepository,
  ordersRepositoryLive,
  type CreateOrderParams,
  type CreateOrderLineParams,
} from "../../src/repositories/OrdersRepository.js"
import {
  ProductsRepository,
  productsRepositoryLive,
} from "../../src/repositories/ProductsRepository.js"

// Custom error types for transaction scenarios
class OrderLineCreationError extends Data.TaggedError("OrderLineCreationError")<{
  readonly message: string
  readonly orderId: number
}> {}

class InsufficientStockError extends Data.TaggedError("InsufficientStockError")<{
  readonly productId: number
  readonly requested: number
  readonly available: number
}> {}

describe("Transaction Tests", () => {
  let container: StartedPostgreSqlContainer
  let testLayer: Layer.Layer<
    OrdersRepository | ProductsRepository | SqlClient.SqlClient,
    SqlError.SqlError
  >

  beforeAll(async () => {
    container = await startPostgres()
    const sqlLayer = makeTestLayer(container)
    testLayer = Layer.mergeAll(
      ordersRepositoryLive,
      productsRepositoryLive,
      sqlLayer
    ).pipe(Layer.provide(sqlLayer))

    // Seed the database
    await Effect.runPromise(seedDatabase.pipe(Effect.provide(sqlLayer)))
  }, 120000)

  afterAll(async () => {
    await stopPostgres()
  })

  describe("Happy Path - Successful Transaction", () => {
    it.effect("creates an order with two order lines within a transaction", () =>
      Effect.gen(function* () {
        const ordersRepo = yield* OrdersRepository
        const sql = yield* SqlClient.SqlClient

        // Define order and order lines data
        const orderParams: CreateOrderParams = {
          customerId: 1,
          status: "pending",
          shippingAddress: "123 Transaction St, Test City, TC 12345",
          billingAddress: "123 Transaction St, Test City, TC 12345",
          notes: "Transaction test order",
        }

        const orderLineParams1: Omit<CreateOrderLineParams, "orderId"> = {
          productId: 1, // Wireless Headphones
          quantity: 2,
          unitPriceCents: 14999,
          discountCents: 0,
        }

        const orderLineParams2: Omit<CreateOrderLineParams, "orderId"> = {
          productId: 2, // USB-C Hub
          quantity: 1,
          unitPriceCents: 4999,
          discountCents: 500,
        }

        // Execute within a transaction
        const result = yield* sql.withTransaction(
          Effect.gen(function* () {
            // Create the order
            const orderOption = yield* ordersRepo.createOrder(orderParams)
            const order = Option.getOrThrow(orderOption)

            // Create first order line
            const line1Option = yield* ordersRepo.createOrderLine({
              ...orderLineParams1,
              orderId: order.id,
            })
            const line1 = Option.getOrThrow(line1Option)

            // Create second order line
            const line2Option = yield* ordersRepo.createOrderLine({
              ...orderLineParams2,
              orderId: order.id,
            })
            const line2 = Option.getOrThrow(line2Option)

            // Update order total
            yield* ordersRepo.updateOrderTotal({ orderId: order.id })

            // Fetch the updated order
            const updatedOrderOption = yield* ordersRepo.getOrder({ id: order.id })
            const updatedOrder = Option.getOrThrow(updatedOrderOption)

            return { order: updatedOrder, lines: [line1, line2] }
          })
        )

        // Verify the transaction committed successfully
        expect(result.order.id).toBeGreaterThan(0)
        expect(result.order.customer_id).toBe(1)
        expect(result.order.status).toBe("pending")
        expect(Option.getOrNull(result.order.shipping_address)).toBe(
          "123 Transaction St, Test City, TC 12345"
        )

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
        const verifyOrder = yield* ordersRepo.getOrder({ id: result.order.id })
        expect(Option.isSome(verifyOrder)).toBe(true)

        const verifyLines = yield* ordersRepo.listOrderLines({
          orderId: result.order.id,
        })
        expect(verifyLines).toHaveLength(2)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("Rollback Scenario - Transaction Failure", () => {
    it.effect("rolls back the entire transaction when an error occurs after creating the order", () =>
      Effect.gen(function* () {
        const ordersRepo = yield* OrdersRepository
        const sql = yield* SqlClient.SqlClient

        // Get initial order count to verify rollback
        const initialOrders = yield* ordersRepo.listOrders()
        const initialOrderCount = initialOrders.length

        const orderParams: CreateOrderParams = {
          customerId: 2,
          status: "pending",
          shippingAddress: "456 Rollback Ave, Test City, TC 12345",
          notes: "This order should be rolled back",
        }

        // Track the order ID that was created (for verification)
        let createdOrderId: number | undefined

        // Execute within a transaction that will fail
        const transactionResult = yield* sql
          .withTransaction(
            Effect.gen(function* () {
              // Create the order - this succeeds
              const orderOption = yield* ordersRepo.createOrder(orderParams)
              const order = Option.getOrThrow(orderOption)
              createdOrderId = order.id

              // Create first order line - this succeeds
              yield* ordersRepo.createOrderLine({
                orderId: order.id,
                productId: 1,
                quantity: 1,
                unitPriceCents: 14999,
                discountCents: 0,
              })

              // Simulate a business logic failure before completing
              // (e.g., inventory check fails, payment processing fails)
              return yield* Effect.fail(
                new OrderLineCreationError({
                  message: "Simulated failure during order line creation",
                  orderId: order.id,
                })
              )
            })
          )
          .pipe(
            Effect.catchTag("OrderLineCreationError", (error) =>
              Effect.succeed({ rolledBack: true, error })
            )
          )

        // Verify the transaction was rolled back
        expect(transactionResult.rolledBack).toBe(true)
        expect(transactionResult.error._tag).toBe("OrderLineCreationError")

        // Verify the order does NOT exist (was rolled back)
        if (createdOrderId !== undefined) {
          const orderCheck = yield* ordersRepo.getOrder({ id: createdOrderId })
          expect(Option.isNone(orderCheck)).toBe(true)
        }

        // Verify order count is unchanged
        const finalOrders = yield* ordersRepo.listOrders()
        expect(finalOrders.length).toBe(initialOrderCount)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("rolls back when a database constraint is violated", () =>
      Effect.gen(function* () {
        const ordersRepo = yield* OrdersRepository
        const sql = yield* SqlClient.SqlClient

        // Get initial order count
        const initialOrders = yield* ordersRepo.listOrders()
        const initialOrderCount = initialOrders.length

        let createdOrderId: number | undefined

        // Execute within a transaction that will fail due to constraint violation
        const transactionResult = yield* sql
          .withTransaction(
            Effect.gen(function* () {
              // Create a valid order
              const orderOption = yield* ordersRepo.createOrder({
                customerId: 1,
                status: "pending",
              })
              const order = Option.getOrThrow(orderOption)
              createdOrderId = order.id

              // Create a valid order line
              yield* ordersRepo.createOrderLine({
                orderId: order.id,
                productId: 1,
                quantity: 1,
                unitPriceCents: 14999,
                discountCents: 0,
              })

              // Try to create a duplicate order line (same order + product)
              // This violates the unique constraint: order_lines_order_product_idx
              yield* ordersRepo.createOrderLine({
                orderId: order.id,
                productId: 1, // Same product - will violate unique constraint
                quantity: 2,
                unitPriceCents: 14999,
                discountCents: 0,
              })

              return { _tag: "Success", order } as const
            })
          )
          .pipe(
            Effect.catchTag("SqlError", (error) =>
              Effect.succeed({ _tag: "RolledBack", error } as const)
            )
          )

        // Verify the transaction was rolled back due to constraint violation using Match
        Match.value(transactionResult).pipe(
          Match.when({ _tag: "RolledBack" }, () => {
            // Expected - constraint violation caused rollback
          }),
          Match.when({ _tag: "Success" }, () => {
            expect.fail("Expected transaction to be rolled back due to constraint violation")
          }),
          Match.exhaustive
        )

        // Verify the order does NOT exist (was rolled back)
        if (createdOrderId !== undefined) {
          const orderCheck = yield* ordersRepo.getOrder({ id: createdOrderId })
          expect(Option.isNone(orderCheck)).toBe(true)
        }

        // Verify order count is unchanged
        const finalOrders = yield* ordersRepo.listOrders()
        expect(finalOrders.length).toBe(initialOrderCount)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("Failure Scenario - Error Handling Patterns", () => {
    it.effect("demonstrates typed error handling with Effect.catchTag", () =>
      Effect.gen(function* () {
        const ordersRepo = yield* OrdersRepository
        const productsRepo = yield* ProductsRepository
        const sql = yield* SqlClient.SqlClient

        const orderParams: CreateOrderParams = {
          customerId: 3,
          status: "pending",
          notes: "Order with stock check",
        }

        // A more realistic scenario: check stock before creating order lines
        const createOrderWithStockCheck = (
          requestedQuantity: number
        ) =>
          sql.withTransaction(
            Effect.gen(function* () {
              // Check product stock first
              const productOption = yield* productsRepo.getProduct({ id: 1 })
              const product = Option.getOrThrow(productOption)

              if (product.stock_quantity < requestedQuantity) {
                return yield* Effect.fail(
                  new InsufficientStockError({
                    productId: product.id,
                    requested: requestedQuantity,
                    available: product.stock_quantity,
                  })
                )
              }

              // Create the order
              const orderOption = yield* ordersRepo.createOrder(orderParams)
              const order = Option.getOrThrow(orderOption)

              // Create the order line
              const lineOption = yield* ordersRepo.createOrderLine({
                orderId: order.id,
                productId: 1,
                quantity: requestedQuantity,
                unitPriceCents: product.price_cents,
                discountCents: 0,
              })

              return { order, line: Option.getOrThrow(lineOption) }
            })
          )

        // Test case 1: Request more stock than available (should fail with InsufficientStockError)
        const failureResult = yield* createOrderWithStockCheck(99999).pipe(
          Effect.match({
            onFailure: (error) => ({ _tag: "Failure", error }) as const,
            onSuccess: (result) => ({ _tag: "Success", result }) as const,
          })
        )

        const failureChecks = Match.value(failureResult).pipe(
          Match.when({ _tag: "Failure" }, ({ error }) =>
            Match.value(error).pipe(
              Match.tag("InsufficientStockError", (e) => ({
                passed: true,
                requested: e.requested,
                available: e.available,
              })),
              Match.orElse(() => ({ passed: false, requested: 0, available: 0 }))
            )
          ),
          Match.orElse(() => ({ passed: false, requested: 0, available: 0 }))
        )

        expect(failureChecks.passed).toBe(true)
        expect(failureChecks.requested).toBe(99999)
        expect(failureChecks.available).toBeLessThan(99999)

        // Test case 2: Request valid quantity (should succeed)
        const successResult = yield* createOrderWithStockCheck(1).pipe(
          Effect.match({
            onFailure: (error) => ({ _tag: "Failure", error }) as const,
            onSuccess: (result) => ({ _tag: "Success", result }) as const,
          })
        )

        // Use Match to handle the result and extract the order ID for verification
        const orderId = Match.value(successResult).pipe(
          Match.when({ _tag: "Success" }, ({ result }) => Option.some(result.order.id)),
          Match.orElse(() => Option.none())
        )

        expect(Option.isSome(orderId)).toBe(true)
        if (Option.isSome(orderId)) {
          // Verify the order was created
          const orderCheck = yield* ordersRepo.getOrder({ id: orderId.value })
          expect(Option.isSome(orderCheck)).toBe(true)
        }
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("demonstrates recovery pattern with separate transactions", () =>
      Effect.gen(function* () {
        const ordersRepo = yield* OrdersRepository
        const sql = yield* SqlClient.SqlClient

        // IMPORTANT: PostgreSQL aborts the entire transaction when an error occurs.
        // Recovery must happen in a NEW transaction, not within the failed one.
        // This test demonstrates the proper pattern for fallback logic.

        // Attempt to create an order for a non-existent customer (will fail FK constraint)
        const primaryAttempt = sql.withTransaction(
          Effect.gen(function* () {
            const orderOption = yield* ordersRepo.createOrder({
              customerId: 99999, // Non-existent customer - will fail
              status: "confirmed",
            })
            return Option.isSome(orderOption)
              ? { _tag: "Primary", order: orderOption.value } as const
              : yield* Effect.fail(new Error("Order creation returned None"))
          })
        )

        // Fallback: Create order for existing customer in a SEPARATE transaction
        const fallbackAttempt = sql.withTransaction(
          Effect.gen(function* () {
            const fallbackOrder = yield* ordersRepo.createOrder({
              customerId: 1, // Valid customer
              status: "pending",
              notes: "Created via fallback due to original order failure",
            })
            return {
              _tag: "Fallback",
              order: Option.getOrThrow(fallbackOrder),
            } as const
          })
        )

        // Try primary, fall back to secondary on SqlError
        const result = yield* primaryAttempt.pipe(
          Effect.catchTag("SqlError", () => fallbackAttempt)
        )

        // Use Match to verify the fallback was used
        Match.value(result).pipe(
          Match.when({ _tag: "Fallback" }, ({ order }) => {
            expect(order.customer_id).toBe(1)
            expect(order.status).toBe("pending")
            expect(Option.getOrNull(order.notes)).toContain("fallback")
          }),
          Match.when({ _tag: "Primary" }, () => {
            // This should not happen - primary should fail
            expect.fail("Expected fallback to be used, but primary succeeded")
          }),
          Match.exhaustive
        )
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("demonstrates nested transaction behavior (savepoints)", () =>
      Effect.gen(function* () {
        const ordersRepo = yield* OrdersRepository
        const sql = yield* SqlClient.SqlClient

        // Note: Effect SQL transactions with withTransaction create savepoints
        // when nested, allowing partial rollbacks
        const result = yield* sql.withTransaction(
          Effect.gen(function* () {
            // Create the main order (outer transaction)
            const orderOption = yield* ordersRepo.createOrder({
              customerId: 1,
              status: "pending",
              notes: "Main order",
            })
            const order = Option.getOrThrow(orderOption)

            // Create first order line (succeeds)
            yield* ordersRepo.createOrderLine({
              orderId: order.id,
              productId: 1,
              quantity: 1,
              unitPriceCents: 14999,
              discountCents: 0,
            })

            // Try a nested transaction that fails, but catch and continue
            yield* sql
              .withTransaction(
                Effect.gen(function* () {
                  // Try to add a duplicate product line (will fail)
                  yield* ordersRepo.createOrderLine({
                    orderId: order.id,
                    productId: 1, // Duplicate - violates unique constraint
                    quantity: 2,
                    unitPriceCents: 14999,
                    discountCents: 0,
                  })
                })
              )
              .pipe(
                Effect.catchTag("SqlError", () =>
                  // Nested transaction failed, but we can continue
                  Effect.void
                )
              )

            // Add a different product (should succeed)
            yield* ordersRepo.createOrderLine({
              orderId: order.id,
              productId: 2, // Different product
              quantity: 1,
              unitPriceCents: 4999,
              discountCents: 0,
            })

            // Update total
            yield* ordersRepo.updateOrderTotal({ orderId: order.id })

            return order
          })
        )

        // Verify the main transaction completed
        const verifyOrder = yield* ordersRepo.getOrder({ id: result.id })
        expect(Option.isSome(verifyOrder)).toBe(true)

        // Verify order lines
        const lines = yield* ordersRepo.listOrderLines({ orderId: result.id })
        expect(lines).toHaveLength(2)
        expect(lines.map((l) => l.product_id).sort()).toEqual([1, 2])
      }).pipe(Effect.provide(testLayer))
    )
  })
})
