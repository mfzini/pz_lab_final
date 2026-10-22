# pz_lab_final

## Descripción general

**PZ Lab** es un sistema de pedidos online para un restaurante chico,
desarrollado como laboratorio final. **El problema que resuelve**: los
restaurantes chicas toman los pedidos a mano (teléfono, papel, memoria):
se piden mal los platos, se pierden pedidos y el dueño no sabe en qué
estado está cada uno. Acá eso pasa a la web:

- **Clientes** (registran con dirección obligatoria): recorren el catálogo
  de platos con foto y precio, arman un carrito, eligen la dirección de
  entrega con búsqueda asíncrona, confirman el pedido y siguen su estado
  en el historial con detalle (platos, subtotales, total).
- **Administrador**: carga y edita platos (foto + categoría), ve **todos**
  los pedidos con filtros por fecha y estado, los va despachando, y
  administra usuarios.

| Pieza | Qué es |
|---|---|
| `backend/` | API REST en **Go** (`net/http` + `pgx` + JWT), 6 tablas PostgreSQL |
| `frontend/` | SPA en **Vue 3 + Vite** (bun) que consume esa API nativamente |
| `.devcontainer/` | Compose del devcontainer: `app` (Go/Bun) + `db` (PostgreSQL) |
| `Makefile` | `make back` / `make front` |

---

## 1. Levantarlo

1. Abrir la carpeta en VS Code y reabrir en el devcontainer
   (levanta `app` + `postgres` con el `.env` de `.devcontainer/.env`).
   El `postCreateCommand` ya corre `go mod tidy` y `bun install`.
2. Crear el esquema y cargar los datos de prueba (ver §3).
3. Servicios:

   ```bash
   make back    # backend → :3000
   make front   # vite   → :5173
   ```

   El compose expone Postgres en `:5432`; los puertos 3000/5173 van
   re-enviados por el devcontainer.

## 2. Variables de entorno

Las carga el container desde `.devcontainer/.env`:

| Variable | Para qué |
|---|---|
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_HOST`, `POSTGRES_PORT` | conexión (`internal/db/connection.go`) |
| `PATH_UPLOADS` | dónde se guardan las fotos de los platos (`GET /uploads/...`) |
| `JWT_SECRET` | firma del access token (default `"secret"`) |
| `JWT_REFRESH_SECRET` | firma del refresh token (default: lo mismo que `JWT_SECRET`) |

## 3. Base de datos

El esquema vive en **`backend/cmd/seeder/script_pg.sql`**: **6 tablas** y
nada más (`users`, `addresses`, `categories`, `products`, `orders`,
`orders_products`). Cada bloque está comentado con qué query de Go lo
alimenta. Relaciones: `addresses.user_id`, `orders.client_id` y
`orders.address_id` son **1:N**; `orders_products` es la **N:M** entre
orders y products (entidad que depende de las dos, con ppu y amount); y
`products.category_id` es **N:1** hacia categories (anulable).

```bash
# reset total (borra tablas y data)
psql -U postgres -d <db> -c \
  "drop table if exists orders_products, orders, products, categories,
    addresses, users cascade;"

