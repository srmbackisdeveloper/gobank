# Golang Bank Project
## Used Stack: Gorilla Mux, PostgresSQL, JWT

Restaurant REST API

- POST /login
- POST /account
- GET /account
- GET /account/:id
- DELETE /account/:id


Table restaurants {
    id bigserial [primary key]
    created_at timestamp
    updated_at timestamp
    title text
    coordinates text
    address text
    cousine text
}

Table menu {
    id bigserial [primary key]
    created_at timestamp
    updated_at timestamp
    title text
    description text
    nutrition_value text
}

// many-to-many
Table restaurants_and_menu {
    id bigserial [primary key]
    created_at timestamp
    updated_at timestamp
    restaurant bigserial
    menu bigserial
}