# Đăng nhập / Đăng ký

**Route:** `/login` · **Vai trò:** public (chưa đăng nhập — nếu đã đăng nhập, tự động điều hướng về trang trước đó hoặc `/`) · **Render mode:** CSR · **Phase:** 1 (MVP)

**Liên quan:** [docs/02-domains/identity/spec.md](../../02-domains/identity/spec.md) · [docs/04-security/authentication.md](../../04-security/authentication.md) · [api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│              (không có Header)           │
│                                           │
│            [Logo TicketFlow]             │
│         Chào mừng bạn quay lại           │
│                                           │
│   ┌───────────────────────────────────┐ │
│   │  [Panel đăng nhập]                 │ │
│   │  - Nút "Tiếp tục với Google"       │ │
│   │  - Nút "Tiếp tục với Facebook"     │ │
│   │  - đường kẻ "hoặc"                 │ │
│   │  - Input email + Input password    │ │
│   │  - Nút "Đăng nhập"                 │ │
│   │  - Link "Quên mật khẩu?"           │ │
│   └───────────────────────────────────┘ │
│                                           │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Panel đăng nhập | Riêng màn hình này |
| Input email/password, Button | Component dùng chung — [`Input`](../../components/buttons-inputs.md#input-text--email--password--textarea--number), [`Button`](../../components/buttons-inputs.md#button) |

Không dùng `Header`/`Footer` chung của layout `public` — trang đăng nhập độc lập, chỉ có logo dẫn về `/`.

## Chi tiết section riêng

**Panel đăng nhập** — khối `rounded-xl shadow-sm border border-gray-200 p-8 max-w-md mx-auto`:

- 2 nút đăng nhập mạng xã hội (Google/Facebook): `Button` variant `secondary`, full-width, icon thương hiệu bên trái — kích hoạt Firebase SDK popup.
- Đường kẻ phân cách với chữ "hoặc" ở giữa (`text-gray-400 text-sm`).
- Form email/password: `Input` variant `email` + `Input` variant `password`.
- `Button` variant `primary`, full-width, label "Đăng nhập".
- Link nhỏ "Quên mật khẩu?" (`text-sm text-blue-600`) — mở luồng reset password của Firebase (form/modal riêng, không đặc tả chi tiết ở phase này).
- Dòng dưới cùng: "Chưa có tài khoản? Đăng ký ngay" — thực chất dùng chung 1 form (email/password chưa tồn tại → tự tạo tài khoản mới), không tách 2 route riêng.

Khi backend chạy với `AUTH_FIREBASE_MODE=mock` (dev cục bộ, không cần Firebase project thật), màn hình nhận trực tiếp email + tên hiển thị để dựng token mà `POST /auth/login` yêu cầu, thay vì mở popup Firebase thật.

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm "Tiếp tục với Google/Facebook" | Mở popup Firebase SDK → nhận Firebase ID token → gọi `POST /auth/login` → thành công: lưu `access_token` vào bộ nhớ (xem [README.md](../../README.md#auth-ở-phía-frontend)), điều hướng về trang trước đó hoặc `/` |
| Nhập email/password, bấm "Đăng nhập" | Validate theo quy tắc chung ([interaction-patterns.md](../../interaction-patterns.md#validate-form)) → gọi Firebase SDK sign-in → `POST /auth/login` → như trên |
| Sai email/mật khẩu | `InlineFormError` dưới field password: "Email hoặc mật khẩu không đúng" |
| `POST /auth/login` lỗi mạng/server | `Toast` `danger`: "Có lỗi xảy ra, vui lòng thử lại" |
| Tài khoản `status=banned` | `InlineFormError` cấp form (trên toàn panel): "Tài khoản của bạn đã bị khoá. Liên hệ hỗ trợ nếu cần trợ giúp." |

## Trạng thái đặc biệt

Nút "Đăng nhập"/nút mạng xã hội chuyển `loading` (xem [components/buttons-inputs.md](../../components/buttons-inputs.md#button)) trong lúc chờ Firebase + `POST /auth/login` phản hồi.

## Responsive

Panel đăng nhập full-width trừ padding `px-4` dưới `md`, bỏ `max-w-md` cố định.
