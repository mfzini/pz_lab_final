SET timezone = '-03:00';

DROP DATABASE IF EXISTS example;
CREATE DATABASE example;

-- Nota: Si usas psql, este comando conecta a la base de datos recién creada:
\c example;

CREATE TABLE IF NOT EXISTS addresses(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    street varchar(36) NOT NULL,
    house varchar(10) NOT NULL,
    floor int,
    door varchar(5) NOT NULL,
    coordinates point,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS people(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(36) NOT NULL,
    surname varchar(36) NOT NULL,
    email varchar(36) UNIQUE NOT NULL,
    phone varchar(20),
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS people_addresses(
    personId uuid NOT NULL,
    CONSTRAINT fk_people_addresses_person FOREIGN KEY (personId) REFERENCES people(id),
    addressId uuid NOT NULL,
    CONSTRAINT fk_people_addresses_address FOREIGN KEY (addressId) REFERENCES addresses(id),
    detail text,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS users(
    personId uuid PRIMARY KEY,
    CONSTRAINT fk_users_person FOREIGN KEY (personId) REFERENCES people(id),
    username varchar(20) NOT NULL,
    password varchar(255) NOT NULL,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS ingredients(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(36) NOT NULL,
    stock decimal(6,3) NOT NULL,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS dishes(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(36) NOT NULL,
    picture varchar(255),
    price decimal(10,2) NOT NULL,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS dishes_ingredients(
    dishId uuid NOT NULL,
    CONSTRAINT fk_dishes_ingredients_dish FOREIGN KEY (dishId) REFERENCES dishes(id),
    ingredientId uuid NOT NULL,
    CONSTRAINT fk_dishes_ingredients_ingredient FOREIGN KEY (ingredientId) REFERENCES ingredients(id),
    ammount decimal(4,3) NOT NULL,
);

CREATE TABLE IF NOT EXISTS orders_status(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(20) NOT NULL,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS orders(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client uuid,
    CONSTRAINT fk_orders_client FOREIGN KEY (client) REFERENCES users(personId),
    detail text,
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);

CREATE TABLE IF NOT EXISTS status_orders(
    orderId uuid NOT NULL,
    CONSTRAINT fk_status_orders_order FOREIGN KEY (orderId) REFERENCES orders(id),
    status uuid NOT NULL,
    CONSTRAINT fk_status_orders_status FOREIGN KEY (status) REFERENCES orders_status(id),
    createdAt timestamp DEFAULT now(),
    updatedAt timestamp,
    deletedAt timestamp
);