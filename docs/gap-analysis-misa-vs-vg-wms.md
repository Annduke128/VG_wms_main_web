# BÁO CÁO PHÂN TÍCH KHOẢNG CÁCH: MISA AMIS vs. VG WMS

> **Ngày tạo:** 2026-04-09  
> **Nguồn dữ liệu:**
>
> - Marketing page amis.misa.vn/amis-kho-hang/
> - Helpamis AMIS Mua hàng (28+ articles)
> - Academy AMIS Kế toán Phân hệ Kho (14 bài)
> - Academy SME.NET Kho (23 bài)
> - Source code analysis VG WMS (6 migrations, services, routes, domain types)
>
> **Lưu ý:** helpact.amis.vn (AMIS Kế toán chi tiết) đang bảo trì, chưa crawl được. Báo cáo có thể thiếu chi tiết về hạch toán kế toán kho.

---

## Tóm tắt

| Chỉ số                       | Giá trị |
| ---------------------------- | ------- |
| Tổng nghiệp vụ MISA đánh giá | ~50+    |
| VG WMS đáp ứng đầy đủ        | ~15-20% |
| Gap Critical                 | 4       |
| Gap High                     | 6       |
| Gap Medium                   | 7       |
| Gap Low                      | 7       |
| Điểm mạnh riêng VG WMS       | 6       |

---

## I. NGHIỆP VỤ NHẬP KHO

### MISA hỗ trợ 8+ loại nhập kho:

| #   | Nghiệp vụ MISA                         | VG WMS                                                                           | Gap                                    |
| --- | -------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------- |
| 1   | Nhập mua hàng hóa (qua kho)            | **Partial** — có inbound_items nhưng không link PO                               | Thiếu PO workflow                      |
| 2   | Nhập hàng mua đang đi đường            | **NO**                                                                           | Thiếu in-transit tracking              |
| 3   | Nhập thành phẩm sản xuất               | **Partial** — combo system tạo combo inventory, nhưng không phải manufacturing   | Thiếu production module                |
| 4   | Nhập hàng bán bị trả lại               | **NO** — chỉ có field `so_luong_tra_lai` trên items, không có quy trình trả hàng | Thiếu returns workflow                 |
| 5   | Nhập NVL sản xuất không dùng hết       | **NO**                                                                           | Thiếu production returns               |
| 6   | Nhập từ chi nhánh phụ thuộc chuyển đến | **NO** — single warehouse                                                        | Thiếu multi-warehouse + stock transfer |
| 7   | Nhập hàng gia công nhận lại            | **NO**                                                                           | Thiếu outsourcing module               |
| 8   | Nhập nhận bán hộ/ký gửi/ký cược        | **NO**                                                                           | Thiếu consignment                      |
| 9   | Nhập nhận giữ hộ/gia công              | **NO**                                                                           | Thiếu custody/outsourcing              |

**VG WMS hiện có:** `CreateInbound` → insert inbound_item + upsert lot + update inventory_main + movement + recalc. Đây là quy trình generic, không phân biệt loại nhập.

---

## II. NGHIỆP VỤ XUẤT KHO

### MISA hỗ trợ 7+ loại xuất kho:

| #   | Nghiệp vụ MISA                        | VG WMS                                       | Gap                          |
| --- | ------------------------------------- | -------------------------------------------- | ---------------------------- |
| 1   | Xuất NVL cho sản xuất                 | **Partial** — combo system xuất NVL qua FIFO | Chỉ cho combo, không general |
| 2   | Xuất bán hàng                         | **YES** — CreateOutbound với FIFO allocation | OK nhưng thiếu SO workflow   |
| 3   | Xuất trả lại hàng mua (qua kho)       | **NO**                                       | Thiếu purchase returns       |
| 4   | Xuất trả lại hàng mua (không qua kho) | **NO**                                       | Thiếu purchase returns       |
| 5   | Xuất giảm giá hàng mua (qua kho)      | **NO**                                       | Thiếu pricing adjustments    |
| 6   | Xuất NVL cho XDCB/sửa chữa TSCĐ       | **NO**                                       | Thiếu asset management       |
| 7   | Xuất NVL/HH góp vốn                   | **NO**                                       | Thiếu capital contribution   |
| 8   | Xuất biếu tặng/tiêu dùng nội bộ       | **NO**                                       | Thiếu internal consumption   |
| 9   | Xuất cho chi nhánh phụ thuộc          | **NO**                                       | Thiếu multi-warehouse        |

