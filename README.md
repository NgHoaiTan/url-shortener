# URL Shortener Service


## 📋 Mục lục

- [Mô tả bài toán](#-mô-tả-bài-toán)
- [Cách chạy project](#-cách-chạy-project)
- [Thiết kế & Quyết định kỹ thuật](#️-thiết-kế--quyết-định-kỹ-thuật)
- [Trade-offs](#️-trade-offs)
- [Challenges](#-challenges)
- [Limitations & Improvements](#-limitations--improvements)
- [API Documentation](#-api-documentation)
- [Architecture](#️-architecture)
- [Tech Stack](#️-tech-stack)

---

## 🎯 Mô tả bài toán


**Vấn đề cốt lõi:** Biến đổi một URL dài thành một URL ngắn gọn, khi truy cập vào URL vừa rút ngắn thì sẽ tự động redirect về URL gốc, đồng thời tracking được số lượt truy cập.

**Chức năng:**
1. **Tạo short URL** từ URL gốc
2. **Redirect** từ short URL về URL gốc
3. **Tracking** số lượt click
4. **Quản lý** danh sách URLs với pagination, filtering, sorting
5. **Validation** để tránh URL không hợp lệ hoặc self-shortening

---

## 🚀 Cách chạy project

### Yêu cầu

**Go 1.25+**, **PostgreSQL 16+**

---

#### Bước 1: Cài đặt dependencies

**PostgreSQL:**
```bash
# Windows: Download từ https://www.postgresql.org/download/
# Mac: brew install postgresql
# Linux: sudo apt-get install postgresql
```

#### Bước 2: Tạo database
```sql
CREATE DATABASE urlshortener;
```

#### Bước 3: Cấu hình .env
```env
DB_HOST=localhost
DB_USER=
DB_PASSWORD=
DB_NAME=urlshortener
DB_PORT=
DB_SSLMODE=disable
DB_TIMEZONE=UTC

BASE_URL=http://localhost:8080
SERVICE_DOMAIN=localhost:8080
APP_PORT=8080
```

#### Bước 4: Install Go dependencies
```bash
go mod download
```

#### Bước 5: Chạy application
```bash
go run main.go
```

#### Bước 6: Test
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://google.com"}'
```
---


## 🏗️ Thiết kế & Quyết định kỹ thuật

### 1. Tại sao chọn PostgreSQL?

**Lý do chọn PostgreSQL:**

- PostgreSQL hỗ trợ UNIQUE constraint, đảm bảo 1 original URL chỉ sinh ra 1 short URL ngay cả khi có nhiều request đồng thời
- PostgreSQL tuân thủ ACID, nên tránh được lỗi race condition khi insert hoặc cập nhật số lượt click.
- Hỗ trợ indexing tốt, giúp tra cứu nhanh short code và original URL.

### 2. RESTful Design:

```
POST   /api/shorten              # Tạo short URL
GET    /api/urls                 # List URLs (với pagination/filter)
GET    /api/urls/:shortCode      # Xem thông tin URL
GET    /:shortCode               # Redirect (không có /api prefix)
```
**Lý do chọn RESTful:**
- Dễ hiểu, dễ sử dụng
- Tách rõ hành động (HTTP method) và tài nguyên (URL)
- Dễ dàng tích hợp với frontend

**RESTful conventions:**
- POST cho create
- GET cho read operations
- HTTP status codes chuẩn (201, 404, 400, 500)

**Pagination built-in:**
```
GET /api/urls?page=1&limit=10&sort_by=click_count&order=desc&keyword=google
```
- Không return toàn bộ data → tránh overload
- Client control được sorting và filtering

### 3. Thuật toán generate mã ngắn

**Implementation trong `utils/codec.go`:**

```go
const base62Chars = "QW8eRTYUIOPmNcpyVtBoSrEwixL5X1M3n6b9DAuvqC7z0Za2Ksd4JfgHhjGklF"
const shortCodeLength = 8

func GenerateShortCode(length int) (string, error) {
    b := make([]byte, length)
	
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
	
    for i := 0; i < length; i++ {
        b[i] = base62Chars[int(b[i]) % len(base62Chars)]
    }
    
    return string(b), nil
}
```

**Có nhiều cách để thực hiện tạo short code, mỗi cách có ưu nhược điểm riêng:**

---

#### Option 1: Hash Functions (MD5, SHA256)

**MD5:**

**Nhược điểm:**
- **Quá dài:** 32 ký tự (hex) hoặc 22 ký tự (base64) - vẫn quá dài cho URL ngắn
- **Collision-prone:** MD5 không collision-resistant, dễ sinh ra cùng hash cho URLs khác nhau
- **Cần truncate:** Cắt ngắn hash → tăng collision rate đáng kể
- **Deterministic:** Cùng input → cùng output → dễ đoán và không có tính bảo mật
- **Security concern:** MD5 đã bị deprecated vì các lỗ hổng bảo mật

**SHA256:**
**Nhược điểm:**
- **Cực kỳ dài:** 64 ký tự (hex) hoặc 43 ký tự (base64)
- **Overkill:** Security features của SHA256 không cần thiết cho URL shortener
- **Performance:** Chậm hơn random generation
- **Vẫn cần truncate:** Cắt xuống 8 chars → mất đi collision resistance

**Vấn đề chung với Hash Functions:**
```
Original hash:  c984d06aafbecf6bc55569f964148ea3
Truncate to 8:  c984d06a

Collision probability tăng từ 2^-128 lên 2^-32
→ Collision sau ~65,000 URLs (Nghịch lý sinh nhật)
```

---

#### Option 2: UUID (Universally Unique Identifier)

**UUIDv4:**

**Nhược điểm:**
- **Quá dài:** 36 ký tự (có dashes) hoặc 32 ký tự (không có dashes)
- **Không URL-friendly:** Chứa dashes và lowercase hex only
- **Overkill:** 122 bits of randomness là quá nhiều cho nhu cầu
- **Không thể customize length:** UUID có format cố định
- **Phí phạm entropy:** Sử dụng nhiều random bits hơn cần thiết

**Truncate UUID:**
```
UUID:           f47ac10b-58cc-4372-a567-0e02b2c3d479
Take first 8:   f47ac10b

→ Mất đi uniqueness guarantee của UUID
→ Không khác gì random 8 chars
```

---

#### Option 3: Sequential ID + Base62 Encoding

**Nhược điểm:**
- **Dễ đoán:** User có thể enumerate tất cả URLs
  ```
  /1 → first URL
  /2 → second URL
  /3 → third URL
  → Scrape toàn bộ database
  ```
- **No privacy:** Biết được tổng số URLs trong hệ thống
- **Scalability issue:** Cần centralized counter → bottleneck khi scale

---

#### Option 4: Random Base62 (Cách đã chọn)

**Tại sao chọn cách này:**

**1. Độ dài tối ưu:**
```
8 chars base62 = 62^8 = 218,340,105,584,896 combinations (~218 trillion)
```
- Ngắn gọn: Chỉ 8 ký tự
- Đủ lớn: Phục vụ hàng triệu URLs mà collision risk thấp
- URL-friendly: Chỉ dùng [a-zA-Z0-9]

**2. Collision resistance:**
```
Birthday Paradox formula:
P(collision) ≈ n^2 / (2 * N)

Với N = 62^8 và n = 1,000,000 URLs:
P(collision) ≈ 1,000,000^2 / (2 * 62^8) ≈ 0.23% 

→ Sau 1 triệu URLs, chỉ có 0.23% khả năng collision
```

**3. Security & Privacy:**
```
Không đoán được: Random generation
Không enumerate được: Không có pattern
Không leak thông tin: Không biết tổng số URLs
```

**4. Flexibility:**
- Có thể tăng length nếu cần
- Balance giữa ngắn gọn và collision risk


### 4. Xử lý conflict/duplicate

**Strategy: Optimistic Locking + Retry với Fallback**

#### Case 1: Race Condition khi generate cùng short code

**Vấn đề:**

Khi có 2 request đồng thời cố gắng tạo short URL và ngẫu nhiên generate ra cùng một mã code, race condition xảy ra như sau:

**Timeline của sự cố:**

```
T1: Request A bắt đầu
    → Gọi hàm GenerateShortCode()
    → Nhận được kết quả: "abc123"
    → Kiểm tra trong database xem "abc123" đã tồn tại chưa
    → Database trả về: KHÔNG TỒN TẠI
    
T2: Request B bắt đầu (cùng lúc với Request A)
    → Gọi hàm GenerateShortCode()
    → May mắn (hoặc không may) cũng nhận được: "abc123"
    → Kiểm tra trong database xem "abc123" đã tồn tại chưa
    → Database trả về: KHÔNG TỒN TẠI (vì Request A chưa kịp insert)
    
T3: Request A tiếp tục
    → Chuẩn bị dữ liệu để insert vào database
    → Bắt đầu transaction INSERT với short_code = "abc123"
    
T4: Request B cũng tiếp tục (gần như cùng lúc)
    → Cũng chuẩn bị dữ liệu để insert
    → Cũng bắt đầu transaction INSERT với short_code = "abc123"
    
T5: CẢ HAI cùng gửi INSERT query xuống database
    → Request A: INSERT INTO short_urls (short_code, ...) VALUES ('abc123', ...)
    → Request B: INSERT INTO short_urls (short_code, ...) VALUES ('abc123', ...)
    
T6: Database nhận 2 requests
    → Request nào đến trước (giả sử A) → INSERT thành công
    → Request còn lại (B) → DUPLICATE KEY ERROR
    → Request B fail và trả về lỗi cho client
```

**Tại sao kiểm tra trước không giúp được gì?**

Nhiều người nghĩ: "Chỉ cần check trước khi insert là được rồi". Nhưng không phải!

**Vấn đề của "Check-Then-Act" pattern:**
- Giữa thời điểm CHECK (kiểm tra) và ACT (insert) có một khoảng thời gian
- Trong khoảng thời gian đó, trạng thái database có thể thay đổi
- Điều này gọi là **Time-of-Check-to-Time-of-Use (TOCTOU) race condition**

```
Request A check → Kết quả: không tồn tại ✓
Request B check → Kết quả: không tồn tại ✓ (vì A chưa insert)
Request A insert → Thành công ✓
Request B insert → FAIL! (vì A đã insert rồi)
```

**Giải pháp đã áp dụng:**

Thay vì cố gắng ngăn chặn race condition, chúng ta **chấp nhận nó có thể xảy ra** và xử lý một cách graceful:

**1. Database-level Uniqueness Constraint:**
- Đặt UNIQUE INDEX trên cột `short_code` trong database
- Database đảm bảo tính duy nhất ở mức atomic (không thể bị race)
- Đây là "single source of truth" duy nhất
- Khi có duplicate insert → database tự động reject và báo lỗi

**2. Optimistic Locking + Retry Mechanism:**
- **Optimistic** = "hy vọng" không có collision, cứ thử insert trước
- Nếu gặp DUPLICATE KEY ERROR → đó là dấu hiệu collision xảy ra
- **Retry** = generate mã mới và thử lại
- Lặp lại tối đa 5 lần

**Flow xử lý:**
```
Lần 1: Generate "abc123" → Insert → DUPLICATE ERROR → Retry
Lần 2: Generate "xyz789" → Insert → DUPLICATE ERROR → Retry
Lần 3: Generate "def456" → Insert → SUCCESS ✓
→ Trả về "def456" cho client
```

**3. Idempotency cho cùng original_url:**
- Trước khi generate mã mới, check xem URL gốc đã tồn tại chưa
- Nếu URL đã tồn tại → trả về mã cũ (không tạo duplicate)
- Đảm bảo: 1 URL gốc → 1 short code duy nhất


#### Case 2: Concurrent requests cho cùng original_url

**Vấn đề:**

Tình huống khác xảy ra khi 2 người dùng khác nhau (hoặc cùng 1 người) cùng lúc cố gắng tạo short URL cho cùng một URL gốc.

**Ví dụ:** Cả User A và User B cùng muốn shorten "https://google.com"

**Timeline của sự cố:**

```
T1: User A gửi request
    → POST /shorten với original_url = "https://google.com"
    → Service kiểm tra xem "https://google.com" đã tồn tại chưa
    → Database trả về: KHÔNG TỒN TẠI
    
T2: User B gửi request (gần như cùng lúc)
    → POST /shorten với original_url = "https://google.com"
    → Service kiểm tra xem "https://google.com" đã tồn tại chưa
    → Database trả về: KHÔNG TỒN TẠI (vì User A chưa kịp insert)
    
T3: User A tiếp tục
    → Generate short code: "abc123"
    → Chuẩn bị insert: ("https://google.com", "abc123")
    
T4: User B cũng tiếp tục
    → Generate short code: "xyz789" (khác với User A)
    → Chuẩn bị insert: ("https://google.com", "xyz789")
    
T5: Cả hai cùng insert vào database
    → User A: INSERT (original_url="https://google.com", short_code="abc123")
    → User B: INSERT (original_url="https://google.com", short_code="xyz789")
    
T6: Database phát hiện conflict
    → Request User A đến trước → INSERT thành công ✓
    → Request User B → DUPLICATE KEY ERROR trên cột original_url
```

**Hậu quả nếu không xử lý:**
- User B nhận lỗi "failed to create short URL"
- User B phải thử lại
- Trải nghiệm người dùng kém
- Có thể tạo ra nhiều short codes cho cùng 1 URL (nếu không có unique constraint)

**Giải pháp đã áp dụng:**

**1. Unique Constraint trên original_url:**
- Database có UNIQUE INDEX trên cột `original_url`
- Đảm bảo 1 URL gốc chỉ có 1 record duy nhất
- Database reject mọi attempt để insert duplicate original_url

**2. Graceful Handling:**

Khi gặp DUPLICATE KEY ERROR trên `original_url`, service không trả lỗi mà:

```
Bước 1: Phát hiện lỗi là do duplicate original_url (không phải short_code)
Bước 2: Query database để lấy record đã tồn tại
        → SELECT * FROM short_urls WHERE original_url = "https://google.com"
Bước 3: Lấy short_code từ record đã tồn tại
        → short_code = "abc123" (của User A)
Bước 4: Trả về short_code này cho User B
        → User B nhận "abc123" (giống User A)
```

**Kết quả cuối cùng:**
- User A nhận: short_code = "abc123"
- User B cũng nhận: short_code = "abc123"
- **Cả hai đều thành công, không có lỗi**
- Cùng URL gốc → cùng short code (idempotent)

Giải pháp này tối ưu hơn lock URL trước khi insert vì:

```
Cách lock phức tạp:
1. Acquire lock trên URL "https://google.com"
2. Check database
3. Insert nếu chưa tồn tại
4. Release lock
```

**Nhược điểm:**
- Tăng complexity và dependencies
- Performance bottleneck (serialized requests)
- Lock timeout và deadlock issues

---

## ⚖️ Trade-offs

### 1. Sync vs Async Click Tracking

**Đã chọn: Async (Goroutine)**

```go
func (s *urlService) GetOriginalURL(shortCode string) (string, error) {
    url, err := s.repo.FindByShortCode(shortCode)
    if err != nil {
        return "", errors.New("short URL not found")
    }
    
    //  Async increment - không block redirect
    go func() {
        if err := s.repo.IncrementClickCount(shortCode); err != nil {
            log.Printf("failed to increment click count: %v", err)
        }
    }()
    
    return url.OriginalURL, nil
}
```

**Trade-off:**

**Lý do chọn Async:**
- User experience > perfect accuracy
- Redirect < 50ms quan trọng hơn là track mọi click
- Click count là metric, không phải critical business data

**Nhược điểm:**
- Nếu server crash ngay sau redirect, click có thể không được count

**Cách cải thiện (future):**
- Dùng message queue để đảm bảo không mất data

---

### 2. Validation: Whitelist vs Blacklist

**Đã chọn: Whitelist (chỉ cho phép http/https)**


**Lý do:**
- **Security:** Ngăn chặn `javascript:`, `file:`, `data:` schemes
- **Simple:** Chỉ cần check 2 schemes
- **Safe default:** Reject unknown schemes

**Nhược điểm:**
- Không hỗ trợ `ftp://` hay custom schemes
---

## 🔥 Challenges

### Challenge 1: Race Condition khi tạo Short Code

**Vấn đề:**
- 2 requests cùng lúc generate cùng short code
- Cả 2 check DB → không tồn tại → cả 2 insert → conflict

**Giải quyết:**
1. **Database unique constraint** làm safety net
2. **Retry mechanism** với max 5 attempts
3. **Double-check** original_url sau khi catch duplicate error

**Học được:**
- Database constraints là tuyến phòng thủ cuối cùng
- Optimistic locking + retry tốt hơn pessimistic locking
- Luôn có chiến lược fallback
---

### Challenge 4: Click Tracking Performance

**Vấn đề:**
- Redirect phải nhanh (< 100ms)
- Tracking click count cần DB write
- Không muốn user chờ tracking complete

**Giải quyết:**
- Async tracking bằng goroutine
- Redirect ngay sau khi lookup

**Trade-off:**
- Có thể miss một số clicks nếu server crash
- Nhưng user experience tốt hơn

**Học được:**
- Không phải mọi operation đều cần sync
- Có thể improve sau bằng message queue

---

## 🚧 Limitations & Improvements

### Limitations hiện tại

#### 1. **Không có Authentication/Authorization**
**Vấn đề:**
- Bất kỳ ai cũng có thể tạo URLs
- Không track được URLs của user nào
- Không có quota/rate limiting

**Impact:** 
- Dễ bị spam
- Không phù hợp production public service

---

#### 3. **No Caching Layer**
**Vấn đề:**
- Mọi redirect đều hit database
- Popular URLs bị query nhiều lần

**Impact:**
- Database bottleneck với high traffic
- Response time chậm hơn mức cần thiết

---

#### 5. **No Analytics/Metrics**
❌ **Vấn đề:**
- Chỉ có click count
- Không track: user agent, referrer, geo location, timestamp

**Impact:**
- Không có insights về traffic
- Không thể optimize

---

#### 6. **No URL Expiration**
❌ **Vấn đề:**
- URLs tồn tại mãi mãi

**Impact:**
- Database grow vô hạn
- Temporary URLs không có cách cleanup

---

### Nếu có thêm thời gian sẽ làm gì?

#### 1. **Redis Caching Layer**

**Benefits:**
- Response time: 50ms → 5ms
- Giảm 90% database load cho popular URLs

---

#### 2. **Authentication & User Management**

**Features:**
- JWT-based authentication
- User registration/login
- URLs belong to users

**Benefits:**
- User có thể manage URLs của mình
- Có thể implement quota per user
- Security tốt hơn

---

#### 3. **Rate Limiting**

**Benefits:**
- Ngăn chặn abuse/spam
- Protect server khỏi overload

---

#### 4. **Custom Short Codes**

**Benefits:**
- Brand-friendly URLs
- Memorable short codes

---

#### 6. **URL Expiration**

**Benefits:**
- Temporary campaigns
- Database không grow vô hạn

---

### Production-ready cần thêm gì?

Để đưa dự án này lên production cần bổ sung các thành phần sau:

---

#### 1. **Security**

---

#### 2. **Monitoring & Logging**

---

#### 3. **Performance Optimization**

---

## 📚 API Documentation

### Base URL
```
http://localhost:8080
```

### Endpoints

#### 1. Create Short URL

**Request:**
```http
POST /api/shorten
Content-Type: application/json

{
  "original_url": "https://example.com/very/long/url"
}
```

**Response: 201 Created**
```json
{
  "id": 1,
  "short_code": "abc12345",
  "original_url": "https://example.com/very/long/url",
  "short_url": "http://localhost:8080/abc12345",
  "click_count": 0,
  "created_at": "2025-12-17T10:00:00Z"
}
```

**Errors:**
```json
// 400 Bad Request - Invalid URL
{
  "error": "URL validation failed",
  "message": "invalid URL format"
}

// 400 Bad Request - Self shortening
{
  "error": "URL validation failed",
  "message": "cannot create short URL for this domain"
}
```

---

#### 2. Redirect to Original URL

**Request:**
```http
GET /:shortCode
```

**Response: 302 Found**
```
Location: https://example.com/very/long/url
```

**Errors:**
```json
// 404 Not Found
{
  "error": "Short URL not found",
  "message": "short URL not found"
}
```

---

#### 3. Get URL Info

**Request:**
```http
GET /api/urls/:shortCode
```

**Response: 200 OK**
```json
{
  "id": 1,
  "short_code": "abc12345",
  "original_url": "https://example.com/very/long/url",
  "short_url": "http://localhost:8080/abc12345",
  "click_count": 42,
  "created_at": "2025-12-17T10:00:00Z",
  "updated_at": "2025-12-17T11:00:00Z"
}
```

---

#### 4. List URLs

**Request:**
```http
GET /api/urls?page=1&page_size=10&sort_by=click_count&order=desc&search=google
```

**Query Parameters:**
- `page` (optional): Page number, default = 1
- `page_size` (optional): Items per page, default = 10, max = 100
- `sort_by` (optional): Sort field (`created_at`, `updated_at`, `click_count`), default = `created_at`
- `order` (optional): Sort order (`asc`, `desc`), default = `desc`
- `search` (optional): Search keyword in original URL

**Response: 200 OK**
```json
{
  "urls": [
    {
      "id": 1,
      "short_code": "abc12345",
      "original_url": "https://google.com",
      "short_url": "http://localhost:8080/abc12345",
      "click_count": 42,
      "created_at": "2025-12-17T10:00:00Z",
      "updated_at": "2025-12-17T11:00:00Z"
    }
  ],
  "total_count": 100,
  "page": 1,
  "page_size": 10,
  "total_pages": 10,
  "sort_by": "click_count",
  "order": "desc",
  "search": "google"
}
```

---

## 🏛️ Architecture

### Project Structure
```
URL-Shortener-Service/
├── config/
│   ├── database.go      # Database connection & migrations
│   └── env.go           # Environment loader
├── controllers/
│   └── url_controller.go # HTTP handlers (request/response)
├── dtos/
│   └── url_dto.go       # Data Transfer Objects
├── models/
│   └── url.go           # Database models (GORM)
├── repositories/
│   └── url_repository.go # Database operations (abstraction)
├── routes/
│   └── routes.go        # Route definitions
├── services/
│   └── url_service.go   # Business logic
├── utils/
│   ├── codec.go         # Short code generator
│   └── url_validator.go # URL validation
├── main.go              # Application entry point
└── .env                 # Configuration
```
---

## 🛠️ Tech Stack

**Backend:**
- **Go 1.25** - Main programming language
- **Gin** - Web framework
- **GORM** - ORM for database operations

**Database:**
- **PostgreSQL 18** - Primary database

