import type { Pool } from "pg"

// Fixed timestamps for deterministic snapshots
export const TIMESTAMPS = {
  // Customers created dates (spread over January 2024)
  customer1: "2024-01-01 10:00:00+00",
  customer2: "2024-01-02 11:30:00+00",
  customer3: "2024-01-03 09:15:00+00",
  customer4: "2024-01-04 14:45:00+00",
  customer5: "2024-01-05 16:20:00+00",
  customer6: "2024-01-06 08:00:00+00",
  customer7: "2024-01-07 12:30:00+00",
  customer8: "2024-01-08 17:45:00+00",
  customer9: "2024-01-09 10:10:00+00",
  customer10: "2024-01-10 15:00:00+00",

  // Products created dates
  product1: "2024-01-01 08:00:00+00",
  product2: "2024-01-01 08:05:00+00",
  product3: "2024-01-01 08:10:00+00",
  product4: "2024-01-01 08:15:00+00",
  product5: "2024-01-01 08:20:00+00",
  product6: "2024-01-02 09:00:00+00",
  product7: "2024-01-02 09:05:00+00",
  product8: "2024-01-02 09:10:00+00",
  product9: "2024-01-02 09:15:00+00",
  product10: "2024-01-02 09:20:00+00",
  product11: "2024-01-03 10:00:00+00",
  product12: "2024-01-03 10:05:00+00",
  product13: "2024-01-03 10:10:00+00",
  product14: "2024-01-03 10:15:00+00",
  product15: "2024-01-03 10:20:00+00",
  product16: "2024-01-04 11:00:00+00",
  product17: "2024-01-04 11:05:00+00",
  product18: "2024-01-04 11:10:00+00",
  product19: "2024-01-04 11:15:00+00",
  product20: "2024-01-04 11:20:00+00",

  // Orders created dates (spread over January-February 2024)
  order1: "2024-01-15 10:00:00+00",
  order2: "2024-01-16 11:30:00+00",
  order3: "2024-01-17 14:00:00+00",
  order4: "2024-01-18 09:45:00+00",
  order5: "2024-01-19 16:30:00+00",
  order6: "2024-01-20 13:15:00+00",
  order7: "2024-01-21 10:45:00+00",
  order8: "2024-01-22 15:20:00+00",
  order9: "2024-01-23 11:00:00+00",
  order10: "2024-01-24 14:30:00+00",
  order11: "2024-01-25 09:00:00+00",
  order12: "2024-01-26 17:00:00+00",
  order13: "2024-01-27 12:45:00+00",
  order14: "2024-01-28 08:30:00+00",
  order15: "2024-01-29 16:15:00+00",

  // Order lines created dates
  orderLine1: "2024-01-15 10:01:00+00",
  orderLine2: "2024-01-15 10:02:00+00",
  orderLine3: "2024-01-16 11:31:00+00",
  orderLine4: "2024-01-17 14:01:00+00",
  orderLine5: "2024-01-17 14:02:00+00",
  orderLine6: "2024-01-17 14:03:00+00",
  orderLine7: "2024-01-18 09:46:00+00",
  orderLine8: "2024-01-19 16:31:00+00",
  orderLine9: "2024-01-19 16:32:00+00",
  orderLine10: "2024-01-20 13:16:00+00",
  orderLine11: "2024-01-21 10:46:00+00",
  orderLine12: "2024-01-21 10:47:00+00",
  orderLine13: "2024-01-22 15:21:00+00",
  orderLine14: "2024-01-22 15:22:00+00",
  orderLine15: "2024-01-22 15:23:00+00",
  orderLine16: "2024-01-23 11:01:00+00",
  orderLine17: "2024-01-24 14:31:00+00",
  orderLine18: "2024-01-24 14:32:00+00",
  orderLine19: "2024-01-25 09:01:00+00",
  orderLine20: "2024-01-26 17:01:00+00",
  orderLine21: "2024-01-26 17:02:00+00",
  orderLine22: "2024-01-27 12:46:00+00",
  orderLine23: "2024-01-27 12:47:00+00",
  orderLine24: "2024-01-28 08:31:00+00",
  orderLine25: "2024-01-29 16:16:00+00",
  orderLine26: "2024-01-29 16:17:00+00",
  orderLine27: "2024-01-29 16:18:00+00",
  orderLine28: "2024-01-29 16:19:00+00",
  orderLine29: "2024-01-29 16:20:00+00",
  orderLine30: "2024-01-29 16:21:00+00",
} as const