**VG WMS hiện có:** `CreateOutbound` → FIFO allocation từ lots → proportional split → update inventory. Generic, 1 loại xuất.

---

## III. NGHIỆP VỤ CHUYỂN KHO

| #   | Nghiệp vụ MISA                   | VG WMS | Gap                              |
| --- | -------------------------------- | ------ | -------------------------------- |
| 1   | Chuyển kho nội bộ (giữa các kho) | **NO** | Thiếu multi-warehouse + transfer |
| 2   | Chuyển hàng gửi bán đại lý       | **NO** | Thiếu consignment sales          |

**VG WMS:** Không có khái niệm warehouse_id. Toàn bộ hệ thống là single warehouse.

---

## IV. KIỂM KÊ KHO

| #   | Nghiệp vụ MISA                 | VG WMS | Gap             |
| --- | ------------------------------ | ------ | --------------- |
| 1   | Kiểm kê kho (stock count)      | **NO** | Thiếu hoàn toàn |
| 2   | So sánh tồn sổ sách vs thực tế | **NO** | Thiếu hoàn toàn |
| 3   | Xử lý chênh lệch (thừa/thiếu)  | **NO** | Thiếu hoàn toàn |

---

## V. LẮP RÁP / THÁO DỠ

| #   | Nghiệp vụ MISA            | VG WMS                                                                      | Gap         |
| --- | ------------------------- | --------------------------------------------------------------------------- | ----------- |
| 1   | Xuất VT lắp ráp + nhập TP | **YES** — ComboService.CreateCombo (FIFO deduct semi + accessories → combo) | Tương đương |
| 2   | Xuất HH tháo dỡ + nhập VT | **YES** — ComboService.CancelCombo (reverse all components back)            | Tương đương |

**Đây là điểm mạnh của VG WMS** — combo system với BOM 2 cấp (semi-finished + accessories), FIFO lot tracking trên component movements.

---

## VI. TÍNH GIÁ XUẤT KHO

| #   | Nghiệp vụ MISA                 | VG WMS                                                     | Gap                      |
| --- | ------------------------------ | ---------------------------------------------------------- | ------------------------ |
| 1   | Bình quân cuối kỳ              | **NO**                                                     | Thiếu                    |
| 2   | Bình quân tức thời             | **NO**                                                     | Thiếu                    |
| 3   | FIFO costing                   | **NO** — có FIFO allocation nhưng không track giá theo lot | Thiếu lot-level costing  |
| 4   | Tính theo kho / không theo kho | **NO**                                                     | Thiếu (single warehouse) |

**VG WMS:** Chỉ có `don_gia` trên products (giá đơn tĩnh), `gia_nhap` (giá nhập). Không có costing engine.

---

## VII. MUA HÀNG (AMIS Mua hàng)

| #   | Nghiệp vụ MISA                     | VG WMS                              | Gap                            |
| --- | ---------------------------------- | ----------------------------------- | ------------------------------ |
| 1   | Yêu cầu mua sắm (YCMS) + phê duyệt | **NO**                              | Thiếu hoàn toàn                |
| 2   | Đề nghị mua hàng                   | **NO**                              | Thiếu                          |
| 3   | Đề nghị báo giá + duyệt báo giá    | **NO**                              | Thiếu                          |
| 4   | Đơn mua hàng (PO)                  | **NO**                              | Thiếu                          |
| 5   | Theo dõi đợt giao theo PO          | **NO**                              | Thiếu                          |
| 6   | Mua theo thỏa thuận khung          | **NO**                              | Thiếu                          |
| 7   | Tra cứu tồn kho từ module mua      | **Partial** — API có inventory/grid | Thiếu cross-module integration |
| 8   | Danh mục NCC                       | **NO**                              | Thiếu suppliers table          |
| 9   | Danh mục kho (multi-warehouse)     | **NO**                              | Single warehouse               |
| 10  | Báo cáo mua hàng (8 loại)          | **NO**                              | Thiếu                          |

---

## VIII. QUẢN LÝ KHO NÂNG CAO

