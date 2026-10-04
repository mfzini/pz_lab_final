set timezone = '-03:00';

create table if not exists people(
    id uuid primary key default gen_random_uuid(),
    name varchar(36) not null,
    surname varchar(36) not null,
    email varchar(36) unique not null,
    phone varchar(20),
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists addresses(
    id uuid primary key default gen_random_uuid(),
    street varchar(36) not null,
    house varchar(10) not null,
    floor int,
    door varchar(5),
    coordinates point,
    detail text,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists people_addresses(
    person_id uuid not null,
    constraint fk_people_addresses_person foreign key (person_id) references people(id),
    address_id uuid not null,
    constraint fk_people_addresses_address foreign key (address_id) references addresses(id),
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists users(
    person_id uuid primary key,
    constraint fk_users_person foreign key (person_id) references people(id),
    username varchar(20) not null,
    password varchar(255) not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists products(
    id uuid primary key default gen_random_uuid(),
    name varchar(36) unique not null,
    price decimal(10,2) not null,
    portrait varchar(250),
    amount int default 0,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists orders(
    id uuid primary key default gen_random_uuid(),
    client_id uuid,
    constraint fk_orders_client foreign key (client_id) references users(person_id),
    address uuid,
    constraint fk_order_address foreign key (address) references addresses(id),
    detail text,
    status varchar(30),
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists orders_products(
    order_id uuid,
    constraint fk_order_products_order foreign key (order_id) references orders(id),
    product_id uuid,
    constraint fk_order_products_product foreign key (product_id) references products(id),
    ppu decimal(10, 2),
    amount int default 1
);


-----------------------------------------------------------------------------------------------------------------------------


create table if not exists products_ingredients(
    product_id uuid not null,
    constraint fk_products_ingredients_product foreign key (product_id) references products(id),
    ingredient_id uuid not null,
    constraint fk_products_ingredients_ingredient foreign key (ingredient_id) references ingredients(id),
    amount decimal(4,3) not null
);

create table if not exists orders_status(
    id uuid primary key default gen_random_uuid(),
    name varchar(20) not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);



create table if not exists status_orders(
    order_id uuid not null,
    constraint fk_status_orders_order foreign key (order_id) references orders(id),
    status uuid not null,
    constraint fk_status_orders_status foreign key (status) references orders_status(id),
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists ingredients(
    id uuid primary key default gen_random_uuid(),
    name varchar(36) not null,
    stock decimal(6,3) not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists images(
    id uuid primary key default gen_random_uuid(),
    original_name varchar(255) not null,
    sha256 varchar(64) not null,
    mimetype varchar(10) not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists products_images(
    product_id uuid,
    constraint fk_product_image foreign key (product_id) references products(id),
    image_id uuid,
    constraint fk_image_product foreign key (image_id) references images(id)
);