psql -U postgres -d <db> -f backend/cmd/seeder/script_pg.sql
```

### Datos de prueba

```bash
cd backend && go run ./cmd/seeder
```

Carga (23 filas) sobre el esquema ya creado:

- **3 usuarios** con la misma contraseña `clave1234`:

  | Rol | Email |
  |---|---|
  | admin | `admin@test.com` |
  | client | `juan@test.com` |
  | client | `maria@test.com` |

- **2 direcciones** (una por cliente; `house` con `742` y `12B` para probar
  que es texto).
- **3 categorías** (`Comida`, `Bebidas`, `Postres`).
- **6 platos** con precios en centavos ($5,00 a $15,00), cada uno con su
  categoría asignada.
- **3 pedidos** — uno por estado y repartidos en los últimos días
  (`shipped` hace 5, `preparing` hace 2, `pending` hoy), para probar los
  filtros de fecha y estado de `GET /ordenes/admin`.

Los fixtures dejan las columnas anulables en `null` de verdad (platos sin
foto, María sin teléfono y sin coordenadas en su dirección, un pedido sin
detalle) — sirven para que `go test ./...` verifique que las lecturas
toleran `null`.

Es **idempotente**: usa ids fijos y `on conflict do nothing`, así que
re-correrlo no duplica nada. Las contraseñas se hashean con bcrypt al
correr (no hay hashes hardcodeados en el script).

Sin este esquema, `POST /direcciones` y `POST /ordenes` fallan en runtime
(las columnas `addresses.user_id`, `orders.address_id` y `orders.status` no
existen en un esquema anterior).

## 4. Arquitectura del backend

Patrón por recurso, en `internal/domain/<recurso>/`:

```
Handler  → HTTP: decodifica, aplica permisos, arma la respuesta
Service  → negocio: validaciones, ownership, qué se puede modificar
Repo     → SQL con pgx (named args / placeholders, nada de string concat)
```

- `cmd/server/main.go` arma todo (wiring) y escucha en `:3000`.
- `internal/security` — JWT, roles, middleware de permisos y CORS.
- `internal/api` — `Error` (respuesta `{"error": "..."}`) y `WriteJson`.
- `internal/domain/common` — `PageResponse` + `PageParams` (paginación).

### Convenciones

| Tema | Cómo |
|---|---|
| IDs | `uuid` v7 (monótono, generado en el server) |
| Borrado | *soft delete* (`deleted_at`) en `users`, `addresses`, `products` |
| Precios | **enteros en centavos**: `1250` = $12,50 (Go no maneja decimales) |
| Nulos | toda columna anulable de la DB mapea a un campo puntero (`*string`, `*int`, `*time.Time`): leer un `null` no rompe, devuelve `null` en el JSON |
| Errores | siempre `{"error": "<mensaje>"}` con el status correspondiente; los fallos internos devuelven un mensaje genérico (nada de errores de DB en la respuesta) |
| Listados | paginados (ver §6) |
| Roles | `client` y `admin`; el rol lo pone el server, nunca el body |
| 401 vs 403 | 401 = sin token (con `WWW-Authenticate`); 403 = con token pero sin permiso |
| Ownership | se lee del token (`claims.Subject`), jamás del body |

Chequeos:

```bash
cd backend && go build ./... && go vet ./... && gofmt -l .

