import { describe, beforeAll, afterAll, expect, it } from "vitest"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import pg from "pg"
import { startPostgres, stopPostgres, makePool } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  getOrder,
  getOrderWithCustomer,
  listOrders,
  listOrdersByCustomer,
  listOrdersByStatus,
  listOrdersPaginated,
  listOrdersByDateRange,
  listOrdersByDateRangeNamed,
  listRecentOrdersWithCustomer,
  searchOrders,
  createOrder,
  updateOrderStatus,
  updateOrderAddresses,
  updateOrderTotal,
  deleteOrder,
  countOrdersByStatus,
  getCustomerOrderStats,
  getOrdersWithLineCount,
  getOrderLine,
  listOrderLines,
  listOrderLinesWithProduct,
  getFullOrderDetails,
  createOrderLine,
  updateOrderLineQuantity,
  deleteOrderLine,
  deleteOrderLines,
  getOrderLineTotal,
  getProductSalesStats,
  getTopSellingProducts,
  getOrdersByProductIds,
} from "../../src/ordersQueries.js"

describe("OrdersQueries", () => {
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

  // ==================== Order Queries ====================

  describe("getOrder", () => {
    it("returns order when found", async () => {
      const result = await getOrder(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })

    it("returns null when order not found", async () => {
      const result = await getOrder(pool, { id: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("getOrderWithCustomer", () => {
    it("returns order with customer details", async () => {
      const result = await getOrderWithCustomer(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.customer_name).toBe("Alice Johnson")
      expect(result.data!.customer_email).toBe("alice@example.com")
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listOrders", () => {
    it("returns all orders ordered by created_at DESC", async () => {
      const result = await listOrders(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(15)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listOrdersByCustomer", () => {
    it("returns orders for a specific customer", async () => {
      const result = await listOrdersByCustomer(pool, { customerId: 1 })

      // Customer 1 (Alice) has 2 orders
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(2)
      expect(result.data.every(o => o.customer_id === 1)).toBe(true)
      expect(result.data).toMatchSnapshot()
    })

    it("returns empty array for customer with no orders", async () => {
      const result = await listOrdersByCustomer(pool, { customerId: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })

  describe("listOrdersByStatus", () => {
    it("returns orders with pending status", async () => {
      const result = await listOrdersByStatus(pool, { status: "pending" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(3)
      expect(result.data.every(o => o.status === "pending")).toBe(true)
      expect(result.data).toMatchSnapshot()
    })

    it("returns orders with delivered status", async () => {
      const result = await listOrdersByStatus(pool, { status: "delivered" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(3)
      expect(result.data).toMatchSnapshot()
    })

    it("returns orders with cancelled status", async () => {
      const result = await listOrdersByStatus(pool, { status: "cancelled" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(1)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listOrdersPaginated", () => {
    it("returns paginated orders - first page", async () => {
      const result = await listOrdersPaginated(pool, { limit: 5, offset: 0 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(5)
      expect(result.data).toMatchSnapshot()
    })

    it("returns paginated orders - second page", async () => {
      const result = await listOrdersPaginated(pool, { limit: 5, offset: 5 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(5)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listOrdersByDateRange", () => {
    it("returns orders within date range", async () => {
      const result = await listOrdersByDateRange(pool, {
        createdAt: new Date("2024-01-20T00:00:00Z"),
        createdAt2: new Date("2024-01-25T00:00:00Z"),
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listOrdersByDateRangeNamed", () => {
    it("returns orders within named date range", async () => {
      const result = await listOrdersByDateRangeNamed(pool, {
        startDate: new Date("2024-01-15T00:00:00Z"),
        endDate: new Date("2024-01-20T00:00:00Z"),
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listRecentOrdersWithCustomer", () => {
    it("returns recent orders with customer info", async () => {
      const result = await listRecentOrdersWithCustomer(pool, { limit: 5 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(5)
      expect(result.data[0]!.customer_name).toBeDefined()
      expect(result.data[0]!.customer_email).toBeDefined()
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("searchOrders", () => {
    it("searches orders by shipping address", async () => {
      const result = await searchOrders(pool, { websearchToTsquery: "New York" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })

    it("searches orders by notes", async () => {
      const result = await searchOrders(pool, { websearchToTsquery: "gift" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("createOrder", () => {
    it("creates a new order with all fields", async () => {
      const result = await createOrder(pool, {
        customerId: 1,
        status: "pending",
        shippingAddress: "123 Test St, Test City, TC 12345",
        billingAddress: "123 Test St, Test City, TC 12345",
        notes: "Test order",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.customer_id).toBe(1)
      expect(result.data!.status).toBe("pending")
      expect(result.data!.total_cents).toBe(0)
      expect(result.data!.shipping_address).toBe("123 Test St, Test City, TC 12345")
    })

    it("creates a new order with minimal fields", async () => {
      const result = await createOrder(pool, {
        customerId: 2,
        status: "confirmed",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.shipping_address).toBeNull()
      expect(result.data!.billing_address).toBeNull()
      expect(result.data!.notes).toBeNull()
    })
  })

  describe("updateOrderStatus", () => {
    it("updates order status", async () => {
      // Update order 4 from pending to confirmed
      await updateOrderStatus(pool, { id: 4, status: "confirmed" })

      // Verify the change
      const order = await getOrder(pool, { id: 4 })
      expect(order.success).toBe(true)
      if (!order.success) throw new Error("expected success")
      expect(order.data!.status).toBe("confirmed")
    })
  })

  describe("updateOrderAddresses", () => {
    it("updates order addresses", async () => {
      await updateOrderAddresses(pool, {
        id: 1,
        shippingAddress: "New Shipping Address",
        billingAddress: "New Billing Address",
      })

      const order = await getOrder(pool, { id: 1 })
      expect(order.success).toBe(true)
      if (!order.success) throw new Error("expected success")
      expect(order.data!.shipping_address).toBe("New Shipping Address")
      expect(order.data!.billing_address).toBe("New Billing Address")
    })
  })

  describe("updateOrderTotal", () => {
    it("recalculates order total from order lines", async () => {
      // Order 1 has 2 order lines:
      // Line 1: 1 x 14999 - 0 = 14999
      // Line 2: 1 x 4999 - 0 = 4999
      // Total should be 19998

      await updateOrderTotal(pool, { orderId: 1 })

      const order = await getOrder(pool, { id: 1 })
      expect(order.success).toBe(true)
      if (!order.success) throw new Error("expected success")
      expect(order.data!.total_cents).toBe(19998)
    })
  })

  describe("deleteOrder", () => {
    it("deletes an order", async () => {
      // Create an order to delete
      const created = await createOrder(pool, {
        customerId: 1,
        status: "pending",
      })
      expect(created.success).toBe(true)
      if (!created.success) throw new Error("expected success")
      const orderId = created.data!.id

      // Delete it
      await deleteOrder(pool, { id: orderId })

      // Verify it's gone
      const result = await getOrder(pool, { id: orderId })
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")
      expect(result.data).toBeNull()
    })
  })

  describe("countOrdersByStatus", () => {
    it("returns order counts grouped by status", async () => {
      const result = await countOrdersByStatus(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("getCustomerOrderStats", () => {
    it("returns order statistics for a customer", async () => {
      const result = await getCustomerOrderStats(pool, { customerId: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(Number(result.data!.total_orders)).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("getOrdersWithLineCount", () => {
    it("returns orders with line item counts", async () => {
      const result = await getOrdersWithLineCount(pool, { limit: 5, offset: 0 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(5)
      expect(result.data[0]!.line_count).toBeDefined()
      // Validate structure without volatile timestamps
      expect(result.data.every(o => o.id && o.status && o.line_count !== undefined)).toBe(true)
    })
  })

  // ==================== Order Line Queries ====================

  describe("getOrderLine", () => {
    it("returns order line when found", async () => {
      const result = await getOrderLine(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })

    it("returns null when order line not found", async () => {
      const result = await getOrderLine(pool, { id: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("listOrderLines", () => {
    it("returns order lines for an order", async () => {
      const result = await listOrderLines(pool, { orderId: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(2)
      expect(result.data).toMatchSnapshot()
    })

    it("returns order lines for order with many items", async () => {
      const result = await listOrderLines(pool, { orderId: 15 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(6)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listOrderLinesWithProduct", () => {
    it("returns order lines with product details", async () => {
      const result = await listOrderLinesWithProduct(pool, { orderId: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(2)
      expect(result.data[0]!.product_name).toBeDefined()
      expect(result.data[0]!.product_sku).toBeDefined()
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("getFullOrderDetails", () => {
    it("returns full order details with customer and products", async () => {
      const result = await getFullOrderDetails(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(2) // 2 order lines
      expect(result.data[0]!.customer_name).toBeDefined()
      expect(result.data[0]!.product_name).toBeDefined()
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("createOrderLine", () => {
    it("creates a new order line", async () => {
      // Create a new order first
      const order = await createOrder(pool, {
        customerId: 1,
        status: "pending",
      })
      expect(order.success).toBe(true)
      if (!order.success) throw new Error("expected success")
      const orderId = order.data!.id

      // Create order line
      const result = await createOrderLine(pool, {
        orderId,
        productId: 1,
        quantity: 2,
        unitPriceCents: 14999,
        discountCents: 500,
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.order_id).toBe(orderId)
      expect(result.data!.product_id).toBe(1)
      expect(result.data!.quantity).toBe(2)
      expect(result.data!.unit_price_cents).toBe(14999)
      expect(result.data!.discount_cents).toBe(500)
    })
  })

  describe("updateOrderLineQuantity", () => {
    it("updates order line quantity", async () => {
      await updateOrderLineQuantity(pool, { id: 1, quantity: 5 })

      const line = await getOrderLine(pool, { id: 1 })
      expect(line.success).toBe(true)
      if (!line.success) throw new Error("expected success")
      expect(line.data!.quantity).toBe(5)
    })
  })

  describe("deleteOrderLine", () => {
    it("deletes an order line", async () => {
      // Create order and line to delete
      const order = await createOrder(pool, {
        customerId: 1,
        status: "pending",
      })
      expect(order.success).toBe(true)
      if (!order.success) throw new Error("expected success")
      const orderId = order.data!.id

      const line = await createOrderLine(pool, {
        orderId,
        productId: 2,
        quantity: 1,
        unitPriceCents: 4999,
        discountCents: 0,
      })
      expect(line.success).toBe(true)
      if (!line.success) throw new Error("expected success")
      const lineId = line.data!.id

      // Delete it
      await deleteOrderLine(pool, { id: lineId })

      // Verify it's gone
      const result = await getOrderLine(pool, { id: lineId })
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")
      expect(result.data).toBeNull()
    })
  })

  describe("deleteOrderLines", () => {
    it("deletes all order lines for an order", async () => {
      // Create order with multiple lines
      const order = await createOrder(pool, {
        customerId: 1,
        status: "pending",
      })
      expect(order.success).toBe(true)
      if (!order.success) throw new Error("expected success")
      const orderId = order.data!.id

      await createOrderLine(pool, {
        orderId,
        productId: 1,
        quantity: 1,
        unitPriceCents: 100,
        discountCents: 0,
      })
      await createOrderLine(pool, {
        orderId,
        productId: 2,
        quantity: 1,
        unitPriceCents: 200,
        discountCents: 0,
      })

      // Verify lines exist
      const beforeLines = await listOrderLines(pool, { orderId })
      expect(beforeLines.success).toBe(true)
      if (!beforeLines.success) throw new Error("expected success")
      expect(beforeLines.data.length).toBe(2)

      // Delete all lines
      const deleted = await deleteOrderLines(pool, { orderId })
      expect(deleted.success).toBe(true)
      if (!deleted.success) throw new Error("expected success")
      expect(deleted.data).toBe(2)

      // Verify lines are gone
      const afterLines = await listOrderLines(pool, { orderId })
      expect(afterLines.success).toBe(true)
      if (!afterLines.success) throw new Error("expected success")
      expect(afterLines.data.length).toBe(0)
    })
  })

  describe("getOrderLineTotal", () => {
    it("returns total for order lines", async () => {
      const result = await getOrderLineTotal(pool, { orderId: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })
  })

  // ==================== Product Sales Stats ====================

  describe("getProductSalesStats", () => {
    it("returns sales statistics for a product", async () => {
      const result = await getProductSalesStats(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.sku).toBe("ELEC-001")
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("getTopSellingProducts", () => {
    it("returns top selling products", async () => {
      const result = await getTopSellingProducts(pool, { limit: 5 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data.length).toBeLessThanOrEqual(5)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("getOrdersByProductIds", () => {
    it("returns orders containing specific products", async () => {
      const result = await getOrdersByProductIds(pool, { productIds: [1, 2] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      // Validate structure without volatile timestamps
      expect(result.data.every(o => o.id && o.status && typeof o.total_cents === 'number')).toBe(true)
    })

    it("returns empty array for products with no orders", async () => {
      // Products 5, 10, 20 are inactive and have no order lines
      const result = await getOrdersByProductIds(pool, { productIds: [5, 10, 20] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })
})
