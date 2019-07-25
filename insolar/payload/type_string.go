// Code generated by "stringer -type=Type"; DO NOT EDIT.

package payload

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeUnknown-0]
	_ = x[TypeMeta-1]
	_ = x[TypeError-2]
	_ = x[TypeID-3]
	_ = x[TypeIDs-4]
	_ = x[TypeJet-5]
	_ = x[TypeState-6]
	_ = x[TypeGetObject-7]
	_ = x[TypePassState-8]
	_ = x[TypeObjIndex-9]
	_ = x[TypeObjState-10]
	_ = x[TypeIndex-11]
	_ = x[TypePass-12]
	_ = x[TypeGetCode-13]
	_ = x[TypeCode-14]
	_ = x[TypeSetCode-15]
	_ = x[TypeSetIncomingRequest-16]
	_ = x[TypeSetOutgoingRequest-17]
	_ = x[TypeSagaCallAcceptNotification-18]
	_ = x[TypeGetFilament-19]
	_ = x[TypeGetRequest-20]
	_ = x[TypeRequest-21]
	_ = x[TypeFilamentSegment-22]
	_ = x[TypeSetResult-23]
	_ = x[TypeActivate-24]
	_ = x[TypeRequestInfo-25]
	_ = x[TypeDeactivate-26]
	_ = x[TypeUpdate-27]
	_ = x[TypeHotObjects-28]
	_ = x[TypeResultInfo-29]
	_ = x[TypeGetPendings-30]
	_ = x[TypeHasPendings-31]
	_ = x[TypePendingsInfo-32]
	_ = x[TypeReplication-33]
	_ = x[TypeGetJet-34]
	_ = x[TypeAbandonedRequestsNotification-35]
	_ = x[TypeGetIndex-36]
	_ = x[TypeReturnResults-37]
	_ = x[TypeCallMethod-38]
	_ = x[TypeExecutorResults-39]
	_ = x[TypePendingFinished-40]
	_ = x[TypeAdditionalCallFromPreviousExecutor-41]
	_ = x[TypeStillExecuting-42]
	_ = x[_latestType-43]
}

const _Type_name = "TypeUnknownTypeMetaTypeErrorTypeIDTypeIDsTypeJetTypeStateTypeGetObjectTypePassStateTypeObjIndexTypeObjStateTypeIndexTypePassTypeGetCodeTypeCodeTypeSetCodeTypeSetIncomingRequestTypeSetOutgoingRequestTypeSagaCallAcceptNotificationTypeGetFilamentTypeGetRequestTypeRequestTypeFilamentSegmentTypeSetResultTypeActivateTypeRequestInfoTypeDeactivateTypeUpdateTypeHotObjectsTypeResultInfoTypeGetPendingsTypeHasPendingsTypePendingsInfoTypeReplicationTypeGetJetTypeAbandonedRequestsNotificationTypeGetIndexTypeReturnResultsTypeCallMethodTypeExecutorResultsTypePendingFinishedTypeAdditionalCallFromPreviousExecutorTypeStillExecuting_latestType"

var _Type_index = [...]uint16{0, 11, 19, 28, 34, 41, 48, 57, 70, 83, 95, 107, 116, 124, 135, 143, 154, 176, 198, 228, 243, 257, 268, 287, 300, 312, 327, 341, 351, 365, 379, 394, 409, 425, 440, 450, 483, 495, 512, 526, 545, 564, 602, 620, 631}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
