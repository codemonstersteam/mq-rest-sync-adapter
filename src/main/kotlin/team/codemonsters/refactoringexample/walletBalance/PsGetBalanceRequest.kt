package team.codemonsters.refactoringexample.walletBalance

import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.OffsetDateTime

data class PsGetBalanceRequest(
    val requestType: String,
    val dateFrom: OffsetDateTime,
    val dateTo: OffsetDateTime?,
    val operTime: OffsetDateTime
) {
    companion object {
        fun fromJsonString(
            jsonString: String
        ): Result<PsGetBalanceRequest> =
            Result.runCatching { JsonUtil.fromString(jsonString, PsGetBalanceRequest::class.java) }

    }
}
