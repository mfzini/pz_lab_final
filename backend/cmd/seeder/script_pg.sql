-- ============================================================================
-- Esquema del backend.
-- Lo que los repos de Go tocan: users, addresses, categories, products,
-- orders y orders_products (6 tablas).
--
-- Reset total (borra tablas y data):
--   drop table if exists orders_products, orders, products, categories,
--     addresses, users cascade;
--
-- Ejecutar:
--   psql -U postgres -d example -f cmd/seeder/script_pg.sql
--
-- Nota: con timestamptz no hace falta setear timezone en la sesión.
-- ============================================================================

-- users <- usersRepo: create (id, name, surname, email, phone, password),
-- update (name, surname), updateField (role), soft delete (deleted_at), select *
create table if not exists users(
    id uuid primary key default gen_random_uuid(),
    name varchar(36) not null,
    surname varchar(36) not null,
    email varchar(60) unique not null,
    phone varchar(20),
    avatar varchar(250),
    role varchar(16) not null default 'client',
    password varchar(255) not null,      -- bcrypt ocupa 60
    created_at timestamptz not null default now(),
    updated_at timestamptz,
    deleted_at timestamptz
);

-- addresses <- addressesRepo: create (id, user_id, street, house, door, floor,
-- coordinates), update (...), soft delete, FindById, FindByStreetAndHouse,
-- FindByUserID (where user_id = $1).
create table if not exists addresses(
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id),
    street varchar(36) not null,
    house varchar(10) not null,           -- texto: "12", "12B", "s/n"
    floor int,
    door varchar(5),
    coordinates point,                   -- el código lo escribe como texto "x,y"
    detail text,
    created_at timestamptz not null default now(),
    updated_at timestamptz,
    deleted_at timestamptz
);

create index if not exists addresses_user_id_idx on addresses(user_id);

-- categories <- categoriesRepo: maestra chica (nombre único). Platos N:1
-- categorías: es la 6ª tabla y da la relación N:1 con products.
create table if not exists categories(
    id uuid primary key default gen_random_uuid(),
    name varchar(40) unique not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz,
    deleted_at timestamptz
);

-- products <- productsRepo: create (id, name, price, portrait, category_id),
-- update (con coalesce en portrait), soft delete, list/findById con join a
-- categories (p.category_id, c.name as category_name).
create table if not exists products(
    id uuid primary key default gen_random_uuid(),
    name varchar(36) unique not null,
    price int not null,                  -- centavos: 1250 = $12,50 (Go no usa decimales)
    portrait varchar(250),
    amount int not null default 0,
    category_id uuid references categories(id),  -- anulable: plato sin categoría
    created_at timestamptz not null default now(),
    updated_at timestamptz,
    deleted_at timestamptz
);

-- orders <- ordersRepo: create (id, client_id, address_id, detail, status),
-- findById/findByuserID (select o.status), updateStatus (update status).
create table if not exists orders(
    id uuid primary key default gen_random_uuid(),
    client_id uuid not null references users(id),
    address_id uuid not null references addresses(id),
    detail text,
    -- pendiente -> en proceso -> despachada
    status varchar(16) not null default 'pending'
        check (status in ('pending', 'preparing', 'shipped')),
    created_at timestamptz not null default now(),
    updated_at timestamptz                -- cuándo cambió de estado la última vez
);

create index if not exists orders_client_id_idx on orders(client_id);

-- orders_products <- ordersRepo: insert (order_id, product_id, amount, ppu) y
-- fetchProducts (od.amount, od.ppu). La PK evita repetir el mismo plato.
create table if not exists orders_products(
    order_id uuid not null references orders(id),
    product_id uuid not null references products(id),
    amount int not null default 1,
    ppu int not null,                    -- centavos, igual que products.price
    primary key (order_id, product_id)
);
