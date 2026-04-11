package constant

const (
	CONFIRMED = "CONFIRMED"
	CANCELLED = "CANCELLED"
	VOIDED    = "VOIDED"
)

func ConfirmedOrderStatuses() []string {
	return []string{ACTIVE, CONFIRMED}
}

func IsConfirmedOrderStatus(status string) bool {
	for _, candidate := range ConfirmedOrderStatuses() {
		if status == candidate {
			return true
		}
	}
	return false
}

func IsConfirmedOrderItemStatus(status string) bool {
	return status == "" || IsConfirmedOrderStatus(status)
}
