package team.codemonsters.refactoringexample.mq

import com.fasterxml.jackson.annotation.JsonFormat
import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import java.time.OffsetDateTime

data class ReceiptPs(
    val requestType: String,
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
    val dateFrom: OffsetDateTime,
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
    val dateTo: OffsetDateTime?,
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
    val operTime: OffsetDateTime,
    val flag: ReceiptStatus,
    val errorCode: Int? = null,
    val description: String? = null
) {
}