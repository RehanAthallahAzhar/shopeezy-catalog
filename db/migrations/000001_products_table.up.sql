CREATE TABLE products (
    id UUID PRIMARY KEY,
    seller_id UUID NOT NULL,
    name TEXT NOT NULL,
    price INT NOT NULL,
    stock INT NOT NULL,
    discount INT,
    "type" TEXT,
    "description" TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);