| #   | Nghiệp vụ MISA                    | VG WMS                                                | Gap                    |
| --- | --------------------------------- | ----------------------------------------------------- | ---------------------- |
| 1   | Đa kho (multi-warehouse)          | **NO**                                                | **Critical**           |
| 2   | Quản lý vị trí kho (bin/location) | **NO**                                                | High                   |
| 3   | HSD / Hạn sử dụng                 | **NO** — lots chỉ có received_at, không expiry        | High                   |
| 4   | FEFO (First Expired First Out)    | **NO** — chỉ FIFO                                     | Medium (cần HSD trước) |
| 5   | Số lô (lot tracking)              | **YES** — batch_code trên inventory_lots              | OK                     |
| 6   | Đa đơn vị tính (UOM conversion)   | **NO**                                                | Medium                 |
| 7   | QR/Barcode                        | **NO**                                                | Medium                 |
| 8   | Hàng đang vận chuyển (in-transit) | **NO**                                                | High                   |
| 9   | Hàng trả lại chưa xử lý           | **NO**                                                | Medium                 |
| 10  | Dự báo tồn kho                    | **Partial** — threshold alerts, LBBQ, so_ngay_ton_ban | Thiếu ML/forecast      |
| 11  | FIFO/LIFO allocation              | **Partial** — FIFO only                               | Thiếu LIFO option      |
| 12  | Sơ đồ kho trực quan               | **NO**                                                | Low                    |
| 13  | AI trợ lý thủ kho                 | **NO**                                                | Low (advanced)         |
| 14  | IoT/RFID                          | **NO**                                                | Low (advanced)         |

---

## IX. BÁO CÁO

| #   | Nghiệp vụ MISA              | VG WMS                              | Gap                     |
| --- | --------------------------- | ----------------------------------- | ----------------------- |
| 1   | Tồn kho tổng hợp            | **YES** — dashboard summary + grid  | OK                      |
| 2   | Tồn kho chi tiết theo số lô | **YES** — /api/inventory/lots       | OK                      |
| 3   | Tồn kho theo HSD            | **NO**                              | Thiếu HSD               |
| 4   | Tồn kho theo vị trí kho     | **NO**                              | Thiếu bin/location      |
| 5   | Báo cáo kiểm kê             | **NO**                              | Thiếu kiểm kê           |
| 6   | Nhập-Xuất-Tồn (NXT)         | **Partial** — charts in/out by week | Thiếu NXT formal report |
| 7   | Đối chiếu kho - bán hàng    | **NO**                              | Thiếu                   |
| 8   | Đối chiếu kho - mua hàng    | **NO**                              | Thiếu                   |
| 9   | Xuất Excel                  | **YES** — /api/inventory/export     | OK                      |
| 10  | Báo cáo tùy chỉnh           | **NO**                              | Thiếu                   |

---

## X. KẾT NỐI HỆ SINH THÁI

| #   | Kết nối MISA                     | VG WMS | Gap                   |
| --- | -------------------------------- | ------ | --------------------- |
| 1   | AMIS Kế toán (hạch toán tự động) | **NO** | Thiếu kế toán         |
| 2   | AMIS Mua hàng (PO → Nhập kho)    | **NO** | Thiếu PO              |
| 3   | AMIS Bán hàng (SO → Xuất kho)    | **NO** | Thiếu SO              |
| 4   | AMIS Quy trình (phê duyệt)       | **NO** | Thiếu approval        |
| 5   | AMIS Hệ thống (user/role/org)    | **NO** | Thiếu user management |
| 6   | Danh mục dùng chung (tập đoàn)   | **NO** | Thiếu multi-org       |

---

## XI. TỔNG HỢP — BẢNG XẾP HẠNG ƯU TIÊN

### CRITICAL (cần thiết để vận hành kho cơ bản cho doanh nghiệp)

| #   | Gap                               | Lý do Critical                                                          | Effort |
| --- | --------------------------------- | ----------------------------------------------------------------------- | ------ |
| 1   | **Multi-warehouse**               | DN có >1 kho là phổ biến. Không có = không thể mở rộng                  | XL     |
| 2   | **Kiểm kê kho**                   | Nghiệp vụ bắt buộc theo luật kế toán VN. Không có = không thể đối chiếu | L      |
| 3   | **Phân loại nghiệp vụ nhập/xuất** | Hiện chỉ có generic in/out, cần phân biệt loại để hạch toán đúng        | M      |
| 4   | **NCC (Suppliers) + Khách hàng**  | Không track được nguồn hàng và đối tượng bán                            | M      |

