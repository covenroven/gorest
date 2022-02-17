create table orders (
	order_id SERIAL primary key,
	customer_name varchar(50) not null,
	ordered_at timestamp not null
)

create table items (
	item_id serial primary key,
	item_code varchar(50) not null,
	description text,
	quantity int not null,
	order_id int not null,
	constraint fk_order_id foreign key(order_id) references orders(order_id)
)
