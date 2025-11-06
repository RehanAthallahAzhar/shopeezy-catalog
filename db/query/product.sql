-- name: InsertProduct :one
INSERT INTO products (
  id, 
  seller_id, 
  "name", 
  price, 
  stock, 
  discount, 
  "type", 
  "description", 
  created_at, 
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
) RETURNING *;

-- name: GetAllProducts :many
SELECT 
  id,
  seller_id,
  "name",
  price,
  stock,
  discount,
  "type",
  "description",
  created_at,
  updated_at
FROM products
WHERE deleted_at IS NULL;

-- name: GetProductByID :one
SELECT 
  id,
  seller_id,
  "name",
  price,
  stock,
  discount,
  "type",
  "description",
  created_at,
  updated_at
FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProductByIDs :many
SELECT 
  id,
  seller_id,
  "name",
  price,
  stock,
  discount,
  "type",
  "description",
  created_at,
  updated_at
FROM products
WHERE id = ANY($1::uuid[]) AND deleted_at IS NULL;

-- name: GetProductsBySellerID :many
SELECT 
  id,
  seller_id,
  "name",
  price,
  stock,
  discount,
  "type",
  "description",
  created_at,
  updated_at
FROM products
WHERE seller_id = $1 AND deleted_at IS NULL;

-- name: GetProductsByName :many
SELECT 
  id,
  seller_id,
  "name",
  price,
  stock,
  discount,
  "type",
  "description",
  created_at,
  updated_at
FROM products
WHERE name LIKE $1 AND deleted_at IS NULL;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, price = $3, stock = $4, discount = $5, type = $6, description = $7, updated_at = NOW()
WHERE id = $1 AND seller_id = $8
RETURNING *;

-- name: DeleteProduct :one
DELETE FROM products WHERE id = $1 
RETURNING *;

-- name: GetProductStock :one
SELECT stock FROM products WHERE id = $1;

-- name: UpdateProductStock :one
UPDATE products SET stock = $2 WHERE id = $1 
RETURNING *;
