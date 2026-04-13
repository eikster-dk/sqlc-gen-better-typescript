import { describe, beforeAll, afterAll, expect, it } from "vitest"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import pg from "pg"
import { startPostgres, stopPostgres, makePool } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  getProduct,
  getProductBySku,
  listProducts,
  listActiveProducts,
  listProductsByCategory,
  listProductsPaginated,
  listProductsCursor,
  searchProducts,
  searchProductsRanked,
  searchProductsWithHighlight,
  searchProductsWebStyle,
  createProduct,
  updateProduct,
  updateProductStock,
  deactivateProduct,
  deleteProduct,
  getProductsByIds,
  countProductsByCategory,
  getLowStockProducts,
} from "../../src/productsQueries.js"

describe("ProductsQueries", () => {
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

  describe("getProduct", () => {
    it("returns product when found", async () => {
      const result = await getProduct(pool, { id: 1 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })

    it("returns null when product not found", async () => {
      const result = await getProduct(pool, { id: 9999 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("getProductBySku", () => {
    it("returns product when found by SKU", async () => {
      const result = await getProductBySku(pool, { sku: "ELEC-001" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data).toMatchSnapshot()
    })

    it("returns null when SKU not found", async () => {
      const result = await getProductBySku(pool, { sku: "NONEXISTENT-SKU" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("listProducts", () => {
    it("returns all products ordered by name", async () => {
      const result = await listProducts(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(20)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listActiveProducts", () => {
    it("returns only active products", async () => {
      const result = await listActiveProducts(pool)

      // 20 total - 3 inactive = 17 active
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(17)
      expect(result.data.every(p => p.is_active)).toBe(true)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listProductsByCategory", () => {
    it("returns products in Electronics category", async () => {
      const result = await listProductsByCategory(pool, { category: "Electronics" })

      // 5 Electronics, but 1 is inactive
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(4)
      expect(result.data.every(p => p.category === "Electronics")).toBe(true)
      expect(result.data).toMatchSnapshot()
    })

    it("returns products in Books category", async () => {
      const result = await listProductsByCategory(pool, { category: "Books" })

      // 5 Books, but 1 is inactive
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(4)
      expect(result.data).toMatchSnapshot()
    })

    it("returns empty array for non-existent category", async () => {
      const result = await listProductsByCategory(pool, { category: "NonExistent" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })

  describe("listProductsPaginated", () => {
    it("returns paginated active products - first page", async () => {
      const result = await listProductsPaginated(pool, { limit: 5, offset: 0 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(5)
      expect(result.data).toMatchSnapshot()
    })

    it("returns paginated active products - second page", async () => {
      const result = await listProductsPaginated(pool, { limit: 5, offset: 5 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(5)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("listProductsCursor", () => {
    it("returns products after cursor ID", async () => {
      const result = await listProductsCursor(pool, { id: 5, limit: 5 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeLessThanOrEqual(5)
      expect(result.data.every(p => p.id > 5)).toBe(true)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("searchProducts", () => {
    it("finds products matching search term", async () => {
      const result = await searchProducts(pool, { plaintoTsquery: "keyboard" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })

    it("finds products with category in search", async () => {
      const result = await searchProducts(pool, { plaintoTsquery: "electronics" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("searchProductsRanked", () => {
    it("returns ranked search results with pagination", async () => {
      const result = await searchProductsRanked(pool, {
        plaintoTsquery: "cotton",
        limit: 10,
        offset: 0,
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("searchProductsWithHighlight", () => {
    it("returns search results with highlighted descriptions", async () => {
      const result = await searchProductsWithHighlight(pool, { plaintoTsquery: "wireless" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("searchProductsWebStyle", () => {
    it("searches with web-style query syntax", async () => {
      const result = await searchProductsWebStyle(pool, { websearchToTsquery: "coffee maker" })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("createProduct", () => {
    it("creates a new product with all fields", async () => {
      const result = await createProduct(pool, {
        sku: "NEW-001",
        name: "New Test Product",
        description: "A test product description",
        priceCents: 9999,
        stockQuantity: 100,
        isActive: true,
        category: "Test",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.sku).toBe("NEW-001")
      expect(result.data!.name).toBe("New Test Product")
      expect(result.data!.price_cents).toBe(9999)
      expect(result.data!.stock_quantity).toBe(100)
      expect(result.data!.is_active).toBe(true)
    })

    it("creates a product without optional fields", async () => {
      const result = await createProduct(pool, {
        sku: "NEW-002",
        name: "Minimal Product",
        priceCents: 1000,
        stockQuantity: 10,
        isActive: true,
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.description).toBeNull()
      expect(result.data!.category).toBeNull()
    })
  })

  describe("updateProduct", () => {
    it("updates an existing product", async () => {
      const result = await updateProduct(pool, {
        id: 1,
        sku: "ELEC-001-UPD",
        name: "Updated Wireless Headphones",
        description: "Updated description",
        priceCents: 15999,
        stockQuantity: 45,
        isActive: true,
        category: "Electronics",
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).not.toBeNull()
      expect(result.data!.sku).toBe("ELEC-001-UPD")
      expect(result.data!.name).toBe("Updated Wireless Headphones")
      expect(result.data!.price_cents).toBe(15999)
    })

    it("returns null when updating non-existent product", async () => {
      const result = await updateProduct(pool, {
        id: 9999,
        sku: "NONEXISTENT",
        name: "Nonexistent",
        priceCents: 100,
        stockQuantity: 0,
        isActive: false,
      })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data).toBeNull()
    })
  })

  describe("updateProductStock", () => {
    it("updates product stock quantity", async () => {
      // Get initial stock
      const before = await getProduct(pool, { id: 2 })
      expect(before.success).toBe(true)
      if (!before.success) throw new Error("expected success")
      const initialStock = before.data!.stock_quantity

      // Update stock (adds to existing)
      await updateProductStock(pool, { id: 2, stockQuantity: 10 })

      // Verify the change
      const after = await getProduct(pool, { id: 2 })
      expect(after.success).toBe(true)
      if (!after.success) throw new Error("expected success")
      expect(after.data!.stock_quantity).toBe(initialStock + 10)
    })
  })

  describe("deactivateProduct", () => {
    it("deactivates a product", async () => {
      // Create a product to deactivate
      const created = await createProduct(pool, {
        sku: "DEACTIVATE-001",
        name: "To Deactivate",
        priceCents: 100,
        stockQuantity: 1,
        isActive: true,
      })
      expect(created.success).toBe(true)
      if (!created.success) throw new Error("expected success")
      const productId = created.data!.id

      // Deactivate it
      await deactivateProduct(pool, { id: productId })

      // Verify it's inactive
      const result = await getProduct(pool, { id: productId })
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")
      expect(result.data!.is_active).toBe(false)
    })
  })

  describe("deleteProduct", () => {
    it("deletes an existing product", async () => {
      // Create a product to delete
      const created = await createProduct(pool, {
        sku: "DELETE-001",
        name: "To Delete",
        priceCents: 100,
        stockQuantity: 1,
        isActive: true,
      })
      expect(created.success).toBe(true)
      if (!created.success) throw new Error("expected success")
      const productId = created.data!.id

      // Delete it
      await deleteProduct(pool, { id: productId })

      // Verify it's gone
      const result = await getProduct(pool, { id: productId })
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")
      expect(result.data).toBeNull()
    })
  })

  describe("getProductsByIds", () => {
    it("returns products for given IDs", async () => {
      const result = await getProductsByIds(pool, { ids: [1, 6, 11, 16] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(4)
      // Validate structure without volatile timestamps
      const ids = result.data.map(p => p.id).sort((a, b) => a - b)
      expect(ids).toEqual([1, 6, 11, 16])
      expect(result.data.every(p => p.sku && p.name)).toBe(true)
    })

    it("returns empty array for non-existent IDs", async () => {
      const result = await getProductsByIds(pool, { ids: [9998, 9999] })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })

  describe("countProductsByCategory", () => {
    it("returns product counts grouped by category", async () => {
      const result = await countProductsByCategory(pool)

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data).toMatchSnapshot()
    })
  })

  describe("getLowStockProducts", () => {
    it("returns products with low stock", async () => {
      const result = await getLowStockProducts(pool, { stockQuantity: 50 })

      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBeGreaterThan(0)
      expect(result.data.every(p => p.stock_quantity < 50)).toBe(true)
      expect(result.data).toMatchSnapshot()
    })

    it("returns empty array when no low stock products", async () => {
      const result = await getLowStockProducts(pool, { stockQuantity: 1 })

      // All seeded products have stock >= 3
      expect(result.success).toBe(true)
      if (!result.success) throw new Error("expected success")

      expect(result.data.length).toBe(0)
    })
  })
})
