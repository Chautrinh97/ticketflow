# Feedback components

Token dùng: [../design-system.md](../design-system.md). Quy tắc **khi nào dùng** Toast/Modal/Confirm/Empty/Error: [../interaction-patterns.md](../interaction-patterns.md) — file này chỉ mô tả **cấu trúc & hành vi riêng** của từng component.

## Toast

**Mục đích:** thông báo kết quả hành động đã hoàn tất. **Dùng ở:** hầu hết mọi màn hình có thao tác ghi dữ liệu.

**Cấu trúc:** khối `rounded-lg shadow-md p-4`, icon trạng thái bên trái (check tròn `success` / dấu X tròn `danger`) + nội dung text + `IconButton ghost` đóng (X) bên phải. Nền: `success` → viền trái `border-l-4 border-green-600` nền `bg-white`; `danger` → viền trái `border-l-4 border-red-600`.

**Hành vi:** tự ẩn sau 4s (`success`) / 6s (`danger`) hoặc khi bấm nút đóng; nhiều toast xếp chồng theo chiều dọc, toast mới nhất ở trên cùng, tối đa 3 toast hiển thị đồng thời (toast cũ nhất tự ẩn nếu vượt quá).

## Modal

**Mục đích:** khung hội thoại nổi che nội dung phía sau, dùng cho form ngắn hoặc làm khung chứa `ConfirmDialog`. **Dùng ở:** event-manage.md (sửa nhanh 1 loại vé, Phase 2), mọi nơi cần `ConfirmDialog`.

**Cấu trúc:** overlay (`bg-black/50`) phủ toàn màn hình, khối modal căn giữa `rounded-xl shadow-xl bg-white p-6`, max-width theo nội dung (`max-w-md` mặc định, `max-w-2xl` cho form). Header modal: tiêu đề (`H3`) + `IconButton ghost` đóng (X) góc phải. Footer modal: các `Button` hành động, canh phải, tối đa 1 `primary`.

**Hành vi:** đóng khi click overlay/nhấn `Esc`/bấm nút X — trừ khi có thay đổi chưa lưu (mở thêm `ConfirmDialog` hỏi trước khi đóng thật, xem interaction-patterns.md). Khi mở modal, khoá cuộn (`scroll lock`) của trang phía sau.

## ConfirmDialog

**Mục đích:** biến thể chuyên biệt của `Modal`, dùng riêng cho xác nhận hành động nguy hiểm (danh sách hành động cụ thể tại [../interaction-patterns.md](../interaction-patterns.md#confirm-dialog-modal-xác-nhận-hành-động-nguy-hiểm)). **Dùng ở:** event-manage.md (huỷ sự kiện, Phase 2), booking-detail.md (huỷ vé, Phase 2), user-management.md (khoá tài khoản, từ chối organizer, Phase 2).

**Cấu trúc:** `Modal` kích thước nhỏ (`max-w-sm`), icon cảnh báo (`triangle-alert`, màu `danger`) phía trên tiêu đề, nội dung xác nhận ngắn gọn, footer 2 nút: `secondary` ("Huỷ", đóng dialog) và `danger` (hành động thật sự, luôn nằm bên phải).

**Hành vi:** nút hành động chuyển `loading` khi đang gọi API; đóng dialog + hiện `Toast` kết quả sau khi API phản hồi.

## LoadingSpinner / Skeleton

**Mục đích:** biểu thị trạng thái đang tải (xem quy tắc chọn loại tại interaction-patterns.md). **Dùng ở:** mọi màn hình có tải dữ liệu.

- `LoadingSpinner`: icon xoay (`loader-circle` xoay 360° liên tục), size khớp ngữ cảnh (16px trong Button, 24px độc lập trong panel nhỏ).
- `Skeleton`: khối `bg-gray-200 animate-pulse rounded-md`, kích thước mô phỏng đúng layout thật (vd `SkeletonBlock` cho `EventCard` gồm 1 khối ảnh 16:9 + 2 dòng text ngắn).

## EmptyState

**Mục đích:** thay thế nội dung khi danh sách/bảng rỗng. **Dùng ở:** mọi danh sách trong hệ thống (my-bookings.md, event-list.md, notifications.md (Phase 2), user-management.md (Phase 2)...).

**Cấu trúc:** căn giữa theo chiều dọc trong khung chứa, icon minh hoạ (`w-6 h-6` theo design-system, màu `text-gray-400`) → tiêu đề ngắn (`H3`) → mô tả phụ (`Body small`) → `Button` CTA (nếu màn hình cho phép tự tạo dữ liệu, vd "Tạo sự kiện đầu tiên").

## InlineFormError

**Mục đích:** hiển thị lỗi validate ngay dưới field liên quan (thay cho helper text khi có lỗi). **Dùng ở:** mọi `Input`/`Select`/`DatePicker`/`FileUpload` trong form.

**Cấu trúc:** `text-xs text-red-600`, có icon cảnh báo nhỏ (`circle-alert`, `w-4 h-4`) đứng trước text.
