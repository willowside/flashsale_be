-- 1. 一致性檢查：DB 庫存扣除量 vs 成功訂單數
-- 邏輯：初始庫存 - 當前 DB 庫存 = Status 為 success 的訂單數
WITH consistency_audit AS (
    SELECT 
        fsp.product_id,
        fsp.sale_stock AS current_db_stock,
        -- init stock 假設為 100
        100 AS initial_allocated_stock, 
        (SELECT COUNT(*) FROM orders o 
         WHERE o.product_id = fsp.product_id 
         AND o.status = 'success') AS success_order_count,
        (SELECT COUNT(*) FROM orders o 
         WHERE o.product_id = fsp.product_id 
         AND o.status = 'failed') AS failed_order_count
    FROM flash_sale_products fsp
    WHERE fsp.product_id = 1001 -- product_id 可調整
)
SELECT 
    product_id,
    initial_allocated_stock,
    current_db_stock,
    success_order_count,
    (initial_allocated_stock - success_order_count) AS expected_db_stock,
    CASE 
        WHEN (initial_allocated_stock - success_order_count) = current_db_stock THEN 'PASS ✅'
        ELSE 'FAIL X - DB Stock Mismatch'
    END AS db_consistency_status,
    failed_order_count AS orders_in_dlq_or_failed
FROM consistency_audit;

-- 2. 虛擬庫存檢查：Redis vs DB (方案 A 允許微小延遲，但最終應接近)
-- 注意：此處需從外部帶入 Redis 數值，或觀察兩者差距
SELECT 
    'Redis vs DB Gap Check' AS metric,
    'Manual Step: Run `redis-cli GET flashsale:stock:1001` and compare with DB current_db_stock' AS instruction;

-- 3. 異常檢測：是否存在超賣（庫存變負數）
SELECT 
    'Over-selling Detection' AS check_type,
    CASE WHEN MIN(sale_stock) < 0 THEN 'FAIL X - Negative Stock!' ELSE 'PASS' END AS result
FROM flash_sale_products;