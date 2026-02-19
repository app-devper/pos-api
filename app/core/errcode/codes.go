package errcode

// ─── Auth / Middleware ───────────────────────────────────────────────────────
const (
	AU_UNAUTHORIZED_001 = "AU-401-001" // missing / invalid authorization header
	AU_UNAUTHORIZED_002 = "AU-401-002" // token invalid or expired
	AU_UNAUTHORIZED_003 = "AU-401-003" // system mismatch
	AU_UNAUTHORIZED_004 = "AU-401-004" // clientId mismatch
	AU_UNAUTHORIZED_005 = "AU-401-005" // session invalid
	AU_FORBIDDEN_001    = "AU-403-001" // employee not found / no branch access
)

// ─── Branch (BR) ────────────────────────────────────────────────────────────
const (
	BR_BAD_REQUEST_001 = "BR-400-001" // invalid request body
	BR_BAD_REQUEST_002 = "BR-400-002" // create/update/delete failed
	BR_INTERNAL_001    = "BR-500-001" // internal server error
)

// ─── Product (PD) ───────────────────────────────────────────────────────────
const (
	PD_BAD_REQUEST_001 = "PD-400-001" // invalid request body
	PD_BAD_REQUEST_002 = "PD-400-002" // create/update/delete failed
	PD_INTERNAL_001    = "PD-500-001" // internal server error
)

// ─── Order (OR) ─────────────────────────────────────────────────────────────
const (
	OR_BAD_REQUEST_001 = "OR-400-001" // invalid request body
	OR_BAD_REQUEST_002 = "OR-400-002" // create/update/delete failed
	OR_INTERNAL_001    = "OR-500-001" // internal server error
)

// ─── Category (CA) ──────────────────────────────────────────────────────────
const (
	CA_BAD_REQUEST_001 = "CA-400-001" // invalid request body
	CA_BAD_REQUEST_002 = "CA-400-002" // create/update/delete failed
	CA_INTERNAL_001    = "CA-500-001" // internal server error
)

// ─── Customer (CU) ──────────────────────────────────────────────────────────
const (
	CU_BAD_REQUEST_001 = "CU-400-001" // invalid request body
	CU_BAD_REQUEST_002 = "CU-400-002" // create/update/delete failed
	CU_INTERNAL_001    = "CU-500-001" // internal server error
)

// ─── Supplier (SU) ──────────────────────────────────────────────────────────
const (
	SU_BAD_REQUEST_001 = "SU-400-001" // invalid request body
	SU_BAD_REQUEST_002 = "SU-400-002" // create/update/delete failed
	SU_INTERNAL_001    = "SU-500-001" // internal server error
)

// ─── Receive (RC) ───────────────────────────────────────────────────────────
const (
	RC_BAD_REQUEST_001 = "RC-400-001" // invalid request body
	RC_BAD_REQUEST_002 = "RC-400-002" // create/update/delete failed
	RC_INTERNAL_001    = "RC-500-001" // internal server error
)

// ─── Employee (EM) ──────────────────────────────────────────────────────────
const (
	EM_BAD_REQUEST_001 = "EM-400-001" // invalid request body
	EM_BAD_REQUEST_002 = "EM-400-002" // create/update/delete failed
	EM_INTERNAL_001    = "EM-500-001" // internal server error
)

// ─── Setting (SE) ───────────────────────────────────────────────────────────
const (
	SE_BAD_REQUEST_001 = "SE-400-001" // invalid request body
	SE_BAD_REQUEST_002 = "SE-400-002" // upsert failed
	SE_INTERNAL_001    = "SE-500-001" // internal server error
)

// ─── Purchase Order (PO) ────────────────────────────────────────────────────
const (
	PO_BAD_REQUEST_001 = "PO-400-001" // invalid request body
	PO_BAD_REQUEST_002 = "PO-400-002" // create/update/delete failed
	PO_INTERNAL_001    = "PO-500-001" // internal server error
)

