CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES products(id),
    flash_sale_id BIGINT NOT NULL REFERENCES flash_sales(id),
    price INT NOT NULL,
    status TEXT NOT NULL,  -- pending / paid / canceled
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
);