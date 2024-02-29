package team.codemonsters.refactoringexample.walletBalance

import org.springframework.http.HttpStatus
import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import java.time.OffsetDateTime

data class TransactionReceipt private constructor(
    val requestType: String,
    val dateFrom: OffsetDateTime,
    val dateTo: OffsetDateTime?,
    val operTime: OffsetDateTime,
    val flag: ReceiptStatus,
    val errorCode: Int? = null,
    val description: String? = null
) {
    companion object {
        private val SUCCESS_HTTP_STATUSES = setOf(HttpStatus.ACCEPTED, HttpStatus.OK)
        fun receipt(walletBalance: WalletBalance): TransactionReceipt {
            val request = walletBalance.validRequest
            var receiptStatus = ReceiptStatus.SUCCESS
            var receiptErrorCode: Int? = null
            var receiptErrorDescription: String? = null
            if (!SUCCESS_HTTP_STATUSES.contains(walletBalance.restServiceResponse.httpCode)) {
                receiptStatus = ReceiptStatus.ERROR
                receiptErrorCode = walletBalance.restServiceResponse.httpCode.value()
                receiptErrorDescription = walletBalance.restServiceResponse.body
            }
            return TransactionReceipt(
                request.balanceRequest.requestType,
                request.balanceRequest.dateFrom,
                request.balanceRequest.dateTo,
                request.balanceRequest.operTime,
                receiptStatus,
                receiptErrorCode,
                receiptErrorDescription
            )
        }
    }
}