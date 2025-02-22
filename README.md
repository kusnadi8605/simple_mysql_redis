# Tutorial Menjalankan CRUD API dengan Golang, MySQL, dan Redis

## 1. Prerequisites
Pastikan Anda telah menginstal:
- Go (Minimal versi 1.18)
- MySQL
- Redis
- Postman atau cURL untuk testing


## 2. Setup Database
Buat database MySQL dengan nama `simple_mysql_redis` dan jalankan query berikut:
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    age INT NOT NULL
);
```

## 3. Clone Repository
```sh
git clone https://github.com/kusnadi8605/simple_mysql_redis.git
cd simple_mysql_redis
```

## 4. Konfigurasi Environment
Sesuaikan konfigurasi DB dan Redis

## 5. Install Dependencies
```sh
go mod init simple_mysql_redis
go mod tidy
go mod vendor
```

## 6. Jalankan Server
```sh
go run main.go
```
Server akan berjalan di `http://localhost:8089`

## 7. API Endpoints dan cURL Commands

### 1. Create User
**Endpoint:** `POST /users`
```sh
curl -X POST http://localhost:8089/users \
     -H "Content-Type: application/json" \
     -d '{
         "name": "Test",
         "email": "test@example.com",
         "age": 25
     }'
```
**Response:**
```json
{
    "returnCode": "00",
    "returnDesc": "User created successfully",
    "data": {
        "id": 1,
        "name": "Test",
        "email": "test@example.com",
        "age": 25
    }
}
```

### 2. Get All Users
**Endpoint:** `GET /users`
```sh
curl -X GET http://localhost:8089/users
```
**Response:**
```json
{
    "returnCode": "00",
    "returnDesc": "Success",
    "data": [
        {
            "id": 1,
            "name": "Test",
            "email": "test@example.com",
            "age": 25
        }
    ]
}
```

### 3. Get Single User
**Endpoint:** `GET /users/{id}`
```sh
curl -X GET http://localhost:8089/users/1
```
**Response:**
```json
{
    "returnCode": "00",
    "returnDesc": "Success",
    "data": {
        "id": 1,
        "name": "Test",
        "email": "test@example.com",
        "age": 25
    }
}
```

### 4. Update User
**Endpoint:** `PUT /users`
```sh
curl -X PUT http://localhost:8089/users \
     -H "Content-Type: application/json" \
     -d '{
         "id": 1,
         "name": "Test Updated",
         "email": "Test.updated@example.com",
         "age": 26
     }'
```
**Response:**
```json
{
    "returnCode": "00",
    "returnDesc": "User updated successfully",
    "data": {
        "id": 1,
        "name": "Test Updated",
        "email": "Test.updated@example.com",
        "age": 26
    }
}
```

### 5. Delete User
**Endpoint:** `DELETE /users/{id}`
```sh
curl -X DELETE http://localhost:8089/users/1
```
**Response:**
```json
{
    "returnCode": "00",
    "returnDesc": "User deleted successfully"
}
```

## 8. Stop Server
Tekan `CTRL+C` di terminal.

---

Sekarang API CRUD Anda siap digunakan! 🚀