### HIGH (cần cho WMS chuyên nghiệp)

| #   | Gap                                                    | Lý do High                                                          | Effort |
| --- | ------------------------------------------------------ | ------------------------------------------------------------------- | ------ |
| 5   | **HSD / Expiry date trên lots**                        | Ngành thực phẩm/dược bắt buộc. Cần thêm field + FEFO logic          | M      |
| 6   | **Stock transfer (chuyển kho)**                        | Phụ thuộc multi-warehouse. Cần transfer document + in-transit state | M      |
| 7   | **In-transit inventory**                               | Hàng đang đi đường cần track riêng                                  | M      |
| 8   | **Purchase Order workflow**                            | PO → Nhận hàng → Nhập kho. Cốt lõi procurement                      | L      |
| 9   | **Returns workflow** (trả hàng mua + hàng bán trả lại) | Cần dedicated process, không chỉ 1 field                            | M      |
| 10  | **Bin/Location management**                            | Kho lớn cần biết hàng ở đâu                                         | L      |

### MEDIUM (nâng cao trải nghiệm)

| #   | Gap                                       | Effort |
| --- | ----------------------------------------- | ------ |
| 11  | Đa đơn vị tính (UOM conversion)           | M      |
| 12  | Costing methods (bình quân, FIFO costing) | L      |
| 13  | QR/Barcode generation + scan              | M      |
| 14  | Báo cáo NXT (Nhập-Xuất-Tồn) chính thức    | S      |
| 15  | User/Role management + approval workflows | L      |
| 16  | FEFO allocation (sau khi có HSD)          | S      |
| 17  | LIFO allocation option                    | S      |

### LOW (advanced, long-term)

| #   | Gap                                   | Effort |
| --- | ------------------------------------- | ------ |
| 18  | Sơ đồ kho trực quan (warehouse map)   | L      |
| 19  | AI trợ lý thủ kho                     | XL     |
| 20  | IoT/RFID integration                  | XL     |
| 21  | Điều phối giao hàng + tối ưu lộ trình | XL     |
| 22  | Kết nối hệ sinh thái kế toán          | XL     |
| 23  | Năng suất nhân viên kho               | M      |
| 24  | Consignment (ký gửi/bán hộ)           | L      |

**Effort legend:** S = <1 tuần, M = 1-2 tuần, L = 3-4 tuần, XL = >1 tháng

---

## XII. ĐIỂM MẠNH CỦA VG WMS (không có trong MISA cơ bản)

| #   | Feature VG WMS                                                 | Ghi chú                                                                                                            |
| --- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| 1   | **Combo/Assembly system** với BOM 2 cấp                        | MISA SME có lắp ráp/tháo dỡ nhưng VG WMS implementation chi tiết hơn (lot-level tracking trên component movements) |
| 2   | **LBBQ + so_ngay_ton_ban** metrics                             | Phân tích tốc độ bán tự động, MISA không nhấn mạnh feature này                                                     |
| 3   | **Versioned thresholds** (min/optimal/max_age + model_version) | Hỗ trợ ML-driven threshold updates                                                                                 |
| 4   | **AG Grid SSR** với filter/sort/pagination                     | UX tốt cho data-heavy views                                                                                        |
| 5   | **Async job queue** (Redis)                                    | Scalable cho bulk operations                                                                                       |
| 6   | **Accessories subsystem** riêng biệt                           | Separate inventory tracking for non-product items                                                                  |

---

## XIII. ĐỀ XUẤT LỘ TRÌNH PHÁT TRIỂN

### Phase 1: Foundation (1-2 tháng)

1. Multi-warehouse support (thêm warehouse_id vào inventory_main, lots, movements)
2. Suppliers + Customers tables
3. Phân loại nghiệp vụ nhập/xuất (movement_type mở rộng)
4. Kiểm kê kho (stock count with variance handling)

### Phase 2: Professional WMS (2-3 tháng)

