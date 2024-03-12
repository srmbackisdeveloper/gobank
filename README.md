# Golang Bank Project
## Used Stack: Gorilla Mux, PostgresSQL, JWT

Bank REST API

- POST /login
- POST /account
- GET /account
- GET /account/:id
- DELETE /account/:id


Database: 
 accounts (
    id SERIAL PRIMARY KEY,
    first_name varchar(50),
    last_name varchar(50),
    number serial,
    balance serial,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


"FantasticFour"
[Bekzhan 21B030933]
[Rauan 22B031182]
[Yerdaulet 22B030389]
