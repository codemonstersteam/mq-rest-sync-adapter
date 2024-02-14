package team.codemonsters.refactoringexample.walletBalance

enum class OperationType(val operation: String) {
    OPERATION_HISTORY("operationHistory")
}

fun getOperationType(fieldName: String): OperationType? {
    for (enumValue in OperationType.values()) {
        if (enumValue.operation == fieldName) {
            return enumValue
        }
    }
    return null
}