# Component dùng chung

Danh mục component dùng chung giữa nhiều màn hình. Đây là **nguồn duy nhất** mô tả các thành phần lặp lại (Header, Footer, Button, Input, Card, Modal, Toast...).

## Nguyên tắc bắt buộc

**Một màn hình dùng tới thành phần đã có trong 4 file dưới đây thì CHỈ được tham chiếu (tên component + variant/props áp dụng), KHÔNG được mô tả lại chi tiết cấu trúc/màu sắc/hành vi của nó.** Mục đích: một component chỉ có một định nghĩa thị giác + hành vi duy nhất trong toàn repo — tránh tình trạng mỗi màn hình tự vẽ lại Header/Button/Input theo cách khác nhau rồi lệch nhau khi implement.

## Danh sách file

| File | Component |
|---|---|
| [layout.md](layout.md) | Header, Footer, Sidebar (dashboard), Breadcrumb, PageContainer |
| [buttons-inputs.md](buttons-inputs.md) | Button, IconButton, Input, QuantityStepper, Select, DatePicker, SearchBar, FileUpload, Checkbox/Radio |
| [feedback.md](feedback.md) | Toast, Modal, ConfirmDialog, LoadingSpinner/Skeleton, EmptyState, InlineFormError |
| [data-display.md](data-display.md) | EventCard, TicketTypeRow, Badge/StatusTag, DataTable, Pagination, Avatar, Tabs, StatTile |

Token màu/spacing/typography dùng trong mọi component đều lấy từ [../design-system.md](../design-system.md). Quy tắc hành vi chung (loading/toast/modal/validate/pagination) lấy từ [../interaction-patterns.md](../interaction-patterns.md) — file component chỉ mô tả hành vi **riêng** của chính component đó (vd Modal đóng khi click overlay).

## Quy trình thêm component mới

1. Rà soát 4 file trên — nếu đã có thành phần tương đương (dù tên khác), ưu tiên mở rộng variant của component sẵn có thay vì tạo mới.
2. Chỉ tạo component mới khi thực sự không có thành phần tương đương.
3. Trước khi thêm, liệt kê **danh sách màn hình dự kiến sẽ dùng** component này. Nếu chỉ 1 màn hình dùng và không có kế hoạch tái sử dụng ở màn hình khác, **không** tách thành component dùng chung — mô tả trực tiếp trong chính screen spec đó (mục "Chi tiết từng section riêng").
4. Thêm component vào file nhóm phù hợp nhất (layout/buttons-inputs/feedback/data-display); nếu không thuộc nhóm nào, cân nhắc tạo file nhóm mới và cập nhật bảng "Danh sách file" ở trên.
5. Mô tả theo đúng khuôn: mục đích — danh sách màn hình dùng — variant/size/state (bảng) — cấu trúc trực quan (bảng phần tử + token) — hành vi tương tác riêng.
6. Rà lại các screen spec đang mô tả tay phần tương đương với component mới — sửa lại thành tham chiếu, tránh trùng lặp mô tả.