# tests (leen contra la DB real con el env cargado; sin env se saltean)
go test ./...
```

---

## 5. Autenticación

JWT con dos fichas (issuer `app`):

| Token | Vida | `aud` | Uso |
|---|---|---|---|
| access | 15 min | `access` | `Authorization: Bearer <token>` en cada request |
| refresh | 7 días | `refresh` | `POST /refresh` para uno nuevo |

```jsonc
// POST /login  →  200
{ "token": "...", "refreshToken": "...", "user": { "id": "...", "role": "client", ... } }
// el access también sale en el header Authorization de la misma respuesta
```

```jsonc
// POST /refresh — acepta body...
{ "refreshToken": "..." }
// ...o el refresh token en el header Authorization; devuelve lo mismo que /login
```

El cliente guarda los dos tokens en `localStorage` y manda el access en el
header de cada request.

`POST /init` crea el primer admin (sólo funciona si no existe ningún admin;
responde `404` si ya hay uno) y devuelve la sesión recién creada:

```jsonc
{ "email": "admin@...", "password": "..." }   // → 201 + tokens + user (role: admin)
```

---

## 6. Paginación y errores

Todo listado usa la misma pauta:

```
GET /recurso?page=1&size=10
```

- `page` default **1**, `size` default **10**, tope `size=100`.
- Respuesta (mismo formato en todos, el de `common.PageResponse`):

```json
{ "Data": [ ... ], "CurrentPage": 1, "TotalPages": 3 }
```

Errores: `{"error": "mensaje"}` con `400` (validación), `401` (sin sesión),
`403` (sin permiso), `404` (no existe / no es tuyo), `406` (email repetido),
`500` (infra).

---

## 7. Matriz de permisos

| Acción | público | client | admin |
|---|:-:|:-:|:-:|
| registrarse, login, refresh | ✅ | — | — |
| ver el menú (`GET /platos`) | ✅ | ✅ | ✅ |
| su perfil, su dirección, su historial | — | ✅ | ✅ |
| crear pedidos | — | ✅ | ✅ |
| cargar/editar platos y categorías | — | ❌ | ✅ |
| cambiar estado de un pedido | — | ❌ | ✅ |
| ver todos los pedidos (fecha/estado) | — | ❌ | ✅ |
| listar/borrar datos de otros usuarios | — | ❌ | ✅ |
| ver el listado de usuarios y asignarles rol | — | ❌ | ✅ |

Los usuarios **no ven datos de otros usuarios**: las lecturas de usuario,
dirección y orden validan *dueño o admin*. El admin gestiona los roles desde
`/admin/usuarios` (`GET /usuarios` + `PATCH /user/{id}/role`); un admin no
quitarse su propio rol (`400`).

---

## 8. Endpoints

### Usuarios

| Método y ruta | Quién | Body | Éxito |
|---|---|---|---|
| `POST /user` | público | `{name, surname, email, password, phone?}` | `201` (rol siempre `client`) |
| `POST /login` | público | `{email, password}` | `200` + tokens |
| `POST /refresh` | público | `{refreshToken}` (o header) | `200` + tokens |
| `POST /init` | público | `{email, password}` | `201` + tokens (admin) |
| `GET /user/{id}` | dueño/admin | — | `200` + user |
| `PATCH /user/{id}` | dueño/admin | `{name?, surname?}` | `200` + user |
| `PATCH /user/{id}/password` | dueño/admin | `{password}` | `204` |
| `PATCH /user/{id}/email` | dueño/admin | `{email}` | `204` |
| `PATCH /user/{id}/phone` | dueño/admin | `{phone}` (≤20) | `204` |
| `PATCH /user/{id}/avatar` | dueño/admin | `{avatar}` (`""`/`null` lo vacía) | `204` |
| `POST /user/{id}/avatar` | dueño/admin | multipart: `picture` (file) | `200` + `{avatar}` (lo sube a `/uploads` y lo setea) |
| `PATCH /user/{id}/role` | **admin** | `{role: "admin"\|"client"}` | `204` |
| `GET /usuarios` | **admin** | `?page&size` | `200` paginado (listado del panel de usuarios; sin `password`) |

`PATCH /user/{id}` sólo toca nombre y apellido: los demás campos tienen su
propio endpoint para que un patch a medias no pise el resto.

### Direcciones

| Método y ruta | Quién | Body / query | Éxito |
|---|---|---|---|
| `POST /direcciones` | cliente | `{street, house, floor?, door?, coordinates?, detail?}` | `201` (dueño y id salen del token/server) |
| `GET /direcciones/{id}` | dueño/admin | — | `200` |
| `GET /usuarios/{id}/direcciones` | dueño/admin | `?street&page&size` (`street` filtra por calle: es la búsqueda asíncrona del checkout) | `200` paginado |
| `PATCH /direcciones/{id}` | dueño/admin | como el create | `200` |
| `DELETE /direcciones/{id}` | dueño/admin | — | `204` |
| `GET /buscar/direcciones` | **admin** | `?street&house&page&size` | `200` paginado |

### Platos

| Método y ruta | Quién | Body / query | Éxito |
|---|---|---|---|
| `GET /platos/list` | público | `?name&sortby=name\|price\|created_at&page&size` | `200` paginado |
| `GET /platos/{id}` | público | — | `200` (`404` si no existe) |
| `POST /platos` | **admin** | multipart: `product` (JSON) + `picture` (file) | `201` |
| `PATCH /platos/{id}` | **admin** | mismo multipart; el id sale del **path** y `picture` es **opcional** (sin foto, se mantiene la actual) | `200` (`404` si no existe) |

```jsonc
// product (multipart) — price en centavos; categoryId anulable
{ "name": "Hamburguesa", "price": 1250, "categoryId": "0000…0211" | null }
```

### Categorías

| Método y ruta | Quién | Body / query | Éxito |
|---|---|---|---|
| `GET /categorias` | público | `?name&page&size` (`name` filtra con ilike: la búsqueda asíncrona del form de platos) | `200` paginado |
| `POST /categorias` | **admin** | `{name}` | `201` (`409` si ya existe, `400` si viene vacío) |

### Órdenes

| Método y ruta | Quién | Body / query | Éxito |
|---|---|---|---|
| `POST /ordenes` | cliente | `{detail?, address: {id}, products: [{id, amount}]}` | `201` |
| `GET /ordenes` | cliente | `?page&size` — **historial propio** | `200` paginado |
| `GET /ordenes/{id}` | dueño/admin | — | `200` |
| `GET /ordenes/admin` | **admin** | `?status&from&to&page&size` | `200` paginado |
| `PATCH /ordenes/{id}/estado` | **admin** | `{status}` | `200` |

- El cliente de la orden sale del **token**, nunca del body; el total se
  calcula al leer (jamás del body).
- La `address.id` tiene que ser del propio usuario (se valida contra la DB
  con el token): si es ajena o no existe → `404`/`400`, y en la respuesta
  queda la dirección canónica de la DB.
- Los platos se validan uno por uno antes de insertar: mínimo un plato,
  `amount ≥ 1`, sin repetir la misma `id` y el plato tiene que existir
  (y no estar borrado) — todo caso contrario responde `400`/`404`.
- `status`: `pending` → `preparing` → `shipped`. Todas entran como `pending`.
- Filtros del admin: `status` (`pending`/`preparing`/`shipped`), `from` y
  `to` en `AAAA-MM-DD` (ambos inclusive). Orden: más recientes primero.

```jsonc
// POST /ordenes
{ "detail": "sin cebolla",
  "address": { "id": "3f2a..." },
  "products": [ { "id": "9b1c...", "amount": 2 } ] }
