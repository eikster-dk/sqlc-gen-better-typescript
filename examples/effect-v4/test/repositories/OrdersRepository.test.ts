import { describe, beforeAll, afterAll, expect, it } from "@effect/vitest"
import { Effect, Layer, Option } from "effect"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import { SqlError } from "effect/unstable/sql"
import { startPostgres, stopPostgres, makeTestLayer } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  OrdersRepository,
  ordersRepositoryLive,
} from "../../src/repositories/OrdersRepository.js"
import {
  CustomersRepository,
  customersRepositoryLive,
} from "../../src/repositories/CustomersRepository.js"
import {
  ProductsRepository,
  productsRepositoryLive,
} from "../../src/repositories/ProductsRepository.js"

describe("OrdersRepository", () => {
  let container: StartedPostgreSqlContainer
  let testLayer: Layer.Layer<OrdersRepository | CustomersRepository | ProductsRepository, SqlError.SqlError>

  beforeAll(async () => {
    container = await startPostgres()
    const sqlLayer = makeTestLayer(container)
    testLayer = Layer.mergeAll(
      ordersRepositoryLive,
      customersRepositoryLive,
      productsRepositoryLive
    ).pipe(Layer.provide(sqlLayer))

    // Seed the database
    await Effect.runPromise(
      seedDatabase.pipe(Effect.provide(sqlLayer))
    )
  }, 120000)

  afterAll(async () => {
    await stopPostgres()
  })

  // ==================== Order Queries ====================

  describe("getOrder", () => {
    it.effect("returns order when found", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrder({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        expect(Option.getOrNull(result)).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when order not found", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrder({ id: 9999 })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getOrderWithCustomer", () => {
    it.effect("returns order with customer details", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrderWithCustomer({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        const order = Option.getOrNull(result)!
        expect(order.customer_name).toBe("Alice Johnson")
        expect(order.customer_email).toBe("alice@example.com")
        expect(order).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrders", () => {
    it.effect("returns all orders ordered by created_at DESC", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrders()

        expect(result.length).toBe(15)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrdersByCustomer", () => {
    it.effect("returns orders for a specific customer", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersByCustomer({ customerId: 1 })

        // Customer 1 (Alice) has 2 orders
        expect(result.length).toBe(2)
        expect(result.every(o => o.customer_id === 1)).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array for customer with no orders", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        // Customer 10 (Jack) only has 1 order (id: 15)
        // Let's check a non-existent customer
        const result = yield* repo.listOrdersByCustomer({ customerId: 9999 })

        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrdersByStatus", () => {
    it.effect("returns orders with pending status", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersByStatus({ status: "pending" })

        expect(result.length).toBe(3)
        expect(result.every(o => o.status === "pending")).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns orders with delivered status", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersByStatus({ status: "delivered" })

        expect(result.length).toBe(3)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns orders with cancelled status", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersByStatus({ status: "cancelled" })

        expect(result.length).toBe(1)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrdersPaginated", () => {
    it.effect("returns paginated orders - first page", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersPaginated({ limit: 5, offset: 0 })

        expect(result.length).toBe(5)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns paginated orders - second page", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersPaginated({ limit: 5, offset: 5 })

        expect(result.length).toBe(5)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrdersByDateRange", () => {
    it.effect("returns orders within date range", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersByDateRange({
          createdAt: new Date("2024-01-20T00:00:00Z"),
          createdAt2: new Date("2024-01-25T00:00:00Z"),
        })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrdersByDateRangeNamed", () => {
    it.effect("returns orders within named date range", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrdersByDateRangeNamed({
          startDate: new Date("2024-01-15T00:00:00Z"),
          endDate: new Date("2024-01-20T00:00:00Z"),
        })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listRecentOrdersWithCustomer", () => {
    it.effect("returns recent orders with customer info", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listRecentOrdersWithCustomer({ limit: 5 })

        expect(result.length).toBe(5)
        expect(result[0]!.customer_name).toBeDefined()
        expect(result[0]!.customer_email).toBeDefined()
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("searchOrders", () => {
    it.effect("searches orders by shipping address", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.searchOrders({ websearchToTsquery: "New York" })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("searches orders by notes", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.searchOrders({ websearchToTsquery: "gift" })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("createOrder", () => {
    it.effect("creates a new order with all fields", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.createOrder({
          customerId: 1,
          status: "pending",
          shippingAddress: "123 Test St, Test City, TC 12345",
          billingAddress: "123 Test St, Test City, TC 12345",
          notes: "Test order",
        })

        expect(Option.isSome(result)).toBe(true)
        const order = Option.getOrNull(result)!
        expect(order.customer_id).toBe(1)
        expect(order.status).toBe("pending")
        expect(order.total_cents).toBe(0)
        expect(Option.getOrNull(order.shipping_address)).toBe("123 Test St, Test City, TC 12345")
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("creates a new order with minimal fields", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.createOrder({
          customerId: 2,
          status: "confirmed",
        })

        expect(Option.isSome(result)).toBe(true)
        const order = Option.getOrNull(result)!
        expect(Option.isNone(order.shipping_address)).toBe(true)
        expect(Option.isNone(order.billing_address)).toBe(true)
        expect(Option.isNone(order.notes)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateOrderStatus", () => {
    it.effect("updates order status", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        // Update order 4 from pending to confirmed
        yield* repo.updateOrderStatus({ id: 4, status: "confirmed" })

        // Verify the change
        const order = yield* repo.getOrder({ id: 4 })
        expect(Option.getOrNull(order)!.status).toBe("confirmed")
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateOrderAddresses", () => {
    it.effect("updates order addresses", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        yield* repo.updateOrderAddresses({
          id: 1,
          shippingAddress: "New Shipping Address",
          billingAddress: "New Billing Address",
        })

        const order = yield* repo.getOrder({ id: 1 })
        expect(Option.getOrNull(Option.getOrNull(order)!.shipping_address)).toBe("New Shipping Address")
        expect(Option.getOrNull(Option.getOrNull(order)!.billing_address)).toBe("New Billing Address")
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateOrderTotal", () => {
    it.effect("recalculates order total from order lines", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        // Order 1 has 2 order lines:
        // Line 1: 1 x 14999 - 0 = 14999
        // Line 2: 1 x 4999 - 0 = 4999
        // Total should be 19998

        yield* repo.updateOrderTotal({ orderId: 1 })

        const order = yield* repo.getOrder({ id: 1 })
        expect(Option.getOrNull(order)!.total_cents).toBe(19998)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("deleteOrder", () => {
    it.effect("deletes an order", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        // Create an order to delete
        const created = yield* repo.createOrder({
          customerId: 1,
          status: "pending",
        })
        const orderId = Option.getOrNull(created)!.id

        // Delete it
        yield* repo.deleteOrder({ id: orderId })

        // Verify it's gone
        const result = yield* repo.getOrder({ id: orderId })
        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("countOrdersByStatus", () => {
    it.effect("returns order counts grouped by status", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.countOrdersByStatus()

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getCustomerOrderStats", () => {
    it.effect("returns order statistics for a customer", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getCustomerOrderStats({ customerId: 1 })

        expect(Option.isSome(result)).toBe(true)
        const stats = Option.getOrNull(result)!
        expect(stats.total_orders).toBeGreaterThan(0n)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getOrdersWithLineCount", () => {
    it.effect("returns orders with line item counts", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrdersWithLineCount({ limit: 5, offset: 0 })

        expect(result.length).toBe(5)
        expect(result[0]!.line_count).toBeDefined()
        // Validate structure without volatile timestamps
        expect(result.every(o => o.id && o.status && typeof o.line_count === 'bigint')).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  // ==================== Order Line Queries ====================

  describe("getOrderLine", () => {
    it.effect("returns order line when found", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrderLine({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        expect(Option.getOrNull(result)).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when order line not found", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrderLine({ id: 9999 })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrderLines", () => {
    it.effect("returns order lines for an order", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrderLines({ orderId: 1 })

        expect(result.length).toBe(2)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns order lines for order with many items", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrderLines({ orderId: 15 })

        expect(result.length).toBe(6)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listOrderLinesWithProduct", () => {
    it.effect("returns order lines with product details", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.listOrderLinesWithProduct({ orderId: 1 })

        expect(result.length).toBe(2)
        expect(result[0]!.product_name).toBeDefined()
        expect(result[0]!.product_sku).toBeDefined()
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getFullOrderDetails", () => {
    it.effect("returns full order details with customer and products", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getFullOrderDetails({ id: 1 })

        expect(result.length).toBe(2) // 2 order lines
        expect(result[0]!.customer_name).toBeDefined()
        expect(result[0]!.product_name).toBeDefined()
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("createOrderLine", () => {
    it.effect("creates a new order line", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        // Create a new order first
        const order = yield* repo.createOrder({
          customerId: 1,
          status: "pending",
        })
        const orderId = Option.getOrNull(order)!.id

        // Create order line
        const result = yield* repo.createOrderLine({
          orderId,
          productId: 1,
          quantity: 2,
          unitPriceCents: 14999,
          discountCents: 500,
        })

        expect(Option.isSome(result)).toBe(true)
        const line = Option.getOrNull(result)!
        expect(line.order_id).toBe(orderId)
        expect(line.product_id).toBe(1)
        expect(line.quantity).toBe(2)
        expect(line.unit_price_cents).toBe(14999)
        expect(line.discount_cents).toBe(500)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateOrderLineQuantity", () => {
    it.effect("updates order line quantity", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        yield* repo.updateOrderLineQuantity({ id: 1, quantity: 5 })

        const line = yield* repo.getOrderLine({ id: 1 })
        expect(Option.getOrNull(line)!.quantity).toBe(5)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("deleteOrderLine", () => {
    it.effect("deletes an order line", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        // Create order and line to delete
        const order = yield* repo.createOrder({
          customerId: 1,
          status: "pending",
        })
        const orderId = Option.getOrNull(order)!.id

        const line = yield* repo.createOrderLine({
          orderId,
          productId: 2,
          quantity: 1,
          unitPriceCents: 4999,
          discountCents: 0,
        })
        const lineId = Option.getOrNull(line)!.id

        // Delete it
        yield* repo.deleteOrderLine({ id: lineId })

        // Verify it's gone
        const result = yield* repo.getOrderLine({ id: lineId })
        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("deleteOrderLines", () => {
    it.effect("deletes all order lines for an order", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository

        // Create order with multiple lines
        const order = yield* repo.createOrder({
          customerId: 1,
          status: "pending",
        })
        const orderId = Option.getOrNull(order)!.id

        yield* repo.createOrderLine({
          orderId,
          productId: 1,
          quantity: 1,
          unitPriceCents: 100,
          discountCents: 0,
        })
        yield* repo.createOrderLine({
          orderId,
          productId: 2,
          quantity: 1,
          unitPriceCents: 200,
          discountCents: 0,
        })

        // Verify lines exist
        const beforeLines = yield* repo.listOrderLines({ orderId })
        expect(beforeLines.length).toBe(2)

        // Delete all lines
        const deleted = yield* repo.deleteOrderLines({ orderId })
        expect(deleted).toBe(2)

        // Verify lines are gone
        const afterLines = yield* repo.listOrderLines({ orderId })
        expect(afterLines.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getOrderLineTotal", () => {
    it.effect("returns total for order lines", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrderLineTotal({ orderId: 1 })

        expect(Option.isSome(result)).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  // ==================== Product Sales Stats ====================

  describe("getProductSalesStats", () => {
    it.effect("returns sales statistics for a product", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getProductSalesStats({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        const stats = Option.getOrNull(result)!
        expect(stats.sku).toBe("ELEC-001")
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getTopSellingProducts", () => {
    it.effect("returns top selling products", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getTopSellingProducts({ limit: 5 })

        expect(result.length).toBeGreaterThan(0)
        expect(result.length).toBeLessThanOrEqual(5)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getOrdersByProductIds", () => {
    it.effect("returns orders containing specific products", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        const result = yield* repo.getOrdersByProductIds({ productIds: [1, 2] })

        expect(result.length).toBeGreaterThan(0)
        // Validate structure without volatile timestamps
        expect(result.every(o => o.id && o.status && typeof o.total_cents === 'number')).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array for products with no orders", () =>
      Effect.gen(function* () {
        const repo = yield* OrdersRepository
        // Product 5 (Old MP3 Player) and Product 20 (Old JavaScript Book) are inactive
        // and likely have no orders
        const result = yield* repo.getOrdersByProductIds({ productIds: [5, 10, 20] })

        // These inactive products have no order lines
        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })
})
