# Casino Wallet Service

## Kullanılan Teknolojiler
- Go 1.24.2
- PostgreSQL (Veritabanı)
- Docker & Docker Compose
- GORM (ORM)
- Gorilla Mux (HTTP Router)
- Zap (Loglama)
- Swagger/OpenAPI (API Dokümantasyonu)
- Migrate (Veritabanı Migrasyonları)
- Mockery (Mock Testing Framework)

## Yerel Ortamda Çalıştırma

1. Gerekli araçların kurulumu:
   - Docker ve Docker Compose
   - Go 1.24.2 veya üzeri
   - Make (opsiyonel, Makefile komutları için)

2. Projeyi klonlayın:
   ```bash
   git clone https://github.com/BarisKilicGsu/casino-wallet-service.git
   cd casino-wallet-service
   ```

3. Docker ile çalıştırma:
   ```bash
   docker-compose up --build -d
   ```

4. Uygulama http://localhost:8080 adresinde çalışmaya başlayacaktır.

5. PostgreSQL'e bağlanmak için: postgresql://postgres:postgres@localhost/casino_wallet?statusColor=686B6F&env=local&name=localhost

6. Veritabanı volümü projedeki volumes/postgres klasörüne bağlıdır. Bu klasörü silerseniz tüm veritabanı temizlenmiş olur. 

Not: Veritabanında 30 örnek kullanıcı bulunmaktadır. Tüm kullanıcıları çekme endpointi test amaçlı istendiği için pagination yapısı eklenmemiştir.

## Örnek İstekler için Curl

### Oyuncu Bakiyesi Sorgulama
```bash
# player1 için bakiye sorgulama
curl -X GET "http://localhost:8080/wallet/player1"

# player2 için bakiye sorgulama
curl -X GET "http://localhost:8080/wallet/player2"
```

### Tüm Oyuncuları Listeleme
```bash
# Tüm oyuncuları listele
curl -X GET "http://localhost:8080/players"
```

### Bahis İşlemi (Bet)
```bash
# player1 için 100 INR'lik bahis
curl -X POST "http://localhost:8080/event" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "currency": "INR",
    "game_code": "ntn_aloha",
    "player_id": "player1",
    "wallet_id": "wallet1",
    "req_id": "bet-001",
    "round_id": "round-001",
    "session_id": "session-001",
    "type": "bet"
  }'
```

### Kazanç İşlemi (Result)
```bash
# player1 için 150 INR'lik kazanç
curl -X POST "http://localhost:8080/event" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 150,
    "currency": "INR",
    "game_code": "ntn_aloha",
    "player_id": "player1",
    "wallet_id": "wallet1",
    "req_id": "result-001",
    "round_id": "round-001",
    "session_id": "session-001",
    "type": "result"
  }'
```



## Tasarım Notları ve Varsayımlar

### Mimari Yapı
- API dokümantasyonu Swagger/OpenAPI ile oluşturulmuştur
- Veritabanı migrasyonları için migrate tool'u kullanılmıştır
- Testler için kullanılması için servis ve repo interfacleri Mockery ile mocklanmıştır
- Servis, mikroservis mimarisi prensiplerine uygun olarak tasarlanmıştır
- Clean Architecture prensipleri uygulanmıştır (Repository, Service, Handler katmanları)
- PostgreSQL veritabanı kullanılarak veri kalıcılığı sağlanmıştır
- Docker container'ları ile izolasyon ve kolay deployment sağlanmıştır

### İş Mantığı ve Validasyonlar
- Her işlem (bet/result) için benzersiz `req_id` kontrolü yapılmaktadır. Her zaman req_id unique olmalıdır.
- İstekler dokümanda belirtildiği gibi gelmelidir, eğer amount result type için yoksa 0 olarak gönderilmelidir.
- Result işlemleri için ilgili bet işleminin varlığı kontrol edilmektedir
- Bet ve result işlemleri arasında tutarlılık kontrolleri:
  - Round ID, Player ID ve Wallet ID kombinasyonu kontrolü (bir round, kullanıcı ve wallet üçlüsü için sadece bir bet ve bir result işlemi yapılabilir)
  - Game code eşleşmesi
  - Wallet ID eşleşmesi
  - Player ID eşleşmesi
  - Round ID eşleşmesi
- Bakiye kontrolleri:
  - Bet işlemlerinde yeterli bakiye kontrolü

### Veritabanı İzolasyon ve Kilitleme Stratejisi
- GORM repository katmanında transaction yönetimi için özel bir implementasyon bulunmaktadır
- Her kritik işlem için SELECT FOR UPDATE ile row-level locking kullanılmaktadır:
  - Player bakiyesi güncellenirken
  - Transaction kayıtları kontrol edilirken
  - Round ID, Type ve Wallet ID kombinasyonu bazlı işlemlerde
- Bu sayede:
  - Aynı oyuncuya ait eşzamanlı işlemler sıralı olarak işlenir
  - Aynı round ID'ye ait işlemler çakışmaz
  - Aynı request ID'ye sahip işlemler tekrar işlenmez
  - Aynı round ID, player ID ve wallet ID kombinasyonuna sahip işlemler çakışmaz
- Repository katmanında transaction yönetimi için:
  - `StartTransaction()`: Yeni transaction başlatma
  - `StartTransactionWithIsolation()`: İzolasyon seviyesi belirterek transaction başlatma
  - `FinishTransaction()`: Transaction'ı commit veya rollback etme
  - `RollbackTransaction()`: Transaction'ı rollback etme
  - `CommitTransaction()`: Transaction'ı commit etme

### Audit Log ve Transaction Yönetimi
- Her işlem (bet/result) için transaction kaydı db'de tutulmaktadır
- Transaction kayıtları user balance ile birlikte atomik olarak işlenmektedir
- Herhangi bir hata durumunda hem transaction hem de kullanıcı bakiyesi rollback edilmektedir



## Potansiyel İyileştirmeler

1. Performans İyileştirmeleri:
   - Redis cache implementasyonu
   - Veritabanı indeksleme optimizasyonları düzeltilebilir
   - Connection pooling eklenerek db connectionları optimize edilebilir
   - Bu kadar fazla veritabanı kilitleme sistemi performansı yavaşlatabilir. Bunu azaltmak için event handler'da transactionların sadece geçerli olduğunu kontrol edip veritabanına tekrar etmeden ekleyebilir. Event sourcing mantığı ile çalışabilir. Kullanıcı bakiyesi çekildiğinde Redis cache'den bakiye çekilebilir. Yeni transaction geldiğinde başka bir mekanizma buradaki transaction miktarlarını alıp Redis'te bakiye güncelleyebilir.
    - Wallet ve kullanıcı birbirinden ayrılarak yeni ilişkisel yapı oluşturulabilir
    - Performansı izlemek için repository işlemlerine parametre eklenerek Prometheus ile bunlar toplanabilir

    



## Ekstra Notlar 

- Swagger dosyası: docs/api.yaml içinde bulunmaktadır
- Veritabanını doldurmak için, eğer kullanıcı yoksa internal/seed/seed.go main tarafından çalıştırılır
- Migrate aracı Docker ayağa kalktığında, eğer eksik çalışmamış migration varsa migrations/* altındaki dosyaları çalıştırır
- Test ve işlemler kolay olsun diye .env kullanılmamıştır. Env değişkenleri docker compose içinden çekilmektedir. 