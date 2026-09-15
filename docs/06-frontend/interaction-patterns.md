# Interaction Patterns

Quy tắc hành vi dùng chung cho toàn bộ frontend — mọi màn hình ([screens/](screens/)) chỉ mô tả **nội dung/text cụ thể** của màn hình đó (vd nội dung thông báo lỗi, tiêu đề modal), còn **quy tắc khi nào/ứng xử ra sao** thì tham chiếu về tài liệu này, không lặp lại.

## Loading

- **Skeleton** (khung xám mô phỏng layout thật): dùng khi tải **dữ liệu chính của trang lần đầu** (danh sách sự kiện, chi tiết sự kiện, bảng dữ liệu dashboard) — giữ nguyên layout, tránh nhảy giật (layout shift) khi dữ liệu về.
- **Spinner** (biểu tượng xoay): dùng cho **hành động do người dùng chủ động kích hoạt** trong phạm vi nhỏ — bên trong Button khi submit, bên trong 1 panel khi refetch lại sau khi filter/đổi trang.
- Không dùng spinner toàn trang (full-page spinner che hết nội dung cũ) khi chỉ refetch lại dữ liệu đã có sẵn — giữ nội dung cũ hiển thị mờ đi (`opacity-60`) trong lúc chờ dữ liệu mới, tránh màn hình trắng/giật.

## Toast (thông báo ngắn hạn)

- Dùng cho kết quả của một hành động **đã hoàn tất**, không cần người dùng phải đọc kỹ hay phản hồi lại: tạo/lưu thành công, xoá thành công, lỗi mạng tạm thời.
- Vị trí: góc trên bên phải màn hình (desktop), full-width dưới header (mobile).
- Thời gian hiển thị: 4 giây tự ẩn với toast thành công (`success`); 6 giây với toast lỗi (`danger`) — có nút đóng thủ công (X) ở cả hai loại.
- **Không dùng toast** cho lỗi mà người dùng cần đọc kỹ để sửa (lỗi validate form, lỗi nghiệp vụ cần giải thích — vd "Không đủ vé") — dùng **inline error** ngay tại vị trí liên quan (xem mục Validate form) hoặc banner lỗi trong trang.

## Modal / Dialog

- Dùng modal khi: (a) hành động **xác nhận** ngắn (xem mục Confirm dialog), hoặc (b) form **ngắn, ít trường**, không cần rời khỏi ngữ cảnh trang hiện tại (vd sửa nhanh 1 loại vé, Phase 2).
- **Không dùng modal** cho form dài nhiều bước hoặc cần SEO/deep-link riêng (vd tạo sự kiện) — điều hướng sang trang riêng.
- Đóng modal khi: bấm nút đóng (X), click vào overlay, hoặc nhấn phím `Esc` — **trừ khi** modal đang có thay đổi chưa lưu (unsaved change), lúc đó phải mở thêm Confirm dialog hỏi "Rời khỏi mà không lưu?" trước khi đóng thật.
- Modal luôn có tối đa 1 CTA chính (`primary`) — các hành động còn lại là `secondary`/text button.

## Confirm dialog (modal xác nhận hành động nguy hiểm)

Bắt buộc hiển thị Confirm dialog trước khi thực hiện — không thực hiện ngay khi bấm nút gốc — với mọi hành động sau (đối chiếu danh sách hành động nhạy cảm tại [../04-security/authorization.md](../04-security/authorization.md)):

| Hành động | Nội dung xác nhận |
|---|---|
| Huỷ/xoá sự kiện (Phase 2) | "Sự kiện sẽ chuyển sang trạng thái Đã huỷ, người mua vé sẽ được thông báo. Tiếp tục?" |
| Huỷ vé (user) (Phase 2) | "Vé sẽ bị huỷ và không thể khôi phục. Tiếp tục?" |
| Khoá tài khoản (admin) (Phase 2) | "Tài khoản sẽ không thể đăng nhập cho tới khi được mở khoá lại. Tiếp tục?" |
| Từ chối yêu cầu organizer (admin) (Phase 2) | "Yêu cầu trở thành organizer sẽ bị từ chối. Tiếp tục?" |

Nút xác nhận trong các dialog này luôn dùng token `danger`, nút còn lại là `secondary` — không dùng `primary` cho hành động phá huỷ (xem [design-system.md](design-system.md)).

## Validate form

- Validate **on blur** (rời khỏi field) cho từng field, và validate lại toàn bộ **on submit**.
- Lỗi hiển thị **inline** ngay dưới field (component `InlineFormError`, xem [components/feedback.md](components/feedback.md)), không dùng toast.
- Schema validate phía client dùng chung **Zod**, phản ánh đúng schema request body trong OpenAPI tương ứng ([../../api-docs/openapi/](../../api-docs/openapi/)) — không tự định nghĩa rule khác biệt với backend.
- Nút submit chuyển sang trạng thái loading (spinner + disable) ngay khi bấm, không cho bấm lặp lại trong lúc chờ phản hồi.

## Empty state

Mọi danh sách/bảng không có dữ liệu phải hiển thị: icon minh hoạ + câu mô tả ngắn + CTA phù hợp nếu người dùng có thể tự tạo ra dữ liệu đó (vd organizer chưa có sự kiện nào → nút "Tạo sự kiện đầu tiên"). Không để trống trắng hoặc chỉ hiện "Không có dữ liệu". Chi tiết cấu trúc component tại [components/feedback.md](components/feedback.md).

## Error state (lỗi gọi API / mất kết nối)

Hiển thị banner/khối lỗi tại đúng vị trí nội dung lẽ ra xuất hiện (không phải toast, không phải trang trắng), kèm nút "Thử lại". Nếu lỗi 401 (access token hết hạn và refresh cũng thất bại): tự động điều hướng về [screens/public/auth-login.md](screens/public/auth-login.md) theo đúng luồng đã mô tả ở [README.md](README.md#auth-ở-phía-frontend).

## Pagination

Chốt **một quy ước duy nhất theo loại danh sách**, không để mỗi màn hình tự chọn kiểu khác nhau:

- **"Xem thêm" (load more, infinite-scroll-like button)**: dùng cho danh sách hướng người dùng cuối — danh sách/tìm kiếm sự kiện, Vé của tôi, Thông báo (Phase 2).
- **Phân trang theo số trang (offset + page number)**: dùng cho bảng dữ liệu organizer/admin — danh sách sự kiện của tôi, danh sách người mua vé, quản lý user, audit log (tất cả: Phase 2). Lý do khác biệt: bảng dữ liệu quản trị cần định vị nhanh theo trang cụ thể và thường đi kèm sort/filter phức tạp hơn, trong khi danh sách hướng người dùng cuối ưu tiên trải nghiệm cuộn liên tục.

Component tương ứng: xem `Pagination` tại [components/data-display.md](components/data-display.md).

## Optimistic update

Chỉ áp dụng optimistic UI (cập nhật giao diện ngay, không chờ response) cho thao tác **không có rủi ro sai lệch nghiêm trọng nếu phải rollback** — vd đánh dấu thông báo đã đọc (Phase 2 — Notification Service chưa có ở Phase 1). **Không** áp dụng optimistic UI cho: đặt vé (có thể fail do hết tồn kho — phải chờ response thật từ `POST /bookings`), mọi hành động ghi dữ liệu tài chính/tồn kho. Khi optimistic update thất bại, rollback UI về trạng thái cũ và hiện toast lỗi.
