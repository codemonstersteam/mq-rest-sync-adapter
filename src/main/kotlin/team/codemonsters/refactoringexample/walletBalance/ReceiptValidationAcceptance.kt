package team.codemonsters.refactoringexample.walletBalance

import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import java.time.OffsetDateTime

data class ReceiptValidationAcceptance private constructor(
    val requestType: String,
    val dateFrom: OffsetDateTime,
    val dateTo: OffsetDateTime?,
    val operTime: OffsetDateTime,
    val flag: ReceiptStatus,
    val errorCode: Int? = null,
    val description: String? = null
) {
    companion object {
        fun accepted(request: ValidRequest) =
            ReceiptValidationAcceptance(
                request.balanceRequest.requestType,
                request.balanceRequest.dateFrom,
                request.balanceRequest.dateTo,
                request.balanceRequest.operTime,
                ReceiptStatus.ACCEPTED,
                null,
                null
            )
    }
}