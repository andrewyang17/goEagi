// Code generated by "stringer -type=CancellationErrorCode -output=cancellation_error_code_string.go"; DO NOT EDIT.

package common

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NoError-0]
	_ = x[AuthenticationFailure-1]
	_ = x[BadRequest-2]
	_ = x[TooManyRequests-3]
	_ = x[Forbidden-4]
	_ = x[ConnectionFailure-5]
	_ = x[ServiceTimeout-6]
	_ = x[ServiceError-7]
	_ = x[ServiceUnavailable-8]
	_ = x[RuntimeError-9]
}

const _CancellationErrorCode_name = "NoErrorAuthenticationFailureBadRequestTooManyRequestsForbiddenConnectionFailureServiceTimeoutServiceErrorServiceUnavailableRuntimeError"

var _CancellationErrorCode_index = [...]uint8{0, 7, 28, 38, 53, 62, 79, 93, 105, 123, 135}

func (i CancellationErrorCode) String() string {
	if i < 0 || i >= CancellationErrorCode(len(_CancellationErrorCode_index)-1) {
		return "CancellationErrorCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _CancellationErrorCode_name[_CancellationErrorCode_index[i]:_CancellationErrorCode_index[i+1]]
}