5. HSD/Expiry trên lots + FEFO allocation
6. Stock transfer + in-transit inventory
7. Purchase Order workflow (PO → Receipt → Inbound)
8. Returns workflow (Sales return + Purchase return)
9. Bin/Location management

### Phase 3: Enterprise (3-4 tháng)

10. UOM conversion
11. Costing methods
12. User/Role + Approval workflows
13. Formal NXT reports
14. QR/Barcode

### Phase 4: Advanced (ongoing)

15. Warehouse map visualization
16. Delivery optimization
17. AI assistant
18. Accounting integration

---

## Phụ lục: Nguồn dữ liệu chi tiết

### A. MISA AMIS Kho hàng (Marketing)

**URL:** amis.misa.vn/amis-kho-hang/

Features extracted:

- Kiểm soát hàng hóa: multi-unit, QR/barcode, mã SP từ NCC, kích thước/số lô/HSD, nhiều vị trí kho
- Quản lý tồn kho: tồn thực tế real-time, hàng đang vận chuyển, hàng trả lại chưa xử lý, dự báo tồn kho
- Tự động hóa: quy trình nhiều bước, nhập-xuất-điều chuyển-kiểm kê, đa kho, FIFO/LIFO/FEFO
- Điều phối giao hàng: auto kế hoạch giao hàng, tối ưu lộ trình, kết nối vận chuyển
- Không gian kho: giao diện trực quan sơ đồ kho, đề xuất vị trí lưu trữ
- Trợ lý AI: đọc biên bản giao hàng NCC, lộ trình soạn/giao hàng, phân tích tồn kho
- Kết nối: Mua hàng, Kế toán, Bán hàng, Sản xuất, Quy trình, Hệ thống

### B. Academy AMIS Kế toán — Phân hệ Kho (14 bài)

- 8 bài Nhập kho (hàng mua đi đường, thành phẩm SX, hàng bán trả lại, NVL SX không dùng hết, chi nhánh phụ thuộc, gia công, bán hộ/ký gửi, giữ hộ/gia công)
- 6 bài Xuất kho (NVL sản xuất, bán hàng, trả lại hàng mua qua kho, không qua kho, giảm giá qua kho, giảm giá không qua kho)

### C. Academy SME.NET Kho (23 bài, 1h17m)

- Nhập kho (4 bài): hàng bán trả lại/gia công/chi nhánh; thành phẩm SX; NVL SX không dùng hết; hàng bán trả lại
- Xuất kho (7 bài): NVL cho SX; XDCB/sửa chữa TSCĐ; góp vốn; bán hàng; biếu tặng/nội bộ; cho chi nhánh; trả lại hàng mua
- Chuyển kho (2 bài): nội bộ; hàng gửi bán đại lý
- Lắp ráp, tháo dỡ (2 bài)
- Kiểm kê kho (1 bài)
- Tính giá xuất kho (1 bài): bình quân cuối kỳ, bình quân tức thời, theo kho/không theo kho
- Lệnh sản xuất (1 bài)
- Báo cáo kho (3 bài)
- Thủ kho (1 bài)

### D. AMIS Mua hàng (helpamis.misa.vn)

- Yêu cầu mua sắm (5+ articles): YCMS via AMIS Quy trình → ký duyệt → đồng bộ
- Quy trình mua hàng (4 types): theo chỉ định, kế hoạch, thỏa thuận, quy trình 4 bước
- Lập ĐMH: 3 cách (theo quy trình, theo thỏa thuận, từ kế hoạch)
- Tra cứu tồn kho: theo mặt hàng/mã quy cách/số lô-HSD
- Danh mục Kho: khai báo kho, đồng bộ kho dùng chung, gộp kho
- Báo cáo: 8 loại (tình hình ĐMH, giá mua, KHMH, YCMS, tổng hợp MH theo NCC, ...)
- Kết nối: Hệ thống, Kế toán, Quy trình, danh mục dùng chung
- Danh mục: 4 nhóm (NCC, mặt hàng, cơ cấu tổ chức, khác)

### E. VG WMS Source Code

- 6 PostgreSQL migrations (001-006)
- 20+ database tables
- 30+ API endpoints
- 5 frontend views (Overview, Inventory, Orders, ComboWarehouse, Settings)
- Key services: InventoryService, OrderService, ComboService, DashboardService, ImporterService
