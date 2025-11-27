-- Insert Clothing category data
INSERT INTO product_categories (product_id, code, name) VALUES
((SELECT id FROM products WHERE code = 'PROD001'), gen_random_uuid(), 'Clothing'),
((SELECT id FROM products WHERE code = 'PROD004'), gen_random_uuid(), 'Clothing'),
((SELECT id FROM products WHERE code = 'PROD007'), gen_random_uuid(), 'Clothing');

-- Insert Shoes category data
INSERT INTO product_categories (product_id, code, name) VALUES
((SELECT id FROM products WHERE code = 'PROD002'), gen_random_uuid(), 'Shoes'),
((SELECT id FROM products WHERE code = 'PROD006'), gen_random_uuid(), 'Shoes');

-- Insert Accessories category data
INSERT INTO product_categories (product_id, code, name) VALUES
((SELECT id FROM products WHERE code = 'PROD003'), gen_random_uuid(), 'Accessories'),
((SELECT id FROM products WHERE code = 'PROD005'), gen_random_uuid(), 'Accessories'),
((SELECT id FROM products WHERE code = 'PROD008'), gen_random_uuid(), 'Accessories');