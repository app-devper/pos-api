package usecase

import "pos/app/domain/constant"

func isValidCustomerStatus(status string) bool {
	switch status {
	case constant.ACTIVE, constant.INACTIVE, constant.ARCHIVED, "DRAFT":
		return true
	default:
		return false
	}
}