// ─── Delivery Order (DO) ────────────────────────────────────────────────────
const (
	DO_BAD_REQUEST_001 = "DO-400-001" // invalid request body
	DO_BAD_REQUEST_002 = "DO-400-002" // create/update/delete failed
	DO_INTERNAL_001    = "DO-500-001" // internal server error
)

// ─── Credit Note (CN) ───────────────────────────────────────────────────────
const (
	CN_BAD_REQUEST_001 = "CN-400-001" // invalid request body
	CN_BAD_REQUEST_002 = "CN-400-002" // create/update/delete failed
	CN_INTERNAL_001    = "CN-500-001" // internal server error
)

// ─── Billing (BL) ───────────────────────────────────────────────────────────
const (
	BL_BAD_REQUEST_001 = "BL-400-001" // invalid request body
	BL_BAD_REQUEST_002 = "BL-400-002" // create/update/delete failed
	BL_INTERNAL_001    = "BL-500-001" // internal server error
)

// ─── Quotation (QT) ─────────────────────────────────────────────────────────
const (
	QT_BAD_REQUEST_001 = "QT-400-001" // invalid request body
	QT_BAD_REQUEST_002 = "QT-400-002" // create/update/delete failed
	QT_INTERNAL_001    = "QT-500-001" // internal server error
)

// ─── Promotion (PM) ─────────────────────────────────────────────────────────
const (
	PM_BAD_REQUEST_001 = "PM-400-001" // invalid request body
	PM_BAD_REQUEST_002 = "PM-400-002" // create/update/delete/apply failed
	PM_INTERNAL_001    = "PM-500-001" // internal server error
)

// ─── Customer History (CH) ──────────────────────────────────────────────────
const (
	CH_BAD_REQUEST_001 = "CH-400-001" // invalid request body
	CH_BAD_REQUEST_002 = "CH-400-002" // create failed
	CH_INTERNAL_001    = "CH-500-001" // internal server error
)

// ─── Patient (PT) ───────────────────────────────────────────────────────────
const (
	PT_BAD_REQUEST_001 = "PT-400-001" // invalid request body
	PT_BAD_REQUEST_002 = "PT-400-002" // create/update/delete/check failed
	PT_INTERNAL_001    = "PT-500-001" // internal server error
)

// ─── Dispensing Log (DI) ────────────────────────────────────────────────────
const (
	DI_BAD_REQUEST_001 = "DI-400-001" // invalid request body
	DI_BAD_REQUEST_002 = "DI-400-002" // create failed
	DI_INTERNAL_001    = "DI-500-001" // internal server error
)

// ─── Stock Transfer (TR) ────────────────────────────────────────────────────
const (
	TR_BAD_REQUEST_001 = "TR-400-001" // invalid request body
	TR_BAD_REQUEST_002 = "TR-400-002" // create/approve/reject failed
	TR_INTERNAL_001    = "TR-500-001" // internal server error
)

// ─── Report (RP) ────────────────────────────────────────────────────────────
const (
	RP_BAD_REQUEST_001 = "RP-400-001" // invalid request / missing params
	RP_BAD_REQUEST_002 = "RP-400-002" // report generation failed
	RP_INTERNAL_001    = "RP-500-001" // internal server error
)

// ─── Dashboard (DA) ─────────────────────────────────────────────────────────
const (
	DA_BAD_REQUEST_001 = "DA-400-001" // invalid request / missing params
	DA_BAD_REQUEST_002 = "DA-400-002" // query failed
	DA_INTERNAL_001    = "DA-500-001" // internal server error
)

// ─── System (SY) ────────────────────────────────────────────────────────────
const (
	SY_NOT_FOUND_001 = "SY-404-001" // route not found
	SY_FORBIDDEN_001 = "SY-403-001" // invalid request / restricted endpoint
	SY_FORBIDDEN_002 = "SY-403-002" // no permission
	SY_INTERNAL_001  = "SY-500-001" // panic recovery / internal server error
)
