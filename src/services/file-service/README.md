# File Service

**Trạng thái:** chưa implement — placeholder scaffold.

Sinh presigned URL upload lên object storage. Stateless — không sở hữu database.

- Domain spec: [../../../docs/02-domains/file-storage/spec.md](../../../docs/02-domains/file-storage/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/file-service.yaml](../../../api-docs/openapi/file-service.yaml)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md) (bỏ qua `repository/` nếu thực sự không cần truy cập DB nào).
