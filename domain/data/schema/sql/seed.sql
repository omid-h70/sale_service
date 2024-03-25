INSERT INTO users (user_id, name, email, roles, password_hash, department, enabled, date_created, date_updated) VALUES
('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', 'admin@example.com', '{ADMIN,USER}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', NULL, true, '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', 'user@example.com', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, '2019-03-24 00:00:00', '2019-03-24 00:00:00')
ON CONFLICT DO NOTHING;
-- ON CONFLICT DO NOTHING -> if data exists do nothing

INSERT INTO products (product_id, user_id, name, cost, quantity, date_created, date_updated) VALUES
('52af2580-428f-11ee-be56-0242ac120002', '5cf37266-3473-4006-984f-9325122678b7', 'Comic Books', 50, 42, '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
('52af2968-428f-11ee-be56-0242ac120002', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'McDonalds Toys', 75, 120, '2019-03-24 00:00:00', '2019-03-24 00:00:00')
ON CONFLICT DO NOTHING;

INSERT INTO sales (sale_id, product_id, quantity, paid, date_created) VALUES
('52af2a8a-428f-11ee-be56-0242ac120002', '52af2580-428f-11ee-be56-0242ac120002', 2, 100, '2019-03-24 00:00:00'),
('52af2b7a-428f-11ee-be56-0242ac120002', '52af2968-428f-11ee-be56-0242ac120002', 5, 250, '2019-03-24 00:00:00'),
('52af2c6a-428f-11ee-be56-0242ac120002', '52af2968-428f-11ee-be56-0242ac120002', 3, 225, '2019-03-24 00:00:00')
ON CONFLICT DO NOTHING;