```

### Imágenes

| Método y ruta | Quién |
|---|---|
| `GET /uploads/{filename}` | público — sirve `PATH_UPLOADS` |

---

## 9. Frontend

Vue 3 + vue-router + Vite, paquetes con **bun**:

```bash
cd frontend
bun install
bun dev      # :5173
bun run lint # eslint + oxlint
```

### Sesión y requests

- **`src/auth.js`** — estado de sesión en `localStorage` (`token`,
  `refreshToken`, `user`) con `login()`, `register()` e `isAdmin`.
- **`src/api.js`** — `apiFetch()`: arma la URL contra `:3000`, agrega el
  `Authorization: Bearer` y, si el access venció (401), renueva con
  `POST /refresh` y reintenta una vez; si no puede, cierra la sesión y manda
  al login. Los errores del backend (`{"error": "..."}`) salen como excepción
  con ese mensaje, para mostrarlos en la vista.

### Rutas y permisos

| Ruta | Quién entra | Qué hace |
|---|---|---|
| `/` | cualquiera | catálogo de platos (público, paginado); con sesión: carrito + confirmación del pedido con **búsqueda asíncrona de dirección** |
| `/about` | cualquiera | página estática "about" |
| `/login` | sin sesión | login; vuelve a `?next=` o al home del rol |
| `/persona` | sin sesión | registro: cuenta (`POST /user`) + **dirección obligatoria** (`POST /direcciones`) + login automático |
| `/historial` | logueado | historial propio; clic en un pedido abre el detalle (platos, subtotales y total) |
| `/direcciones` | logueado | CRUD de sus direcciones (alta/edición en modal, con botón propio) |
| `/perfil` | logueado | perfil propio: nombre/apellido + **subir o quitar el avatar** (multipart `POST /user/{id}/avatar`) |
| `/admin/platos` | admin | platos: alta y edición con foto y **categoría por búsqueda asíncrona** en modal (`POST/PATCH /platos`); `/platos` redirige acá |
| `/admin/pedidos` | admin | todas las órdenes con filtros de estado/fechas + cambio de estado |
| `/admin/usuarios` | admin | listado paginado de usuarios con su rol: el admin **asigna o quita el rol admin** (`PATCH /user/{id}/role`); no puede cambiarse a sí mismo |

Los guards (`meta.requiresAuth` / `requiresAdmin` / `guestOnly`) viven en
`src/router/index.js`. Detalles: los precios se muestran en pesos
(`Price / 100`), los platos sin `Portrait` muestran "Sin foto" y el carrito
(`src/order.js`, reactivo) vive en memoria: se arma en el índice con cantidades
y se limpia al confirmar el pedido, con un modal de éxito que se cierra solo a
los 4 segundos.

**Entidades relacionadas por búsqueda asíncrona**: la dirección del checkout y
la categoría del form de platos se eligen tipeando (debounce de 300 ms contra
`?street=` / `?name=` del servidor, con sugerencias y opción de crear la
categoría al toque) — nunca un `<select>` llenado con todos los registros. El
perfil (`/perfil`) permite editar nombre/apellido y subir o quitar el avatar
del usuario, que se muestra en el header. El layout es responsive: en
pantallas chicas el header y los filtros se apilan, la tienda pasa a una sola
columna y las vistas usan una columna centrada (direcciones, platos, perfil)
con listas y modales fluidos.

---

## 10. Matriz de evidencias de requerimientos mínimos

Certificación punto por punto del enunciado, con la evidencia concreta en el
código (rutas del backend y vistas del frontend):

| # | Requerimiento | Cumple | Evidencia |
|---|---|:-:|---|
| 1 | Al menos 6 tablas principales relacionadas; 1:N + N:M (o entidad que dependa de dos) | ✅ | **6 tablas**: `users`, `addresses`, `categories`, `products`, `orders`, `orders_products` ([`script_pg.sql`](backend/cmd/seeder/script_pg.sql), [§3](#3-base-de-datos)). **1:N**: `addresses.user_id`, `orders.client_id`, `orders.address_id`, `products.category_id`. **N:M**: `orders_products` une orders ↔ products (con `ppu` y `amount`) — entidad que depende de dos tablas |
| 2a | Login / Logout funcional | ✅ | `POST /login` + `POST /refresh` + `POST /user` (§8); logout = `clearSession()` en [`frontend/src/auth.js`](frontend/src/auth.js) que limpia los tokens (§9) |
| 2b | Seguridad en el backend y manejo de roles | ✅ | Middlewares [`security.OnlyAdmin` / `AllowClients`](backend/internal/security/middleware.go): 401 sin token, 403 sin rol; matriz completa en [§7](#7-matriz-de-permisos) |
| 2c | Al menos una funcionalidad relevante restringida exclusivamente por rol | ✅ | **Sólo admin** (403 para client): `POST /platos`, `PATCH /platos/{id}`, `POST /categorias`, `GET /ordenes/admin`, `GET /buscar/direcciones`, `GET /usuarios`, `PATCH /user/{id}/role`. Con UI en [`/admin/usuarios`](frontend/src/views/AdminUsersView.vue): el admin asigna o quita el rol admin a otros usuarios (el backend rechaza auto-degradarse con `400`) |
| 2d | Gestión de avatar personalizado en el perfil de usuario | ✅ | Vista [`/perfil`](frontend/src/views/ProfileView.vue): previsualizar, **subir** (`POST /user/{id}/avatar`, multipart, dueño/admin) y **quitar** (`PATCH` con `avatar: null`); el avatar se muestra en el header (§9) |
| 3 | Procesamiento y almacenamiento de archivos/imágenes, adicional al avatar | ✅ | Catálogo de fotos de platos: `picture` en `POST/PATCH /platos` → [`images.Save`](backend/internal/domain/images/imagesService.go) (valida que sea imagen, hash sha256, disco en `PATH_UPLOADS`), servidas por `GET /uploads/{filename}` ([§8 Imágenes](#imágenes)) |
| 4 | ABM/CRUD completo en Vue, todo vía AJAX/JSON sin recargar | ✅ | **Direcciones** (`/direcciones`): alta (`POST`), baja (`DELETE`), modificación (`PATCH`) y listado (`GET`) — [`AddressesView.vue`](frontend/src/views/AddressesView.vue), [§9](#rutas-y-permisos) |
| 5 | Paginado por servidor en todos los listados principales | ✅ | Todos los endpoints de listado devuelven `PageResponse{Data, CurrentPage, TotalPages}` y aceptan `?page&size`: `/platos/list`, `/usuarios/{id}/direcciones`, `/ordenes`, `/ordenes/admin`, `/buscar/direcciones`, `/categorias`, `/usuarios` ([§6](#6-paginación-y-errores)) |
| 6 | Búsqueda asíncrona de entidades relacionadas (prohibido traer todos los registros para un select masivo) | ✅ | Checkout en [`HomeView.vue`](frontend/src/views/HomeView.vue): la dirección se elige tipeando (debounce 300 ms → `GET /usuarios/{id}/direcciones?street=`). Form de platos en [`AdminProductsView.vue`](frontend/src/views/AdminProductsView.vue): la categoría tipeando (→ `GET /categorias?name=` + opción de crear). **No existe ningún `<select>` cargado con todos los registros** |
| 7 | Web API REST con JWT | ✅ | Endpoints autenticados con Bearer JWT (access 15 min + refresh 7 días, [§5](#5-autenticación)); la SPA consume la API nativamente con `apiFetch` ([frontend/src/api.js](frontend/src/api.js)) → no aplica colección de pruebas adjunta |
