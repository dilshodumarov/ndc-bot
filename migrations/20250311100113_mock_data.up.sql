INSERT INTO client_type (name) VALUES
  ('Retail'),
  ('Wholesale'),
  ('VIP'),
  ('Corporate'),
  ('Partner');

INSERT INTO role (name, client_type_id) VALUES
  ('Admin',       (SELECT id FROM client_type WHERE name = 'Retail')),
  ('Manager',     (SELECT id FROM client_type WHERE name = 'Wholesale')),
  ('Sales',       (SELECT id FROM client_type WHERE name = 'VIP')),
  ('Support',     (SELECT id FROM client_type WHERE name = 'Corporate')),
  ('Guest',       (SELECT id FROM client_type WHERE name = 'Partner'));

INSERT INTO "user" (name, surname, username, password, bith_date, tg_user_name, phone, instagram, client_from, role_id) VALUES
  ('John', 'Doe', 'johndoe', 'hashedpass1', '1990-01-01', '@johndoe', '998901234567', 'johndoe_insta', 'Facebook', (SELECT id FROM role WHERE name = 'Admin')),
  ('Jane', 'Smith', 'janesmith', 'hashedpass2', '1991-02-02', '@janesmith', '998901234568', 'janesmith_insta', 'Instagram', (SELECT id FROM role WHERE name = 'Manager')),
  ('Bob', 'Johnson', 'bobj', 'hashedpass3', '1989-03-03', '@bobj', '998901234569', 'bobj_insta', 'Telegram', (SELECT id FROM role WHERE name = 'Sales')),
  ('Alice', 'Williams', 'alicew', 'hashedpass4', '1995-04-04', '@alicew', '998901234570', 'alicew_insta', 'Referral', (SELECT id FROM role WHERE name = 'Support')),
  ('Charlie', 'Brown', 'charlieb', 'hashedpass5', '1988-05-05', '@charlieb', '998901234571', 'charlieb_insta', 'Website', (SELECT id FROM role WHERE name = 'Guest'));

INSERT INTO category (name) VALUES
  ('Electronics'),
  ('Clothing'),
  ('Books'),
  ('Home Appliances'),
  ('Toys');

INSERT INTO attribute (name, category_id) VALUES
  ('Color', (SELECT id FROM category WHERE name = 'Clothing')),
  ('Size', (SELECT id FROM category WHERE name = 'Clothing')),
  ('Author', (SELECT id FROM category WHERE name = 'Books')),
  ('Brand', (SELECT id FROM category WHERE name = 'Electronics')),
  ('Material', (SELECT id FROM category WHERE name = 'Home Appliances'));

INSERT INTO product (name, category_id, short_info, description, cost, discount_cost, discount, count) VALUES
  ('Smartphone', (SELECT id FROM category WHERE name = 'Electronics'), 'Latest model', 'A brand new smartphone with great features.', 1000, 900, 10, 10),
  ('Jeans', (SELECT id FROM category WHERE name = 'Clothing'), 'Denim jeans', 'Comfortable and stylish jeans.', 50, 45, 10, 5),
  ('Cookbook', (SELECT id FROM category WHERE name = 'Books'), 'Healthy recipes', 'A collection of healthy recipes.', 30, 25, 17, 15),
  ('Blender', (SELECT id FROM category WHERE name = 'Home Appliances'), 'High-speed blender', 'Perfect for smoothies.', 150, 120, 20, 45),
  ('Toy Car', (SELECT id FROM category WHERE name = 'Toys'), 'Remote controlled', 'A fun remote-controlled car.', 60, 50, 17, 0);

INSERT INTO integration (name) VALUES
  ('Telegram Bot'),
  ('CRM System'),
  ('Payment Gateway'),
  ('Email Marketing'),
  ('SMS Notification');

INSERT INTO "order" (user_id, integration_id, status, status_changed_time) VALUES
  ((SELECT id FROM "user" WHERE username = 'johndoe'), (SELECT id FROM integration WHERE name = 'Telegram Bot'), 'Pending', NOW()),
  ((SELECT id FROM "user" WHERE username = 'janesmith'), (SELECT id FROM integration WHERE name = 'CRM System'), 'Shipped', NOW()),
  ((SELECT id FROM "user" WHERE username = 'bobj'), (SELECT id FROM integration WHERE name = 'Payment Gateway'), 'Completed', NOW()),
  ((SELECT id FROM "user" WHERE username = 'alicew'), (SELECT id FROM integration WHERE name = 'Email Marketing'), 'Cancelled', NOW()),
  ((SELECT id FROM "user" WHERE username = 'charlieb'), (SELECT id FROM integration WHERE name = 'SMS Notification'), 'Processing', NOW());

INSERT INTO order_products (order_id, product_id, count) VALUES
  ((SELECT id FROM "order" LIMIT 1 OFFSET 0), (SELECT id FROM product WHERE name = 'Smartphone'), 1),
  ((SELECT id FROM "order" LIMIT 1 OFFSET 1), (SELECT id FROM product WHERE name = 'Jeans'), 2),
  ((SELECT id FROM "order" LIMIT 1 OFFSET 2), (SELECT id FROM product WHERE name = 'Cookbook'), 3),
  ((SELECT id FROM "order" LIMIT 1 OFFSET 3), (SELECT id FROM product WHERE name = 'Blender'), 1),
  ((SELECT id FROM "order" LIMIT 1 OFFSET 4), (SELECT id FROM product WHERE name = 'Toy Car'), 5);