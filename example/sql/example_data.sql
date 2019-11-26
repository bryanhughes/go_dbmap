INSERT INTO test_schema.user (first_name, last_name, email) VALUES ('Big', 'Man', 'bigman@gmail.com');

INSERT INTO test_schema.address ( address1, city, state, country, postcode ) VALUES ( '2030 3rd Street #14', 'San Francisco', 'CA', 'US', '94107');
INSERT INTO test_schema.address ( address1, city, state, country, postcode ) VALUES ( '2002 3rd Street #02', 'San Francisco', 'CA', 'US', '94107');


INSERT INTO public.product ( product_name, sku, produced) VALUES ( 'foo', '129837', (TIMESTAMP '1967-06-07 15:36:38'));
INSERT INTO public.product ( product_name, sku, produced) VALUES ( 'bar', '983457', (TIMESTAMP '2011-05-16 15:36:38'));
INSERT INTO example_a (column_a, column_b, column_e, column_g) VALUES ('Captain', 'Kirk', ARRAY['a', 'b', 'c'], ARRAY[true, false, true]);
-- Postgres will allow inserting an array into a non-array column, but you can not do array operations on it
INSERT INTO example_a (column_a, column_b, column_e, column_g) VALUES (ARRAY['Captain', 'Jim'], 'Kirk', ARRAY['d', 'e', 'c'], ARRAY[true, false, true]);