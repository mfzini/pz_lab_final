set timezone = '-03:00';

drop database if exists example;
create database example;

create table if not exists addresses(
    id uuid primary key default gen_random_uuid(),
    street varchar(36) not null,
    house varchar(10) not null,
    floor int,
    door varchar(5),
    coordinates point,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

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

create table if not exists people_addresses(
    person_id uuid not null,
    constraint fk_people_addresses_person foreign key (person_id) references people(id),
    address_id uuid not null,
    constraint fk_people_addresses_address foreign key (address_id) references addresses(id),
    detail text,
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

create table if not exists users(
    person_id uuid primary key,
    constraint fk_users_person foreign key (person_id) references people(id),
    username varchar(20) not null,
    password varchar(255) not null,
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

create table if not exists dishes(
    id uuid primary key default gen_random_uuid(),
    name varchar(36) not null,
    price decimal(10,2) not null,
    portrait uuid,
    constraint fk_dish_image foreign key (portrait) references images(id),
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists dishes_images(
    dish_id uuid,
    constraint fk_dish_image foreign key (dish_id) references dishes(id),
    image_id uuid,
    constraint fk_image_dish foreign key (image_id) references images(id)
);

create table if not exists dishes_ingredients(
    dish_id uuid not null,
    constraint fk_dishes_ingredients_dish foreign key (dish_id) references dishes(id),
    ingredient_id uuid not null,
    constraint fk_dishes_ingredients_ingredient foreign key (ingredient_id) references ingredients(id),
    ammount decimal(4,3) not null
);

create table if not exists orders_status(
    id uuid primary key default gen_random_uuid(),
    name varchar(20) not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table if not exists orders(
    id uuid primary key default gen_random_uuid(),
    client uuid,
    constraint fk_orders_client foreign key (client) references users(person_id),
    detail text,
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