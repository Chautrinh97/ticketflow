# Domain: AI Integration *(nice-to-have — Phase 4)*

**Service sở hữu:** chưa xác định tên service cụ thể — có thể là một endpoint bổ sung trong Event Service (auto mô tả) và một service mỏng riêng cho chatbot/gợi ý · **Database:** PostgreSQL với extension `pgvector` (hoặc `dense_vector` nếu đã có Elasticsearch) · **Phase:** 4 (Nice-to-have).

Chỉ gọi API suy luận có sẵn (OpenAI/Claude/Gemini...) qua inference API — **không** tự huấn luyện hay fine-tune model.

## 1. Chatbot hỗ trợ khách hàng

RAG (Retrieval-Augmented Generation) đơn giản: retrieve FAQ/thông tin sự kiện liên quan (từ `event_catalog.faq` — xem [../../03-data/mongodb-schema.md](../../03-data/mongodb-schema.md) — và dữ liệu `events` công khai) rồi đưa vào prompt cho LLM sinh câu trả lời. Không cho phép chatbot trả lời dựa trên dữ liệu nhạy cảm (thông tin thanh toán, dữ liệu user khác) — retrieval chỉ giới hạn ở dữ liệu công khai của sự kiện.

## 2. Gợi ý sự kiện cá nhân hoá

Sinh embedding cho mô tả sự kiện (`events.description` + `event_catalog.tags`) và lịch sử mua vé của user (từ `orders`/`order_items`), lưu vào PostgreSQL (`pgvector`) hoặc `dense_vector` của Elasticsearch nếu Search Service đã tồn tại. Tìm kiếm sự kiện gợi ý theo độ tương đồng vector (cosine similarity) với embedding lịch sử của user.

- Embedding sự kiện được tính lại khi sự kiện được tạo/cập nhật mô tả (trigger từ Event Service, tương tự cách `search_vector` được tính lại khi xuất bản — xem [../event-catalog/spec.md](../event-catalog/spec.md)).
- Đây là tính năng đọc-only đối với dữ liệu nghiệp vụ gốc — không ghi ngược thay đổi vào `events`/`orders`.

## 3. Tự sinh mô tả/tag sự kiện

Organizer nhập vài dòng ngắn mô tả ý tưởng sự kiện → gọi LLM API sinh mô tả đầy đủ + gợi ý tag/SEO meta, hiển thị cho organizer **xem trước và chỉnh sửa** trước khi lưu — không tự động lưu thẳng nội dung do LLM sinh ra vào `events.description` mà không qua xác nhận của organizer (tránh nội dung sai lệch/không phù hợp được xuất bản mà không ai review).

## Nguyên tắc chung khi tích hợp AI

- Luôn gọi qua inference API bên thứ ba (không tự host model) — chi phí và độ trễ tính theo API, cần có timeout/fallback hợp lý (vd: nếu chatbot không phản hồi kịp, hiển thị thông báo lỗi thay vì treo request).
- Không gửi dữ liệu nhạy cảm của user (thông tin thanh toán, email, số điện thoại) vào prompt gửi cho LLM bên thứ ba trừ khi thực sự cần thiết và đã xác nhận chính sách dữ liệu của provider phù hợp.

## Phân theo phase

| Tính năng | Phase |
|---|---|
| Chatbot RAG, gợi ý cá nhân hoá (embedding + pgvector), tự sinh mô tả/tag | 4 (Nice-to-have) |
