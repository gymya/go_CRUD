# Go CRUD Backend Service

本專案為 Go 後端 CRUD 服務，涵蓋使用者驗證、產品與分類管理、快取與一致性錯誤處理等核心功能。突顯架構、可維護性、效能與可擴展性。

---

## 專案摘要

- **語言**：Go
- **架構分層**：Delivery / Service / Repository / Domain
- **資料庫**：SQLite
- **快取**：Redis
- **API 文件**：Swagger (OpenAPI)
- **設定管理**：YAML + environment variables

---

## 快速開始

### 需求

- Go 1.20+
- Redis
- SQLite

---

## 專案結構

```
cmd/
  server/              # 服務入口
configs/
  app.yaml             # 應用設定
  config.go            # 設定載入
internal/
  delivery/http/       # HTTP handlers & middlewares
  service/             # 業務邏輯
  repository/SQLite/   # 資料存取層
  domain/              # 實體與錯誤定義
  cache/               # Redis 快取
migrations/            # DB migration
pkg/                   # 可重用元件
```

---

## API 文件

- docs/swagger.json
- docs/swagger.yaml

---

## 面試導向能力展示

### 1) 架構與可維護性

- 明確分層，避免關注點耦合
- Handler 專注於傳輸層與輸入驗證
- Service 負責商業規則與流程控管
- Repository 隔離資料來源，便於替換與測試

### 2) 錯誤處理一致性

- 集中式錯誤定義與轉譯
- 可區分業務錯誤與系統錯誤
- 中介層統一輸出錯誤回應格式

### 3) 效能與快取策略

- Redis 快取降低熱點查詢成本
- 快取與資料層解耦，便於替換與測試

### 4) 安全與存取控制

- JWT 驗證中介層
- 可擴展為 RBAC 或權限分級

### 5) 可擴展性與可移植性

- SQLite 方便本機與 CI 開發
- Repository 介面可替換為 MySQL / PostgreSQL
- 可擴充 gRPC / GraphQL 或事件驅動架構


## 專案說明

> 這個專案的設計重點是分層與邊界清晰。Service 將業務規則集中管理，Repository 負責資料存取，降低跨層耦合。透過統一錯誤處理與快取策略，提升穩定性與效能，同時為後續擴展（例如替換資料庫或加入 gRPC）預留清楚的擴充點。