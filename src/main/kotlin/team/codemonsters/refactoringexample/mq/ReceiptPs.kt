package team.codemonsters.refactoringexample.mq

import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import java.time.OffsetDateTime

data class ReceiptPs(
    val requestType: String,
    val dateFrom: OffsetDateTime,
    val dateTo: OffsetDateTime?,
    val operTime: OffsetDateTime,
    val flag: ReceiptStatus,
    val errorCode: Int? = null,
    val description: String? = null
)