export const seedDatabase = async (pool: Pool): Promise<void> => {
  // Seed customers (10 customers with varied data)
  await pool.query(`
    INSERT INTO customers (id, email, name, phone, created_at, updated_at) VALUES
    (1, 'alice@example.com', 'Alice Johnson', '+1-555-0101', '${TIMESTAMPS.customer1}', '${TIMESTAMPS.customer1}'),
    (2, 'bob@example.com', 'Bob Smith', '+1-555-0102', '${TIMESTAMPS.customer2}', '${TIMESTAMPS.customer2}'),
    (3, 'carol@example.com', 'Carol Williams', NULL, '${TIMESTAMPS.customer3}', '${TIMESTAMPS.customer3}'),
    (4, 'david@example.com', 'David Brown', '+1-555-0104', '${TIMESTAMPS.customer4}', '${TIMESTAMPS.customer4}'),
    (5, 'eve@example.com', 'Eve Davis', NULL, '${TIMESTAMPS.customer5}', '${TIMESTAMPS.customer5}'),
    (6, 'frank@example.com', 'Frank Miller', '+1-555-0106', '${TIMESTAMPS.customer6}', '${TIMESTAMPS.customer6}'),
    (7, 'grace@example.com', 'Grace Wilson', '+1-555-0107', '${TIMESTAMPS.customer7}', '${TIMESTAMPS.customer7}'),
    (8, 'henry@example.com', 'Henry Moore', NULL, '${TIMESTAMPS.customer8}', '${TIMESTAMPS.customer8}'),
    (9, 'ivy@example.com', 'Ivy Taylor', '+1-555-0109', '${TIMESTAMPS.customer9}', '${TIMESTAMPS.customer9}'),
    (10, 'jack@example.com', 'Jack Anderson', '+1-555-0110', '${TIMESTAMPS.customer10}', '${TIMESTAMPS.customer10}')
  `)

  // Reset customers sequence
  await pool.query(`SELECT setval('customers_id_seq', 10)`)

  // Seed products (20 products across categories)
  await pool.query(`
    INSERT INTO products (id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at) VALUES
    -- Electronics (5 products)
    (1, 'ELEC-001', 'Wireless Headphones', 'Premium noise-canceling wireless headphones with 30-hour battery life', 14999, 50, true, 'Electronics', '${TIMESTAMPS.product1}', '${TIMESTAMPS.product1}'),
    (2, 'ELEC-002', 'USB-C Hub', 'Multi-port USB-C hub with HDMI, USB-A, and SD card reader', 4999, 100, true, 'Electronics', '${TIMESTAMPS.product2}', '${TIMESTAMPS.product2}'),
    (3, 'ELEC-003', 'Mechanical Keyboard', 'RGB mechanical gaming keyboard with Cherry MX switches', 12999, 30, true, 'Electronics', '${TIMESTAMPS.product3}', '${TIMESTAMPS.product3}'),
    (4, 'ELEC-004', 'Wireless Mouse', 'Ergonomic wireless mouse with adjustable DPI', 3999, 75, true, 'Electronics', '${TIMESTAMPS.product4}', '${TIMESTAMPS.product4}'),
    (5, 'ELEC-005', 'Old MP3 Player', 'Discontinued MP3 player', 2999, 5, false, 'Electronics', '${TIMESTAMPS.product5}', '${TIMESTAMPS.product5}'),
    
    -- Clothing (5 products)
    (6, 'CLTH-001', 'Cotton T-Shirt', 'Comfortable 100% cotton t-shirt in various colors', 1999, 200, true, 'Clothing', '${TIMESTAMPS.product6}', '${TIMESTAMPS.product6}'),
    (7, 'CLTH-002', 'Denim Jeans', 'Classic fit denim jeans with stretch fabric', 4999, 80, true, 'Clothing', '${TIMESTAMPS.product7}', '${TIMESTAMPS.product7}'),
    (8, 'CLTH-003', 'Winter Jacket', 'Insulated winter jacket with water-resistant exterior', 9999, 40, true, 'Clothing', '${TIMESTAMPS.product8}', '${TIMESTAMPS.product8}'),
    (9, 'CLTH-004', 'Running Shoes', 'Lightweight running shoes with cushioned sole', 7999, 60, true, 'Clothing', '${TIMESTAMPS.product9}', '${TIMESTAMPS.product9}'),
    (10, 'CLTH-005', 'Summer Shorts', 'Seasonal summer shorts - clearance', 1499, 10, false, 'Clothing', '${TIMESTAMPS.product10}', '${TIMESTAMPS.product10}'),
    
    -- Home & Garden (5 products)
    (11, 'HOME-001', 'Coffee Maker', 'Programmable 12-cup coffee maker with thermal carafe', 7999, 25, true, 'Home & Garden', '${TIMESTAMPS.product11}', '${TIMESTAMPS.product11}'),
    (12, 'HOME-002', 'Plant Pot Set', 'Set of 3 ceramic plant pots in various sizes', 2999, 150, true, 'Home & Garden', '${TIMESTAMPS.product12}', '${TIMESTAMPS.product12}'),
    (13, 'HOME-003', 'LED Desk Lamp', 'Adjustable LED desk lamp with USB charging port', 3499, 90, true, 'Home & Garden', '${TIMESTAMPS.product13}', '${TIMESTAMPS.product13}'),
    (14, 'HOME-004', 'Throw Blanket', 'Soft fleece throw blanket 50x60 inches', 2499, 120, true, 'Home & Garden', '${TIMESTAMPS.product14}', '${TIMESTAMPS.product14}'),
    (15, 'HOME-005', 'Garden Hose', 'Expandable garden hose with spray nozzle', 3999, 45, true, 'Home & Garden', '${TIMESTAMPS.product15}', '${TIMESTAMPS.product15}'),
    
    -- Books (5 products)
    (16, 'BOOK-001', 'TypeScript Handbook', 'Comprehensive guide to TypeScript programming', 3999, 100, true, 'Books', '${TIMESTAMPS.product16}', '${TIMESTAMPS.product16}'),
    (17, 'BOOK-002', 'Effect Programming', 'Functional programming with Effect in TypeScript', 4499, 75, true, 'Books', '${TIMESTAMPS.product17}', '${TIMESTAMPS.product17}'),
    (18, 'BOOK-003', 'Database Design', 'Principles of relational database design', 5499, 50, true, 'Books', '${TIMESTAMPS.product18}', '${TIMESTAMPS.product18}'),
    (19, 'BOOK-004', 'Clean Code', 'Writing maintainable and readable code', 4999, 85, true, 'Books', '${TIMESTAMPS.product19}', '${TIMESTAMPS.product19}'),
    (20, 'BOOK-005', 'Old JavaScript Book', 'Outdated JavaScript reference', 1999, 3, false, 'Books', '${TIMESTAMPS.product20}', '${TIMESTAMPS.product20}')
  `)

  // Reset products sequence
  await pool.query(`SELECT setval('products_id_seq', 20)`)

  // Seed orders (15 orders with various statuses)
  await pool.query(`
    INSERT INTO orders (id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at) VALUES
    (1, 1, 'delivered', 19998, '123 Main St, New York, NY 10001', '123 Main St, New York, NY 10001', 'Leave at door', '${TIMESTAMPS.order1}', '${TIMESTAMPS.order1}'),
    (2, 1, 'shipped', 4999, '123 Main St, New York, NY 10001', '123 Main St, New York, NY 10001', NULL, '${TIMESTAMPS.order2}', '${TIMESTAMPS.order2}'),
    (3, 2, 'processing', 25997, '456 Oak Ave, Los Angeles, CA 90001', '456 Oak Ave, Los Angeles, CA 90001', 'Gift wrapping requested', '${TIMESTAMPS.order3}', '${TIMESTAMPS.order3}'),
    (4, 2, 'pending', 7999, '456 Oak Ave, Los Angeles, CA 90001', NULL, NULL, '${TIMESTAMPS.order4}', '${TIMESTAMPS.order4}'),
    (5, 3, 'confirmed', 11998, '789 Pine Rd, Chicago, IL 60601', '789 Pine Rd, Chicago, IL 60601', 'Express delivery', '${TIMESTAMPS.order5}', '${TIMESTAMPS.order5}'),
    (6, 4, 'delivered', 3999, '321 Elm St, Houston, TX 77001', '321 Elm St, Houston, TX 77001', NULL, '${TIMESTAMPS.order6}', '${TIMESTAMPS.order6}'),
    (7, 4, 'cancelled', 14999, '321 Elm St, Houston, TX 77001', '321 Elm St, Houston, TX 77001', 'Customer requested cancellation', '${TIMESTAMPS.order7}', '${TIMESTAMPS.order7}'),
    (8, 5, 'delivered', 15996, '654 Maple Dr, Phoenix, AZ 85001', '654 Maple Dr, Phoenix, AZ 85001', 'Second floor apartment', '${TIMESTAMPS.order8}', '${TIMESTAMPS.order8}'),
    (9, 5, 'shipped', 2999, '654 Maple Dr, Phoenix, AZ 85001', '654 Maple Dr, Phoenix, AZ 85001', NULL, '${TIMESTAMPS.order9}', '${TIMESTAMPS.order9}'),
    (10, 6, 'processing', 9498, '987 Cedar Ln, Philadelphia, PA 19101', '987 Cedar Ln, Philadelphia, PA 19101', 'Call before delivery', '${TIMESTAMPS.order10}', '${TIMESTAMPS.order10}'),
    (11, 7, 'pending', 4999, '147 Birch Way, San Antonio, TX 78201', NULL, NULL, '${TIMESTAMPS.order11}', '${TIMESTAMPS.order11}'),
    (12, 8, 'refunded', 7998, '258 Spruce Ct, San Diego, CA 92101', '258 Spruce Ct, San Diego, CA 92101', 'Refunded due to damage', '${TIMESTAMPS.order12}', '${TIMESTAMPS.order12}'),
    (13, 8, 'confirmed', 10998, '258 Spruce Ct, San Diego, CA 92101', '258 Spruce Ct, San Diego, CA 92101', NULL, '${TIMESTAMPS.order13}', '${TIMESTAMPS.order13}'),
    (14, 9, 'shipped', 3999, '369 Willow Blvd, Dallas, TX 75201', '369 Willow Blvd, Dallas, TX 75201', 'Fragile items', '${TIMESTAMPS.order14}', '${TIMESTAMPS.order14}'),
    (15, 10, 'pending', 27495, '741 Aspen Pl, San Jose, CA 95101', '741 Aspen Pl, San Jose, CA 95101', 'Large order - verify stock', '${TIMESTAMPS.order15}', '${TIMESTAMPS.order15}')
  `)

  // Reset orders sequence
  await pool.query(`SELECT setval('orders_id_seq', 15)`)

  // Seed order lines (30 order lines)
  await pool.query(`
    INSERT INTO order_lines (id, order_id, product_id, quantity, unit_price_cents, discount_cents, created_at) VALUES
    -- Order 1 (delivered): 2 items
    (1, 1, 1, 1, 14999, 0, '${TIMESTAMPS.orderLine1}'),
    (2, 1, 4, 1, 4999, 0, '${TIMESTAMPS.orderLine2}'),
    
    -- Order 2 (shipped): 1 item
    (3, 2, 2, 1, 4999, 0, '${TIMESTAMPS.orderLine3}'),
    
    -- Order 3 (processing): 3 items
    (4, 3, 3, 1, 12999, 0, '${TIMESTAMPS.orderLine4}'),
    (5, 3, 6, 2, 1999, 0, '${TIMESTAMPS.orderLine5}'),
    (6, 3, 12, 3, 2999, 0, '${TIMESTAMPS.orderLine6}'),
    
    -- Order 4 (pending): 1 item
    (7, 4, 11, 1, 7999, 0, '${TIMESTAMPS.orderLine7}'),
    
    -- Order 5 (confirmed): 2 items
    (8, 5, 7, 1, 4999, 0, '${TIMESTAMPS.orderLine8}'),
    (9, 5, 8, 1, 9999, 3000, '${TIMESTAMPS.orderLine9}'),
    
    -- Order 6 (delivered): 1 item
    (10, 6, 4, 1, 3999, 0, '${TIMESTAMPS.orderLine10}'),
    
    -- Order 7 (cancelled): 2 items (but order was cancelled)
    (11, 7, 1, 1, 14999, 0, '${TIMESTAMPS.orderLine11}'),
    (12, 7, 2, 1, 4999, 4999, '${TIMESTAMPS.orderLine12}'),
    
    -- Order 8 (delivered): 3 items
    (13, 8, 6, 3, 1999, 0, '${TIMESTAMPS.orderLine13}'),
    (14, 8, 13, 2, 3499, 0, '${TIMESTAMPS.orderLine14}'),
    (15, 8, 14, 1, 2499, 500, '${TIMESTAMPS.orderLine15}'),
    
    -- Order 9 (shipped): 1 item
    (16, 9, 12, 1, 2999, 0, '${TIMESTAMPS.orderLine16}'),
    
    -- Order 10 (processing): 2 items
    (17, 10, 16, 1, 3999, 0, '${TIMESTAMPS.orderLine17}'),
    (18, 10, 17, 1, 4499, 0, '${TIMESTAMPS.orderLine18}'),
    
    -- Order 11 (pending): 1 item
    (19, 11, 7, 1, 4999, 0, '${TIMESTAMPS.orderLine19}'),
    
    -- Order 12 (refunded): 2 items
    (20, 12, 6, 2, 1999, 0, '${TIMESTAMPS.orderLine20}'),
    (21, 12, 4, 1, 3999, 0, '${TIMESTAMPS.orderLine21}'),
    
    -- Order 13 (confirmed): 2 items
    (22, 13, 18, 1, 5499, 0, '${TIMESTAMPS.orderLine22}'),
    (23, 13, 19, 1, 4999, 0, '${TIMESTAMPS.orderLine23}'),
    
    -- Order 14 (shipped): 1 item
    (24, 14, 15, 1, 3999, 0, '${TIMESTAMPS.orderLine24}'),
    
    -- Order 15 (pending): 6 items - large order
    (25, 15, 1, 1, 14999, 1000, '${TIMESTAMPS.orderLine25}'),
    (26, 15, 3, 1, 12999, 500, '${TIMESTAMPS.orderLine26}'),
    (27, 15, 9, 1, 7999, 0, '${TIMESTAMPS.orderLine27}'),
    (28, 15, 11, 1, 7999, 1000, '${TIMESTAMPS.orderLine28}'),
    (29, 15, 16, 2, 3999, 500, '${TIMESTAMPS.orderLine29}'),
    (30, 15, 17, 1, 4499, 500, '${TIMESTAMPS.orderLine30}')
  `)

  // Reset order_lines sequence
  await pool.query(`SELECT setval('order_lines_id_seq', 30)`)
}