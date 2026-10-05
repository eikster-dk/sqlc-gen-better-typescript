import { PgTypes } from "@effect/sql-pg"
import { Option, Result, Schema } from "effect"
import { describe, expect, it } from "vitest"
import { EchoBigIntsParams } from "../src/repositories/BinaryTypesRepositoryRequest.js"
import { EchoBigIntsResult } from "../src/repositories/BinaryTypesRepositoryResponse.js"
import {
  GetCustomerOrderStatsResult,
  GetOrderLineTotalResult,
  GetProductSalesStatsResult,
} from "../src/repositories/OrdersRepositoryResponse.js"

describe("generated bigint wire schemas", () => {
  it.each([-9223372036854775808n, -9007199254740993n, 0n, 9007199254740993n, 9223372036854775807n])(
    "preserves %s through request encoding, binary transport, and result decoding",
    (value) => {
      const params = Schema.encodeSync(EchoBigIntsParams)({ value })
      const bytes = Result.getOrThrow(PgTypes.encode(params.value, PgTypes.OID.int8))
      const arrayBytes = Result.getOrThrow(PgTypes.encode([params.value], PgTypes.OID.int8Array))

      const result = Schema.decodeUnknownSync(EchoBigIntsResult)({
        value: Result.getOrThrow(PgTypes.decode(bytes, PgTypes.OID.int8, 1)),
        values: Result.getOrThrow(PgTypes.decode(arrayBytes, PgTypes.OID.int8Array, 1)),
      })

      expect(result).toEqual({ value, values: [value] })
    }
  )
})

describe("aggregate wire schemas", () => {
  const binaryRoundTrip = (value: unknown, oid: number) =>
    Result.getOrThrow(PgTypes.decode(Result.getOrThrow(PgTypes.encode(value, oid)), oid, 1))

  it("decodes integer sums as bigint and numeric averages as decimal strings", () => {
    const stats = Schema.decodeUnknownSync(GetCustomerOrderStatsResult)({
      total_orders: binaryRoundTrip(3n, PgTypes.OID.int8),
      total_spent: binaryRoundTrip(24997n, PgTypes.OID.int8),
      avg_order_value: binaryRoundTrip("8332.3333333333333333", PgTypes.OID.numeric),
    })
    expect(stats).toEqual({
      total_orders: 3n,
      total_spent: 24997n,
      avg_order_value: Option.some("8332.3333333333333333"),
    })
  })

  it.each([0n, 79994n])("decodes a coalesced line total of %s", (total) => {
    expect(Schema.decodeUnknownSync(GetOrderLineTotalResult)({
      total: binaryRoundTrip(total, PgTypes.OID.int8),
    })).toEqual({ total })
  })

  it("decodes coalesced product sales aggregates", () => {
    expect(Schema.decodeUnknownSync(GetProductSalesStatsResult)({
      id: 1,
      sku: "ELEC-001",
      name: "Wireless Headphones",
      total_sold: binaryRoundTrip(9n, PgTypes.OID.int8),
      total_revenue: binaryRoundTrip(134991n, PgTypes.OID.int8),
    })).toEqual({
      id: 1,
      sku: "ELEC-001",
      name: "Wireless Headphones",
      total_sold: 9n,
      total_revenue: 134991n,
    })
  })
})
