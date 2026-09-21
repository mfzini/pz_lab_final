set global time_zone = '-03:00';

drop database if exists example;
create database if not exists example;

use example;

create table if not exists addresses(
    id uuid primary key default uuid(),
    street varchar(36) not null,
    house  varchar(10) not null,
    floor  int,
    door varchar(5) not null,
    coordinates point,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists people(
    id uuid primary key default uuid(),
    name varchar(36) not null,
    surname varchar(36) not null,
    email varchar(36) not null,
    phone varchar(20),
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists people_addresses(
    personId uuid not null,
    constraint foreign key (personId) references people(id),
    addressId uuid not null,
    constraint foreign key (addressId) references addresses(id),
    detail text,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists users(
	personId uuid primary key,
	constraint foreign key (personId) references people(id),
    username varchar(20) not null,
    password varchar(255) not null,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists ingredients(
    id uuid primary key default uuid(),
    name varchar(36) not null,
    stock decimal(6,3) not null
);

create table if not exists dishes(
    id uuid primary key default uuid(),
    name varchar(36) not null,
    picture varchar(255),
    price decimal(10,2) not null,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists dishes_ingredients(
    dishId uuid not null,
    constraint foreign key (dishId) references dishes(id),
    ingredientId uuid not null,
    constraint foreign key (ingredientId) references ingredients(id),
    ammount decimal(4,3) not null,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists orders_status(
    id uuid primary key default uuid(),
    name varchar(20) not null,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists orders(
    id uuid primary key default uuid(),
    client uuid,
    constraint foreign key (client) references users(personId),
    detail text,
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);

create table if not exists status_orders(
    orderId uuid not null,
    constraint foreign key (orderId) references orders(id),
    status uuid not null,
    constraint foreign key (status) references orders_status(id),
    createdAt timestamp default now(),
    updatedAt timestamp,
    deletedAt timestamp
);