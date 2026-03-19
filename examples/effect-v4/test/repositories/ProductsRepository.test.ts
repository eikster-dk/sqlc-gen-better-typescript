import { describe, beforeAll, afterAll, expect, it } from "@effect/vitest"
import { Effect, Layer, Option } from "effect"
import type { StartedPostgreSqlContainer } from "@testcontainers/postgresql"
import { SqlError } from "effect/unstable/sql"
import { startPostgres, stopPostgres, makeTestLayer } from "../setup/testcontainers.js"
import { seedDatabase } from "../setup/seed.js"
import {
  ProductsRepository,
  productsRepositoryLive,
} from "../../src/repositories/ProductsRepository.js"

describe("ProductsRepository", () => {
  let container: StartedPostgreSqlContainer
  let testLayer: Layer.Layer<ProductsRepository, SqlError.SqlError>

  beforeAll(async () => {
    container = await startPostgres()
    const sqlLayer = makeTestLayer(container)
    testLayer = productsRepositoryLive.pipe(Layer.provide(sqlLayer))

    // Seed the database
    await Effect.runPromise(
      seedDatabase.pipe(Effect.provide(sqlLayer))
    )
  }, 120000)

  afterAll(async () => {
    await stopPostgres()
  })

  describe("getProduct", () => {
    it.effect("returns product when found", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getProduct({ id: 1 })

        expect(Option.isSome(result)).toBe(true)
        expect(Option.getOrNull(result)).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when product not found", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getProduct({ id: 9999 })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getProductBySku", () => {
    it.effect("returns product when found by SKU", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getProductBySku({ sku: "ELEC-001" })

        expect(Option.isSome(result)).toBe(true)
        expect(Option.getOrNull(result)).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when SKU not found", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getProductBySku({ sku: "NONEXISTENT-SKU" })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listProducts", () => {
    it.effect("returns all products ordered by name", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProducts()

        expect(result.length).toBe(20)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listActiveProducts", () => {
    it.effect("returns only active products", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listActiveProducts()

        // 20 total - 3 inactive = 17 active
        expect(result.length).toBe(17)
        expect(result.every(p => p.is_active)).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listProductsByCategory", () => {
    it.effect("returns products in Electronics category", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProductsByCategory({ category: "Electronics" })

        // 5 Electronics, but 1 is inactive
        expect(result.length).toBe(4)
        expect(result.every(p => Option.getOrNull(p.category) === "Electronics")).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns products in Books category", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProductsByCategory({ category: "Books" })

        // 5 Books, but 1 is inactive
        expect(result.length).toBe(4)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array for non-existent category", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProductsByCategory({ category: "NonExistent" })

        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listProductsPaginated", () => {
    it.effect("returns paginated active products - first page", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProductsPaginated({ limit: 5, offset: 0 })

        expect(result.length).toBe(5)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns paginated active products - second page", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProductsPaginated({ limit: 5, offset: 5 })

        expect(result.length).toBe(5)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("listProductsCursor", () => {
    it.effect("returns products after cursor ID", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.listProductsCursor({ id: 5, limit: 5 })

        expect(result.length).toBeLessThanOrEqual(5)
        expect(result.every(p => p.id > 5)).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("searchProducts", () => {
    it.effect("finds products matching search term", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.searchProducts({ plaintoTsquery: "keyboard" })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("finds products with category in search", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.searchProducts({ plaintoTsquery: "electronics" })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("searchProductsRanked", () => {
    it.effect("returns ranked search results with pagination", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.searchProductsRanked({
          plaintoTsquery: "cotton",
          limit: 10,
          offset: 0,
        })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("searchProductsWithHighlight", () => {
    it.effect("returns search results with highlighted descriptions", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.searchProductsWithHighlight({ plaintoTsquery: "wireless" })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("searchProductsWebStyle", () => {
    it.effect("searches with web-style query syntax", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.searchProductsWebStyle({ websearchToTsquery: "coffee maker" })

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("createProduct", () => {
    it.effect("creates a new product with all fields", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.createProduct({
          sku: "NEW-001",
          name: "New Test Product",
          description: "A test product description",
          priceCents: 9999,
          stockQuantity: 100,
          isActive: true,
          category: "Test",
        })

        expect(Option.isSome(result)).toBe(true)
        const product = Option.getOrNull(result)!
        expect(product.sku).toBe("NEW-001")
        expect(product.name).toBe("New Test Product")
        expect(product.price_cents).toBe(9999)
        expect(product.stock_quantity).toBe(100)
        expect(product.is_active).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("creates a product without optional fields", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.createProduct({
          sku: "NEW-002",
          name: "Minimal Product",
          priceCents: 1000,
          stockQuantity: 10,
          isActive: true,
        })

        expect(Option.isSome(result)).toBe(true)
        const product = Option.getOrNull(result)!
        expect(Option.isNone(product.description)).toBe(true)
        expect(Option.isNone(product.category)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateProduct", () => {
    it.effect("updates an existing product", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.updateProduct({
          id: 1,
          sku: "ELEC-001-UPD",
          name: "Updated Wireless Headphones",
          description: "Updated description",
          priceCents: 15999,
          stockQuantity: 45,
          isActive: true,
          category: "Electronics",
        })

        expect(Option.isSome(result)).toBe(true)
        const product = Option.getOrNull(result)!
        expect(product.sku).toBe("ELEC-001-UPD")
        expect(product.name).toBe("Updated Wireless Headphones")
        expect(product.price_cents).toBe(15999)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns None when updating non-existent product", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.updateProduct({
          id: 9999,
          sku: "NONEXISTENT",
          name: "Nonexistent",
          priceCents: 100,
          stockQuantity: 0,
          isActive: false,
        })

        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("updateProductStock", () => {
    it.effect("updates product stock quantity", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository

        // Get initial stock
        const before = yield* repo.getProduct({ id: 2 })
        const initialStock = Option.getOrNull(before)!.stock_quantity

        // Update stock (adds to existing)
        yield* repo.updateProductStock({ id: 2, stockQuantity: 10 })

        // Verify the change
        const after = yield* repo.getProduct({ id: 2 })
        expect(Option.getOrNull(after)!.stock_quantity).toBe(initialStock + 10)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("deactivateProduct", () => {
    it.effect("deactivates a product", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository

        // Create a product to deactivate
        const created = yield* repo.createProduct({
          sku: "DEACTIVATE-001",
          name: "To Deactivate",
          priceCents: 100,
          stockQuantity: 1,
          isActive: true,
        })
        const productId = Option.getOrNull(created)!.id

        // Deactivate it
        yield* repo.deactivateProduct({ id: productId })

        // Verify it's inactive
        const result = yield* repo.getProduct({ id: productId })
        expect(Option.getOrNull(result)!.is_active).toBe(false)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("deleteProduct", () => {
    it.effect("deletes an existing product", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository

        // Create a product to delete
        const created = yield* repo.createProduct({
          sku: "DELETE-001",
          name: "To Delete",
          priceCents: 100,
          stockQuantity: 1,
          isActive: true,
        })
        const productId = Option.getOrNull(created)!.id

        // Delete it
        yield* repo.deleteProduct({ id: productId })

        // Verify it's gone
        const result = yield* repo.getProduct({ id: productId })
        expect(Option.isNone(result)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getProductsByIds", () => {
    it.effect("returns products for given IDs", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getProductsByIds({ ids: [1, 6, 11, 16] })

        expect(result.length).toBe(4)
        // Validate structure without volatile timestamps
        const ids = result.map(p => p.id).sort((a, b) => a - b)
        expect(ids).toEqual([1, 6, 11, 16])
        expect(result.every(p => p.sku && p.name)).toBe(true)
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array for non-existent IDs", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getProductsByIds({ ids: [9998, 9999] })

        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("countProductsByCategory", () => {
    it.effect("returns product counts grouped by category", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.countProductsByCategory()

        expect(result.length).toBeGreaterThan(0)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )
  })

  describe("getLowStockProducts", () => {
    it.effect("returns products with low stock", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getLowStockProducts({ stockQuantity: 50 })

        expect(result.length).toBeGreaterThan(0)
        expect(result.every(p => p.stock_quantity < 50)).toBe(true)
        expect(result).toMatchSnapshot()
      }).pipe(Effect.provide(testLayer))
    )

    it.effect("returns empty array when no low stock products", () =>
      Effect.gen(function* () {
        const repo = yield* ProductsRepository
        const result = yield* repo.getLowStockProducts({ stockQuantity: 1 })

        // All seeded products have stock >= 3
        expect(result.length).toBe(0)
      }).pipe(Effect.provide(testLayer))
    )
  })
})
